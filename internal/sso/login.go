package sso

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/AndreZiviani/aws-fuzzy/internal/awsprofile"
	"github.com/AndreZiviani/aws-fuzzy/internal/securestorage"
	"github.com/AndreZiviani/aws-fuzzy/internal/tracing"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/common-fate/clio"
	opentracing "github.com/opentracing/opentracing-go"
)

func NewLogin(profile, token string, ask, verbose, url, noCache bool) *Login {
	login := Login{
		Profile: profile,
		MFATOTP: token,
		Ask:     ask,
		Verbose: verbose,
		Url:     url,
		NoCache: noCache,
	}

	return &login
}

func (p *Login) LoadProfiles() {
	awsProfiles, _ := awsprofile.LoadProfiles()
	p.profiles = *awsProfiles
}

func (p *Login) Execute(ctx context.Context) error {
	closer, err := tracing.InitTracing()
	if err != nil {
		return fmt.Errorf("failed to initialize tracing, %s", err)
	}
	defer func() { _ = closer.Close() }()

	tracer := opentracing.GlobalTracer()
	spanSso, ctx := opentracing.StartSpanFromContextWithTracer(ctx, tracer, "ssologincmd")
	defer spanSso.Finish()

	if p.Verbose {
		clio.SetLevelFromString("debug")
	}

	creds, err := p.GetCredentials(ctx)
	if err != nil {
		return err
	}

	p.PrintCredentials(creds)

	return err
}

func (p *Login) AskAuth() bool {
	if p.Ask {
		fmt.Fprintf(os.Stderr, "Authenticate again? (y/N) ")

		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')

		if text[0] != 'y' && text[0] != 'Y' {
			os.Exit(0)
		}

		return true
	}

	return false
}

func (p *Login) GetCredentials(ctx context.Context) (*aws.Credentials, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ssorolecreds")
	defer span.Finish()

	p.LoadProfiles()

	profile, err := p.GetProfile(p.Profile)
	if err != nil {
		return nil, err
	}

	credstore := securestorage.NewSecureSSOTokenStorage()

	// Check cached role credentials (unless NoCache)
	if p.NoCache {
		_ = credstore.SecureStorage.Clear(profile.Name)
	} else {
		var creds aws.Credentials
		if credstore.SecureStorage.Retrieve(profile.Name, &creds) == nil {
			if creds.HasKeys() && !creds.Expired() {
				return &creds, nil
			}
			_ = credstore.SecureStorage.Clear(profile.Name)
		}

		// Check parent profile cached credentials
		if len(profile.Parents) > 0 {
			var parentCreds aws.Credentials
			if credstore.SecureStorage.Retrieve(profile.Parents[0].Name, &parentCreds) == nil {
				creds, err := p.AssumeRoleWithCreds(ctx, &parentCreds)
				if err == nil {
					return &creds, nil
				}
				clio.Debugf("failed to use cached parent credentials: %s", err)
			}
		}
	}

	p.AskAuth()
	p.checkExpiredCreds(ctx)

	clio.Infof("Could not find cached credentials, refreshing...")

	// Get fresh credentials — AssumeTerminal handles SSO token management internally
	var creds aws.Credentials
	if profile.ProfileType == awsprofile.ProfileTypeIAM && profile.AWSConfig.MFASerial != "" {
		creds, err = p.LoginMFA(ctx)
	} else {
		creds, err = profile.AssumeTerminal(ctx, awsprofile.ConfigOpts{Duration: time.Hour, PrintOnly: p.Url})
	}
	if err != nil {
		return nil, err
	}

	// cache role credentials
	_ = credstore.SecureStorage.Store(profile.Name, creds)

	return &creds, nil
}

func (p *Login) AssumeRoleWithCreds(ctx context.Context, parentcreds *aws.Credentials) (aws.Credentials, error) {
	profile, err := p.GetProfile(p.Profile)
	if err != nil {
		return aws.Credentials{}, err
	}

	cfg, _ := NewAwsConfig(ctx, parentcreds)

	stsClient := sts.NewFromConfig(cfg)

	session, err := stsClient.AssumeRole(ctx,
		&sts.AssumeRoleInput{
			RoleArn:         &profile.AWSConfig.RoleARN,
			RoleSessionName: &profile.Name,
		})

	if err != nil {
		return aws.Credentials{}, err
	}

	creds := awsprofile.TypeCredsToAwsCreds(*session.Credentials)
	credstore := securestorage.NewSecureSSOTokenStorage()
	_ = credstore.SecureStorage.Store(profile.Name, creds)

	return creds, err

}
func (p *Login) LoginMFA(ctx context.Context) (aws.Credentials, error) {
	// Granted/AWS CLI only implements MFA authentication when assuming a role
	// Here we want to just login and get a session token for the user without assuming a role

	p.LoadProfiles()
	profile, err := p.GetProfile(p.Profile)
	if err != nil {
		return aws.Credentials{}, err
	}

	var creds aws.Credentials

	if len(profile.AWSConfig.SourceProfileName) > 0 {
		// check if we have cached credentials for source profile
		credstore := securestorage.NewSecureSSOTokenStorage()
		_ = credstore.SecureStorage.Retrieve(profile.AWSConfig.SourceProfileName, &creds)
	}

	cfg, _ := NewAwsConfig(ctx, &profile.AWSConfig.Credentials)
	stsClient := sts.NewFromConfig(cfg)

	if len(creds.AccessKeyID) == 0 {
		var mfatotp string
		if len(p.MFATOTP) == 0 {
			reader := bufio.NewReader(os.Stdin)
			fmt.Fprintf(os.Stderr, "MFA Token: ")
			mfatotp, _ = reader.ReadString('\n')
			mfatotp = strings.TrimSuffix(mfatotp, "\n")
		} else {
			mfatotp = p.MFATOTP
		}

		params := sts.GetSessionTokenInput{
			SerialNumber: &profile.AWSConfig.MFASerial,
			TokenCode:    &mfatotp,
		}
		session, err := stsClient.GetSessionToken(ctx, &params)
		if err != nil {
			return aws.Credentials{}, err
		}

		creds = awsprofile.TypeCredsToAwsCreds(*session.Credentials)
	} else {
		params := sts.AssumeRoleInput{
			RoleArn:         &profile.AWSConfig.RoleARN,
			RoleSessionName: &profile.Name,
		}
		session, err := stsClient.AssumeRole(ctx, &params)
		if err != nil {
			return aws.Credentials{}, err
		}

		creds = awsprofile.TypeCredsToAwsCreds(*session.Credentials)
	}

	return creds, err
}

func (p *Login) checkExpiredCreds(ctx context.Context) {
	if len(os.Getenv("AWS_ACCESS_KEY_ID")) == 0 {
		return
	}

	cfg, _ := NewAwsConfig(ctx, nil)

	stsClient := sts.NewFromConfig(cfg)

	_, err := stsClient.GetCallerIdentity(ctx, nil)

	if err != nil {
		// remove credentials from env var if expired
		_ = os.Unsetenv("AWS_SECURITY_TOKEN")
		_ = os.Unsetenv("AWS_SECRET_ACCESS_KEY")
		_ = os.Unsetenv("AWS_SESSION_TOKEN")
		_ = os.Unsetenv("AWS_ACCESS_KEY_ID")
	}
}
