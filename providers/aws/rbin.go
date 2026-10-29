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
	"github.com/aws/aws-sdk-go-v2/service/rbin"
	"github.com/aws/aws-sdk-go-v2/service/rbin/types"
)

// Recycle Bin sweep. aws_rbin_rule imports with the rule identifier
// (verified against provider docs). ListRules requires a resource type,
// so both supported types are enumerated.

var rbinAllowEmptyValues = []string{"tags."}

type RbinGenerator struct {
	AWSService
}

func (g *RbinGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := rbin.NewFromConfig(config)
	ctx := context.TODO()
	for _, rt := range []types.ResourceType{types.ResourceTypeEbsSnapshot, types.ResourceTypeEc2Image} {
		p := rbin.NewListRulesPaginator(svc, &rbin.ListRulesInput{ResourceType: rt})
		for p.HasMorePages() {
			page, err := p.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, rule := range page.Rules {
				id := StringValue(rule.Identifier)
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(id, id, "aws_rbin_rule", "aws", rbinAllowEmptyValues))
			}
		}
	}
	return nil
}
