package vpc

import (
	"context"
	"strings"

	"github.com/AndreZiviani/aws-fuzzy/internal/common"
	"github.com/AndreZiviani/aws-fuzzy/internal/sso"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	opentracing "github.com/opentracing/opentracing-go"
	//"github.com/opentracing/opentracing-go/log"
)

type DescribeTGRegistrationsOutput struct {
	Name        string
	Region      string
	Attachments []ec2types.TransitGatewayAttachment
	Resource    string
}

type DescribeTransitGatewayRouteTablesOutput struct {
	Name   string
	Routes []ec2types.TransitGatewayRoute
}

func NewEC2Client(ctx context.Context, profile string, region *string) (*ec2.Client, error) {
	spanNewEC2Client, ctx := opentracing.StartSpanFromContext(ctx, "newec2client")
	defer spanNewEC2Client.Finish()

	login := sso.Login{Profile: profile}
	creds, err := login.GetCredentials(ctx)
	if err != nil {
		return nil, err
	}

	cfg, err := sso.NewAwsConfig(ctx, creds, config.WithRegion(*region))
	if err != nil {
		return nil, err
	}

	return ec2.NewFromConfig(cfg), nil
}

func GetEC2Client(ctx context.Context, clients map[string]map[string]*ec2.Client, profile string, region *string) (*ec2.Client, error) {
	// check if we already have a client
	if _, ok := clients[profile]; ok {
		if client, ok := clients[profile][*region]; ok {
			return client, nil
		}
	} else {
		// we have nothing
		clients[profile] = map[string]*ec2.Client{}
	}

	// creating a client on the specified region
	client, err := NewEC2Client(ctx, profile, region)
	if err != nil {
		return nil, err
	}
	clients[profile][*region] = client

	return client, nil
}

func DescribeTransitGateway(ctx context.Context, ec2client *ec2.Client, tg *string) (*ec2types.TransitGateway, error) {
	spanDescribeTGs, ctx := opentracing.StartSpanFromContext(ctx, "describetransitgws")
	defer spanDescribeTGs.Finish()

	tgid := strings.Split(aws.ToString(tg), "/")

	// get all transit gateways
	tginfo, err := ec2client.DescribeTransitGateways(ctx, &ec2.DescribeTransitGatewaysInput{
		Filters: []ec2types.Filter{
			{
				Name:   aws.String("transit-gateway-id"),
				Values: []string{tgid[1]},
			},
		},
	})

	if err != nil {
		return &ec2types.TransitGateway{}, err
	}

	return &tginfo.TransitGateways[0], nil
}

func GetTransitGatewayAttachmentsByTG(ctx context.Context, clients map[string]map[string]*ec2.Client, attachmentARN arn.ARN) ([]ec2types.TransitGatewayAttachment, error) {
	spanDescribeTGAttachments, ctx := opentracing.StartSpanFromContext(ctx, "describetransitgwattachments")
	defer spanDescribeTGAttachments.Finish()

	login := sso.Login{}
	login.LoadProfiles()

	tgwProfile, err := login.GetProfileFromID(attachmentARN.AccountID)
	if err != nil {
		return nil, nil // TODO
	}

	// get a ec2 client instance on this region using the specified profile
	tgwClient, err := GetEC2Client(ctx, clients, tgwProfile.Name, &attachmentARN.Region)
	if err != nil {
		return nil, err
	}

	tgid := strings.Split(attachmentARN.Resource, "/")

	tgattach, err := tgwClient.DescribeTransitGatewayAttachments(ctx,
		&ec2.DescribeTransitGatewayAttachmentsInput{
			Filters: []ec2types.Filter{
				{
					Name:   aws.String("transit-gateway-id"),
					Values: []string{tgid[1]},
				},
			},
		},
	)

	if err != nil {
		return []ec2types.TransitGatewayAttachment{}, err
	}

	detailedAttachments := make([]ec2types.TransitGatewayAttachment, len(tgattach.TransitGatewayAttachments))

	for i, attach := range tgattach.TransitGatewayAttachments {
		detailedAttachments[i] = attach

		profile, err := login.GetProfileFromID(*attach.ResourceOwnerId)
		if err != nil {
			return nil, err // TODO
		}
		client, err := GetEC2Client(ctx, clients, profile.Name, &attachmentARN.Region)

		if attach.ResourceType == ec2types.TransitGatewayAttachmentResourceTypeVpc {
			vpc, err := client.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{
				VpcIds: []string{aws.ToString(attach.ResourceId)},
			})
			if err != nil {
				return nil, err // TODO
			}
			name := common.GetEC2Tag(vpc.Vpcs[0].Tags, "Name", aws.ToString(attach.ResourceId))
			cidr := vpc.Vpcs[0].CidrBlock

			detailedAttachments[i].Tags = common.SetEC2Tag(detailedAttachments[i].Tags, "Name", &name)
			detailedAttachments[i].Tags = common.SetEC2Tag(detailedAttachments[i].Tags, "cidr", cidr)
		}
	}

	return detailedAttachments, nil
}

