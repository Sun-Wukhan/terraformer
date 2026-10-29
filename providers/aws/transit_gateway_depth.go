// Copyright 2024 The Terraformer Authors.
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
	"log"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// getTransitGatewayDepth extends TransitGatewayGenerator with connects,
// connect peers, multicast domains, peering attachments, policy tables,
// route-table associations/propagations/routes, prefix list references,
// metering policies and VPN concentrators. Failures are logged and skipped
// so one unsupported API never blocks the base transit gateway import.
//
// Notable skips in this family (see the cluster ledger for the full list):
//   - aws_ec2_transit_gateway_multicast_domain_association,
//     aws_ec2_transit_gateway_multicast_group_member and
//     aws_ec2_transit_gateway_multicast_group_source: no provider importer.
//   - aws_ec2_transit_gateway_default_route_table_association and
//     aws_ec2_transit_gateway_default_route_table_propagation: no provider
//     importer.
//   - aws_ec2_transit_gateway_vpc_attachment_accepter and
//     aws_ec2_transit_gateway_peering_attachment_accepter: accepter-side
//     views of attachments already enumerated by this generator; only
//     meaningful under the accepter account's provider.
func (g *TransitGatewayGenerator) getTransitGatewayDepth(svc *ec2.Client) {
	g.getTransitGatewayRouteTableRelations(svc)
	g.getTransitGatewayPolicyTables(svc)
	g.getTransitGatewayConnects(svc)
	g.getTransitGatewayMulticastDomains(svc)
	g.getTransitGatewayPeeringAttachments(svc)
	g.getTransitGatewayMeteringPolicies(svc)
	g.getVpnConcentrators(svc)
}

func (g *TransitGatewayGenerator) getTransitGatewayRouteTableRelations(svc *ec2.Client) {
	p := ec2.NewDescribeTransitGatewayRouteTablesPaginator(svc, &ec2.DescribeTransitGatewayRouteTablesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, routeTable := range page.TransitGatewayRouteTables {
			if routeTable.TransitGatewayRouteTableId == nil {
				continue
			}
			routeTableID := *routeTable.TransitGatewayRouteTableId
			g.getTransitGatewayRouteTableAssociations(svc, routeTableID)
			g.getTransitGatewayRouteTablePropagations(svc, routeTableID)
			g.getTransitGatewayPrefixListReferences(svc, routeTableID)
			g.getTransitGatewayRoutes(svc, routeTableID)
		}
	}
}

