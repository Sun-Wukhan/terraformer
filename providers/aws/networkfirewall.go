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
	"errors"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall/types"
)

var networkFirewallAllowEmptyValues = []string{"tags."}

type NetworkFirewallGenerator struct {
	AWSService
}

// InitResources enumerates the AWS Network Firewall objects in the
// configured region: firewalls (with their VPC endpoint associations,
// logging configuration and resource policy), firewall policies, rule
// groups, TLS inspection configurations and container associations.
// Policy-carrying objects are also probed for a resource policy,
// which is keyed by the ARN of the object it is attached to.
//
// aws_networkfirewall_firewall_transit_gateway_attachment_accepter
// is intentionally not generated: it is an action-style accepter
// resource with no importer in the Terraform AWS provider and no
// independent listable object (the attachment itself is an attribute
// of the firewall's transit gateway attachment).
func (g *NetworkFirewallGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	client := networkfirewall.NewFromConfig(config)
	for _, add := range []func(*networkfirewall.Client) error{
		g.addFirewalls,
		g.addFirewallPolicies,
		g.addRuleGroups,
		g.addTLSInspectionConfigurations,
		g.addContainerAssociations,
	} {
		if err := add(client); err != nil {
			return err
		}
	}
	return nil
}

func (g *NetworkFirewallGenerator) addFirewalls(client *networkfirewall.Client) error {
	input := &networkfirewall.ListFirewallsInput{}
	for {
		out, err := client.ListFirewalls(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, firewall := range out.Firewalls {
			arn := StringValue(firewall.FirewallArn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn,
				StringValue(firewall.FirewallName),
				"aws_networkfirewall_firewall",
				"aws",
				networkFirewallAllowEmptyValues,
			))
			if err := g.addVpcEndpointAssociations(client, arn); err != nil {
				return err
			}
			if err := g.addLoggingConfiguration(client, arn, StringValue(firewall.FirewallName)); err != nil {
				return err
			}
			if err := g.addResourcePolicy(client, arn); err != nil {
				return err
			}
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *NetworkFirewallGenerator) addVpcEndpointAssociations(client *networkfirewall.Client, firewallArn string) error {
	input := &networkfirewall.ListVpcEndpointAssociationsInput{FirewallArn: aws.String(firewallArn)}
	for {
		out, err := client.ListVpcEndpointAssociations(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, association := range out.VpcEndpointAssociations {
			associationArn := StringValue(association.VpcEndpointAssociationArn)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				associationArn,
				associationArn,
				"aws_networkfirewall_vpc_endpoint_association",
				"aws",
				map[string]string{"firewall_arn": firewallArn},
				networkFirewallAllowEmptyValues,
				map[string]interface{}{},
			))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

// The logging configuration is a per-firewall singleton keyed by the
// firewall ARN; it is generated only when at least one log destination
// is configured.
func (g *NetworkFirewallGenerator) addLoggingConfiguration(client *networkfirewall.Client, firewallArn, firewallName string) error {
	out, err := client.DescribeLoggingConfiguration(context.TODO(), &networkfirewall.DescribeLoggingConfigurationInput{
		FirewallArn: aws.String(firewallArn),
	})
	if err != nil {
		return err
	}
	if out.LoggingConfiguration == nil || len(out.LoggingConfiguration.LogDestinationConfigs) == 0 {
		return nil
	}
	g.Resources = append(g.Resources, terraformutils.NewResource(
		firewallArn,
		firewallName,
		"aws_networkfirewall_logging_configuration",
		"aws",
		map[string]string{"firewall_arn": firewallArn},
		networkFirewallAllowEmptyValues,
		map[string]interface{}{},
	))
	return nil
}

func (g *NetworkFirewallGenerator) addFirewallPolicies(client *networkfirewall.Client) error {
	input := &networkfirewall.ListFirewallPoliciesInput{}
	for {
		out, err := client.ListFirewallPolicies(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, policy := range out.FirewallPolicies {
			arn := StringValue(policy.Arn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn,
				StringValue(policy.Name),
				"aws_networkfirewall_firewall_policy",
				"aws",
				networkFirewallAllowEmptyValues,
			))
			if err := g.addResourcePolicy(client, arn); err != nil {
				return err
			}
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *NetworkFirewallGenerator) addRuleGroups(client *networkfirewall.Client) error {
	input := &networkfirewall.ListRuleGroupsInput{}
	for {
		out, err := client.ListRuleGroups(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, group := range out.RuleGroups {
			arn := StringValue(group.Arn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn,
				StringValue(group.Name),
				"aws_networkfirewall_rule_group",
				"aws",
				networkFirewallAllowEmptyValues,
			))
			if err := g.addResourcePolicy(client, arn); err != nil {
				return err
			}
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *NetworkFirewallGenerator) addTLSInspectionConfigurations(client *networkfirewall.Client) error {
	input := &networkfirewall.ListTLSInspectionConfigurationsInput{}
	for {
		out, err := client.ListTLSInspectionConfigurations(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, tlsConfig := range out.TLSInspectionConfigurations {
			arn := StringValue(tlsConfig.Arn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn,
				StringValue(tlsConfig.Name),
				"aws_networkfirewall_tls_inspection_configuration",
				"aws",
				networkFirewallAllowEmptyValues,
			))
			if err := g.addResourcePolicy(client, arn); err != nil {
				return err
			}
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *NetworkFirewallGenerator) addContainerAssociations(client *networkfirewall.Client) error {
	input := &networkfirewall.ListContainerAssociationsInput{}
	for {
		out, err := client.ListContainerAssociations(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, association := range out.ContainerAssociations {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(association.Arn),
				StringValue(association.Name),
				"aws_networkfirewall_container_association",
				"aws",
				networkFirewallAllowEmptyValues,
			))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

// The resource policy resource is keyed by the ARN of the object the
// policy is attached to and exists only when a policy document is
// actually attached; the probe reports ResourceNotFoundException
// otherwise.
func (g *NetworkFirewallGenerator) addResourcePolicy(client *networkfirewall.Client, resourceArn string) error {
	out, err := client.DescribeResourcePolicy(context.TODO(), &networkfirewall.DescribeResourcePolicyInput{
		ResourceArn: aws.String(resourceArn),
	})
	if err != nil {
		var notFound *types.ResourceNotFoundException
		if errors.As(err, &notFound) {
			return nil
		}
		return err
	}
	if out.Policy == nil || *out.Policy == "" {
		return nil
	}
	g.Resources = append(g.Resources, terraformutils.NewResource(
		resourceArn,
		resourceArn,
		"aws_networkfirewall_resource_policy",
		"aws",
		map[string]string{"resource_arn": resourceArn},
		networkFirewallAllowEmptyValues,
		map[string]interface{}{},
	))
	return nil
}
