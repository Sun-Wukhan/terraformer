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
	"testing"

	"google.golang.org/api/compute/v1"
)

// TestAgentGenerator feeds canned Agent Registry / Agent Identity
// responses in the global location and verifies a registry service,
// a registry binding, an auth provider and its IAM trio.
func TestAgentGenerator(t *testing.T) {
	parent := "projects/test-project/locations/global"
	provider := parent + "/authProviders/ap1"
	transport := &cannedTransport{routes: []cannedRoute{
		{":getIamPolicy", `{"bindings":[{"role":"roles/agentidentity.user","members":["user:jane@example.com"]}]}`},
		{"/locations/global/services", `{"services":[{"name":"` + parent + `/services/svc1"}]}`},
		{"/locations/global/bindings", `{"bindings":[{"name":"` + parent + `/bindings/b1"}]}`},
		{"/locations/global/authProviders", `{"authProviders":[{"name":"` + provider + `"}]}`},
	}}
	g := &AgentGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	g.initResources(context.Background(), transport.client())
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_agent_registry_service":                   {parent + "/services/svc1"},
		"google_agent_registry_binding":                   {parent + "/bindings/b1"},
		"google_agent_identity_auth_provider":             {provider},
		"google_agent_identity_auth_provider_iam_policy":  {provider},
		"google_agent_identity_auth_provider_iam_binding": {provider},
		"google_agent_identity_auth_provider_iam_member":  {provider},
	})
}

// TestAgenticApplicationsGenerator pins the documented limitation:
// the generator is registered but emits nothing, because the
// agenticapplications API has no verifiable enumeration surface.
func TestAgenticApplicationsGenerator(t *testing.T) {
	g := &AgenticApplicationsGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	if err := g.InitResources(); err != nil {
		t.Fatalf("InitResources: %v", err)
	}
	if len(g.Resources) != 0 {
		t.Errorf("resources = %d, want 0", len(g.Resources))
	}
}

// TestVectorSearchGenerator feeds canned Vector Search 2.0
// responses and verifies a collection with a data object and an
// index.
func TestVectorSearchGenerator(t *testing.T) {
	parent := "projects/test-project/locations/us-central1"
	collection := parent + "/collections/col1"
	transport := &cannedTransport{routes: []cannedRoute{
		{"/collections/col1/dataObjects", `{"dataObjects":[{"name":"` + collection + `/dataObjects/1"}]}`},
		{"/collections/col1/indexes", `{"indexes":[{"name":"` + collection + `/indexes/idx1"}]}`},
		{"/locations/us-central1/collections", `{"collections":[{"name":"` + collection + `"}]}`},
	}}
	g := &VectorSearchGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	g.initResources(context.Background(), transport.client())
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_vector_search_collection":  {collection},
		"google_vector_search_data_object": {collection + "/dataObjects/1"},
		"google_vector_search_index":       {collection + "/indexes/idx1"},
	})
}
