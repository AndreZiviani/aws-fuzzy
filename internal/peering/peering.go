package peering

import (
	"context"

	"github.com/AndreZiviani/aws-fuzzy/internal/config"
	opentracing "github.com/opentracing/opentracing-go"
	//"github.com/opentracing/opentracing-go/log"
)

func Peering(ctx context.Context, profile string, account string, region string) ([]string, []string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "peering")
	defer span.Finish()

	peering := config.Config{
		Profile: profile,
		Account: account,
		Region:  region,
		Pager:   false,
		Service: "EC2",
		Type:    "VPCPeeringConnection",
		Select: "configuration.requesterVpcInfo.ownerId" +
			", configuration.requesterVpcInfo.vpcId" +
			", configuration.requesterVpcInfo.region" +
			", configuration.accepterVpcInfo.vpcId" +
			", configuration.accepterVpcInfo.ownerId" +
			", configuration.accepterVpcInfo.region" +
			", configuration.vpcPeeringConnectionId" +
			", configuration.status" +
			", tags",
		Filter: "",
		Limit:  0,
	}

	vpc := config.Config{
		Profile: profile,
		Account: account,
		Region:  region,
		Pager:   false,
		Service: "EC2",
		Type:    "VPC",
		Select: "resourceId" +
			", configuration.ownerId" +
			", tags",
		Filter: "",
		Limit:  0,
	}

	peeringResult, err := peering.QueryConfig(ctx)
	if err != nil {
		return nil, nil, err
	}

	vpcResult, err := vpc.QueryConfig(ctx)
	if err != nil {
		return nil, nil, err
	}

	return peeringResult, vpcResult, nil
}
