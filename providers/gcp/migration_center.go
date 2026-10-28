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
	migrationcenterv1 "google.golang.org/api/migrationcenter/v1"
)

var migrationCenterAllowEmptyValues = []string{""}

var migrationCenterAdditionalFields = map[string]interface{}{}

type MigrationCenterGenerator struct {
	GCPService
}

// InitResources enumerates Migration Center resources for the
// configured project and region: groups, preference sets, sources,
// discovery clients, import jobs with their import data files,
// report configs with their reports, and the per-location settings
// singleton (read through locations.getSettings; its provider ID is
// projects/{project}/locations/{location}/settings). Provider IDs
// are the full API resource names for the rest.
//
// Deliberately not generated:
//   - google_migration_center_assets_export_job: the Migration
//     Center v1 API in the pinned generated client and its discovery
//     document has no assetsExportJobs collection (no list/get), so
//     there is no verified endpoint to enumerate export jobs
//     through.
func (g *MigrationCenterGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := migrationcenterv1.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, svc)
}

func (g *MigrationCenterGenerator) initResources(ctx context.Context, svc *migrationcenterv1.Service) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region)
	locationParent := "projects/" + project + "/locations/" + region.Name
	baseAttrs := map[string]string{"project": project, "location": region.Name}

	g.loadMigrationCenterList(func() error {
		return svc.Projects.Locations.Groups.List(locationParent).Pages(ctx, func(page *migrationcenterv1.ListGroupsResponse) error {
			for _, group := range page.Groups {
				g.Resources = append(g.Resources, g.newMigrationCenterResource(group.Name, "google_migration_center_group", baseAttrs))
			}
			return nil
		})
	})
	g.loadMigrationCenterList(func() error {
		return svc.Projects.Locations.PreferenceSets.List(locationParent).Pages(ctx, func(page *migrationcenterv1.ListPreferenceSetsResponse) error {
			for _, preferenceSet := range page.PreferenceSets {
				g.Resources = append(g.Resources, g.newMigrationCenterResource(preferenceSet.Name, "google_migration_center_preference_set", baseAttrs))
			}
			return nil
		})
	})
	g.loadMigrationCenterList(func() error {
		return svc.Projects.Locations.Sources.List(locationParent).Pages(ctx, func(page *migrationcenterv1.ListSourcesResponse) error {
			for _, source := range page.Sources {
				g.Resources = append(g.Resources, g.newMigrationCenterResource(source.Name, "google_migration_center_source", baseAttrs))
			}
			return nil
		})
	})
	g.loadMigrationCenterList(func() error {
		return svc.Projects.Locations.DiscoveryClients.List(locationParent).Pages(ctx, func(page *migrationcenterv1.ListDiscoveryClientsResponse) error {
			for _, client := range page.DiscoveryClients {
				g.Resources = append(g.Resources, g.newMigrationCenterResource(client.Name, "google_migration_center_discovery_client", baseAttrs))
			}
			return nil
		})
	})
	g.loadImportJobs(ctx, svc, locationParent, baseAttrs)
	g.loadReportConfigs(ctx, svc, locationParent, baseAttrs)

	if _, err := svc.Projects.Locations.GetSettings(locationParent).Context(ctx).Do(); err != nil {
		log.Println(err)
	} else {
		g.Resources = append(g.Resources, g.newMigrationCenterResource(locationParent+"/settings", "google_migration_center_settings", baseAttrs))
	}
	return nil
}

func (g *MigrationCenterGenerator) loadMigrationCenterList(listFn func() error) {
	if err := listFn(); err != nil {
		log.Println(err)
	}
}

func (g *MigrationCenterGenerator) newMigrationCenterResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
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
		migrationCenterAllowEmptyValues,
		migrationCenterAdditionalFields,
	)
}

func (g *MigrationCenterGenerator) loadImportJobs(ctx context.Context, svc *migrationcenterv1.Service, locationParent string, baseAttrs map[string]string) {
	err := svc.Projects.Locations.ImportJobs.List(locationParent).Pages(ctx, func(page *migrationcenterv1.ListImportJobsResponse) error {
		for _, job := range page.ImportJobs {
			g.Resources = append(g.Resources, g.newMigrationCenterResource(job.Name, "google_migration_center_import_job", baseAttrs))
			fileErr := svc.Projects.Locations.ImportJobs.ImportDataFiles.List(job.Name).Pages(ctx, func(page *migrationcenterv1.ListImportDataFilesResponse) error {
				for _, file := range page.ImportDataFiles {
					g.Resources = append(g.Resources, g.newMigrationCenterResource(file.Name, "google_migration_center_import_data_file", baseAttrs))
				}
				return nil
			})
			if fileErr != nil {
				log.Println(fileErr)
			}
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (g *MigrationCenterGenerator) loadReportConfigs(ctx context.Context, svc *migrationcenterv1.Service, locationParent string, baseAttrs map[string]string) {
	err := svc.Projects.Locations.ReportConfigs.List(locationParent).Pages(ctx, func(page *migrationcenterv1.ListReportConfigsResponse) error {
		for _, config := range page.ReportConfigs {
			g.Resources = append(g.Resources, g.newMigrationCenterResource(config.Name, "google_migration_center_report_config", baseAttrs))
			reportErr := svc.Projects.Locations.ReportConfigs.Reports.List(config.Name).Pages(ctx, func(page *migrationcenterv1.ListReportsResponse) error {
				for _, report := range page.Reports {
					g.Resources = append(g.Resources, g.newMigrationCenterResource(report.Name, "google_migration_center_report", baseAttrs))
				}
				return nil
			})
			if reportErr != nil {
				log.Println(reportErr)
			}
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}
