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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	sstypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

// loadTailResources adds SSM activations, associations, documents, maintenance
// windows (with targets and tasks), patch baselines, patch groups and resource
// data syncs.
// Provider import IDs (verified): activation = ID; association = association ID;
// document = name; maintenance window = ID; window target = "<window ID>/<target ID>";
// window task = "<window ID>/<task ID>"; patch baseline = ID;
// patch group = "<patch group>,<baseline ID>"; resource data sync = name.
// Skipped in this family:
//   - aws_ssm_default_patch_baseline: a per-operating-system pointer to whichever
//     baseline is registered as default (RegisterDefaultPatchBaseline); it is a
//     derived settings mapping over the fixed OS enum, not an independent object.
//   - aws_ssm_service_setting: singleton account/region service settings object;
//     the provider ID is the setting ARN from a fixed, non-enumerable set of
//     setting IDs and the SSM API exposes only GetServiceSetting per ID.
func (g *SsmGenerator) loadTailResources(svc *ssm.Client) {
	ap := ssm.NewDescribeActivationsPaginator(svc, &ssm.DescribeActivationsInput{})
	for ap.HasMorePages() {
		page, err := ap.NextPage(context.TODO())
		if err != nil {
			break
		}
		for _, act := range page.ActivationList {
			id := StringValue(act.ActivationId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_ssm_activation", "aws", ssmAllowEmptyValues))
		}
	}

	assocP := ssm.NewListAssociationsPaginator(svc, &ssm.ListAssociationsInput{})
	for assocP.HasMorePages() {
		page, err := assocP.NextPage(context.TODO())
		if err != nil {
			break
		}
		for _, assoc := range page.Associations {
			id := StringValue(assoc.AssociationId)
			name := StringValue(assoc.Name)
			if name == "" {
				name = id
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, name, "aws_ssm_association", "aws", ssmAllowEmptyValues))
		}
	}

	dp := ssm.NewListDocumentsPaginator(svc, &ssm.ListDocumentsInput{
		DocumentFilterList: []sstypes.DocumentFilter{{Key: sstypes.DocumentFilterKeyOwner, Value: aws.String("Self")}},
	})
	for dp.HasMorePages() {
		page, err := dp.NextPage(context.TODO())
		if err != nil {
			break
		}
		for _, doc := range page.DocumentIdentifiers {
			name := StringValue(doc.Name)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_ssm_document", "aws", ssmAllowEmptyValues))
		}
	}

	wp := ssm.NewDescribeMaintenanceWindowsPaginator(svc, &ssm.DescribeMaintenanceWindowsInput{})
	for wp.HasMorePages() {
		page, err := wp.NextPage(context.TODO())
		if err != nil {
			break
		}
		for _, window := range page.WindowIdentities {
			windowID := StringValue(window.WindowId)
			name := StringValue(window.Name)
			if name == "" {
				name = windowID
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				windowID, name, "aws_ssm_maintenance_window", "aws", ssmAllowEmptyValues))

			targets, err := svc.DescribeMaintenanceWindowTargets(context.TODO(),
				&ssm.DescribeMaintenanceWindowTargetsInput{WindowId: window.WindowId})
			if err == nil {
				for _, target := range targets.Targets {
					id := windowID + "/" + StringValue(target.WindowTargetId)
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						id, id, "aws_ssm_maintenance_window_target", "aws", ssmAllowEmptyValues))
				}
			}
			tasks, err := svc.DescribeMaintenanceWindowTasks(context.TODO(),
				&ssm.DescribeMaintenanceWindowTasksInput{WindowId: window.WindowId})
			if err == nil {
				for _, task := range tasks.Tasks {
					id := windowID + "/" + StringValue(task.WindowTaskId)
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						id, id, "aws_ssm_maintenance_window_task", "aws", ssmAllowEmptyValues))
				}
			}
		}
	}

	bp := ssm.NewDescribePatchBaselinesPaginator(svc, &ssm.DescribePatchBaselinesInput{})
	for bp.HasMorePages() {
		page, err := bp.NextPage(context.TODO())
		if err != nil {
			break
		}
		for _, baseline := range page.BaselineIdentities {
			id := StringValue(baseline.BaselineId)
			name := StringValue(baseline.BaselineName)
			if name == "" {
				name = id
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, name, "aws_ssm_patch_baseline", "aws", ssmAllowEmptyValues))
		}
	}

	groups, err := svc.DescribePatchGroups(context.TODO(), &ssm.DescribePatchGroupsInput{})
	if err == nil {
		for _, mapping := range groups.Mappings {
			if mapping.BaselineIdentity == nil {
				continue
			}
			id := StringValue(mapping.PatchGroup) + "," + StringValue(mapping.BaselineIdentity.BaselineId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_ssm_patch_group", "aws", ssmAllowEmptyValues))
		}
	}

	syncs, err := svc.ListResourceDataSync(context.TODO(), &ssm.ListResourceDataSyncInput{})
	if err == nil {
		for _, sync := range syncs.ResourceDataSyncItems {
			name := StringValue(sync.SyncName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_ssm_resource_data_sync", "aws", ssmAllowEmptyValues))
		}
	}
}
