package sso

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/AndreZiviani/aws-fuzzy/internal/cache"
	"github.com/AndreZiviani/aws-fuzzy/internal/tracing"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials/ssocreds"
	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
	opentracing "github.com/opentracing/opentracing-go"
	"io/ioutil"
	"os"
	"time"
)

func checkExpired(kind string, path string) (interface{}, error) {
	jsonFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	jsonBytes, _ := ioutil.ReadAll(jsonFile)

	now := time.Now()
	switch kind {
	case "device":
		creds := SsoDeviceCredentials{}
		json.Unmarshal(jsonBytes, &creds)
		if creds.ExpiresAt.Before(now) {
			// invalid credentials, show login again
			return nil, errors.New("expired credentials")
		} else {
			return creds, nil
		}
	case "session":
		creds := SsoSessionCredentials{}
		json.Unmarshal(jsonBytes, &creds)
		if creds.ExpiresAt.Before(now) {
			// invalid credentials, show login again
			return nil, errors.New("expired credentials")
		} else {
			return creds, nil
		}
	default:
		return nil, errors.New("unsupported file type")
	}

}

func checkCachedDevice(cfg aws.Config) (SsoDeviceCredentials, error) {
	creds, err := checkExpired("device", fmt.Sprintf("%s/.aws/sso/cache/botocore-client-id-%s.json", os.Getenv("HOME"), cfg.Region))

	if creds == nil {
		return SsoDeviceCredentials{}, err
	}

	return creds.(SsoDeviceCredentials), err
}

func getSessionFileName(startUrl *string) string {
	sha := sha1.New()
	sha.Write([]byte(*startUrl))

	hash := sha.Sum(nil)

	return string(hash)
}

func checkCachedSession(cfg aws.Config, startUrl *string) (SsoSessionCredentials, error) {
	hash := getSessionFileName(startUrl)

	creds, err := checkExpired("session", fmt.Sprintf("%s/.aws/sso/cache/%x.json", os.Getenv("HOME"), hash))

	if creds == nil {
		return SsoSessionCredentials{}, err
	}

	return creds.(SsoSessionCredentials), err
}

func cacheCredentials(device *SsoDeviceCredentials, session *SsoSessionCredentials) error {
	hash := getSessionFileName(session.StartUrl)

	file, _ := json.Marshal(session)
	_ = ioutil.WriteFile(fmt.Sprintf("%s/.aws/sso/cache/%x.json", os.Getenv("HOME"), hash), file, 0600)

	file, _ = json.Marshal(device)
	_ = ioutil.WriteFile(fmt.Sprintf("%s/.aws/sso/cache/botocore-client-id-%s.json", os.Getenv("HOME"), session.Region), file, 0600)
	return nil

}

