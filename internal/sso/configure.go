package sso

import (
	"bufio"
	"context"
	"fmt"
	"github.com/AndreZiviani/aws-fuzzy/internal/tracing"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sso"
	opentracing "github.com/opentracing/opentracing-go"
	"os"
	"strings"
)

func ConfigureProfiles(ctx context.Context) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ssoprofiles")
	defer span.Finish()

	reader := bufio.NewReader(os.Stdin)

	//configPath := fmt.Sprintf("%s/.aws/config", common.UserHomeDir)

	fmt.Print("Enter SSO start url: ")
	startUrl, _ := reader.ReadString('\n')
	startUrl = strings.Replace(startUrl, "\n", "", -1)

	fmt.Print("Default region: ")
	region, _ := reader.ReadString('\n')
	region = strings.Replace(region, "\n", "", -1)

	// Create a dummy default in order to be able to authenticate and list accounts
	profiles := NewSsoProfiles()
	profiles["default"] = AwsProfile{startUrl, region, "000000000000", "dummy"}

	err := WriteSsoProfiles(profiles)
	if err != nil {
		return err
	}

	// Authenticate
	ssocreds, err := SsoLogin(ctx, false)
	if err != nil {
		return err
	}

	cfg, err := NewAwsConfig(ctx, nil)
	if err != nil {
		return err
	}

	ssoclient := sso.NewFromConfig(cfg)

	// List available account
	accounts, err := ssoclient.ListAccounts(ctx,
		&sso.ListAccountsInput{
			AccessToken: ssocreds.AccessToken,
			MaxResults:  aws.Int32(100),
		},
	)
	if err != nil {
		fmt.Printf("failed to list accounts, %s\n", err)
		return err
	}

	for _, account := range accounts.AccountList {
		// List Available roles in each account
		roles, err := ssoclient.ListAccountRoles(ctx,
			&sso.ListAccountRolesInput{
				AccessToken: ssocreds.AccessToken,
				AccountId:   account.AccountId,
				MaxResults:  aws.Int32(100),
			},
		)
		if err != nil {
			fmt.Printf("failed to list account roles, %s\n", err)
			return err
		}

		// TODO: ask user which role to use if they have more than one role on this account
		role := roles.RoleList[0].RoleName

		tmp := strings.ReplaceAll(*account.AccountName, " ", "_")
		profile := fmt.Sprintf("profile %s", tmp)
		profiles[profile] = AwsProfile{startUrl, region, *account.AccountId, *role}
	}

	// Save complete profile config
	err = WriteSsoProfiles(profiles)
	if err != nil {
		return err
	}

	return nil
}

func (p *ConfigureCommand) Execute(args []string) error {
	ctx := context.Background()

	closer, err := tracing.InitTracing()
	if err != nil {
		fmt.Printf("failed to initialize tracing, %s\n", err)
	}
	defer closer.Close()

	tracer := opentracing.GlobalTracer()
	spanSso, ctx := opentracing.StartSpanFromContextWithTracer(ctx, tracer, "ssoconfigure")
	defer spanSso.Finish()
	return ConfigureProfiles(ctx)
}
