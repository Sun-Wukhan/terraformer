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
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/compute/v1"
)

var computePoliciesAllowEmptyValues = []string{""}

var computePoliciesAdditionalFields = map[string]interface{}{}

type ComputePoliciesGenerator struct {
	GCPService
}

// InitResources enumerates the Compute Engine firewall and security
// policy families:
//
//   - Network firewall policies (global and regional) with their
//     embedded rules, packet mirroring rules (global only) and
//     associations, from the generated client.
//   - Cloud Armor security policy rules: the global policies are
//     emitted by the securityPolicies generator, so only their rules
//     are added here; regional policies are emitted with their rules.
//   - Hierarchy firewall policies (google_compute_firewall_policy and
//     its rules and associations) for every parent the project can
//     see — the organization and its ancestor folders, resolved
//     through Cloud Resource Manager ancestry. The list endpoint is
//     locations/global/firewallPolicies with a parentId per parent.
//   - Organization security policies with their rules and
//     associations, over raw REST (the pinned client has no service
//     for the GA locations/global/securityPolicies endpoints), again
//     scoped to the resolved organization.
//   - Region network policies and their traffic classification rules,
//     over raw REST against the beta endpoint: the family exists only
//     in the beta compute API, which is also what backs the provider's
//     GA resources (the Network Security v1beta1 precedent).
//
// Deliberately not generated:
//   - google_compute_firewall_policy_with_rules,
//     google_compute_network_firewall_policy_with_rules and
//     google_compute_region_network_firewall_policy_with_rules: the
//     same API object and the same provider ID as the policy resource
//     emitted here, with the rules managed inline instead of as
//     separate rule resources. Emitting both would double-generate.
//   - google_compute_router_nat_address,
//     google_compute_router_route_policy and
//     google_compute_router_named_set: documented in
//     routerSubresources.go (NAT addresses are fields of the NAT
//     config; route policies and named sets are not exposed on the GA
//     Router resource — re-verified against the pinned client for
//     this batch).
func (g *ComputePoliciesGenerator) InitResources() error {
	ctx := context.Background()
	computeService, err := compute.NewService(ctx)
	if err != nil {
		return err
	}
	crmService, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		return err
	}
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, computeService, crmService, restClient)
}

func (g *ComputePoliciesGenerator) initResources(ctx context.Context, computeService *compute.Service, crmService *cloudresourcemanager.Service, restClient *http.Client) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name

	if err := g.loadNetworkFirewallPolicies(ctx, computeService, project); err != nil {
		return err
	}
	g.loadRegionNetworkFirewallPolicies(ctx, computeService, project, region)
	g.loadRegionSecurityPolicies(ctx, computeService, project, region)
	g.loadGlobalSecurityPolicyRules(ctx, computeService, project)
	g.loadRegionNetworkPolicies(ctx, restClient, project, region)

	orgID, folderIDs := "", []string{}
	if crmService != nil {
		if ancestry, err := crmService.Projects.GetAncestry(project, &cloudresourcemanager.GetAncestryRequest{}).Context(ctx).Do(); err == nil {
			for _, ancestor := range ancestry.Ancestor {
				if ancestor.ResourceId == nil {
					continue
				}
				switch ancestor.ResourceId.Type {
				case "organization":
					orgID = ancestor.ResourceId.Id
				case "folder":
					folderIDs = append(folderIDs, ancestor.ResourceId.Id)
				}
			}
		} else {
			log.Println(err)
		}
	}
	parents := []string{}
	for _, folderID := range folderIDs {
		parents = append(parents, "folders/"+folderID)
	}
	if orgID != "" {
		parents = append(parents, "organizations/"+orgID)
	}
	if len(parents) == 0 {
		log.Println("compute policies: could not resolve organization or folders for project " + project + "; skipping hierarchy and organization policies")
		return nil
	}
	for _, parent := range parents {
		g.loadHierarchyFirewallPolicies(ctx, computeService, parent)
	}
	if orgID != "" {
		g.loadOrganizationSecurityPolicies(ctx, restClient, "organizations/"+orgID)
	}
	return nil
}

