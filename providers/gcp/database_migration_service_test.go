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
	datamigrationv1 "google.golang.org/api/datamigration/v1"
	"google.golang.org/api/option"
)

// TestDatabaseMigrationServiceGenerator feeds canned DMS responses
// and verifies connection profiles, migration jobs and private
// connections with full resource names as IDs.
func TestDatabaseMigrationServiceGenerator(t *testing.T) {
	loc := "projects/test-project/locations/us-central1"
	transport := &cannedTransport{routes: []cannedRoute{
		{"/locations/us-central1/connectionProfiles", `{"connectionProfiles":[{"name":"` + loc + `/connectionProfiles/cp1"}]}`},
		{"/locations/us-central1/migrationJobs", `{"migrationJobs":[{"name":"` + loc + `/migrationJobs/mj1"}]}`},
		{"/locations/us-central1/privateConnections", `{"privateConnections":[{"name":"` + loc + `/privateConnections/pc1"}]}`},
	}}
	ctx := context.Background()
	svc, err := datamigrationv1.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("datamigration.NewService: %v", err)
	}
	g := &DatabaseMigrationServiceGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	if err := g.initResources(ctx, svc); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_database_migration_service_connection_profile": {loc + "/connectionProfiles/cp1"},
		"google_database_migration_service_migration_job":      {loc + "/migrationJobs/mj1"},
		"google_database_migration_service_private_connection": {loc + "/privateConnections/pc1"},
	})
}
