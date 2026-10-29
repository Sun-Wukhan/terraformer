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
	"github.com/aws/aws-sdk-go-v2/service/shield"
)

var shieldAllowEmptyValues = []string{"tags."}

type ShieldGenerator struct {
	AWSService
}

// InitResources enumerates the AWS Shield Advanced objects visible
// to the account: protections with their health check associations,
// protection groups, and the DRT access log bucket associations.
//
// Intentionally not generated:
//   - aws_shield_application_layer_automatic_response: the provider
//     documentation has no Import section for this resource (no
//     provider importer), although the per-protection configuration
//     is visible in ListProtections output.
//
// Single account-level objects with no list API (each is the one
// result of a Describe call):
//   - aws_shield_subscription: the account's single Shield Advanced
//     subscription (DescribeSubscription).
//   - aws_shield_proactive_engagement: the account's single proactive
//     engagement configuration (DescribeProactiveEngagement).
//   - aws_shield_drt_access_role_arn_association: the account's
//     single DRT access role (the RoleArn field of DescribeDRTAccess).
func (g *ShieldGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	client := shield.NewFromConfig(config)
	if err := g.addProtections(client); err != nil {
		return err
	}
	if err := g.addProtectionGroups(client); err != nil {
		return err
	}
	return g.addDRTAccessLogBucketAssociations(client)
}

func (g *ShieldGenerator) addProtections(client *shield.Client) error {
	input := &shield.ListProtectionsInput{}
	for {
		out, err := client.ListProtections(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, protection := range out.Protections {
			protectionID := StringValue(protection.Id)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				protectionID,
				StringValue(protection.Name),
				"aws_shield_protection",
				"aws",
				shieldAllowEmptyValues,
			))
			// Provider ID format: protection_id+health_check_arn.
			// Route 53 health check ARNs are global and are built from
			// the health check ID returned by ListProtections.
			for _, healthCheckID := range protection.HealthCheckIds {
				healthCheckArn := "arn:aws:route53:::healthcheck/" + healthCheckID
				g.Resources = append(g.Resources, terraformutils.NewResource(
					protectionID+"+"+healthCheckArn,
					healthCheckID,
					"aws_shield_protection_health_check_association",
					"aws",
					map[string]string{"shield_protection_id": protectionID, "health_check_arn": healthCheckArn},
					shieldAllowEmptyValues,
					map[string]interface{}{},
				))
			}
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *ShieldGenerator) addProtectionGroups(client *shield.Client) error {
	input := &shield.ListProtectionGroupsInput{}
	for {
		out, err := client.ListProtectionGroups(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, group := range out.ProtectionGroups {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(group.ProtectionGroupId),
				StringValue(group.ProtectionGroupId),
				"aws_shield_protection_group",
				"aws",
				shieldAllowEmptyValues,
			))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

// Provider ID format: log_bucket.
func (g *ShieldGenerator) addDRTAccessLogBucketAssociations(client *shield.Client) error {
	out, err := client.DescribeDRTAccess(context.TODO(), &shield.DescribeDRTAccessInput{})
	if err != nil {
		return err
	}
	for _, bucket := range out.LogBucketList {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			bucket,
			bucket,
			"aws_shield_drt_access_log_bucket_association",
			"aws",
			shieldAllowEmptyValues,
		))
	}
	return nil
}
