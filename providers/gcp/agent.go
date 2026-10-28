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
)

// AgentGenerator enumerates the Gemini Enterprise Agent Platform
// control-plane surface of one project/location over raw REST —
// neither API has a generated client in the pinned
// google.golang.org/api module:
//
//   - Agent Registry (agentregistry.googleapis.com/v1beta1):
//     manually registered services and registry bindings.
//   - Agent Identity (agentidentity.googleapis.com/v1beta1):
//     auth providers, each with its IAM trio expanded with literal
//     type names from the standard :getIamPolicy endpoint.
//
// Registry resources live in both regional locations and the
// "global" location (Google-managed MCP servers register only in
// "global"), so both the configured region and "global" are
// enumerated. Every provider ID is the resource's full API name;
// the APIs' resource names carry the project number rather than
// the project ID, which is why the IDs are taken verbatim from the
// list responses.
type AgentGenerator struct {
	GCPService
}

const (
	agentRegistryBaseURL  = "https://agentregistry.googleapis.com/v1beta1/"
	agentIdentityBaseURL  = "https://agentidentity.googleapis.com/v1beta1/"
	agentPlatformLocation = "global"
)

func (g *AgentGenerator) InitResources() error {
	ctx := context.Background()
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		log.Println(err)
		return nil
	}
	g.initResources(ctx, restClient)
	return nil
}

func (g *AgentGenerator) initResources(ctx context.Context, restClient *http.Client) {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	locations := []string{agentPlatformLocation}
	if region != "" && region != agentPlatformLocation {
		locations = append(locations, region)
	}
	for _, location := range locations {
		parent := "projects/" + project + "/locations/" + location
		services, err := restListResources(ctx, restClient, agentRegistryBaseURL+parent+"/services")
		if err != nil {
			log.Println(err)
		} else {
			for _, service := range services {
				g.Resources = append(g.Resources, g.newAgentResource(service.Name, "google_agent_registry_service"))
			}
		}
		bindings, err := restListResources(ctx, restClient, agentRegistryBaseURL+parent+"/bindings")
		if err != nil {
			log.Println(err)
		} else {
			for _, binding := range bindings {
				g.Resources = append(g.Resources, g.newAgentResource(binding.Name, "google_agent_registry_binding"))
			}
		}
		providers, err := restListResources(ctx, restClient, agentIdentityBaseURL+parent+"/authProviders")
		if err != nil {
			log.Println(err)
			continue
		}
		for _, provider := range providers {
			g.Resources = append(g.Resources, g.newAgentResource(provider.Name, "google_agent_identity_auth_provider"))
			policy, err := restGetIamPolicy(ctx, restClient, agentIdentityBaseURL+provider.Name+":getIamPolicy")
			if err != nil {
				log.Println(err)
				continue
			}
			g.Resources = append(g.Resources, expandIamPolicy(policy, provider.Name,
				"google_agent_identity_auth_provider_iam_policy",
				"google_agent_identity_auth_provider_iam_binding",
				"google_agent_identity_auth_provider_iam_member",
				map[string]string{"auth_provider": provider.Name},
				func(id, resourceType string, attrs map[string]string) terraformutils.Resource {
					return g.newAgentResourceWithAttrs(id, resourceType, attrs)
				})...)
		}
	}
}

func (g *AgentGenerator) newAgentResource(id, resourceType string) terraformutils.Resource {
	return g.newAgentResourceWithAttrs(id, resourceType, map[string]string{})
}

func (g *AgentGenerator) newAgentResourceWithAttrs(id, resourceType string, attrs map[string]string) terraformutils.Resource {
	parts := strings.Split(id, "/")
	merged := map[string]string{"project": g.GetArgs()["project"].(string)}
	for k, v := range attrs {
		merged[k] = v
	}
	return terraformutils.NewResource(
		id,
		parts[len(parts)-1],
		resourceType,
		g.ProviderName,
		merged,
		[]string{},
		map[string]interface{}{},
	)
}
