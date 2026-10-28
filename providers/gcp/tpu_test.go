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
	"google.golang.org/api/option"
	tpuv2 "google.golang.org/api/tpu/v2"
)

// TestTpuGenerator feeds canned TPU v2 responses for one zone and
// verifies VMs and queued resources.
func TestTpuGenerator(t *testing.T) {
	zoneParent := "projects/test-project/locations/us-central1-a"
	transport := &cannedTransport{routes: []cannedRoute{
		{"/nodes", `{"nodes":[{"name":"` + zoneParent + `/nodes/vm1"}]}`},
		{"/queuedResources", `{"queuedResources":[{"name":"` + zoneParent + `/queuedResources/q1"}]}`},
	}}
	ctx := context.Background()
	clientOpts := []option.ClientOption{option.WithHTTPClient(transport.client()), option.WithoutAuthentication()}
	tpuService, err := tpuv2.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("tpu.NewService: %v", err)
	}
	g := &TpuGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region": compute.Region{Name: "us-central1", Zones: []string{
			"https://www.googleapis.com/compute/v1/projects/test-project/zones/us-central1-a",
		}},
	})
	if err := g.initResources(ctx, tpuService); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_tpu_v2_vm":              {zoneParent + "/nodes/vm1"},
		"google_tpu_v2_queued_resource": {zoneParent + "/queuedResources/q1"},
	})
}
