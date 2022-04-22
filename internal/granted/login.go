package granted

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/AndreZiviani/aws-fuzzy/internal/sso"
	"github.com/AndreZiviani/aws-fuzzy/internal/tracing"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/common-fate/granted/pkg/cfaws"
	"github.com/common-fate/granted/pkg/credstore"
	opentracing "github.com/opentracing/opentracing-go"
)

func (p *LoginCommand) Execute(args []string) error {

	ctx := context.Background()

	closer, err := tracing.InitTracing()
	if err != nil {
		fmt.Printf("failed to initialize tracing, %s\n", err)
	}
	defer closer.Close()

	tracer := opentracing.GlobalTracer()
	spanSso, ctx := opentracing.StartSpanFromContextWithTracer(ctx, tracer, "ssologincmd")
	defer spanSso.Finish()

	creds, err := p.GetCredentials(ctx)
	if err != nil {
		return err
	}

	sso.PrintCredentials(creds)

	return err
}

func (p *LoginCommand) GetCredentials(ctx context.Context) (*aws.Credentials, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ssorolecreds")
	defer span.Finish()

	awsProfiles, err := cfaws.GetProfilesFromDefaultSharedConfig(ctx)

	var profile *cfaws.CFSharedConfig
	var ok bool

	if profile, ok = awsProfiles[p.Profile]; !ok {
		return nil, errors.New(fmt.Sprintf("profile %s not found!", p.Profile))
	}

	// if profile == nil {...
	// prompt for profile using fzf

	var creds aws.Credentials
	err = credstore.Retrieve(profile.Name, &creds)
	if err != nil {
		if p.Ask {
			reader := bufio.NewReader(os.Stdin)
			fmt.Fprintf(os.Stderr, "Authenticate again? (y/N) ")
			text, _ := reader.ReadString('\n')
			if text[0] != 'y' && text[0] != 'Y' {
				os.Exit(0)
			}
		}

		fmt.Fprintf(os.Stderr, "Could not find cached credentials, refreshing...\n")
		creds, err = profile.AssumeTerminal(ctx, cfaws.ConfigOpts{Duration: time.Hour})
		if err != nil {
			return nil, err
		}
		credstore.Store(profile.Name, creds)
	}

	return &creds, err
}
