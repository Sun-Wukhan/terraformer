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
	appenginev1 "google.golang.org/api/appengine/v1"
	"google.golang.org/api/compute/v1"
	iapv1 "google.golang.org/api/iap/v1"
	runv2 "google.golang.org/api/run/v2"
)

var iapAllowEmptyValues = []string{""}

var iapAdditionalFields = map[string]interface{}{}

// IapGenerator is the IAP IAM-attachment pass: for every parent the
// IAP API can address it reads the live policy through the iap/v1
// getIamPolicy call and expands it into the provider's
// _iam_policy/_iam_binding/_iam_member trio (expandIamPolicy), with
// the parent IDs in the provider's documented formats:
//
//   - project web policy: projects/{p}/iap_web
//   - web type compute: projects/{p}/iap_web/compute
//   - web type App Engine: projects/{p}/iap_web/appengine-{appId}
//     (the App Engine app ID is the project ID), plus per-service and
//     per-version policies under it, enumerated via the App Engine API
//   - web backend services (global and regional) and forwarding-rule
//     services (global and regional), enumerated via Compute:
//     projects/{p}/iap_web/compute/services/{name},
//     .../compute/regions/{region}/services/{name},
//     .../iap_web/forwarding_rule/services/{name},
//     .../iap_web/forwarding_rule/regions/{region}/services/{name}
//   - web Cloud Run services in the configured region, enumerated via
//     the Cloud Run API: projects/{p}/iap_web/cloud_run-{region}/services/{name}
//   - tunnel policy: projects/{p}/iap_tunnel, and per-instance tunnel
//     policies for every Compute instance in the configured region's
//     zones: projects/{p}/iap_tunnel/zones/{zone}/instances/{name}
//   - tunnel destination groups in the configured region (iap/v1 list)
//     and their IAM trios, addressed by the dest group resource name
//   - the project IAP settings singleton: projects/{p}/iap_settings
//
// Deliberately not generated:
//   - google_iap_agent_registry_iam, google_iap_agent_registry_agent_iam,
//     google_iap_agent_registry_endpoint_iam and
//     google_iap_agent_registry_mcp_server_iam: the Agent Registry
//     surface has no service in the pinned iap/v1 or iap/v1beta1
//     clients (only Brands, IapTunnel and DestGroups exist) and no
//     list endpoint for registry collections is present in the IAP
//     discovery document used by this package, so registry parents
//     cannot be enumerated.
//   - google_iap_location_web_iam: the location-scoped web IAM
//     variant is addressed per location (projects/{p}/locations/{loc}/iap_web);
//     the iap/v1 client exposes no location list for a project, so
//     the locations holding a distinct web policy cannot be
//     discovered. The global web policy is generated.
type IapGenerator struct {
	GCPService
}

func (g *IapGenerator) InitResources() error {
	ctx := context.Background()
	iapSvc, err := iapv1.NewService(ctx)
	if err != nil {
		return err
	}
	aeSvc, err := appenginev1.NewService(ctx)
	if err != nil {
		log.Println(err)
		aeSvc = nil
	}
	computeSvc, err := compute.NewService(ctx)
	if err != nil {
		log.Println(err)
		computeSvc = nil
	}
	runSvc, err := runv2.NewService(ctx)
	if err != nil {
		log.Println(err)
		runSvc = nil
	}
	return g.initResources(ctx, iapSvc, aeSvc, computeSvc, runSvc)
}

func (g *IapGenerator) newIapResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
	if attrs == nil {
		attrs = map[string]string{}
	}
	parts := strings.Split(id, "/")
	return terraformutils.NewResource(id, parts[len(parts)-1], resourceType, g.ProviderName, attrs, iapAllowEmptyValues, iapAdditionalFields)
}

// expandIapPolicy reads one parent's IAP policy and expands the trio
// named literally at the call site via named arguments.
func (g *IapGenerator) expandIapPolicy(ctx context.Context, iapSvc *iapv1.Service, parentID, policyType, bindingType, memberType string, attrs map[string]string) {
	if iapSvc == nil {
		return
	}
	policy, err := iapSvc.V1.GetIamPolicy(parentID, &iapv1.GetIamPolicyRequest{}).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		return
	}
	bindings := []iamBinding{}
	for _, b := range policy.Bindings {
		bindings = append(bindings, iamBinding{Role: b.Role, Members: b.Members})
	}
	g.Resources = append(g.Resources, expandIamPolicy(bindings, parentID, policyType, bindingType, memberType, attrs, g.newIapResource)...)
}

