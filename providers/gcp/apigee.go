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
	apigee "google.golang.org/api/apigee/v1"
)

var apigeeAllowEmptyValues = []string{"tags."}

var apigeeAdditionalFields = map[string]interface{}{}

type ApigeeGenerator struct {
	GCPService
}

// InitResources enumerates the Apigee organization bound to the configured
// project: the organization itself (with its add-ons config, control plane
// access and sync authorization singletons), environments (discovered from
// the organization resource, since the GA client has no environments list
// call) with their IAM policies, security actions, flow hooks, debug mask
// and add-ons config, and the org-level collections: API proxies and their
// deployments, API products, app groups, developers and their apps, data
// collectors, analytics datastores, DNS zones, endpoint attachments,
// environment groups and their attachments, instances with their
// attachments and NAT addresses, security monitoring conditions, security
// profiles, shared flows and their deployments, and spaces.
//
// Deliberately not generated (no list call exists for these collections in
// the GA apigee v1 client, verified against the generated client):
//   - google_apigee_env_keystore and
//     google_apigee_keystores_aliases_key_cert_file,
//     google_apigee_keystores_aliases_pkcs12 and
//     google_apigee_keystores_aliases_self_signed_cert: environment
//     keystores and their aliases expose create/get/delete only.
//   - google_apigee_env_references: environment references expose
//     create/get/delete/update only.
//   - google_apigee_environment_keyvaluemaps and
//     google_apigee_environment_keyvaluemaps_entries: the key value map
//     collection itself cannot be listed (only entries of a known map
//     can), so neither the maps nor their entries are discoverable.
//   - google_apigee_target_server: environment target servers expose
//     create/get/delete/update only.
//   - google_apigee_security_feedback: no security feedback collection
//     exists in the GA apigee v1 client at all.
func (g *ApigeeGenerator) InitResources() error {
	ctx := context.Background()
	service, err := apigee.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, service)
}

func (g *ApigeeGenerator) initResources(ctx context.Context, service *apigee.Service) error {
	project := g.GetArgs()["project"].(string)
	orgNames, err := g.discoverOrganizations(ctx, service, project)
	if err != nil {
		return err
	}
	for _, orgName := range orgNames {
		if err := g.loadOrganization(ctx, service, orgName); err != nil {
			return err
		}
	}
	return nil
}

// discoverOrganizations resolves the Apigee organizations bound to the
// configured project. The canonical Apigee X layout names the organization
// after its project, so a direct get is tried first; the organizations
// list (which returns organization-to-project mappings) is the fallback.
func (g *ApigeeGenerator) discoverOrganizations(ctx context.Context, service *apigee.Service, project string) ([]string, error) {
	orgName := "organizations/" + project
	if _, err := service.Organizations.Get(orgName).Context(ctx).Do(); err == nil {
		return []string{orgName}, nil
	}
	resp, err := service.Organizations.List("organizations").Context(ctx).Do()
	if err != nil {
		return nil, err
	}
	orgs := []string{}
	seen := map[string]bool{}
	for _, mapping := range resp.Organizations {
		match := mapping.ProjectId == project
		for _, pid := range mapping.ProjectIds {
			if pid == project {
				match = true
			}
		}
		if !match || mapping.Organization == "" {
			continue
		}
		name := mapping.Organization
		if !strings.HasPrefix(name, "organizations/") {
			name = "organizations/" + name
		}
		if !seen[name] {
			seen[name] = true
			orgs = append(orgs, name)
		}
	}
	return orgs, nil
}

func apigeeLastSegment(name string) string {
	parts := strings.Split(name, "/")
	return parts[len(parts)-1]
}

// apigeeResourceID tolerates the Apigee API's inconsistent list responses:
// some collections return short names, others full resource names.
func apigeeResourceID(parent, collection, name string) string {
	if strings.Contains(name, "/") {
		return name
	}
	return parent + "/" + collection + "/" + name
}

func (g *ApigeeGenerator) newApigeeResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
	if attrs == nil {
		attrs = map[string]string{}
	}
	return terraformutils.NewResource(
		id,
		apigeeLastSegment(id),
		resourceType,
		g.ProviderName,
		attrs,
		apigeeAllowEmptyValues,
		apigeeAdditionalFields,
	)
}