func GetTransitGatewayAttachmentsByAttachment(ctx context.Context, ec2client *ec2.Client, attachment string) ([]ec2types.TransitGatewayAttachment, error) {
	spanDescribeTGAttachments, ctx := opentracing.StartSpanFromContext(ctx, "describetransitgwattachments")
	defer spanDescribeTGAttachments.Finish()

	tgattach, err := ec2client.DescribeTransitGatewayAttachments(ctx,
		&ec2.DescribeTransitGatewayAttachmentsInput{
			Filters: []ec2types.Filter{
				{
					Name:   aws.String("transit-gateway-attachment-id"),
					Values: []string{attachment},
				},
			},
		},
	)

	if err != nil {
		return []ec2types.TransitGatewayAttachment{}, err
	}

	return tgattach.TransitGatewayAttachments, nil
}

func DescribeTransitGatewayRegistrationsFromARN(ctx context.Context, transitGatewaysARN []*string) ([]*DescribeTGRegistrationsOutput, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "describetransitgateway")
	defer span.Finish()

	clients := make(map[string]map[string]*ec2.Client)

	output := make([]*DescribeTGRegistrationsOutput, len(transitGatewaysARN))

	login := sso.Login{}
	login.LoadProfiles()

	for i, tg := range transitGatewaysARN {

		arn, _ := arn.Parse(aws.ToString(tg))

		profile, err := login.GetProfileFromID(arn.AccountID)
		if err != nil {
			// could not find a profile with this account id so we cant get more information about this TGW
			output[i] = &DescribeTGRegistrationsOutput{
				Name:     arn.Resource,
				Region:   arn.Region,
				Resource: arn.Resource,
			}
			continue
		}

		// get a ec2 client instance on this region using the specified profile
		client, err := GetEC2Client(ctx, clients, profile.Name, &arn.Region)
		if err != nil {
			return nil, err
		}

		attachments, err := GetTransitGatewayAttachmentsByTG(ctx, clients, arn)

		// get transit gateway Name tag
		tginfo, _ := DescribeTransitGateway(ctx, client, &arn.Resource)
		name := common.GetEC2Tag(tginfo.Tags, "Name", strings.Split(arn.Resource, "/")[1])

		output[i] = &DescribeTGRegistrationsOutput{
			Name:        name,
			Region:      arn.Region,
			Attachments: attachments,
			Resource:    arn.Resource,
		}
	}

	return output, nil
}

func DescribeTransitGatewayRouteTables(ctx context.Context, ec2client *ec2.Client) ([]*DescribeTransitGatewayRouteTablesOutput, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "describe transitgateway route tables")
	defer span.Finish()

	tmp, err := ec2client.DescribeTransitGatewayRouteTables(ctx, &ec2.DescribeTransitGatewayRouteTablesInput{})
	if err != nil {
		return nil, err
	}

	routeTables := tmp.TransitGatewayRouteTables

	output := make([]*DescribeTransitGatewayRouteTablesOutput, len(routeTables))

	for i, table := range routeTables {
		tmp, _ := ec2client.SearchTransitGatewayRoutes(ctx,
			&ec2.SearchTransitGatewayRoutesInput{
				Filters: []ec2types.Filter{
					{
						Name:   aws.String("state"),
						Values: []string{"active"},
					},
				},
				TransitGatewayRouteTableId: table.TransitGatewayRouteTableId,
			},
		)

		routes := tmp.Routes

		tableName := common.GetEC2Tag(table.Tags, "Name", aws.ToString(table.TransitGatewayRouteTableId))

		output[i] = &DescribeTransitGatewayRouteTablesOutput{
			Name:   tableName,
			Routes: routes,
		}
	}

	return output, nil
}
