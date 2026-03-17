package awsprofile

import (
	"context"
	"os/exec"
	"time"

	"github.com/AndreZiviani/aws-fuzzy/internal/afconfig"
	"github.com/AndreZiviani/aws-fuzzy/internal/securestorage"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sso"
	ssotypes "github.com/aws/aws-sdk-go-v2/service/sso/types"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
	ssooidctypes "github.com/aws/aws-sdk-go-v2/service/ssooidc/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/common-fate/clio"
	"github.com/hako/durafmt"
	"github.com/pkg/errors"
	"gopkg.in/ini.v1"
)

// Implements Assumer
type AwsSsoAssumer struct {
}

func (asa *AwsSsoAssumer) AssumeTerminal(ctx context.Context, c *Profile, configOpts ConfigOpts) (aws.Credentials, error) {
	return c.SSOLogin(ctx, configOpts)
}

func (asa *AwsSsoAssumer) AssumeConsole(ctx context.Context, c *Profile, configOpts ConfigOpts) (aws.Credentials, error) {
	return c.SSOLogin(ctx, configOpts)
}

func (asa *AwsSsoAssumer) Type() string {
	return ProfileTypeSSO
}

// Matches the profile type on whether it is an sso profile by checking for ssoaccountid.
func (asa *AwsSsoAssumer) ProfileMatchesType(rawProfile *ini.Section, parsedProfile config.SharedConfig) bool {
	return parsedProfile.SSOAccountID != ""
}

func (c *Profile) SSOLogin(ctx context.Context, configOpts ConfigOpts) (aws.Credentials, error) {

	rootProfile := c
	requiresAssuming := false
	if len(c.Parents) > 0 {
		rootProfile = c.Parents[0]

		requiresAssuming = true
	}

	var startURL, ssoTokenKey string
	cfg := aws.NewConfig()

	if c.AWSConfig.SSOSession != nil {
		startURL = c.AWSConfig.SSOSession.SSOStartURL
		ssoTokenKey = c.AWSConfig.SSOSession.Name
		cfg.Region = c.AWSConfig.SSOSession.SSORegion
	} else {
		startURL = rootProfile.AWSConfig.SSOStartURL
		ssoTokenKey = startURL
		cfg.Region = rootProfile.AWSConfig.SSORegion
	}

	secureSSOTokenStorage := securestorage.NewSecureSSOTokenStorage()
	cachedToken := secureSSOTokenStorage.GetValidSSOToken(ctx, ssoTokenKey)
	var accessToken *string
	if cachedToken == nil {
		newSSOToken, err := SSODeviceCodeFlowFromStartUrl(ctx, *cfg, startURL, c.Name, configOpts.PrintOnly)
		if err != nil {
			return aws.Credentials{}, err
		}

		secureSSOTokenStorage.StoreSSOToken(ssoTokenKey, *newSSOToken)
		accessToken = &newSSOToken.AccessToken
	} else {
		accessToken = &cachedToken.AccessToken
	}

	// create sso client
	ssoClient := sso.NewFromConfig(*cfg)
	res, err := ssoClient.GetRoleCredentials(ctx, &sso.GetRoleCredentialsInput{AccessToken: accessToken, AccountId: &rootProfile.AWSConfig.SSOAccountID, RoleName: &rootProfile.AWSConfig.SSORoleName})
	if err != nil {
		var unauthorised *ssotypes.UnauthorizedException
		if errors.As(err, &unauthorised) {
			// possible error with the access token we used, in this case we should clear our cached token and request a new one if the user tries again
			secureSSOTokenStorage.ClearSSOToken(ssoTokenKey)
		}
		return aws.Credentials{}, err
	}
	rootCreds := TypeRoleCredsToAwsCreds(*res.RoleCredentials)
	credProvider := &CredProv{rootCreds}

	if requiresAssuming {
		// return creds, nil
		toAssume := append([]*Profile{}, c.Parents[1:]...)
		toAssume = append(toAssume, c)
		for i, p := range toAssume {
			region, err := c.Region(ctx)
			if err != nil {
				return aws.Credentials{}, err
			}
			// in order to support profiles which do not specify a region, we use the default region when assuming the role
			stsClient := sts.New(sts.Options{Credentials: aws.NewCredentialsCache(credProvider), Region: region})
			stsp := stscreds.NewAssumeRoleProvider(stsClient, p.AWSConfig.RoleARN, func(aro *stscreds.AssumeRoleOptions) {
				// all configuration goes in here for this profile
				if p.AWSConfig.RoleSessionName != "" {
					aro.RoleSessionName = p.AWSConfig.RoleSessionName
				} else {
					aro.RoleSessionName = sessionName()
				}
				if p.AWSConfig.MFASerial != "" {
					aro.SerialNumber = &p.AWSConfig.MFASerial
					aro.TokenProvider = MfaTokenProvider
				} else if c.AWSConfig.MFASerial != "" {
					aro.SerialNumber = &c.AWSConfig.MFASerial
					aro.TokenProvider = MfaTokenProvider
				}
				aro.Duration = configOpts.Duration
				if p.AWSConfig.ExternalID != "" {
					aro.ExternalID = &p.AWSConfig.ExternalID
				}
			})
			stsCreds, err := stsp.Retrieve(ctx)
			if err != nil {
				return aws.Credentials{}, err
			}
			// only print for sub assumes because the final credentials are printed at the end of the assume command
			// this is here for visibility in to role traversals when assuming a final profile with sso
			if i < len(toAssume)-1 {
				durationDescription := durafmt.Parse(time.Until(stsCreds.Expires) * time.Second).LimitFirstN(1).String()
				clio.Successf("Assumed parent profile: [%s](%s) session credentials will expire %s", p.Name, region, durationDescription)
			}
			credProvider = &CredProv{stsCreds}

		}
	}
	return credProvider.Credentials, nil

}

