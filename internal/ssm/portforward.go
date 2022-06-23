package ssm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/AndreZiviani/aws-fuzzy/internal/ssm_plugin"
	"github.com/AndreZiviani/aws-fuzzy/internal/sso"
	"github.com/AndreZiviani/aws-fuzzy/internal/tracing"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	awsssm "github.com/aws/aws-sdk-go-v2/service/ssm"
	opentracing "github.com/opentracing/opentracing-go"
)

func (p *PortForward) DoPortForward(ctx context.Context, id, local, remote string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ssmportforward")
	defer span.Finish()

	login := sso.Login{Profile: p.Profile}
	creds, err := login.GetCredentials(ctx)
	if err != nil {
		return err
	}

	cfg, err := sso.NewAwsConfig(ctx, creds, config.WithRegion(p.Region))
	if err != nil {
		return err
	}

	docName := "AWS-StartPortForwardingSession" // https://us-east-1.console.aws.amazon.com/systems-manager/documents/AWS-StartPortForwardingSession/description?region=us-east-1
	input := &awsssm.StartSessionInput{
		DocumentName: &docName,
		Parameters: map[string][]string{
			"portNumber":      []string{remote},
			"localPortNumber": []string{local},
		},
		Target: &id,
	}
	inputJson, err := json.Marshal(input)

	ssmclient := awsssm.NewFromConfig(cfg)

	session, err := ssmclient.StartSession(ctx, input)
	sessionJson, _ := json.Marshal(session)

	if err != nil {
		return err
	}

	/*
		// we cant incluse the plugin directly here because it is mostly unmaintained
		// maybe this issue will be fixed and we could try to embed the plugin here
		// https://github.com/aws/session-manager-plugin/issues/1

		input = []string{
			"ignored",
			string(sessionJson),
			p.Region,
			"StartSession",
			p.Profile,
			fmt.Sprintf("{\"Target\":\"%s\"}",
			id,
			fmt.Sprintf("https://ssm.%s.amazonaws.com", p.Region,
		}

		session.ValidadeInputAndStartSession(input, os.Stdout)
	*/

	// for now we have to use the embeded the binary
	ssm_plugin.RunPlugin(
		string(sessionJson),
		p.Region,
		"StartSession",
		p.Profile,
		string(inputJson),
	)

	_, err = ssmclient.TerminateSession(ctx, &awsssm.TerminateSessionInput{
		SessionId: session.SessionId,
	})

	return err
}

func (p *PortForward) Execute(args []string) error {

	ctx := context.Background()

	closer, err := tracing.InitTracing()
	if err != nil {
		fmt.Printf("failed to initialize tracing, %s\n", err)
	}
	defer closer.Close()

	tracer := opentracing.GlobalTracer()
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, tracer, "ssm")

	login := sso.Login{Profile: p.Profile}

	creds, err := login.GetCredentials(ctx)
	if err != nil {
		return err
	}

	cfg, err := sso.NewAwsConfig(ctx, creds, config.WithRegion(p.Region))
	if err != nil {
		return err
	}

	instances, err := GetInstances(ctx, cfg)
	if err != nil {
		return err
	}

	span.Finish()

	instance, err := tui(instances)
	if err != nil {
		return err
	}

	ports := strings.Split(p.Ports, ":")

	return p.DoPortForward(ctx, aws.ToString(instance.InstanceId), ports[0], ports[1])
}
