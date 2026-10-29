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

// getNatGatewayDepth extends NatGatewayGenerator with the NAT gateway EIP
// association resources, derived from the gateway's addresses. Failures
// are logged and skipped so they never block the base aws_nat_gateway
// import.
func (g *NatGatewayGenerator) getNatGatewayDepth(svc *ec2.Client) {
	p := ec2.NewDescribeNatGatewaysPaginator(svc, &ec2.DescribeNatGatewaysInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, gateway := range page.NatGateways {
			if gateway.NatGatewayId == nil {
				continue
			}
			gatewayID := *gateway.NatGatewayId
			// aws_nat_gateway_eip_association is imported by
			// "<nat gateway id>,<allocation id>".
			for _, address := range gateway.NatGatewayAddresses {
				if address.AllocationId == nil || address.AssociationId == nil {
					continue
				}
				id := gatewayID + "," + *address.AllocationId
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id,
					id,
					"aws_nat_gateway_eip_association",
					"aws",
					awsNetCoreAllowEmptyValues,
				))
			}
		}
	}
}
