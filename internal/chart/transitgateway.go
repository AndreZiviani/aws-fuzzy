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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	nm "github.com/aws/aws-sdk-go-v2/service/networkmanager"
	nmtypes "github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	opentracing "github.com/opentracing/opentracing-go"
	//"github.com/opentracing/opentracing-go/log"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
)

type GlobalNetwork struct {
	Connections                           []nmtypes.Connection
	CustomerGatewayAssociations           []nmtypes.CustomerGatewayAssociation
	Devices                               []nmtypes.Device
	LinkAssociations                      []nmtypes.LinkAssociation
	Links                                 []nmtypes.Link
	Sites                                 []nmtypes.Site
	TransitGatewayConnectPeerAssociations []nmtypes.TransitGatewayConnectPeerAssociation
	TransitGatewayRegistrations           []nmtypes.TransitGatewayRegistration
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

func NewTree() *charts.Tree {

	graph := charts.NewTree()
	graph.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Width: "100%", Height: "95vh"}),
		charts.WithTitleOpts(opts.Title{Title: "Global Networks"}),
		charts.WithToolboxOpts(opts.Toolbox{
			Show: true,
			Feature: &opts.ToolBoxFeature{
				SaveAsImage: &opts.ToolBoxFeatureSaveAsImage{
					Show:  true,
					Type:  "png",
					Title: "Download as PNG",
				},
			},
		}),
		//charts.WithLegendOpts(opts.Legend{Show: true}),
		charts.WithTooltipOpts(opts.Tooltip{Show: false}),
	)

	return graph
}

func GetTransitGatewayAttachments(ctx context.Context, ec2client *ec2.Client, tg *string) ([]ec2types.TransitGatewayAttachment, error) {
	spanDescribeTGAttachments, ctx := opentracing.StartSpanFromContext(ctx, "describetransitgwattachments")
	defer spanDescribeTGAttachments.Finish()

	tgid := strings.Split(aws.ToString(tg), "/")

	tgattach, err := ec2client.DescribeTransitGatewayAttachments(ctx,
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

	return tgattach.TransitGatewayAttachments, nil
}

func GetGlobalNetworks(ctx context.Context, nmclient *nm.Client) ([]nmtypes.GlobalNetwork, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "getglobalnetworks")
	defer span.Finish()

	globalnetworks, err := nmclient.DescribeGlobalNetworks(ctx,
		&nm.DescribeGlobalNetworksInput{})

	if err != nil {
		fmt.Printf("failed to describe global networks, %s\n", err)
		return nil, err
	}

	return globalnetworks.GlobalNetworks, nil
}

func DescribeGlobalNetwork(ctx context.Context, nmclient *nm.Client, network nmtypes.GlobalNetwork) (*GlobalNetwork, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "describeglobalnetwork")
	defer span.Finish()

	/* TODO:
	connections, err := nmclient.GetConnections(ctx,
		&nm.GetConnectionsInput{
			GlobalNetworkId: network.GlobalNetworkId,
		},
	)
	if err != nil {
		return nil, err
	}

	customerGatewaysAssociations, err := nmclient.GetCustomerGatewayAssociations(ctx,
		&nm.GetCustomerGatewayAssociationsInput{
			GlobalNetworkId: network.GlobalNetworkId,
		},
	)
	if err != nil {
		return nil, err
	}

	devices, err := nmclient.GetDevices(ctx,
		&nm.GetDevicesInput{
			GlobalNetworkId: network.GlobalNetworkId,
		},
	)
	if err != nil {
		return nil, err
	}

	linkAssociations, err := nmclient.GetLinkAssociations(ctx,
		&nm.GetLinkAssociationsInput{
			GlobalNetworkId: network.GlobalNetworkId,
		},
	)
	if err != nil {
		return nil, err
	}

	links, err := nmclient.GetLinks(ctx,
		&nm.GetLinksInput{
			GlobalNetworkId: network.GlobalNetworkId,
		},
	)
	if err != nil {
		return nil, err
	}

	sites, err := nmclient.GetSites(ctx,
		&nm.GetSitesInput{
			GlobalNetworkId: network.GlobalNetworkId,
		},
	)
	if err != nil {
		return nil, err
	}

	transitGatewayAssociations, err := nmclient.GetTransitGatewayConnectPeerAssociations(ctx,
		&nm.GetTransitGatewayConnectPeerAssociationsInput{
			GlobalNetworkId: network.GlobalNetworkId,
		},
	)
	if err != nil {
		return nil, err
	}
	*/

	transitGatewayRegistrations, err := nmclient.GetTransitGatewayRegistrations(ctx,
		&nm.GetTransitGatewayRegistrationsInput{
			GlobalNetworkId: network.GlobalNetworkId,
		},
	)
	if err != nil {
		return nil, err
	}

	return &GlobalNetwork{
		/* TODO:
		Connections:                           connections.Connections,
		CustomerGatewayAssociations:           customerGatewaysAssociations.CustomerGatewayAssociations,
		Devices:                               devices.Devices,
		LinkAssociations:                      linkAssociations.LinkAssociations,
		Links:                                 links.Links,
		Sites:                                 sites.Sites,
		TransitGatewayConnectPeerAssociations: transitGatewayAssociations.TransitGatewayConnectPeerAssociations,
		*/
		TransitGatewayRegistrations: transitGatewayRegistrations.TransitGatewayRegistrations,
	}, nil

}
func GetEC2Client(ctx context.Context, clients map[string]map[string]*ec2.Client, profile *string, region *string) (*ec2.Client, error) {
	// check if we already have a client
	if _, ok := clients[*profile]; ok {
		if client, ok := clients[*profile][*region]; ok {
			return client, nil
		}
	} else {
		// we have nothing
		clients[*profile] = map[string]*ec2.Client{}
	}

	// creating a client on the specified region
	client, _ := NewEC2Client(ctx, profile, region)
	clients[*profile][*region] = client

	return client, nil
}

