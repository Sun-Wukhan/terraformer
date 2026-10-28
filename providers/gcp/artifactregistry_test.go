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

	artifactregistryv1 "google.golang.org/api/artifactregistry/v1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

// TestArtifactRegistryGenerator feeds canned Artifact Registry
// responses and verifies repositories, rules, the IAM trio and the
// two location singletons.
func TestArtifactRegistryGenerator(t *testing.T) {
	loc := "projects/test-project/locations/us-central1"
	repo := loc + "/repositories/r1"
	policy := `{"bindings":[{"role":"roles/artifactregistry.reader","members":["user:a@example.com"]}]}`
	transport := &cannedTransport{routes: []cannedRoute{
		{":getIamPolicy", policy},
		{"/repositories/r1/rules", `{"rules":[{"name":"` + repo + `/rules/rule1"}]}`},
		{"/repositories", `{"repositories":[{"name":"` + repo + `"}]}`},
		{"/projectConfig", `{"name":"` + loc + `/projectConfig"}`},
		{"/vpcscConfig", `{"name":"` + loc + `/vpcscConfig"}`},
	}}
	ctx := context.Background()
	clientOpts := []option.ClientOption{option.WithHTTPClient(transport.client()), option.WithoutAuthentication()}
	registryService, err := artifactregistryv1.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("artifactregistry.NewService: %v", err)
	}
	g := &ArtifactRegistryGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project", "region": compute.Region{Name: "us-central1"}})
	if err := g.initResources(ctx, registryService, transport.client()); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_artifact_registry_repository":             {repo},
		"google_artifact_registry_repository_iam_policy":  {repo},
		"google_artifact_registry_repository_iam_binding": {repo},
		"google_artifact_registry_repository_iam_member":  {repo},
		"google_artifact_registry_rule":                   {repo + "/rules/rule1"},
		"google_artifact_registry_project_config":         {loc + "/projectConfig"},
		"google_artifact_registry_vpcsc_config":           {loc + "/vpcscConfig"},
	})
}
