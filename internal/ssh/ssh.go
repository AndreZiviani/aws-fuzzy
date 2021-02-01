package ssh

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/AndreZiviani/aws-fuzzy/internal/cache"
	"github.com/AndreZiviani/aws-fuzzy/internal/sso"
	"github.com/AndreZiviani/aws-fuzzy/internal/tracing"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"os"
	"os/exec"
	"time"
)

func GetTag(tags []ec2types.Tag, key string) string {
	// Get tag Name
	for _, tag := range tags {
		if aws.ToString(tag.Key) == key {
			return aws.ToString(tag.Value)
		}
	}
	return "<missing Name>"
}

func DoSsh(user, key, ip string) {
	cmd := exec.Command("ssh", "-l", user, "-i", key, ip)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func GetInstances(ctx context.Context, profile string) (*ec2.DescribeInstancesOutput, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "sshgetinstances")
	defer span.Finish()

	c, _ := cache.New("ssh")
	j, err := c.Fetch(profile)

	instances := &ec2.DescribeInstancesOutput{}
	if err == nil {
		// We have valid cached credentials
		err = json.Unmarshal([]byte(j), &instances)

		//return instances, nil
	}

	creds, err := sso.GetCredentials(ctx, profile)
	if err != nil {
		return nil, err
	}

	cfg, err := sso.NewAwsConfig(ctx, creds)
	if err != nil {
		return nil, err
	}

	spanDescribeInstances, ctx := opentracing.StartSpanFromContext(ctx, "ec2describe")
	ec2client := ec2.NewFromConfig(cfg)

	instances, err = ec2client.DescribeInstances(ctx,
		&ec2.DescribeInstancesInput{
			Filters: []ec2types.Filter{
				{
					Name:   aws.String("instance-state-name"),
					Values: []string{"running"},
				},
			},
			MaxResults: 1000,
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
	tmp, _ := json.Marshal(instances)
	c.Save(profile, string(tmp), time.Duration(10)*time.Minute)
	spanDescribeInstances.Finish()

	return instances, nil

}

func (p *SshCommand) Execute(args []string) error {

	ctx := context.Background()

	closer, err := tracing.InitTracing()
	if err != nil {
		fmt.Printf("failed to initialize tracing, %s\n", err)
	}
	defer closer.Close()

	tracer := opentracing.GlobalTracer()
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, tracer, "ssh")

	// Expand $USER env variable
	if p.User == "$USER" {
		p.User = os.Getenv("USER")
	}

	// Expand ~ if present
	if p.Key[0] == '~' {
		p.Key = fmt.Sprintf("%s/%s", os.Getenv("HOME"), p.Key[2:])
	}

	instances, err := GetInstances(ctx, p.Profile)

	span.Finish()

	instance, err := FuzzyFind(instances)
	if err != nil {
		fmt.Printf("failed to select instance, %s\n", err)
		return err
	}

	return nil
	DoSsh(p.User, p.Key, aws.ToString(instance.PrivateIpAddress))
	return nil

}
