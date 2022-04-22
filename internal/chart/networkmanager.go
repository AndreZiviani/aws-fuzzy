package chart

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/AndreZiviani/aws-fuzzy/internal/common"
	"github.com/AndreZiviani/aws-fuzzy/internal/sso"
	"github.com/AndreZiviani/aws-fuzzy/internal/tracing"
	"github.com/AndreZiviani/aws-fuzzy/internal/vpc"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/config"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	opentracing "github.com/opentracing/opentracing-go"

	//"github.com/opentracing/opentracing-go/log"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

func mapRegistrations(tgs []*vpc.DescribeTGRegistrationsOutput) *opts.TreeData {
	transitgatewaynodes := make([]*opts.TreeData, 0)
	regions := make(map[string][]*opts.TreeData)

	login := sso.Login{}

	for _, tg := range tgs {
		tgwchildren := make([]*opts.TreeData, 0)
		vpcnodes := make([]*opts.TreeData, 0)
		vpnnodes := make([]*opts.TreeData, 0)
		dxgwnodes := make([]*opts.TreeData, 0)
		connectnodes := make([]*opts.TreeData, 0)
		peeringnodes := make([]*opts.TreeData, 0)
		tgwpeeringnodes := make([]*opts.TreeData, 0)

		// Create Tree nodes
		for _, attachment := range tg.Attachments {
			profile, err := login.GetProfileFromID(aws.ToString(attachment.ResourceOwnerId))
			account := profile.Name
			if err != nil {
				account = *attachment.ResourceOwnerId
			}

			name := fmt.Sprintf("%s\n%s\n%s",
				common.GetEC2Tag(attachment.Tags, "Name", aws.ToString(attachment.TransitGatewayAttachmentId)),
				account,
				aws.ToString(attachment.ResourceId),
			)

			node := &opts.TreeData{
				Name: name,
			}

			switch attachment.ResourceType {

			case ec2types.TransitGatewayAttachmentResourceTypeVpc:
				vpcnodes = append(vpcnodes, node)

			case ec2types.TransitGatewayAttachmentResourceTypeVpn:
				vpnnodes = append(vpnnodes, node)

			case ec2types.TransitGatewayAttachmentResourceTypeDirectConnectGateway:
				dxgwnodes = append(dxgwnodes, node)

			case ec2types.TransitGatewayAttachmentResourceTypeConnect:
				connectnodes = append(connectnodes, node)

			case ec2types.TransitGatewayAttachmentResourceTypePeering:
				peeringnodes = append(peeringnodes, node)

			case ec2types.TransitGatewayAttachmentResourceTypeTgwPeering:
				tgwpeeringnodes = append(tgwpeeringnodes, node)

			default:
				continue

			}
		}

		if len(vpcnodes) > 0 {
			tgwchildren = append(tgwchildren,
				&opts.TreeData{
					Name:     "vpcs",
					Children: vpcnodes,
				},
			)
		}
		if len(vpnnodes) > 0 {
			tgwchildren = append(tgwchildren,
				&opts.TreeData{
					Name:     "vpns",
					Children: vpnnodes,
				},
			)
		}
		if len(dxgwnodes) > 0 {
			tgwchildren = append(tgwchildren,
				&opts.TreeData{
					Name:     "dxs",
					Children: dxgwnodes,
				},
			)
		}
		if len(connectnodes) > 0 {
			tgwchildren = append(tgwchildren,
				&opts.TreeData{
					Name:     "connections",
					Children: connectnodes,
				},
			)
		}
		if len(peeringnodes) > 0 {
			tgwchildren = append(tgwchildren,
				&opts.TreeData{
					Name:     "peerings",
					Children: peeringnodes,
				},
			)
		}
		if len(tgwpeeringnodes) > 0 {
			tgwchildren = append(tgwchildren,
				&opts.TreeData{
					Name:     "tgwpeerings",
					Children: tgwpeeringnodes,
				},
			)
		}

		tgw := &opts.TreeData{
			Name:     tg.Name,
			Children: tgwchildren,
		}

		if _, ok := regions[tg.Region]; !ok {
			regionnodes := make([]*opts.TreeData, 0)
			regions[tg.Region] = regionnodes
		}
		regions[tg.Region] = append(regions[tg.Region], tgw)

	}

	for k, v := range regions {
		transitgatewaynodes = append(transitgatewaynodes,
			&opts.TreeData{
				Name:     k,
				Children: v,
			},
		)
	}

	return &opts.TreeData{
		Name:     "Transit Gateways",
		Children: transitgatewaynodes,
	}
}

