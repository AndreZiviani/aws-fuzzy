package ssm

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	awsssm "github.com/aws/aws-sdk-go-v2/service/ssm"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

func GetInstances(ctx context.Context, cfg aws.Config) (*ec2.DescribeInstancesOutput, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ssmgetinstances")
	defer span.Finish()

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
