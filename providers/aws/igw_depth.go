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
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// getIgwDepth extends IgwGenerator with egress-only internet gateways and
// carrier gateways. Failures are logged and skipped so they never block
// the base aws_internet_gateway import.
//
// Notable skips in this family: aws_vpn_gateway_attachment,
// aws_vpn_gateway_route_propagation and aws_vpn_connection_route have no
// provider importer (the VPN gateway itself is covered by
// VpnGatewayGenerator).
func (g *IgwGenerator) getIgwDepth(svc *ec2.Client) {
	g.getCarrierGateways(svc)
	p := ec2.NewDescribeEgressOnlyInternetGatewaysPaginator(svc, &ec2.DescribeEgressOnlyInternetGatewaysInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, gateway := range page.EgressOnlyInternetGateways {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(gateway.EgressOnlyInternetGatewayId),
				StringValue(gateway.EgressOnlyInternetGatewayId),
				"aws_egress_only_internet_gateway",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
}

// aws_ec2_carrier_gateway is imported by the carrier gateway's ID.
func (g *IgwGenerator) getCarrierGateways(svc *ec2.Client) {
	p := ec2.NewDescribeCarrierGatewaysPaginator(svc, &ec2.DescribeCarrierGatewaysInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, gateway := range page.CarrierGateways {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(gateway.CarrierGatewayId),
				StringValue(gateway.CarrierGatewayId),
				"aws_ec2_carrier_gateway",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
}

// Depth enumeration for internet gateways, called from
// IgwGenerator.InitResources: one aws_internet_gateway_attachment
// per VPC attachment of every internet gateway (attached or not —
// unlike aws_internet_gateway itself, which the base generator only
// emits for attached gateways). Provider ID format:
// internet_gateway_id:vpc_id.
func (g *IgwGenerator) addDepthResources(svc *ec2.Client) error {
	p := ec2.NewDescribeInternetGatewaysPaginator(svc, &ec2.DescribeInternetGatewaysInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, igw := range page.InternetGateways {
			igwID := StringValue(igw.InternetGatewayId)
			for _, attachment := range igw.Attachments {
				vpcID := StringValue(attachment.VpcId)
				g.Resources = append(g.Resources, terraformutils.NewResource(
					igwID+":"+vpcID,
					igwID,
					"aws_internet_gateway_attachment",
					"aws",
					map[string]string{"internet_gateway_id": igwID, "vpc_id": vpcID},
					IgwAllowEmptyValues,
					map[string]interface{}{},
				))
			}
		}
	}
	return nil
}
