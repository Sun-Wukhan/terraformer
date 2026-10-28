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
	netappv1 "google.golang.org/api/netapp/v1"
	"google.golang.org/api/option"
)

// TestNetappGenerator feeds canned NetApp Volumes responses
// (generated client plus raw REST host groups) and verifies all
// eleven resource types.
func TestNetappGenerator(t *testing.T) {
	loc := "projects/test-project/locations/us-central1"
	vault := loc + "/backupVaults/vault1"
	volume := loc + "/volumes/vol1"
	transport := &cannedTransport{routes: []cannedRoute{
		{"/backupVaults/vault1/backups", `{"backups":[{"name":"` + vault + `/backups/b1"}]}`},
		{"/volumes/vol1/quotaRules", `{"quotaRules":[{"name":"` + volume + `/quotaRules/q1"}]}`},
		{"/volumes/vol1/replications", `{"replications":[{"name":"` + volume + `/replications/r1"}]}`},
		{"/volumes/vol1/snapshots", `{"snapshots":[{"name":"` + volume + `/snapshots/s1"}]}`},
		{"/activeDirectories", `{"activeDirectories":[{"name":"` + loc + `/activeDirectories/ad1"}]}`},
		{"/kmsConfigs", `{"kmsConfigs":[{"name":"` + loc + `/kmsConfigs/k1"}]}`},
		{"/storagePools", `{"storagePools":[{"name":"` + loc + `/storagePools/sp1"}]}`},
		{"/backupPolicies", `{"backupPolicies":[{"name":"` + loc + `/backupPolicies/bp1"}]}`},
		{"/backupVaults", `{"backupVaults":[{"name":"` + vault + `"}]}`},
		{"/volumes", `{"volumes":[{"name":"` + volume + `"}]}`},
		{"/hostGroups", `{"hostGroups":[{"name":"` + loc + `/hostGroups/hg1"}]}`},
	}}
	ctx := context.Background()
	clientOpts := []option.ClientOption{option.WithHTTPClient(transport.client()), option.WithoutAuthentication()}
	netappService, err := netappv1.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("netapp.NewService: %v", err)
	}
	g := &NetappGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project", "region": compute.Region{Name: "us-central1"}})
	if err := g.initResources(ctx, netappService, transport.client()); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_netapp_active_directory":   {loc + "/activeDirectories/ad1"},
		"google_netapp_kmsconfig":          {loc + "/kmsConfigs/k1"},
		"google_netapp_storage_pool":       {loc + "/storagePools/sp1"},
		"google_netapp_backup_policy":      {loc + "/backupPolicies/bp1"},
		"google_netapp_backup_vault":       {vault},
		"google_netapp_backup":             {vault + "/backups/b1"},
		"google_netapp_volume":             {volume},
		"google_netapp_volume_quota_rule":  {volume + "/quotaRules/q1"},
		"google_netapp_volume_replication": {volume + "/replications/r1"},
		"google_netapp_volume_snapshot":    {volume + "/snapshots/s1"},
		"google_netapp_host_group":         {loc + "/hostGroups/hg1"},
	})
}
