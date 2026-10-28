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
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// TestOracleDatabaseEnumeration asserts the generated-client
// collections, the raw-REST collections (including ODB subnets under
// their network) and the GoldenGate collections.
func TestOracleDatabaseEnumeration(t *testing.T) {
	ctx := context.Background()
	parent := "projects/test-project/locations/us-central1"
	odb := parent + "/odbNetworks/odb1"
	deployment := parent + "/deployments/dep1"
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: "/autonomousDatabases", body: `{"autonomousDatabases":[{"name":"` + parent + `/autonomousDatabases/adb1"}]}`},
		{substr: "/cloudExadataInfrastructures", body: `{"cloudExadataInfrastructures":[{"name":"` + parent + `/cloudExadataInfrastructures/exa1"}]}`},
		{substr: "/cloudVmClusters", body: `{"cloudVmClusters":[{"name":"` + parent + `/cloudVmClusters/vmc1"}]}`},
		{substr: "/dbSystems", body: `{"dbSystems":[{"name":"` + parent + `/dbSystems/dbs1"}]}`},
		{substr: "/exadbVmClusters", body: `{"exadbVmClusters":[{"name":"` + parent + `/exadbVmClusters/evc1"}]}`},
		{substr: "/exascaleDbStorageVaults", body: `{"exascaleDbStorageVaults":[{"name":"` + parent + `/exascaleDbStorageVaults/vault1"}]}`},
		{substr: "/odbSubnets", body: `{"odbSubnets":[{"name":"` + odb + `/odbSubnets/sub1"}]}`},
		{substr: "/odbNetworks", body: `{"odbNetworks":[{"name":"` + odb + `"}]}`},
		{substr: "/connections", body: `{"connections":[{"name":"` + parent + `/connections/ggc1"}]}`},
		{substr: "/connectionAssignments", body: `{"connectionAssignments":[{"name":"` + deployment + `/connectionAssignments/ca1"}]}`},
		{substr: "/deployments", body: `{"deployments":[{"name":"` + deployment + `"}]}`},
	}}
	service, err := oracledatabase.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	g := &OracleDatabaseGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	g.initResources(ctx, service, transport.client())
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_oracle_database_autonomous_database":              {parent + "/autonomousDatabases/adb1"},
		"google_oracle_database_cloud_exadata_infrastructure":     {parent + "/cloudExadataInfrastructures/exa1"},
		"google_oracle_database_cloud_vm_cluster":                 {parent + "/cloudVmClusters/vmc1"},
		"google_oracle_database_db_system":                        {parent + "/dbSystems/dbs1"},
		"google_oracle_database_exadb_vm_cluster":                 {parent + "/exadbVmClusters/evc1"},
		"google_oracle_database_exascale_db_storage_vault":        {parent + "/exascaleDbStorageVaults/vault1"},
		"google_oracle_database_odb_network":                      {odb},
		"google_oracle_database_odb_subnet":                       {odb + "/odbSubnets/sub1"},
		"google_oracle_database_goldengate_connection":            {parent + "/connections/ggc1"},
		"google_oracle_database_goldengate_deployment":            {deployment},
		"google_oracle_database_goldengate_connection_assignment": {deployment + "/connectionAssignments/ca1"},
	})
}
