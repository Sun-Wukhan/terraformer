// Copyright 2026 The Terraformer Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aws

import (
	"context"
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
)

var networkManagerAllowEmptyValues = []string{"tags."}

type NetworkManagerGenerator struct {
	AWSService
}

// arnResourceName derives a Terraform resource name from an ARN by taking
// the segment after the last "/" (or the whole ARN when it has none).
func arnResourceName(arn string) string {
	if i := strings.LastIndex(arn, "/"); i >= 0 && i < len(arn)-1 {
		return arn[i+1:]
	}
	return arn
}

// InitResources enumerates the AWS Network Manager objects visible in the
// configured region: global networks with their sites, devices, links,
// connections, customer gateway associations, transit gateway
// registrations, transit gateway Connect peer associations and link
// associations; core networks with their attachments (VPC, Connect with
// its Connect peers, site-to-site VPN, transit gateway route table and
// Direct Connect gateway), core network policy attachments, prefix list
// associations and attachment routing policy associations; and transit
// gateway peerings.
//
// aws_networkmanager_attachment_accepter is intentionally not generated:
// it is an action-style resource with no importer in the Terraform AWS
// provider and no independent existence in the Network Manager API (its
// Read derives everything from the attachment itself), so there is no
// object to discover.
func (g *NetworkManagerGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	client := networkmanager.NewFromConfig(config)

	if err := g.addGlobalNetworks(client); err != nil {
		return err
	}
	if err := g.addCoreNetworks(client); err != nil {
		return err
	}
	return g.addTransitGatewayPeerings(client)
}

