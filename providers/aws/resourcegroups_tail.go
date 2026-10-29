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
	"github.com/aws/aws-sdk-go-v2/service/resourcegroups"
)

// loadGroupResources adds aws_resourcegroups_resource membership resources.
// Provider import ID (verified): "<group ARN>,<resource ARN>".
func (g *ResourceGroupsGenerator) loadGroupResources(svc *resourcegroups.Client) {
	p := resourcegroups.NewListGroupsPaginator(svc, &resourcegroups.ListGroupsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return
		}
		for _, group := range page.GroupIdentifiers {
			groupARN := StringValue(group.GroupArn)
			out, err := svc.ListGroupResources(context.TODO(), &resourcegroups.ListGroupResourcesInput{Group: group.GroupArn})
			if err != nil {
				continue
			}
			for _, res := range out.Resources {
				if res.Identifier == nil {
					continue
				}
				id := groupARN + "," + StringValue(res.Identifier.ResourceArn)
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, id, "aws_resourcegroups_resource", "aws", resourcegroupsAllowEmptyValues))
			}
		}
	}
}