func apigeeAddonsConfigured(cfg *apigee.GoogleCloudApigeeV1AddonsConfig) bool {
	if cfg == nil {
		return false
	}
	return cfg.AdvancedApiOpsConfig != nil ||
		cfg.ApiSecurityConfig != nil ||
		cfg.ConnectorsPlatformConfig != nil ||
		cfg.IntegrationConfig != nil ||
		cfg.MonetizationConfig != nil
}

func (g *ApigeeGenerator) loadOrganization(ctx context.Context, service *apigee.Service, orgName string) error {
	org, err := service.Organizations.Get(orgName).Context(ctx).Do()
	if err != nil {
		return err
	}
	orgID := apigeeLastSegment(orgName)
	g.Resources = append(g.Resources, g.newApigeeResource(
		orgName,
		"google_apigee_organization",
		map[string]string{"project_id": org.ProjectId},
	))
	if apigeeAddonsConfigured(org.AddonsConfig) {
		g.Resources = append(g.Resources, g.newApigeeResource(
			orgName,
			"google_apigee_addons_config",
			map[string]string{"org": orgID},
		))
	}
	// Control plane access is a hybrid-only singleton; a failed get on
	// other organization types is not an enumeration error.
	if _, err := service.Organizations.GetControlPlaneAccess(orgName + "/controlPlaneAccess").Context(ctx).Do(); err == nil {
		g.Resources = append(g.Resources, g.newApigeeResource(
			orgName,
			"google_apigee_control_plane_access",
			map[string]string{"org_id": orgID},
		))
	} else {
		log.Println(err)
	}
	if sync, err := service.Organizations.GetSyncAuthorization(orgName, &apigee.GoogleCloudApigeeV1GetSyncAuthorizationRequest{}).Context(ctx).Do(); err == nil {
		if len(sync.Identities) > 0 {
			g.Resources = append(g.Resources, g.newApigeeResource(
				orgName,
				"google_apigee_sync_authorization",
				map[string]string{"org_id": orgID},
			))
		}
	} else {
		log.Println(err)
	}

	for _, env := range org.Environments {
		if err := g.loadEnvironment(ctx, service, orgName, env); err != nil {
			return err
		}
	}
	for _, load := range []func(context.Context, *apigee.Service, string) error{
		g.loadApisAndDeployments,
		g.loadAPIProducts,
		g.loadAppGroups,
		g.loadDevelopers,
		g.loadDataCollectors,
		g.loadDatastores,
		g.loadDNSZones,
		g.loadEndpointAttachments,
		g.loadEnvGroups,
		g.loadInstances,
		g.loadSecurityMonitoringConditions,
		g.loadSecurityProfiles,
		g.loadSharedFlows,
		g.loadSpaces,
	} {
		if err := load(ctx, service, orgName); err != nil {
			return err
		}
	}
	return nil
}

var apigeeFlowHookPoints = []string{
	"PreProxyFlowHook",
	"PostProxyFlowHook",
	"PreTargetFlowHook",
	"PostTargetFlowHook",
}

