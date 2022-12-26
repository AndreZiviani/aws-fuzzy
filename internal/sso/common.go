package sso

import (
	"context"
	"errors"
	"fmt"

	"github.com/AndreZiviani/aws-fuzzy/internal/awsprofile"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

func NewAwsConfig(ctx context.Context, creds *aws.Credentials, opts ...func(*config.LoadOptions) error) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		fmt.Printf("unable to load SDK config, %v\n", err)
		return aws.Config{}, err
	}

	if creds != nil {
		cfg.Credentials = credentials.NewStaticCredentialsProvider(
			creds.AccessKeyID,
			creds.SecretAccessKey,
			creds.SessionToken,
		)
	}

	return cfg, nil
}

func (p *Login) GetProfile(profile string) (*awsprofile.Profile, error) {
	p.profiles.LoadInitialisedProfile(context.TODO(), "default")
	return p.profiles.LoadInitialisedProfile(context.TODO(), profile)
}

func (p *Login) GetProfileFromID(id string) (*awsprofile.Profile, error) {
	for _, v := range p.profiles.ProfileNames {
		profile, _ := p.GetProfile(v)
		if profile.ProfileType == "AWS_SSO" {
			if id == profile.AWSConfig.SSOAccountID {
				return profile, nil
			}
		} else if profile.ProfileType == "AWS_IAM" {
			// we dont have a simple way to find out the account id
			// we could extract the account id from the role arn but not all profiles assume a role
			// another option is querying STS but it will probably be slow
			continue
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
