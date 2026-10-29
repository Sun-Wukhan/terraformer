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
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	"github.com/aws/aws-sdk-go-v2/service/wafv2/types"
)

// Depth enumeration for WAFv2, called from Wafv2Generator.InitResources
// for both scopes: the individual rules of every web ACL
// (aws_wafv2_web_acl_rule, keyed by web_acl_arn,rule_name) and, for
// rules whose statement references a rule group, the corresponding
// aws_wafv2_web_acl_rule_group_association (keyed by
// web_acl_arn,rule_name,custom,rule_group_arn for custom rule groups
// and web_acl_arn,rule_name,managed,vendor:name[:version] for managed
// rule groups). Both adopter resources describe the same underlying
// web ACL rules that the aws_wafv2_web_acl resource carries inline;
// they are generated so the rules can be adopted under either
// management style.
//
// aws_wafv2_api_key is intentionally not generated: its provider ID
// is the generated API key value itself (api_key,scope), and the
// ListAPIKeys API returns only key summaries (token domains, creation
// time) — the key value is never returned by any list or describe
// call (GetDecryptedAPIKey requires the encrypted key as input), so
// the import ID cannot be reconstructed.
func (g *Wafv2Generator) addDepthResources(svc *wafv2.Client) error {
	output, err := svc.ListWebACLs(context.TODO(), &wafv2.ListWebACLsInput{Scope: g.scope})
	if err != nil {
		return err
	}
	for _, acl := range output.WebACLs {
		if err := g.addWebACLRules(svc, acl); err != nil {
			return err
		}
	}
	return nil
}

func (g *Wafv2Generator) addWebACLRules(svc *wafv2.Client, summary types.WebACLSummary) error {
	out, err := svc.GetWebACL(context.TODO(), &wafv2.GetWebACLInput{
		Name:  summary.Name,
		Scope: g.scope,
		Id:    summary.Id,
	})
	if err != nil {
		return err
	}
	if out.WebACL == nil {
		return nil
	}
	webACLArn := StringValue(out.WebACL.ARN)
	for _, rule := range out.WebACL.Rules {
		ruleName := StringValue(rule.Name)
		g.Resources = append(g.Resources, terraformutils.NewResource(
			webACLArn+","+ruleName,
			ruleName,
			"aws_wafv2_web_acl_rule",
			"aws",
			map[string]string{"web_acl_arn": webACLArn},
			wafv2AllowEmptyValues,
			map[string]interface{}{},
		))
		if rule.Statement == nil {
			continue
		}
		var associationID string
		switch {
		case rule.Statement.RuleGroupReferenceStatement != nil:
			associationID = webACLArn + "," + ruleName + ",custom," + StringValue(rule.Statement.RuleGroupReferenceStatement.ARN)
		case rule.Statement.ManagedRuleGroupStatement != nil:
			managed := rule.Statement.ManagedRuleGroupStatement
			associationID = webACLArn + "," + ruleName + ",managed," + StringValue(managed.VendorName) + ":" + StringValue(managed.Name)
			if managed.Version != nil && *managed.Version != "" {
				associationID += ":" + *managed.Version
			}
		default:
			continue
		}
		g.Resources = append(g.Resources, terraformutils.NewResource(
			associationID,
			ruleName,
			"aws_wafv2_web_acl_rule_group_association",
			"aws",
			map[string]string{"web_acl_arn": webACLArn},
			wafv2AllowEmptyValues,
			map[string]interface{}{},
		))
	}
	return nil
}
