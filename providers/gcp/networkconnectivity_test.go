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
	"sort"
	"testing"

	"google.golang.org/api/compute/v1"
	ncc "google.golang.org/api/networkconnectivity/v1"
	"google.golang.org/api/option"
)

// TestNetworkConnectivityGenerator feeds canned Network Connectivity
// responses through the real enumeration code and verifies the emitted
// resource types and IDs, including hub groups nested under hubs and
// the hub IAM expansion.
func TestNetworkConnectivityGenerator(t *testing.T) {
	global := "projects/test-project/locations/global/"
	regional := "projects/test-project/locations/us-central1/"
	hub := global + "hubs/hub-1"
	transport := &cannedTransport{routes: []cannedRoute{
		{":getIamPolicy", `{"bindings":[{"role":"roles/networkconnectivity.hubAdmin","members":["user:a@example.com","user:b@example.com"]}]}`},
		{"/groups", `{"groups":[{"name":"` + hub + `/groups/group-1"}]}`},
		{"/hubs", `{"hubs":[{"name":"` + hub + `"}]}`},
		{"/policyBasedRoutes", `{"policyBasedRoutes":[{"name":"` + global + `policyBasedRoutes/pbr-1"}]}`},
		{"/locations/global/internalRanges", `{"internalRanges":[{"name":"` + global + `internalRanges/ir-global"}]}`},
		{"/locations/us-central1/internalRanges", `{"internalRanges":[{"name":"` + regional + `internalRanges/ir-regional"}]}`},
		{"/regionalEndpoints", `{"regionalEndpoints":[{"name":"` + regional + `regionalEndpoints/re-1"}]}`},
		{"/serviceConnectionPolicies", `{"serviceConnectionPolicies":[{"name":"` + regional + `serviceConnectionPolicies/scp-1"}]}`},
		{"/locations/us-central1/spokes", `{"spokes":[{"name":"` + regional + `spokes/spoke-regional"}]}`},
		{"/locations/global/spokes", `{"spokes":[{"name":"` + global + `spokes/spoke-global"}]}`},
	}}

	ctx := context.Background()
	service, err := ncc.NewService(ctx,
		option.WithHTTPClient(transport.client()),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatalf("ncc.NewService: %v", err)
	}

	g := &NetworkConnectivityGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	if err := g.initResources(ctx, service); err != nil {
		t.Fatalf("initResources: %v", err)
	}

	expected := map[string][]string{
		"google_network_connectivity_hub":                       {hub},
		"google_network_connectivity_hub_iam_policy":            {hub},
		"google_network_connectivity_hub_iam_binding":           {hub},
		"google_network_connectivity_hub_iam_member":            {hub},
		"google_network_connectivity_group":                     {hub + "/groups/group-1"},
		"google_network_connectivity_policy_based_route":        {global + "policyBasedRoutes/pbr-1"},
		"google_network_connectivity_internal_range":            {global + "internalRanges/ir-global", regional + "internalRanges/ir-regional"},
		"google_network_connectivity_regional_endpoint":         {regional + "regionalEndpoints/re-1"},
		"google_network_connectivity_service_connection_policy": {regional + "serviceConnectionPolicies/scp-1"},
		"google_network_connectivity_spoke":                     {regional + "spokes/spoke-regional", global + "spokes/spoke-global"},
	}

	got := map[string]map[string]bool{}
	for _, resource := range g.Resources {
		if got[resource.InstanceInfo.Type] == nil {
			got[resource.InstanceInfo.Type] = map[string]bool{}
		}
		got[resource.InstanceInfo.Type][resource.InstanceState.ID] = true
	}
	if len(got) != len(expected) {
		t.Errorf("resource types = %d, want %d; got %v", len(got), len(expected), got)
	}
	for resourceType, ids := range expected {
		gotIDs := []string{}
		for id := range got[resourceType] {
			gotIDs = append(gotIDs, id)
		}
		sort.Strings(gotIDs)
		wantIDs := append([]string{}, ids...)
		sort.Strings(wantIDs)
		if len(gotIDs) != len(wantIDs) {
			t.Errorf("%s IDs = %v, want %v", resourceType, gotIDs, wantIDs)
			continue
		}
		for i := range wantIDs {
			if gotIDs[i] != wantIDs[i] {
				t.Errorf("%s IDs = %v, want %v", resourceType, gotIDs, wantIDs)
				break
			}
		}
	}
}