func (g *TransitGatewayGenerator) getTransitGatewayRouteTableAssociations(svc *ec2.Client, routeTableID string) {
	p := ec2.NewGetTransitGatewayRouteTableAssociationsPaginator(svc, &ec2.GetTransitGatewayRouteTableAssociationsInput{
		TransitGatewayRouteTableId: aws.String(routeTableID),
	})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, association := range page.Associations {
			if association.TransitGatewayAttachmentId == nil {
				continue
			}
			id := routeTableID + "_" + *association.TransitGatewayAttachmentId
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				id,
				"aws_ec2_transit_gateway_route_table_association",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
}

func (g *TransitGatewayGenerator) getTransitGatewayRouteTablePropagations(svc *ec2.Client, routeTableID string) {
	p := ec2.NewGetTransitGatewayRouteTablePropagationsPaginator(svc, &ec2.GetTransitGatewayRouteTablePropagationsInput{
		TransitGatewayRouteTableId: aws.String(routeTableID),
	})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, propagation := range page.TransitGatewayRouteTablePropagations {
			if propagation.TransitGatewayAttachmentId == nil {
				continue
			}
			id := routeTableID + "_" + *propagation.TransitGatewayAttachmentId
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				id,
				"aws_ec2_transit_gateway_route_table_propagation",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
}

func (g *TransitGatewayGenerator) getTransitGatewayPrefixListReferences(svc *ec2.Client, routeTableID string) {
	p := ec2.NewGetTransitGatewayPrefixListReferencesPaginator(svc, &ec2.GetTransitGatewayPrefixListReferencesInput{
		TransitGatewayRouteTableId: aws.String(routeTableID),
	})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, reference := range page.TransitGatewayPrefixListReferences {
			if reference.PrefixListId == nil {
				continue
			}
			id := routeTableID + "_" + *reference.PrefixListId
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				id,
				"aws_ec2_transit_gateway_prefix_list_reference",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
}

// Only static routes become aws_ec2_transit_gateway_route resources;
// propagated routes are owned by their propagations.
func (g *TransitGatewayGenerator) getTransitGatewayRoutes(svc *ec2.Client, routeTableID string) {
	out, err := svc.SearchTransitGatewayRoutes(context.TODO(), &ec2.SearchTransitGatewayRoutesInput{
		TransitGatewayRouteTableId: aws.String(routeTableID),
		Filters: []types.Filter{
			{
				Name:   aws.String("type"),
				Values: []string{"static"},
			},
		},
	})
	if err != nil {
		log.Println(err)
		return
	}
	for _, route := range out.Routes {
		if route.Type != types.TransitGatewayRouteTypeStatic {
			continue
		}
		destination := StringValue(route.DestinationCidrBlock)
		if destination == "" {
			destination = StringValue(route.PrefixListId)
		}
		if destination == "" {
			continue
		}
		id := routeTableID + "_" + destination
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			id,
			id,
			"aws_ec2_transit_gateway_route",
			"aws",
			awsNetCoreAllowEmptyValues,
		))
	}
}

func (g *TransitGatewayGenerator) getTransitGatewayPolicyTables(svc *ec2.Client) {
	p := ec2.NewDescribeTransitGatewayPolicyTablesPaginator(svc, &ec2.DescribeTransitGatewayPolicyTablesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, policyTable := range page.TransitGatewayPolicyTables {
			if policyTable.TransitGatewayPolicyTableId == nil {
				continue
			}
			policyTableID := *policyTable.TransitGatewayPolicyTableId
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				policyTableID,
				policyTableID,
				"aws_ec2_transit_gateway_policy_table",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
			g.getTransitGatewayPolicyTableAssociations(svc, policyTableID)
			g.getTransitGatewayPolicyTableEntries(svc, policyTableID)
		}
	}
}

func (g *TransitGatewayGenerator) getTransitGatewayPolicyTableAssociations(svc *ec2.Client, policyTableID string) {
	p := ec2.NewGetTransitGatewayPolicyTableAssociationsPaginator(svc, &ec2.GetTransitGatewayPolicyTableAssociationsInput{
		TransitGatewayPolicyTableId: aws.String(policyTableID),
	})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, association := range page.Associations {
			if association.TransitGatewayAttachmentId == nil {
				continue
			}
			id := policyTableID + "_" + *association.TransitGatewayAttachmentId
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				id,
				"aws_ec2_transit_gateway_policy_table_association",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
}

// aws_ec2_transit_gateway_policy_table_entry is imported by
// "<policy table id>,<rule number>".
func (g *TransitGatewayGenerator) getTransitGatewayPolicyTableEntries(svc *ec2.Client, policyTableID string) {
	out, err := svc.GetTransitGatewayPolicyTableEntries(context.TODO(), &ec2.GetTransitGatewayPolicyTableEntriesInput{
		TransitGatewayPolicyTableId: aws.String(policyTableID),
	})
	if err != nil {
		log.Println(err)
		return
	}
	for _, entry := range out.TransitGatewayPolicyTableEntries {
		if entry.PolicyRuleNumber == nil {
			continue
		}
		id := policyTableID + "," + *entry.PolicyRuleNumber
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			id,
			id,
			"aws_ec2_transit_gateway_policy_table_entry",
			"aws",
			awsNetCoreAllowEmptyValues,
		))
	}
}

func (g *TransitGatewayGenerator) getTransitGatewayConnects(svc *ec2.Client) {
	p := ec2.NewDescribeTransitGatewayConnectsPaginator(svc, &ec2.DescribeTransitGatewayConnectsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, connect := range page.TransitGatewayConnects {
			if connect.TransitGatewayAttachmentId == nil {
				continue
			}
			attachmentID := *connect.TransitGatewayAttachmentId
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				attachmentID,
				attachmentID,
				"aws_ec2_transit_gateway_connect",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
			g.getTransitGatewayConnectPeers(svc, attachmentID)
		}
	}
}

