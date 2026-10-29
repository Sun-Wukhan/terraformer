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
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// EBS sweep additions to EbsGenerator. Importer IDs verified against
// provider docs: aws_ebs_snapshot = the snapshot ID;
// aws_ebs_fast_snapshot_restore = the Availability Zone name.
//
// Skipped in this family (see results ledger):
//   - aws_ebs_default_kms_key, aws_ebs_encryption_by_default and
//     aws_ebs_snapshot_block_public_access: region-level singleton
//     settings; importable by region/account in the provider but not
//     enumerable (no list call; exactly one implicit object per region) —
//     singleton precedent.
//   - aws_ebs_snapshot_copy, aws_ebs_snapshot_import and
//     aws_ebs_volume_copy: one-off copy/import operations whose results
//     are ordinary snapshots/volumes (already covered by aws_ebs_snapshot
//     / aws_ebs_volume); there is no API that enumerates resources *as*
//     copies/imports, so they cannot be discovered distinctly.

func (g *EbsGenerator) loadSweepResources(config aws.Config, svc *ec2.Client) error {
	ctx := context.TODO()
	acct, err := g.getAccountNumber(config)
	if err != nil {
		return err
	}
	snapP := ec2.NewDescribeSnapshotsPaginator(svc, &ec2.DescribeSnapshotsInput{OwnerIds: []string{StringValue(acct)}})
	for snapP.HasMorePages() {
		page, err := snapP.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, snap := range page.Snapshots {
			id := StringValue(snap.SnapshotId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(id, id, "aws_ebs_snapshot", "aws", ebsAllowEmptyValues))
		}
	}
	// Fast snapshot restore is managed per Availability Zone in the
	// provider (one resource per AZ), so dedupe the per-snapshot entries.
	seenAZ := map[string]bool{}
	fsrP := ec2.NewDescribeFastSnapshotRestoresPaginator(svc, &ec2.DescribeFastSnapshotRestoresInput{})
	for fsrP.HasMorePages() {
		page, err := fsrP.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, fsr := range page.FastSnapshotRestores {
			az := StringValue(fsr.AvailabilityZone)
			if string(fsr.State) == "enabled" && az != "" && !seenAZ[az] {
				seenAZ[az] = true
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(az, az, "aws_ebs_fast_snapshot_restore", "aws", ebsAllowEmptyValues))
			}
		}
	}
	return nil
}
