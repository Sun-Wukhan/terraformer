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

func TestComputeIamGenerator(t *testing.T) {
	ctx := context.Background()
	policyBody := `{"bindings":[{"role":"roles/compute.viewer","members":["user:a@example.com"]}]}`
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: ":getAncestry", body: testAncestryBody},
		// Every getIamPolicy call (client or raw REST) returns the
		// same one-binding policy; this route must precede the lists.
		{substr: "getIamPolicy", body: policyBody},
		{substr: "/locations/global/firewallPolicies", body: `{"items":[{"name":"locations/global/firewallPolicies/111"}]}`},
		{substr: "/global/backendBuckets", body: `{"items":[{"name":"bb1"}]}`},
		{substr: "/global/backendServices", body: `{"items":[{"name":"bs1"}]}`},
		{substr: "/global/images", body: `{"items":[{"name":"img1"}]}`},
		{substr: "/global/instanceTemplates", body: `{"items":[{"name":"it1"}]}`},
		{substr: "/global/machineImages", body: `{"items":[{"name":"mi1"}]}`},
		{substr: "/global/snapshots", body: `{"items":[{"name":"snap1"}]}`},
		{substr: "/global/firewallPolicies", body: `{"items":[{"name":"nfw1"}]}`},
		{substr: "/regions/us-central1/backendServices", body: `{"items":[{"name":"rbs1"}]}`},
		{substr: "/regions/us-central1/disks", body: `{"items":[{"name":"rd1"}]}`},
		{substr: "/regions/us-central1/instantSnapshots", body: `{"items":[{"name":"ris1"}]}`},
		{substr: "/regions/us-central1/subnetworks", body: `{"items":[{"name":"sub1"}]}`},
		{substr: "/regions/us-central1/firewallPolicies", body: `{"items":[{"name":"rnfw1"}]}`},
		{substr: "/regions/us-central1/backendBuckets", body: `{"items":[{"name":"rbb1"}]}`},
		{substr: "/zones/us-central1-a/disks", body: `{"items":[{"name":"d1"}]}`},
		{substr: "/zones/us-central1-a/instances", body: `{"items":[{"name":"vm1"}]}`},
		{substr: "/zones/us-central1-a/instantSnapshots", body: `{"items":[{"name":"is1"}]}`},
		{substr: "/zones/us-central1-a/storagePools", body: `{"items":[{"name":"sp1"}]}`},
	}}
	computeService := computeTestService(t, transport)
	crmService := crmTestService(t, transport)

	g := &ComputeIamGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region": compute.Region{
			Name:  "us-central1",
			Zones: []string{"https://www.googleapis.com/compute/v1/projects/test-project/zones/us-central1-a"},
		},
	})
	if err := g.initResources(ctx, computeService, crmService, transport.client()); err != nil {
		t.Fatalf("initResources: %v", err)
	}

	// Every parent expands to the _iam_policy/_iam_binding/_iam_member
	// trio, all carrying the parent's provider ID.
	parents := map[string]string{
		"google_compute_backend_bucket":                 "projects/test-project/global/backendBuckets/bb1",
		"google_compute_backend_service":                "projects/test-project/global/backendServices/bs1",
		"google_compute_image":                          "projects/test-project/global/images/img1",
		"google_compute_instance_template":              "projects/test-project/global/instanceTemplates/it1",
		"google_compute_machine_image":                  "projects/test-project/global/machineImages/mi1",
		"google_compute_snapshot":                       "projects/test-project/global/snapshots/snap1",
		"google_compute_network_firewall_policy":        "projects/test-project/global/firewallPolicies/nfw1",
		"google_compute_region_backend_service":         "projects/test-project/regions/us-central1/backendServices/rbs1",
		"google_compute_region_disk":                    "projects/test-project/regions/us-central1/disks/rd1",
		"google_compute_region_instant_snapshot":        "projects/test-project/regions/us-central1/instantSnapshots/ris1",
		"google_compute_subnetwork":                     "projects/test-project/regions/us-central1/subnetworks/sub1",
		"google_compute_region_network_firewall_policy": "projects/test-project/regions/us-central1/firewallPolicies/rnfw1",
		"google_compute_region_backend_bucket":          "projects/test-project/regions/us-central1/backendBuckets/rbb1",
		"google_compute_disk":                           "projects/test-project/zones/us-central1-a/disks/d1",
		"google_compute_instance":                       "projects/test-project/zones/us-central1-a/instances/vm1",
		"google_compute_instant_snapshot":               "projects/test-project/zones/us-central1-a/instantSnapshots/is1",
		"google_compute_storage_pool":                   "projects/test-project/zones/us-central1-a/storagePools/sp1",
		"google_compute_firewall_policy":                "locations/global/firewallPolicies/111",
	}
	expected := map[string][]string{}
	for prefix, parentID := range parents {
		expected[prefix+"_iam_policy"] = []string{parentID}
		expected[prefix+"_iam_binding"] = []string{parentID}
		expected[prefix+"_iam_member"] = []string{parentID}
	}
	assertResourceIDs(t, g.Resources, expected)
}
