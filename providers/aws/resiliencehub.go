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

// AWS Resilience Hub (v1) generator (service key "resiliencehub").

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/resiliencehub"
)

var resiliencehubAllowEmptyValues = []string{"tags."}

type ResilienceHubGenerator struct {
	AWSService
}

// aws_resiliencehub_resiliency_policy — importer ID is the policy ARN.
func (g *ResilienceHubGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := resiliencehub.NewFromConfig(config)
	var nextToken *string
	for {
		out, err := svc.ListResiliencyPolicies(context.TODO(), &resiliencehub.ListResiliencyPoliciesInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, policy := range out.ResiliencyPolicies {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(policy.PolicyArn), StringValue(policy.PolicyName),
				"aws_resiliencehub_resiliency_policy", "aws", resiliencehubAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