func (g *ApigeeGenerator) loadEnvironment(ctx context.Context, service *apigee.Service, orgName, env string) error {
	orgID := apigeeLastSegment(orgName)
	envName := orgName + "/environments/" + env
	g.Resources = append(g.Resources, g.newApigeeResource(
		envName,
		"google_apigee_environment",
		map[string]string{"org_id": orgName, "name": env},
	))

	if cfg, err := service.Organizations.Environments.GetAddonsConfig(envName + "/addonsConfig").Context(ctx).Do(); err == nil {
		if apigeeAddonsConfigured(cfg) {
			g.Resources = append(g.Resources, g.newApigeeResource(
				envName,
				"google_apigee_environment_addons_config",
				map[string]string{"env_id": envName},
			))
		}
	} else {
		log.Println(err)
	}
	if mask, err := service.Organizations.Environments.GetDebugmask(envName + "/debugmask").Context(ctx).Do(); err == nil {
		if len(mask.FaultJSONPaths) > 0 || len(mask.FaultXPaths) > 0 ||
			len(mask.RequestJSONPaths) > 0 || len(mask.RequestXPaths) > 0 ||
			len(mask.ResponseJSONPaths) > 0 || len(mask.ResponseXPaths) > 0 ||
			len(mask.Variables) > 0 {
			g.Resources = append(g.Resources, g.newApigeeResource(
				envName,
				"google_apigee_environment_debugmask",
				map[string]string{"org_id": orgName, "env_id": envName},
			))
		}
	} else {
		log.Println(err)
	}

	policy, err := service.Organizations.Environments.GetIamPolicy(envName).Context(ctx).Do()
	if err != nil {
		log.Println(err)
	} else {
		g.Resources = appendApigeeEnvironmentIam(g, g.Resources, envName, orgID, policy)
	}

	for _, point := range apigeeFlowHookPoints {
		hookName := envName + "/flowhooks/" + point
		hook, err := service.Organizations.Environments.Flowhooks.Get(hookName).Context(ctx).Do()
		if err != nil {
			log.Println(err)
			continue
		}
		if hook.SharedFlow == "" {
			continue
		}
		g.Resources = append(g.Resources, g.newApigeeResource(
			hookName,
			"google_apigee_flowhook",
			map[string]string{
				"org_id":          orgID,
				"environment":     env,
				"flow_hook_point": point,
				"sharedflow":      hook.SharedFlow,
			},
		))
	}

	return service.Organizations.Environments.SecurityActions.List(envName).Pages(ctx, func(page *apigee.GoogleCloudApigeeV1ListSecurityActionsResponse) error {
		for _, action := range page.SecurityActions {
			g.Resources = append(g.Resources, g.newApigeeResource(
				apigeeResourceID(envName, "securityActions", action.Name),
				"google_apigee_security_action",
				map[string]string{"org_id": orgName, "environment": env},
			))
		}
		return nil
	})
}

// appendApigeeEnvironmentIam mirrors the Cloud Run IAM pattern: the policy
// resource ID is the environment's full name, and bindings and members are
// expanded from the environment's live IAM policy.
func appendApigeeEnvironmentIam(g *ApigeeGenerator, resources []terraformutils.Resource, envName, orgID string, policy *apigee.GoogleIamV1Policy) []terraformutils.Resource {
	parentAttrs := map[string]string{"org_id": orgID, "env_id": envName}
	resources = append(resources, g.newApigeeResource(
		envName,
		"google_apigee_environment_iam_policy",
		parentAttrs,
	))
	for _, binding := range policy.Bindings {
		bindingAttrs := map[string]string{"role": binding.Role}
		for k, v := range parentAttrs {
			bindingAttrs[k] = v
		}
		resources = append(resources, g.newApigeeResource(
			envName,
			"google_apigee_environment_iam_binding",
			bindingAttrs,
		))
		for _, member := range binding.Members {
			memberAttrs := map[string]string{"role": binding.Role, "member": member}
			for k, v := range parentAttrs {
				memberAttrs[k] = v
			}
			resources = append(resources, g.newApigeeResource(
				envName,
				"google_apigee_environment_iam_member",
				memberAttrs,
			))
		}
	}
	return resources
}

