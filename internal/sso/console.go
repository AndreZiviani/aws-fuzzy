package sso

import (
	"context"
	"fmt"

	"github.com/AndreZiviani/aws-fuzzy/internal/awsprofile"
	"github.com/AndreZiviani/aws-fuzzy/internal/tracing"
	gconsole "github.com/common-fate/granted/pkg/console"
	opentracing "github.com/opentracing/opentracing-go"
)

func NewConsole(profile, region, service string, url, verbose, noCache bool) *Console {
	console := Console{
		Profile: profile,
		Region:  region,
		Service: service,
		Url:     url,
		Verbose: verbose,
		NoCache: noCache,
	}

	return &console
}

func (p *Console) Execute(ctx context.Context) error {
	closer, err := tracing.InitTracing()
	if err != nil {
		fmt.Printf("failed to initialize tracing, %s\n", err)
	}
	defer closer.Close()

	tracer := opentracing.GlobalTracer()
	spanSso, ctx := opentracing.StartSpanFromContextWithTracer(ctx, tracer, "ssoconsolecmd")
	defer spanSso.Finish()

	return p.OpenBrowser(ctx)
}

func (p *Console) OpenBrowser(ctx context.Context) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ssorolecreds")
	defer span.Finish()

	login := Login{Profile: p.Profile, NoCache: p.NoCache, Url: p.Url}
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

	return awsprofile.LaunchBrowser(session, p.Profile, "console", p.Url)
}
