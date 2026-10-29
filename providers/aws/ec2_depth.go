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
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// awsNetCoreAllowEmptyValues is shared by the net-core depth generators.
var awsNetCoreAllowEmptyValues = []string{"tags."}

// getEc2Depth extends Ec2Generator with the remaining importable EC2-core
// resource types. Failures are logged and skipped so one unsupported API in
// a region never blocks the base aws_instance import.
//
// Notable skips in this family (see the cluster ledger for the full list):
//   - aws_ec2_tag: tags are attributes of their parent resource, already
//     imported inline; standalone enumeration would duplicate every tag.
//   - aws_ec2_instance_state: overlay on aws_instance with no independent
//     identity.
//   - aws_ec2_serial_console_access, aws_ec2_image_block_public_access,
//     aws_ec2_allowed_images_settings: singleton account/region settings.
//   - aws_ec2_instance_metadata_defaults: provider does not support import.
//   - aws_ec2_capacity_block_reservation, aws_ami_copy, aws_ami_from_instance,
//     aws_spot_instance_request, aws_snapshot_create_volume_permission:
//     no provider importer.
func (g *Ec2Generator) getEc2Depth(svc *ec2.Client) {
	g.getCapacityReservations(svc)
	g.getAvailabilityZoneGroups(svc)
	g.getDefaultCreditSpecifications(svc)
	g.getFleets(svc)
	g.getHosts(svc)
	g.getInstanceConnectEndpoints(svc)
	g.getManagedPrefixLists(svc)
	g.getNetworkInsights(svc)
	g.getTrafficMirror(svc)
	g.getAmis(svc)
	g.getKeyPairs(svc)
	g.getPlacementGroups(svc)
	g.getSpotFleetRequests(svc)
	g.getSpotDatafeedSubscription(svc)
	g.getSecondaryNetworks(svc)
}

func (g *Ec2Generator) getCapacityReservations(svc *ec2.Client) {
	p := ec2.NewDescribeCapacityReservationsPaginator(svc, &ec2.DescribeCapacityReservationsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, reservation := range page.CapacityReservations {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(reservation.CapacityReservationId),
				StringValue(reservation.CapacityReservationId),
				"aws_ec2_capacity_reservation",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
}

// aws_ec2_availability_zone_group manages opt-in for Local/Wavelength Zone
// groups; plain availability zones share the region group and are not
// emitted.
func (g *Ec2Generator) getAvailabilityZoneGroups(svc *ec2.Client) {
	out, err := svc.DescribeAvailabilityZones(context.TODO(), &ec2.DescribeAvailabilityZonesInput{})
	if err != nil {
		log.Println(err)
		return
	}
	seen := map[string]bool{}
	for _, zone := range out.AvailabilityZones {
		if zone.GroupName == nil || (zone.ZoneType != nil && *zone.ZoneType == "availability-zone") {
			continue
		}
		group := *zone.GroupName
		if seen[group] {
			continue
		}
		seen[group] = true
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			group,
			group,
			"aws_ec2_availability_zone_group",
			"aws",
			awsNetCoreAllowEmptyValues,
		))
	}
}

