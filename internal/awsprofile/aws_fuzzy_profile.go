package awsprofile

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/AndreZiviani/aws-fuzzy/internal/afconfig"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/common-fate/clio"
	"gopkg.in/ini.v1"
)

func (p *Profiles) ParseCustomSSOProfile(ctx context.Context, profile *Profile) (*config.SharedConfig, error) {
	afcfg := afconfig.NewDefaultConfig()
	err := IsValidCustomProfile(profile.RawConfig)
	if err != nil {
		return nil, err
	}
	cfg, err := config.LoadSharedConfigProfile(ctx, profile.Name, func(lsco *config.LoadSharedConfigOptions) { lsco.ConfigFiles = []string{profile.File} })
	if err != nil {
		return nil, err
	}
	item, err := profile.RawConfig.GetKey(afcfg.AppNameConfig + "_sso_account_id")
	if err != nil {
		return nil, err
	}
	cfg.SSOAccountID = item.Value()
	item, err = profile.RawConfig.GetKey(afcfg.AppNameConfig + "_sso_role_name")
	if err != nil {
		return nil, err
	}
	cfg.SSORoleName = item.Value()
	item, err = profile.RawConfig.GetKey(afcfg.AppNameConfig + "_sso_session")
	if err == nil {
		// New profile with SSO Session
		s, ok := p.sessions[item.Value()]
		if ok == false {
			return nil, err
		}
		err = s.init(ctx)
		if err != nil {
			return nil, err
		}

		cfg.SSOSession = &s.AWSConfig
	} else {
		// Legacy profile
		item, err = profile.RawConfig.GetKey(afcfg.AppNameConfig + "_sso_region")
		if err != nil {
			return nil, err
		}
		cfg.SSORegion = item.Value()
		item, err = profile.RawConfig.GetKey(afcfg.AppNameConfig + "_sso_start_url")
		if err != nil {
			return nil, err
		}

		// sanity check to verify if the provided value is a valid url
		_, err = url.ParseRequestURI(item.Value())
		if err != nil {
			clio.Debug(err)
			return nil, fmt.Errorf("invalid value '%s' provided for '%s_sso_start_url'", item.Value(), afcfg.AppNameConfig)
		}

		cfg.SSOStartURL = item.Value()

	}

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
	afcfg := afconfig.NewDefaultConfig()
	var requiredCustomCredentials []string
	if rawConfig.HasKey(afcfg.AppNameConfig + "_sso_session") {
		requiredCustomCredentials = []string{
			afcfg.AppNameConfig + "_sso_session",
			afcfg.AppNameConfig + "_sso_account_id",
			afcfg.AppNameConfig + "_sso_role_name",
		}
	} else {
		requiredCustomCredentials = []string{
			afcfg.AppNameConfig + "_sso_start_url",
			afcfg.AppNameConfig + "_sso_region",
			afcfg.AppNameConfig + "_sso_account_id",
			afcfg.AppNameConfig + "_sso_role_name",
		}
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
	afcfg := afconfig.NewDefaultConfig()
	for _, v := range rawConfig.KeyStrings() {
		if strings.HasPrefix(v, afcfg.AppNameConfig+"_sso_") {
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
	afcfg := afconfig.NewDefaultConfig()
	appName := "(" + afcfg.AppName + "|" + strings.Replace(afcfg.AppName, "-", "_", -1) + ")"
	regex := regexp.MustCompile(`^(\s+)?` + appName + `\s+sso\s+credential-process.*--profile\s+(?P<PName>([^\s]+))`)

	if regex.MatchString(arg) {
		matches := regex.FindStringSubmatch(arg)
		pNameIndex := regex.SubexpIndex("PName")

		profileName := matches[pNameIndex]

		if profileName == "" {
			return fmt.Errorf("profile name not provided. Try adding profile name like '" + afcfg.AppNameConfig + " credential-process --profile <profile-name>'")
		}

		// if matches then do nth.
		if profileName == awsProfileName {
			return nil
		}

		return fmt.Errorf("unmatched profile names. The profile name '%s' provided to '"+afcfg.AppNameConfig+" credential-process' does not match AWS profile name '%s'", profileName, awsProfileName)
	}

	return fmt.Errorf("unable to parse 'credential_process'. Looks like your credential_process isn't configured correctly. \n You need to add '" + afcfg.AppNameConfig + " credential-process --profile <profile-name>'")
}
