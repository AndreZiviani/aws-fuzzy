package ssm

import (
	"context"
	"fmt"

	"github.com/AndreZiviani/aws-fuzzy/internal/sso"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	awsssm "github.com/aws/aws-sdk-go-v2/service/ssm"
	awsssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

const (
	docPortForwardRemoteHost = "AWS-StartPortForwardingSessionToRemoteHost"
	docInteractiveCommand    = "AWS-StartInteractiveCommand"
	docRunShellScript        = "AWS-RunShellScript"
)

// newConfig builds an AWS config for the given profile and region, logging in
// via SSO if needed.
func newConfig(ctx context.Context, profile, region string) (aws.Config, error) {
	login := sso.Login{Profile: profile}

	creds, err := login.GetCredentials(ctx)
	if err != nil {
		return aws.Config{}, err
	}

	return sso.NewAwsConfig(ctx, creds, config.WithRegion(region))
}

// GetInstances lists running EC2 instances that are registered with SSM and
// reachable. When instanceIDs is non-empty the SSM lookup is narrowed to those
// instances instead of enumerating every managed instance in the account.
func GetInstances(ctx context.Context, cfg aws.Config, instanceIDs []string) (*ec2.DescribeInstancesOutput, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ssmgetinstances")
	defer span.Finish()

	filters := []awsssmtypes.InstanceInformationStringFilter{
		{
			Key:    aws.String("PingStatus"),
			Values: []string{"Online"},
		},
		{
			Key:    aws.String("AssociationStatus"),
			Values: []string{"Success"},
		},
	}

	if len(instanceIDs) > 0 {
		filters = append(filters, awsssmtypes.InstanceInformationStringFilter{
			Key:    aws.String("InstanceIds"),
			Values: instanceIDs,
		})
	}

	ssmclient := awsssm.NewFromConfig(cfg)
	ssmPag := awsssm.NewDescribeInstanceInformationPaginator(
		ssmclient,
		&awsssm.DescribeInstanceInformationInput{
			MaxResults: aws.Int32(50),
			Filters:    filters,
		},
	)
	ssmInstances := make([]*awsssm.DescribeInstanceInformationOutput, 0)
	for ssmPag.HasMorePages() {
		tmpInstances, err := ssmPag.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		ssmInstances = append(ssmInstances, tmpInstances)
	}

	managedIDs := SSMGetInstanceID(ssmInstances)
	if len(managedIDs) == 0 {
		return &ec2.DescribeInstancesOutput{}, nil
	}

	spanDescribeInstances, ctx := opentracing.StartSpanFromContext(ctx, "ec2getinstances")
	defer spanDescribeInstances.Finish()

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
			InstanceIds: managedIDs,
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

	return instances, nil
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
