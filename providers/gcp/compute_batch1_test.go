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
)

func computeTestService(t *testing.T, transport *cannedTransport) *compute.Service {
	t.Helper()
	service, err := compute.NewService(context.Background(),
		option.WithHTTPClient(transport.client()),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatalf("compute.NewService: %v", err)
	}
	return service
}

func TestComputeBatch1Generators(t *testing.T) {
	ctx := context.Background()
	args := map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	}

	t.Run("snapshots", func(t *testing.T) {
		transport := &cannedTransport{routes: []cannedRoute{
			{substr: "/snapshots", body: `{"items":[{"name":"snap1"}]}`},
		}}
		service := computeTestService(t, transport)
		g := &SnapshotsGenerator{}
		g.ProviderName = "google"
		g.SetArgs(args)
		resources := g.createResources(ctx, service.Snapshots.List("test-project"))
		if len(resources) != 1 || resources[0].InstanceInfo.Type != "google_compute_snapshot" || resources[0].InstanceState.ID != "snap1" {
			t.Errorf("snapshots: got %+v", resources)
		}
	})

	t.Run("machineImages", func(t *testing.T) {
		transport := &cannedTransport{routes: []cannedRoute{
			{substr: "/machineImages", body: `{"items":[{"name":"mi1"}]}`},
		}}
		service := computeTestService(t, transport)
		g := &MachineImagesGenerator{}
		g.ProviderName = "google"
		g.SetArgs(args)
		resources := g.createResources(ctx, service.MachineImages.List("test-project"))
		if len(resources) != 1 || resources[0].InstanceInfo.Type != "google_compute_machine_image" || resources[0].InstanceState.ID != "mi1" {
			t.Errorf("machineImages: got %+v", resources)
		}
	})

	t.Run("globalNetworkEndpointGroups", func(t *testing.T) {
		transport := &cannedTransport{routes: []cannedRoute{
			{substr: "/global/networkEndpointGroups", body: `{"items":[{"name":"gneg1"}]}`},
		}}
		service := computeTestService(t, transport)
		g := &GlobalNetworkEndpointGroupsGenerator{}
		g.ProviderName = "google"
		g.SetArgs(args)
		resources := g.createResources(ctx, service.GlobalNetworkEndpointGroups.List("test-project"))
		if len(resources) != 1 || resources[0].InstanceInfo.Type != "google_compute_global_network_endpoint_group" || resources[0].InstanceState.ID != "gneg1" {
			t.Errorf("globalNetworkEndpointGroups: got %+v", resources)
		}
	})
}

func TestRouterSubresourcesEnumeration(t *testing.T) {
	ctx := context.Background()
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: "/routers", body: `{"items":[{"name":"router1","nats":[{"name":"nat1"}],"bgpPeers":[{"name":"peer1"}],"interfaces":[{"name":"if1"}]}]}`},
	}}
	service := computeTestService(t, transport)

	g := &RouterSubresourcesGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	if err := g.initResources(ctx, service); err != nil {
		t.Fatalf("initResources: %v", err)
	}

	got := map[string]string{}
	for _, r := range g.Resources {
		got[r.InstanceInfo.Type] = r.InstanceState.ID
	}
	want := map[string]string{
		"google_compute_router_nat":       "test-project/us-central1/router1/nat1",
		"google_compute_router_peer":      "test-project/us-central1/router1/peer1",
		"google_compute_router_interface": "test-project/us-central1/router1/if1",
	}
	for resourceType, wantID := range want {
		if got[resourceType] != wantID {
			t.Errorf("%s: got ID %q, want %q", resourceType, got[resourceType], wantID)
		}
	}
}
