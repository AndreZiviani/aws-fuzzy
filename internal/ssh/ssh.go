package ssh

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/AndreZiviani/aws-fuzzy/internal/sso"
	"github.com/AndreZiviani/aws-fuzzy/internal/tracing"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

func New(profile, user, key string) *Ssh {
	keyPath := key
	// Expand ~ if present
	if key[0] == '~' {
		homeDir, _ := os.UserHomeDir()
		keyPath = fmt.Sprintf("%s/%s", homeDir, key[2:])
	}

	ssh := Ssh{
		Profile: profile,
		User:    user,
		Key:     keyPath,
	}

	return &ssh
}

func (p *Ssh) DoSsh(ip string) {
	cmd := exec.Command("ssh", "-l", p.User, "-i", p.Key, ip)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}

func (p *Ssh) GetInstances(ctx context.Context) (*ec2.DescribeInstancesOutput, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "sshgetinstances")
	defer span.Finish()

	login := sso.Login{Profile: p.Profile}

	creds, err := login.GetCredentials(ctx)
	if err != nil {
		return nil, err
	}

	cfg, err := sso.NewAwsConfig(ctx, creds)
	if err != nil {
		return nil, err
	}

	spanDescribeInstances, ctx := opentracing.StartSpanFromContext(ctx, "ec2describe")
	ec2client := ec2.NewFromConfig(cfg)

	instances, err := ec2client.DescribeInstances(ctx,
		&ec2.DescribeInstancesInput{
			Filters: []ec2types.Filter{
				{
					Name:   aws.String("instance-state-name"),
					Values: []string{"running"},
				},
			},
			MaxResults: aws.Int32(1000),
		},
	)
	if err != nil {
		fmt.Printf("failed to describe instances, %s\n", err)
		return nil, err
	}

	spanDescribeInstances.SetTag("service", "ec2")
	spanDescribeInstances.LogFields(
		log.String("event", "describe instances"),
	)
	spanDescribeInstances.Finish()

	return instances, nil

}

func (p *Ssh) Execute(ctx context.Context) error {

	closer, err := tracing.InitTracing()
	if err != nil {
		return fmt.Errorf("failed to initialize tracing, %s", err)
	}
	defer func() { _ = closer.Close() }()

	tracer := opentracing.GlobalTracer()
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, tracer, "ssh")

	instances, err := p.GetInstances(ctx)
	if err != nil {
		return err
	}

	span.Finish()

	instance, err := tui(instances)
	if err != nil {
		return err
	}

	p.DoSsh(aws.ToString(instance.PrivateIpAddress))
	return nil
}