func (g *ComputePoliciesGenerator) newComputePolicyResource(id, name, resourceType string, attrs map[string]string) terraformutils.Resource {
	if attrs == nil {
		attrs = map[string]string{}
	}
	return terraformutils.NewResource(
		id,
		name,
		resourceType,
		g.ProviderName,
		attrs,
		computePoliciesAllowEmptyValues,
		computePoliciesAdditionalFields,
	)
}

// emitFirewallPolicyChildren emits the rules, packet mirroring rules
// and associations embedded in a FirewallPolicy object, under the
// provider ID prefix of the policy. kind selects the resource family
// ("network", "region" or "hierarchy"); ruleSegment is "rules" for the
// global families, while the regional network firewall policy rule ID
// omits the segment (a documented provider quirk).
func (g *ComputePoliciesGenerator) emitFirewallPolicyChildren(policy *compute.FirewallPolicy, policyID, policyName string, attrs map[string]string, kind, ruleSegment string, packetMirroring bool) {
	ruleType := map[string]string{
		"network":   "google_compute_network_firewall_policy_rule",
		"region":    "google_compute_region_network_firewall_policy_rule",
		"hierarchy": "google_compute_firewall_policy_rule",
	}[kind]
	associationType := map[string]string{
		"network":   "google_compute_network_firewall_policy_association",
		"region":    "google_compute_region_network_firewall_policy_association",
		"hierarchy": "google_compute_firewall_policy_association",
	}[kind]
	for _, rule := range policy.Rules {
		ruleAttrs := map[string]string{"firewall_policy": policyName, "priority": fmt.Sprint(rule.Priority)}
		for k, v := range attrs {
			ruleAttrs[k] = v
		}
		ruleID := policyID + "/" + fmt.Sprint(rule.Priority)
		if ruleSegment != "" {
			ruleID = policyID + "/" + ruleSegment + "/" + fmt.Sprint(rule.Priority)
		}
		g.Resources = append(g.Resources, g.newComputePolicyResource(
			ruleID, fmt.Sprint(rule.Priority), ruleType, ruleAttrs))
	}
	if packetMirroring {
		for _, rule := range policy.PacketMirroringRules {
			ruleAttrs := map[string]string{"firewall_policy": policyName, "priority": fmt.Sprint(rule.Priority)}
			for k, v := range attrs {
				ruleAttrs[k] = v
			}
			g.Resources = append(g.Resources, g.newComputePolicyResource(
				policyID+"/packetMirroringRules/"+fmt.Sprint(rule.Priority), fmt.Sprint(rule.Priority),
				"google_compute_network_firewall_policy_packet_mirroring_rule", ruleAttrs))
		}
	}
	for _, association := range policy.Associations {
		assocAttrs := map[string]string{"name": association.Name, "firewall_policy": policyName, "attachment_target": association.AttachmentTarget}
		for k, v := range attrs {
			assocAttrs[k] = v
		}
		g.Resources = append(g.Resources, g.newComputePolicyResource(
			policyID+"/associations/"+association.Name, association.Name, associationType, assocAttrs))
	}
}

func (g *ComputePoliciesGenerator) loadNetworkFirewallPolicies(ctx context.Context, computeService *compute.Service, project string) error {
	return computeService.NetworkFirewallPolicies.List(project).Pages(ctx, func(page *compute.FirewallPolicyList) error {
		for _, policy := range page.Items {
			policyID := "projects/" + project + "/global/firewallPolicies/" + policy.Name
			attrs := map[string]string{"project": project}
			g.Resources = append(g.Resources, g.newComputePolicyResource(
				policy.Name, policy.Name, "google_compute_network_firewall_policy",
				map[string]string{"name": policy.Name, "project": project}))
			g.emitFirewallPolicyChildren(policy, policyID, policy.Name, attrs, "network", "rules", true)
		}
		return nil
	})
}

