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
	"log"
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
)

// Security Hub depth pass (AWS full-coverage sweep, security
// cluster). Called from SecurityhubGenerator.InitResources once the
// hub is known to be enabled.
//
// Implemented here (provider importer ID formats per the
// terraform-provider-aws resource docs):
//   - aws_securityhub_action_target: the action target ARN, from
//     DescribeActionTargets.
//   - aws_securityhub_automation_rule: the rule ARN, from
//     ListAutomationRules.
//   - aws_securityhub_automation_rule_v2: the rule ARN, from
//     ListAutomationRulesV2.
//   - aws_securityhub_aggregator_v2: the aggregator ARN, from
//     ListAggregatorsV2.
//   - aws_securityhub_connector_v2: the connector ARN, from
//     ListConnectorsV2.
//   - aws_securityhub_configuration_policy: the policy UUID, from
//     ListConfigurationPolicies.
//   - aws_securityhub_finding_aggregator: the aggregator ARN, from
//     ListFindingAggregators.
//   - aws_securityhub_insight: the insight ARN, from GetInsights.
//   - aws_securityhub_product_subscription:
//     product_arn,product_subscription_arn (verified); the product
//     ARN is derived from each subscription ARN returned by
//     ListEnabledProductsForImport by dropping the account segment
//     and replacing "product-subscription/" with "product/".
//   - aws_securityhub_standards_control: the control ARN, from
//     DescribeStandardsControls for every enabled standard.
//
// Intentionally not generated:
//   - aws_securityhub_account_v2 is the singleton Security Hub V2
//     enablement for the account/region.
//   - aws_securityhub_configuration_policy_association is association
//     glue between a configuration policy and an account/OU/root
//     target; the API summary carries no association ID of its own.
//   - aws_securityhub_feature_v2 is a per-feature enablement toggle;
//     the SDK exposes only Enable/DisableSecurityHubFeatureV2 and no
//     list/get enumeration for it.
//   - aws_securityhub_invite_accepter is the accepter side of a
//     member invitation in the invited account.
//   - aws_securityhub_organization_admin_account is the singleton
//     delegated-administrator designation for the organization.
//   - aws_securityhub_organization_configuration is the singleton
//     organization configuration of the administrator account
//     (DescribeOrganizationConfiguration).
//   - aws_securityhub_standards_control_association is the newer
//     association form of the same standard/control enablement that
//     aws_securityhub_standards_control already covers; emitting both
//     would double-manage control enablement.
func (g *SecurityhubGenerator) addDepthResources(client *securityhub.Client) {
	g.addActionTargets(client)
	g.addAutomationRules(client)
	g.addAutomationRulesV2(client)
	g.addAggregatorsV2(client)
	g.addConnectorsV2(client)
	g.addConfigurationPolicies(client)
	g.addFindingAggregators(client)
	g.addInsights(client)
	g.addProductSubscriptions(client)
	g.addStandardsControls(client)
}

func (g *SecurityhubGenerator) addActionTargets(client *securityhub.Client) {
	p := securityhub.NewDescribeActionTargetsPaginator(client, &securityhub.DescribeActionTargetsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, target := range page.ActionTargets {
			arn := StringValue(target.ActionTargetArn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn,
				StringValue(target.Name),
				"aws_securityhub_action_target",
				"aws",
				securityhubAllowEmptyValues))
		}
	}
}

func (g *SecurityhubGenerator) addAutomationRules(client *securityhub.Client) {
	input := &securityhub.ListAutomationRulesInput{}
	for {
		page, err := client.ListAutomationRules(context.TODO(), input)
		if err != nil {
			log.Println(err)
			return
		}
		for _, rule := range page.AutomationRulesMetadata {
			arn := StringValue(rule.RuleArn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn,
				StringValue(rule.RuleName),
				"aws_securityhub_automation_rule",
				"aws",
				securityhubAllowEmptyValues))
		}
		if page.NextToken == nil {
			return
		}
		input.NextToken = page.NextToken
	}
}

func (g *SecurityhubGenerator) addAutomationRulesV2(client *securityhub.Client) {
	out, err := client.ListAutomationRulesV2(context.TODO(), &securityhub.ListAutomationRulesV2Input{})
	if err != nil {
		log.Println(err)
		return
	}
	for _, rule := range out.Rules {
		arn := StringValue(rule.RuleArn)
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			arn,
			StringValue(rule.RuleName),
			"aws_securityhub_automation_rule_v2",
			"aws",
			securityhubAllowEmptyValues))
	}
}

