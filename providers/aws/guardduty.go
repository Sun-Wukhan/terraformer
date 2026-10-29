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

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/guardduty"
)

var guarddutyAllowEmptyValues = []string{"tags."}

type GuardDutyGenerator struct {
	AWSService
}

// InitResources enumerates GuardDuty detectors in the configured
// region and, per detector, its filters, IP sets, threat intel sets,
// publishing destinations and members, plus account-level malware
// protection plans.
//
// Provider importer ID formats (terraform-provider-aws docs):
//   - aws_guardduty_detector: the detector ID.
//   - aws_guardduty_filter: detector_id:filter_name (verified).
//   - aws_guardduty_ipset / aws_guardduty_threatintelset /
//     aws_guardduty_publishing_destination / aws_guardduty_member:
//     detector_id:set_or_account_id, the same colon-separated pattern.
//   - aws_guardduty_malware_protection_plan: the plan ID.
//
// Intentionally not generated:
//   - aws_guardduty_detector_feature has no provider importer (its
//     docs carry no Import section).
//   - aws_guardduty_invite_accepter is the accepter side of a member
//     invitation; it belongs to the invited account, not to an object
//     enumerable in the inviting account.
//   - aws_guardduty_organization_admin_account is the singleton
//     delegated-administrator designation for the organization.
//   - aws_guardduty_organization_configuration and
//     aws_guardduty_organization_configuration_feature are the
//     singleton organization configuration of the administrator
//     account (DescribeOrganizationConfiguration).
//   - aws_guardduty_member_detector_feature is per-member feature
//     configuration glue under a member account, not an independent
//     enumerable object.
func (g *GuardDutyGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := guardduty.NewFromConfig(config)

	detectors, err := svc.ListDetectors(context.TODO(), &guardduty.ListDetectorsInput{})
	if err != nil {
		return err
	}
	for _, detectorID := range detectors.DetectorIds {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			detectorID,
			detectorID,
			"aws_guardduty_detector",
			"aws",
			guarddutyAllowEmptyValues))
		g.addFilters(svc, detectorID)
		g.addIPSets(svc, detectorID)
		g.addThreatIntelSets(svc, detectorID)
		g.addPublishingDestinations(svc, detectorID)
		g.addMembers(svc, detectorID)
	}
	g.addMalwareProtectionPlans(svc)
	return nil
}

func (g *GuardDutyGenerator) addFilters(svc *guardduty.Client, detectorID string) {
	out, err := svc.ListFilters(context.TODO(), &guardduty.ListFiltersInput{DetectorId: aws.String(detectorID)})
	if err != nil {
		log.Println(err)
		return
	}
	for _, name := range out.FilterNames {
		g.Resources = append(g.Resources, terraformutils.NewResource(
			detectorID+":"+name,
			name,
			"aws_guardduty_filter",
			"aws",
			map[string]string{"detector_id": detectorID},
			guarddutyAllowEmptyValues,
			map[string]interface{}{}))
	}
}

func (g *GuardDutyGenerator) addIPSets(svc *guardduty.Client, detectorID string) {
	out, err := svc.ListIPSets(context.TODO(), &guardduty.ListIPSetsInput{DetectorId: aws.String(detectorID)})
	if err != nil {
		log.Println(err)
		return
	}
	for _, id := range out.IpSetIds {
		g.Resources = append(g.Resources, terraformutils.NewResource(
			detectorID+":"+id,
			id,
			"aws_guardduty_ipset",
			"aws",
			map[string]string{"detector_id": detectorID},
			guarddutyAllowEmptyValues,
			map[string]interface{}{}))
	}
}

func (g *GuardDutyGenerator) addThreatIntelSets(svc *guardduty.Client, detectorID string) {
	out, err := svc.ListThreatIntelSets(context.TODO(), &guardduty.ListThreatIntelSetsInput{DetectorId: aws.String(detectorID)})
	if err != nil {
		log.Println(err)
		return
	}
	for _, id := range out.ThreatIntelSetIds {
		g.Resources = append(g.Resources, terraformutils.NewResource(
			detectorID+":"+id,
			id,
			"aws_guardduty_threatintelset",
			"aws",
			map[string]string{"detector_id": detectorID},
			guarddutyAllowEmptyValues,
			map[string]interface{}{}))
	}
}

func (g *GuardDutyGenerator) addPublishingDestinations(svc *guardduty.Client, detectorID string) {
	out, err := svc.ListPublishingDestinations(context.TODO(), &guardduty.ListPublishingDestinationsInput{DetectorId: aws.String(detectorID)})
	if err != nil {
		log.Println(err)
		return
	}
	for _, dest := range out.Destinations {
		id := StringValue(dest.DestinationId)
		g.Resources = append(g.Resources, terraformutils.NewResource(
			detectorID+":"+id,
			id,
			"aws_guardduty_publishing_destination",
			"aws",
			map[string]string{"detector_id": detectorID},
			guarddutyAllowEmptyValues,
			map[string]interface{}{}))
	}
}

func (g *GuardDutyGenerator) addMembers(svc *guardduty.Client, detectorID string) {
	p := guardduty.NewListMembersPaginator(svc, &guardduty.ListMembersInput{DetectorId: aws.String(detectorID)})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, member := range page.Members {
			accountID := StringValue(member.AccountId)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				detectorID+":"+accountID,
				accountID,
				"aws_guardduty_member",
				"aws",
				map[string]string{"detector_id": detectorID},
				guarddutyAllowEmptyValues,
				map[string]interface{}{}))
		}
	}
}

func (g *GuardDutyGenerator) addMalwareProtectionPlans(svc *guardduty.Client) {
	input := &guardduty.ListMalwareProtectionPlansInput{}
	for {
		page, err := svc.ListMalwareProtectionPlans(context.TODO(), input)
		if err != nil {
			log.Println(err)
			return
		}
		for _, plan := range page.MalwareProtectionPlans {
			id := StringValue(plan.MalwareProtectionPlanId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				id,
				"aws_guardduty_malware_protection_plan",
				"aws",
				guarddutyAllowEmptyValues))
		}
		if page.NextToken == nil {
			return
		}
		input.NextToken = page.NextToken
	}
}
