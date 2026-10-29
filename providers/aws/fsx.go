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
	"github.com/aws/aws-sdk-go-v2/service/fsx"
	"github.com/aws/aws-sdk-go-v2/service/fsx/types"
)

// FSx sweep. One DescribeFileSystems enumeration is split by
// FileSystemType into the per-engine provider types (all import with the
// fs-* file system ID, verified against provider docs); volumes are split
// by VolumeType (ONTAP fsvol-* IDs / OPENZFS IDs), snapshots are the
// OpenZFS snapshot type (FSx snapshots only exist for OpenZFS), and the
// S3 access point attachment imports with its name.

var fsxAllowEmptyValues = []string{"tags."}

type FsxGenerator struct {
	AWSService
}

func (g *FsxGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := fsx.NewFromConfig(config)
	ctx := context.TODO()
	add := func(id, name, typ string) {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(id, name, typ, "aws", fsxAllowEmptyValues))
	}

	fsP := fsx.NewDescribeFileSystemsPaginator(svc, &fsx.DescribeFileSystemsInput{})
	for fsP.HasMorePages() {
		page, err := fsP.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, fs := range page.FileSystems {
			id := StringValue(fs.FileSystemId)
			switch fs.FileSystemType {
			case types.FileSystemTypeLustre:
				add(id, id, "aws_fsx_lustre_file_system")
			case types.FileSystemTypeWindows:
				add(id, id, "aws_fsx_windows_file_system")
			case types.FileSystemTypeOntap:
				add(id, id, "aws_fsx_ontap_file_system")
			case types.FileSystemTypeOpenzfs:
				add(id, id, "aws_fsx_openzfs_file_system")
			}
		}
	}
	volP := fsx.NewDescribeVolumesPaginator(svc, &fsx.DescribeVolumesInput{})
	for volP.HasMorePages() {
		page, err := volP.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, vol := range page.Volumes {
			id := StringValue(vol.VolumeId)
			switch vol.VolumeType {
			case types.VolumeTypeOntap:
				add(id, id, "aws_fsx_ontap_volume")
			case types.VolumeTypeOpenzfs:
				add(id, id, "aws_fsx_openzfs_volume")
			}
		}
	}
	svmP := fsx.NewDescribeStorageVirtualMachinesPaginator(svc, &fsx.DescribeStorageVirtualMachinesInput{})
	for svmP.HasMorePages() {
		page, err := svmP.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, svm := range page.StorageVirtualMachines {
			id := StringValue(svm.StorageVirtualMachineId)
			add(id, id, "aws_fsx_ontap_storage_virtual_machine")
		}
	}
	snapP := fsx.NewDescribeSnapshotsPaginator(svc, &fsx.DescribeSnapshotsInput{})
	for snapP.HasMorePages() {
		page, err := snapP.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, snap := range page.Snapshots {
			id := StringValue(snap.SnapshotId)
			add(id, id, "aws_fsx_openzfs_snapshot")
		}
	}
	bakP := fsx.NewDescribeBackupsPaginator(svc, &fsx.DescribeBackupsInput{})
	for bakP.HasMorePages() {
		page, err := bakP.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, bak := range page.Backups {
			id := StringValue(bak.BackupId)
			add(id, id, "aws_fsx_backup")
		}
	}
	draP := fsx.NewDescribeDataRepositoryAssociationsPaginator(svc, &fsx.DescribeDataRepositoryAssociationsInput{})
	for draP.HasMorePages() {
		page, err := draP.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, dra := range page.Associations {
			id := StringValue(dra.AssociationId)
			add(id, id, "aws_fsx_data_repository_association")
		}
	}
	fcP := fsx.NewDescribeFileCachesPaginator(svc, &fsx.DescribeFileCachesInput{})
	for fcP.HasMorePages() {
		page, err := fcP.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, fc := range page.FileCaches {
			id := StringValue(fc.FileCacheId)
			add(id, id, "aws_fsx_file_cache")
		}
	}
	attInput := &fsx.DescribeS3AccessPointAttachmentsInput{}
	for {
		out, err := svc.DescribeS3AccessPointAttachments(ctx, attInput)
		if err != nil {
			break
		}
		for _, att := range out.S3AccessPointAttachments {
			add(StringValue(att.Name), StringValue(att.Name), "aws_fsx_s3_access_point_attachment")
		}
		if out.NextToken == nil {
			break
		}
		attInput.NextToken = out.NextToken
	}
	return nil
}
