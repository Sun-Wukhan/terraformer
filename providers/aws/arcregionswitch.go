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
	"github.com/aws/aws-sdk-go-v2/service/arcregionswitch"
)

var arcRegionSwitchAllowEmptyValues = []string{"tags."}

type ArcRegionSwitchGenerator struct {
	AWSService
}

// InitResources enumerates the Amazon Application Recovery
// Controller region switch plans in the configured region. Plans are
// keyed by their ARN.
func (g *ArcRegionSwitchGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	client := arcregionswitch.NewFromConfig(config)
	input := &arcregionswitch.ListPlansInput{}
	for {
		out, err := client.ListPlans(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, plan := range out.Plans {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(plan.Arn),
				StringValue(plan.Name),
				"aws_arcregionswitch_plan",
				"aws",
				arcRegionSwitchAllowEmptyValues,
			))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}