func (g *SecurityhubGenerator) addAggregatorsV2(client *securityhub.Client) {
	out, err := client.ListAggregatorsV2(context.TODO(), &securityhub.ListAggregatorsV2Input{})
	if err != nil {
		log.Println(err)
		return
	}
	for _, aggregator := range out.AggregatorsV2 {
		arn := StringValue(aggregator.AggregatorV2Arn)
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			arn,
			arnResourceName(arn),
			"aws_securityhub_aggregator_v2",
			"aws",
			securityhubAllowEmptyValues))
	}
}

func (g *SecurityhubGenerator) addConnectorsV2(client *securityhub.Client) {
	out, err := client.ListConnectorsV2(context.TODO(), &securityhub.ListConnectorsV2Input{})
	if err != nil {
		log.Println(err)
		return
	}
	for _, connector := range out.Connectors {
		arn := StringValue(connector.ConnectorArn)
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			arn,
			StringValue(connector.Name),
			"aws_securityhub_connector_v2",
			"aws",
			securityhubAllowEmptyValues))
	}
}

func (g *SecurityhubGenerator) addConfigurationPolicies(client *securityhub.Client) {
	p := securityhub.NewListConfigurationPoliciesPaginator(client, &securityhub.ListConfigurationPoliciesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, policy := range page.ConfigurationPolicySummaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(policy.Id),
				StringValue(policy.Name),
				"aws_securityhub_configuration_policy",
				"aws",
				securityhubAllowEmptyValues))
		}
	}
}

func (g *SecurityhubGenerator) addFindingAggregators(client *securityhub.Client) {
	p := securityhub.NewListFindingAggregatorsPaginator(client, &securityhub.ListFindingAggregatorsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, aggregator := range page.FindingAggregators {
			arn := StringValue(aggregator.FindingAggregatorArn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn,
				arnResourceName(arn),
				"aws_securityhub_finding_aggregator",
				"aws",
				securityhubAllowEmptyValues))
		}
	}
}

func (g *SecurityhubGenerator) addInsights(client *securityhub.Client) {
	p := securityhub.NewGetInsightsPaginator(client, &securityhub.GetInsightsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, insight := range page.Insights {
			arn := StringValue(insight.InsightArn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn,
				StringValue(insight.Name),
				"aws_securityhub_insight",
				"aws",
				securityhubAllowEmptyValues))
		}
	}
}

func (g *SecurityhubGenerator) addProductSubscriptions(client *securityhub.Client) {
	out, err := client.ListEnabledProductsForImport(context.TODO(), &securityhub.ListEnabledProductsForImportInput{})
	if err != nil {
		log.Println(err)
		return
	}
	for _, subscriptionArn := range out.ProductSubscriptions {
		// subscription ARN: arn:aws:securityhub:<region>:<account>:product-subscription/<vendor>/<product>
		// product ARN:      arn:aws:securityhub:<region>::product/<vendor>/<product>
		productArn := strings.Replace(subscriptionArn, "product-subscription/", "product/", 1)
		parts := strings.SplitN(productArn, ":", 6)
		if len(parts) == 6 {
			parts[4] = ""
			productArn = strings.Join(parts, ":")
		}
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			productArn+","+subscriptionArn,
			arnResourceName(subscriptionArn),
			"aws_securityhub_product_subscription",
			"aws",
			securityhubAllowEmptyValues))
	}
}

func (g *SecurityhubGenerator) addStandardsControls(client *securityhub.Client) {
	standards, err := client.GetEnabledStandards(context.TODO(), &securityhub.GetEnabledStandardsInput{})
	if err != nil {
		log.Println(err)
		return
	}
	for _, subscription := range standards.StandardsSubscriptions {
		p := securityhub.NewDescribeStandardsControlsPaginator(client, &securityhub.DescribeStandardsControlsInput{
			StandardsSubscriptionArn: subscription.StandardsSubscriptionArn,
		})
		for p.HasMorePages() {
			page, err := p.NextPage(context.TODO())
			if err != nil {
				log.Println(err)
				break
			}
			for _, control := range page.Controls {
				arn := StringValue(control.StandardsControlArn)
				g.Resources = append(g.Resources, terraformutils.NewResource(
					arn,
					StringValue(control.ControlId),
					"aws_securityhub_standards_control",
					"aws",
					map[string]string{},
					securityhubAllowEmptyValues,
					map[string]interface{}{}))
			}
		}
	}
}
