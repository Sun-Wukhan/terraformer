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

	aiplatformbeta "google.golang.org/api/aiplatform/v1beta1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

// TestVertexAiDepthGenerator feeds canned v1beta1 and raw REST
// responses and verifies the depth pass: an endpoint IAM trio from
// the v1beta1 getIamPolicy, the RAG engine config singleton, a
// semantic governance policy and the policy engine singleton.
func TestVertexAiDepthGenerator(t *testing.T) {
	parent := "projects/test-project/locations/us-central1"
	endpoint := parent + "/endpoints/456"
	transport := &cannedTransport{routes: []cannedRoute{
		{":getIamPolicy", `{"bindings":[{"role":"roles/aiplatform.admin","members":["user:jane@example.com"]}]}`},
		{"/endpoints", `{"endpoints":[{"name":"` + endpoint + `"}]}`},
		{"/ragEngineConfig", `{"name":"` + parent + `/ragEngineConfig"}`},
		{"/semanticGovernancePolicies", `{"semanticGovernancePolicies":[{"name":"` + parent + `/semanticGovernancePolicies/p1"}]}`},
		{"/semanticGovernancePolicyEngine", `{"name":"` + parent + `/semanticGovernancePolicyEngine"}`},
	}}
	ctx := context.Background()
	betaService, err := aiplatformbeta.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("aiplatform v1beta1 NewService: %v", err)
	}
	g := &VertexAiGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	g.initDepthResources(ctx, betaService, transport.client())
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_vertex_ai_endpoint_iam_policy":  {endpoint},
		"google_vertex_ai_endpoint_iam_binding": {endpoint},
		"google_vertex_ai_endpoint_iam_member":  {endpoint},
		"google_vertex_ai_rag_engine_config":    {parent + "/ragEngineConfig"},
		"google_vertex_ai_semantic_governance_policy": {
			parent + "/semanticGovernancePolicies/p1",
		},
		"google_vertex_ai_semantic_governance_policy_engine": {
			parent + "/semanticGovernancePolicyEngine",
		},
	})
}
