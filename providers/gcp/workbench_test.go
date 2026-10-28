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
	notebooksv2 "google.golang.org/api/notebooks/v2"
	"google.golang.org/api/option"
)

// TestWorkbenchGenerator feeds a canned Notebooks v2 instance list
// and IAM policy and verifies the instance plus its IAM trio.
func TestWorkbenchGenerator(t *testing.T) {
	parent := "projects/test-project/locations/us-central1"
	instance := parent + "/instances/wb1"
	transport := &cannedTransport{routes: []cannedRoute{
		{":getIamPolicy", `{"bindings":[{"role":"roles/notebooks.admin","members":["user:jane@example.com"]}]}`},
		{"/locations/us-central1/instances", `{"instances":[{"name":"` + instance + `"}]}`},
	}}
	ctx := context.Background()
	service, err := notebooksv2.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("notebooks v2 NewService: %v", err)
	}
	g := &WorkbenchGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	g.initResources(ctx, service)
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_workbench_instance":             {instance},
		"google_workbench_instance_iam_policy":  {instance},
		"google_workbench_instance_iam_binding": {instance},
		"google_workbench_instance_iam_member":  {instance},
	})
}
