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
	"github.com/aws/aws-sdk-go-v2/service/efs"
	"github.com/aws/aws-sdk-go-v2/service/efs/types"
)

// EFS sweep additions to EfsGenerator. Both provider types are keyed by
// file system ID: aws_efs_backup_policy imports with the file system ID
// (emitted per file system whose backup policy exists — Describe always
// returns a policy object, so gate on its status being set), and
// aws_efs_replication_configuration is keyed by the source file system ID
// in the provider (its only identifying argument).

func (g *EfsGenerator) loadSweepResources(svc *efs.Client) error {
	ctx := context.TODO()
	fsP := efs.NewDescribeFileSystemsPaginator(svc, &efs.DescribeFileSystemsInput{})
	for fsP.HasMorePages() {
		page, err := fsP.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, fs := range page.FileSystems {
			if bp, err := svc.DescribeBackupPolicy(ctx, &efs.DescribeBackupPolicyInput{FileSystemId: fs.FileSystemId}); err == nil && bp.BackupPolicy != nil && bp.BackupPolicy.Status != "" {
				id := StringValue(fs.FileSystemId)
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(id, id, "aws_efs_backup_policy", "aws", efsAllowEmptyValues))
			}
		}
	}
	repP := efs.NewDescribeReplicationConfigurationsPaginator(svc, &efs.DescribeReplicationConfigurationsInput{})
	for repP.HasMorePages() {
		page, err := repP.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, rep := range page.Replications {
			// Only configurations whose destination is in this region are
			// managed here; cross-region destinations belong to their own
			// region's import.
			for _, dest := range rep.Destinations {
				if dest.Status == types.ReplicationStatusEnabled || dest.Status == types.ReplicationStatusEnabling {
					id := StringValue(rep.SourceFileSystemId)
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(id, id, "aws_efs_replication_configuration", "aws", efsAllowEmptyValues))
					break
				}
			}
		}
	}
	return nil
}
