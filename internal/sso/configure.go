package sso

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/AndreZiviani/aws-fuzzy/internal/common"
	"github.com/AndreZiviani/aws-fuzzy/internal/tracing"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/common-fate/granted/pkg/cfaws"
	opentracing "github.com/opentracing/opentracing-go"
	"gopkg.in/ini.v1"
)

type AwsProfile struct {
	StartUrl          string `ini:"granted_sso_start_url"`
	Region            string `ini:"granted_sso_region"`
	AccountId         string `ini:"granted_sso_account_id"`
	Role              string `ini:"granted_sso_role_name"`
	CredentialProcess string `ini:"credential_process"`
}

func (p *Configure) GetAccountAccess(ctx context.Context, startURL string, region string) (map[string]AwsProfile, error) {
	cfg, err := NewAwsConfig(ctx, nil)
	ssoclient := sso.NewFromConfig(cfg)

	secureSSOTokenStorage := NewSecureSSOTokenStorage()
	ssoToken := secureSSOTokenStorage.GetValidSSOToken(startURL)
	if ssoToken == nil {
		ssoToken, err = cfaws.SSODeviceCodeFlowFromStartUrl(ctx, cfg, startURL)
		if err != nil {
			return nil, err
		}
	}

	// List available account
	accounts, err := ssoclient.ListAccounts(ctx,
		&sso.ListAccountsInput{
			AccessToken: &ssoToken.AccessToken,
			MaxResults:  aws.Int32(100),
		},
	)
	if err != nil {
		fmt.Printf("failed to list accounts, %s\n", err)
		return map[string]AwsProfile{}, err
	}

	profiles := NewSsoProfiles()
	for _, account := range accounts.AccountList {
		// List Available roles in each account
		roles, err := ssoclient.ListAccountRoles(ctx,
			&sso.ListAccountRolesInput{
				AccessToken: &ssoToken.AccessToken,
				AccountId:   account.AccountId,
				MaxResults:  aws.Int32(100),
			},
		)
		if err != nil {
			fmt.Printf("failed to list account roles, %s\n", err)
			return map[string]AwsProfile{}, err
		}

		role := aws.ToString(roles.RoleList[0].RoleName)
		reader := bufio.NewReader(os.Stdin)

		if len(roles.RoleList) > 1 {
			fmt.Printf("Found %d roles for account %s:\n", len(roles.RoleList), *account.AccountName)
			sort.Slice(roles.RoleList, func(i, j int) bool {
				return strings.ToLower(aws.ToString(roles.RoleList[i].RoleName)) < strings.ToLower(aws.ToString(roles.RoleList[j].RoleName))
			})
			for _, v := range roles.RoleList {
				fmt.Println(aws.ToString(v.RoleName))
			}

			fmt.Printf("Which one do you want? (default: %s) ", aws.ToString(roles.RoleList[0].RoleName))
			text, _ := reader.ReadString('\n')

			if text != "\n" {
				role = text
			}

			// TODO: check for typo in role name?
		}

		fmt.Printf("Profile name for account %s: (defaults to account name) ", *account.AccountName)
		text, _ := reader.ReadString('\n')
		text = strings.TrimSuffix(text, "\n")

		tmp := strings.ReplaceAll(*account.AccountName, " ", "_")
		if len(text) != 0 {
			tmp = text
		}

		profile := fmt.Sprintf("profile %s", tmp)
		process := fmt.Sprintf("%s sso credential-process --profile %s", os.Args[0], tmp)
		profiles[profile] = AwsProfile{startURL, region, *account.AccountId, role, process}
	}

	return profiles, err
}

func (p *Configure) ConfigureProfiles(ctx context.Context) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ssoprofiles")
	defer span.Finish()

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("SSO start url: ")
	startURL, _ := reader.ReadString('\n')
	startURL = strings.Replace(startURL, "\n", "", -1)

	fmt.Print("SSO region: ")
	region, _ := reader.ReadString('\n')
	region = strings.Replace(region, "\n", "", -1)

	profiles, err := p.GetAccountAccess(ctx, startURL, region)

	// Save complete profile config
	err = WriteSsoProfiles(profiles)
	if err != nil {
		return err
	}

	return nil
}

func NewSsoProfiles() map[string]AwsProfile {
	return make(map[string]AwsProfile)
}

func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	err = out.Sync()
	return err
}

func WriteSsoProfiles(profiles map[string]AwsProfile) error {
	configDir := fmt.Sprintf("%s/.aws", common.UserHomeDir)
	configPath := fmt.Sprintf("%s/config", configDir)

	if _, err := os.Stat(configPath); err == nil {
		// found existing config file, backup before proceeding
		err := CopyFile(configPath, fmt.Sprintf("%s.bkp", configPath))
		if err != nil {
			fmt.Printf("could not backup config, %v\n", err)
			return err
		}
	} else {
		err = os.Mkdir(configDir, 0700)
		if err != nil {
			return err
		}
	}

	c := ini.Empty()

	for k, v := range profiles {
		s, _ := c.NewSection(k)
		_ = s.ReflectFrom(&v)
	}

	f, err := os.Create(configPath)

	if err != nil {
		fmt.Printf("failed to write SSO profiles, %s\n", err)
		return err
	}
	defer f.Close()

	_, err = c.WriteTo(f)
	if err != nil {
		fmt.Printf("failed to write SSO profiles, %s\n", err)
		return err
	}

	return nil
}

func (p *Configure) Execute(args []string) error {
	ctx := context.Background()

	closer, err := tracing.InitTracing()
	if err != nil {
		fmt.Printf("failed to initialize tracing, %s\n", err)
	}
	defer closer.Close()

	tracer := opentracing.GlobalTracer()
	spanSso, ctx := opentracing.StartSpanFromContextWithTracer(ctx, tracer, "ssoconfigure")
	defer spanSso.Finish()

	return p.ConfigureProfiles(ctx)
}
