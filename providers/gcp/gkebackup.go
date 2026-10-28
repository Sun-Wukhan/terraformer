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
	gkebackupv1 "google.golang.org/api/gkebackup/v1"
)

var gkeBackupAllowEmptyValues = []string{""}

var gkeBackupAdditionalFields = map[string]interface{}{}

// GkeBackupGenerator enumerates Backup for GKE in the configured
// region: backup channels, backup plans (with IAM), restore
// channels and restore plans (with IAM).
type GkeBackupGenerator struct {
	GCPService
}

func (g *GkeBackupGenerator) InitResources() error {
	ctx := context.Background()
	backupService, err := gkebackupv1.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, backupService)
}

func (g *GkeBackupGenerator) initResources(ctx context.Context, backupService *gkebackupv1.Service) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + region
	baseAttrs := map[string]string{"project": project, "location": region}

	err := backupService.Projects.Locations.BackupChannels.List(parent).Pages(ctx, func(page *gkebackupv1.ListBackupChannelsResponse) error {
		for _, channel := range page.BackupChannels {
			g.Resources = append(g.Resources, g.newGkeBackupResource(channel.Name, "google_gke_backup_backup_channel", baseAttrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = backupService.Projects.Locations.RestoreChannels.List(parent).Pages(ctx, func(page *gkebackupv1.ListRestoreChannelsResponse) error {
		for _, channel := range page.RestoreChannels {
			g.Resources = append(g.Resources, g.newGkeBackupResource(channel.Name, "google_gke_backup_restore_channel", baseAttrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = backupService.Projects.Locations.BackupPlans.List(parent).Pages(ctx, func(page *gkebackupv1.ListBackupPlansResponse) error {
		for _, plan := range page.BackupPlans {
			g.Resources = append(g.Resources, g.newGkeBackupResource(plan.Name, "google_gke_backup_backup_plan", baseAttrs))
			g.loadGkeBackupIam(ctx, backupService, plan.Name, true, baseAttrs)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return backupService.Projects.Locations.RestorePlans.List(parent).Pages(ctx, func(page *gkebackupv1.ListRestorePlansResponse) error {
		for _, plan := range page.RestorePlans {
			g.Resources = append(g.Resources, g.newGkeBackupResource(plan.Name, "google_gke_backup_restore_plan", baseAttrs))
			g.loadGkeBackupIam(ctx, backupService, plan.Name, false, baseAttrs)
		}
		return nil
	})
}

func (g *GkeBackupGenerator) newGkeBackupResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
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
		gkeBackupAllowEmptyValues,
		gkeBackupAdditionalFields,
	)
}

func (g *GkeBackupGenerator) loadGkeBackupIam(ctx context.Context, backupService *gkebackupv1.Service, name string, backupPlan bool, baseAttrs map[string]string) {
	var policy *gkebackupv1.Policy
	var err error
	if backupPlan {
		policy, err = backupService.Projects.Locations.BackupPlans.GetIamPolicy(name).Context(ctx).Do()
	} else {
		policy, err = backupService.Projects.Locations.RestorePlans.GetIamPolicy(name).Context(ctx).Do()
	}
	if err != nil {
		log.Println(err)
		return
	}
	bindings := []iamBinding{}
	for _, binding := range policy.Bindings {
		bindings = append(bindings, iamBinding{Role: binding.Role, Members: binding.Members})
	}
	if backupPlan {
		parentAttrs := map[string]string{"backup_plan_id": name}
		for k, v := range baseAttrs {
			parentAttrs[k] = v
		}
		g.Resources = append(g.Resources, expandIamPolicy(
			bindings, name,
			"google_gke_backup_backup_plan_iam_policy", "google_gke_backup_backup_plan_iam_binding", "google_gke_backup_backup_plan_iam_member",
			parentAttrs, g.newGkeBackupResource)...)
		return
	}
	parentAttrs := map[string]string{"restore_plan_id": name}
	for k, v := range baseAttrs {
		parentAttrs[k] = v
	}
	g.Resources = append(g.Resources, expandIamPolicy(
		bindings, name,
		"google_gke_backup_restore_plan_iam_policy", "google_gke_backup_restore_plan_iam_binding", "google_gke_backup_restore_plan_iam_member",
		parentAttrs, g.newGkeBackupResource)...)
}
