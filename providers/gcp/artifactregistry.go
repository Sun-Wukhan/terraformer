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
	artifactregistryv1 "google.golang.org/api/artifactregistry/v1"
	"google.golang.org/api/compute/v1"
)

var artifactRegistryAllowEmptyValues = []string{""}

var artifactRegistryAdditionalFields = map[string]interface{}{}

const artifactRegistryBaseURL = "https://artifactregistry.googleapis.com/v1/"

// ArtifactRegistryGenerator enumerates Artifact Registry in the
// configured region: repositories (with IAM) and their rules, plus
// the two per-location singletons — the project config (read over
// raw REST: the pinned artifactregistry/v1 client has no project
// config methods) and the VPCSC config.
type ArtifactRegistryGenerator struct {
	GCPService
}

func (g *ArtifactRegistryGenerator) InitResources() error {
	ctx := context.Background()
	registryService, err := artifactregistryv1.NewService(ctx)
	if err != nil {
		return err
	}
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, registryService, restClient)
}

func (g *ArtifactRegistryGenerator) initResources(ctx context.Context, registryService *artifactregistryv1.Service, restClient *http.Client) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + region
	baseAttrs := map[string]string{"project": project, "location": region}

	err := registryService.Projects.Locations.Repositories.List(parent).Pages(ctx, func(page *artifactregistryv1.ListRepositoriesResponse) error {
		for _, repository := range page.Repositories {
			g.Resources = append(g.Resources, g.newArtifactRegistryResource(repository.Name, "google_artifact_registry_repository", baseAttrs))
			policy, err := registryService.Projects.Locations.Repositories.GetIamPolicy(repository.Name).Context(ctx).Do()
			if err != nil {
				log.Println(err)
			} else {
				bindings := []iamBinding{}
				for _, binding := range policy.Bindings {
					bindings = append(bindings, iamBinding{Role: binding.Role, Members: binding.Members})
				}
				parentAttrs := map[string]string{"repository": repository.Name}
				for k, v := range baseAttrs {
					parentAttrs[k] = v
				}
				g.Resources = append(g.Resources, expandIamPolicy(
					bindings, repository.Name,
					"google_artifact_registry_repository_iam_policy", "google_artifact_registry_repository_iam_binding", "google_artifact_registry_repository_iam_member",
					parentAttrs, g.newArtifactRegistryResource)...)
			}
			if err := g.loadRepositoryRules(ctx, registryService, repository.Name, baseAttrs); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	g.loadProjectConfig(ctx, restClient, parent, baseAttrs)
	g.loadVpcscConfig(ctx, registryService, parent, baseAttrs)
	return nil
}

func (g *ArtifactRegistryGenerator) newArtifactRegistryResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
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
		artifactRegistryAllowEmptyValues,
		artifactRegistryAdditionalFields,
	)
}

func (g *ArtifactRegistryGenerator) loadRepositoryRules(ctx context.Context, registryService *artifactregistryv1.Service, repositoryName string, baseAttrs map[string]string) error {
	return registryService.Projects.Locations.Repositories.Rules.List(repositoryName).Pages(ctx, func(page *artifactregistryv1.ListRulesResponse) error {
		for _, rule := range page.Rules {
			g.Resources = append(g.Resources, g.newArtifactRegistryResource(rule.Name, "google_artifact_registry_rule", baseAttrs))
		}
		return nil
	})
}

// loadProjectConfig reads the per-location project config singleton
// over raw REST and emits it when the location has one.
func (g *ArtifactRegistryGenerator) loadProjectConfig(ctx context.Context, restClient *http.Client, parent string, baseAttrs map[string]string) {
	body, err := restDo(ctx, restClient, http.MethodGet, artifactRegistryBaseURL+parent+"/projectConfig", nil)
	if err != nil {
		log.Println(err)
		return
	}
	var config struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(body, &config); err != nil {
		log.Println(err)
		return
	}
	id := config.Name
	if id == "" {
		id = parent + "/projectConfig"
	}
	g.Resources = append(g.Resources, g.newArtifactRegistryResource(id, "google_artifact_registry_project_config", baseAttrs))
}

func (g *ArtifactRegistryGenerator) loadVpcscConfig(ctx context.Context, registryService *artifactregistryv1.Service, parent string, baseAttrs map[string]string) {
	config, err := registryService.Projects.Locations.GetVpcscConfig(parent + "/vpcscConfig").Context(ctx).Do()
	if err != nil {
		log.Println(err)
		return
	}
	id := config.Name
	if id == "" {
		id = parent + "/vpcscConfig"
	}
	g.Resources = append(g.Resources, g.newArtifactRegistryResource(id, "google_artifact_registry_vpcsc_config", baseAttrs))
}
