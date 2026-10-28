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
	memcachev1 "google.golang.org/api/memcache/v1"
	"google.golang.org/api/option"
)

// TestMemcacheGenerator feeds canned Memcache responses and verifies
// instances are emitted with full resource names as IDs.
func TestMemcacheGenerator(t *testing.T) {
	loc := "projects/test-project/locations/us-central1"
	transport := &cannedTransport{routes: []cannedRoute{
		{"/locations/us-central1/instances", `{"instances":[{"name":"` + loc + `/instances/mc1"}]}`},
	}}
	ctx := context.Background()
	svc, err := memcachev1.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("memcache.NewService: %v", err)
	}
	g := &MemcacheGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	if err := g.initResources(ctx, svc); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_memcache_instance": {loc + "/instances/mc1"},
	})
}