func (g *ApigeeGenerator) loadApisAndDeployments(ctx context.Context, service *apigee.Service, orgName string) error {
	orgID := apigeeLastSegment(orgName)
	resp, err := service.Organizations.Apis.List(orgName).Context(ctx).Do()
	if err != nil {
		return err
	}
	for _, proxy := range resp.Proxies {
		g.Resources = append(g.Resources, g.newApigeeResource(
			apigeeResourceID(orgName, "apis", proxy.Name),
			"google_apigee_api",
			map[string]string{"org_id": orgID, "name": proxy.Name},
		))
	}
	// One org-wide deployments list feeds both deployment resources:
	// google_apigee_api_deployment IDs end in "/deployments", while
	// google_apigee_environment_api_revision_deployment IDs end at the
	// revision.
	deployments, err := service.Organizations.Deployments.List(orgName).Context(ctx).Do()
	if err != nil {
		return err
	}
	seen := map[string]bool{}
	for _, deployment := range deployments.Deployments {
		revName := orgName + "/environments/" + deployment.Environment +
			"/apis/" + deployment.ApiProxy + "/revisions/" + deployment.Revision
		if seen[revName] {
			continue
		}
		seen[revName] = true
		attrs := map[string]string{
			"org_id":      orgID,
			"environment": deployment.Environment,
			"proxy_id":    deployment.ApiProxy,
			"api":         deployment.ApiProxy,
			"revision":    deployment.Revision,
		}
		g.Resources = append(g.Resources, g.newApigeeResource(
			revName+"/deployments",
			"google_apigee_api_deployment",
			attrs,
		))
		g.Resources = append(g.Resources, g.newApigeeResource(
			revName,
			"google_apigee_environment_api_revision_deployment",
			attrs,
		))
	}
	return nil
}

func (g *ApigeeGenerator) loadAPIProducts(ctx context.Context, service *apigee.Service, orgName string) error {
	orgID := apigeeLastSegment(orgName)
	resp, err := service.Organizations.Apiproducts.List(orgName).Context(ctx).Do()
	if err != nil {
		return err
	}
	for _, product := range resp.ApiProduct {
		g.Resources = append(g.Resources, g.newApigeeResource(
			apigeeResourceID(orgName, "apiproducts", product.Name),
			"google_apigee_api_product",
			map[string]string{"org_id": orgID, "name": product.Name},
		))
	}
	return nil
}

func (g *ApigeeGenerator) loadAppGroups(ctx context.Context, service *apigee.Service, orgName string) error {
	orgID := apigeeLastSegment(orgName)
	return service.Organizations.Appgroups.List(orgName).Pages(ctx, func(page *apigee.GoogleCloudApigeeV1ListAppGroupsResponse) error {
		for _, group := range page.AppGroups {
			g.Resources = append(g.Resources, g.newApigeeResource(
				apigeeResourceID(orgName, "appgroups", group.Name),
				"google_apigee_app_group",
				map[string]string{"org_id": orgID, "name": group.Name},
			))
		}
		return nil
	})
}

func (g *ApigeeGenerator) loadDevelopers(ctx context.Context, service *apigee.Service, orgName string) error {
	orgID := apigeeLastSegment(orgName)
	resp, err := service.Organizations.Developers.List(orgName).Context(ctx).Do()
	if err != nil {
		return err
	}
	for _, developer := range resp.Developer {
		devName := orgName + "/developers/" + developer.Email
		g.Resources = append(g.Resources, g.newApigeeResource(
			devName,
			"google_apigee_developer",
			map[string]string{"org_id": orgID, "email": developer.Email},
		))
		apps, err := service.Organizations.Developers.Apps.List(devName).Context(ctx).Do()
		if err != nil {
			return err
		}
		for _, app := range apps.App {
			g.Resources = append(g.Resources, g.newApigeeResource(
				apigeeResourceID(devName, "apps", app.Name),
				"google_apigee_developer_app",
				map[string]string{"org_id": orgID, "developer_id": developer.Email, "name": app.Name},
			))
		}
	}
	return nil
}

func (g *ApigeeGenerator) loadDataCollectors(ctx context.Context, service *apigee.Service, orgName string) error {
	orgID := apigeeLastSegment(orgName)
	return service.Organizations.Datacollectors.List(orgName).Pages(ctx, func(page *apigee.GoogleCloudApigeeV1ListDataCollectorsResponse) error {
		for _, collector := range page.DataCollectors {
			g.Resources = append(g.Resources, g.newApigeeResource(
				apigeeResourceID(orgName, "datacollectors", collector.Name),
				"google_apigee_data_collector",
				map[string]string{"org_id": orgID, "name": collector.Name},
			))
		}
		return nil
	})
}

