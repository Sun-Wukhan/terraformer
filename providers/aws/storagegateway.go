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
	"github.com/aws/aws-sdk-go-v2/service/storagegateway"
	"github.com/aws/aws-sdk-go-v2/service/storagegateway/types"
)

// Storage Gateway sweep. Importer IDs verified against provider docs:
// gateway and the per-gateway disk sub-resources (cache, upload buffer,
// working storage) import with the gateway ARN; cached/stored iSCSI
// volumes, NFS/SMB file shares and file system associations import with
// their own ARNs; tape pools import with the pool ARN.

var storageGatewayAllowEmptyValues = []string{"tags."}

type StorageGatewayGenerator struct {
	AWSService
}

func (g *StorageGatewayGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := storagegateway.NewFromConfig(config)
	ctx := context.TODO()
	add := func(id, name, typ string) {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(id, name, typ, "aws", storageGatewayAllowEmptyValues))
	}

	gwInput := &storagegateway.ListGatewaysInput{}
	for {
		out, err := svc.ListGateways(ctx, gwInput)
		if err != nil {
			return err
		}
		for _, gw := range out.Gateways {
			arn := StringValue(gw.GatewayARN)
			add(arn, StringValue(gw.GatewayName), "aws_storagegateway_gateway")
			// Per-gateway disk sub-resources exist only when the
			// corresponding Describe call succeeds.
			if _, err := svc.DescribeCache(ctx, &storagegateway.DescribeCacheInput{GatewayARN: gw.GatewayARN}); err == nil {
				add(arn, StringValue(gw.GatewayName), "aws_storagegateway_cache")
			}
			if _, err := svc.DescribeUploadBuffer(ctx, &storagegateway.DescribeUploadBufferInput{GatewayARN: gw.GatewayARN}); err == nil {
				add(arn, StringValue(gw.GatewayName), "aws_storagegateway_upload_buffer")
			}
			if _, err := svc.DescribeWorkingStorage(ctx, &storagegateway.DescribeWorkingStorageInput{GatewayARN: gw.GatewayARN}); err == nil {
				add(arn, StringValue(gw.GatewayName), "aws_storagegateway_working_storage")
			}
			// iSCSI volumes are enumerated via ListVolumes (the Describe*
			// calls take volume ARNs, not a gateway ARN); VolumeType
			// distinguishes cached from stored volumes.
			volInput := &storagegateway.ListVolumesInput{GatewayARN: gw.GatewayARN}
			for {
				vols, err := svc.ListVolumes(ctx, volInput)
				if err != nil {
					break
				}
				for _, vol := range vols.VolumeInfos {
					switch StringValue(vol.VolumeType) {
					case "CACHED":
						add(StringValue(vol.VolumeARN), StringValue(vol.VolumeId), "aws_storagegateway_cached_iscsi_volume")
					case "STORED":
						add(StringValue(vol.VolumeARN), StringValue(vol.VolumeId), "aws_storagegateway_stored_iscsi_volume")
					}
				}
				if vols.Marker == nil {
					break
				}
				volInput.Marker = vols.Marker
			}
		}
		if out.Marker == nil {
			break
		}
		gwInput.Marker = out.Marker
	}

	fsInput := &storagegateway.ListFileSharesInput{}
	for {
		out, err := svc.ListFileShares(ctx, fsInput)
		if err != nil {
			break
		}
		for _, share := range out.FileShareInfoList {
			switch share.FileShareType {
			case types.FileShareTypeNfs:
				add(StringValue(share.FileShareARN), StringValue(share.FileShareId), "aws_storagegateway_nfs_file_share")
			case types.FileShareTypeSmb:
				add(StringValue(share.FileShareARN), StringValue(share.FileShareId), "aws_storagegateway_smb_file_share")
			}
		}
		if out.NextMarker == nil {
			break
		}
		fsInput.Marker = out.NextMarker
	}

	fsaInput := &storagegateway.ListFileSystemAssociationsInput{}
	for {
		out, err := svc.ListFileSystemAssociations(ctx, fsaInput)
		if err != nil {
			break
		}
		for _, fsa := range out.FileSystemAssociationSummaryList {
			add(StringValue(fsa.FileSystemAssociationARN), StringValue(fsa.FileSystemAssociationARN), "aws_storagegateway_file_system_association")
		}
		if out.NextMarker == nil {
			break
		}
		fsaInput.Marker = out.NextMarker
	}

	poolInput := &storagegateway.ListTapePoolsInput{}
	for {
		out, err := svc.ListTapePools(ctx, poolInput)
		if err != nil {
			break
		}
		for _, pool := range out.PoolInfos {
			add(StringValue(pool.PoolARN), StringValue(pool.PoolName), "aws_storagegateway_tape_pool")
		}
		if out.Marker == nil {
			break
		}
		poolInput.Marker = out.Marker
	}
	return nil
}
