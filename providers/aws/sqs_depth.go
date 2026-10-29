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
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// Depth coverage for SQS: the per-queue policy resources. Each imports with
// the queue URL as ID (verified against the Terraform AWS provider docs) and
// is emitted only when the corresponding queue attribute is set.

func (g *SqsGenerator) loadQueuePolicies(svc *sqs.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListQueues(context.TODO(), &sqs.ListQueuesInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, queueURL := range out.QueueUrls {
			attrs, err := svc.GetQueueAttributes(context.TODO(), &sqs.GetQueueAttributesInput{
				QueueUrl:       &queueURL,
				AttributeNames: []types.QueueAttributeName{types.QueueAttributeNameAll},
			})
			if err != nil {
				return err
			}
			urlParts := strings.Split(queueURL, "/")
			queueName := urlParts[len(urlParts)-1]
			if _, ok := attrs.Attributes["Policy"]; ok {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					queueURL,
					queueName,
					"aws_sqs_queue_policy",
					"aws",
					sqsAllowEmptyValues))
			}
			if _, ok := attrs.Attributes["RedrivePolicy"]; ok {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					queueURL,
					queueName,
					"aws_sqs_queue_redrive_policy",
					"aws",
					sqsAllowEmptyValues))
			}
			if _, ok := attrs.Attributes["RedriveAllowPolicy"]; ok {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					queueURL,
					queueName,
					"aws_sqs_queue_redrive_allow_policy",
					"aws",
					sqsAllowEmptyValues))
			}
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
