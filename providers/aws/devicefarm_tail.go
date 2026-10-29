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
	"github.com/aws/aws-sdk-go-v2/service/devicefarm"
	dftypes "github.com/aws/aws-sdk-go-v2/service/devicefarm/types"
)

// loadTailResources adds Device Farm device pools, instance profiles, network
// profiles, test grid projects and uploads.
// Provider import IDs (verified): each resource's ARN.
func (g *DeviceFarmGenerator) loadTailResources(svc *devicefarm.Client) {
	profiles, err := svc.ListInstanceProfiles(context.TODO(), &devicefarm.ListInstanceProfilesInput{})
	if err == nil {
		for _, prof := range profiles.InstanceProfiles {
			arn := StringValue(prof.Arn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(prof.Name), "aws_devicefarm_instance_profile", "aws", devicefarmAllowEmptyValues))
		}
	}

	tgp := devicefarm.NewListTestGridProjectsPaginator(svc, &devicefarm.ListTestGridProjectsInput{})
	for tgp.HasMorePages() {
		page, err := tgp.NextPage(context.TODO())
		if err != nil {
			break
		}
		for _, proj := range page.TestGridProjects {
			arn := StringValue(proj.Arn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(proj.Name), "aws_devicefarm_test_grid_project", "aws", devicefarmAllowEmptyValues))
		}
	}

	pp := devicefarm.NewListProjectsPaginator(svc, &devicefarm.ListProjectsInput{})
	for pp.HasMorePages() {
		page, err := pp.NextPage(context.TODO())
		if err != nil {
			break
		}
		for _, project := range page.Projects {
			pools, err := svc.ListDevicePools(context.TODO(), &devicefarm.ListDevicePoolsInput{
				Arn: project.Arn, Type: dftypes.DevicePoolTypePrivate,
			})
			if err == nil {
				for _, pool := range pools.DevicePools {
					arn := StringValue(pool.Arn)
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						arn, StringValue(pool.Name), "aws_devicefarm_device_pool", "aws", devicefarmAllowEmptyValues))
				}
			}
			networks, err := svc.ListNetworkProfiles(context.TODO(), &devicefarm.ListNetworkProfilesInput{
				Arn: project.Arn, Type: dftypes.NetworkProfileTypePrivate,
			})
			if err == nil {
				for _, prof := range networks.NetworkProfiles {
					arn := StringValue(prof.Arn)
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						arn, StringValue(prof.Name), "aws_devicefarm_network_profile", "aws", devicefarmAllowEmptyValues))
				}
			}
			uploads, err := svc.ListUploads(context.TODO(), &devicefarm.ListUploadsInput{Arn: project.Arn})
			if err == nil {
				for _, up := range uploads.Uploads {
					arn := StringValue(up.Arn)
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						arn, StringValue(up.Name), "aws_devicefarm_upload", "aws", devicefarmAllowEmptyValues))
				}
			}
			_ = aws.String
		}
	}
}
