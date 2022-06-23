package sso

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/AndreZiviani/aws-fuzzy/internal/tracing"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/common-fate/granted/pkg/cfaws"
	"github.com/common-fate/granted/pkg/credstore"
	"github.com/common-fate/granted/pkg/debug"
	opentracing "github.com/opentracing/opentracing-go"
)

func (p *Login) LoadProfiles() {
	awsProfiles, _ := cfaws.GetProfilesFromDefaultSharedConfig(context.TODO())
	p.profiles = awsProfiles
}

func (p *Login) Execute(args []string) error {

	ctx := context.Background()

	closer, err := tracing.InitTracing()
	if err != nil {
		fmt.Printf("failed to initialize tracing, %s\n", err)
	}
	defer closer.Close()

	tracer := opentracing.GlobalTracer()
	spanSso, ctx := opentracing.StartSpanFromContextWithTracer(ctx, tracer, "ssologincmd")
	defer spanSso.Finish()

	if p.Verbose {
		// enable granted debug
		debug.CliVerbosity = debug.VerbosityDebug
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

	var creds aws.Credentials
	var profile *cfaws.CFSharedConfig
	var ok bool

	if profile, ok = p.profiles[p.Profile]; !ok {
		return nil, errors.New(fmt.Sprintf("profile %s not found!", p.Profile))
	}

	// if profile == nil {...
	// prompt for profile using fzf

	err := credstore.Retrieve(profile.Name, &creds)

	if err == nil {
		// return cached credentials

		// check if credentials are expired
		if creds.HasKeys() && !creds.Expired() {
			return &creds, nil
		}

		credstore.Clear(profile.Name)

		return p.GetCredentials(ctx)
	}

	// we dont have cached credentials for this profile

	if len(profile.Parents) > 0 {
		// this profile uses a parent profile, check if we have cached credentials for that
		err := credstore.Retrieve(profile.Parents[0].Name, &creds)
		if err == nil {
			// yes we have
			creds, err = p.AssumeRoleWithCreds(ctx, &creds)
			if err == nil {
				return &creds, err
			}

			fmt.Fprintf(os.Stderr, "Something went wrong... Trying again. Error:\n%s\n", err)
		}
	}

	p.AskAuth()

	// check if we have expired credential set in env vars
	p.checkExpiredCreds(ctx)

	fmt.Fprintf(os.Stderr, "Could not find cached credentials, refreshing...\n")

	if profile.ProfileType == "AWS_IAM" && profile.AWSConfig.MFASerial != "" {
		// IAM profile with MFA
		creds, err = p.LoginMFA(ctx)
	} else {
		// Everything else
		creds, err = profile.AssumeTerminal(ctx, cfaws.ConfigOpts{Duration: time.Hour})
	}

	if err != nil {
		return nil, err
	}

	// cache credentials
	credstore.Store(profile.Name, creds)

	return &creds, nil

}

func (p *Login) AssumeRoleWithCreds(ctx context.Context, parentcreds *aws.Credentials) (aws.Credentials, error) {
	profile := p.profiles[p.Profile]

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

	creds := cfaws.TypeCredsToAwsCreds(*session.Credentials)
	credstore.Store(profile.Name, creds)

	return creds, err

}
func (p *Login) LoginMFA(ctx context.Context) (aws.Credentials, error) {
	// Granted/AWS CLI only implements MFA authentication when assuming a role
	// Here we want to just login and get a session token for the user without assuming a role

	p.LoadProfiles()
	profile := p.profiles[p.Profile]

	var creds aws.Credentials
	var err error

	if len(profile.AWSConfig.SourceProfileName) > 0 {
		// check if we have cached credentials for source profile
		credstore.Retrieve(profile.AWSConfig.SourceProfileName, &creds)
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

		creds = cfaws.TypeCredsToAwsCreds(*session.Credentials)
	} else {
		params := sts.AssumeRoleInput{
			RoleArn:         &profile.AWSConfig.RoleARN,
			RoleSessionName: &profile.Name,
		}
		session, err := stsClient.AssumeRole(ctx, &params)
		if err != nil {
			return aws.Credentials{}, err
		}

		creds = cfaws.TypeCredsToAwsCreds(*session.Credentials)
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
		os.Unsetenv("AWS_SECURITY_TOKEN")
		os.Unsetenv("AWS_SECRET_ACCESS_KEY")
		os.Unsetenv("AWS_SESSION_TOKEN")
		os.Unsetenv("AWS_ACCESS_KEY_ID")
	}
}
