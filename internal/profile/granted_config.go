package cfaws

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/common-fate/clio"
	"gopkg.in/ini.v1"
)

func ParseCustomSSOProfile(ctx context.Context, profile *Profile) (*config.SharedConfig, error) {
	err := IsValidCustomProfile(profile.RawConfig)
	if err != nil {
		return nil, err
	}
	cfg, err := config.LoadSharedConfigProfile(ctx, profile.Name, func(lsco *config.LoadSharedConfigOptions) { lsco.ConfigFiles = []string{profile.File} })
	if err != nil {
		return nil, err
	}
	item, err := profile.RawConfig.GetKey(cfg.AppName + "_sso_account_id")
	if err != nil {
		return nil, err
	}
	cfg.SSOAccountID = item.Value()
	item, err = profile.RawConfig.GetKey(cfg.AppName + "_sso_region")
	if err != nil {
		return nil, err
	}
	cfg.SSORegion = item.Value()
	item, err = profile.RawConfig.GetKey(cfg.AppName + "_sso_role_name")
	if err != nil {
		return nil, err
	}
	cfg.SSORoleName = item.Value()
	item, err = profile.RawConfig.GetKey(cfg.AppName + "_sso_start_url")
	if err != nil {
		return nil, err
	}

	// sanity check to verify if the provided value is a valid url
	_, err = url.ParseRequestURI(item.Value())
	if err != nil {
		clio.Debug(err)
		return nil, fmt.Errorf("invalid value '%s' provided for '%s_sso_start_url'", item.Value(), cfg.AppName)
	}

	cfg.SSOStartURL = item.Value()

	item, err = profile.RawConfig.GetKey("credential_process")
	if err != nil {
		return nil, err
	}

	err = validateCredentialProcess(item.Value(), profile.Name)
	if err != nil {
		return nil, err
	}

	return &cfg, err
}

// We have to make sure the custom prefix is added to the aws config file.
func IsValidCustomProfile(rawConfig *ini.Section) error {
	requiredCustomCredentials := []string{
		cfg.AppName + "_sso_start_url",
		cfg.AppName + "_sso_region",
		cfg.AppName + "_sso_account_id",
		cfg.AppName + "_sso_role_name",
	}
	for _, value := range requiredCustomCredentials {
		if !rawConfig.HasKey(value) {
			return fmt.Errorf("invalid custom aws config. '%s' field must be provided", value)
		}
	}
	return nil
}

// check if the config section has any keys with custom prefix
func hasCustomSSOPrefix(rawConfig *ini.Section) bool {
	for _, v := range rawConfig.KeyStrings() {
		if strings.HasPrefix(v, cfg.AppName+"_sso_") {
			return true
		}
	}
	return false
}

// validateCredentialProcess checks whether the custom prefixed AWS profiles
// are correctly using the credential-process override or not.
// also check whether the provided flag to 'credential-process --profile pname'
// matches the AWS config profile name. If it doesn't then return an err
// as the user will certainly run into unexpected behaviour.
func validateCredentialProcess(arg string, awsProfileName string) error {
	regex := regexp.MustCompile(`^(\s+)?` + cfg.AppName + `\s+credential-process.*--profile\s+(?P<PName>([^\s]+))`)

	if regex.MatchString(arg) {
		matches := regex.FindStringSubmatch(arg)
		pNameIndex := regex.SubexpIndex("PName")

		profileName := matches[pNameIndex]

		if profileName == "" {
			return fmt.Errorf("profile name not provided. Try adding profile name like '" + cfg.AppName + " credential-process --profile <profile-name>'")
		}

		// if matches then do nth.
		if profileName == awsProfileName {
			return nil
		}

		return fmt.Errorf("unmatched profile names. The profile name '%s' provided to '"+cfg.AppName+" credential-process' does not match AWS profile name '%s'", profileName, awsProfileName)
	}

	return fmt.Errorf("unable to parse 'credential_process'. Looks like your credential_process isn't configured correctly. \n You need to add '" + cfg.AppName + " credential-process --profile <profile-name>'")
}