func (g *TransitGatewayGenerator) getTransitGatewayConnectPeers(svc *ec2.Client, attachmentID string) {
	p := ec2.NewDescribeTransitGatewayConnectPeersPaginator(svc, &ec2.DescribeTransitGatewayConnectPeersInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("transit-gateway-attachment-id"),
				Values: []string{attachmentID},
			},
		},
	})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, peer := range page.TransitGatewayConnectPeers {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(peer.TransitGatewayConnectPeerId),
				StringValue(peer.TransitGatewayConnectPeerId),
				"aws_ec2_transit_gateway_connect_peer",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
}

func (g *TransitGatewayGenerator) getTransitGatewayMulticastDomains(svc *ec2.Client) {
	p := ec2.NewDescribeTransitGatewayMulticastDomainsPaginator(svc, &ec2.DescribeTransitGatewayMulticastDomainsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, domain := range page.TransitGatewayMulticastDomains {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(domain.TransitGatewayMulticastDomainId),
				StringValue(domain.TransitGatewayMulticastDomainId),
				"aws_ec2_transit_gateway_multicast_domain",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
}

func (g *TransitGatewayGenerator) getTransitGatewayPeeringAttachments(svc *ec2.Client) {
	p := ec2.NewDescribeTransitGatewayPeeringAttachmentsPaginator(svc, &ec2.DescribeTransitGatewayPeeringAttachmentsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, attachment := range page.TransitGatewayPeeringAttachments {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(attachment.TransitGatewayAttachmentId),
				StringValue(attachment.TransitGatewayAttachmentId),
				"aws_ec2_transit_gateway_peering_attachment",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
}

func (g *TransitGatewayGenerator) getTransitGatewayMeteringPolicies(svc *ec2.Client) {
	input := &ec2.DescribeTransitGatewayMeteringPoliciesInput{}
	for {
		out, err := svc.DescribeTransitGatewayMeteringPolicies(context.TODO(), input)
		if err != nil {
			log.Println(err)
			return
		}
		for _, policy := range out.TransitGatewayMeteringPolicies {
			if policy.TransitGatewayMeteringPolicyId == nil {
				continue
			}
			policyID := *policy.TransitGatewayMeteringPolicyId
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				policyID,
				policyID,
				"aws_ec2_transit_gateway_metering_policy",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
			g.getTransitGatewayMeteringPolicyEntries(svc, policyID)
		}
		if out.NextToken == nil {
			break
		}
		input.NextToken = out.NextToken
	}
}

// aws_ec2_transit_gateway_metering_policy_entry is imported by
// "<policy id>,<rule number>".
func (g *TransitGatewayGenerator) getTransitGatewayMeteringPolicyEntries(svc *ec2.Client, policyID string) {
	out, err := svc.GetTransitGatewayMeteringPolicyEntries(context.TODO(), &ec2.GetTransitGatewayMeteringPolicyEntriesInput{
		TransitGatewayMeteringPolicyId: aws.String(policyID),
	})
	if err != nil {
		log.Println(err)
		return
	}
	for _, entry := range out.TransitGatewayMeteringPolicyEntries {
		if entry.PolicyRuleNumber == nil {
			continue
		}
		id := policyID + "," + *entry.PolicyRuleNumber
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			id,
			id,
			"aws_ec2_transit_gateway_metering_policy_entry",
			"aws",
			awsNetCoreAllowEmptyValues,
		))
	}
}

func (g *TransitGatewayGenerator) getVpnConcentrators(svc *ec2.Client) {
	p := ec2.NewDescribeVpnConcentratorsPaginator(svc, &ec2.DescribeVpnConcentratorsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, concentrator := range page.VpnConcentrators {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(concentrator.VpnConcentratorId),
				StringValue(concentrator.VpnConcentratorId),
				"aws_vpn_concentrator",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
}