func (g *ComputePoliciesGenerator) loadRegionNetworkFirewallPolicies(ctx context.Context, computeService *compute.Service, project, region string) {
	err := computeService.RegionNetworkFirewallPolicies.List(project, region).Pages(ctx, func(page *compute.FirewallPolicyList) error {
		for _, policy := range page.Items {
			policyID := "projects/" + project + "/regions/" + region + "/firewallPolicies/" + policy.Name
			attrs := map[string]string{"project": project, "region": region}
			g.Resources = append(g.Resources, g.newComputePolicyResource(
				policy.Name, policy.Name, "google_compute_region_network_firewall_policy",
				map[string]string{"name": policy.Name, "project": project, "region": region}))
			g.emitFirewallPolicyChildren(policy, policyID, policy.Name, attrs, "region", "", false)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (g *ComputePoliciesGenerator) loadRegionSecurityPolicies(ctx context.Context, computeService *compute.Service, project, region string) {
	err := computeService.RegionSecurityPolicies.List(project, region).Pages(ctx, func(page *compute.SecurityPolicyList) error {
		for _, policy := range page.Items {
			attrs := map[string]string{"name": policy.Name, "project": project, "region": region}
			g.Resources = append(g.Resources, g.newComputePolicyResource(
				policy.Name, policy.Name, "google_compute_region_security_policy", attrs))
			for _, rule := range policy.Rules {
				ruleID := "projects/" + project + "/regions/" + region + "/securityPolicies/" + policy.Name + "/priority/" + fmt.Sprint(rule.Priority)
				ruleAttrs := map[string]string{"security_policy": policy.Name, "priority": fmt.Sprint(rule.Priority), "project": project, "region": region}
				g.Resources = append(g.Resources, g.newComputePolicyResource(
					ruleID, fmt.Sprint(rule.Priority), "google_compute_region_security_policy_rule", ruleAttrs))
			}
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (g *ComputePoliciesGenerator) loadGlobalSecurityPolicyRules(ctx context.Context, computeService *compute.Service, project string) {
	err := computeService.SecurityPolicies.List(project).Pages(ctx, func(page *compute.SecurityPolicyList) error {
		for _, policy := range page.Items {
			for _, rule := range policy.Rules {
				ruleID := "projects/" + project + "/global/securityPolicies/" + policy.Name + "/priority/" + fmt.Sprint(rule.Priority)
				ruleAttrs := map[string]string{"security_policy": policy.Name, "priority": fmt.Sprint(rule.Priority), "project": project}
				g.Resources = append(g.Resources, g.newComputePolicyResource(
					ruleID, fmt.Sprint(rule.Priority), "google_compute_security_policy_rule", ruleAttrs))
			}
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

// loadHierarchyFirewallPolicies lists the hierarchy firewall policies
// of one parent (an organization or a folder). The policy's own ID is
// the locations/global/firewallPolicies/{id} form the provider uses.
func (g *ComputePoliciesGenerator) loadHierarchyFirewallPolicies(ctx context.Context, computeService *compute.Service, parent string) {
	err := computeService.FirewallPolicies.List().ParentId(parent).Pages(ctx, func(page *compute.FirewallPolicyList) error {
		for _, policy := range page.Items {
			id := policy.Name[strings.LastIndex(policy.Name, "/")+1:]
			policyID := "locations/global/firewallPolicies/" + id
			name := policy.ShortName
			if name == "" {
				name = id
			}
			g.Resources = append(g.Resources, g.newComputePolicyResource(
				policyID, name, "google_compute_firewall_policy",
				map[string]string{"name": id, "parent": parent}))
			// The rule and association resources identify their
			// policy through firewall_policy alone; no parent field.
			g.emitFirewallPolicyChildren(policy, policyID, id, map[string]string{}, "hierarchy", "rules", false)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

type orgSecurityPolicyAssociation struct {
	Name             string `json:"name"`
	AttachmentTarget string `json:"attachmentTarget"`
}

// loadOrganizationSecurityPolicies enumerates the organization
// security policies of the project's organization over raw REST: the
// GA endpoints exist in the live compute discovery document but the
// pinned generated client has no service for them.
func (g *ComputePoliciesGenerator) loadOrganizationSecurityPolicies(ctx context.Context, restClient *http.Client, parent string) {
	base := "https://compute.googleapis.com/compute/v1/locations/global/securityPolicies"
	pageToken := ""
	for {
		listURL := base + "?parentId=" + url.QueryEscape(parent)
		if pageToken != "" {
			listURL += "&pageToken=" + url.QueryEscape(pageToken)
		}
		body, err := restDo(ctx, restClient, http.MethodGet, listURL, nil)
		if err != nil {
			log.Println(err)
			return
		}
		var page struct {
			Items         []*compute.SecurityPolicy `json:"items"`
			NextPageToken string                    `json:"nextPageToken"`
		}
		if err := json.Unmarshal(body, &page); err != nil {
			log.Println(err)
			return
		}
		for _, policy := range page.Items {
			id := policy.Name[strings.LastIndex(policy.Name, "/")+1:]
			policyID := "locations/global/securityPolicies/" + id
			g.Resources = append(g.Resources, g.newComputePolicyResource(
				policyID, id, "google_compute_organization_security_policy",
				map[string]string{"policy_id": policyID, "parent": parent}))
			for _, rule := range policy.Rules {
				ruleAttrs := map[string]string{"policy_id": policyID, "priority": fmt.Sprint(rule.Priority)}
				g.Resources = append(g.Resources, g.newComputePolicyResource(
					policyID+"/priority/"+fmt.Sprint(rule.Priority), fmt.Sprint(rule.Priority),
					"google_compute_organization_security_policy_rule", ruleAttrs))
			}
			g.loadOrganizationSecurityPolicyAssociations(ctx, restClient, base, policyID)
		}
		if page.NextPageToken == "" {
			return
		}
		pageToken = page.NextPageToken
	}
}

func (g *ComputePoliciesGenerator) loadOrganizationSecurityPolicyAssociations(ctx context.Context, restClient *http.Client, base, policyID string) {
	listURL := base + "/listAssociations?targetResource=" + url.QueryEscape(policyID)
	body, err := restDo(ctx, restClient, http.MethodGet, listURL, nil)
	if err != nil {
		log.Println(err)
		return
	}
	var response struct {
		Associations []orgSecurityPolicyAssociation `json:"associations"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		log.Println(err)
		return
	}
	for _, association := range response.Associations {
		attrs := map[string]string{"policy_id": policyID, "name": association.Name, "attachment_target": association.AttachmentTarget}
		g.Resources = append(g.Resources, g.newComputePolicyResource(
			policyID+"/association/"+association.Name, association.Name,
			"google_compute_organization_security_policy_association", attrs))
	}
}

type regionNetworkPolicy struct {
	Name                       string `json:"name"`
	TrafficClassificationRules []struct {
		Priority int64 `json:"priority"`
	} `json:"trafficClassificationRules"`
}

// loadRegionNetworkPolicies enumerates region network policies and
// their embedded traffic classification rules over raw REST against
// the beta compute endpoint (the family has no GA surface).
func (g *ComputePoliciesGenerator) loadRegionNetworkPolicies(ctx context.Context, restClient *http.Client, project, region string) {
	base := "https://compute.googleapis.com/compute/beta/projects/" + project + "/regions/" + region + "/networkPolicies"
	pageToken := ""
	for {
		listURL := base
		if pageToken != "" {
			listURL += "?pageToken=" + url.QueryEscape(pageToken)
		}
		body, err := restDo(ctx, restClient, http.MethodGet, listURL, nil)
		if err != nil {
			log.Println(err)
			return
		}
		var page struct {
			Items         []regionNetworkPolicy `json:"items"`
			NextPageToken string                `json:"nextPageToken"`
		}
		if err := json.Unmarshal(body, &page); err != nil {
			log.Println(err)
			return
		}
		for _, policy := range page.Items {
			attrs := map[string]string{"name": policy.Name, "project": project, "region": region}
			g.Resources = append(g.Resources, g.newComputePolicyResource(
				policy.Name, policy.Name, "google_compute_region_network_policy", attrs))
			for _, rule := range policy.TrafficClassificationRules {
				ruleID := "projects/" + project + "/regions/" + region + "/networkPolicies/" + policy.Name + "/" + fmt.Sprint(rule.Priority)
				ruleAttrs := map[string]string{"network_policy": policy.Name, "priority": fmt.Sprint(rule.Priority), "project": project, "region": region}
				g.Resources = append(g.Resources, g.newComputePolicyResource(
					ruleID, fmt.Sprint(rule.Priority), "google_compute_region_network_policy_traffic_classification_rule", ruleAttrs))
			}
		}
		if page.NextPageToken == "" {
			return
		}
		pageToken = page.NextPageToken
	}
}