func (g *NetworkManagerGenerator) addGlobalNetworks(client *networkmanager.Client) error {
	paginator := networkmanager.NewDescribeGlobalNetworksPaginator(client, &networkmanager.DescribeGlobalNetworksInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, globalNetwork := range page.GlobalNetworks {
			globalNetworkID := *globalNetwork.GlobalNetworkId
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				globalNetworkID,
				globalNetworkID,
				"aws_networkmanager_global_network",
				"aws",
				networkManagerAllowEmptyValues,
			))
			for _, add := range []func(*networkmanager.Client, string) error{
				g.addSites,
				g.addDevices,
				g.addLinks,
				g.addConnections,
				g.addCustomerGatewayAssociations,
				g.addTransitGatewayRegistrations,
				g.addTransitGatewayConnectPeerAssociations,
				g.addLinkAssociations,
			} {
				if err := add(client, globalNetworkID); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (g *NetworkManagerGenerator) addSites(client *networkmanager.Client, globalNetworkID string) error {
	paginator := networkmanager.NewGetSitesPaginator(client, &networkmanager.GetSitesInput{
		GlobalNetworkId: aws.String(globalNetworkID),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, site := range page.Sites {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				*site.SiteId,
				*site.SiteId,
				"aws_networkmanager_site",
				"aws",
				map[string]string{"global_network_id": globalNetworkID},
				networkManagerAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

func (g *NetworkManagerGenerator) addDevices(client *networkmanager.Client, globalNetworkID string) error {
	paginator := networkmanager.NewGetDevicesPaginator(client, &networkmanager.GetDevicesInput{
		GlobalNetworkId: aws.String(globalNetworkID),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, device := range page.Devices {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				*device.DeviceId,
				*device.DeviceId,
				"aws_networkmanager_device",
				"aws",
				map[string]string{"global_network_id": globalNetworkID},
				networkManagerAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

func (g *NetworkManagerGenerator) addLinks(client *networkmanager.Client, globalNetworkID string) error {
	paginator := networkmanager.NewGetLinksPaginator(client, &networkmanager.GetLinksInput{
		GlobalNetworkId: aws.String(globalNetworkID),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, link := range page.Links {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				*link.LinkId,
				*link.LinkId,
				"aws_networkmanager_link",
				"aws",
				map[string]string{"global_network_id": globalNetworkID},
				networkManagerAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

func (g *NetworkManagerGenerator) addConnections(client *networkmanager.Client, globalNetworkID string) error {
	paginator := networkmanager.NewGetConnectionsPaginator(client, &networkmanager.GetConnectionsInput{
		GlobalNetworkId: aws.String(globalNetworkID),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, connection := range page.Connections {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				*connection.ConnectionId,
				*connection.ConnectionId,
				"aws_networkmanager_connection",
				"aws",
				map[string]string{"global_network_id": globalNetworkID},
				networkManagerAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

// Provider ID format: global network ID, customer gateway ARN.
func (g *NetworkManagerGenerator) addCustomerGatewayAssociations(client *networkmanager.Client, globalNetworkID string) error {
	paginator := networkmanager.NewGetCustomerGatewayAssociationsPaginator(client, &networkmanager.GetCustomerGatewayAssociationsInput{
		GlobalNetworkId: aws.String(globalNetworkID),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, association := range page.CustomerGatewayAssociations {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				globalNetworkID+","+*association.CustomerGatewayArn,
				arnResourceName(*association.CustomerGatewayArn),
				"aws_networkmanager_customer_gateway_association",
				"aws",
				map[string]string{"global_network_id": globalNetworkID},
				networkManagerAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

// Provider ID format: global network ID, transit gateway ARN.
func (g *NetworkManagerGenerator) addTransitGatewayRegistrations(client *networkmanager.Client, globalNetworkID string) error {
	paginator := networkmanager.NewGetTransitGatewayRegistrationsPaginator(client, &networkmanager.GetTransitGatewayRegistrationsInput{
		GlobalNetworkId: aws.String(globalNetworkID),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, registration := range page.TransitGatewayRegistrations {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				globalNetworkID+","+*registration.TransitGatewayArn,
				arnResourceName(*registration.TransitGatewayArn),
				"aws_networkmanager_transit_gateway_registration",
				"aws",
				map[string]string{"global_network_id": globalNetworkID},
				networkManagerAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

// Provider ID format: global network ID, transit gateway Connect peer ARN.
func (g *NetworkManagerGenerator) addTransitGatewayConnectPeerAssociations(client *networkmanager.Client, globalNetworkID string) error {
	paginator := networkmanager.NewGetTransitGatewayConnectPeerAssociationsPaginator(client, &networkmanager.GetTransitGatewayConnectPeerAssociationsInput{
		GlobalNetworkId: aws.String(globalNetworkID),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, association := range page.TransitGatewayConnectPeerAssociations {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				globalNetworkID+","+*association.TransitGatewayConnectPeerArn,
				arnResourceName(*association.TransitGatewayConnectPeerArn),
				"aws_networkmanager_transit_gateway_connect_peer_association",
				"aws",
				map[string]string{"global_network_id": globalNetworkID},
				networkManagerAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

// Provider ID format: global network ID, link ID, device ID.
func (g *NetworkManagerGenerator) addLinkAssociations(client *networkmanager.Client, globalNetworkID string) error {
	paginator := networkmanager.NewGetLinkAssociationsPaginator(client, &networkmanager.GetLinkAssociationsInput{
		GlobalNetworkId: aws.String(globalNetworkID),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, association := range page.LinkAssociations {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				globalNetworkID+","+*association.LinkId+","+*association.DeviceId,
				*association.LinkId,
				"aws_networkmanager_link_association",
				"aws",
				map[string]string{"global_network_id": globalNetworkID},
				networkManagerAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

func (g *NetworkManagerGenerator) addCoreNetworks(client *networkmanager.Client) error {
	paginator := networkmanager.NewListCoreNetworksPaginator(client, &networkmanager.ListCoreNetworksInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, coreNetwork := range page.CoreNetworks {
			coreNetworkID := *coreNetwork.CoreNetworkId
			attributes := map[string]string{}
			if coreNetwork.GlobalNetworkId != nil {
				attributes["global_network_id"] = *coreNetwork.GlobalNetworkId
			}
			g.Resources = append(g.Resources, terraformutils.NewResource(
				coreNetworkID,
				coreNetworkID,
				"aws_networkmanager_core_network",
				"aws",
				attributes,
				networkManagerAllowEmptyValues,
				map[string]interface{}{},
			))
			if err := g.addAttachments(client, coreNetworkID); err != nil {
				return err
			}
			if err := g.addCoreNetworkPolicyAttachment(client, coreNetworkID); err != nil {
				return err
			}
			if err := g.addPrefixListAssociations(client, coreNetworkID); err != nil {
				return err
			}
			if err := g.addAttachmentRoutingPolicyAssociations(client, coreNetworkID); err != nil {
				return err
			}
		}
	}
	return nil
}

// Every attachment flavour is keyed by the attachment ID alone; the
// attachment type filter of ListAttachments separates them.
func (g *NetworkManagerGenerator) addAttachments(client *networkmanager.Client, coreNetworkID string) error {
	for attachmentType, resourceType := range map[types.AttachmentType]string{
		types.AttachmentTypeVpc:                      "aws_networkmanager_vpc_attachment",
		types.AttachmentTypeConnect:                  "aws_networkmanager_connect_attachment",
		types.AttachmentTypeSiteToSiteVpn:            "aws_networkmanager_site_to_site_vpn_attachment",
		types.AttachmentTypeTransitGatewayRouteTable: "aws_networkmanager_transit_gateway_route_table_attachment",
		types.AttachmentTypeDirectConnectGateway:     "aws_networkmanager_dx_gateway_attachment",
	} {
		paginator := networkmanager.NewListAttachmentsPaginator(client, &networkmanager.ListAttachmentsInput{
			CoreNetworkId:  aws.String(coreNetworkID),
			AttachmentType: attachmentType,
		})
		for paginator.HasMorePages() {
			page, err := paginator.NextPage(context.TODO())
			if err != nil {
				return err
			}
			for _, attachment := range page.Attachments {
				attachmentID := *attachment.AttachmentId
				g.Resources = append(g.Resources, terraformutils.NewResource(
					attachmentID,
					attachmentID,
					resourceType,
					"aws",
					map[string]string{"core_network_id": coreNetworkID},
					networkManagerAllowEmptyValues,
					map[string]interface{}{},
				))
				if attachmentType == types.AttachmentTypeConnect {
					if err := g.addConnectPeers(client, coreNetworkID, attachmentID); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func (g *NetworkManagerGenerator) addConnectPeers(client *networkmanager.Client, coreNetworkID, connectAttachmentID string) error {
	paginator := networkmanager.NewListConnectPeersPaginator(client, &networkmanager.ListConnectPeersInput{
		CoreNetworkId:       aws.String(coreNetworkID),
		ConnectAttachmentId: aws.String(connectAttachmentID),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, peer := range page.ConnectPeers {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				*peer.ConnectPeerId,
				*peer.ConnectPeerId,
				"aws_networkmanager_connect_peer",
				"aws",
				map[string]string{"connect_attachment_id": connectAttachmentID},
				networkManagerAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

// The policy attachment is a per-core-network singleton keyed by the core
// network ID; it exists only once at least one policy version does.
func (g *NetworkManagerGenerator) addCoreNetworkPolicyAttachment(client *networkmanager.Client, coreNetworkID string) error {
	paginator := networkmanager.NewListCoreNetworkPolicyVersionsPaginator(client, &networkmanager.ListCoreNetworkPolicyVersionsInput{
		CoreNetworkId: aws.String(coreNetworkID),
	})
	hasVersion := false
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		if len(page.CoreNetworkPolicyVersions) > 0 {
			hasVersion = true
		}
	}
	if !hasVersion {
		return nil
	}
	g.Resources = append(g.Resources, terraformutils.NewResource(
		coreNetworkID,
		coreNetworkID,
		"aws_networkmanager_core_network_policy_attachment",
		"aws",
		map[string]string{"core_network_id": coreNetworkID},
		networkManagerAllowEmptyValues,
		map[string]interface{}{},
	))
	return nil
}

// Provider ID format: core network ID, prefix list ARN.
func (g *NetworkManagerGenerator) addPrefixListAssociations(client *networkmanager.Client, coreNetworkID string) error {
	paginator := networkmanager.NewListCoreNetworkPrefixListAssociationsPaginator(client, &networkmanager.ListCoreNetworkPrefixListAssociationsInput{
		CoreNetworkId: aws.String(coreNetworkID),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, association := range page.PrefixListAssociations {
			name := arnResourceName(*association.PrefixListArn)
			if association.PrefixListAlias != nil && *association.PrefixListAlias != "" {
				name = *association.PrefixListAlias
			}
			g.Resources = append(g.Resources, terraformutils.NewResource(
				coreNetworkID+","+*association.PrefixListArn,
				name,
				"aws_networkmanager_prefix_list_association",
				"aws",
				map[string]string{"core_network_id": coreNetworkID},
				networkManagerAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

// Provider ID format: core network ID, attachment ID. Only attachments
// carrying a routing policy label have this resource.
func (g *NetworkManagerGenerator) addAttachmentRoutingPolicyAssociations(client *networkmanager.Client, coreNetworkID string) error {
	paginator := networkmanager.NewListAttachmentRoutingPolicyAssociationsPaginator(client, &networkmanager.ListAttachmentRoutingPolicyAssociationsInput{
		CoreNetworkId: aws.String(coreNetworkID),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, association := range page.AttachmentRoutingPolicyAssociations {
			if association.RoutingPolicyLabel == nil || *association.RoutingPolicyLabel == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewResource(
				coreNetworkID+","+*association.AttachmentId,
				*association.RoutingPolicyLabel,
				"aws_networkmanager_attachment_routing_policy_label",
				"aws",
				map[string]string{"core_network_id": coreNetworkID},
				networkManagerAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

func (g *NetworkManagerGenerator) addTransitGatewayPeerings(client *networkmanager.Client) error {
	paginator := networkmanager.NewListPeeringsPaginator(client, &networkmanager.ListPeeringsInput{
		PeeringType: types.PeeringTypeTransitGateway,
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, peering := range page.Peerings {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*peering.PeeringId,
				*peering.PeeringId,
				"aws_networkmanager_transit_gateway_peering",
				"aws",
				networkManagerAllowEmptyValues,
			))
		}
	}
	return nil
}
