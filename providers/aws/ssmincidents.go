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
	"github.com/aws/aws-sdk-go-v2/service/ssmincidents"
)

var ssmincidentsAllowEmptyValues = []string{"tags."}

// SsmIncidentsGenerator covers Incident Manager response plans.
// Provider import ID (verified): response plan ARN.
// Skipped in this family:
//   - aws_ssmincidents_replication_set: singleton per account/region — the
//     provider import ID is the literal string "import" and at most one
//     replication set exists, so it is an account-level settings object rather
//     than an enumerable inventory.
type SsmIncidentsGenerator struct {
	AWSService
}

func (g *SsmIncidentsGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := ssmincidents.NewFromConfig(config)
	p := ssmincidents.NewListResponsePlansPaginator(svc, &ssmincidents.ListResponsePlansInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, plan := range page.ResponsePlanSummaries {
			arn := StringValue(plan.Arn)
			name := StringValue(plan.Name)
			if name == "" {
				name = arn
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, name, "aws_ssmincidents_response_plan", "aws", ssmincidentsAllowEmptyValues))
		}
	}
	return nil
}
