// Copyright 2026 The Terraformer Authors.
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
	"fmt"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
)

// Depth enumeration for the core Route 53 API, called from
// Route53Generator.InitResources.
//
// Intentionally not generated:
//   - aws_route53_records_exclusive: an exclusive-management resource
//     whose state is the full record set of a zone; it has no
//     independent API object (its content is exactly the
//     aws_route53_record set already generated per zone) and generating
//     it alongside the individual records would make the two fight
//     over the same record sets.
func (g *Route53Generator) addDepthResources(svc *route53.Client) error {
	for _, add := range []func(*route53.Client) error{
		g.addDelegationSets,
		g.addQueryLogs,
		g.addCidrCollections,
		g.addTrafficPolicies,
		g.addTrafficPolicyInstances,
		g.addZoneDepth,
	} {
		if err := add(svc); err != nil {
			return err
		}
	}
	return nil
}

func (g *Route53Generator) addDelegationSets(svc *route53.Client) error {
	input := &route53.ListReusableDelegationSetsInput{}
	for {
		out, err := svc.ListReusableDelegationSets(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, set := range out.DelegationSets {
			id := cleanPrefix(StringValue(set.Id), "/delegationset/")
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				id,
				"aws_route53_delegation_set",
				"aws",
				route53AllowEmptyValues,
			))
		}
		if !out.IsTruncated {
			return nil
		}
		input.Marker = out.NextMarker
	}
}

