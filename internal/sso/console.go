package sso

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path"

	"github.com/AndreZiviani/aws-fuzzy/internal/afconfig"
	"github.com/AndreZiviani/aws-fuzzy/internal/tracing"
	gassume "github.com/common-fate/granted/pkg/assume"
	gbrowser "github.com/common-fate/granted/pkg/browser"
	gconsole "github.com/common-fate/granted/pkg/console"
	"github.com/common-fate/granted/pkg/forkprocess"
	glauncher "github.com/common-fate/granted/pkg/launcher"
	opentracing "github.com/opentracing/opentracing-go"
)

func (p *Console) Execute(args []string) error {

	ctx := context.Background()

	closer, err := tracing.InitTracing()
	if err != nil {
		fmt.Printf("failed to initialize tracing, %s\n", err)
	}
	defer closer.Close()

	tracer := opentracing.GlobalTracer()
	spanSso, ctx := opentracing.StartSpanFromContextWithTracer(ctx, tracer, "ssoconsolecmd")
	defer spanSso.Finish()

	p.OpenBrowser(ctx)

	return err
}

func (p *Console) OpenBrowser(ctx context.Context) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ssorolecreds")
	defer span.Finish()

	login := Login{Profile: p.Profile}
	login.LoadProfiles()
	profile, err := login.GetProfile(p.Profile)
	if err != nil {
		return err
	}

	credentials, err := login.GetCredentials(ctx)
	if err != nil {
		return err
	}

	region := p.Region
	if len(profile.AWSConfig.Region) > 0 {
		region = profile.AWSConfig.Region
	}

	con := gconsole.AWS{
		Profile: p.Profile,
		Region:  region,
		Service: p.Service,
	}
	session, err := con.URL(*credentials)
	if err != nil {
		return err
	}

	cfg := afconfig.NewDefaultConfig()
	err = cfg.Load()

	if cfg.DefaultBrowser == gbrowser.FirefoxKey {
		session = fmt.Sprintf("ext+granted-containers:name=%s&url=%s", p.Profile, url.QueryEscape(session))
	}

	if p.Url {
		fmt.Println(session)
		return nil
	}

	return p.LaunchConsoleSession(session)
}

func (p *Console) LaunchConsoleSession(con string) error {
	cfg := afconfig.NewDefaultConfig()
	err := cfg.Load()
	if err != nil {
		return err
	}

	browserPath := cfg.CustomBrowserPath
	if browserPath == "" {
		return fmt.Errorf("default browser not configured. run `aws-fuzzy sso browser` to configure")
	}

	configDir, _ := cfg.ConfigFolder()
	if err != nil {
		return err
	}

	var l gassume.Launcher
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
	default:
		l = glauncher.Open{}
	}

	// now build the actual command to run - e.g. 'firefox --new-tab <URL>'
	args := l.LaunchCommand(con, p.Profile)
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
	return nil
}
