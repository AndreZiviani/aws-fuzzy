package afconfig

import (
	"fmt"
	neturl "net/url"
	"os"
	"path"

	gassume "github.com/common-fate/granted/pkg/assume"
	gbrowser "github.com/common-fate/granted/pkg/browser"
	"github.com/common-fate/granted/pkg/forkprocess"
	glauncher "github.com/common-fate/granted/pkg/launcher"
)

func LaunchBrowser(url string, profile string, printOnly bool) error {
	cfg, err := NewLoadedConfig()
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
		finalUrl = fmt.Sprintf("ext+granted-containers:name=%s&url=%s", profile, neturl.QueryEscape(url))
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
		args := l.LaunchCommand(finalUrl, profile)
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