func NewSsoCredentials(ctx context.Context, cfg aws.Config, oidc *ssooidc.Client, startUrl *string) (*SsoSessionCredentials, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ssocreds")
	defer span.Finish()

	sessionCredential, err := checkCachedSession(cfg, startUrl)

	if err == nil {
		return &sessionCredential, nil
	}

	// we do not have valid credential, authenticating again
	now := time.Now()

	deviceCredential, err := checkCachedDevice(cfg)

	if err != nil {
		// device credential is expired, renewing
		oidcClient, err := oidc.RegisterClient(ctx,
			&ssooidc.RegisterClientInput{
				ClientName: aws.String(fmt.Sprintf("botocore-client-%d", now.Unix())),
				ClientType: aws.String("public"), // only time available
			},
		)
		if err != nil {
			fmt.Printf("unable to register client, %v\n", err)
			return nil, err
		}

		deviceCredential.ClientId = oidcClient.ClientId
		deviceCredential.ClientSecret = oidcClient.ClientSecret
		deviceCredential.ExpiresAt = rfc3339{time.Unix(oidcClient.ClientSecretExpiresAt, 0)}

	}

	oidcAuth, err := oidc.StartDeviceAuthorization(ctx,
		&ssooidc.StartDeviceAuthorizationInput{
			ClientId:     deviceCredential.ClientId,
			ClientSecret: deviceCredential.ClientSecret,
			StartUrl:     startUrl,
		},
	)
	if err != nil {
		fmt.Printf("unable to start device registration, %v\n", err)
		return nil, err
	}

	// Prompt user to signin using the browser
	fmt.Printf(
		"AWS SSO login required.\n"+
			"Attempting to open the SSO authorization page in your default browser.\n"+
			"If the browser does not open or you wish to use a different device to\n"+
			"authorize this request, open the following URL:\n\n"+
			"%v\n\n", *oidcAuth.VerificationUriComplete)

	openBrowser(*oidcAuth.VerificationUriComplete)

	token, err := oidc.CreateToken(ctx,
		&ssooidc.CreateTokenInput{
			ClientId:     deviceCredential.ClientId,
			ClientSecret: deviceCredential.ClientSecret,
			DeviceCode:   oidcAuth.DeviceCode,
			GrantType:    aws.String("urn:ietf:params:oauth:grant-type:device_code"), // Only grant type available
		},
	)
	if err != nil {
		fmt.Printf("unable to create token, %v\n", err)
		return nil, err
	}

	sessionCredential = SsoSessionCredentials{startUrl,
		cfg.Region,
		token.AccessToken,
		rfc3339{now.Add(time.Duration(token.ExpiresIn) * time.Second)},
	}

	err = cacheCredentials(&deviceCredential, &sessionCredential)
	if err != nil {
		fmt.Printf("unable to cache credentials, %v\n", err)
	}

	return &sessionCredential, nil

}

func SsoLogin(ctx context.Context) (*SsoSessionCredentials, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ssologin")
	defer span.Finish()

	// Load profiles to get StartUrl
	profiles, err := LoadSsoProfiles()
	if err != nil {
		return nil, err
	}

	var startUrl *string
	if defaultProfile, ok := profiles["default"]; ok {
		startUrl = &defaultProfile.StartUrl
	} else {
		fmt.Println("missing starturl from aws config, aborting...")
		return nil, errors.New("missing starturl from aws config, aborting...")
	}

	// New AWS config with custom retryer
	cfg, err := NewAwsConfig(ctx, nil, "AuthorizationPendingException")
	if err != nil {
		return nil, err
	}

	oidc := ssooidc.NewFromConfig(cfg)

	return NewSsoCredentials(ctx, cfg, oidc, startUrl)
}

func GetCredentials(ctx context.Context, profile string) (*aws.Credentials, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ssorolecreds")
	defer span.Finish()

	c, _ := cache.New("sso")
	j, err := c.Fetch(profile)

	creds := aws.Credentials{}
	if err == nil {
		// We have valid cached credentials
		err = json.Unmarshal([]byte(j), &creds)
		//PrintCredentials(creds)
		return &creds, nil
	}

	cfg, err := NewAwsConfig(ctx, nil)
	if err != nil {
		return nil, err
	}

	profiles, err := LoadSsoProfiles()
	if err != nil {
		return nil, err
	}

	ssoclient := sso.NewFromConfig(cfg)

	account := profiles[profile]
	provider := ssocreds.New(ssoclient, account.AccountId, account.Role, account.StartUrl)

	creds, err = provider.Retrieve(ctx)
	if err != nil {
		fmt.Printf("failed to get role credentials, %s\n", err)
		SsoLogin(ctx)
		return GetCredentials(ctx, profile)
	}

	tmp, _ := json.Marshal(creds)
	c.Save(profile, string(tmp), creds.Expires.Sub(time.Now()))

	return &creds, nil

}

func PrintCredentials(creds *aws.Credentials) {
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

func (p *LoginCommand) Execute(args []string) error {

	ctx := context.Background()

	closer, err := tracing.InitTracing()
	if err != nil {
		fmt.Printf("failed to initialize tracing, %s\n", err)
	}
	defer closer.Close()

	tracer := opentracing.GlobalTracer()
	spanSso, ctx := opentracing.StartSpanFromContextWithTracer(ctx, tracer, "ssologincmd")
	defer spanSso.Finish()

	creds, err := GetCredentials(ctx, p.Profile)
	PrintCredentials(creds)

	return err
}
