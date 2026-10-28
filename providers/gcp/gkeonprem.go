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
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"google.golang.org/api/compute/v1"
	gkeonpremv1 "google.golang.org/api/gkeonprem/v1"
)

var gkeOnpremAllowEmptyValues = []string{""}

var gkeOnpremAdditionalFields = map[string]interface{}{}

// GkeOnpremGenerator enumerates GKE on-prem (Google Distributed
// Cloud) clusters in the configured region: bare metal admin
// clusters, bare metal clusters with their node pools, VMware admin
// clusters, and VMware clusters with their node pools.
type GkeOnpremGenerator struct {
	GCPService
}

func (g *GkeOnpremGenerator) InitResources() error {
	ctx := context.Background()
	onpremService, err := gkeonpremv1.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, onpremService)
}

func (g *GkeOnpremGenerator) initResources(ctx context.Context, onpremService *gkeonpremv1.Service) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + region
	baseAttrs := map[string]string{"project": project, "location": region}

	err := onpremService.Projects.Locations.BareMetalAdminClusters.List(parent).Pages(ctx, func(page *gkeonpremv1.ListBareMetalAdminClustersResponse) error {
		for _, cluster := range page.BareMetalAdminClusters {
			g.Resources = append(g.Resources, g.newGkeOnpremResource(cluster.Name, "google_gkeonprem_bare_metal_admin_cluster", baseAttrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = onpremService.Projects.Locations.VmwareAdminClusters.List(parent).Pages(ctx, func(page *gkeonpremv1.ListVmwareAdminClustersResponse) error {
		for _, cluster := range page.VmwareAdminClusters {
			g.Resources = append(g.Resources, g.newGkeOnpremResource(cluster.Name, "google_gkeonprem_vmware_admin_cluster", baseAttrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = onpremService.Projects.Locations.BareMetalClusters.List(parent).Pages(ctx, func(page *gkeonpremv1.ListBareMetalClustersResponse) error {
		for _, cluster := range page.BareMetalClusters {
			g.Resources = append(g.Resources, g.newGkeOnpremResource(cluster.Name, "google_gkeonprem_bare_metal_cluster", baseAttrs))
			if err := g.loadBareMetalNodePools(ctx, onpremService, cluster.Name, baseAttrs); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return onpremService.Projects.Locations.VmwareClusters.List(parent).Pages(ctx, func(page *gkeonpremv1.ListVmwareClustersResponse) error {
		for _, cluster := range page.VmwareClusters {
			g.Resources = append(g.Resources, g.newGkeOnpremResource(cluster.Name, "google_gkeonprem_vmware_cluster", baseAttrs))
			if err := g.loadVmwareNodePools(ctx, onpremService, cluster.Name, baseAttrs); err != nil {
				return err
			}
		}
		return nil
	})
}

func (g *GkeOnpremGenerator) newGkeOnpremResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
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
		gkeOnpremAllowEmptyValues,
		gkeOnpremAdditionalFields,
	)
}

func (g *GkeOnpremGenerator) loadBareMetalNodePools(ctx context.Context, onpremService *gkeonpremv1.Service, clusterName string, baseAttrs map[string]string) error {
	return onpremService.Projects.Locations.BareMetalClusters.BareMetalNodePools.List(clusterName).Pages(ctx, func(page *gkeonpremv1.ListBareMetalNodePoolsResponse) error {
		for _, pool := range page.BareMetalNodePools {
			g.Resources = append(g.Resources, g.newGkeOnpremResource(pool.Name, "google_gkeonprem_bare_metal_node_pool", baseAttrs))
		}
		return nil
	})
}

func (g *GkeOnpremGenerator) loadVmwareNodePools(ctx context.Context, onpremService *gkeonpremv1.Service, clusterName string, baseAttrs map[string]string) error {
	return onpremService.Projects.Locations.VmwareClusters.VmwareNodePools.List(clusterName).Pages(ctx, func(page *gkeonpremv1.ListVmwareNodePoolsResponse) error {
		for _, pool := range page.VmwareNodePools {
			g.Resources = append(g.Resources, g.newGkeOnpremResource(pool.Name, "google_gkeonprem_vmware_node_pool", baseAttrs))
		}
		return nil
	})
}
