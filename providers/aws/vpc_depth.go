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
	"github.com/aws/aws-sdk-go-v2/service/organizations"
)

// getVpcDepth extends VpcGenerator with DHCP options, CIDR block
// associations, the default-VPC adoption resources, IPAM, VPC encryption
// controls, block-public-access exclusions and VPC route servers.
// Failures are logged and skipped so one unsupported API in a region never
// blocks the base aws_vpc import.
//
// Notable skips in this family (see the cluster ledger for the full list):
//   - aws_vpc_block_public_access_options: singleton account/region
//     settings object.
//   - aws_vpc_ipam_preview_next_cidr: computed preview, no provider
//     importer.
//   - aws_vpc_endpoint_service_allowed_principal and
//     aws_vpc_network_performance_metric_subscription: no provider
//     importer.
func (g *VpcGenerator) getVpcDepth(svc *ec2.Client, config aws.Config) {
	g.getDhcpOptions(svc)
	g.getVpcScopedResources(svc)
	g.getIpam(svc, config)
	g.getVpcEncryptionControls(svc)
	g.getVpcBlockPublicAccessExclusions(svc)
	g.getRouteServers(svc)
}

func (g *VpcGenerator) getDhcpOptions(svc *ec2.Client) {
	p := ec2.NewDescribeDhcpOptionsPaginator(svc, &ec2.DescribeDhcpOptionsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, options := range page.DhcpOptions {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(options.DhcpOptionsId),
				StringValue(options.DhcpOptionsId),
				"aws_vpc_dhcp_options",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
}

// getVpcScopedResources derives the per-VPC resources from DescribeVpcs:
// DHCP options associations (imported by VPC id), the default VPC and its
// default DHCP options set, and the IPv4/IPv6 CIDR block associations.
func (g *VpcGenerator) getVpcScopedResources(svc *ec2.Client) {
	p := ec2.NewDescribeVpcsPaginator(svc, &ec2.DescribeVpcsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, vpc := range page.Vpcs {
			if vpc.VpcId == nil {
				continue
			}
			vpcID := *vpc.VpcId
			if vpc.DhcpOptionsId != nil {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					vpcID,
					vpcID,
					"aws_vpc_dhcp_options_association",
					"aws",
					awsNetCoreAllowEmptyValues,
				))
			}
			if vpc.IsDefault != nil && *vpc.IsDefault {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					vpcID,
					vpcID,
					"aws_default_vpc",
					"aws",
					awsNetCoreAllowEmptyValues,
				))
				if vpc.DhcpOptionsId != nil {
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						StringValue(vpc.DhcpOptionsId),
						StringValue(vpc.DhcpOptionsId),
						"aws_default_vpc_dhcp_options",
						"aws",
						awsNetCoreAllowEmptyValues,
					))
				}
			}
			for _, association := range vpc.CidrBlockAssociationSet {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					StringValue(association.AssociationId),
					StringValue(association.AssociationId),
					"aws_vpc_ipv4_cidr_block_association",
					"aws",
					awsNetCoreAllowEmptyValues,
				))
			}
			for _, association := range vpc.Ipv6CidrBlockAssociationSet {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					StringValue(association.AssociationId),
					StringValue(association.AssociationId),
					"aws_vpc_ipv6_cidr_block_association",
					"aws",
					awsNetCoreAllowEmptyValues,
				))
			}
		}
	}
}