func (g *IapGenerator) initResources(ctx context.Context, iapSvc *iapv1.Service, aeSvc *appenginev1.APIService, computeSvc *compute.Service, runSvc *runv2.Service) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region)
	regionName := region.Name
	zones := computeZoneNames(region)
	base := "projects/" + project
	projAttrs := map[string]string{"project": project}

	// Project-level singletons.
	g.expandIapPolicy(ctx, iapSvc, base+"/iap_web",
		"google_iap_web_iam_policy", "google_iap_web_iam_binding", "google_iap_web_iam_member", projAttrs)
	g.expandIapPolicy(ctx, iapSvc, base+"/iap_web/compute",
		"google_iap_web_type_compute_iam_policy", "google_iap_web_type_compute_iam_binding", "google_iap_web_type_compute_iam_member", projAttrs)
	g.expandIapPolicy(ctx, iapSvc, base+"/iap_web/appengine-"+project,
		"google_iap_web_type_app_engine_iam_policy", "google_iap_web_type_app_engine_iam_binding", "google_iap_web_type_app_engine_iam_member", projAttrs)
	g.expandIapPolicy(ctx, iapSvc, base+"/iap_tunnel",
		"google_iap_tunnel_iam_policy", "google_iap_tunnel_iam_binding", "google_iap_tunnel_iam_member", projAttrs)
	if iapSvc != nil {
		if settings, err := iapSvc.V1.GetIapSettings(base + "/iap_settings").Context(ctx).Do(); err == nil && settings != nil && settings.Name != "" {
			g.Resources = append(g.Resources, g.newIapResource(settings.Name, "google_iap_settings", projAttrs))
		} else if err != nil {
			log.Println(err)
		}
	}

	g.loadIapAppEngine(ctx, iapSvc, aeSvc, project)
	g.loadIapComputeParents(ctx, iapSvc, computeSvc, project, regionName, zones)
	g.loadIapCloudRun(ctx, iapSvc, runSvc, project, regionName)
	g.loadIapDestGroups(ctx, iapSvc, project, regionName)
	return nil
}

