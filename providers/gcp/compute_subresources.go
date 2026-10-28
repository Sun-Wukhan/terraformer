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
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/compute/v1"
)

var computeSubresourcesAllowEmptyValues = []string{""}

var computeSubresourcesAdditionalFields = map[string]interface{}{}

type ComputeSubresourcesGenerator struct {
	GCPService
}

// InitResources enumerates the Compute Engine sub-resources that live
// inside (or directly under) resources other generators already emit,
// following the routerSubresources precedent:
//
//   - Network peerings, embedded in the network resource.
//   - Instance group named ports and group memberships, from the
//     zonal instance groups of the configured region.
//   - Per-instance configs and MIG resize requests, listed per
//     instance group manager (zonal and regional; the regional resize
//     request list is raw REST — the pinned client has no service for
//     it, but the GA discovery document has the endpoint).
//   - Resource policy attachments, from the resourcePolicies fields
//     of zonal disks, regional disks and instances.
//   - Project-level singletons read from the project resource itself:
//     common instance metadata, the Cloud Armor tier, the default
//     network tier, and the Shared VPC host/service project relations
//     (the service projects come from getXpnResources).
//   - The snapshot settings singleton and the per-zone instance
//     settings singletons.
//   - The snapshot recycle bin policies, over raw REST against the
//     beta endpoint: get/patch singletons that exist only in the beta
//     compute API, which is what backs the provider's GA resources.
//
// Deliberately not generated:
//   - google_compute_attached_disk: an instance's disks are fields of
//     the google_compute_instance resource already emitted; the
//     attachment has no identity separate from its parent (the same
//     position Batch 1 took for router NAT addresses).
//   - google_compute_bulk_per_instance_config: a bulk view over the
//     same per-instance config objects emitted individually here;
//     its provider ID is the instance group manager itself.
//   - google_compute_network_peering_routes_config: the same API
//     object and provider ID as the google_compute_network_peering
//     emitted here, managing an overlapping facet (custom route
//     import/export) of that peering.
//   - google_compute_project_metadata_item: a per-key facet of the
//     project's common instance metadata; the provider documents
//     that it conflicts with the whole-set
//     google_compute_project_metadata emitted here, and the keys
//     remain part of that resource.
func (g *ComputeSubresourcesGenerator) InitResources() error {
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

func (g *ComputeSubresourcesGenerator) initResources(ctx context.Context, computeService *compute.Service, crmService *cloudresourcemanager.Service, restClient *http.Client) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region)
	regionName := region.Name
	zones := computeZoneNames(region)

	g.loadNetworkPeerings(ctx, computeService, project)
	for _, zone := range zones {
		g.loadInstanceGroupFacets(ctx, computeService, project, zone)
		g.loadZonalMigFacets(ctx, computeService, project, zone)
		g.loadDiskPolicyAttachments(ctx, computeService, project, zone)
		g.loadInstancePolicyAttachments(ctx, computeService, project, zone)
		g.loadInstanceSettings(ctx, computeService, project, zone)
	}
	g.loadRegionalMigFacets(ctx, computeService, restClient, project, regionName)
	g.loadRegionDiskPolicyAttachments(ctx, computeService, project, regionName)
	g.loadProjectSingletons(ctx, computeService, project)
	g.loadRecycleBinPolicies(ctx, crmService, restClient, project)
	return nil
}

func (g *ComputeSubresourcesGenerator) newComputeSubresource(id, name, resourceType string, attrs map[string]string) terraformutils.Resource {
	if attrs == nil {
		attrs = map[string]string{}
	}
	return terraformutils.NewResource(
		id,
		name,
		resourceType,
		g.ProviderName,
		attrs,
		computeSubresourcesAllowEmptyValues,
		computeSubresourcesAdditionalFields,
	)
}

func lastSegment(urlOrName string) string {
	return urlOrName[strings.LastIndex(urlOrName, "/")+1:]
}

