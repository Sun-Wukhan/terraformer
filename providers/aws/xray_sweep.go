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

// Sweep extension for the X-Ray generator (service key "xray").
//
// Skipped types in this family (see results ledger):
//   - aws_xray_encryption_config: account/Region singleton; the provider ID
//     is the Region name and only GetEncryptionConfig exists.
//   - aws_xray_trace_segment_destination: account/Region singleton; the
//     provider ID is the Region name and only GetTraceSegmentDestination
//     exists.

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/xray"
)

func (g *XrayGenerator) createSweepResources(svc *xray.Client) error {
	if err := g.createGroups(svc); err != nil {
		return err
	}
	if err := g.createIndexingRules(svc); err != nil {
		return err
	}
	return g.createResourcePolicies(svc)
}

// aws_xray_group — importer ID is the group ARN.
func (g *XrayGenerator) createGroups(svc *xray.Client) error {
	var nextToken *string
	for {
		out, err := svc.GetGroups(context.TODO(), &xray.GetGroupsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, group := range out.Groups {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(group.GroupARN), StringValue(group.GroupName),
				"aws_xray_group", "aws", xrayAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

// aws_xray_indexing_rule — importer ID is the rule name. GetIndexingRules
// returns every rule (including the built-in "Default") in one paginated call.
func (g *XrayGenerator) createIndexingRules(svc *xray.Client) error {
	var nextToken *string
	for {
		out, err := svc.GetIndexingRules(context.TODO(), &xray.GetIndexingRulesInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, rule := range out.IndexingRules {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(rule.Name), StringValue(rule.Name),
				"aws_xray_indexing_rule", "aws", xrayAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

// aws_xray_resource_policy — importer ID is the policy name.
func (g *XrayGenerator) createResourcePolicies(svc *xray.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListResourcePolicies(context.TODO(), &xray.ListResourcePoliciesInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, policy := range out.ResourcePolicies {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(policy.PolicyName), StringValue(policy.PolicyName),
				"aws_xray_resource_policy", "aws", xrayAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
