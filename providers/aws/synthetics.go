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

// CloudWatch Synthetics generator (service key "synthetics").

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/synthetics"
)

var syntheticsAllowEmptyValues = []string{"tags."}

type SyntheticsGenerator struct {
	AWSService
}

func (g *SyntheticsGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := synthetics.NewFromConfig(config)
	if err := g.createCanaries(svc); err != nil {
		return err
	}
	return g.createGroups(svc)
}

// aws_synthetics_canary — importer ID is the canary name.
func (g *SyntheticsGenerator) createCanaries(svc *synthetics.Client) error {
	var nextToken *string
	for {
		out, err := svc.DescribeCanaries(context.TODO(), &synthetics.DescribeCanariesInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, canary := range out.Canaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(canary.Name), StringValue(canary.Name),
				"aws_synthetics_canary", "aws", syntheticsAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

// aws_synthetics_group — importer ID is the group name.
// aws_synthetics_group_association — importer ID "canary_arn,group_name";
// ListGroupResources returns the ARNs of the resources (canaries) in a group.
func (g *SyntheticsGenerator) createGroups(svc *synthetics.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListGroups(context.TODO(), &synthetics.ListGroupsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, group := range out.Groups {
			groupName := StringValue(group.Name)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				groupName, groupName, "aws_synthetics_group", "aws", syntheticsAllowEmptyValues))
			var resourceNextToken *string
			for {
				resources, err := svc.ListGroupResources(context.TODO(), &synthetics.ListGroupResourcesInput{GroupIdentifier: group.Name, NextToken: resourceNextToken})
				if err != nil {
					return err
				}
				for _, resourceARN := range resources.Resources {
					id := resourceARN + "," + groupName
					g.Resources = append(g.Resources, terraformutils.NewResource(
						id, groupName, "aws_synthetics_group_association", "aws",
						map[string]string{"group_name": groupName, "canary_arn": resourceARN},
						syntheticsAllowEmptyValues, map[string]interface{}{}))
				}
				resourceNextToken = resources.NextToken
				if resourceNextToken == nil {
					break
				}
			}
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
