package sso

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	ssotypes "github.com/aws/aws-sdk-go-v2/service/sso/types"
	ststypes "github.com/aws/aws-sdk-go-v2/service/sts/types"
	"gopkg.in/ini.v1"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

var (
	configPath = fmt.Sprintf("%s/.aws/config", os.Getenv("HOME"))
)

type AwsProfile struct {
	StartUrl  string `ini:"sso_start_url"`
	Region    string `ini:"sso_region"`
	AccountId string `ini:"sso_account_id"`
	Role      string `ini:"sso_role_name"`
}

type RoleCredentials struct {
	AccessKey    *string
	SecretKey    *string
	SessionToken *string
	Expires      CustomTime
}

type SsoDeviceCredentials struct {
	ClientId     *string    `json:"clientId"`
	ClientSecret *string    `json:"clientSecret"`
	ExpiresAt    CustomTime `json:"expiresAt"`
}

type SsoSessionCredentials struct {
	StartUrl    *string    `json:"startUrl"`
	Region      string     `json:"region"`
	AccessToken *string    `json:"accessToken"`
	ExpiresAt   CustomTime `json:"expiresAt"`
}

type CustomTime struct {
	time.Time
}

type CredentialsProvider struct {
	//*ssotypes.RoleCredentials
	*ststypes.Credentials
}

const ctLayout = "2006-01-02T15:04:05UTC"

var nilTime = (time.Time{}).UnixNano()

func (ct *CustomTime) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		ct.Time = time.Time{}
		return
	}
	ct.Time, err = time.Parse(ctLayout, s)
	return
}

func (ct *CustomTime) MarshalJSON() ([]byte, error) {
	if ct.Time.UnixNano() == nilTime {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", ct.Time.Format(ctLayout))), nil
}

func (cp CredentialsProvider) Retrieve(c context.Context) (aws.Credentials, error) {
	if cp.Credentials == nil {
		return aws.Credentials{}, errors.New("invalid credentials")
	}

	return aws.Credentials{
		AccessKeyID:     aws.ToString(cp.AccessKeyId),
		SecretAccessKey: aws.ToString(cp.SecretAccessKey),
		SessionToken:    aws.ToString(cp.SessionToken),
		Expires:         aws.ToTime(cp.Expiration),
	}, nil
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

func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		fmt.Printf("unable to open browser, %v\n", err)
	}

}

func GetAccount(account string) (*AwsProfile, error) {
	profiles, _ := LoadSsoProfiles()

	if profile, ok := profiles[account]; ok {
		// account is a profile name
		return &profile, nil
	}

	// account is an account id
	for _, v := range profiles {
		if v.AccountId == account {
			return &v, nil
		}
	}

	return nil, errors.New("could not find account")
}

func NewSsoProfiles() map[string]AwsProfile {
	return make(map[string]AwsProfile)
}

func WriteSsoProfiles(profiles map[string]AwsProfile) error {

	if _, err := os.Stat(configPath); err == nil {
		// found existing config file, backup before proceeding
		err := CopyFile(configPath, fmt.Sprintf("%s.bkp", configPath))
		if err != nil {
			fmt.Printf("could not backup config, %v\n", err)
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
func LoadSsoProfiles() (map[string]AwsProfile, error) {
	// Load aws config
	cfg, _ := ini.Load(configPath)

	cfg.DeleteSection("DEFAULT")

	profiles := NewSsoProfiles()

	for _, v := range cfg.Sections() {
		profileName := v.Name()
		profileName = strings.ReplaceAll(profileName, "profile ", "")

		profile := AwsProfile{}
		err := v.MapTo(&profile)
		if err != nil {
			fmt.Printf("failed to load profiles, %s\n", err)
			return nil, err
		}

		profiles[profileName] = profile
	}

	return profiles, nil

}

func NewAwsConfig(ctx context.Context, creds *ssotypes.RoleCredentials, retryables ...string) (aws.Config, error) {
	var cfg aws.Config
	var err error

	if len(retryables) > 0 {

		cfg, err = config.LoadDefaultConfig(ctx,
			// default if not specified
			config.WithRegion("us-east-1"), config.WithRetryer(func() aws.Retryer {
				tmp := retry.AddWithMaxBackoffDelay(retry.NewStandard(), time.Second*1)
				tmp = retry.AddWithMaxAttempts(tmp, 0)
				return retry.AddWithErrorCodes(tmp, retryables...)
			}))

		if err != nil {
			fmt.Printf("unable to load SDK config, %v\n", err)
			return aws.Config{}, err
		}

	} else {

		// Load AWS config defaulting to us-east-1 region if not specified
		cfg, err = config.LoadDefaultConfig(ctx,
			// default if not specified
			config.WithRegion("us-east-1"))
		if err != nil {
			fmt.Printf("unable to load SDK config, %v\n", err)
			return aws.Config{}, err
		}
	}

	if creds != nil {
		tmp := time.Unix(creds.Expiration, 0)
		cfg.Credentials = CredentialsProvider{
			Credentials: &ststypes.Credentials{
				AccessKeyId:     creds.AccessKeyId,
				SecretAccessKey: creds.SecretAccessKey,
				SessionToken:    creds.SessionToken,
				Expiration:      &tmp,
			},
		}
	}

	return cfg, nil
}
