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
	"github.com/aws/aws-sdk-go-v2/service/ses"
	sestypes "github.com/aws/aws-sdk-go-v2/service/ses/types"
)

// Depth coverage for SES v1. Import IDs verified against the Terraform AWS
// provider docs: identity-scoped resources import with the identity,
// aws_ses_event_destination with configuration_set_name/destination_name,
// aws_ses_identity_policy with identity/policy_name, aws_ses_receipt_filter
// with the filter name.
//
// Skipped in this family:
//   - aws_ses_active_receipt_rule_set: singleton pointer naming the one
//     active receipt rule set per account/region; the rule set itself is
//     already emitted as aws_ses_receipt_rule_set.
//   - aws_ses_domain_identity_verification: verification-waiter glue with no
//     configuration of its own beyond the domain identity (already emitted
//     as aws_ses_domain_identity).

func (g *SesGenerator) loadSesDepth(svc *ses.Client) error {
	if err := g.loadSesIdentityDepth(svc); err != nil {
		return err
	}
	if err := g.loadSesEventDestinations(svc); err != nil {
		return err
	}
	if err := g.loadReceiptFilters(svc); err != nil {
		return err
	}
	return nil
}

func (g *SesGenerator) loadSesIdentityDepth(svc *ses.Client) error {
	p := ses.NewListIdentitiesPaginator(svc, &ses.ListIdentitiesInput{IdentityType: "Domain"})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, identity := range page.Identities {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				identity,
				identity,
				"aws_ses_domain_dkim",
				"aws",
				sesAllowEmptyValues))

			mailFrom, err := svc.GetIdentityMailFromDomainAttributes(context.TODO(), &ses.GetIdentityMailFromDomainAttributesInput{
				Identities: []string{identity},
			})
			if err != nil {
				return err
			}
			if attrs, ok := mailFrom.MailFromDomainAttributes[identity]; ok && attrs.MailFromDomain != nil && *attrs.MailFromDomain != "" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					identity,
					identity,
					"aws_ses_domain_mail_from",
					"aws",
					sesAllowEmptyValues))
			}

			notif, err := svc.GetIdentityNotificationAttributes(context.TODO(), &ses.GetIdentityNotificationAttributesInput{
				Identities: []string{identity},
			})
			if err != nil {
				return err
			}
			if attrs, ok := notif.NotificationAttributes[identity]; ok &&
				(attrs.BounceTopic != nil || attrs.ComplaintTopic != nil || attrs.DeliveryTopic != nil) {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					identity,
					identity,
					"aws_ses_identity_notification_topic",
					"aws",
					sesAllowEmptyValues))
			}

			policies, err := svc.ListIdentityPolicies(context.TODO(), &ses.ListIdentityPoliciesInput{Identity: &identity})
			if err != nil {
				return err
			}
			for _, policyName := range policies.PolicyNames {
				id := identity + "/" + policyName
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id,
					policyName,
					"aws_ses_identity_policy",
					"aws",
					sesAllowEmptyValues))
			}
		}
	}
	return nil
}

func (g *SesGenerator) loadSesEventDestinations(svc *ses.Client) error {
	configurationSets, err := svc.ListConfigurationSets(context.TODO(), &ses.ListConfigurationSetsInput{})
	if err != nil {
		return err
	}
	for _, configurationSet := range configurationSets.ConfigurationSets {
		setName := StringValue(configurationSet.Name)
		described, err := svc.DescribeConfigurationSet(context.TODO(), &ses.DescribeConfigurationSetInput{
			ConfigurationSetName:           configurationSet.Name,
			ConfigurationSetAttributeNames: []sestypes.ConfigurationSetAttribute{sestypes.ConfigurationSetAttributeEventDestinations},
		})
		if err != nil {
			return err
		}
		for _, destination := range described.EventDestinations {
			id := setName + "/" + StringValue(destination.Name)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				StringValue(destination.Name),
				"aws_ses_event_destination",
				"aws",
				sesAllowEmptyValues))
		}
	}
	return nil
}

func (g *SesGenerator) loadReceiptFilters(svc *ses.Client) error {
	out, err := svc.ListReceiptFilters(context.TODO(), &ses.ListReceiptFiltersInput{})
	if err != nil {
		return err
	}
	for _, filter := range out.Filters {
		name := StringValue(filter.Name)
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			name,
			name,
			"aws_ses_receipt_filter",
			"aws",
			sesAllowEmptyValues))
	}
	return nil
}
