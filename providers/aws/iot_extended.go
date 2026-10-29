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
	"github.com/aws/aws-sdk-go-v2/service/iot"
)

// IoT depth for the AWS sweep. Provider importer IDs verified against the
// terraform-provider-aws docs: authorizers, billing groups, domain
// configurations, policies, provisioning templates and thing groups are
// keyed by name; CA and device certificates by certificate ID; topic rule
// destinations by ARN; the account event configurations singleton by region.
//
// Deliberately not generated:
//   - aws_iot_indexing_configuration: the provider docs carry no Import
//     section for this singleton (no provider importer).
//   - aws_iot_logging_options: singleton account logging settings; no
//     list API (only GetLoggingOptions for the account singleton).
//   - aws_iot_policy_attachment, aws_iot_thing_principal_attachment and
//     aws_iot_thing_group_membership: policy/principal and thing/group
//     association glue, not standalone resources.
func (g *IotGenerator) loadExtended(svc *iot.Client, region string) error {
	for _, load := range []func(*iot.Client) error{
		g.loadAuthorizers, g.loadBillingGroups, g.loadCACertificates,
		g.loadCertificatesIoT, g.loadDomainConfigurations, g.loadPoliciesIoT,
		g.loadProvisioningTemplates, g.loadThingGroups, g.loadTopicRuleDestinations,
	} {
		if err := load(svc); err != nil {
			return err
		}
	}
	if region != "" {
		out, err := svc.DescribeEventConfigurations(context.TODO(), &iot.DescribeEventConfigurationsInput{})
		if err != nil {
			return err
		}
		if len(out.EventConfigurations) > 0 {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				region, "iot_event_configurations", "aws_iot_event_configurations", "aws", iotAllowEmptyValues))
		}
	}
	return nil
}

func (g *IotGenerator) loadAuthorizers(svc *iot.Client) error {
	input := &iot.ListAuthorizersInput{}
	for {
		out, err := svc.ListAuthorizers(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, a := range out.Authorizers {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(a.AuthorizerName), StringValue(a.AuthorizerName),
				"aws_iot_authorizer", "aws", iotAllowEmptyValues))
		}
		if out.NextMarker == nil {
			return nil
		}
		input.Marker = out.NextMarker
	}
}

func (g *IotGenerator) loadBillingGroups(svc *iot.Client) error {
	input := &iot.ListBillingGroupsInput{}
	for {
		out, err := svc.ListBillingGroups(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, b := range out.BillingGroups {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(b.GroupName), StringValue(b.GroupName),
				"aws_iot_billing_group", "aws", iotAllowEmptyValues))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *IotGenerator) loadCACertificates(svc *iot.Client) error {
	input := &iot.ListCACertificatesInput{}
	for {
		out, err := svc.ListCACertificates(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, c := range out.Certificates {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(c.CertificateId), StringValue(c.CertificateId),
				"aws_iot_ca_certificate", "aws", iotAllowEmptyValues))
		}
		if out.NextMarker == nil {
			return nil
		}
		input.Marker = out.NextMarker
	}
}

func (g *IotGenerator) loadCertificatesIoT(svc *iot.Client) error {
	input := &iot.ListCertificatesInput{}
	for {
		out, err := svc.ListCertificates(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, c := range out.Certificates {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(c.CertificateId), StringValue(c.CertificateId),
				"aws_iot_certificate", "aws", iotAllowEmptyValues))
		}
		if out.NextMarker == nil {
			return nil
		}
		input.Marker = out.NextMarker
	}
}

func (g *IotGenerator) loadDomainConfigurations(svc *iot.Client) error {
	input := &iot.ListDomainConfigurationsInput{}
	for {
		out, err := svc.ListDomainConfigurations(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, d := range out.DomainConfigurations {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(d.DomainConfigurationName), StringValue(d.DomainConfigurationName),
				"aws_iot_domain_configuration", "aws", iotAllowEmptyValues))
		}
		if out.NextMarker == nil {
			return nil
		}
		input.Marker = out.NextMarker
	}
}

func (g *IotGenerator) loadPoliciesIoT(svc *iot.Client) error {
	input := &iot.ListPoliciesInput{}
	for {
		out, err := svc.ListPolicies(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, p := range out.Policies {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(p.PolicyName), StringValue(p.PolicyName),
				"aws_iot_policy", "aws", iotAllowEmptyValues))
		}
		if out.NextMarker == nil {
			return nil
		}
		input.Marker = out.NextMarker
	}
}

func (g *IotGenerator) loadProvisioningTemplates(svc *iot.Client) error {
	input := &iot.ListProvisioningTemplatesInput{}
	for {
		out, err := svc.ListProvisioningTemplates(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, t := range out.Templates {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(t.TemplateName), StringValue(t.TemplateName),
				"aws_iot_provisioning_template", "aws", iotAllowEmptyValues))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *IotGenerator) loadThingGroups(svc *iot.Client) error {
	input := &iot.ListThingGroupsInput{}
	for {
		out, err := svc.ListThingGroups(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, t := range out.ThingGroups {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(t.GroupName), StringValue(t.GroupName),
				"aws_iot_thing_group", "aws", iotAllowEmptyValues))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *IotGenerator) loadTopicRuleDestinations(svc *iot.Client) error {
	input := &iot.ListTopicRuleDestinationsInput{}
	for {
		out, err := svc.ListTopicRuleDestinations(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, d := range out.DestinationSummaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(d.Arn), StringValue(d.Arn),
				"aws_iot_topic_rule_destination", "aws", iotAllowEmptyValues))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}
