package awsprofile

import (
	"context"
	"fmt"
	neturl "net/url"
	"os"
	"path"

	"github.com/AndreZiviani/aws-fuzzy/internal/afconfig"
	gassume "github.com/common-fate/granted/pkg/assume"
	gbrowser "github.com/common-fate/granted/pkg/browser"
	"github.com/common-fate/granted/pkg/forkprocess"
	glauncher "github.com/common-fate/granted/pkg/launcher"
)

func LaunchBrowser(url string, profile string, flow string, printOnly bool) error {
	cfg, err := afconfig.NewLoadedConfig()
	if err != nil {
		return err
	}

	browserPath := cfg.CustomBrowserPath
	if browserPath == "" && cfg.DefaultBrowser != gbrowser.StdoutKey {
		return fmt.Errorf("default browser not configured. run `aws-fuzzy sso browser` to configure")
	}

	configDir, _ := cfg.ConfigFolder()
	if err != nil {
		return err
	}

	profiles, err := LoadProfiles()
	if err != nil {
		return err
	}

	p, err := profiles.LoadInitialisedProfile(context.TODO(), profile)
	if err != nil {
		return err
	}

	var containerName string
	if flow == "sso" {
		if p.AWSConfig.SSOSession != nil {
			containerName = p.AWSConfig.SSOSession.Name
		} else {
			containerName = p.AWSConfig.SSOStartURL
		}
	} else {
		containerName = p.Name
	}

	var l gassume.Launcher
	finalUrl := url

	switch cfg.DefaultBrowser {
	case gbrowser.ChromeKey:
		l = glauncher.ChromeProfile{
			ExecutablePath: browserPath,
			UserDataPath:   path.Join(configDir, "chromium-profiles", "1"), // held over for backwards compatibility, "1" indicates Chrome profiles
		}
	case gbrowser.BraveKey:
		l = glauncher.ChromeProfile{
			ExecutablePath: browserPath,
			UserDataPath:   path.Join(configDir, "chromium-profiles", "2"), // held over for backwards compatibility, "2" indicates Brave profiles
		}
	case gbrowser.EdgeKey:
		l = glauncher.ChromeProfile{
			ExecutablePath: browserPath,
			UserDataPath:   path.Join(configDir, "chromium-profiles", "3"), // held over for backwards compatibility, "3" indicates Edge profiles
		}
	case gbrowser.ChromiumKey:
		l = glauncher.ChromeProfile{
			ExecutablePath: browserPath,
			UserDataPath:   path.Join(configDir, "chromium-profiles", "4"), // held over for backwards compatibility, "4" indicates Chromium profiles
		}
	case gbrowser.FirefoxKey:
		l = glauncher.Firefox{
			ExecutablePath: browserPath,
		}

		color := ""
		icon := ""
		if flow == "sso" && p.AWSConfig.SSOSession != nil {
			s := profiles.sessions[p.AWSConfig.SSOSession.Name]
			item, err := s.RawConfig.GetKey(cfg.AppNameConfig + "_firefox_color")
			if err == nil {
				// https://github.com/onebytegone/granted-containers/blob/main/src/opener/parser.ts#L14
				// allowed values:
				// "blue", "turquoise", "green", "yellow", "orange", "red", "pink", "purple"
				color = item.Value()
			}

			item, err = s.RawConfig.GetKey(cfg.AppNameConfig + "_firefox_icon")
			if err == nil {
				// https://github.com/onebytegone/granted-containers/blob/main/src/opener/parser.ts#L14
				// allowed values:
				// "fingerprint", "briefcase", "dollar", "cart", "circle", "gift", "vacation", "food", "fruit", "pet", "tree", "chill"
				icon = item.Value()
			}
		} else {
			item, err := p.RawConfig.GetKey(cfg.AppNameConfig + "_firefox_color")
			if err == nil {
				color = item.Value()
			}

			item, err = p.RawConfig.GetKey(cfg.AppNameConfig + "_firefox_icon")
			if err == nil {
				icon = item.Value()
			}
		}

		finalUrl = fmt.Sprintf("ext+granted-containers:name=%s&url=%s&color=%s&icon=%s", containerName, neturl.QueryEscape(url), color, icon)
	case gbrowser.StdoutKey:
		fmt.Println(finalUrl)
		return nil
	default:
		l = glauncher.Open{}
	}

	if printOnly {
		fmt.Println(finalUrl)
	} else {
		// now build the actual command to run - e.g. 'firefox --new-tab <URL>'
		args := l.LaunchCommand(finalUrl, containerName)
		cmd, err := forkprocess.New(args...)
		if err != nil {
			return err
		}
		err = cmd.Start()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Granted was unable to open a browser session automatically: %s", err.Error())
			fmt.Fprintf(os.Stderr, "\nOpen session manually using the following url:\n")
			return err
		}
	}
	return nil
}