func (g *Route53Generator) addQueryLogs(svc *route53.Client) error {
	input := &route53.ListQueryLoggingConfigsInput{}
	for {
		out, err := svc.ListQueryLoggingConfigs(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, queryLog := range out.QueryLoggingConfigs {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(queryLog.Id),
				StringValue(queryLog.Id),
				"aws_route53_query_log",
				"aws",
				route53AllowEmptyValues,
			))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

// Provider ID format for locations: cidr_collection_id,location_name.
func (g *Route53Generator) addCidrCollections(svc *route53.Client) error {
	input := &route53.ListCidrCollectionsInput{}
	for {
		out, err := svc.ListCidrCollections(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, collection := range out.CidrCollections {
			collectionID := StringValue(collection.Id)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				collectionID,
				StringValue(collection.Name),
				"aws_route53_cidr_collection",
				"aws",
				route53AllowEmptyValues,
			))
			if err := g.addCidrLocations(svc, collectionID); err != nil {
				return err
			}
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *Route53Generator) addCidrLocations(svc *route53.Client, collectionID string) error {
	input := &route53.ListCidrLocationsInput{CollectionId: aws.String(collectionID)}
	for {
		out, err := svc.ListCidrLocations(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, location := range out.CidrLocations {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				collectionID+","+StringValue(location.LocationName),
				StringValue(location.LocationName),
				"aws_route53_cidr_location",
				"aws",
				map[string]string{"cidr_collection_id": collectionID},
				route53AllowEmptyValues,
				route53AdditionalFields,
			))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *Route53Generator) addTrafficPolicies(svc *route53.Client) error {
	input := &route53.ListTrafficPoliciesInput{}
	for {
		out, err := svc.ListTrafficPolicies(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, policy := range out.TrafficPolicySummaries {
			// Provider ID format: traffic_policy_id/latest_version.
			version := int32(1)
			if policy.LatestVersion != nil {
				version = *policy.LatestVersion
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				fmt.Sprintf("%s/%d", StringValue(policy.Id), version),
				StringValue(policy.Name),
				"aws_route53_traffic_policy",
				"aws",
				route53AllowEmptyValues,
			))
		}
		if !out.IsTruncated {
			return nil
		}
		input.TrafficPolicyIdMarker = out.TrafficPolicyIdMarker
	}
}

func (g *Route53Generator) addTrafficPolicyInstances(svc *route53.Client) error {
	input := &route53.ListTrafficPolicyInstancesInput{}
	for {
		out, err := svc.ListTrafficPolicyInstances(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, instance := range out.TrafficPolicyInstances {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(instance.Id),
				StringValue(instance.Name),
				"aws_route53_traffic_policy_instance",
				"aws",
				route53AllowEmptyValues,
			))
		}
		if !out.IsTruncated {
			return nil
		}
		input.HostedZoneIdMarker = out.HostedZoneIdMarker
		input.TrafficPolicyInstanceNameMarker = out.TrafficPolicyInstanceNameMarker
		input.TrafficPolicyInstanceTypeMarker = out.TrafficPolicyInstanceTypeMarker
	}
}

// Per-zone depth: DNSSEC configuration and its key-signing keys,
// VPC association authorizations, and the zone's VPC associations.
// Provider ID formats: hosted_zone_dnssec is keyed by the zone ID;
// key_signing_key is zone_id,ksk_name; both association resources are
// zone_id:vpc_id (the zone association carries the VPC region as a
// third colon-separated segment, which the API returns per VPC).
func (g *Route53Generator) addZoneDepth(svc *route53.Client) error {
	p := route53.NewListHostedZonesPaginator(svc, &route53.ListHostedZonesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, zone := range page.HostedZones {
			zoneID := cleanZoneID(StringValue(zone.Id))
			if err := g.addZoneDNSSEC(svc, zoneID); err != nil {
				return err
			}
			if err := g.addVPCAssociationAuthorizations(svc, zoneID); err != nil {
				return err
			}
			if err := g.addZoneAssociations(svc, zoneID, zone.Config != nil && zone.Config.PrivateZone); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *Route53Generator) addZoneDNSSEC(svc *route53.Client, zoneID string) error {
	out, err := svc.GetDNSSEC(context.TODO(), &route53.GetDNSSECInput{HostedZoneId: aws.String(zoneID)})
	if err != nil {
		return err
	}
	if out.Status != nil && StringValue(out.Status.ServeSignature) == "SIGNING" {
		g.Resources = append(g.Resources, terraformutils.NewResource(
			zoneID,
			zoneID,
			"aws_route53_hosted_zone_dnssec",
			"aws",
			map[string]string{"hosted_zone_id": zoneID},
			route53AllowEmptyValues,
			route53AdditionalFields,
		))
	}
	for _, ksk := range out.KeySigningKeys {
		g.Resources = append(g.Resources, terraformutils.NewResource(
			zoneID+","+StringValue(ksk.Name),
			StringValue(ksk.Name),
			"aws_route53_key_signing_key",
			"aws",
			map[string]string{"hosted_zone_id": zoneID},
			route53AllowEmptyValues,
			route53AdditionalFields,
		))
	}
	return nil
}

func (g *Route53Generator) addVPCAssociationAuthorizations(svc *route53.Client, zoneID string) error {
	input := &route53.ListVPCAssociationAuthorizationsInput{HostedZoneId: aws.String(zoneID)}
	for {
		out, err := svc.ListVPCAssociationAuthorizations(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, vpc := range out.VPCs {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				zoneID+":"+StringValue(vpc.VPCId),
				StringValue(vpc.VPCId),
				"aws_route53_vpc_association_authorization",
				"aws",
				map[string]string{"zone_id": zoneID},
				route53AllowEmptyValues,
				route53AdditionalFields,
			))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *Route53Generator) addZoneAssociations(svc *route53.Client, zoneID string, private bool) error {
	if !private {
		return nil
	}
	out, err := svc.GetHostedZone(context.TODO(), &route53.GetHostedZoneInput{Id: aws.String(zoneID)})
	if err != nil {
		return err
	}
	for _, vpc := range out.VPCs {
		id := zoneID + ":" + StringValue(vpc.VPCId)
		if vpc.VPCRegion != "" {
			id += ":" + string(vpc.VPCRegion)
		}
		g.Resources = append(g.Resources, terraformutils.NewResource(
			id,
			StringValue(vpc.VPCId),
			"aws_route53_zone_association",
			"aws",
			map[string]string{"zone_id": zoneID},
			route53AllowEmptyValues,
			route53AdditionalFields,
		))
	}
	return nil
}
