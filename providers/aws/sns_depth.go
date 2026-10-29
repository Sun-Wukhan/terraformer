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
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

// Depth coverage for SNS. Import IDs verified against the Terraform AWS
// provider docs: platform application and both topic policy resources import
// with the topic/application ARN.
//
// Skipped in this family:
//   - aws_sns_sms_preferences: singleton account-level SMS settings; not an
//     enumerable resource.

func (g *SnsGenerator) loadSnsDepth(svc *sns.Client) error {
	if err := g.loadPlatformApplications(svc); err != nil {
		return err
	}
	if err := g.loadTopicPolicies(svc); err != nil {
		return err
	}
	return nil
}

func (g *SnsGenerator) loadPlatformApplications(svc *sns.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListPlatformApplications(context.TODO(), &sns.ListPlatformApplicationsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, app := range out.PlatformApplications {
			arn := StringValue(app.PlatformApplicationArn)
			arnParts := strings.Split(arn, "/")
			name := arnParts[len(arnParts)-1]
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn,
				name,
				"aws_sns_platform_application",
				"aws",
				snsAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

func (g *SnsGenerator) loadTopicPolicies(svc *sns.Client) error {
	p := sns.NewListTopicsPaginator(svc, &sns.ListTopicsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, topic := range page.Topics {
			topicArn := StringValue(topic.TopicArn)
			attrs, err := svc.GetTopicAttributes(context.TODO(), &sns.GetTopicAttributesInput{TopicArn: topic.TopicArn})
			if err != nil {
				return err
			}
			arnParts := strings.Split(topicArn, ":")
			topicName := arnParts[len(arnParts)-1]
			if _, ok := attrs.Attributes["Policy"]; ok {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					topicArn,
					topicName,
					"aws_sns_topic_policy",
					"aws",
					snsAllowEmptyValues))
			}
			if _, ok := attrs.Attributes["DataProtectionPolicy"]; ok {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					topicArn,
					topicName,
					"aws_sns_topic_data_protection_policy",
					"aws",
					snsAllowEmptyValues))
			}
		}
	}
	return nil
}
