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
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"google.golang.org/api/compute/v1"
	datamigrationv1 "google.golang.org/api/datamigration/v1"
)

var dmsAllowEmptyValues = []string{""}

var dmsAdditionalFields = map[string]interface{}{}

type DatabaseMigrationServiceGenerator struct {
	GCPService
}

// InitResources enumerates Database Migration Service resources for
// the configured project and region: connection profiles, migration
// jobs and private connections. Provider IDs are the full API
// resource names, the canonical import format for all of these.
func (g *DatabaseMigrationServiceGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := datamigrationv1.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, svc)
}

func (g *DatabaseMigrationServiceGenerator) initResources(ctx context.Context, svc *datamigrationv1.Service) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region)
	locationParent := "projects/" + project + "/locations/" + region.Name
	baseAttrs := map[string]string{"project": project, "location": region.Name}
	err := svc.Projects.Locations.ConnectionProfiles.List(locationParent).Pages(ctx, func(page *datamigrationv1.ListConnectionProfilesResponse) error {
		for _, profile := range page.ConnectionProfiles {
			g.Resources = append(g.Resources, g.newDMSResource(profile.Name, "google_database_migration_service_connection_profile", baseAttrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = svc.Projects.Locations.MigrationJobs.List(locationParent).Pages(ctx, func(page *datamigrationv1.ListMigrationJobsResponse) error {
		for _, job := range page.MigrationJobs {
			g.Resources = append(g.Resources, g.newDMSResource(job.Name, "google_database_migration_service_migration_job", baseAttrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = svc.Projects.Locations.PrivateConnections.List(locationParent).Pages(ctx, func(page *datamigrationv1.ListPrivateConnectionsResponse) error {
		for _, connection := range page.PrivateConnections {
			g.Resources = append(g.Resources, g.newDMSResource(connection.Name, "google_database_migration_service_private_connection", baseAttrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	return nil
}

func (g *DatabaseMigrationServiceGenerator) newDMSResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
	if attrs == nil {
		attrs = map[string]string{}
	}
	parts := strings.Split(id, "/")
	return terraformutils.NewResource(
		id,
		parts[len(parts)-1],
		resourceType,
		g.ProviderName,
		attrs,
		dmsAllowEmptyValues,
		dmsAdditionalFields,
	)
}
