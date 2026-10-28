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

	alloydbv1 "google.golang.org/api/alloydb/v1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

// TestAlloydbGenerator feeds canned AlloyDB responses and verifies
// backups, clusters, instances and users with full names as IDs.
func TestAlloydbGenerator(t *testing.T) {
	loc := "projects/test-project/locations/us-central1"
	cluster := loc + "/clusters/cl1"
	transport := &cannedTransport{routes: []cannedRoute{
		{"/clusters/cl1/instances", `{"instances":[{"name":"` + cluster + `/instances/inst1"}]}`},
		{"/clusters/cl1/users", `{"users":[{"name":"` + cluster + `/users/u1"}]}`},
		{"/locations/us-central1/clusters", `{"clusters":[{"name":"` + cluster + `"}]}`},
		{"/locations/us-central1/backups", `{"backups":[{"name":"` + loc + `/backups/bk1"}]}`},
	}}
	ctx := context.Background()
	svc, err := alloydbv1.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("alloydb.NewService: %v", err)
	}
	g := &AlloydbGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	if err := g.initResources(ctx, svc); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_alloydb_backup":   {loc + "/backups/bk1"},
		"google_alloydb_cluster":  {cluster},
		"google_alloydb_instance": {cluster + "/instances/inst1"},
		"google_alloydb_user":     {cluster + "/users/u1"},
	})
}