func (g *VpcGenerator) getIpam(svc *ec2.Client, config aws.Config) {
	ipamPaginator := ec2.NewDescribeIpamsPaginator(svc, &ec2.DescribeIpamsInput{})
	for ipamPaginator.HasMorePages() {
		page, err := ipamPaginator.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			break
		}
		for _, ipam := range page.Ipams {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(ipam.IpamId),
				StringValue(ipam.IpamId),
				"aws_vpc_ipam",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
	scopePaginator := ec2.NewDescribeIpamScopesPaginator(svc, &ec2.DescribeIpamScopesInput{})
	for scopePaginator.HasMorePages() {
		page, err := scopePaginator.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			break
		}
		for _, scope := range page.IpamScopes {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(scope.IpamScopeId),
				StringValue(scope.IpamScopeId),
				"aws_vpc_ipam_scope",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
	poolPaginator := ec2.NewDescribeIpamPoolsPaginator(svc, &ec2.DescribeIpamPoolsInput{})
	for poolPaginator.HasMorePages() {
		page, err := poolPaginator.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			break
		}
		for _, pool := range page.IpamPools {
			if pool.IpamPoolId == nil {
				continue
			}
			poolID := *pool.IpamPoolId
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				poolID,
				poolID,
				"aws_vpc_ipam_pool",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
			g.getIpamPoolCidrs(svc, poolID)
			g.getIpamPoolAllocations(svc, poolID)
		}
	}
	discoveryPaginator := ec2.NewDescribeIpamResourceDiscoveriesPaginator(svc, &ec2.DescribeIpamResourceDiscoveriesInput{})
	for discoveryPaginator.HasMorePages() {
		page, err := discoveryPaginator.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			break
		}
		for _, discovery := range page.IpamResourceDiscoveries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(discovery.IpamResourceDiscoveryId),
				StringValue(discovery.IpamResourceDiscoveryId),
				"aws_vpc_ipam_resource_discovery",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
	associationPaginator := ec2.NewDescribeIpamResourceDiscoveryAssociationsPaginator(svc, &ec2.DescribeIpamResourceDiscoveryAssociationsInput{})
	for associationPaginator.HasMorePages() {
		page, err := associationPaginator.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			break
		}
		for _, association := range page.IpamResourceDiscoveryAssociations {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(association.IpamResourceDiscoveryAssociationId),
				StringValue(association.IpamResourceDiscoveryAssociationId),
				"aws_vpc_ipam_resource_discovery_association",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
	g.getIpamOrganizationAdminAccounts(config)
}

// aws_vpc_ipam_pool_cidr is imported by "<cidr>_<pool id>" (not by the
// pool CIDR id).
func (g *VpcGenerator) getIpamPoolCidrs(svc *ec2.Client, poolID string) {
	p := ec2.NewGetIpamPoolCidrsPaginator(svc, &ec2.GetIpamPoolCidrsInput{
		IpamPoolId: aws.String(poolID),
	})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, poolCidr := range page.IpamPoolCidrs {
			if poolCidr.Cidr == nil {
				continue
			}
			id := *poolCidr.Cidr + "_" + poolID
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				id,
				"aws_vpc_ipam_pool_cidr",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
}

func (g *VpcGenerator) getIpamPoolAllocations(svc *ec2.Client, poolID string) {
	p := ec2.NewGetIpamPoolAllocationsPaginator(svc, &ec2.GetIpamPoolAllocationsInput{
		IpamPoolId: aws.String(poolID),
	})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, allocation := range page.IpamPoolAllocations {
			if allocation.IpamPoolAllocationId == nil {
				continue
			}
			id := *allocation.IpamPoolAllocationId + "_" + poolID
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				id,
				"aws_vpc_ipam_pool_cidr_allocation",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
}

// The IPAM delegated administrator is recorded in AWS Organizations under
// the ipam.amazonaws.com service principal; the resource id is the account
// id.
func (g *VpcGenerator) getIpamOrganizationAdminAccounts(config aws.Config) {
	orgClient := organizations.NewFromConfig(config)
	p := organizations.NewListDelegatedAdministratorsPaginator(orgClient, &organizations.ListDelegatedAdministratorsInput{
		ServicePrincipal: aws.String("ipam.amazonaws.com"),
	})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, admin := range page.DelegatedAdministrators {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(admin.Id),
				StringValue(admin.Id),
				"aws_vpc_ipam_organization_admin_account",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
}

func (g *VpcGenerator) getVpcEncryptionControls(svc *ec2.Client) {
	input := &ec2.DescribeVpcEncryptionControlsInput{}
	for {
		out, err := svc.DescribeVpcEncryptionControls(context.TODO(), input)
		if err != nil {
			log.Println(err)
			return
		}
		for _, control := range out.VpcEncryptionControls {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(control.VpcEncryptionControlId),
				StringValue(control.VpcEncryptionControlId),
				"aws_vpc_encryption_control",
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

func (g *VpcGenerator) getVpcBlockPublicAccessExclusions(svc *ec2.Client) {
	input := &ec2.DescribeVpcBlockPublicAccessExclusionsInput{}
	for {
		out, err := svc.DescribeVpcBlockPublicAccessExclusions(context.TODO(), input)
		if err != nil {
			log.Println(err)
			return
		}
		for _, exclusion := range out.VpcBlockPublicAccessExclusions {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(exclusion.ExclusionId),
				StringValue(exclusion.ExclusionId),
				"aws_vpc_block_public_access_exclusion",
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

func (g *VpcGenerator) getRouteServers(svc *ec2.Client) {
	serverPaginator := ec2.NewDescribeRouteServersPaginator(svc, &ec2.DescribeRouteServersInput{})
	for serverPaginator.HasMorePages() {
		page, err := serverPaginator.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, server := range page.RouteServers {
			if server.RouteServerId == nil {
				continue
			}
			serverID := *server.RouteServerId
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				serverID,
				serverID,
				"aws_vpc_route_server",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
			g.getRouteServerAssociations(svc, serverID)
			g.getRouteServerPropagations(svc, serverID)
		}
	}
	endpointPaginator := ec2.NewDescribeRouteServerEndpointsPaginator(svc, &ec2.DescribeRouteServerEndpointsInput{})
	for endpointPaginator.HasMorePages() {
		page, err := endpointPaginator.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			break
		}
		for _, endpoint := range page.RouteServerEndpoints {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(endpoint.RouteServerEndpointId),
				StringValue(endpoint.RouteServerEndpointId),
				"aws_vpc_route_server_endpoint",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
	peerPaginator := ec2.NewDescribeRouteServerPeersPaginator(svc, &ec2.DescribeRouteServerPeersInput{})
	for peerPaginator.HasMorePages() {
		page, err := peerPaginator.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			break
		}
		for _, peer := range page.RouteServerPeers {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(peer.RouteServerPeerId),
				StringValue(peer.RouteServerPeerId),
				"aws_vpc_route_server_peer",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
}

// aws_vpc_route_server_vpc_association is imported by "<route server
// id>,<vpc id>".
func (g *VpcGenerator) getRouteServerAssociations(svc *ec2.Client, serverID string) {
	out, err := svc.GetRouteServerAssociations(context.TODO(), &ec2.GetRouteServerAssociationsInput{
		RouteServerId: aws.String(serverID),
	})
	if err != nil {
		log.Println(err)
		return
	}
	for _, association := range out.RouteServerAssociations {
		if association.VpcId == nil {
			continue
		}
		id := serverID + "," + *association.VpcId
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			id,
			id,
			"aws_vpc_route_server_vpc_association",
			"aws",
			awsNetCoreAllowEmptyValues,
		))
	}
}

// aws_vpc_route_server_propagation is imported by "<route server
// id>,<route table id>".
func (g *VpcGenerator) getRouteServerPropagations(svc *ec2.Client, serverID string) {
	out, err := svc.GetRouteServerPropagations(context.TODO(), &ec2.GetRouteServerPropagationsInput{
		RouteServerId: aws.String(serverID),
	})
	if err != nil {
		log.Println(err)
		return
	}
	for _, propagation := range out.RouteServerPropagations {
		if propagation.RouteTableId == nil {
			continue
		}
		id := serverID + "," + *propagation.RouteTableId
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			id,
			id,
			"aws_vpc_route_server_propagation",
			"aws",
			awsNetCoreAllowEmptyValues,
		))
	}
}
