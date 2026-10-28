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
)

// TestMemorystoreGenerator feeds canned Memorystore REST responses
// and verifies instances and ACL policies with full names as IDs.
func TestMemorystoreGenerator(t *testing.T) {
	loc := "projects/test-project/locations/us-central1"
	transport := &cannedTransport{routes: []cannedRoute{
		{"/locations/us-central1/instances", `{"instances":[{"name":"` + loc + `/instances/cache1"}]}`},
		{"/locations/us-central1/aclPolicies", `{"aclPolicies":[{"name":"` + loc + `/aclPolicies/acl1"}]}`},
	}}
	g := &MemorystoreGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	if err := g.initResources(context.Background(), transport.client()); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_memorystore_instance":   {loc + "/instances/cache1"},
		"google_memorystore_acl_policy": {loc + "/aclPolicies/acl1"},
	})
}
