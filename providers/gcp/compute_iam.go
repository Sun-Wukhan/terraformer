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
	"log"
	"net/http"
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/compute/v1"
)

var computeIamAllowEmptyValues = []string{""}

var computeIamAdditionalFields = map[string]interface{}{}

type ComputeIamGenerator struct {
	GCPService
}

// InitResources is the IAM attachment pass over the Compute Engine
// resources the other generators emit: for every parent resource it
// reads the live IAM policy through the compute client's getIamPolicy
// methods and expands it into the provider's _iam_policy, _iam_binding
// and _iam_member resources (the shared expandIamPolicy helper, as in
// the Network Security and Dataplex generators). The parents are
// re-listed here so the pass is independent of which other services
// the user selected:
//
//   - Global: backend buckets, backend services, images, instance
//     templates, machine images, snapshots, network firewall policies
//     and — scoped by organization/folder ancestry — hierarchy
//     firewall policies.
//   - Regional: backend services, disks, instant snapshots,
//     subnetworks, region network firewall policies, and region
//     backend buckets (whose getIamPolicy is read over raw REST: the
//     pinned client has no region backend bucket service at all, but
//     the GA discovery document has both endpoints).
//   - Zonal, per zone of the configured region: disks, instances,
//     instant snapshots and storage pools.
//
// A getIamPolicy failure for one parent (for example a resource type
// that rejects the call in some state) is logged and skipped; it
// never aborts the pass.
func (g *ComputeIamGenerator) InitResources() error {
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

func (g *ComputeIamGenerator) initResources(ctx context.Context, computeService *compute.Service, crmService *cloudresourcemanager.Service, restClient *http.Client) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region)
	regionName := region.Name
	zones := computeZoneNames(region)

	g.loadGlobalIam(ctx, computeService, project)
	g.loadRegionalIam(ctx, computeService, restClient, project, regionName)
	for _, zone := range zones {
		g.loadZonalIam(ctx, computeService, project, zone)
	}
	g.loadHierarchyFirewallPolicyIam(ctx, computeService, crmService, project)
	return nil
}

func (g *ComputeIamGenerator) newComputeIamResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
	if attrs == nil {
		attrs = map[string]string{}
	}
	return terraformutils.NewResource(
		id,
		id,
		resourceType,
		g.ProviderName,
		attrs,
		computeIamAllowEmptyValues,
		computeIamAdditionalFields,
	)
}

// expandComputeIamPolicy expands one parent's live policy into the
// provider IAM trio (the _iam_policy, _iam_binding and _iam_member
// resource types, named literally at every call site so the emitted
// type strings stay greppable, as elsewhere in this package), all
// carrying the parent's provider ID.
func (g *ComputeIamGenerator) expandComputeIamPolicy(policy *compute.Policy, parentID, policyType, bindingType, memberType string, parentAttrs map[string]string) {
	if policy == nil {
		return
	}
	bindings := []iamBinding{}
	for _, binding := range policy.Bindings {
		bindings = append(bindings, iamBinding{Role: binding.Role, Members: binding.Members})
	}
	g.Resources = append(g.Resources, expandIamPolicy(
		bindings, parentID, policyType, bindingType, memberType,
		parentAttrs, g.newComputeIamResource)...)
}

