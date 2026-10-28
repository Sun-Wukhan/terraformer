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
	gkeonpremv1 "google.golang.org/api/gkeonprem/v1"
	"google.golang.org/api/option"
)

// TestGkeOnpremGenerator feeds canned GKE on-prem responses and
// verifies admin clusters, clusters and node pools.
func TestGkeOnpremGenerator(t *testing.T) {
	loc := "projects/test-project/locations/us-central1"
	bm := loc + "/bareMetalClusters/bm1"
	vmw := loc + "/vmwareClusters/v1"
	transport := &cannedTransport{routes: []cannedRoute{
		{"/bareMetalClusters/bm1/bareMetalNodePools", `{"bareMetalNodePools":[{"name":"` + bm + `/bareMetalNodePools/np1"}]}`},
		{"/vmwareClusters/v1/vmwareNodePools", `{"vmwareNodePools":[{"name":"` + vmw + `/vmwareNodePools/np1"}]}`},
		{"/bareMetalAdminClusters", `{"bareMetalAdminClusters":[{"name":"` + loc + `/bareMetalAdminClusters/admin1"}]}`},
		{"/vmwareAdminClusters", `{"vmwareAdminClusters":[{"name":"` + loc + `/vmwareAdminClusters/admin1"}]}`},
		{"/bareMetalClusters", `{"bareMetalClusters":[{"name":"` + bm + `"}]}`},
		{"/vmwareClusters", `{"vmwareClusters":[{"name":"` + vmw + `"}]}`},
	}}
	ctx := context.Background()
	clientOpts := []option.ClientOption{option.WithHTTPClient(transport.client()), option.WithoutAuthentication()}
	onpremService, err := gkeonpremv1.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("gkeonprem.NewService: %v", err)
	}
	g := &GkeOnpremGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project", "region": compute.Region{Name: "us-central1"}})
	if err := g.initResources(ctx, onpremService); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_gkeonprem_bare_metal_admin_cluster": {loc + "/bareMetalAdminClusters/admin1"},
		"google_gkeonprem_bare_metal_cluster":       {bm},
		"google_gkeonprem_bare_metal_node_pool":     {bm + "/bareMetalNodePools/np1"},
		"google_gkeonprem_vmware_admin_cluster":     {loc + "/vmwareAdminClusters/admin1"},
		"google_gkeonprem_vmware_cluster":           {vmw},
		"google_gkeonprem_vmware_node_pool":         {vmw + "/vmwareNodePools/np1"},
	})
}