// aws_ec2_default_credit_specification is imported by instance family.
// The default credit specification itself is only readable per family
// (GetDefaultCreditSpecification), so the family set is derived from
// DescribeInstanceTypes: the family is the instance type name's prefix
// before the first dot.
func (g *Ec2Generator) getDefaultCreditSpecifications(svc *ec2.Client) {
	seen := map[string]bool{}
	p := ec2.NewDescribeInstanceTypesPaginator(svc, &ec2.DescribeInstanceTypesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, instanceType := range page.InstanceTypes {
			family, _, found := strings.Cut(string(instanceType.InstanceType), ".")
			if !found || family == "" || seen[family] {
				continue
			}
			seen[family] = true
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				family,
				family,
				"aws_ec2_default_credit_specification",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
}

func (g *Ec2Generator) getFleets(svc *ec2.Client) {
	p := ec2.NewDescribeFleetsPaginator(svc, &ec2.DescribeFleetsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, fleet := range page.Fleets {
			state := string(fleet.FleetState)
			if strings.HasPrefix(state, "deleted") || state == "failed" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(fleet.FleetId),
				StringValue(fleet.FleetId),
				"aws_ec2_fleet",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
}

func (g *Ec2Generator) getHosts(svc *ec2.Client) {
	p := ec2.NewDescribeHostsPaginator(svc, &ec2.DescribeHostsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, host := range page.Hosts {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(host.HostId),
				StringValue(host.HostId),
				"aws_ec2_host",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
}

func (g *Ec2Generator) getInstanceConnectEndpoints(svc *ec2.Client) {
	p := ec2.NewDescribeInstanceConnectEndpointsPaginator(svc, &ec2.DescribeInstanceConnectEndpointsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, endpoint := range page.InstanceConnectEndpoints {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(endpoint.InstanceConnectEndpointId),
				StringValue(endpoint.InstanceConnectEndpointId),
				"aws_ec2_instance_connect_endpoint",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
}

// AWS-managed prefix lists (OwnerId "AWS") cannot be managed as
// aws_ec2_managed_prefix_list resources and are skipped.
func (g *Ec2Generator) getManagedPrefixLists(svc *ec2.Client) {
	p := ec2.NewDescribeManagedPrefixListsPaginator(svc, &ec2.DescribeManagedPrefixListsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, prefixList := range page.PrefixLists {
			if StringValue(prefixList.OwnerId) == "AWS" || prefixList.PrefixListId == nil {
				continue
			}
			listID := *prefixList.PrefixListId
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				listID,
				listID,
				"aws_ec2_managed_prefix_list",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
			ep := ec2.NewGetManagedPrefixListEntriesPaginator(svc, &ec2.GetManagedPrefixListEntriesInput{
				PrefixListId: prefixList.PrefixListId,
			})
			for ep.HasMorePages() {
				entriesPage, err := ep.NextPage(context.TODO())
				if err != nil {
					log.Println(err)
					break
				}
				for _, entry := range entriesPage.Entries {
					if entry.Cidr == nil {
						continue
					}
					id := listID + "," + *entry.Cidr
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						id,
						id,
						"aws_ec2_managed_prefix_list_entry",
						"aws",
						awsNetCoreAllowEmptyValues,
					))
				}
			}
		}
	}
}

func (g *Ec2Generator) getNetworkInsights(svc *ec2.Client) {
	scopePaginator := ec2.NewDescribeNetworkInsightsAccessScopesPaginator(svc, &ec2.DescribeNetworkInsightsAccessScopesInput{})
	for scopePaginator.HasMorePages() {
		page, err := scopePaginator.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			break
		}
		for _, scope := range page.NetworkInsightsAccessScopes {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(scope.NetworkInsightsAccessScopeId),
				StringValue(scope.NetworkInsightsAccessScopeId),
				"aws_ec2_network_insights_access_scope",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
	pathPaginator := ec2.NewDescribeNetworkInsightsPathsPaginator(svc, &ec2.DescribeNetworkInsightsPathsInput{})
	for pathPaginator.HasMorePages() {
		page, err := pathPaginator.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			break
		}
		for _, path := range page.NetworkInsightsPaths {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(path.NetworkInsightsPathId),
				StringValue(path.NetworkInsightsPathId),
				"aws_ec2_network_insights_path",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
	analysisPaginator := ec2.NewDescribeNetworkInsightsAnalysesPaginator(svc, &ec2.DescribeNetworkInsightsAnalysesInput{})
	for analysisPaginator.HasMorePages() {
		page, err := analysisPaginator.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			break
		}
		for _, analysis := range page.NetworkInsightsAnalyses {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(analysis.NetworkInsightsAnalysisId),
				StringValue(analysis.NetworkInsightsAnalysisId),
				"aws_ec2_network_insights_analysis",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
}

func (g *Ec2Generator) getTrafficMirror(svc *ec2.Client) {
	filterPaginator := ec2.NewDescribeTrafficMirrorFiltersPaginator(svc, &ec2.DescribeTrafficMirrorFiltersInput{})
	for filterPaginator.HasMorePages() {
		page, err := filterPaginator.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			break
		}
		for _, filter := range page.TrafficMirrorFilters {
			if filter.TrafficMirrorFilterId == nil {
				continue
			}
			filterID := *filter.TrafficMirrorFilterId
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				filterID,
				filterID,
				"aws_ec2_traffic_mirror_filter",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
			rules := append(append([]types.TrafficMirrorFilterRule{}, filter.IngressFilterRules...), filter.EgressFilterRules...)
			for _, rule := range rules {
				if rule.TrafficMirrorFilterRuleId == nil {
					continue
				}
				id := filterID + ":" + *rule.TrafficMirrorFilterRuleId
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id,
					id,
					"aws_ec2_traffic_mirror_filter_rule",
					"aws",
					awsNetCoreAllowEmptyValues,
				))
			}
		}
	}
	sessionPaginator := ec2.NewDescribeTrafficMirrorSessionsPaginator(svc, &ec2.DescribeTrafficMirrorSessionsInput{})
	for sessionPaginator.HasMorePages() {
		page, err := sessionPaginator.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			break
		}
		for _, session := range page.TrafficMirrorSessions {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(session.TrafficMirrorSessionId),
				StringValue(session.TrafficMirrorSessionId),
				"aws_ec2_traffic_mirror_session",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
	targetPaginator := ec2.NewDescribeTrafficMirrorTargetsPaginator(svc, &ec2.DescribeTrafficMirrorTargetsInput{})
	for targetPaginator.HasMorePages() {
		page, err := targetPaginator.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			break
		}
		for _, target := range page.TrafficMirrorTargets {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(target.TrafficMirrorTargetId),
				StringValue(target.TrafficMirrorTargetId),
				"aws_ec2_traffic_mirror_target",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
}

// Only self-owned images are emitted; public/shared images are not the
// account's aws_ami resources. Launch permissions are read per image.
func (g *Ec2Generator) getAmis(svc *ec2.Client) {
	p := ec2.NewDescribeImagesPaginator(svc, &ec2.DescribeImagesInput{
		Owners: []string{"self"},
	})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, image := range page.Images {
			if image.ImageId == nil {
				continue
			}
			imageID := *image.ImageId
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				imageID,
				imageID,
				"aws_ami",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
			attr, err := svc.DescribeImageAttribute(context.TODO(), &ec2.DescribeImageAttributeInput{
				Attribute: types.ImageAttributeNameLaunchPermission,
				ImageId:   image.ImageId,
			})
			if err != nil {
				log.Println(err)
				continue
			}
			for _, permission := range attr.LaunchPermissions {
				principal := ""
				switch {
				case permission.UserId != nil:
					principal = *permission.UserId
				case permission.Group != "":
					principal = string(permission.Group)
				case permission.OrganizationArn != nil:
					principal = *permission.OrganizationArn
				case permission.OrganizationalUnitArn != nil:
					principal = *permission.OrganizationalUnitArn
				}
				if principal == "" {
					continue
				}
				id := principal + "/" + imageID
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id,
					id,
					"aws_ami_launch_permission",
					"aws",
					awsNetCoreAllowEmptyValues,
				))
			}
		}
	}
}

func (g *Ec2Generator) getKeyPairs(svc *ec2.Client) {
	out, err := svc.DescribeKeyPairs(context.TODO(), &ec2.DescribeKeyPairsInput{})
	if err != nil {
		log.Println(err)
		return
	}
	for _, keyPair := range out.KeyPairs {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			StringValue(keyPair.KeyName),
			StringValue(keyPair.KeyName),
			"aws_key_pair",
			"aws",
			awsNetCoreAllowEmptyValues,
		))
	}
}

func (g *Ec2Generator) getPlacementGroups(svc *ec2.Client) {
	out, err := svc.DescribePlacementGroups(context.TODO(), &ec2.DescribePlacementGroupsInput{})
	if err != nil {
		log.Println(err)
		return
	}
	for _, group := range out.PlacementGroups {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			StringValue(group.GroupName),
			StringValue(group.GroupName),
			"aws_placement_group",
			"aws",
			awsNetCoreAllowEmptyValues,
		))
	}
}

func (g *Ec2Generator) getSpotFleetRequests(svc *ec2.Client) {
	p := ec2.NewDescribeSpotFleetRequestsPaginator(svc, &ec2.DescribeSpotFleetRequestsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, request := range page.SpotFleetRequestConfigs {
			switch string(request.SpotFleetRequestState) {
			case "submitted", "active", "modifying":
			default:
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(request.SpotFleetRequestId),
				StringValue(request.SpotFleetRequestId),
				"aws_spot_fleet_request",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
}

// The spot datafeed subscription is a per-account singleton, but it has a
// Describe API and a provider importer (fixed id), so it is emitted when a
// subscription exists.
func (g *Ec2Generator) getSpotDatafeedSubscription(svc *ec2.Client) {
	out, err := svc.DescribeSpotDatafeedSubscription(context.TODO(), &ec2.DescribeSpotDatafeedSubscriptionInput{})
	if err != nil || out.SpotDatafeedSubscription == nil || out.SpotDatafeedSubscription.Bucket == nil {
		return
	}
	g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
		"spot-datafeed-subscription",
		"spot-datafeed-subscription",
		"aws_spot_datafeed_subscription",
		"aws",
		awsNetCoreAllowEmptyValues,
	))
}

func (g *Ec2Generator) getSecondaryNetworks(svc *ec2.Client) {
	networkPaginator := ec2.NewDescribeSecondaryNetworksPaginator(svc, &ec2.DescribeSecondaryNetworksInput{})
	for networkPaginator.HasMorePages() {
		page, err := networkPaginator.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			break
		}
		for _, network := range page.SecondaryNetworks {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(network.SecondaryNetworkId),
				StringValue(network.SecondaryNetworkId),
				"aws_ec2_secondary_network",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
	subnetPaginator := ec2.NewDescribeSecondarySubnetsPaginator(svc, &ec2.DescribeSecondarySubnetsInput{})
	for subnetPaginator.HasMorePages() {
		page, err := subnetPaginator.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			break
		}
		for _, subnet := range page.SecondarySubnets {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(subnet.SecondarySubnetId),
				StringValue(subnet.SecondarySubnetId),
				"aws_ec2_secondary_subnet",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
}
