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
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cftypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

// loadTailResources adds aws_cloudformation_stack_instances and aws_cloudformation_type.
// Provider import IDs (verified): stack instances = "<stack set name>,SELF"
// (the SELF call-as form for the management account); type = type version ARN.
// Skipped in this family:
//   - aws_cloudcontrolapi_resource: the provider documentation has no Import
//     section (no provider importer), and Cloud Control ListResources requires a
//     caller-supplied resource type name, so there is no global inventory to enumerate.
func (g *CloudFormationGenerator) loadTailResources(svc *cloudformation.Client) {
	stackSets, err := svc.ListStackSets(context.TODO(), &cloudformation.ListStackSetsInput{})
	if err == nil {
		for _, summary := range stackSets.Summaries {
			if summary.Status == cftypes.StackSetStatusDeleted {
				continue
			}
			id := StringValue(summary.StackSetName) + ",SELF"
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_cloudformation_stack_instances", "aws", cloudFormationAllowEmptyValues))
		}
	}

	p := cloudformation.NewListTypesPaginator(svc, &cloudformation.ListTypesInput{
		ProvisioningType: cftypes.ProvisioningTypeFullyMutable,
		Visibility:       cftypes.VisibilityPrivate,
	})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			break
		}
		for _, summary := range page.TypeSummaries {
			versions, err := svc.ListTypeVersions(context.TODO(), &cloudformation.ListTypeVersionsInput{
				Type:     summary.Type,
				TypeName: summary.TypeName,
				Arn:      summary.TypeArn,
			})
			if err != nil {
				continue
			}
			for _, version := range versions.TypeVersionSummaries {
				arn := StringValue(version.Arn)
				if arn == "" {
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					arn, StringValue(summary.TypeName), "aws_cloudformation_type", "aws", cloudFormationAllowEmptyValues))
				_ = aws.String
			}
		}
	}
}
