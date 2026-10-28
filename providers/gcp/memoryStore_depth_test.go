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
	"google.golang.org/api/redis/v1"
)

// TestMemoryStoreDepthGenerator feeds canned Redis responses and
// verifies clusters (generated client) and cluster ACL policies
// (raw REST) with full resource names as IDs.
func TestMemoryStoreDepthGenerator(t *testing.T) {
	loc := "projects/test-project/locations/us-central1"
	transport := &cannedTransport{routes: []cannedRoute{
		{"/locations/us-central1/clusters", `{"clusters":[{"name":"` + loc + `/clusters/rc1"}]}`},
		{"/locations/us-central1/aclPolicies", `{"aclPolicies":[{"name":"` + loc + `/aclPolicies/acl1"}]}`},
	}}
	ctx := context.Background()
	svc, err := redis.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("redis.NewService: %v", err)
	}
	g := &MemoryStoreGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	g.initDepthResources(ctx, svc, transport.client())
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_redis_cluster":            {loc + "/clusters/rc1"},
		"google_redis_cluster_acl_policy": {loc + "/aclPolicies/acl1"},
	})
}