func (g *ComputeIamGenerator) loadGlobalIam(ctx context.Context, computeService *compute.Service, project string) {
	err := computeService.BackendBuckets.List(project).Pages(ctx, func(page *compute.BackendBucketList) error {
		for _, item := range page.Items {
			policy, err := computeService.BackendBuckets.GetIamPolicy(project, item.Name).Context(ctx).Do()
			if err != nil {
				log.Println(err)
				continue
			}
			g.expandComputeIamPolicy(policy, "projects/"+project+"/global/backendBuckets/"+item.Name,
				"google_compute_backend_bucket_iam_policy", "google_compute_backend_bucket_iam_binding", "google_compute_backend_bucket_iam_member", map[string]string{"name": item.Name, "project": project})
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = computeService.BackendServices.List(project).Pages(ctx, func(page *compute.BackendServiceList) error {
		for _, item := range page.Items {
			policy, err := computeService.BackendServices.GetIamPolicy(project, item.Name).Context(ctx).Do()
			if err != nil {
				log.Println(err)
				continue
			}
			g.expandComputeIamPolicy(policy, "projects/"+project+"/global/backendServices/"+item.Name,
				"google_compute_backend_service_iam_policy", "google_compute_backend_service_iam_binding", "google_compute_backend_service_iam_member", map[string]string{"name": item.Name, "project": project})
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = computeService.Images.List(project).Pages(ctx, func(page *compute.ImageList) error {
		for _, item := range page.Items {
			policy, err := computeService.Images.GetIamPolicy(project, item.Name).Context(ctx).Do()
			if err != nil {
				log.Println(err)
				continue
			}
			g.expandComputeIamPolicy(policy, "projects/"+project+"/global/images/"+item.Name,
				"google_compute_image_iam_policy", "google_compute_image_iam_binding", "google_compute_image_iam_member", map[string]string{"image": item.Name, "project": project})
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = computeService.InstanceTemplates.List(project).Pages(ctx, func(page *compute.InstanceTemplateList) error {
		for _, item := range page.Items {
			policy, err := computeService.InstanceTemplates.GetIamPolicy(project, item.Name).Context(ctx).Do()
			if err != nil {
				log.Println(err)
				continue
			}
			g.expandComputeIamPolicy(policy, "projects/"+project+"/global/instanceTemplates/"+item.Name,
				"google_compute_instance_template_iam_policy", "google_compute_instance_template_iam_binding", "google_compute_instance_template_iam_member", map[string]string{"name": item.Name, "project": project})
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = computeService.MachineImages.List(project).Pages(ctx, func(page *compute.MachineImageList) error {
		for _, item := range page.Items {
			policy, err := computeService.MachineImages.GetIamPolicy(project, item.Name).Context(ctx).Do()
			if err != nil {
				log.Println(err)
				continue
			}
			g.expandComputeIamPolicy(policy, "projects/"+project+"/global/machineImages/"+item.Name,
				"google_compute_machine_image_iam_policy", "google_compute_machine_image_iam_binding", "google_compute_machine_image_iam_member", map[string]string{"machine_image": item.Name, "project": project})
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = computeService.Snapshots.List(project).Pages(ctx, func(page *compute.SnapshotList) error {
		for _, item := range page.Items {
			policy, err := computeService.Snapshots.GetIamPolicy(project, item.Name).Context(ctx).Do()
			if err != nil {
				log.Println(err)
				continue
			}
			g.expandComputeIamPolicy(policy, "projects/"+project+"/global/snapshots/"+item.Name,
				"google_compute_snapshot_iam_policy", "google_compute_snapshot_iam_binding", "google_compute_snapshot_iam_member", map[string]string{"name": item.Name, "project": project})
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = computeService.NetworkFirewallPolicies.List(project).Pages(ctx, func(page *compute.FirewallPolicyList) error {
		for _, item := range page.Items {
			policy, err := computeService.NetworkFirewallPolicies.GetIamPolicy(project, item.Name).Context(ctx).Do()
			if err != nil {
				log.Println(err)
				continue
			}
			g.expandComputeIamPolicy(policy, "projects/"+project+"/global/firewallPolicies/"+item.Name,
				"google_compute_network_firewall_policy_iam_policy", "google_compute_network_firewall_policy_iam_binding", "google_compute_network_firewall_policy_iam_member", map[string]string{"name": item.Name, "project": project})
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (g *ComputeIamGenerator) loadRegionalIam(ctx context.Context, computeService *compute.Service, restClient *http.Client, project, region string) {
	err := computeService.RegionBackendServices.List(project, region).Pages(ctx, func(page *compute.BackendServiceList) error {
		for _, item := range page.Items {
			policy, err := computeService.RegionBackendServices.GetIamPolicy(project, region, item.Name).Context(ctx).Do()
			if err != nil {
				log.Println(err)
				continue
			}
			g.expandComputeIamPolicy(policy, "projects/"+project+"/regions/"+region+"/backendServices/"+item.Name,
				"google_compute_region_backend_service_iam_policy", "google_compute_region_backend_service_iam_binding", "google_compute_region_backend_service_iam_member", map[string]string{"name": item.Name, "region": region, "project": project})
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = computeService.RegionDisks.List(project, region).Pages(ctx, func(page *compute.DiskList) error {
		for _, item := range page.Items {
			policy, err := computeService.RegionDisks.GetIamPolicy(project, region, item.Name).Context(ctx).Do()
			if err != nil {
				log.Println(err)
				continue
			}
			g.expandComputeIamPolicy(policy, "projects/"+project+"/regions/"+region+"/disks/"+item.Name,
				"google_compute_region_disk_iam_policy", "google_compute_region_disk_iam_binding", "google_compute_region_disk_iam_member", map[string]string{"name": item.Name, "region": region, "project": project})
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = computeService.RegionInstantSnapshots.List(project, region).Pages(ctx, func(page *compute.InstantSnapshotList) error {
		for _, item := range page.Items {
			policy, err := computeService.RegionInstantSnapshots.GetIamPolicy(project, region, item.Name).Context(ctx).Do()
			if err != nil {
				log.Println(err)
				continue
			}
			g.expandComputeIamPolicy(policy, "projects/"+project+"/regions/"+region+"/instantSnapshots/"+item.Name,
				"google_compute_region_instant_snapshot_iam_policy", "google_compute_region_instant_snapshot_iam_binding", "google_compute_region_instant_snapshot_iam_member", map[string]string{"name": item.Name, "region": region, "project": project})
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = computeService.Subnetworks.List(project, region).Pages(ctx, func(page *compute.SubnetworkList) error {
		for _, item := range page.Items {
			policy, err := computeService.Subnetworks.GetIamPolicy(project, region, item.Name).Context(ctx).Do()
			if err != nil {
				log.Println(err)
				continue
			}
			g.expandComputeIamPolicy(policy, "projects/"+project+"/regions/"+region+"/subnetworks/"+item.Name,
				"google_compute_subnetwork_iam_policy", "google_compute_subnetwork_iam_binding", "google_compute_subnetwork_iam_member", map[string]string{"subnetwork": item.Name, "region": region, "project": project})
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = computeService.RegionNetworkFirewallPolicies.List(project, region).Pages(ctx, func(page *compute.FirewallPolicyList) error {
		for _, item := range page.Items {
			policy, err := computeService.RegionNetworkFirewallPolicies.GetIamPolicy(project, region, item.Name).Context(ctx).Do()
			if err != nil {
				log.Println(err)
				continue
			}
			g.expandComputeIamPolicy(policy, "projects/"+project+"/regions/"+region+"/firewallPolicies/"+item.Name,
				"google_compute_region_network_firewall_policy_iam_policy", "google_compute_region_network_firewall_policy_iam_binding", "google_compute_region_network_firewall_policy_iam_member", map[string]string{"name": item.Name, "region": region, "project": project})
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	g.loadRegionBackendBucketIam(ctx, restClient, project, region)
}

// loadRegionBackendBucketIam reads region backend bucket policies
// over raw REST: the pinned client has no region backend bucket
// service, but the GA discovery document exposes both the list and
// the getIamPolicy endpoints.
func (g *ComputeIamGenerator) loadRegionBackendBucketIam(ctx context.Context, restClient *http.Client, project, region string) {
	base := "https://compute.googleapis.com/compute/v1/projects/" + project + "/regions/" + region + "/backendBuckets"
	items, err := listComputeRest(ctx, restClient, base)
	if err != nil {
		log.Println(err)
		return
	}
	for _, item := range items {
		body, err := restDo(ctx, restClient, http.MethodGet, base+"/"+item.Name+"/getIamPolicy", nil)
		if err != nil {
			log.Println(err)
			continue
		}
		var policy struct {
			Bindings []iamBinding `json:"bindings"`
		}
		if err := json.Unmarshal(body, &policy); err != nil {
			log.Println(err)
			continue
		}
		parentID := "projects/" + project + "/regions/" + region + "/backendBuckets/" + item.Name
		parentAttrs := map[string]string{"name": item.Name, "region": region, "project": project}
		g.Resources = append(g.Resources, expandIamPolicy(
			policy.Bindings, parentID,
			"google_compute_region_backend_bucket_iam_policy",
			"google_compute_region_backend_bucket_iam_binding",
			"google_compute_region_backend_bucket_iam_member",
			parentAttrs, g.newComputeIamResource)...)
	}
}

func (g *ComputeIamGenerator) loadZonalIam(ctx context.Context, computeService *compute.Service, project, zone string) {
	err := computeService.Disks.List(project, zone).Pages(ctx, func(page *compute.DiskList) error {
		for _, item := range page.Items {
			policy, err := computeService.Disks.GetIamPolicy(project, zone, item.Name).Context(ctx).Do()
			if err != nil {
				log.Println(err)
				continue
			}
			g.expandComputeIamPolicy(policy, "projects/"+project+"/zones/"+zone+"/disks/"+item.Name,
				"google_compute_disk_iam_policy", "google_compute_disk_iam_binding", "google_compute_disk_iam_member", map[string]string{"name": item.Name, "zone": zone, "project": project})
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = computeService.Instances.List(project, zone).Pages(ctx, func(page *compute.InstanceList) error {
		for _, item := range page.Items {
			policy, err := computeService.Instances.GetIamPolicy(project, zone, item.Name).Context(ctx).Do()
			if err != nil {
				log.Println(err)
				continue
			}
			g.expandComputeIamPolicy(policy, "projects/"+project+"/zones/"+zone+"/instances/"+item.Name,
				"google_compute_instance_iam_policy", "google_compute_instance_iam_binding", "google_compute_instance_iam_member", map[string]string{"instance_name": item.Name, "zone": zone, "project": project})
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = computeService.InstantSnapshots.List(project, zone).Pages(ctx, func(page *compute.InstantSnapshotList) error {
		for _, item := range page.Items {
			policy, err := computeService.InstantSnapshots.GetIamPolicy(project, zone, item.Name).Context(ctx).Do()
			if err != nil {
				log.Println(err)
				continue
			}
			g.expandComputeIamPolicy(policy, "projects/"+project+"/zones/"+zone+"/instantSnapshots/"+item.Name,
				"google_compute_instant_snapshot_iam_policy", "google_compute_instant_snapshot_iam_binding", "google_compute_instant_snapshot_iam_member", map[string]string{"name": item.Name, "zone": zone, "project": project})
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = computeService.StoragePools.List(project, zone).Pages(ctx, func(page *compute.StoragePoolList) error {
		for _, item := range page.Items {
			policy, err := computeService.StoragePools.GetIamPolicy(project, zone, item.Name).Context(ctx).Do()
			if err != nil {
				log.Println(err)
				continue
			}
			g.expandComputeIamPolicy(policy, "projects/"+project+"/zones/"+zone+"/storagePools/"+item.Name,
				"google_compute_storage_pool_iam_policy", "google_compute_storage_pool_iam_binding", "google_compute_storage_pool_iam_member", map[string]string{"name": item.Name, "zone": zone, "project": project})
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

// loadHierarchyFirewallPolicyIam expands the IAM policies of the
// hierarchy firewall policies visible from the project's organization
// and ancestor folders (the parents the policies generator uses).
func (g *ComputeIamGenerator) loadHierarchyFirewallPolicyIam(ctx context.Context, computeService *compute.Service, crmService *cloudresourcemanager.Service, project string) {
	if crmService == nil {
		return
	}
	ancestry, err := crmService.Projects.GetAncestry(project, &cloudresourcemanager.GetAncestryRequest{}).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		return
	}
	parents := []string{}
	for _, ancestor := range ancestry.Ancestor {
		if ancestor.ResourceId == nil {
			continue
		}
		switch ancestor.ResourceId.Type {
		case "organization":
			parents = append(parents, "organizations/"+ancestor.ResourceId.Id)
		case "folder":
			parents = append(parents, "folders/"+ancestor.ResourceId.Id)
		}
	}
	for _, parent := range parents {
		err := computeService.FirewallPolicies.List().ParentId(parent).Pages(ctx, func(page *compute.FirewallPolicyList) error {
			for _, item := range page.Items {
				id := item.Name[strings.LastIndex(item.Name, "/")+1:]
				policy, err := computeService.FirewallPolicies.GetIamPolicy(id).Context(ctx).Do()
				if err != nil {
					log.Println(err)
					continue
				}
				g.expandComputeIamPolicy(policy, "locations/global/firewallPolicies/"+id,
					"google_compute_firewall_policy_iam_policy", "google_compute_firewall_policy_iam_binding", "google_compute_firewall_policy_iam_member", map[string]string{"name": id})
			}
			return nil
		})
		if err != nil {
			log.Println(err)
		}
	}
}
