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
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/compute/v1"
	netsecv1 "google.golang.org/api/networksecurity/v1"
	netsecbeta "google.golang.org/api/networksecurity/v1beta1"
)

var networkSecurityAllowEmptyValues = []string{"tags."}

var networkSecurityAdditionalFields = map[string]interface{}{}

type NetworkSecurityGenerator struct {
	GCPService
}

// InitResources enumerates Cloud Network Security resources for the
// configured project and region: address groups (with the IAM policy,
// bindings and members expanded from each project-scoped group's live IAM
// policy), authorization policies, authz policies, client and server TLS
// policies, gateway security policies and their rules, TLS inspection
// policies and URL lists at both the configured region and the global
// location; packet mirroring and intercept deployment groups, endpoint
// groups and their regional associations; firewall endpoint associations
// per zone of the configured region; and — when the project's organization
// can be resolved through Cloud Resource Manager ancestry — the
// organization-scoped address groups, zonal firewall endpoints, security
// profiles and security profile groups.
//
// backend_authentication_config and the intercept resources are served by
// the v1beta1 client: they have no GA list surface, matching the
// google provider, which also exposes them as GA resources backed by the
// beta API.
//
// Deliberately not generated (no list call exists in either the v1 or the
// v1beta1 networksecurity client, verified against the generated
// clients):
//   - google_network_security_dns_threat_detector: there is no DNS
//     threat detector service at all.
//   - google_network_security_sac_attachment and
//     google_network_security_sac_realm: no Secure Access Connect
//     services exist in either client.
//   - google_network_security_ull_mirroring_collector,
//     google_network_security_ull_mirroring_collector_rule and
//     google_network_security_ull_mirroring_engine: no ULL mirroring
//     services exist in either client.
//   - google_network_security_mirroring_endpoint: only mirroring
//     endpoint groups and endpoint group associations exist; there is no
//     standalone endpoint collection.
//   - Project-scoped google_network_security_firewall_endpoint resources
//     are not listed: the client exposes firewall endpoint listing only
//     under organizations. Organization-scoped endpoints are enumerated.
func (g *NetworkSecurityGenerator) InitResources() error {
	ctx := context.Background()
	v1Service, err := netsecv1.NewService(ctx)
	if err != nil {
		return err
	}
	betaService, err := netsecbeta.NewService(ctx)
	if err != nil {
		return err
	}
	crmService, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, v1Service, betaService, crmService)
}

func (g *NetworkSecurityGenerator) initResources(ctx context.Context, v1Service *netsecv1.Service, betaService *netsecbeta.Service, crmService *cloudresourcemanager.Service) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region)
	regionName := region.Name
	zones := []string{}
	for _, zone := range region.Zones {
		zones = append(zones, zone[strings.LastIndex(zone, "/")+1:])
	}
	projectParent := "projects/" + project

	if err := g.loadRegionalCollections(ctx, v1Service, betaService, projectParent, project, regionName); err != nil {
		return err
	}
	// Most Network Security collections also exist at the global
	// location; a resource lives in exactly one location, so listing
	// both never duplicates.
	if err := g.loadRegionalCollections(ctx, v1Service, betaService, projectParent, project, "global"); err != nil {
		log.Println(err)
	}
	if err := g.loadGlobalCollections(ctx, v1Service, betaService, projectParent, project); err != nil {
		return err
	}
	if err := g.loadZonalCollections(ctx, v1Service, betaService, projectParent, project, regionName, zones); err != nil {
		return err
	}

	orgID := ""
	if crmService != nil {
		if ancestry, err := crmService.Projects.GetAncestry(project, &cloudresourcemanager.GetAncestryRequest{}).Context(ctx).Do(); err == nil {
			for _, ancestor := range ancestry.Ancestor {
				if ancestor.ResourceId != nil && ancestor.ResourceId.Type == "organization" {
					orgID = ancestor.ResourceId.Id
				}
			}
		} else {
			log.Println(err)
		}
	}
	if orgID == "" {
		log.Println("network security: could not resolve organization for project " + project + "; skipping organization-scoped collections")
		return nil
	}
	orgParent := "organizations/" + orgID
	g.loadOrgCollections(ctx, v1Service, orgParent, orgID, regionName, zones)
	return nil
}

