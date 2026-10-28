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

package gcp

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/dns/v1"
)

const dnsBaseURL = "https://dns.googleapis.com/v1/"

// initDnsDepthResources is the Cloud DNS depth pass (the dns_depth.go
// sibling of clouddns.go, following the storage_depth.go pattern): the
// managed zone IAM policy/binding/member trio expanded from each
// zone's live IAM policy, DNS server policies, response policies and
// their rules, and the region's outbound endpoints. Outbound endpoints
// are listed over raw REST: the live dns v1 API exposes
// projects.locations.outboundEndpoints (verified against the Cloud
// DNS REST reference), but the pinned generated client predates the
// collection.
func (g *CloudDNSGenerator) initDnsDepthResources(ctx context.Context, svc *dns.Service, restClient *http.Client) {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region)

	err := svc.ManagedZones.List(project).Pages(ctx, func(page *dns.ManagedZonesListResponse) error {
		for _, zone := range page.ManagedZones {
			g.loadDnsZoneIam(ctx, svc, project, zone.Name)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = svc.Policies.List(project).Pages(ctx, func(page *dns.PoliciesListResponse) error {
		for _, policy := range page.Policies {
			g.Resources = append(g.Resources, g.newDnsDepthResource(
				"projects/"+project+"/policies/"+policy.Name,
				policy.Name,
				"google_dns_policy",
				map[string]string{"project": project, "name": policy.Name},
			))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = svc.ResponsePolicies.List(project).Pages(ctx, func(page *dns.ResponsePoliciesListResponse) error {
		for _, responsePolicy := range page.ResponsePolicies {
			responsePolicyID := "projects/" + project + "/responsePolicies/" + responsePolicy.ResponsePolicyName
			g.Resources = append(g.Resources, g.newDnsDepthResource(
				responsePolicyID,
				responsePolicy.ResponsePolicyName,
				"google_dns_response_policy",
				map[string]string{"project": project, "name": responsePolicy.ResponsePolicyName},
			))
			ruleErr := svc.ResponsePolicyRules.List(project, responsePolicy.ResponsePolicyName).Pages(ctx, func(page *dns.ResponsePolicyRulesListResponse) error {
				for _, rule := range page.ResponsePolicyRules {
					g.Resources = append(g.Resources, g.newDnsDepthResource(
						responsePolicyID+"/rules/"+rule.RuleName,
						rule.RuleName,
						"google_dns_response_policy_rule",
						map[string]string{"project": project, "response_policy": responsePolicy.ResponsePolicyName, "name": rule.RuleName},
					))
				}
				return nil
			})
			if ruleErr != nil {
				log.Println(ruleErr)
			}
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}

	if restClient != nil && region.Name != "" {
		endpoints, err := restListResources(ctx, restClient, dnsBaseURL+"projects/"+project+"/locations/"+region.Name+"/outboundEndpoints")
		if err != nil {
			log.Println(err)
		}
		for _, endpoint := range endpoints {
			parts := strings.Split(endpoint.Name, "/")
			g.Resources = append(g.Resources, g.newDnsDepthResource(
				endpoint.Name,
				parts[len(parts)-1],
				"google_dns_outbound_endpoint",
				map[string]string{"project": project, "location": region.Name},
			))
		}
	}
}

func (g *CloudDNSGenerator) newDnsDepthResource(id, name, resourceType string, attrs map[string]string) terraformutils.Resource {
	return terraformutils.NewResource(
		id,
		name,
		resourceType,
		g.ProviderName,
		attrs,
		cloudDNSAllowEmptyValues,
		cloudDNSAdditionalFields,
	)
}

func (g *CloudDNSGenerator) loadDnsZoneIam(ctx context.Context, svc *dns.Service, project, zoneName string) {
	zoneID := "projects/" + project + "/managedZones/" + zoneName
	policy, err := svc.ManagedZones.GetIamPolicy(zoneID, &dns.GoogleIamV1GetIamPolicyRequest{}).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		return
	}
	bindings := []iamBinding{}
	for _, binding := range policy.Bindings {
		bindings = append(bindings, iamBinding{Role: binding.Role, Members: binding.Members})
	}
	g.Resources = append(g.Resources, expandIamPolicy(bindings, zoneID,
		"google_dns_managed_zone_iam_policy", "google_dns_managed_zone_iam_binding", "google_dns_managed_zone_iam_member",
		map[string]string{"project": project, "managed_zone": zoneName},
		func(id, resourceType string, attrs map[string]string) terraformutils.Resource {
			return g.newDnsDepthResource(id, zoneName, resourceType, attrs)
		})...)
}