func (g *IapGenerator) loadIapAppEngine(ctx context.Context, iapSvc *iapv1.Service, aeSvc *appenginev1.APIService, project string) {
	if aeSvc == nil || iapSvc == nil {
		return
	}
	err := aeSvc.Apps.Services.List(project).Pages(ctx, func(page *appenginev1.ListServicesResponse) error {
		for _, svc := range page.Services {
			svcID := svc.Id
			if svcID == "" {
				svcID = svc.Name[strings.LastIndex(svc.Name, "/")+1:]
			}
			parent := "projects/" + project + "/iap_web/appengine-" + project + "/services/" + svcID
			g.expandIapPolicy(ctx, iapSvc, parent,
				"google_iap_app_engine_service_iam_policy", "google_iap_app_engine_service_iam_binding", "google_iap_app_engine_service_iam_member",
				map[string]string{"project": project, "app_id": project, "service": svcID})
			_ = aeSvc.Apps.Services.Versions.List(project, svcID).Pages(ctx, func(vpage *appenginev1.ListVersionsResponse) error {
				for _, version := range vpage.Versions {
					verID := version.Id
					if verID == "" {
						verID = version.Name[strings.LastIndex(version.Name, "/")+1:]
					}
					g.expandIapPolicy(ctx, iapSvc, parent+"/versions/"+verID,
						"google_iap_app_engine_version_iam_policy", "google_iap_app_engine_version_iam_binding", "google_iap_app_engine_version_iam_member",
						map[string]string{"project": project, "app_id": project, "service": svcID, "version_id": verID})
				}
				return nil
			})
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (g *IapGenerator) loadIapComputeParents(ctx context.Context, iapSvc *iapv1.Service, computeSvc *compute.Service, project, regionName string, zones []string) {
	if computeSvc == nil || iapSvc == nil {
		return
	}
	_ = computeSvc.BackendServices.List(project).Pages(ctx, func(page *compute.BackendServiceList) error {
		for _, item := range page.Items {
			g.expandIapPolicy(ctx, iapSvc, "projects/"+project+"/iap_web/compute/services/"+item.Name,
				"google_iap_web_backend_service_iam_policy", "google_iap_web_backend_service_iam_binding", "google_iap_web_backend_service_iam_member",
				map[string]string{"project": project, "web_backend_service": item.Name})
		}
		return nil
	})
	_ = computeSvc.RegionBackendServices.List(project, regionName).Pages(ctx, func(page *compute.BackendServiceList) error {
		for _, item := range page.Items {
			g.expandIapPolicy(ctx, iapSvc, "projects/"+project+"/iap_web/compute/regions/"+regionName+"/services/"+item.Name,
				"google_iap_web_region_backend_service_iam_policy", "google_iap_web_region_backend_service_iam_binding", "google_iap_web_region_backend_service_iam_member",
				map[string]string{"project": project, "region": regionName, "web_region_backend_service": item.Name})
		}
		return nil
	})
	_ = computeSvc.GlobalForwardingRules.List(project).Pages(ctx, func(page *compute.ForwardingRuleList) error {
		for _, item := range page.Items {
			g.expandIapPolicy(ctx, iapSvc, "projects/"+project+"/iap_web/forwarding_rule/services/"+item.Name,
				"google_iap_web_forwarding_rule_service_iam_policy", "google_iap_web_forwarding_rule_service_iam_binding", "google_iap_web_forwarding_rule_service_iam_member",
				map[string]string{"project": project, "web_forwarding_rule_service": item.Name})
		}
		return nil
	})
	_ = computeSvc.ForwardingRules.List(project, regionName).Pages(ctx, func(page *compute.ForwardingRuleList) error {
		for _, item := range page.Items {
			g.expandIapPolicy(ctx, iapSvc, "projects/"+project+"/iap_web/forwarding_rule/regions/"+regionName+"/services/"+item.Name,
				"google_iap_web_region_forwarding_rule_service_iam_policy", "google_iap_web_region_forwarding_rule_service_iam_binding", "google_iap_web_region_forwarding_rule_service_iam_member",
				map[string]string{"project": project, "region": regionName, "web_region_forwarding_rule_service": item.Name})
		}
		return nil
	})
	for _, zone := range zones {
		zone := zone
		_ = computeSvc.Instances.List(project, zone).Pages(ctx, func(page *compute.InstanceList) error {
			for _, item := range page.Items {
				g.expandIapPolicy(ctx, iapSvc, "projects/"+project+"/iap_tunnel/zones/"+zone+"/instances/"+item.Name,
					"google_iap_tunnel_instance_iam_policy", "google_iap_tunnel_instance_iam_binding", "google_iap_tunnel_instance_iam_member",
					map[string]string{"project": project, "zone": zone, "instance_name": item.Name})
			}
			return nil
		})
	}
}

func (g *IapGenerator) loadIapCloudRun(ctx context.Context, iapSvc *iapv1.Service, runSvc *runv2.Service, project, regionName string) {
	if runSvc == nil || iapSvc == nil || regionName == "" {
		return
	}
	parent := "projects/" + project + "/locations/" + regionName
	err := runSvc.Projects.Locations.Services.List(parent).Pages(ctx, func(page *runv2.GoogleCloudRunV2ListServicesResponse) error {
		for _, svc := range page.Services {
			name := svc.Name[strings.LastIndex(svc.Name, "/")+1:]
			g.expandIapPolicy(ctx, iapSvc, "projects/"+project+"/iap_web/cloud_run-"+regionName+"/services/"+name,
				"google_iap_web_cloud_run_service_iam_policy", "google_iap_web_cloud_run_service_iam_binding", "google_iap_web_cloud_run_service_iam_member",
				map[string]string{"project": project, "location": regionName, "cloud_run_service_name": name})
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (g *IapGenerator) loadIapDestGroups(ctx context.Context, iapSvc *iapv1.Service, project, regionName string) {
	if iapSvc == nil || regionName == "" {
		return
	}
	parent := "projects/" + project + "/locations/" + regionName
	err := iapSvc.Projects.IapTunnel.Locations.DestGroups.List(parent).Pages(ctx, func(page *iapv1.ListTunnelDestGroupsResponse) error {
		for _, group := range page.TunnelDestGroups {
			g.Resources = append(g.Resources, g.newIapResource(group.Name, "google_iap_tunnel_dest_group", map[string]string{"project": project, "region": regionName}))
			groupID := group.Name[strings.LastIndex(group.Name, "/")+1:]
			g.expandIapPolicy(ctx, iapSvc, group.Name,
				"google_iap_tunnel_dest_group_iam_policy", "google_iap_tunnel_dest_group_iam_binding", "google_iap_tunnel_dest_group_iam_member",
				map[string]string{"project": project, "region": regionName, "dest_group": groupID})
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}
