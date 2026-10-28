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
	tpuv2 "google.golang.org/api/tpu/v2"
)

var tpuAllowEmptyValues = []string{""}

var tpuAdditionalFields = map[string]interface{}{}

// TpuGenerator enumerates Cloud TPU (v2 API): VMs (nodes) and
// queued resources. TPU resources are zonal, so every zone of the
// configured region is enumerated.
type TpuGenerator struct {
	GCPService
}

func (g *TpuGenerator) InitResources() error {
	ctx := context.Background()
	tpuService, err := tpuv2.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, tpuService)
}

func (g *TpuGenerator) initResources(ctx context.Context, tpuService *tpuv2.Service) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region)
	for _, zone := range computeZoneNames(region) {
		parent := "projects/" + project + "/locations/" + zone
		baseAttrs := map[string]string{"project": project, "location": zone}
		err := tpuService.Projects.Locations.Nodes.List(parent).Pages(ctx, func(page *tpuv2.ListNodesResponse) error {
			for _, node := range page.Nodes {
				g.Resources = append(g.Resources, g.newTpuResource(node.Name, "google_tpu_v2_vm", baseAttrs))
			}
			return nil
		})
		if err != nil {
			return err
		}
		err = tpuService.Projects.Locations.QueuedResources.List(parent).Pages(ctx, func(page *tpuv2.ListQueuedResourcesResponse) error {
			for _, resource := range page.QueuedResources {
				g.Resources = append(g.Resources, g.newTpuResource(resource.Name, "google_tpu_v2_queued_resource", baseAttrs))
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *TpuGenerator) newTpuResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
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
		tpuAllowEmptyValues,
		tpuAdditionalFields,
	)
}