func NewEC2Client(ctx context.Context, profile *string, region *string) (*ec2.Client, error) {
	spanNewEC2Client, ctx := opentracing.StartSpanFromContext(ctx, "newec2client")
	defer spanNewEC2Client.Finish()

	creds, err := sso.GetCredentials(ctx, *profile, false)
	if err != nil {
		return nil, err
	}

	cfg, err := sso.NewAwsConfig(ctx, creds, config.WithRegion(*region))
	if err != nil {
		return nil, err
	}

	return ec2.NewFromConfig(cfg), nil
}

func DescribeTransitGatewayRegistrationsFromARN(ctx context.Context, transitGateways []nmtypes.TransitGatewayRegistration) (*opts.TreeData, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "describetransitgateways")
	defer span.Finish()

	clients := make(map[string]map[string]*ec2.Client)
	regions := make(map[string][]*opts.TreeData, 0)

	transitgatewaynodes := make([]*opts.TreeData, 0)

	for _, tg := range transitGateways {
		tgwchildren := make([]*opts.TreeData, 0)
		vpcnodes := make([]*opts.TreeData, 0)
		vpnnodes := make([]*opts.TreeData, 0)
		dxgwnodes := make([]*opts.TreeData, 0)
		connectnodes := make([]*opts.TreeData, 0)
		peeringnodes := make([]*opts.TreeData, 0)
		tgwpeeringnodes := make([]*opts.TreeData, 0)

		arn, _ := arn.Parse(*tg.TransitGatewayArn)

		profile, _, err := sso.GetAccount(arn.AccountID)
		if err != nil {
			fmt.Printf("failed to get account, %s\n", err)
			return nil, err
		}

		client, err := GetEC2Client(ctx, clients, profile, &arn.Region)
		if err != nil {
			fmt.Printf("failed to create ec2 client, %s\n", err)
			return nil, err
		}

		attachments, err := GetTransitGatewayAttachments(ctx, client, &arn.Resource)

		// Create Tree nodes
		for _, attachment := range attachments {
			account, _, err := sso.GetAccount(aws.ToString(attachment.ResourceOwnerId))
			if err != nil {
				account = attachment.ResourceOwnerId
			}

			name := fmt.Sprintf("%s\n%s\n%s",
				common.GetEC2Tag(attachment.Tags, "Name", aws.ToString(attachment.TransitGatewayAttachmentId)),
				aws.ToString(account),
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

		// get transit gateway Name tag
		tginfo, _ := DescribeTransitGateway(ctx, client, &arn.Resource)
		name := common.GetEC2Tag(tginfo.Tags, "Name", strings.Split(arn.Resource, "/")[1])

		tgw := &opts.TreeData{
			Name:     name,
			Children: tgwchildren,
		}

		if _, ok := regions[arn.Region]; !ok {
			regionnodes := make([]*opts.TreeData, 0)
			regions[arn.Region] = regionnodes
		}
		regions[arn.Region] = append(regions[arn.Region], tgw)

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
	}, nil
}

func NetworkManager(ctx context.Context, p *NMCommand) ([]opts.TreeData, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "networkmanager")
	defer span.Finish()

	creds, err := sso.GetCredentials(ctx, p.Profile, false)
	if err != nil {
		return nil, err
	}

	cfg, err := sso.NewAwsConfig(ctx, creds, config.WithRegion(p.Region))
	if err != nil {
		return nil, err
	}

	nmclient := nm.NewFromConfig(cfg)

	globalnetworks, err := GetGlobalNetworks(ctx, nmclient)

	globalnetworknodes := make([]*opts.TreeData, 0)
	for _, network := range globalnetworks {
		arn, _ := arn.Parse(*network.GlobalNetworkArn)
		name := common.GetNMTag(network.Tags, "Name", strings.Split(arn.Resource, "/")[1])

		gnetworkchildren := make([]*opts.TreeData, 0)

		networkinfo, err := DescribeGlobalNetwork(ctx, nmclient, network)
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
		tgregistrations, err := DescribeTransitGatewayRegistrationsFromARN(ctx, networkinfo.TransitGatewayRegistrations)
		if err != nil {
			return nil, err
		}

		gnetworkchildren = append(gnetworkchildren, tgregistrations)

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

func (p *NMCommand) Execute(args []string) error {

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
		panic(err)
	}

	g := NewTree()
	g.AddSeries("tree", tree).
		SetSeriesOptions(
			charts.WithTreeOpts(
				opts.TreeChart{
					Layout:           "orthogonal",
					Orient:           "LR",
					InitialTreeDepth: -1,
					Right:            "250px",
					Left:             "150px",
					Leaves: &opts.TreeLeaves{
						Label: &opts.Label{Show: true, Position: "right", Color: "Black"},
					},
				},
			),
			charts.WithLabelOpts(opts.Label{Show: true, Position: "top", Color: "Black"}),
		)

	page := components.NewPage()
	page.AddCharts(g)
	f, err := os.Create("tree.html")
	if err != nil {
		panic(err)
	}

	page.Render(io.MultiWriter(f))
	return nil

}
