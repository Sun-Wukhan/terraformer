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
	"github.com/aws/aws-sdk-go-v2/service/bcmdataexports"
	"github.com/aws/aws-sdk-go-v2/service/billing"
	"github.com/aws/aws-sdk-go-v2/service/pricingplanmanager"
	"github.com/aws/aws-sdk-go-v2/service/savingsplans"
)

var billingTailAllowEmptyValues = []string{"tags."}

// Skipped in the billing family:
//   - aws_cur_report_definition: the Cost and Usage Report service has no
//     aws-sdk-go-v2 client at all (DescribeReportDefinitions exists only in
//     SDK v1 / other language SDKs), so report definitions cannot be enumerated
//     with the SDK generation this codebase uses.
//   - aws_computeoptimizer_enrollment_status / aws_computeoptimizer_recommendation_preferences:
//     singleton account-level enrollment/preferences objects; provider ID is the
//     account ID (or a scope tuple), with no list API.
//   - aws_costoptimizationhub_enrollment_status / aws_costoptimizationhub_preferences:
//     singleton account-level enrollment/preferences objects; provider ID is the
//     account ID, with no list API.
//   - aws_uxc_account_customizations: singleton account-level customization object;
//     provider ID is the account ID and the uxc SDK exposes only
//     GetAccountCustomizations for the caller's account.

// BcmDataExportsGenerator covers BCM Data Exports. Import ID = export ARN.
type BcmDataExportsGenerator struct {
	AWSService
}

func (g *BcmDataExportsGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := bcmdataexports.NewFromConfig(config)
	p := bcmdataexports.NewListExportsPaginator(svc, &bcmdataexports.ListExportsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, exp := range page.Exports {
			arn := StringValue(exp.ExportArn)
			name := StringValue(exp.ExportName)
			if name == "" {
				name = arn
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, name, "aws_bcmdataexports_export", "aws", billingTailAllowEmptyValues))
		}
	}
	return nil
}

// BillingGenerator covers AWS Billing views. Import ID = billing view ARN.
type BillingGenerator struct {
	AWSService
}

func (g *BillingGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := billing.NewFromConfig(config)
	p := billing.NewListBillingViewsPaginator(svc, &billing.ListBillingViewsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, view := range page.BillingViews {
			arn := StringValue(view.Arn)
			name := StringValue(view.Name)
			if name == "" {
				name = arn
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, name, "aws_billing_view", "aws", billingTailAllowEmptyValues))
		}
	}
	return nil
}

// PricingPlanManagerGenerator covers Pricing Plan Manager subscriptions.
// Import ID = subscription ARN.
type PricingPlanManagerGenerator struct {
	AWSService
}

func (g *PricingPlanManagerGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := pricingplanmanager.NewFromConfig(config)
	p := pricingplanmanager.NewListSubscriptionsPaginator(svc, &pricingplanmanager.ListSubscriptionsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, sub := range page.SubscriptionSummaries {
			arn := StringValue(sub.Arn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_pricingplanmanager_subscription", "aws", billingTailAllowEmptyValues))
		}
	}
	return nil
}

// SavingsPlansGenerator covers Savings Plans. Import ID = savings plan ID.
type SavingsPlansGenerator struct {
	AWSService
}

func (g *SavingsPlansGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := savingsplans.NewFromConfig(config)
	out, err := svc.DescribeSavingsPlans(context.TODO(), &savingsplans.DescribeSavingsPlansInput{})
	if err != nil {
		return err
	}
	for _, plan := range out.SavingsPlans {
		id := StringValue(plan.SavingsPlanId)
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			id, id, "aws_savingsplans_savings_plan", "aws", billingTailAllowEmptyValues))
	}
	return nil
}
