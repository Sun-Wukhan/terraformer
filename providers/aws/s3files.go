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
	"github.com/aws/aws-sdk-go-v2/service/s3files"
)

// S3 Files sweep. Importer IDs verified against provider docs:
// file system = its fs-* ID; synchronization configuration = the file
// system ID; mount target = its mount target ID; access point = its
// access point ID; file system policy = the file system ID (the policy
// resource is keyed by file_system_id in the provider).

var s3FilesAllowEmptyValues = []string{"tags."}

type S3FilesGenerator struct {
	AWSService
}

func (g *S3FilesGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := s3files.NewFromConfig(config)
	ctx := context.TODO()
	add := func(id, name, typ string) {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(id, name, typ, "aws", s3FilesAllowEmptyValues))
	}
	fsInput := &s3files.ListFileSystemsInput{}
	for {
		out, err := svc.ListFileSystems(ctx, fsInput)
		if err != nil {
			return err
		}
		for _, fs := range out.FileSystems {
			fsID := StringValue(fs.FileSystemId)
			add(fsID, fsID, "aws_s3files_file_system")
			if _, err := svc.GetFileSystemPolicy(ctx, &s3files.GetFileSystemPolicyInput{FileSystemId: fs.FileSystemId}); err == nil {
				add(fsID, fsID, "aws_s3files_file_system_policy")
			}
			if _, err := svc.GetSynchronizationConfiguration(ctx, &s3files.GetSynchronizationConfigurationInput{FileSystemId: fs.FileSystemId}); err == nil {
				add(fsID, fsID, "aws_s3files_synchronization_configuration")
			}
			mtInput := &s3files.ListMountTargetsInput{FileSystemId: fs.FileSystemId}
			for {
				mtOut, err := svc.ListMountTargets(ctx, mtInput)
				if err != nil {
					break
				}
				for _, mt := range mtOut.MountTargets {
					add(StringValue(mt.MountTargetId), StringValue(mt.MountTargetId), "aws_s3files_mount_target")
				}
				if mtOut.NextToken == nil {
					break
				}
				mtInput.NextToken = mtOut.NextToken
			}
			apInput := &s3files.ListAccessPointsInput{FileSystemId: fs.FileSystemId}
			for {
				apOut, err := svc.ListAccessPoints(ctx, apInput)
				if err != nil {
					break
				}
				for _, ap := range apOut.AccessPoints {
					add(StringValue(ap.AccessPointId), StringValue(ap.AccessPointId), "aws_s3files_access_point")
				}
				if apOut.NextToken == nil {
					break
				}
				apInput.NextToken = apOut.NextToken
			}
		}
		if out.NextToken == nil {
			break
		}
		fsInput.NextToken = out.NextToken
	}
	return nil
}
