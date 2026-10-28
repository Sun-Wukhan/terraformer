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
	dataformv1beta1 "google.golang.org/api/dataform/v1beta1"
)

var dataformAllowEmptyValues = []string{""}

var dataformAdditionalFields = map[string]interface{}{}

type DataformGenerator struct {
	GCPService
}

// InitResources enumerates Dataform resources for the configured
// project and region: repositories with their IAM trio, release
// configs and workflow configs per repository, and the per-location
// config singleton (read through locations.getConfig — the config
// always exists for a location that has the API enabled; its
// provider ID is projects/{project}/locations/{location}/config).
//
// Deliberately not generated:
//   - google_dataform_folder and google_dataform_team_folder: the
//     Dataform REST discovery document (and with it the pinned
//     generated client) has no folders or teamFolders resource at
//     all — those objects exist only in the GAPIC surface, which
//     this module deliberately does not depend on — so there is no
//     verified list endpoint to enumerate them through.
func (g *DataformGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := dataformv1beta1.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, svc)
}

func (g *DataformGenerator) initResources(ctx context.Context, svc *dataformv1beta1.Service) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region)
	locationParent := "projects/" + project + "/locations/" + region.Name
	baseAttrs := map[string]string{"project": project, "location": region.Name}
	err := svc.Projects.Locations.Repositories.List(locationParent).Pages(ctx, func(page *dataformv1beta1.ListRepositoriesResponse) error {
		for _, repository := range page.Repositories {
			g.Resources = append(g.Resources, g.newDataformResource(repository.Name, "google_dataform_repository", baseAttrs))
			g.loadDataformIam(ctx, svc, repository.Name, baseAttrs)
			g.loadReleaseConfigs(ctx, svc, repository.Name, baseAttrs)
			g.loadWorkflowConfigs(ctx, svc, repository.Name, baseAttrs)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	if _, err := svc.Projects.Locations.GetConfig(locationParent).Context(ctx).Do(); err != nil {
		log.Println(err)
	} else {
		g.Resources = append(g.Resources, g.newDataformResource(locationParent+"/config", "google_dataform_config", baseAttrs))
	}
	return nil
}

func (g *DataformGenerator) newDataformResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
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
		dataformAllowEmptyValues,
		dataformAdditionalFields,
	)
}

func (g *DataformGenerator) loadDataformIam(ctx context.Context, svc *dataformv1beta1.Service, repositoryName string, baseAttrs map[string]string) {
	policy, err := svc.Projects.Locations.Repositories.GetIamPolicy(repositoryName).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		return
	}
	bindings := []iamBinding{}
	for _, binding := range policy.Bindings {
		bindings = append(bindings, iamBinding{Role: binding.Role, Members: binding.Members})
	}
	g.Resources = append(g.Resources, expandIamPolicy(
		bindings,
		repositoryName,
		"google_dataform_repository_iam_policy",
		"google_dataform_repository_iam_binding",
		"google_dataform_repository_iam_member",
		baseAttrs,
		g.newDataformResource,
	)...)
}

func (g *DataformGenerator) loadReleaseConfigs(ctx context.Context, svc *dataformv1beta1.Service, repositoryName string, baseAttrs map[string]string) {
	err := svc.Projects.Locations.Repositories.ReleaseConfigs.List(repositoryName).Pages(ctx, func(page *dataformv1beta1.ListReleaseConfigsResponse) error {
		for _, config := range page.ReleaseConfigs {
			g.Resources = append(g.Resources, g.newDataformResource(config.Name, "google_dataform_repository_release_config", baseAttrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (g *DataformGenerator) loadWorkflowConfigs(ctx context.Context, svc *dataformv1beta1.Service, repositoryName string, baseAttrs map[string]string) {
	err := svc.Projects.Locations.Repositories.WorkflowConfigs.List(repositoryName).Pages(ctx, func(page *dataformv1beta1.ListWorkflowConfigsResponse) error {
		for _, config := range page.WorkflowConfigs {
			g.Resources = append(g.Resources, g.newDataformResource(config.Name, "google_dataform_repository_workflow_config", baseAttrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}
