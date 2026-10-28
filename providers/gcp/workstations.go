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
	"google.golang.org/api/compute/v1"
	workstationsv1 "google.golang.org/api/workstations/v1"
)

var workstationsAllowEmptyValues = []string{""}

var workstationsAdditionalFields = map[string]interface{}{}

// WorkstationsGenerator enumerates Cloud Workstations in the
// configured region: workstation clusters, their workstation
// configs (with IAM), and the workstations in each config (with
// IAM).
type WorkstationsGenerator struct {
	GCPService
}

func (g *WorkstationsGenerator) InitResources() error {
	ctx := context.Background()
	wsService, err := workstationsv1.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, wsService)
}

func (g *WorkstationsGenerator) initResources(ctx context.Context, wsService *workstationsv1.Service) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + region
	baseAttrs := map[string]string{"project": project, "location": region}
	return wsService.Projects.Locations.WorkstationClusters.List(parent).Pages(ctx, func(page *workstationsv1.ListWorkstationClustersResponse) error {
		for _, cluster := range page.WorkstationClusters {
			g.Resources = append(g.Resources, g.newWorkstationsResource(cluster.Name, "google_workstations_workstation_cluster", baseAttrs))
			if err := g.loadWorkstationConfigs(ctx, wsService, cluster.Name, baseAttrs); err != nil {
				return err
			}
		}
		return nil
	})
}

func (g *WorkstationsGenerator) newWorkstationsResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
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
		workstationsAllowEmptyValues,
		workstationsAdditionalFields,
	)
}

func (g *WorkstationsGenerator) loadWorkstationConfigs(ctx context.Context, wsService *workstationsv1.Service, clusterName string, baseAttrs map[string]string) error {
	return wsService.Projects.Locations.WorkstationClusters.WorkstationConfigs.List(clusterName).Pages(ctx, func(page *workstationsv1.ListWorkstationConfigsResponse) error {
		for _, config := range page.WorkstationConfigs {
			g.Resources = append(g.Resources, g.newWorkstationsResource(config.Name, "google_workstations_workstation_config", baseAttrs))
			policy, err := wsService.Projects.Locations.WorkstationClusters.WorkstationConfigs.GetIamPolicy(config.Name).Context(ctx).Do()
			if err != nil {
				log.Println(err)
			} else {
				g.expandWorkstationsIam(policy, config.Name,
					"google_workstations_workstation_config_iam_policy", "google_workstations_workstation_config_iam_binding", "google_workstations_workstation_config_iam_member",
					"workstation_config_id", baseAttrs)
			}
			if err := g.loadWorkstations(ctx, wsService, config.Name, baseAttrs); err != nil {
				return err
			}
		}
		return nil
	})
}

func (g *WorkstationsGenerator) loadWorkstations(ctx context.Context, wsService *workstationsv1.Service, configName string, baseAttrs map[string]string) error {
	return wsService.Projects.Locations.WorkstationClusters.WorkstationConfigs.Workstations.List(configName).Pages(ctx, func(page *workstationsv1.ListWorkstationsResponse) error {
		for _, workstation := range page.Workstations {
			g.Resources = append(g.Resources, g.newWorkstationsResource(workstation.Name, "google_workstations_workstation", baseAttrs))
			policy, err := wsService.Projects.Locations.WorkstationClusters.WorkstationConfigs.Workstations.GetIamPolicy(workstation.Name).Context(ctx).Do()
			if err != nil {
				log.Println(err)
				continue
			}
			g.expandWorkstationsIam(policy, workstation.Name,
				"google_workstations_workstation_iam_policy", "google_workstations_workstation_iam_binding", "google_workstations_workstation_iam_member",
				"workstation_id", baseAttrs)
		}
		return nil
	})
}

func (g *WorkstationsGenerator) expandWorkstationsIam(policy *workstationsv1.Policy, name, policyType, bindingType, memberType, parentKey string, baseAttrs map[string]string) {
	bindings := []iamBinding{}
	for _, binding := range policy.Bindings {
		bindings = append(bindings, iamBinding{Role: binding.Role, Members: binding.Members})
	}
	parentAttrs := map[string]string{parentKey: name}
	for k, v := range baseAttrs {
		parentAttrs[k] = v
	}
	g.Resources = append(g.Resources, expandIamPolicy(
		bindings, name, policyType, bindingType, memberType, parentAttrs, g.newWorkstationsResource)...)
}