func (g *ApigeeGenerator) loadDatastores(ctx context.Context, service *apigee.Service, orgName string) error {
	resp, err := service.Organizations.Analytics.Datastores.List(orgName).Context(ctx).Do()
	if err != nil {
		return err
	}
	for _, datastore := range resp.Datastores {
		// The datastore carries no name field; the server-assigned UUID
		// is the last segment of its self link.
		id := orgName + "/analytics/datastores/" + apigeeLastSegment(datastore.Self)
		g.Resources = append(g.Resources, g.newApigeeResource(
			id,
			"google_apigee_datastore",
			map[string]string{"org_id": orgName},
		))
	}
	return nil
}

func (g *ApigeeGenerator) loadDNSZones(ctx context.Context, service *apigee.Service, orgName string) error {
	orgID := apigeeLastSegment(orgName)
	return service.Organizations.DnsZones.List(orgName).Pages(ctx, func(page *apigee.GoogleCloudApigeeV1ListDnsZonesResponse) error {
		for _, zone := range page.DnsZones {
			g.Resources = append(g.Resources, g.newApigeeResource(
				apigeeResourceID(orgName, "dnsZones", zone.Name),
				"google_apigee_dns_zone",
				map[string]string{"org_id": orgID, "name": zone.Name},
			))
		}
		return nil
	})
}

func (g *ApigeeGenerator) loadEndpointAttachments(ctx context.Context, service *apigee.Service, orgName string) error {
	orgID := apigeeLastSegment(orgName)
	return service.Organizations.EndpointAttachments.List(orgName).Pages(ctx, func(page *apigee.GoogleCloudApigeeV1ListEndpointAttachmentsResponse) error {
		for _, attachment := range page.EndpointAttachments {
			g.Resources = append(g.Resources, g.newApigeeResource(
				apigeeResourceID(orgName, "endpointAttachments", attachment.Name),
				"google_apigee_endpoint_attachment",
				map[string]string{"org_id": orgID, "name": attachment.Name},
			))
		}
		return nil
	})
}

