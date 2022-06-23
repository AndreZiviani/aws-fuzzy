package sso

import (
	"context"
	"fmt"

	"github.com/AndreZiviani/aws-fuzzy/internal/tracing"
	"github.com/common-fate/granted/pkg/browsers"
	"github.com/common-fate/granted/pkg/debug"
	opentracing "github.com/opentracing/opentracing-go"
)

func (p *Browser) Execute(args []string) error {

	ctx := context.Background()

	closer, err := tracing.InitTracing()
	if err != nil {
		fmt.Printf("failed to initialize tracing, %s\n", err)
	}
	defer closer.Close()

	tracer := opentracing.GlobalTracer()
	spanSso, ctx := opentracing.StartSpanFromContextWithTracer(ctx, tracer, "ssobrowsercmd")
	defer spanSso.Finish()

	if p.Verbose {
		// enable granted debug
		debug.CliVerbosity = debug.VerbosityDebug
	}

	browser := p.Browser

	if browser == "" {
		browser, err = browsers.HandleManualBrowserSelection()
		if err != nil {
			return err
		}
	}

	return browsers.ConfigureBrowserSelection(browser, "")

}
