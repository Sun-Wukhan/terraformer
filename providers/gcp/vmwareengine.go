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
	vmwareenginev1 "google.golang.org/api/vmwareengine/v1"
)

var vmwareEngineAllowEmptyValues = []string{""}

var vmwareEngineAdditionalFields = map[string]interface{}{}

const vmwareEngineBaseURL = "https://vmwareengine.googleapis.com/v1/"

// VmwareEngineGenerator enumerates Google Cloud VMware Engine in
// the configured region: VMware Engine networks, network peerings,
// network policies with their external access rules, private clouds
// with their clusters, subnets and external addresses — plus
// datastores over raw REST (the pinned vmwareengine/v1 client has
// no datastores service; the provider addresses a datastore by
// location and name, projects/{project}/locations/{location}/
// datastores/{name}).
type VmwareEngineGenerator struct {
	GCPService
}

func (g *VmwareEngineGenerator) InitResources() error {
	ctx := context.Background()
	engineService, err := vmwareenginev1.NewService(ctx)
	if err != nil {
		return err
	}
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, engineService, restClient)
}

func (g *VmwareEngineGenerator) initResources(ctx context.Context, engineService *vmwareenginev1.Service, restClient *http.Client) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + region
	baseAttrs := map[string]string{"project": project, "location": region}

	err := engineService.Projects.Locations.VmwareEngineNetworks.List(parent).Pages(ctx, func(page *vmwareenginev1.ListVmwareEngineNetworksResponse) error {
		for _, network := range page.VmwareEngineNetworks {
			g.Resources = append(g.Resources, g.newVmwareEngineResource(network.Name, "google_vmwareengine_network", baseAttrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = engineService.Projects.Locations.NetworkPeerings.List(parent).Pages(ctx, func(page *vmwareenginev1.ListNetworkPeeringsResponse) error {
		for _, peering := range page.NetworkPeerings {
			g.Resources = append(g.Resources, g.newVmwareEngineResource(peering.Name, "google_vmwareengine_network_peering", baseAttrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = engineService.Projects.Locations.NetworkPolicies.List(parent).Pages(ctx, func(page *vmwareenginev1.ListNetworkPoliciesResponse) error {
		for _, policy := range page.NetworkPolicies {
			g.Resources = append(g.Resources, g.newVmwareEngineResource(policy.Name, "google_vmwareengine_network_policy", baseAttrs))
			if err := g.loadExternalAccessRules(ctx, engineService, policy.Name, baseAttrs); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = engineService.Projects.Locations.PrivateClouds.List(parent).Pages(ctx, func(page *vmwareenginev1.ListPrivateCloudsResponse) error {
		for _, cloud := range page.PrivateClouds {
			g.Resources = append(g.Resources, g.newVmwareEngineResource(cloud.Name, "google_vmwareengine_private_cloud", baseAttrs))
			if err := g.loadPrivateCloudChildren(ctx, engineService, cloud.Name, baseAttrs); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	g.loadDatastores(ctx, restClient, parent, baseAttrs)
	return nil
}

func (g *VmwareEngineGenerator) newVmwareEngineResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
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
		vmwareEngineAllowEmptyValues,
		vmwareEngineAdditionalFields,
	)
}

func (g *VmwareEngineGenerator) loadExternalAccessRules(ctx context.Context, engineService *vmwareenginev1.Service, policyName string, baseAttrs map[string]string) error {
	return engineService.Projects.Locations.NetworkPolicies.ExternalAccessRules.List(policyName).Pages(ctx, func(page *vmwareenginev1.ListExternalAccessRulesResponse) error {
		for _, rule := range page.ExternalAccessRules {
			g.Resources = append(g.Resources, g.newVmwareEngineResource(rule.Name, "google_vmwareengine_external_access_rule", baseAttrs))
		}
		return nil
	})
}

func (g *VmwareEngineGenerator) loadPrivateCloudChildren(ctx context.Context, engineService *vmwareenginev1.Service, cloudName string, baseAttrs map[string]string) error {
	err := engineService.Projects.Locations.PrivateClouds.Clusters.List(cloudName).Pages(ctx, func(page *vmwareenginev1.ListClustersResponse) error {
		for _, cluster := range page.Clusters {
			g.Resources = append(g.Resources, g.newVmwareEngineResource(cluster.Name, "google_vmwareengine_cluster", baseAttrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = engineService.Projects.Locations.PrivateClouds.Subnets.List(cloudName).Pages(ctx, func(page *vmwareenginev1.ListSubnetsResponse) error {
		for _, subnet := range page.Subnets {
			g.Resources = append(g.Resources, g.newVmwareEngineResource(subnet.Name, "google_vmwareengine_subnet", baseAttrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	return engineService.Projects.Locations.PrivateClouds.ExternalAddresses.List(cloudName).Pages(ctx, func(page *vmwareenginev1.ListExternalAddressesResponse) error {
		for _, address := range page.ExternalAddresses {
			g.Resources = append(g.Resources, g.newVmwareEngineResource(address.Name, "google_vmwareengine_external_address", baseAttrs))
		}
		return nil
	})
}

func (g *VmwareEngineGenerator) loadDatastores(ctx context.Context, restClient *http.Client, parent string, baseAttrs map[string]string) {
	items, err := restListResources(ctx, restClient, vmwareEngineBaseURL+parent+"/datastores")
	if err != nil {
		log.Println(err)
		return
	}
	for _, item := range items {
		g.Resources = append(g.Resources, g.newVmwareEngineResource(item.Name, "google_vmwareengine_datastore", baseAttrs))
	}
}
