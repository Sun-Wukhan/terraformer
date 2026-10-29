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
	"github.com/aws/aws-sdk-go-v2/service/datasync"
)

// DataSync sweep. Agents, tasks and every location type import with
// their ARN (verified against provider docs). ListLocations does not say
// which location type an ARN is, so each location is probed with the
// per-type Describe calls; exactly one succeeds for a given location and
// determines the emitted provider type.

var dataSyncAllowEmptyValues = []string{"tags."}

type DataSyncGenerator struct {
	AWSService
}

func (g *DataSyncGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := datasync.NewFromConfig(config)
	ctx := context.TODO()
	add := func(id, name, typ string) {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(id, name, typ, "aws", dataSyncAllowEmptyValues))
	}

	agentP := datasync.NewListAgentsPaginator(svc, &datasync.ListAgentsInput{})
	for agentP.HasMorePages() {
		page, err := agentP.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, agent := range page.Agents {
			add(StringValue(agent.AgentArn), StringValue(agent.Name), "aws_datasync_agent")
		}
	}
	taskP := datasync.NewListTasksPaginator(svc, &datasync.ListTasksInput{})
	for taskP.HasMorePages() {
		page, err := taskP.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, task := range page.Tasks {
			add(StringValue(task.TaskArn), StringValue(task.Name), "aws_datasync_task")
		}
	}
	locP := datasync.NewListLocationsPaginator(svc, &datasync.ListLocationsInput{})
	for locP.HasMorePages() {
		page, err := locP.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, loc := range page.Locations {
			arn := StringValue(loc.LocationArn)
			switch {
			case datasyncProbe(func() error {
				_, err := svc.DescribeLocationS3(ctx, &datasync.DescribeLocationS3Input{LocationArn: loc.LocationArn})
				return err
			}):
				add(arn, arn, "aws_datasync_location_s3")
			case datasyncProbe(func() error {
				_, err := svc.DescribeLocationNfs(ctx, &datasync.DescribeLocationNfsInput{LocationArn: loc.LocationArn})
				return err
			}):
				add(arn, arn, "aws_datasync_location_nfs")
			case datasyncProbe(func() error {
				_, err := svc.DescribeLocationSmb(ctx, &datasync.DescribeLocationSmbInput{LocationArn: loc.LocationArn})
				return err
			}):
				add(arn, arn, "aws_datasync_location_smb")
			case datasyncProbe(func() error {
				_, err := svc.DescribeLocationEfs(ctx, &datasync.DescribeLocationEfsInput{LocationArn: loc.LocationArn})
				return err
			}):
				add(arn, arn, "aws_datasync_location_efs")
			case datasyncProbe(func() error {
				_, err := svc.DescribeLocationFsxLustre(ctx, &datasync.DescribeLocationFsxLustreInput{LocationArn: loc.LocationArn})
				return err
			}):
				add(arn, arn, "aws_datasync_location_fsx_lustre_file_system")
			case datasyncProbe(func() error {
				_, err := svc.DescribeLocationFsxOntap(ctx, &datasync.DescribeLocationFsxOntapInput{LocationArn: loc.LocationArn})
				return err
			}):
				add(arn, arn, "aws_datasync_location_fsx_ontap_file_system")
			case datasyncProbe(func() error {
				_, err := svc.DescribeLocationFsxOpenZfs(ctx, &datasync.DescribeLocationFsxOpenZfsInput{LocationArn: loc.LocationArn})
				return err
			}):
				add(arn, arn, "aws_datasync_location_fsx_openzfs_file_system")
			case datasyncProbe(func() error {
				_, err := svc.DescribeLocationFsxWindows(ctx, &datasync.DescribeLocationFsxWindowsInput{LocationArn: loc.LocationArn})
				return err
			}):
				add(arn, arn, "aws_datasync_location_fsx_windows_file_system")
			case datasyncProbe(func() error {
				_, err := svc.DescribeLocationHdfs(ctx, &datasync.DescribeLocationHdfsInput{LocationArn: loc.LocationArn})
				return err
			}):
				add(arn, arn, "aws_datasync_location_hdfs")
			case datasyncProbe(func() error {
				_, err := svc.DescribeLocationObjectStorage(ctx, &datasync.DescribeLocationObjectStorageInput{LocationArn: loc.LocationArn})
				return err
			}):
				add(arn, arn, "aws_datasync_location_object_storage")
			case datasyncProbe(func() error {
				_, err := svc.DescribeLocationAzureBlob(ctx, &datasync.DescribeLocationAzureBlobInput{LocationArn: loc.LocationArn})
				return err
			}):
				add(arn, arn, "aws_datasync_location_azure_blob")
			}
		}
	}
	return nil
}

func datasyncProbe(f func() error) bool { return f() == nil }
