package peering

import (
	"context"

	"github.com/AndreZiviani/aws-fuzzy/internal/config"
	opentracing "github.com/opentracing/opentracing-go"
	//"github.com/opentracing/opentracing-go/log"
)

func Peering(ctx context.Context, profile string, account string) ([]string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "peering")
	defer span.Finish()

	query := config.Config{
		Profile: profile,
		Account: account,
		Pager:   false,
		Service: "EC2",
		Select: "configuration.requesterVpcInfo.ownerId" +
			", configuration.requesterVpcInfo.vpcId" +
			", configuration.accepterVpcInfo.vpcId" +
			", configuration.accepterVpcInfo.ownerId" +
			", configuration.vpcPeeringConnectionId" +
			", tags.key, tags.value",
		Filter: "",
		Limit:  0,
	}

	result, err := config.QueryConfig(ctx, &query, "VPCPeeringConnection%")

	return result, err
}
