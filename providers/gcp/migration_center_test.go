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
	migrationcenterv1 "google.golang.org/api/migrationcenter/v1"
	"google.golang.org/api/option"
)

// TestMigrationCenterGenerator feeds canned Migration Center
// responses and verifies groups, preference sets, sources, discovery
// clients, import jobs and data files, report configs and reports,
// and the settings singleton.
func TestMigrationCenterGenerator(t *testing.T) {
	loc := "projects/test-project/locations/us-central1"
	job := loc + "/importJobs/job1"
	config := loc + "/reportConfigs/rc1"
	transport := &cannedTransport{routes: []cannedRoute{
		{"/locations/us-central1/groups", `{"groups":[{"name":"` + loc + `/groups/g1"}]}`},
		{"/locations/us-central1/preferenceSets", `{"preferenceSets":[{"name":"` + loc + `/preferenceSets/ps1"}]}`},
		{"/locations/us-central1/sources", `{"sources":[{"name":"` + loc + `/sources/src1"}]}`},
		{"/locations/us-central1/discoveryClients", `{"discoveryClients":[{"name":"` + loc + `/discoveryClients/dc1"}]}`},
		{"/importJobs/job1/importDataFiles", `{"importDataFiles":[{"name":"` + job + `/importDataFiles/file1"}]}`},
		{"/locations/us-central1/importJobs", `{"importJobs":[{"name":"` + job + `"}]}`},
		{"/reportConfigs/rc1/reports", `{"reports":[{"name":"` + config + `/reports/rep1"}]}`},
		{"/locations/us-central1/reportConfigs", `{"reportConfigs":[{"name":"` + config + `"}]}`},
		{"/locations/us-central1:getSettings", `{}`},
		{"/locations/us-central1", `{}`},
	}}
	ctx := context.Background()
	svc, err := migrationcenterv1.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("migrationcenter.NewService: %v", err)
	}
	g := &MigrationCenterGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	if err := g.initResources(ctx, svc); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_migration_center_group":            {loc + "/groups/g1"},
		"google_migration_center_preference_set":   {loc + "/preferenceSets/ps1"},
		"google_migration_center_source":           {loc + "/sources/src1"},
		"google_migration_center_discovery_client": {loc + "/discoveryClients/dc1"},
		"google_migration_center_import_job":       {job},
		"google_migration_center_import_data_file": {job + "/importDataFiles/file1"},
		"google_migration_center_report_config":    {config},
		"google_migration_center_report":           {config + "/reports/rep1"},
		"google_migration_center_settings":         {loc + "/settings"},
	})
}
