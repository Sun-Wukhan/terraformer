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
	"log"
	"net/http"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"google.golang.org/api/compute/v1"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
)

// OracleDatabaseGenerator enumerates Oracle Database at Google Cloud
// in the configured region:
//
//   - via the generated oracledatabase v1 client: autonomous
//     databases, cloud Exadata infrastructures and cloud VM clusters;
//   - over raw REST (the pinned generated client predates these
//     collections; endpoints verified against the Oracle Database
//     REST reference): DB systems, ExaDB VM clusters, Exascale DB
//     storage vaults, ODB networks and their ODB subnets;
//   - GoldenGate (goldengate.googleapis.com/v1, raw REST; no
//     generated client exists): connections, deployments, and the
//     connection assignments of each deployment.
//
// Deliberately not generated:
//   - google_oracle_database_cloud_exadata_infrastructure_exascale_config:
//     the exascale configuration is not a separate API object — it is
//     a section of the cloud Exadata infrastructure resource itself,
//     with no dedicated get or list endpoint in the discovery
//     document, so there is no independently enumerable object.
type OracleDatabaseGenerator struct {
	GCPService
}

const (
	oracleDatabaseBaseURL = "https://oracledatabase.googleapis.com/v1/"
	goldengateBaseURL     = "https://goldengate.googleapis.com/v1/"
)

func (g *OracleDatabaseGenerator) InitResources() error {
	ctx := context.Background()
	service, err := oracledatabase.NewService(ctx)
	if err != nil {
		return err
	}
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		log.Println(err)
	}
	g.initResources(ctx, service, restClient)
	return nil
}

func (g *OracleDatabaseGenerator) newResource(id, resourceType string) terraformutils.Resource {
	return terraformutils.NewResource(
		id, opsLastSegment(id), resourceType, g.ProviderName,
		map[string]string{}, []string{}, map[string]interface{}{},
	)
}

func (g *OracleDatabaseGenerator) initResources(ctx context.Context, service *oracledatabase.Service, restClient *http.Client) {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + region

	if err := service.Projects.Locations.AutonomousDatabases.List(parent).Pages(ctx, func(page *oracledatabase.ListAutonomousDatabasesResponse) error {
		for _, db := range page.AutonomousDatabases {
			g.Resources = append(g.Resources, g.newResource(db.Name, "google_oracle_database_autonomous_database"))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := service.Projects.Locations.CloudExadataInfrastructures.List(parent).Pages(ctx, func(page *oracledatabase.ListCloudExadataInfrastructuresResponse) error {
		for _, infra := range page.CloudExadataInfrastructures {
			g.Resources = append(g.Resources, g.newResource(infra.Name, "google_oracle_database_cloud_exadata_infrastructure"))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := service.Projects.Locations.CloudVmClusters.List(parent).Pages(ctx, func(page *oracledatabase.ListCloudVmClustersResponse) error {
		for _, cluster := range page.CloudVmClusters {
			g.Resources = append(g.Resources, g.newResource(cluster.Name, "google_oracle_database_cloud_vm_cluster"))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}

	if restClient == nil {
		return
	}
	restCollections := []struct {
		segment      string
		resourceType string
	}{
		{"dbSystems", "google_oracle_database_db_system"},
		{"exadbVmClusters", "google_oracle_database_exadb_vm_cluster"},
		{"exascaleDbStorageVaults", "google_oracle_database_exascale_db_storage_vault"},
		{"odbNetworks", "google_oracle_database_odb_network"},
	}
	for _, collection := range restCollections {
		items, err := restListResources(ctx, restClient, oracleDatabaseBaseURL+parent+"/"+collection.segment)
		if err != nil {
			log.Println(err)
			continue
		}
		for _, item := range items {
			g.Resources = append(g.Resources, g.newResource(item.Name, collection.resourceType))
			if collection.segment == "odbNetworks" {
				subnets, err := restListResources(ctx, restClient, oracleDatabaseBaseURL+item.Name+"/odbSubnets")
				if err != nil {
					log.Println(err)
					continue
				}
				for _, subnet := range subnets {
					g.Resources = append(g.Resources, g.newResource(subnet.Name, "google_oracle_database_odb_subnet"))
				}
			}
		}
	}
	g.loadGoldengate(ctx, restClient, parent)
}

func (g *OracleDatabaseGenerator) loadGoldengate(ctx context.Context, restClient *http.Client, parent string) {
	connections, err := restListResources(ctx, restClient, goldengateBaseURL+parent+"/connections")
	if err != nil {
		log.Println(err)
	} else {
		for _, conn := range connections {
			g.Resources = append(g.Resources, g.newResource(conn.Name, "google_oracle_database_goldengate_connection"))
		}
	}
	deployments, err := restListResources(ctx, restClient, goldengateBaseURL+parent+"/deployments")
	if err != nil {
		log.Println(err)
		return
	}
	for _, deployment := range deployments {
		g.Resources = append(g.Resources, g.newResource(deployment.Name, "google_oracle_database_goldengate_deployment"))
		assignments, err := restListResources(ctx, restClient, goldengateBaseURL+deployment.Name+"/connectionAssignments")
		if err != nil {
			log.Println(err)
			continue
		}
		for _, assignment := range assignments {
			g.Resources = append(g.Resources, g.newResource(assignment.Name, "google_oracle_database_goldengate_connection_assignment"))
		}
	}
}
