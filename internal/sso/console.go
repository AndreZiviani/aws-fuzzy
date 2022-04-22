package sso

import (
	"context"
	"fmt"

	"github.com/AndreZiviani/aws-fuzzy/internal/tracing"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/common-fate/granted/pkg/browsers"
	"github.com/common-fate/granted/pkg/cfaws"
	"github.com/common-fate/granted/pkg/config"
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

	creds, err := login.GetCredentials(ctx)
	if err != nil {
		return err
	}

	p.OpenBrowser(ctx, creds)

	return err
}

func (p *Console) OpenBrowser(ctx context.Context, credentials *aws.Credentials) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ssorolecreds")
	defer span.Finish()

	awsProfiles, _ := cfaws.GetProfilesFromDefaultSharedConfig(ctx)

	profile := awsProfiles[p.Profile]

	browserOpts := browsers.BrowserOpts{Profile: profile.Name}
	url, err := browsers.MakeUrl(browsers.SessionFromCredentials(*credentials), browserOpts, "", profile.AWSConfig.Region)
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

	return browsers.LaunchConsoleSession(browsers.SessionFromCredentials(*credentials), browserOpts, "", profile.AWSConfig.Region)
}
