package vpc

import (
	"context"
	"fmt"

	nm "github.com/aws/aws-sdk-go-v2/service/networkmanager"
	nmtypes "github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	opentracing "github.com/opentracing/opentracing-go"
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
