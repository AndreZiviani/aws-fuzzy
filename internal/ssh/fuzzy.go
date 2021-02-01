package ssh

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	fuzzyfinder "github.com/ktr0731/go-fuzzyfinder"
)

func FuzzyFind(list *ec2.DescribeInstancesOutput) (*ec2types.Instance, error) {
	instances := []ec2types.Instance{}

	for _, r := range list.Reservations {
		instances = append(instances, r.Instances...)
	}

	idx, err := fuzzyfinder.Find(
		instances,
		func(i int) string {
			main := fmt.Sprintf("%s (%s)", GetTag(instances[i].Tags, "Name"), aws.ToString(instances[i].PrivateIpAddress))
			return main
		},
		fuzzyfinder.WithPreviewWindow(
			func(i, width, _ int) string {
				if i == -1 {
					return "no results"
				}

				iam := "<no role attached>"
				if instances[i].IamInstanceProfile != nil {
					iam = aws.ToString(instances[i].IamInstanceProfile.Arn)
				}

				return fmt.Sprintf("Name: %s\nInstanceId: %s\nPrivateIP: %s\nInstanceType: %s\nIAM: %s\nImageId: %s\n",
					GetTag(instances[i].Tags, "Name"), aws.ToString(instances[i].InstanceId),
					aws.ToString(instances[i].PrivateIpAddress), instances[i].InstanceType,
					iam, aws.ToString(instances[i].ImageId),
				)
			},
		),
	)

	return &instances[idx], err
}
