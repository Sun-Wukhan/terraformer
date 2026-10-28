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

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/compute/v1"
	osconfigv1 "google.golang.org/api/osconfig/v1"
	osconfigv2 "google.golang.org/api/osconfig/v2"
)

// OsConfigGenerator enumerates VM Manager (OS Config):
//
//   - v1: patch deployments (project scope) and OS policy assignments
//     (per zone of the configured region), via the generated v1
//     client;
//   - v2: policy orchestrators at project scope and, through CRM
//     ancestry, at folder and organization scope (locations/global),
//     via the generated v2 client.
//
// Deliberately not generated:
//   - google_os_config_guest_policies: the legacy plural resource is
//     a deprecated project-level aggregate (its provider ID is the
//     project itself, not an individual guest policy) superseded by
//     google_os_config_os_policy_assignment; emitting it would
//     double-manage the policies the assignment resources cover.
type OsConfigGenerator struct {
	GCPService
}

func (g *OsConfigGenerator) InitResources() error {
	ctx := context.Background()
	v1Service, err := osconfigv1.NewService(ctx)
	if err != nil {
		return err
	}
	v2Service, err := osconfigv2.NewService(ctx)
	if err != nil {
		return err
	}
	crmService, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		log.Println(err)
	}
	g.initResources(ctx, v1Service, v2Service, crmService)
	return nil
}

func (g *OsConfigGenerator) newResource(id, resourceType string) terraformutils.Resource {
	return terraformutils.NewResource(
		id, opsLastSegment(id), resourceType, g.ProviderName,
		map[string]string{"project": g.GetArgs()["project"].(string)},
		[]string{}, map[string]interface{}{},
	)
}

func (g *OsConfigGenerator) initResources(ctx context.Context, v1Service *osconfigv1.Service, v2Service *osconfigv2.Service, crmService *cloudresourcemanager.Service) {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region)

	if err := v1Service.Projects.PatchDeployments.List("projects/"+project).Pages(ctx, func(page *osconfigv1.ListPatchDeploymentsResponse) error {
		for _, deployment := range page.PatchDeployments {
			g.Resources = append(g.Resources, g.newResource(deployment.Name, "google_os_config_patch_deployment"))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}

	for _, zone := range opsZones(region) {
		parent := "projects/" + project + "/locations/" + zone
		if err := v1Service.Projects.Locations.OsPolicyAssignments.List(parent).Pages(ctx, func(page *osconfigv1.ListOSPolicyAssignmentsResponse) error {
			for _, assignment := range page.OsPolicyAssignments {
				g.Resources = append(g.Resources, g.newResource(assignment.Name, "google_os_config_os_policy_assignment"))
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
	}

	for _, scope := range opsScopes(ctx, crmService, project) {
		parent := scope.parent + "/locations/global"
		var err error
		switch scope.kind {
		case "project":
			err = v2Service.Projects.Locations.Global.PolicyOrchestrators.List(parent).Pages(ctx, func(page *osconfigv2.GoogleCloudOsconfigV2__ListPolicyOrchestratorsResponse) error {
				for _, orchestrator := range page.PolicyOrchestrators {
					g.Resources = append(g.Resources, g.newResource(orchestrator.Name, "google_os_config_v2_policy_orchestrator"))
				}
				return nil
			})
		case "folder":
			err = v2Service.Folders.Locations.Global.PolicyOrchestrators.List(parent).Pages(ctx, func(page *osconfigv2.GoogleCloudOsconfigV2__ListPolicyOrchestratorsResponse) error {
				for _, orchestrator := range page.PolicyOrchestrators {
					g.Resources = append(g.Resources, g.newResource(orchestrator.Name, "google_os_config_v2_policy_orchestrator_for_folder"))
				}
				return nil
			})
		case "organization":
			err = v2Service.Organizations.Locations.Global.PolicyOrchestrators.List(parent).Pages(ctx, func(page *osconfigv2.GoogleCloudOsconfigV2__ListPolicyOrchestratorsResponse) error {
				for _, orchestrator := range page.PolicyOrchestrators {
					g.Resources = append(g.Resources, g.newResource(orchestrator.Name, "google_os_config_v2_policy_orchestrator_for_organization"))
				}
				return nil
			})
		}
		if err != nil {
			log.Println(err)
		}
	}
}
