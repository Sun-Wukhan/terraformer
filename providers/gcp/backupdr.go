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
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	backupdr "google.golang.org/api/backupdr/v1"
	"google.golang.org/api/compute/v1"
)

var backupDrAllowEmptyValues = []string{""}

var backupDrAdditionalFields = map[string]interface{}{}

type BackupDrGenerator struct {
	GCPService
}

// InitResources enumerates Backup and DR in the configured region:
// backup vaults, backup plans, backup plan associations and
// management servers. Provider IDs are the full resource names.
//
// Deliberately not generated:
//   - google_backup_dr_service_config: the BackupDR API exposes only
//     an initialize method on the per-location serviceConfig singleton
//     (no get or list in the generated client or the discovery
//     document), so an existing configuration cannot be discovered or
//     read back.
//   - google_backup_dr_auto_protection_policy: the BackupDR API has
//     no autoProtectionPolicies collection in the generated client or
//     the discovery document (no list/get surface at all).
//   - google_backup_dr_restore_workload: a restore is a transient
//     action on a backup (the backups :restore method); the API has
//     no restoreWorkloads collection to enumerate.
func (g *BackupDrGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := backupdr.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, svc)
}

func (g *BackupDrGenerator) initResources(ctx context.Context, svc *backupdr.Service) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + region
	if err := svc.Projects.Locations.BackupVaults.List(parent).Pages(ctx, func(page *backupdr.ListBackupVaultsResponse) error {
		for _, v := range page.BackupVaults {
			g.Resources = append(g.Resources, g.newResource(v.Name, "google_backup_dr_backup_vault"))
		}
		return nil
	}); err != nil {
		return err
	}
	if err := svc.Projects.Locations.BackupPlans.List(parent).Pages(ctx, func(page *backupdr.ListBackupPlansResponse) error {
		for _, p := range page.BackupPlans {
			g.Resources = append(g.Resources, g.newResource(p.Name, "google_backup_dr_backup_plan"))
		}
		return nil
	}); err != nil {
		return err
	}
	if err := svc.Projects.Locations.BackupPlanAssociations.List(parent).Pages(ctx, func(page *backupdr.ListBackupPlanAssociationsResponse) error {
		for _, a := range page.BackupPlanAssociations {
			g.Resources = append(g.Resources, g.newResource(a.Name, "google_backup_dr_backup_plan_association"))
		}
		return nil
	}); err != nil {
		return err
	}
	return svc.Projects.Locations.ManagementServers.List(parent).Pages(ctx, func(page *backupdr.ListManagementServersResponse) error {
		for _, s := range page.ManagementServers {
			g.Resources = append(g.Resources, g.newResource(s.Name, "google_backup_dr_management_server"))
		}
		return nil
	})
}

func (g *BackupDrGenerator) newResource(id, resourceType string) terraformutils.Resource {
	parts := strings.Split(id, "/")
	return terraformutils.NewResource(id, parts[len(parts)-1], resourceType, g.ProviderName,
		map[string]string{}, backupDrAllowEmptyValues, backupDrAdditionalFields)
}
