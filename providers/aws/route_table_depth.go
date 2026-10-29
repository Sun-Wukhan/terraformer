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
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// getRouteTableDepth extends RouteTableGenerator with standalone routes
// and the default route table adoption resource. Failures are logged and
// skipped so they never block the base aws_route_table import.
func (g *RouteTableGenerator) getRouteTableDepth(svc *ec2.Client) {
	p := ec2.NewDescribeRouteTablesPaginator(svc, &ec2.DescribeRouteTablesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, table := range page.RouteTables {
			if table.RouteTableId == nil {
				continue
			}
			tableID := *table.RouteTableId
			for _, association := range table.Associations {
				// aws_default_route_table is imported by VPC id.
				if association.Main != nil && *association.Main && table.VpcId != nil {
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						StringValue(table.VpcId),
						StringValue(table.VpcId),
						"aws_default_route_table",
						"aws",
						awsNetCoreAllowEmptyValues,
					))
				}
			}
			for _, route := range table.Routes {
				// Only routes created via CreateRoute map to aws_route;
				// the local route (CreateRouteTable) and propagated
				// routes (EnableVgwRoutePropagation) do not.
				if route.Origin != types.RouteOriginCreateRoute {
					continue
				}
				destination := StringValue(route.DestinationCidrBlock)
				if destination == "" {
					destination = StringValue(route.DestinationIpv6CidrBlock)
				}
				if destination == "" {
					destination = StringValue(route.DestinationPrefixListId)
				}
				if destination == "" {
					continue
				}
				id := tableID + "_" + destination
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id,
					id,
					"aws_route",
					"aws",
					awsNetCoreAllowEmptyValues,
				))
			}
		}
	}
}