// getOrRegisterClient returns a cached client registration if available, or registers a new one.
func getOrRegisterClient(ctx context.Context, ssooidcClient *ssooidc.Client, cfg aws.Config, startUrl string, grantTypes []string) (*securestorage.ClientRegistration, error) {
	regStore := securestorage.NewSecureClientRegistrationStorage()
	cached := regStore.GetValidRegistration(startUrl)
	if cached != nil {
		clio.Debugf("using cached SSO client registration (expires %s)", cached.RegistrationExpiresAt.Format(time.RFC3339))
		return cached, nil
	}

	register, err := ssooidcClient.RegisterClient(ctx, &ssooidc.RegisterClientInput{
		ClientName: aws.String("aws-fuzzy"),
		ClientType: aws.String("public"),
		GrantTypes: grantTypes,
		Scopes:     []string{securestorage.ScopeAccountAccess},
	})
	if err != nil {
		return nil, err
	}

	reg := &securestorage.ClientRegistration{
		ClientID:              aws.ToString(register.ClientId),
		ClientSecret:          aws.ToString(register.ClientSecret),
		RegistrationExpiresAt: time.Unix(register.ClientSecretExpiresAt, 0),
		Region:                cfg.Region,
		AuthorizationEndpoint: aws.ToString(register.AuthorizationEndpoint),
		TokenEndpoint:         aws.ToString(register.TokenEndpoint),
	}

	regStore.StoreRegistration(startUrl, *reg)
	return reg, nil
}

// openBrowserForSSO opens the given URL in the user's browser (custom or default).
func openBrowserForSSO(afcfg afconfig.Config, url, profile string, printOnly bool) error {
	if afcfg.CustomSSOBrowserPath != "" {
		cmd := exec.Command(afcfg.CustomSSOBrowserPath, url)
		err := cmd.Start()
		if err != nil {
			clio.Debug(err.Error())
		} else {
			err = cmd.Process.Release()
			if err != nil {
				clio.Debug(err.Error())
			}
		}
		return nil
	}
	return LaunchBrowser(url, profile, "sso", printOnly)
}

// SSODeviceCodeFlowFromStartUrl contains all the steps to complete a device code flow to retrieve an SSO token
func SSODeviceCodeFlowFromStartUrl(ctx context.Context, cfg aws.Config, startUrl string, profile string, printOnly bool) (*securestorage.SSOToken, error) {
	afcfg, err := afconfig.NewLoadedConfig()
	if err != nil {
		return nil, err
	}

	ssooidcClient := ssooidc.NewFromConfig(cfg)

	reg, err := getOrRegisterClient(ctx, ssooidcClient, cfg, startUrl, []string{GrantTypeDeviceCode, GrantTypeRefreshToken})
	if err != nil {
		return nil, err
	}

	deviceAuth, err := ssooidcClient.StartDeviceAuthorization(ctx, &ssooidc.StartDeviceAuthorizationInput{
		ClientId:     &reg.ClientID,
		ClientSecret: &reg.ClientSecret,
		StartUrl:     aws.String(startUrl),
	})
	if err != nil {
		return nil, err
	}

	url := aws.ToString(deviceAuth.VerificationUriComplete)
	if !printOnly {
		clio.Info("If the browser does not open automatically, please open this link:")
		clio.Info(url)
	}

	if err := openBrowserForSSO(afcfg, url, profile, printOnly); err != nil {
		return nil, err
	}

	clio.Info("Awaiting authentication in the browser...")
	token, err := PollToken(ctx, ssooidcClient, reg.ClientSecret, reg.ClientID, *deviceAuth.DeviceCode, PollingConfig{CheckInterval: time.Second * 2, TimeoutAfter: time.Minute * 2})
	if err != nil {
		return nil, err
	}

	return &securestorage.SSOToken{
		AccessToken:           aws.ToString(token.AccessToken),
		Expiry:                time.Now().Add(time.Duration(token.ExpiresIn) * time.Second),
		ClientID:              reg.ClientID,
		ClientSecret:          reg.ClientSecret,
		RegistrationExpiresAt: reg.RegistrationExpiresAt,
		Region:                cfg.Region,
		RefreshToken:          token.RefreshToken,
	}, nil
}

var ErrTimeout error = errors.New("polling for device authorization token timed out")

type PollingConfig struct {
	CheckInterval time.Duration
	TimeoutAfter  time.Duration
}

// PollToken will poll for a token and return it once the authentication/authorization flow has been completed in the browser
func PollToken(ctx context.Context, c *ssooidc.Client, clientSecret string, clientID string, deviceCode string, cfg PollingConfig) (*ssooidc.CreateTokenOutput, error) {
	start := time.Now()
	for {
		time.Sleep(cfg.CheckInterval)

		token, err := c.CreateToken(ctx, &ssooidc.CreateTokenInput{
			ClientId:     &clientID,
			ClientSecret: &clientSecret,
			DeviceCode:   &deviceCode,
			GrantType:    aws.String(GrantTypeDeviceCode),
			Scope:        []string{securestorage.ScopeAccountAccess},
		})
		var pendingAuth *ssooidctypes.AuthorizationPendingException
		if err == nil {
			return token, nil
		} else if !errors.As(err, &pendingAuth) {
			return nil, err
		}

		if time.Now().After(start.Add(cfg.TimeoutAfter)) {
			return nil, ErrTimeout
		}
	}
}
