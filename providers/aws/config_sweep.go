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

// Sweep extension for the AWS Config generator (service key "config").
// All ten remaining Config types are enumerable via Describe* list APIs.

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/configservice"
)

func (g *ConfigGenerator) createSweepResources(client *configservice.Client) error {
	if err := g.addAggregateAuthorizations(client); err != nil {
		return err
	}
	if err := g.addConfigurationAggregators(client); err != nil {
		return err
	}
	if err := g.addConfigurationRecorderStatuses(client); err != nil {
		return err
	}
	if err := g.addConformancePacks(client); err != nil {
		return err
	}
	if err := g.addOrganizationConformancePacks(client); err != nil {
		return err
	}
	ruleNames, err := g.addOrganizationConfigRules(client)
	if err != nil {
		return err
	}
	if err := g.addRemediationConfigurations(client, ruleNames); err != nil {
		return err
	}
	return g.addRetentionConfigurations(client)
}

// aws_config_aggregate_authorization — importer ID "account_id:authorized_aws_region".
func (g *ConfigGenerator) addAggregateAuthorizations(client *configservice.Client) error {
	var nextToken *string
	for {
		out, err := client.DescribeAggregationAuthorizations(context.TODO(), &configservice.DescribeAggregationAuthorizationsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, auth := range out.AggregationAuthorizations {
			id := StringValue(auth.AuthorizedAccountId) + ":" + StringValue(auth.AuthorizedAwsRegion)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				id, StringValue(auth.AuthorizedAccountId), "aws_config_aggregate_authorization", "aws",
				map[string]string{"account_id": StringValue(auth.AuthorizedAccountId), "authorized_aws_region": StringValue(auth.AuthorizedAwsRegion)},
				configAllowEmptyValues, map[string]interface{}{}))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

// aws_config_configuration_aggregator — importer ID is the aggregator name.
func (g *ConfigGenerator) addConfigurationAggregators(client *configservice.Client) error {
	var nextToken *string
	for {
		out, err := client.DescribeConfigurationAggregators(context.TODO(), &configservice.DescribeConfigurationAggregatorsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, agg := range out.ConfigurationAggregators {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(agg.ConfigurationAggregatorName), StringValue(agg.ConfigurationAggregatorName),
				"aws_config_configuration_aggregator", "aws", configAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

// aws_config_configuration_recorder_status — importer ID is the recorder name.
func (g *ConfigGenerator) addConfigurationRecorderStatuses(client *configservice.Client) error {
	out, err := client.DescribeConfigurationRecorderStatus(context.TODO(), &configservice.DescribeConfigurationRecorderStatusInput{})
	if err != nil {
		return err
	}
	for _, status := range out.ConfigurationRecordersStatus {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			StringValue(status.Name), StringValue(status.Name),
			"aws_config_configuration_recorder_status", "aws", configAllowEmptyValues))
	}
	return nil
}

// aws_config_conformance_pack — importer ID is the conformance pack name.
func (g *ConfigGenerator) addConformancePacks(client *configservice.Client) error {
	var nextToken *string
	for {
		out, err := client.DescribeConformancePacks(context.TODO(), &configservice.DescribeConformancePacksInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, pack := range out.ConformancePackDetails {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(pack.ConformancePackName), StringValue(pack.ConformancePackName),
				"aws_config_conformance_pack", "aws", configAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

// aws_config_organization_conformance_pack — importer ID is the pack name.
func (g *ConfigGenerator) addOrganizationConformancePacks(client *configservice.Client) error {
	var nextToken *string
	for {
		out, err := client.DescribeOrganizationConformancePacks(context.TODO(), &configservice.DescribeOrganizationConformancePacksInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, pack := range out.OrganizationConformancePacks {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(pack.OrganizationConformancePackName), StringValue(pack.OrganizationConformancePackName),
				"aws_config_organization_conformance_pack", "aws", configAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

// aws_config_organization_managed_rule / aws_config_organization_custom_rule /
// aws_config_organization_custom_policy_rule — importer IDs are the rule name.
// DescribeOrganizationConfigRules returns every organization rule; the variant
// is selected by which metadata block is populated. Returns all rule names so
// remediation configurations can be probed per rule.
func (g *ConfigGenerator) addOrganizationConfigRules(client *configservice.Client) ([]string, error) {
	var ruleNames []string
	var nextToken *string
	for {
		out, err := client.DescribeOrganizationConfigRules(context.TODO(), &configservice.DescribeOrganizationConfigRulesInput{NextToken: nextToken})
		if err != nil {
			return nil, err
		}
		for _, rule := range out.OrganizationConfigRules {
			name := StringValue(rule.OrganizationConfigRuleName)
			ruleNames = append(ruleNames, name)
			resourceType := ""
			switch {
			case rule.OrganizationCustomPolicyRuleMetadata != nil:
				resourceType = "aws_config_organization_custom_policy_rule"
			case rule.OrganizationCustomRuleMetadata != nil:
				resourceType = "aws_config_organization_custom_rule"
			case rule.OrganizationManagedRuleMetadata != nil:
				resourceType = "aws_config_organization_managed_rule"
			}
			if resourceType == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, resourceType, "aws", configAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return ruleNames, nil
}

// aws_config_remediation_configuration — importer ID is the config rule name.
// Remediation configurations exist per config rule; probe every rule returned
// by DescribeConfigRules plus the organization rules above.
func (g *ConfigGenerator) addRemediationConfigurations(client *configservice.Client, extraRuleNames []string) error {
	seen := map[string]bool{}
	var ruleNames []string
	add := func(name string) {
		if name != "" && !seen[name] {
			seen[name] = true
			ruleNames = append(ruleNames, name)
		}
	}
	var nextToken *string
	for {
		out, err := client.DescribeConfigRules(context.TODO(), &configservice.DescribeConfigRulesInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, rule := range out.ConfigRules {
			add(StringValue(rule.ConfigRuleName))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	for _, name := range extraRuleNames {
		add(name)
	}
	// DescribeRemediationConfigurations accepts multiple rule names per call.
	for i := 0; i < len(ruleNames); i += 25 {
		end := i + 25
		if end > len(ruleNames) {
			end = len(ruleNames)
		}
		out, err := client.DescribeRemediationConfigurations(context.TODO(), &configservice.DescribeRemediationConfigurationsInput{ConfigRuleNames: ruleNames[i:end]})
		if err != nil {
			return err
		}
		for _, remediation := range out.RemediationConfigurations {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(remediation.ConfigRuleName), StringValue(remediation.ConfigRuleName),
				"aws_config_remediation_configuration", "aws", configAllowEmptyValues))
		}
	}
	return nil
}

// aws_config_retention_configuration — importer ID is the configuration name
// (always "default"); DescribeRetentionConfigurations returns it as a list.
func (g *ConfigGenerator) addRetentionConfigurations(client *configservice.Client) error {
	var nextToken *string
	for {
		out, err := client.DescribeRetentionConfigurations(context.TODO(), &configservice.DescribeRetentionConfigurationsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, retention := range out.RetentionConfigurations {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(retention.Name), StringValue(retention.Name),
				"aws_config_retention_configuration", "aws", configAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
