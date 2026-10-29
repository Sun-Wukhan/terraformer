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

// LocalGatewayGenerator covers EC2 local gateways (Outposts): route
// tables, routes, and the route tables' VPC and virtual interface group
// associations. Sub-resource failures are logged and skipped.
type LocalGatewayGenerator struct {
	AWSService
}

func (g *LocalGatewayGenerator) InitResources() error {
	config, err := g.generateConfig()
	if err != nil {
		return err
	}
	svc := ec2.NewFromConfig(config)
	g.getLocalGatewayRouteTables(svc)
	g.getLocalGatewayRouteTableVpcAssociations(svc)
	g.getLocalGatewayRouteTableVirtualInterfaceGroupAssociations(svc)
	return nil
}

func (g *LocalGatewayGenerator) getLocalGatewayRouteTables(svc *ec2.Client) {
	p := ec2.NewDescribeLocalGatewayRouteTablesPaginator(svc, &ec2.DescribeLocalGatewayRouteTablesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, routeTable := range page.LocalGatewayRouteTables {
			if routeTable.LocalGatewayRouteTableId == nil {
				continue
			}
			routeTableID := *routeTable.LocalGatewayRouteTableId
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				routeTableID,
				routeTableID,
				"aws_ec2_local_gateway_route_table",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
			g.getLocalGatewayRoutes(svc, routeTableID)
		}
	}
}

// aws_ec2_local_gateway_route is imported by
// "<route table id>_<destination cidr>".
func (g *LocalGatewayGenerator) getLocalGatewayRoutes(svc *ec2.Client, routeTableID string) {
	p := ec2.NewSearchLocalGatewayRoutesPaginator(svc, &ec2.SearchLocalGatewayRoutesInput{
		LocalGatewayRouteTableId: aws.String(routeTableID),
	})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, route := range page.Routes {
			if route.DestinationCidrBlock == nil {
				continue
			}
			id := routeTableID + "_" + *route.DestinationCidrBlock
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				id,
				"aws_ec2_local_gateway_route",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
}

func (g *LocalGatewayGenerator) getLocalGatewayRouteTableVpcAssociations(svc *ec2.Client) {
	p := ec2.NewDescribeLocalGatewayRouteTableVpcAssociationsPaginator(svc, &ec2.DescribeLocalGatewayRouteTableVpcAssociationsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, association := range page.LocalGatewayRouteTableVpcAssociations {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(association.LocalGatewayRouteTableVpcAssociationId),
				StringValue(association.LocalGatewayRouteTableVpcAssociationId),
				"aws_ec2_local_gateway_route_table_vpc_association",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
}

func (g *LocalGatewayGenerator) getLocalGatewayRouteTableVirtualInterfaceGroupAssociations(svc *ec2.Client) {
	p := ec2.NewDescribeLocalGatewayRouteTableVirtualInterfaceGroupAssociationsPaginator(svc, &ec2.DescribeLocalGatewayRouteTableVirtualInterfaceGroupAssociationsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, association := range page.LocalGatewayRouteTableVirtualInterfaceGroupAssociations {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(association.LocalGatewayRouteTableVirtualInterfaceGroupAssociationId),
				StringValue(association.LocalGatewayRouteTableVirtualInterfaceGroupAssociationId),
				"aws_ec2_local_gateway_route_table_virtual_interface_group_association",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
}
