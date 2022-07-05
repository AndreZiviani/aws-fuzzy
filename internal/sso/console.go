package sso

import (
	"context"
	"fmt"

	"github.com/AndreZiviani/aws-fuzzy/internal/tracing"
	"github.com/common-fate/granted/pkg/browsers"
	"github.com/common-fate/granted/pkg/config"
	"github.com/common-fate/granted/pkg/debug"
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

	if p.Verbose {
		// enable granted debug
		debug.CliVerbosity = debug.VerbosityDebug
	}

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

	session := browsers.SessionFromCredentials(*credentials)

	region := p.Region
	if len(profile.AWSConfig.Region) > 0 {
		region = profile.AWSConfig.Region
	}

	browserOpts := browsers.BrowserOpts{Profile: profile.Name}
	url, err := browsers.MakeUrl(session, browserOpts, "", region)
	if err != nil {
		return err
	}

	cfg, err := config.Load()

	if p.Url {
		if cfg.DefaultBrowser == browsers.FirefoxKey || cfg.DefaultBrowser == browsers.FirefoxStdoutKey {
			url = browsers.MakeFirefoxContainerURL(url, browserOpts)
			if err != nil {
				return err
			}
		}

		fmt.Println(url)
		return nil
	}

	return browsers.LaunchConsoleSession(session, browserOpts, "", region)
}
