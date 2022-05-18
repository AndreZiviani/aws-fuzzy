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
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/common-fate/granted/pkg/cfaws"
	"github.com/common-fate/granted/pkg/credstore"
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

	creds, err := p.GetCredentials(ctx)
	if err != nil {
		return err
	}

	p.PrintCredentials(creds)

	return err
}

func (p *Login) GetCredentials(ctx context.Context) (*aws.Credentials, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ssorolecreds")
	defer span.Finish()

	var profile *cfaws.CFSharedConfig
	var ok bool

	if profile, ok = p.profiles[p.Profile]; !ok {
		// mais p.profiles isnt initialized yet?
		p.LoadProfiles()

		// try again...
		if profile, ok = p.profiles[p.Profile]; !ok {
			return nil, errors.New(fmt.Sprintf("profile %s not found!", p.Profile))
		}

		// got it!
	}

	// if profile == nil {...
	// prompt for profile using fzf

	var creds aws.Credentials
	err := credstore.Retrieve(profile.Name, &creds)
	if err != nil {
		if p.Ask {
			reader := bufio.NewReader(os.Stdin)
			fmt.Fprintf(os.Stderr, "Authenticate again? (y/N) ")
			text, _ := reader.ReadString('\n')
			if text[0] != 'y' && text[0] != 'Y' {
				os.Exit(0)
			}
		}

		fmt.Fprintf(os.Stderr, "Could not find cached credentials, refreshing...\n")

		var creds aws.Credentials
		if profile.ProfileType == "AWS_IAM" && profile.AWSConfig.MFASerial != "" {
			creds, err = p.LoginMFA(ctx)
		} else {
			creds, err = profile.AssumeTerminal(ctx, cfaws.ConfigOpts{Duration: time.Hour})
		}

		if err != nil {
			return nil, err
		}

		credstore.Store(profile.Name, creds)

		return &creds, nil
	}

	if creds.HasKeys() && !creds.Expired() {
		return &creds, nil
	}

	credstore.Clear(profile.Name)

	return p.GetCredentials(ctx)
}

func (p *Login) GetProfile(profile string) (*cfaws.CFSharedConfig, error) {
	var tmp *cfaws.CFSharedConfig
	var ok bool
	if tmp, ok = p.profiles[profile]; !ok {
		return nil, errors.New(fmt.Sprintf("profile %s not found!", profile))
	}

	return tmp, nil
}

func (p *Login) GetProfileFromID(id string) (*cfaws.CFSharedConfig, error) {
	for _, v := range p.profiles {
		if id == v.AWSConfig.SSOAccountID {
			return v, nil
		}
	}

	return nil, errors.New(fmt.Sprintf("could not find a profile with account id %s!", id))
}

func (p *Login) PrintCredentials(creds *aws.Credentials) {
	fmt.Printf(
		"export AWS_ACCESS_KEY_ID='%s'\n"+
			"export AWS_SECRET_ACCESS_KEY='%s'\n"+
			"export AWS_SESSION_TOKEN='%s'\n"+
			"export AWS_SECURITY_TOKEN='%s'\n"+
			"export AWS_EXPIRES='%s'\n",
		creds.AccessKeyID,
		creds.SecretAccessKey,
		creds.SessionToken,
		creds.SessionToken,
		creds.Expires.String(),
	)
}

func (p *Login) LoginMFA(ctx context.Context) (aws.Credentials, error) {
	// Granted/AWS CLI only implements MFA authentication when assuming a role
	// Here we want to just login and get a session token for the user without assuming a role

	p.LoadProfiles()
	profile := p.profiles[p.Profile]

	reader := bufio.NewReader(os.Stdin)
	fmt.Fprintf(os.Stderr, "MFA Token: ")
	mfatotp, _ := reader.ReadString('\n')
	mfatotp = strings.TrimSuffix(mfatotp, "\n")

	cfg, _ := config.LoadDefaultConfig(ctx)
	stsClient := sts.NewFromConfig(cfg)

	params := sts.GetSessionTokenInput{
		SerialNumber: &profile.AWSConfig.MFASerial,
		TokenCode:    &mfatotp,
	}
	session, err := stsClient.GetSessionToken(ctx, &params)
	if err != nil {
		return aws.Credentials{}, err
	}

	c := session.Credentials

	creds := aws.Credentials{
		AccessKeyID:     *c.AccessKeyId,
		SecretAccessKey: *c.SecretAccessKey,
		SessionToken:    *c.SessionToken,
		Expires:         *c.Expiration,
	}

	return creds, err
}
