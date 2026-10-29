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
	"github.com/aws/aws-sdk-go-v2/service/cloud9"
)

// loadEnvironmentMemberships adds aws_cloud9_environment_membership resources.
// Provider import ID (verified): "<environment ID>#<user ARN>".
func (g *Cloud9Generator) loadEnvironmentMemberships(svc *cloud9.Client) {
	output, err := svc.ListEnvironments(context.TODO(), &cloud9.ListEnvironmentsInput{})
	if err != nil {
		return
	}
	for _, environmentID := range output.EnvironmentIds {
		members, err := svc.DescribeEnvironmentMemberships(context.TODO(),
			&cloud9.DescribeEnvironmentMembershipsInput{EnvironmentId: &environmentID})
		if err != nil {
			continue
		}
		for _, m := range members.Memberships {
			id := environmentID + "#" + StringValue(m.UserArn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_cloud9_environment_membership", "aws", cloud9AllowEmptyValues))
		}
	}
}
