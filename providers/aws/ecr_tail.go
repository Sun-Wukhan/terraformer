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
	"github.com/aws/aws-sdk-go-v2/service/ecr"
)

// loadTailResources adds ECR pull-through cache rules, repository creation
// templates and pull-time update exclusions.
// Provider import IDs (verified): pull-through rule = repository prefix;
// creation template = prefix; pull-time exclusion = principal ARN.
// Skipped in this family (singleton registry/account objects — the provider ID
// is the registry ID or a fixed setting name and there is no list API; existence
// is a single Get call for the caller's own registry):
//   - aws_ecr_account_setting (ID = fixed setting name, e.g. BASIC_SCAN_TYPE_VERSION)
//   - aws_ecr_registry_policy (ID = registry ID)
//   - aws_ecr_registry_scanning_configuration (ID = registry ID)
//   - aws_ecr_replication_configuration (ID = registry ID)
func (g *EcrGenerator) loadTailResources(svc *ecr.Client) {
	rules, err := svc.DescribePullThroughCacheRules(context.TODO(), &ecr.DescribePullThroughCacheRulesInput{})
	if err == nil {
		for _, rule := range rules.PullThroughCacheRules {
			prefix := StringValue(rule.EcrRepositoryPrefix)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				prefix, prefix, "aws_ecr_pull_through_cache_rule", "aws", ecrAllowEmptyValues))
		}
	}
	templates, err := svc.DescribeRepositoryCreationTemplates(context.TODO(), &ecr.DescribeRepositoryCreationTemplatesInput{})
	if err == nil {
		for _, tpl := range templates.RepositoryCreationTemplates {
			prefix := StringValue(tpl.Prefix)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				prefix, prefix, "aws_ecr_repository_creation_template", "aws", ecrAllowEmptyValues))
		}
	}
	var nextToken *string
	for {
		page, err := svc.ListPullTimeUpdateExclusions(context.TODO(), &ecr.ListPullTimeUpdateExclusionsInput{NextToken: nextToken})
		if err != nil {
			break
		}
		for _, principalARN := range page.PullTimeUpdateExclusions {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				principalARN, principalARN, "aws_ecr_pull_time_update_exclusion", "aws", ecrAllowEmptyValues))
		}
		if page.NextToken == nil {
			break
		}
		nextToken = page.NextToken
	}
}