func NetworkManager(ctx context.Context, p *NM) ([]opts.TreeData, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "networkmanager")
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

	nmclient := networkmanager.NewFromConfig(cfg)

	globalnetworks, err := vpc.GetGlobalNetworks(ctx, nmclient)
	if err != nil {
		return nil, err
	}

	globalnetworknodes := make([]*opts.TreeData, 0)
	for _, network := range globalnetworks {
		arn, _ := arn.Parse(*network.GlobalNetworkArn)
		name := common.GetNMTag(network.Tags, "Name", strings.Split(arn.Resource, "/")[1])

		gnetworkchildren := make([]*opts.TreeData, 0)

		networkinfo, err := vpc.DescribeGlobalNetwork(ctx, nmclient, network)
		if err != nil {
			return nil, err
		}

		/* TODO:
		DescribeConnectionsFromARN(ctx, networkinfo.Connections)
		DescribeCustomerGatewayAssociationsFromARN(ctx, networkinfo.CustomerGatewayAssociations)
		DescribeDevicesFromARN(ctx, networkinfo.Devices)
		DescribeLinkAssociationsFromARN(ctx, networkinfo.LinkAssociations)
		DescribeLinksFromARN(ctx, networkinfo.Links)
		DescribeSitesFromARN(ctx, networkinfo.Sites)
		DescribeTransitGatewayConnectPeerAssociationsFromARN(ctx, networkinfo.TransitGatewayConnectPeerAssociations)
		*/

		tgARNs := make([]*string, len(networkinfo.TransitGatewayRegistrations))
		for i, arn := range networkinfo.TransitGatewayRegistrations {
			tgARNs[i] = arn.TransitGatewayArn
		}

		tgs, err := vpc.DescribeTransitGatewayRegistrationsFromARN(ctx, tgARNs)
		if err != nil {
			return nil, err
		}

		gnetworkchildren = append(gnetworkchildren, mapRegistrations(tgs))

		gnetwork := &opts.TreeData{
			Name:     name,
			Children: gnetworkchildren,
		}

		globalnetworknodes = append(globalnetworknodes, gnetwork)
	}

	global := make([]opts.TreeData, 0)

	switch len(globalnetworknodes) {
	case 0:
		return nil, errors.New("could not find any global network")

	case 1:
		global = append(global, *globalnetworknodes[0])

	default:
		global = append(global,
			opts.TreeData{
				Name:     "Global Networks",
				Children: globalnetworknodes,
			},
		)

	}

	return global, nil

}

func (p *NM) Execute(args []string) error {

	ctx := context.Background()

	closer, err := tracing.InitTracing()
	if err != nil {
		fmt.Printf("failed to initialize tracing, %s\n", err)
	}
	defer closer.Close()

	tracer := opentracing.GlobalTracer()
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, tracer, "chart")
	defer span.Finish()

	// NetworkManager is only available in us-west-2 (for now...)
	p.Region = "us-west-2"
	tree, err := NetworkManager(ctx, p)

	if err != nil {
		return err
	}

	g := NewTree("Global Networks")
	g.AddSeries("tree", tree).
		SetSeriesOptions(
			charts.WithTreeOpts(
				opts.TreeChart{
					Layout:           "orthogonal",
					Orient:           "LR",
					InitialTreeDepth: -1,
					Right:            "250px",
					Left:             "150px",
					Roam:             true,
					Leaves: &opts.TreeLeaves{
						Label: &opts.Label{Show: true, Position: "right", Color: "Black"},
					},
				},
			),
			charts.WithLabelOpts(opts.Label{Show: true, Position: "top", Color: "Black"}),
		)

	page := NewPage()
	page.AddCharts(g)
	f, err := os.Create("globalnetworks.html")
	if err != nil {
		panic(err)
	}

	page.Render(io.MultiWriter(f))
	return nil

}
