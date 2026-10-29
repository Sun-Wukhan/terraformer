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

// getVpcEndpointDepth extends VpcEndpointGenerator with endpoint services,
// connection notifications, per-endpoint policy/private-DNS resources, the
// route-table/subnet/security-group association resources and
// service-owner-side connection accepters. Failures are logged and skipped
// so one unsupported API never blocks the base aws_vpc_endpoint import.
//
// Notable skips in this family (see the cluster ledger for the full list):
//   - aws_vpc_endpoint_service_allowed_principal: no provider importer.
//   - aws_vpc_endpoint_service_private_dns_verification: provider
//     explicitly does not support import.
func (g *VpcEndpointGenerator) getVpcEndpointDepth(svc *ec2.Client) {
	g.getVpcEndpointServices(svc)
	g.getVpcEndpointConnectionNotifications(svc)
	g.getVpcEndpointRelations(svc)
}

func (g *VpcEndpointGenerator) getVpcEndpointServices(svc *ec2.Client) {
	p := ec2.NewDescribeVpcEndpointServiceConfigurationsPaginator(svc, &ec2.DescribeVpcEndpointServiceConfigurationsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, service := range page.ServiceConfigurations {
			if service.ServiceId == nil {
				continue
			}
			serviceID := *service.ServiceId
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				serviceID,
				serviceID,
				"aws_vpc_endpoint_service",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
			g.getVpcEndpointConnectionAccepters(svc, serviceID)
		}
	}
}

// aws_vpc_endpoint_connection_accepter is imported by
// "<service id>_<endpoint id>" and exists on the service-owner side of an
// endpoint connection.
func (g *VpcEndpointGenerator) getVpcEndpointConnectionAccepters(svc *ec2.Client, serviceID string) {
	p := ec2.NewDescribeVpcEndpointConnectionsPaginator(svc, &ec2.DescribeVpcEndpointConnectionsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("service-id"),
				Values: []string{serviceID},
			},
		},
	})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, connection := range page.VpcEndpointConnections {
			if connection.VpcEndpointId == nil {
				continue
			}
			switch string(connection.VpcEndpointState) {
			case "deleted", "deleting", "expired", "rejected":
				continue
			}
			id := serviceID + "_" + *connection.VpcEndpointId
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				id,
				"aws_vpc_endpoint_connection_accepter",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
}

func (g *VpcEndpointGenerator) getVpcEndpointConnectionNotifications(svc *ec2.Client) {
	p := ec2.NewDescribeVpcEndpointConnectionNotificationsPaginator(svc, &ec2.DescribeVpcEndpointConnectionNotificationsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, notification := range page.ConnectionNotificationSet {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(notification.ConnectionNotificationId),
				StringValue(notification.ConnectionNotificationId),
				"aws_vpc_endpoint_connection_notification",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
}

// The policy, private-DNS and association resources are all keyed by the
// endpoint id (associations by "<endpoint id>/<target id>") and are derived
// from the endpoint object itself.
func (g *VpcEndpointGenerator) getVpcEndpointRelations(svc *ec2.Client) {
	p := ec2.NewDescribeVpcEndpointsPaginator(svc, &ec2.DescribeVpcEndpointsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, endpoint := range page.VpcEndpoints {
			if endpoint.VpcEndpointId == nil {
				continue
			}
			endpointID := *endpoint.VpcEndpointId
			if StringValue(endpoint.PolicyDocument) != "" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					endpointID,
					endpointID,
					"aws_vpc_endpoint_policy",
					"aws",
					awsNetCoreAllowEmptyValues,
				))
			}
			if endpoint.VpcEndpointType == types.VpcEndpointTypeInterface {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					endpointID,
					endpointID,
					"aws_vpc_endpoint_private_dns",
					"aws",
					awsNetCoreAllowEmptyValues,
				))
			}
			for _, routeTableID := range endpoint.RouteTableIds {
				id := endpointID + "/" + routeTableID
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id,
					id,
					"aws_vpc_endpoint_route_table_association",
					"aws",
					awsNetCoreAllowEmptyValues,
				))
			}
			for _, subnetID := range endpoint.SubnetIds {
				id := endpointID + "/" + subnetID
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id,
					id,
					"aws_vpc_endpoint_subnet_association",
					"aws",
					awsNetCoreAllowEmptyValues,
				))
			}
			for _, group := range endpoint.Groups {
				if group.GroupId == nil {
					continue
				}
				id := endpointID + "/" + *group.GroupId
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id,
					id,
					"aws_vpc_endpoint_security_group_association",
					"aws",
					awsNetCoreAllowEmptyValues,
				))
			}
		}
	}
}
