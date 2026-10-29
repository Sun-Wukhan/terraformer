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
)

// ClientVpnGenerator covers AWS Client VPN: endpoints, authorization
// rules, target network associations and routes. Sub-resource failures are
// logged and skipped.
type ClientVpnGenerator struct {
	AWSService
}

func (g *ClientVpnGenerator) InitResources() error {
	config, err := g.generateConfig()
	if err != nil {
		return err
	}
	svc := ec2.NewFromConfig(config)
	p := ec2.NewDescribeClientVpnEndpointsPaginator(svc, &ec2.DescribeClientVpnEndpointsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, endpoint := range page.ClientVpnEndpoints {
			if endpoint.ClientVpnEndpointId == nil {
				continue
			}
			endpointID := *endpoint.ClientVpnEndpointId
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				endpointID,
				endpointID,
				"aws_ec2_client_vpn_endpoint",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
			g.getClientVpnAuthorizationRules(svc, endpointID)
			g.getClientVpnNetworkAssociations(svc, endpointID)
			g.getClientVpnRoutes(svc, endpointID)
		}
	}
	return nil
}

// aws_ec2_client_vpn_authorization_rule is imported by
// "<endpoint id>,<destination cidr>[,<group id>]".
func (g *ClientVpnGenerator) getClientVpnAuthorizationRules(svc *ec2.Client, endpointID string) {
	p := ec2.NewDescribeClientVpnAuthorizationRulesPaginator(svc, &ec2.DescribeClientVpnAuthorizationRulesInput{
		ClientVpnEndpointId: aws.String(endpointID),
	})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, rule := range page.AuthorizationRules {
			if rule.DestinationCidr == nil {
				continue
			}
			id := endpointID + "," + *rule.DestinationCidr
			if rule.GroupId != nil && *rule.GroupId != "" {
				id += "," + *rule.GroupId
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				id,
				"aws_ec2_client_vpn_authorization_rule",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
}

// aws_ec2_client_vpn_network_association is imported by
// "<endpoint id>,<association id>".
func (g *ClientVpnGenerator) getClientVpnNetworkAssociations(svc *ec2.Client, endpointID string) {
	p := ec2.NewDescribeClientVpnTargetNetworksPaginator(svc, &ec2.DescribeClientVpnTargetNetworksInput{
		ClientVpnEndpointId: aws.String(endpointID),
	})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, targetNetwork := range page.ClientVpnTargetNetworks {
			if targetNetwork.AssociationId == nil {
				continue
			}
			id := endpointID + "," + *targetNetwork.AssociationId
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				id,
				"aws_ec2_client_vpn_network_association",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
}

// aws_ec2_client_vpn_route is imported by
// "<endpoint id>,<subnet id>,<destination cidr>". Only routes with origin
// "add-route" are managed routes; association/propagated routes are not.
func (g *ClientVpnGenerator) getClientVpnRoutes(svc *ec2.Client, endpointID string) {
	p := ec2.NewDescribeClientVpnRoutesPaginator(svc, &ec2.DescribeClientVpnRoutesInput{
		ClientVpnEndpointId: aws.String(endpointID),
	})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, route := range page.Routes {
			if StringValue(route.Origin) != "add-route" {
				continue
			}
			if route.TargetSubnet == nil || route.DestinationCidr == nil {
				continue
			}
			id := endpointID + "," + *route.TargetSubnet + "," + *route.DestinationCidr
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				id,
				"aws_ec2_client_vpn_route",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
}
