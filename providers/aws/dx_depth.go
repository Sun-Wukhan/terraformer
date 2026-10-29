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

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directconnect"
)

// Depth enumeration for Direct Connect, called from
// DirectConnectGenerator.InitResources.
//
// Intentionally not generated (no Import section in the Terraform AWS
// provider documentation, i.e. no provider importer):
//   - aws_dx_bgp_peer: BGP peers are sub-objects of a virtual interface
//     and the provider resource has no import format.
//   - aws_dx_connection_association: the connection/LAG association has
//     no import format; the association is visible as the connection's
//     lag_id attribute.
//   - aws_dx_connection_confirmation: an action-style confirmation
//     resource with no import format and no independent API object.
//   - aws_dx_hosted_connection: partner-side hosted connections have
//     no import format in the provider documentation.
//   - aws_dx_macsec_key_association: the provider documentation gives
//     no import ID format for the MACSec key association.
func (g *DirectConnectGenerator) addDepthResources(config aws.Config, svc *directconnect.Client) error {
	if err := g.addLags(svc); err != nil {
		return err
	}
	if err := g.addGatewayAssociations(svc); err != nil {
		return err
	}
	if err := g.addGatewayAssociationProposals(svc); err != nil {
		return err
	}
	return g.addHostedAndTransitVirtualInterfaces(config, svc)
}

func (g *DirectConnectGenerator) addLags(svc *directconnect.Client) error {
	out, err := svc.DescribeLags(context.TODO(), &directconnect.DescribeLagsInput{})
	if err != nil {
		return err
	}
	for _, lag := range out.Lags {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			StringValue(lag.LagId),
			StringValue(lag.LagName),
			"aws_dx_lag",
			"aws",
			dxAllowEmptyValues,
		))
	}
	return nil
}

// Provider ID format: dx_gateway_id/associated_gateway_id.
func (g *DirectConnectGenerator) addGatewayAssociations(svc *directconnect.Client) error {
	gateways, err := svc.DescribeDirectConnectGateways(context.TODO(), &directconnect.DescribeDirectConnectGatewaysInput{})
	if err != nil {
		return err
	}
	for _, gateway := range gateways.DirectConnectGateways {
		input := &directconnect.DescribeDirectConnectGatewayAssociationsInput{
			DirectConnectGatewayId: gateway.DirectConnectGatewayId,
		}
		for {
			out, err := svc.DescribeDirectConnectGatewayAssociations(context.TODO(), input)
			if err != nil {
				return err
			}
			for _, association := range out.DirectConnectGatewayAssociations {
				if association.AssociatedGateway == nil || association.AssociatedGateway.Id == nil {
					continue
				}
				gatewayID := StringValue(association.AssociatedGateway.Id)
				g.Resources = append(g.Resources, terraformutils.NewResource(
					StringValue(gateway.DirectConnectGatewayId)+"/"+gatewayID,
					gatewayID,
					"aws_dx_gateway_association",
					"aws",
					map[string]string{"dx_gateway_id": StringValue(gateway.DirectConnectGatewayId)},
					dxAllowEmptyValues,
					map[string]interface{}{},
				))
			}
			if out.NextToken == nil {
				break
			}
			input.NextToken = out.NextToken
		}
	}
	return nil
}

// Provider ID format: the proposal ID alone.
func (g *DirectConnectGenerator) addGatewayAssociationProposals(svc *directconnect.Client) error {
	input := &directconnect.DescribeDirectConnectGatewayAssociationProposalsInput{}
	for {
		out, err := svc.DescribeDirectConnectGatewayAssociationProposals(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, proposal := range out.DirectConnectGatewayAssociationProposals {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(proposal.ProposalId),
				StringValue(proposal.ProposalId),
				"aws_dx_gateway_association_proposal",
				"aws",
				dxAllowEmptyValues,
			))
		}
		if out.NextToken == nil {
			break
		}
		input.NextToken = out.NextToken
	}
	return nil
}

// Transit virtual interfaces are keyed by the VIF ID. A virtual
// interface whose owner account differs from its connection's owner
// account is a hosted virtual interface: from the connection owner's
// side it is an aws_dx_hosted_*_virtual_interface, and from the VIF
// owner's side it is the matching *_accepter resource. Which side this
// account is on is decided by comparing both owners with the caller
// account. Directly owned VIFs (owner == connection owner) keep the
// plain aws_dx_*_virtual_interface types generated by dx.go (private
// and public) or the transit type generated here.
func (g *DirectConnectGenerator) addHostedAndTransitVirtualInterfaces(config aws.Config, svc *directconnect.Client) error {
	accountID, err := g.getAccountNumber(config)
	if err != nil {
		return err
	}
	connections, err := svc.DescribeConnections(context.TODO(), &directconnect.DescribeConnectionsInput{})
	if err != nil {
		return err
	}
	connectionOwner := map[string]string{}
	for _, connection := range connections.Connections {
		connectionOwner[StringValue(connection.ConnectionId)] = StringValue(connection.OwnerAccount)
	}
	// Virtual interfaces can ride directly on a LAG; LAGs are not
	// returned by DescribeConnections, so register their owners too.
	lags, err := svc.DescribeLags(context.TODO(), &directconnect.DescribeLagsInput{})
	if err != nil {
		return err
	}
	for _, lag := range lags.Lags {
		connectionOwner[StringValue(lag.LagId)] = StringValue(lag.OwnerAccount)
	}
	out, err := svc.DescribeVirtualInterfaces(context.TODO(), &directconnect.DescribeVirtualInterfacesInput{})
	if err != nil {
		return err
	}
	for _, vif := range out.VirtualInterfaces {
		vifID := StringValue(vif.VirtualInterfaceId)
		vifType := StringValue(vif.VirtualInterfaceType)
		owner := StringValue(vif.OwnerAccount)
		hosted := owner != connectionOwner[StringValue(vif.ConnectionId)]
		var resourceType string
		switch {
		case !hosted && vifType == "transit":
			resourceType = "aws_dx_transit_virtual_interface"
		case hosted && owner == StringValue(accountID):
			// This account accepted a VIF allocated by another account.
			switch vifType {
			case "private":
				resourceType = "aws_dx_hosted_private_virtual_interface_accepter"
			case "public":
				resourceType = "aws_dx_hosted_public_virtual_interface_accepter"
			case "transit":
				resourceType = "aws_dx_hosted_transit_virtual_interface_accepter"
			}
		case hosted:
			// This account allocated the VIF for another account.
			switch vifType {
			case "private":
				resourceType = "aws_dx_hosted_private_virtual_interface"
			case "public":
				resourceType = "aws_dx_hosted_public_virtual_interface"
			case "transit":
				resourceType = "aws_dx_hosted_transit_virtual_interface"
			}
		}
		if resourceType == "" {
			continue
		}
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			vifID,
			StringValue(vif.VirtualInterfaceName),
			resourceType,
			"aws",
			dxAllowEmptyValues,
		))
	}
	return nil
}