func (g *NetworkSecurityGenerator) newNetSecResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
	if attrs == nil {
		attrs = map[string]string{}
	}
	parts := strings.Split(id, "/")
	return terraformutils.NewResource(
		id,
		parts[len(parts)-1],
		resourceType,
		g.ProviderName,
		attrs,
		networkSecurityAllowEmptyValues,
		networkSecurityAdditionalFields,
	)
}

func (g *NetworkSecurityGenerator) loadRegionalCollections(ctx context.Context, v1Service *netsecv1.Service, betaService *netsecbeta.Service, parent, project, location string) error {
	locationParent := parent + "/locations/" + location
	attrs := map[string]string{"project": project, "location": location}

	// Address groups additionally expand their IAM policy, following the
	// Cloud Run IAM pattern.
	err := v1Service.Projects.Locations.AddressGroups.List(locationParent).Pages(ctx, func(page *netsecv1.ListAddressGroupsResponse) error {
		for _, group := range page.AddressGroups {
			g.Resources = append(g.Resources, g.newNetSecResource(group.Name, "google_network_security_address_group", attrs))
			g.loadAddressGroupIam(ctx, v1Service, group.Name)
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = v1Service.Projects.Locations.AuthorizationPolicies.List(locationParent).Pages(ctx, func(page *netsecv1.ListAuthorizationPoliciesResponse) error {
		for _, policy := range page.AuthorizationPolicies {
			g.Resources = append(g.Resources, g.newNetSecResource(policy.Name, "google_network_security_authorization_policy", attrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = v1Service.Projects.Locations.AuthzPolicies.List(locationParent).Pages(ctx, func(page *netsecv1.ListAuthzPoliciesResponse) error {
		for _, policy := range page.AuthzPolicies {
			g.Resources = append(g.Resources, g.newNetSecResource(policy.Name, "google_network_security_authz_policy", attrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = v1Service.Projects.Locations.ClientTlsPolicies.List(locationParent).Pages(ctx, func(page *netsecv1.ListClientTlsPoliciesResponse) error {
		for _, policy := range page.ClientTlsPolicies {
			g.Resources = append(g.Resources, g.newNetSecResource(policy.Name, "google_network_security_client_tls_policy", attrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = v1Service.Projects.Locations.ServerTlsPolicies.List(locationParent).Pages(ctx, func(page *netsecv1.ListServerTlsPoliciesResponse) error {
		for _, policy := range page.ServerTlsPolicies {
			g.Resources = append(g.Resources, g.newNetSecResource(policy.Name, "google_network_security_server_tls_policy", attrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = v1Service.Projects.Locations.GatewaySecurityPolicies.List(locationParent).Pages(ctx, func(page *netsecv1.ListGatewaySecurityPoliciesResponse) error {
		for _, policy := range page.GatewaySecurityPolicies {
			g.Resources = append(g.Resources, g.newNetSecResource(policy.Name, "google_network_security_gateway_security_policy", attrs))
			ruleErr := v1Service.Projects.Locations.GatewaySecurityPolicies.Rules.List(policy.Name).Pages(ctx, func(page *netsecv1.ListGatewaySecurityPolicyRulesResponse) error {
				for _, rule := range page.GatewaySecurityPolicyRules {
					g.Resources = append(g.Resources, g.newNetSecResource(rule.Name, "google_network_security_gateway_security_policy_rule", attrs))
				}
				return nil
			})
			if ruleErr != nil {
				return ruleErr
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = v1Service.Projects.Locations.TlsInspectionPolicies.List(locationParent).Pages(ctx, func(page *netsecv1.ListTlsInspectionPoliciesResponse) error {
		for _, policy := range page.TlsInspectionPolicies {
			g.Resources = append(g.Resources, g.newNetSecResource(policy.Name, "google_network_security_tls_inspection_policy", attrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = v1Service.Projects.Locations.UrlLists.List(locationParent).Pages(ctx, func(page *netsecv1.ListUrlListsResponse) error {
		for _, list := range page.UrlLists {
			g.Resources = append(g.Resources, g.newNetSecResource(list.Name, "google_network_security_url_lists", attrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	return betaService.Projects.Locations.BackendAuthenticationConfigs.List(locationParent).Pages(ctx, func(page *netsecbeta.ListBackendAuthenticationConfigsResponse) error {
		for _, config := range page.BackendAuthenticationConfigs {
			g.Resources = append(g.Resources, g.newNetSecResource(config.Name, "google_network_security_backend_authentication_config", attrs))
		}
		return nil
	})
}

func (g *NetworkSecurityGenerator) loadAddressGroupIam(ctx context.Context, v1Service *netsecv1.Service, groupName string) {
	policy, err := v1Service.Projects.Locations.AddressGroups.GetIamPolicy(groupName).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		return
	}
	parentAttrs := map[string]string{"address_group": groupName}
	g.Resources = append(g.Resources, g.newNetSecResource(
		groupName,
		"google_network_security_address_group_iam_policy",
		parentAttrs,
	))
	for _, binding := range policy.Bindings {
		bindingAttrs := map[string]string{"role": binding.Role}
		for k, v := range parentAttrs {
			bindingAttrs[k] = v
		}
		g.Resources = append(g.Resources, g.newNetSecResource(
			groupName,
			"google_network_security_address_group_iam_binding",
			bindingAttrs,
		))
		for _, member := range binding.Members {
			memberAttrs := map[string]string{"role": binding.Role, "member": member}
			for k, v := range parentAttrs {
				memberAttrs[k] = v
			}
			g.Resources = append(g.Resources, g.newNetSecResource(
				groupName,
				"google_network_security_address_group_iam_member",
				memberAttrs,
			))
		}
	}
}

func (g *NetworkSecurityGenerator) loadGlobalCollections(ctx context.Context, v1Service *netsecv1.Service, betaService *netsecbeta.Service, parent, project string) error {
	locationParent := parent + "/locations/global"
	attrs := map[string]string{"project": project, "location": "global"}

	err := v1Service.Projects.Locations.MirroringDeploymentGroups.List(locationParent).Pages(ctx, func(page *netsecv1.ListMirroringDeploymentGroupsResponse) error {
		for _, group := range page.MirroringDeploymentGroups {
			g.Resources = append(g.Resources, g.newNetSecResource(group.Name, "google_network_security_mirroring_deployment_group", attrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = v1Service.Projects.Locations.MirroringEndpointGroups.List(locationParent).Pages(ctx, func(page *netsecv1.ListMirroringEndpointGroupsResponse) error {
		for _, group := range page.MirroringEndpointGroups {
			g.Resources = append(g.Resources, g.newNetSecResource(group.Name, "google_network_security_mirroring_endpoint_group", attrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = betaService.Projects.Locations.InterceptDeploymentGroups.List(locationParent).Pages(ctx, func(page *netsecbeta.ListInterceptDeploymentGroupsResponse) error {
		for _, group := range page.InterceptDeploymentGroups {
			g.Resources = append(g.Resources, g.newNetSecResource(group.Name, "google_network_security_intercept_deployment_group", attrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	return betaService.Projects.Locations.InterceptEndpointGroups.List(locationParent).Pages(ctx, func(page *netsecbeta.ListInterceptEndpointGroupsResponse) error {
		for _, group := range page.InterceptEndpointGroups {
			g.Resources = append(g.Resources, g.newNetSecResource(group.Name, "google_network_security_intercept_endpoint_group", attrs))
		}
		return nil
	})
}

func (g *NetworkSecurityGenerator) loadZonalCollections(ctx context.Context, v1Service *netsecv1.Service, betaService *netsecbeta.Service, parent, project, regionName string, zones []string) error {
	for _, zone := range zones {
		locationParent := parent + "/locations/" + zone
		attrs := map[string]string{"project": project, "location": zone}
		err := v1Service.Projects.Locations.FirewallEndpointAssociations.List(locationParent).Pages(ctx, func(page *netsecv1.ListFirewallEndpointAssociationsResponse) error {
			for _, association := range page.FirewallEndpointAssociations {
				g.Resources = append(g.Resources, g.newNetSecResource(association.Name, "google_network_security_firewall_endpoint_association", attrs))
			}
			return nil
		})
		if err != nil {
			log.Println(err)
		}
		err = v1Service.Projects.Locations.MirroringDeployments.List(locationParent).Pages(ctx, func(page *netsecv1.ListMirroringDeploymentsResponse) error {
			for _, deployment := range page.MirroringDeployments {
				g.Resources = append(g.Resources, g.newNetSecResource(deployment.Name, "google_network_security_mirroring_deployment", attrs))
			}
			return nil
		})
		if err != nil {
			log.Println(err)
		}
		err = betaService.Projects.Locations.InterceptDeployments.List(locationParent).Pages(ctx, func(page *netsecbeta.ListInterceptDeploymentsResponse) error {
			for _, deployment := range page.InterceptDeployments {
				g.Resources = append(g.Resources, g.newNetSecResource(deployment.Name, "google_network_security_intercept_deployment", attrs))
			}
			return nil
		})
		if err != nil {
			log.Println(err)
		}
	}
	// Endpoint group associations live in the region of the consumer
	// network they attach to.
	locationParent := parent + "/locations/" + regionName
	attrs := map[string]string{"project": project, "location": regionName}
	err := v1Service.Projects.Locations.MirroringEndpointGroupAssociations.List(locationParent).Pages(ctx, func(page *netsecv1.ListMirroringEndpointGroupAssociationsResponse) error {
		for _, association := range page.MirroringEndpointGroupAssociations {
			g.Resources = append(g.Resources, g.newNetSecResource(association.Name, "google_network_security_mirroring_endpoint_group_association", attrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = betaService.Projects.Locations.InterceptEndpointGroupAssociations.List(locationParent).Pages(ctx, func(page *netsecbeta.ListInterceptEndpointGroupAssociationsResponse) error {
		for _, association := range page.InterceptEndpointGroupAssociations {
			g.Resources = append(g.Resources, g.newNetSecResource(association.Name, "google_network_security_intercept_endpoint_group_association", attrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	return nil
}

// loadOrgCollections enumerates the organization-scoped collections.
// Individual list failures are logged rather than returned: these
// collections are best-effort on top of the project-scoped enumeration.
func (g *NetworkSecurityGenerator) loadOrgCollections(ctx context.Context, v1Service *netsecv1.Service, orgParent, orgID, regionName string, zones []string) {
	for _, location := range []string{regionName, "global"} {
		locationParent := orgParent + "/locations/" + location
		attrs := map[string]string{"organization": orgID, "location": location}
		err := v1Service.Organizations.Locations.AddressGroups.List(locationParent).Pages(ctx, func(page *netsecv1.ListAddressGroupsResponse) error {
			for _, group := range page.AddressGroups {
				g.Resources = append(g.Resources, g.newNetSecResource(group.Name, "google_network_security_address_group", attrs))
			}
			return nil
		})
		if err != nil {
			log.Println(err)
		}
	}
	globalParent := orgParent + "/locations/global"
	attrs := map[string]string{"organization": orgID, "location": "global"}
	err := v1Service.Organizations.Locations.SecurityProfiles.List(globalParent).Pages(ctx, func(page *netsecv1.ListSecurityProfilesResponse) error {
		for _, profile := range page.SecurityProfiles {
			g.Resources = append(g.Resources, g.newNetSecResource(profile.Name, "google_network_security_security_profile", attrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = v1Service.Organizations.Locations.SecurityProfileGroups.List(globalParent).Pages(ctx, func(page *netsecv1.ListSecurityProfileGroupsResponse) error {
		for _, group := range page.SecurityProfileGroups {
			g.Resources = append(g.Resources, g.newNetSecResource(group.Name, "google_network_security_security_profile_group", attrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	for _, zone := range zones {
		locationParent := orgParent + "/locations/" + zone
		zoneAttrs := map[string]string{"organization": orgID, "location": zone}
		err := v1Service.Organizations.Locations.FirewallEndpoints.List(locationParent).Pages(ctx, func(page *netsecv1.ListFirewallEndpointsResponse) error {
			for _, endpoint := range page.FirewallEndpoints {
				g.Resources = append(g.Resources, g.newNetSecResource(endpoint.Name, "google_network_security_firewall_endpoint", zoneAttrs))
			}
			return nil
		})
		if err != nil {
			log.Println(err)
		}
	}
}
