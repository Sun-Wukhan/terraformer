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
	datafusionv1 "google.golang.org/api/datafusion/v1"
	"google.golang.org/api/option"
)

// TestDataFusionGenerator feeds canned Data Fusion responses and
// verifies instances expand their IAM trio with full names as IDs.
func TestDataFusionGenerator(t *testing.T) {
	loc := "projects/test-project/locations/us-central1"
	inst := loc + "/instances/df1"
	transport := &cannedTransport{routes: []cannedRoute{
		{":getIamPolicy", `{"bindings":[{"role":"roles/datafusion.admin","members":["user:a@example.com"]}]}`},
		{"/locations/us-central1/instances", `{"instances":[{"name":"` + inst + `"}]}`},
	}}
	ctx := context.Background()
	svc, err := datafusionv1.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("datafusion.NewService: %v", err)
	}
	g := &DataFusionGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	if err := g.initResources(ctx, svc); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_data_fusion_instance":             {inst},
		"google_data_fusion_instance_iam_policy":  {inst},
		"google_data_fusion_instance_iam_binding": {inst},
		"google_data_fusion_instance_iam_member":  {inst},
	})
}
