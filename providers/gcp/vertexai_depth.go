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

	aiplatform "google.golang.org/api/aiplatform/v1beta1"
	"google.golang.org/api/compute/v1"
)

// Depth pass for the Vertex AI generator (see vertexai.go), covering
// the types the GA v1 client cannot enumerate:
//
//   - Endpoint IAM: endpoints expose getIamPolicy only in the
//     v1beta1 aiplatform client, so the endpoints are re-listed
//     through v1beta1 and each policy is expanded into the
//     google_vertex_ai_endpoint_iam_policy/_binding/_member trio
//     with literal type names.
//   - The per-location RAG engine config singleton, read over raw
//     REST (GET projects/{project}/locations/{location}/
//     ragEngineConfig in aiplatform v1 — the generated clients do
//     not surface the method).
//   - Semantic governance: the policy collection and the
//     per-location policy engine singleton, over raw REST against
//     the regional aiplatform v1 endpoint (the collections exist in
//     the current v1 discovery document but not in the pinned
//     generated clients).
func (g *VertexAiGenerator) initDepthResources(ctx context.Context, betaService *aiplatform.Service, restClient *http.Client) {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + region
	g.loadEndpointIam(ctx, betaService, parent)
	g.loadRagEngineConfig(ctx, restClient, parent)
	g.loadSemanticGovernance(ctx, restClient, parent)
}

func (g *VertexAiGenerator) vertexRestBase() string {
	region := g.GetArgs()["region"].(compute.Region).Name
	if region == "" {
		return "https://aiplatform.googleapis.com/v1/"
	}
	return "https://" + region + "-aiplatform.googleapis.com/v1/"
}

func (g *VertexAiGenerator) loadEndpointIam(ctx context.Context, betaService *aiplatform.Service, parent string) {
	err := betaService.Projects.Locations.Endpoints.List(parent).Pages(ctx, func(page *aiplatform.GoogleCloudAiplatformV1beta1ListEndpointsResponse) error {
		for _, endpoint := range page.Endpoints {
			policy, err := betaService.Projects.Locations.Endpoints.GetIamPolicy(endpoint.Name).Context(ctx).Do()
			if err != nil {
				log.Println(err)
				continue
			}
			bindings := []iamBinding{}
			for _, binding := range policy.Bindings {
				bindings = append(bindings, iamBinding{Role: binding.Role, Members: binding.Members})
			}
			g.Resources = append(g.Resources, expandIamPolicy(bindings, endpoint.Name,
				"google_vertex_ai_endpoint_iam_policy",
				"google_vertex_ai_endpoint_iam_binding",
				"google_vertex_ai_endpoint_iam_member",
				map[string]string{"endpoint": endpoint.Name},
				g.vertexResourceFunc())...)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

// loadRagEngineConfig emits the per-location RAG engine config
// singleton when one has been configured for the location.
func (g *VertexAiGenerator) loadRagEngineConfig(ctx context.Context, restClient *http.Client, parent string) {
	name := parent + "/ragEngineConfig"
	returned, err := aiRestSingletonName(ctx, restClient, g.vertexRestBase()+name)
	if err != nil {
		log.Println(err)
		return
	}
	if returned != "" {
		name = returned
	}
	g.Resources = append(g.Resources, g.newVertexResource(name, "rag_engine_config", "google_vertex_ai_rag_engine_config", map[string]string{"name": name}))
}

func (g *VertexAiGenerator) loadSemanticGovernance(ctx context.Context, restClient *http.Client, parent string) {
	policies, err := restListResources(ctx, restClient, g.vertexRestBase()+parent+"/semanticGovernancePolicies")
	if err != nil {
		log.Println(err)
	} else {
		for _, policy := range policies {
			g.Resources = append(g.Resources, g.newVertexResource(policy.Name, vertexLabel(policy.Name, ""), "google_vertex_ai_semantic_governance_policy", map[string]string{"name": policy.Name}))
		}
	}
	engine := parent + "/semanticGovernancePolicyEngine"
	returned, err := aiRestSingletonName(ctx, restClient, g.vertexRestBase()+engine)
	if err != nil {
		log.Println(err)
		return
	}
	if returned == "" {
		return
	}
	g.Resources = append(g.Resources, g.newVertexResource(returned, "semantic_governance_policy_engine", "google_vertex_ai_semantic_governance_policy_engine", map[string]string{"name": returned}))
}
