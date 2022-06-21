package ssm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/AndreZiviani/aws-fuzzy/internal/ssm_plugin"
	"github.com/AndreZiviani/aws-fuzzy/internal/sso"
	"github.com/AndreZiviani/aws-fuzzy/internal/tracing"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	awsssm "github.com/aws/aws-sdk-go-v2/service/ssm"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

func (p *Session) DoSsm(ctx context.Context, id string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ssmsession")
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

	input := &awsssm.StartSessionInput{Target: &id}
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

func (p *Session) GetInstances(ctx context.Context) (*ec2.DescribeInstancesOutput, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "sshgetinstances")
	defer span.Finish()

	login := sso.Login{Profile: p.Profile}

	creds, err := login.GetCredentials(ctx)
	if err != nil {
		return nil, err
	}

	cfg, err := sso.NewAwsConfig(ctx, creds, config.WithRegion(p.Region))
	if err != nil {
		return nil, err
	}

	spanDescribeInstances, ctx := opentracing.StartSpanFromContext(ctx, "ssmgetinstances")

	ssmclient := awsssm.NewFromConfig(cfg)
	ssmPag := awsssm.NewDescribeInstanceInformationPaginator(
		ssmclient,
		&awsssm.DescribeInstanceInformationInput{MaxResults: 50},
	)
	ssmInstances := make([]*awsssm.DescribeInstanceInformationOutput, 0)
	for ssmPag.HasMorePages() {
		tmpInstances, err := ssmPag.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		ssmInstances = append(ssmInstances, tmpInstances)
	}

	instanceIDs := SSMGetInstanceID(ssmInstances)

	ec2client := ec2.NewFromConfig(cfg)

	// TODO: paginate if instance list is big
	instances, err := ec2client.DescribeInstances(ctx,
		&ec2.DescribeInstancesInput{
			Filters: []ec2types.Filter{
				{
					Name:   aws.String("instance-state-name"),
					Values: []string{"running"},
				},
			},
			InstanceIds: instanceIDs,
		},
	)
	if err != nil {
		fmt.Printf("failed to describe instances, %s\n", err)
		return nil, err
	}

	spanDescribeInstances.SetTag("service", "ssm")
	spanDescribeInstances.LogFields(
		log.String("event", "describe instances"),
	)
	spanDescribeInstances.Finish()

	return instances, nil

}

func (p *Session) Execute(args []string) error {

	ctx := context.Background()

	closer, err := tracing.InitTracing()
	if err != nil {
		fmt.Printf("failed to initialize tracing, %s\n", err)
	}
	defer closer.Close()

	tracer := opentracing.GlobalTracer()
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, tracer, "ssm")

	instances, err := p.GetInstances(ctx)
	if err != nil {
		return err
	}

	span.Finish()

	instance, err := tui(instances)
	if err != nil {
		return err
	}

	return p.DoSsm(ctx, aws.ToString(instance.InstanceId))
}

func SSMGetInstanceID(ssmOutputs []*awsssm.DescribeInstanceInformationOutput) []string {
	ec2List := make([]string, 0)

	for _, list := range ssmOutputs {
		for _, instance := range list.InstanceInformationList {
			ec2List = append(ec2List, aws.ToString(instance.InstanceId))
		}
	}

	return ec2List
}
