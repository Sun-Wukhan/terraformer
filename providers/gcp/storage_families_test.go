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
	parallelstorev1 "google.golang.org/api/parallelstore/v1"
)

// TestFilestoreGenerator feeds canned Filestore raw REST responses
// and verifies instances, snapshots and backups.
func TestFilestoreGenerator(t *testing.T) {
	loc := "projects/test-project/locations/us-central1"
	instance := loc + "/instances/i1"
	transport := &cannedTransport{routes: []cannedRoute{
		{"/instances/i1/snapshots", `{"snapshots":[{"name":"` + instance + `/snapshots/s1"}]}`},
		{"/instances", `{"instances":[{"name":"` + instance + `"}]}`},
		{"/backups", `{"backups":[{"name":"` + loc + `/backups/b1"}]}`},
	}}
	g := &FilestoreGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project", "region": compute.Region{Name: "us-central1"}})
	g.initResources(context.Background(), transport.client())
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_filestore_instance": {instance},
		"google_filestore_snapshot": {instance + "/snapshots/s1"},
		"google_filestore_backup":   {loc + "/backups/b1"},
	})
}

// TestParallelstoreGenerator feeds a canned Parallelstore response
// and verifies the instance type and ID.
func TestParallelstoreGenerator(t *testing.T) {
	loc := "projects/test-project/locations/us-central1"
	transport := &cannedTransport{routes: []cannedRoute{
		{"/instances", `{"instances":[{"name":"` + loc + `/instances/p1"}]}`},
	}}
	ctx := context.Background()
	clientOpts := []option.ClientOption{option.WithHTTPClient(transport.client()), option.WithoutAuthentication()}
	storeService, err := parallelstorev1.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("parallelstore.NewService: %v", err)
	}
	g := &ParallelstoreGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project", "region": compute.Region{Name: "us-central1"}})
	if err := g.initResources(ctx, storeService); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_parallelstore_instance": {loc + "/instances/p1"},
	})
}

// TestLustreGenerator feeds a canned Managed Lustre raw REST
// response and verifies the instance type and ID.
func TestLustreGenerator(t *testing.T) {
	loc := "projects/test-project/locations/us-central1"
	transport := &cannedTransport{routes: []cannedRoute{
		{"/instances", `{"instances":[{"name":"` + loc + `/instances/l1"}]}`},
	}}
	g := &LustreGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project", "region": compute.Region{Name: "us-central1"}})
	g.initResources(context.Background(), transport.client())
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_lustre_instance": {loc + "/instances/l1"},
	})
}

// TestHypercomputeClusterGenerator verifies the documented skip:
// the generator emits nothing.
func TestHypercomputeClusterGenerator(t *testing.T) {
	g := &HypercomputeClusterGenerator{}
	g.ProviderName = "google"
	if err := g.InitResources(); err != nil {
		t.Fatalf("InitResources: %v", err)
	}
	if len(g.Resources) != 0 {
		t.Errorf("resources = %v, want none (documented skip)", g.Resources)
	}
}
