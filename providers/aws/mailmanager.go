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
	"github.com/aws/aws-sdk-go-v2/service/mailmanager"
)

var mailmanagerAllowEmptyValues = []string{"tags."}

// MailManagerGenerator covers SES Mail Manager archives, ingress points, relays,
// rule sets and traffic policies. Provider import IDs (verified): each resource's ID.
type MailManagerGenerator struct {
	AWSService
}

func (g *MailManagerGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := mailmanager.NewFromConfig(config)

	ap := mailmanager.NewListArchivesPaginator(svc, &mailmanager.ListArchivesInput{})
	for ap.HasMorePages() {
		page, err := ap.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, a := range page.Archives {
			id := StringValue(a.ArchiveId)
			name := StringValue(a.ArchiveName)
			if name == "" {
				name = id
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, name, "aws_mailmanager_archive", "aws", mailmanagerAllowEmptyValues))
		}
	}
	ip := mailmanager.NewListIngressPointsPaginator(svc, &mailmanager.ListIngressPointsInput{})
	for ip.HasMorePages() {
		page, err := ip.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, pt := range page.IngressPoints {
			id := StringValue(pt.IngressPointId)
			name := StringValue(pt.IngressPointName)
			if name == "" {
				name = id
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, name, "aws_mailmanager_ingress_point", "aws", mailmanagerAllowEmptyValues))
		}
	}
	rp := mailmanager.NewListRelaysPaginator(svc, &mailmanager.ListRelaysInput{})
	for rp.HasMorePages() {
		page, err := rp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, r := range page.Relays {
			id := StringValue(r.RelayId)
			name := StringValue(r.RelayName)
			if name == "" {
				name = id
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, name, "aws_mailmanager_relay", "aws", mailmanagerAllowEmptyValues))
		}
	}
	sp := mailmanager.NewListRuleSetsPaginator(svc, &mailmanager.ListRuleSetsInput{})
	for sp.HasMorePages() {
		page, err := sp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, rs := range page.RuleSets {
			id := StringValue(rs.RuleSetId)
			name := StringValue(rs.RuleSetName)
			if name == "" {
				name = id
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, name, "aws_mailmanager_rule_set", "aws", mailmanagerAllowEmptyValues))
		}
	}
	tp := mailmanager.NewListTrafficPoliciesPaginator(svc, &mailmanager.ListTrafficPoliciesInput{})
	for tp.HasMorePages() {
		page, err := tp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, pol := range page.TrafficPolicies {
			id := StringValue(pol.TrafficPolicyId)
			name := StringValue(pol.TrafficPolicyName)
			if name == "" {
				name = id
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, name, "aws_mailmanager_traffic_policy", "aws", mailmanagerAllowEmptyValues))
		}
	}
	return nil
}