func (g *ComputeSubresourcesGenerator) loadNetworkPeerings(ctx context.Context, computeService *compute.Service, project string) {
	err := computeService.Networks.List(project).Pages(ctx, func(page *compute.NetworkList) error {
		for _, network := range page.Items {
			for _, peering := range network.Peerings {
				id := project + "/" + network.Name + "/" + peering.Name
				attrs := map[string]string{"name": peering.Name, "network": network.Name, "project": project}
				g.Resources = append(g.Resources, g.newComputeSubresource(
					id, peering.Name, "google_compute_network_peering", attrs))
			}
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (g *ComputeSubresourcesGenerator) loadInstanceGroupFacets(ctx context.Context, computeService *compute.Service, project, zone string) {
	err := computeService.InstanceGroups.List(project, zone).Pages(ctx, func(page *compute.InstanceGroupList) error {
		for _, group := range page.Items {
			for _, namedPort := range group.NamedPorts {
				id := project + "/" + zone + "/" + group.Name + "/" + fmt.Sprint(namedPort.Port) + "/" + namedPort.Name
				attrs := map[string]string{
					"name":  namedPort.Name,
					"port":  fmt.Sprint(namedPort.Port),
					"group": group.Name,
					"zone":  zone,
				}
				g.Resources = append(g.Resources, g.newComputeSubresource(
					id, namedPort.Name, "google_compute_instance_group_named_port", attrs))
			}
			g.loadInstanceGroupMemberships(ctx, computeService, project, zone, group.Name)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (g *ComputeSubresourcesGenerator) loadInstanceGroupMemberships(ctx context.Context, computeService *compute.Service, project, zone, group string) {
	response, err := computeService.InstanceGroups.ListInstances(project, zone, group, &compute.InstanceGroupsListInstancesRequest{}).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		return
	}
	for _, member := range response.Items {
		instance := lastSegment(member.Instance)
		id := project + "/" + zone + "/" + group + "/" + instance
		attrs := map[string]string{"instance_group": group, "instance": instance, "zone": zone, "project": project}
		g.Resources = append(g.Resources, g.newComputeSubresource(
			id, instance, "google_compute_instance_group_membership", attrs))
	}
}

func (g *ComputeSubresourcesGenerator) loadZonalMigFacets(ctx context.Context, computeService *compute.Service, project, zone string) {
	err := computeService.InstanceGroupManagers.List(project, zone).Pages(ctx, func(page *compute.InstanceGroupManagerList) error {
		for _, mig := range page.Items {
			g.loadZonalPerInstanceConfigs(ctx, computeService, project, zone, mig.Name)
			g.loadZonalResizeRequests(ctx, computeService, project, zone, mig.Name)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (g *ComputeSubresourcesGenerator) loadZonalPerInstanceConfigs(ctx context.Context, computeService *compute.Service, project, zone, mig string) {
	response, err := computeService.InstanceGroupManagers.ListPerInstanceConfigs(project, zone, mig).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		return
	}
	for _, config := range response.Items {
		id := project + "/" + zone + "/" + mig + "/" + config.Name
		attrs := map[string]string{"name": config.Name, "instance_group_manager": mig, "zone": zone, "project": project}
		g.Resources = append(g.Resources, g.newComputeSubresource(
			id, config.Name, "google_compute_per_instance_config", attrs))
	}
}

func (g *ComputeSubresourcesGenerator) loadZonalResizeRequests(ctx context.Context, computeService *compute.Service, project, zone, mig string) {
	err := computeService.InstanceGroupManagerResizeRequests.List(project, zone, mig).Pages(ctx, func(page *compute.InstanceGroupManagerResizeRequestsListResponse) error {
		for _, request := range page.Items {
			id := project + "/" + zone + "/" + mig + "/" + request.Name
			attrs := map[string]string{"name": request.Name, "instance_group_manager": mig, "zone": zone, "project": project}
			g.Resources = append(g.Resources, g.newComputeSubresource(
				id, request.Name, "google_compute_resize_request", attrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (g *ComputeSubresourcesGenerator) loadRegionalMigFacets(ctx context.Context, computeService *compute.Service, restClient *http.Client, project, region string) {
	err := computeService.RegionInstanceGroupManagers.List(project, region).Pages(ctx, func(page *compute.RegionInstanceGroupManagerList) error {
		for _, mig := range page.Items {
			g.loadRegionalPerInstanceConfigs(ctx, computeService, project, region, mig.Name)
			g.loadRegionalResizeRequests(ctx, restClient, project, region, mig.Name)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (g *ComputeSubresourcesGenerator) loadRegionalPerInstanceConfigs(ctx context.Context, computeService *compute.Service, project, region, mig string) {
	response, err := computeService.RegionInstanceGroupManagers.ListPerInstanceConfigs(project, region, mig).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		return
	}
	for _, config := range response.Items {
		id := project + "/" + region + "/" + mig + "/" + config.Name
		attrs := map[string]string{"name": config.Name, "region_instance_group_manager": mig, "region": region, "project": project}
		g.Resources = append(g.Resources, g.newComputeSubresource(
			id, config.Name, "google_compute_region_per_instance_config", attrs))
	}
}

// loadRegionalResizeRequests reads the regional MIG resize requests
// over raw REST: the GA discovery document has the endpoint but the
// pinned generated client has no service for it.
func (g *ComputeSubresourcesGenerator) loadRegionalResizeRequests(ctx context.Context, restClient *http.Client, project, region, mig string) {
	url := "https://compute.googleapis.com/compute/v1/projects/" + project + "/regions/" + region + "/instanceGroupManagers/" + mig + "/resizeRequests"
	items, err := listComputeRest(ctx, restClient, url)
	if err != nil {
		log.Println(err)
		return
	}
	for _, item := range items {
		id := project + "/" + region + "/" + mig + "/" + item.Name
		attrs := map[string]string{"name": item.Name, "instance_group_manager": mig, "region": region, "project": project}
		g.Resources = append(g.Resources, g.newComputeSubresource(
			id, item.Name, "google_compute_region_resize_request", attrs))
	}
}

func (g *ComputeSubresourcesGenerator) loadDiskPolicyAttachments(ctx context.Context, computeService *compute.Service, project, zone string) {
	err := computeService.Disks.List(project, zone).Pages(ctx, func(page *compute.DiskList) error {
		for _, disk := range page.Items {
			for _, policyURL := range disk.ResourcePolicies {
				policy := lastSegment(policyURL)
				id := project + "/" + zone + "/" + disk.Name + "/" + policy
				attrs := map[string]string{"name": policy, "disk": disk.Name, "zone": zone, "project": project}
				g.Resources = append(g.Resources, g.newComputeSubresource(
					id, policy, "google_compute_disk_resource_policy_attachment", attrs))
			}
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (g *ComputeSubresourcesGenerator) loadRegionDiskPolicyAttachments(ctx context.Context, computeService *compute.Service, project, region string) {
	err := computeService.RegionDisks.List(project, region).Pages(ctx, func(page *compute.DiskList) error {
		for _, disk := range page.Items {
			for _, policyURL := range disk.ResourcePolicies {
				policy := lastSegment(policyURL)
				id := project + "/" + region + "/" + disk.Name + "/" + policy
				attrs := map[string]string{"name": policy, "disk": disk.Name, "region": region, "project": project}
				g.Resources = append(g.Resources, g.newComputeSubresource(
					id, policy, "google_compute_region_disk_resource_policy_attachment", attrs))
			}
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (g *ComputeSubresourcesGenerator) loadInstancePolicyAttachments(ctx context.Context, computeService *compute.Service, project, zone string) {
	err := computeService.Instances.List(project, zone).Pages(ctx, func(page *compute.InstanceList) error {
		for _, instance := range page.Items {
			for _, policyURL := range instance.ResourcePolicies {
				policy := lastSegment(policyURL)
				id := project + "/" + zone + "/" + instance.Name + "/" + policy
				attrs := map[string]string{"name": policy, "instance": instance.Name, "zone": zone, "project": project}
				g.Resources = append(g.Resources, g.newComputeSubresource(
					id, policy, "google_compute_resource_policy_attachment", attrs))
			}
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (g *ComputeSubresourcesGenerator) loadInstanceSettings(ctx context.Context, computeService *compute.Service, project, zone string) {
	if _, err := computeService.InstanceSettings.Get(project, zone).Context(ctx).Do(); err != nil {
		log.Println(err)
		return
	}
	id := "projects/" + project + "/zones/" + zone + "/instanceSettings"
	attrs := map[string]string{"project": project, "zone": zone}
	g.Resources = append(g.Resources, g.newComputeSubresource(
		id, "instanceSettings", "google_compute_instance_settings", attrs))
}

func (g *ComputeSubresourcesGenerator) loadProjectSingletons(ctx context.Context, computeService *compute.Service, project string) {
	projectResource, err := computeService.Projects.Get(project).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		return
	}
	if projectResource.CommonInstanceMetadata != nil {
		g.Resources = append(g.Resources, g.newComputeSubresource(
			project, project, "google_compute_project_metadata",
			map[string]string{"project": project}))
	}
	if projectResource.CloudArmorTier != "" {
		g.Resources = append(g.Resources, g.newComputeSubresource(
			project, project, "google_compute_project_cloud_armor_tier",
			map[string]string{"project": project, "cloud_armor_tier": projectResource.CloudArmorTier}))
	}
	if projectResource.DefaultNetworkTier != "" {
		g.Resources = append(g.Resources, g.newComputeSubresource(
			project, project, "google_compute_project_default_network_tier",
			map[string]string{"project": project, "default_network_tier": projectResource.DefaultNetworkTier}))
	}
	if projectResource.XpnProjectStatus == "HOST" {
		g.Resources = append(g.Resources, g.newComputeSubresource(
			project, project, "google_compute_shared_vpc_host_project",
			map[string]string{"project": project}))
	}
	if _, err := computeService.SnapshotSettings.Get(project).Context(ctx).Do(); err == nil {
		id := "projects/" + project + "/global/snapshotSettings"
		g.Resources = append(g.Resources, g.newComputeSubresource(
			id, "snapshotSettings", "google_compute_snapshot_settings",
			map[string]string{"project": project}))
	} else {
		log.Println(err)
	}
	g.loadSharedVpcServiceProjects(ctx, computeService, project)
}

func (g *ComputeSubresourcesGenerator) loadSharedVpcServiceProjects(ctx context.Context, computeService *compute.Service, project string) {
	response, err := computeService.Projects.GetXpnResources(project).Context(ctx).Do()
	if err != nil {
		// getXpnResources only succeeds for Shared VPC host projects.
		log.Println(err)
		return
	}
	for _, resource := range response.Resources {
		id := project + "/" + resource.Id
		attrs := map[string]string{"host_project": project, "service_project": resource.Id}
		g.Resources = append(g.Resources, g.newComputeSubresource(
			id, resource.Id, "google_compute_shared_vpc_service_project", attrs))
	}
}

// loadRecycleBinPolicies reads the snapshot recycle bin policy
// singletons over raw REST against the beta compute endpoint (the
// only surface these get/patch singletons have). A non-2xx response
// (for example when no policy exists) is logged and nothing is
// emitted for that scope.
func (g *ComputeSubresourcesGenerator) loadRecycleBinPolicies(ctx context.Context, crmService *cloudresourcemanager.Service, restClient *http.Client, project string) {
	projectURL := "https://compute.googleapis.com/compute/beta/projects/" + project + "/global/snapshotRecycleBinPolicy"
	if _, err := restDo(ctx, restClient, http.MethodGet, projectURL, nil); err == nil {
		g.Resources = append(g.Resources, g.newComputeSubresource(
			"projects/"+project+"/global/snapshotRecycleBinPolicy", "snapshotRecycleBinPolicy",
			"google_compute_snapshot_recycle_bin_policy", map[string]string{"project": project}))
	} else {
		log.Println(err)
	}
	if crmService == nil {
		return
	}
	ancestry, err := crmService.Projects.GetAncestry(project, &cloudresourcemanager.GetAncestryRequest{}).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		return
	}
	orgID := ""
	for _, ancestor := range ancestry.Ancestor {
		if ancestor.ResourceId != nil && ancestor.ResourceId.Type == "organization" {
			orgID = ancestor.ResourceId.Id
		}
	}
	if orgID == "" {
		return
	}
	orgURL := "https://compute.googleapis.com/compute/beta/organizations/" + orgID + "/global/snapshotRecycleBinPolicy"
	if _, err := restDo(ctx, restClient, http.MethodGet, orgURL, nil); err == nil {
		g.Resources = append(g.Resources, g.newComputeSubresource(
			"organizations/"+orgID+"/global/snapshotRecycleBinPolicy", "snapshotRecycleBinPolicy",
			"google_compute_organization_snapshot_recycle_bin_policy", map[string]string{"organization": orgID}))
	} else {
		log.Println(err)
	}
}
