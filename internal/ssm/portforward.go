package ssm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/AndreZiviani/aws-fuzzy/internal/ssm_plugin"
	"github.com/AndreZiviani/aws-fuzzy/internal/tracing"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsssm "github.com/aws/aws-sdk-go-v2/service/ssm"
	opentracing "github.com/opentracing/opentracing-go"
)

func NewPortForward(profile, region, ports, instance string, nonInteractive bool) *PortForward {
	pf := PortForward{
		Profile:        profile,
		Region:         region,
		Ports:          ports,
		Instance:       instance,
		NonInteractive: nonInteractive,
	}

	return &pf
}

func (p *PortForward) DoPortForward(ctx context.Context, cfg aws.Config, id, local, host, remote string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ssmportforward")
	defer span.Finish()

	docName := docPortForwardRemoteHost
	input := &awsssm.StartSessionInput{
		DocumentName: &docName,
		Parameters: map[string][]string{
			"portNumber":      []string{remote},
			"localPortNumber": []string{local},
			"host":            []string{host},
		},
		Target: &id,
	}
	inputJson, _ := json.Marshal(input)

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
	_ = ssm_plugin.RunPlugin(
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

func (p *PortForward) Execute(ctx context.Context) error {
	closer, err := tracing.InitTracing()
	if err != nil {
		fmt.Printf("failed to initialize tracing, %s\n", err)
	}
	defer func() { _ = closer.Close() }()

	tracer := opentracing.GlobalTracer()
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, tracer, "ssm")

	ports := strings.Split(p.Ports, ":")
	if len(ports) != 3 {
		return fmt.Errorf("invalid --ports %q, expected '<local port>:<remote host>:<remote port>'", p.Ports)
	}

	if err := validateTarget(p.Instance, p.NonInteractive); err != nil {
		return err
	}

	cfg, err := newConfig(ctx, p.Profile, p.Region)
	if err != nil {
		return err
	}

	instance, err := resolveInstance(ctx, cfg, p.Instance, p.NonInteractive)
	if err != nil {
		return err
	}

	span.Finish()

	return p.DoPortForward(ctx, cfg, aws.ToString(instance.InstanceId), ports[0], ports[1], ports[2])
}
