package sso

import (
	"context"
	"fmt"

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
