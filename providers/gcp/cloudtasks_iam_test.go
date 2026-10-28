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

	"google.golang.org/api/cloudtasks/v2"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

// TestCloudTasksQueueIam asserts the combined queue IAM resource is
// emitted for a queue carrying IAM bindings.
func TestCloudTasksQueueIam(t *testing.T) {
	ctx := context.Background()
	queue := "projects/test-project/locations/us-central1/queues/q1"
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: ":getIamPolicy", body: `{"bindings":[{"role":"roles/cloudtasks.enqueuer","members":["user:a@example.com"]}]}`},
		{substr: "/queues", body: `{"queues":[{"name":"` + queue + `"}]}`},
	}}
	service, err := cloudtasks.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	g := &CloudTaskGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	if err := g.loadCloudTaskQueueIam(ctx, service); err != nil {
		t.Fatalf("loadCloudTaskQueueIam: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_cloud_tasks_queue_iam": {queue},
	})
}
