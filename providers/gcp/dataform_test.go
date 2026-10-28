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
	dataformv1beta1 "google.golang.org/api/dataform/v1beta1"
	"google.golang.org/api/option"
)

// TestDataformGenerator feeds canned Dataform responses and verifies
// repositories expand their IAM trio, release and workflow configs
// are emitted, and the location config singleton is emitted.
func TestDataformGenerator(t *testing.T) {
	loc := "projects/test-project/locations/us-central1"
	repo := loc + "/repositories/repo1"
	transport := &cannedTransport{routes: []cannedRoute{
		{":getIamPolicy", `{"bindings":[{"role":"roles/dataform.editor","members":["user:a@example.com"]}]}`},
		{"/repositories/repo1/releaseConfigs", `{"releaseConfigs":[{"name":"` + repo + `/releaseConfigs/rc1"}]}`},
		{"/repositories/repo1/workflowConfigs", `{"workflowConfigs":[{"name":"` + repo + `/workflowConfigs/wc1"}]}`},
		{"/locations/us-central1/repositories", `{"repositories":[{"name":"` + repo + `"}]}`},
		{"/locations/us-central1:getConfig", `{}`},
		{"/locations/us-central1", `{}`},
	}}
	ctx := context.Background()
	svc, err := dataformv1beta1.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("dataform.NewService: %v", err)
	}
	g := &DataformGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	if err := g.initResources(ctx, svc); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_dataform_repository":                 {repo},
		"google_dataform_repository_iam_policy":      {repo},
		"google_dataform_repository_iam_binding":     {repo},
		"google_dataform_repository_iam_member":      {repo},
		"google_dataform_repository_release_config":  {repo + "/releaseConfigs/rc1"},
		"google_dataform_repository_workflow_config": {repo + "/workflowConfigs/wc1"},
		"google_dataform_config":                     {loc + "/config"},
	})
}
