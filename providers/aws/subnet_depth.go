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

// getSubnetDepth extends SubnetGenerator with the default subnet adoption
// resource and subnet CIDR reservations. Failures are logged and skipped
// so they never block the base aws_subnet import.
func (g *SubnetGenerator) getSubnetDepth(svc *ec2.Client) {
	p := ec2.NewDescribeSubnetsPaginator(svc, &ec2.DescribeSubnetsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, subnet := range page.Subnets {
			if subnet.SubnetId == nil {
				continue
			}
			subnetID := *subnet.SubnetId
			if subnet.DefaultForAz != nil && *subnet.DefaultForAz {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					subnetID,
					subnetID,
					"aws_default_subnet",
					"aws",
					awsNetCoreAllowEmptyValues,
				))
			}
			g.getSubnetCidrReservations(svc, subnetID)
		}
	}
}

// aws_ec2_subnet_cidr_reservation is imported by
// "<subnet id>:<reservation id>".
func (g *SubnetGenerator) getSubnetCidrReservations(svc *ec2.Client, subnetID string) {
	input := &ec2.GetSubnetCidrReservationsInput{
		SubnetId: aws.String(subnetID),
	}
	for {
		out, err := svc.GetSubnetCidrReservations(context.TODO(), input)
		if err != nil {
			log.Println(err)
			return
		}
		reservations := append(append([]types.SubnetCidrReservation{}, out.SubnetIpv4CidrReservations...), out.SubnetIpv6CidrReservations...)
		for _, reservation := range reservations {
			if reservation.SubnetCidrReservationId == nil {
				continue
			}
			id := subnetID + ":" + *reservation.SubnetCidrReservationId
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				id,
				"aws_ec2_subnet_cidr_reservation",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
		if out.NextToken == nil {
			break
		}
		input.NextToken = out.NextToken
	}
}
