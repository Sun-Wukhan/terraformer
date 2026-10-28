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

func TestComputeSubresourcesGenerator(t *testing.T) {
	ctx := context.Background()
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: ":getAncestry", body: testAncestryBody},
		{substr: "/global/snapshotRecycleBinPolicy", body: `{}`},
		// Per-group / per-MIG sub-collections precede their parents.
		{substr: "/instanceGroups/ig1/listInstances", body: `{"items":[{"instance":"https://www.googleapis.com/compute/v1/projects/test-project/zones/us-central1-a/instances/vm1","status":"RUNNING"}]}`},
		{substr: "/zones/us-central1-a/instanceGroupManagers/mig1/listPerInstanceConfigs", body: `{"items":[{"name":"pic1"}]}`},
		{substr: "/zones/us-central1-a/instanceGroupManagers/mig1/resizeRequests", body: `{"items":[{"name":"rr1"}]}`},
		{substr: "/regions/us-central1/instanceGroupManagers/rmig1/listPerInstanceConfigs", body: `{"items":[{"name":"rpic1"}]}`},
		{substr: "/regions/us-central1/instanceGroupManagers/rmig1/resizeRequests", body: `{"items":[{"name":"rrr1"}]}`},
		{substr: "/zones/us-central1-a/instanceGroups", body: `{"items":[{"name":"ig1","namedPorts":[{"name":"http","port":80}]}]}`},
		{substr: "/zones/us-central1-a/instanceGroupManagers", body: `{"items":[{"name":"mig1"}]}`},
		{substr: "/regions/us-central1/instanceGroupManagers", body: `{"items":[{"name":"rmig1"}]}`},
		{substr: "/global/networks", body: `{"items":[{"name":"net1","peerings":[{"name":"peer1","network":"https://www.googleapis.com/compute/v1/projects/other/global/networks/theirs"}]}]}`},
		{substr: "/zones/us-central1-a/disks", body: `{"items":[{"name":"disk1","resourcePolicies":["https://www.googleapis.com/compute/v1/projects/test-project/regions/us-central1/resourcePolicies/pol1"]}]}`},
		{substr: "/regions/us-central1/disks", body: `{"items":[{"name":"rdisk1","resourcePolicies":["https://www.googleapis.com/compute/v1/projects/test-project/regions/us-central1/resourcePolicies/rpol1"]}]}`},
		{substr: "/zones/us-central1-a/instances", body: `{"items":[{"name":"vm1","resourcePolicies":["https://www.googleapis.com/compute/v1/projects/test-project/regions/us-central1/resourcePolicies/pol2"]}]}`},
		{substr: "/instanceSettings", body: `{}`},
		{substr: "/global/snapshotSettings", body: `{}`},
		{substr: "/getXpnResources", body: `{"resources":[{"id":"svc-proj-1","type":"PROJECT"}]}`},
		// Catch-all last: the bare project GET.
		{substr: "/compute/v1/projects/test-project", body: `{"name":"test-project","commonInstanceMetadata":{"fingerprint":"abc","items":[{"key":"ssh-keys","value":"x"}]},"cloudArmorTier":"CA_ENTERPRISE_PAYGO","defaultNetworkTier":"PREMIUM","xpnProjectStatus":"HOST"}`},
	}}
	computeService := computeTestService(t, transport)
	crmService := crmTestService(t, transport)

	g := &ComputeSubresourcesGenerator{}
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

	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_compute_network_peering":                          {"test-project/net1/peer1"},
		"google_compute_instance_group_named_port":                {"test-project/us-central1-a/ig1/80/http"},
		"google_compute_instance_group_membership":                {"test-project/us-central1-a/ig1/vm1"},
		"google_compute_per_instance_config":                      {"test-project/us-central1-a/mig1/pic1"},
		"google_compute_resize_request":                           {"test-project/us-central1-a/mig1/rr1"},
		"google_compute_region_per_instance_config":               {"test-project/us-central1/rmig1/rpic1"},
		"google_compute_region_resize_request":                    {"test-project/us-central1/rmig1/rrr1"},
		"google_compute_disk_resource_policy_attachment":          {"test-project/us-central1-a/disk1/pol1"},
		"google_compute_region_disk_resource_policy_attachment":   {"test-project/us-central1/rdisk1/rpol1"},
		"google_compute_resource_policy_attachment":               {"test-project/us-central1-a/vm1/pol2"},
		"google_compute_instance_settings":                        {"projects/test-project/zones/us-central1-a/instanceSettings"},
		"google_compute_project_metadata":                         {"test-project"},
		"google_compute_project_cloud_armor_tier":                 {"test-project"},
		"google_compute_project_default_network_tier":             {"test-project"},
		"google_compute_shared_vpc_host_project":                  {"test-project"},
		"google_compute_shared_vpc_service_project":               {"test-project/svc-proj-1"},
		"google_compute_snapshot_settings":                        {"projects/test-project/global/snapshotSettings"},
		"google_compute_snapshot_recycle_bin_policy":              {"projects/test-project/global/snapshotRecycleBinPolicy"},
		"google_compute_organization_snapshot_recycle_bin_policy": {"organizations/789/global/snapshotRecycleBinPolicy"},
	})
}
