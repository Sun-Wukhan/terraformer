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

package aws

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/backup"
)

// AWS Backup sweep. Importer IDs verified against provider docs:
// plan = plan ID; selection = "plan-id|selection-id"; vault, vault lock
// configuration, vault notifications and vault policy = vault name;
// framework / report plan = their names; restore testing plan = its name;
// restore testing selection = "selection-name:plan-name".
// Logically air-gapped vaults are the ListBackupVaults entries whose
// VaultType is LOGICALLY_AIR_GAPPED_BACKUP_VAULT; vault lock configurations are the
// vaults reporting Locked / retention bounds in the same list.
//
// Skipped in this family (see results ledger):
//   - aws_backup_global_settings: account-level singleton (Describe exists
//     but there is exactly one implicit object; not enumerable) —
//     singleton precedent.
//   - aws_backup_region_settings: region-level singleton, same reasoning.

var backupAllowEmptyValues = []string{"tags."}

type BackupGenerator struct {
	AWSService
}

func (g *BackupGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := backup.NewFromConfig(config)
	ctx := context.TODO()
	add := func(id, name, typ string) {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(id, name, typ, "aws", backupAllowEmptyValues))
	}

	planP := backup.NewListBackupPlansPaginator(svc, &backup.ListBackupPlansInput{})
	for planP.HasMorePages() {
		page, err := planP.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, plan := range page.BackupPlansList {
			planID := StringValue(plan.BackupPlanId)
			add(planID, StringValue(plan.BackupPlanName), "aws_backup_plan")
			selInput := &backup.ListBackupSelectionsInput{BackupPlanId: plan.BackupPlanId}
			for {
				selOut, err := svc.ListBackupSelections(ctx, selInput)
				if err != nil {
					break
				}
				for _, sel := range selOut.BackupSelectionsList {
					id := planID + "|" + StringValue(sel.SelectionId)
					add(id, StringValue(sel.SelectionName), "aws_backup_selection")
				}
				if selOut.NextToken == nil {
					break
				}
				selInput.NextToken = selOut.NextToken
			}
		}
	}

	vaultInput := &backup.ListBackupVaultsInput{}
	for {
		out, err := svc.ListBackupVaults(ctx, vaultInput)
		if err != nil {
			return err
		}
		for _, vault := range out.BackupVaultList {
			name := StringValue(vault.BackupVaultName)
			if string(vault.VaultType) == "LOGICALLY_AIR_GAPPED_BACKUP_VAULT" {
				add(name, name, "aws_backup_logically_air_gapped_vault")
			} else {
				add(name, name, "aws_backup_vault")
			}
			if (vault.Locked != nil && *vault.Locked) || vault.MinRetentionDays != nil {
				add(name, name, "aws_backup_vault_lock_configuration")
			}
			if _, err := svc.GetBackupVaultNotifications(ctx, &backup.GetBackupVaultNotificationsInput{BackupVaultName: vault.BackupVaultName}); err == nil {
				add(name, name, "aws_backup_vault_notifications")
			}
			if _, err := svc.GetBackupVaultAccessPolicy(ctx, &backup.GetBackupVaultAccessPolicyInput{BackupVaultName: vault.BackupVaultName}); err == nil {
				add(name, name, "aws_backup_vault_policy")
			}
		}
		if out.NextToken == nil {
			break
		}
		vaultInput.NextToken = out.NextToken
	}

	fwInput := &backup.ListFrameworksInput{}
	for {
		out, err := svc.ListFrameworks(ctx, fwInput)
		if err != nil {
			break
		}
		for _, fw := range out.Frameworks {
			add(StringValue(fw.FrameworkName), StringValue(fw.FrameworkName), "aws_backup_framework")
		}
		if out.NextToken == nil {
			break
		}
		fwInput.NextToken = out.NextToken
	}

	rpInput := &backup.ListReportPlansInput{}
	for {
		out, err := svc.ListReportPlans(ctx, rpInput)
		if err != nil {
			break
		}
		for _, rp := range out.ReportPlans {
			add(StringValue(rp.ReportPlanName), StringValue(rp.ReportPlanName), "aws_backup_report_plan")
		}
		if out.NextToken == nil {
			break
		}
		rpInput.NextToken = out.NextToken
	}

	rtInput := &backup.ListRestoreTestingPlansInput{}
	for {
		out, err := svc.ListRestoreTestingPlans(ctx, rtInput)
		if err != nil {
			break
		}
		for _, plan := range out.RestoreTestingPlans {
			planName := StringValue(plan.RestoreTestingPlanName)
			add(planName, planName, "aws_backup_restore_testing_plan")
			selInput := &backup.ListRestoreTestingSelectionsInput{RestoreTestingPlanName: plan.RestoreTestingPlanName}
			for {
				selOut, err := svc.ListRestoreTestingSelections(ctx, selInput)
				if err != nil {
					break
				}
				for _, sel := range selOut.RestoreTestingSelections {
					id := StringValue(sel.RestoreTestingSelectionName) + ":" + planName
					add(id, StringValue(sel.RestoreTestingSelectionName), "aws_backup_restore_testing_selection")
				}
				if selOut.NextToken == nil {
					break
				}
				selInput.NextToken = selOut.NextToken
			}
		}
		if out.NextToken == nil {
			break
		}
		rtInput.NextToken = out.NextToken
	}
	return nil
}