func (g *ApigeeGenerator) loadEnvGroups(ctx context.Context, service *apigee.Service, orgName string) error {
	orgID := apigeeLastSegment(orgName)
	return service.Organizations.Envgroups.List(orgName).Pages(ctx, func(page *apigee.GoogleCloudApigeeV1ListEnvironmentGroupsResponse) error {
		for _, group := range page.EnvironmentGroups {
			groupName := apigeeResourceID(orgName, "envgroups", group.Name)
			g.Resources = append(g.Resources, g.newApigeeResource(
				groupName,
				"google_apigee_envgroup",
				map[string]string{"org_id": orgID, "name": group.Name},
			))
			err := service.Organizations.Envgroups.Attachments.List(groupName).Pages(ctx, func(page *apigee.GoogleCloudApigeeV1ListEnvironmentGroupAttachmentsResponse) error {
				for _, attachment := range page.EnvironmentGroupAttachments {
					g.Resources = append(g.Resources, g.newApigeeResource(
						apigeeResourceID(groupName, "attachments", attachment.Name),
						"google_apigee_envgroup_attachment",
						map[string]string{"org_id": orgID, "envgroup_id": group.Name},
					))
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (g *ApigeeGenerator) loadInstances(ctx context.Context, service *apigee.Service, orgName string) error {
	orgID := apigeeLastSegment(orgName)
	return service.Organizations.Instances.List(orgName).Pages(ctx, func(page *apigee.GoogleCloudApigeeV1ListInstancesResponse) error {
		for _, instance := range page.Instances {
			instanceName := apigeeResourceID(orgName, "instances", instance.Name)
			g.Resources = append(g.Resources, g.newApigeeResource(
				instanceName,
				"google_apigee_instance",
				map[string]string{"org_id": orgID, "name": instance.Name},
			))
			err := service.Organizations.Instances.Attachments.List(instanceName).Pages(ctx, func(page *apigee.GoogleCloudApigeeV1ListInstanceAttachmentsResponse) error {
				for _, attachment := range page.Attachments {
					g.Resources = append(g.Resources, g.newApigeeResource(
						apigeeResourceID(instanceName, "attachments", attachment.Name),
						"google_apigee_instance_attachment",
						map[string]string{"org_id": orgID, "instance_id": instance.Name},
					))
				}
				return nil
			})
			if err != nil {
				return err
			}
			err = service.Organizations.Instances.NatAddresses.List(instanceName).Pages(ctx, func(page *apigee.GoogleCloudApigeeV1ListNatAddressesResponse) error {
				for _, nat := range page.NatAddresses {
					g.Resources = append(g.Resources, g.newApigeeResource(
						apigeeResourceID(instanceName, "natAddresses", nat.Name),
						"google_apigee_nat_address",
						map[string]string{"org_id": orgID, "instance_id": instance.Name, "name": nat.Name},
					))
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (g *ApigeeGenerator) loadSecurityMonitoringConditions(ctx context.Context, service *apigee.Service, orgName string) error {
	orgID := apigeeLastSegment(orgName)
	return service.Organizations.SecurityMonitoringConditions.List(orgName).Pages(ctx, func(page *apigee.GoogleCloudApigeeV1ListSecurityMonitoringConditionsResponse) error {
		for _, condition := range page.SecurityMonitoringConditions {
			g.Resources = append(g.Resources, g.newApigeeResource(
				apigeeResourceID(orgName, "securityMonitoringConditions", condition.Name),
				"google_apigee_security_monitoring_condition",
				map[string]string{"org_id": orgID, "name": condition.Name},
			))
		}
		return nil
	})
}

func (g *ApigeeGenerator) loadSecurityProfiles(ctx context.Context, service *apigee.Service, orgName string) error {
	orgID := apigeeLastSegment(orgName)
	return service.Organizations.SecurityProfiles.List(orgName).Pages(ctx, func(page *apigee.GoogleCloudApigeeV1ListSecurityProfilesResponse) error {
		for _, profile := range page.SecurityProfiles {
			// Profile names may carry a "@revision" suffix; the
			// provider resource addresses the profile itself.
			name := strings.Split(profile.Name, "@")[0]
			g.Resources = append(g.Resources, g.newApigeeResource(
				apigeeResourceID(orgName, "securityProfiles", name),
				"google_apigee_security_profile_v2",
				map[string]string{"org_id": orgID, "profile_id": apigeeLastSegment(name)},
			))
		}
		return nil
	})
}

func (g *ApigeeGenerator) loadSharedFlows(ctx context.Context, service *apigee.Service, orgName string) error {
	orgID := apigeeLastSegment(orgName)
	resp, err := service.Organizations.Sharedflows.List(orgName).Context(ctx).Do()
	if err != nil {
		return err
	}
	for _, flow := range resp.SharedFlows {
		flowName := apigeeResourceID(orgName, "sharedflows", flow.Name)
		g.Resources = append(g.Resources, g.newApigeeResource(
			flowName,
			"google_apigee_sharedflow",
			map[string]string{"org_id": orgID, "name": flow.Name},
		))
		deployments, err := service.Organizations.Sharedflows.Deployments.List(flowName).Context(ctx).Do()
		if err != nil {
			return err
		}
		seen := map[string]bool{}
		for _, deployment := range deployments.Deployments {
			id := orgName + "/environments/" + deployment.Environment +
				"/sharedflows/" + flow.Name + "/revisions/" + deployment.Revision + "/deployments"
			if seen[id] {
				continue
			}
			seen[id] = true
			g.Resources = append(g.Resources, g.newApigeeResource(
				id,
				"google_apigee_sharedflow_deployment",
				map[string]string{
					"org_id":        orgID,
					"environment":   deployment.Environment,
					"sharedflow_id": flow.Name,
					"revision":      deployment.Revision,
				},
			))
		}
	}
	return nil
}

func (g *ApigeeGenerator) loadSpaces(ctx context.Context, service *apigee.Service, orgName string) error {
	orgID := apigeeLastSegment(orgName)
	return service.Organizations.Spaces.List(orgName).Pages(ctx, func(page *apigee.GoogleCloudApigeeV1ListSpacesResponse) error {
		for _, space := range page.Spaces {
			g.Resources = append(g.Resources, g.newApigeeResource(
				apigeeResourceID(orgName, "spaces", space.Name),
				"google_apigee_space",
				map[string]string{"org_id": orgID, "name": space.Name},
			))
		}
		return nil
	})
}
