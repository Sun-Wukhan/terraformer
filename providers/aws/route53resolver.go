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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53resolver"
)

var route53ResolverAllowEmptyValues = []string{"tags."}

type Route53ResolverGenerator struct {
	AWSService
}

// InitResources enumerates the Route 53 Resolver objects in the
// configured region: endpoints, rules and rule associations, query
// log configs and their associations, DNS Firewall domain lists, rule
// groups (with their rules and VPC associations), per-VPC firewall
// configs, and per-VPC DNSSEC configs.
//
// aws_route53_resolver_config is intentionally not generated: it is
// the single account/region-level Resolver settings object (one
// ListResolverConfigs result per account and region), a singleton
// settings resource rather than an enumerable object.
func (g *Route53ResolverGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	client := route53resolver.NewFromConfig(config)
	for _, add := range []func(*route53resolver.Client) error{
		g.addEndpoints,
		g.addRules,
		g.addRuleAssociations,
		g.addQueryLogConfigs,
		g.addQueryLogConfigAssociations,
		g.addFirewallDomainLists,
		g.addFirewallRuleGroups,
		g.addFirewallConfigs,
		g.addDNSSECConfigs,
	} {
		if err := add(client); err != nil {
			return err
		}
	}
	return nil
}

func (g *Route53ResolverGenerator) addEndpoints(client *route53resolver.Client) error {
	input := &route53resolver.ListResolverEndpointsInput{}
	for {
		out, err := client.ListResolverEndpoints(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, endpoint := range out.ResolverEndpoints {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(endpoint.Id),
				StringValue(endpoint.Name),
				"aws_route53_resolver_endpoint",
				"aws",
				route53ResolverAllowEmptyValues,
			))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *Route53ResolverGenerator) addRules(client *route53resolver.Client) error {
	input := &route53resolver.ListResolverRulesInput{}
	for {
		out, err := client.ListResolverRules(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, rule := range out.ResolverRules {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(rule.Id),
				StringValue(rule.Name),
				"aws_route53_resolver_rule",
				"aws",
				route53ResolverAllowEmptyValues,
			))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *Route53ResolverGenerator) addRuleAssociations(client *route53resolver.Client) error {
	input := &route53resolver.ListResolverRuleAssociationsInput{}
	for {
		out, err := client.ListResolverRuleAssociations(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, association := range out.ResolverRuleAssociations {
			name := StringValue(association.Name)
			if name == "" {
				name = StringValue(association.Id)
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(association.Id),
				name,
				"aws_route53_resolver_rule_association",
				"aws",
				route53ResolverAllowEmptyValues,
			))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *Route53ResolverGenerator) addQueryLogConfigs(client *route53resolver.Client) error {
	input := &route53resolver.ListResolverQueryLogConfigsInput{}
	for {
		out, err := client.ListResolverQueryLogConfigs(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, queryLogConfig := range out.ResolverQueryLogConfigs {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(queryLogConfig.Id),
				StringValue(queryLogConfig.Name),
				"aws_route53_resolver_query_log_config",
				"aws",
				route53ResolverAllowEmptyValues,
			))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *Route53ResolverGenerator) addQueryLogConfigAssociations(client *route53resolver.Client) error {
	input := &route53resolver.ListResolverQueryLogConfigAssociationsInput{}
	for {
		out, err := client.ListResolverQueryLogConfigAssociations(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, association := range out.ResolverQueryLogConfigAssociations {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(association.Id),
				StringValue(association.Id),
				"aws_route53_resolver_query_log_config_association",
				"aws",
				route53ResolverAllowEmptyValues,
			))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *Route53ResolverGenerator) addFirewallDomainLists(client *route53resolver.Client) error {
	input := &route53resolver.ListFirewallDomainListsInput{}
	for {
		out, err := client.ListFirewallDomainLists(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, domainList := range out.FirewallDomainLists {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(domainList.Id),
				StringValue(domainList.Name),
				"aws_route53_resolver_firewall_domain_list",
				"aws",
				route53ResolverAllowEmptyValues,
			))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *Route53ResolverGenerator) addFirewallRuleGroups(client *route53resolver.Client) error {
	input := &route53resolver.ListFirewallRuleGroupsInput{}
	for {
		out, err := client.ListFirewallRuleGroups(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, group := range out.FirewallRuleGroups {
			groupID := StringValue(group.Id)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				groupID,
				StringValue(group.Name),
				"aws_route53_resolver_firewall_rule_group",
				"aws",
				route53ResolverAllowEmptyValues,
			))
			if err := g.addFirewallRules(client, groupID); err != nil {
				return err
			}
		}
		if out.NextToken == nil {
			break
		}
		input.NextToken = out.NextToken
	}
	return g.addFirewallRuleGroupAssociations(client)
}

// Provider ID format: firewall_rule_group_id:firewall_domain_list_id
// for standard rules, firewall_rule_group_id:firewall_threat_protection_id
// for DNS Firewall Advanced rules.
func (g *Route53ResolverGenerator) addFirewallRules(client *route53resolver.Client, groupID string) error {
	input := &route53resolver.ListFirewallRulesInput{FirewallRuleGroupId: aws.String(groupID)}
	for {
		out, err := client.ListFirewallRules(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, rule := range out.FirewallRules {
			second := StringValue(rule.FirewallDomainListId)
			if second == "" {
				second = StringValue(rule.FirewallThreatProtectionId)
			}
			if second == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewResource(
				groupID+":"+second,
				StringValue(rule.Name),
				"aws_route53_resolver_firewall_rule",
				"aws",
				map[string]string{"firewall_rule_group_id": groupID},
				route53ResolverAllowEmptyValues,
				map[string]interface{}{},
			))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *Route53ResolverGenerator) addFirewallRuleGroupAssociations(client *route53resolver.Client) error {
	input := &route53resolver.ListFirewallRuleGroupAssociationsInput{}
	for {
		out, err := client.ListFirewallRuleGroupAssociations(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, association := range out.FirewallRuleGroupAssociations {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(association.Id),
				StringValue(association.Name),
				"aws_route53_resolver_firewall_rule_group_association",
				"aws",
				route53ResolverAllowEmptyValues,
			))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *Route53ResolverGenerator) addFirewallConfigs(client *route53resolver.Client) error {
	input := &route53resolver.ListFirewallConfigsInput{}
	for {
		out, err := client.ListFirewallConfigs(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, firewallConfig := range out.FirewallConfigs {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				StringValue(firewallConfig.Id),
				StringValue(firewallConfig.Id),
				"aws_route53_resolver_firewall_config",
				"aws",
				map[string]string{"resource_id": StringValue(firewallConfig.ResourceId)},
				route53ResolverAllowEmptyValues,
				map[string]interface{}{},
			))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *Route53ResolverGenerator) addDNSSECConfigs(client *route53resolver.Client) error {
	input := &route53resolver.ListResolverDnssecConfigsInput{}
	for {
		out, err := client.ListResolverDnssecConfigs(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, dnssecConfig := range out.ResolverDnssecConfigs {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				StringValue(dnssecConfig.Id),
				StringValue(dnssecConfig.Id),
				"aws_route53_resolver_dnssec_config",
				"aws",
				map[string]string{"resource_id": StringValue(dnssecConfig.ResourceId)},
				route53ResolverAllowEmptyValues,
				map[string]interface{}{},
			))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}
