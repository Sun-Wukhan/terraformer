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

// getVpcPeeringDepth extends VpcPeeringConnectionGenerator with the
// per-peering options resource and, for cross-account peerings where this
// account is purely the accepter, the accepter resource. Both are imported
// by the peering connection id. Failures are logged and skipped so they
// never block the base aws_vpc_peering_connection import.
func (g *VpcPeeringConnectionGenerator) getVpcPeeringDepth(svc *ec2.Client, config aws.Config) {
	accountID, err := g.getAccountNumber(config)
	if err != nil {
		log.Println(err)
	}
	p := ec2.NewDescribeVpcPeeringConnectionsPaginator(svc, &ec2.DescribeVpcPeeringConnectionsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, connection := range page.VpcPeeringConnections {
			if connection.VpcPeeringConnectionId == nil {
				continue
			}
			connectionID := *connection.VpcPeeringConnectionId
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				connectionID,
				connectionID,
				"aws_vpc_peering_connection_options",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
			// The accepter resource only applies when another account
			// requested the peering; same-account peerings are fully
			// described by aws_vpc_peering_connection itself.
			if accountID != nil && connection.RequesterVpcInfo != nil &&
				connection.RequesterVpcInfo.OwnerId != nil &&
				*connection.RequesterVpcInfo.OwnerId != *accountID {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					connectionID,
					connectionID,
					"aws_vpc_peering_connection_accepter",
					"aws",
					awsNetCoreAllowEmptyValues,
				))
			}
		}
	}
}
