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

package gcp

import (
	"context"
	"log"
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	billingbudgetsv1 "google.golang.org/api/billingbudgets/v1"
	cloudbillingv1 "google.golang.org/api/cloudbilling/v1"
)

var billingAllowEmptyValues = []string{""}

var billingAdditionalFields = map[string]interface{}{}

// BillingGenerator enumerates the billing estate reachable from the
// configured project: the project's billing account is resolved from
// its ProjectBillingInfo, then the billing account's IAM trio
// (google_billing_account_iam_policy/_binding/_member, provider ID
// billingAccounts/{id}), its budgets (google_billing_budget, ID
// billingAccounts/{id}/budgets/{id}), its subaccounts
// (google_billing_subaccount, ID = the billing account ID segment)
// and the project's billing link singleton
// (google_billing_project_info, ID projects/{project}).
type BillingGenerator struct {
	GCPService
}

func (g *BillingGenerator) InitResources() error {
	ctx := context.Background()
	cbSvc, err := cloudbillingv1.NewService(ctx)
	if err != nil {
		return err
	}
	bbSvc, err := billingbudgetsv1.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, cbSvc, bbSvc)
}

func (g *BillingGenerator) newBillingResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
	if attrs == nil {
		attrs = map[string]string{}
	}
	parts := strings.Split(id, "/")
	return terraformutils.NewResource(id, parts[len(parts)-1], resourceType, g.ProviderName, attrs, billingAllowEmptyValues, billingAdditionalFields)
}

func (g *BillingGenerator) initResources(ctx context.Context, cbSvc *cloudbillingv1.APIService, bbSvc *billingbudgetsv1.Service) error {
	project := g.GetArgs()["project"].(string)
	info, err := cbSvc.Projects.GetBillingInfo("projects/" + project).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		return nil
	}
	if info == nil || info.BillingAccountName == "" {
		log.Println("billing: project " + project + " has no billing account; nothing to enumerate")
		return nil
	}
	baName := info.BillingAccountName // billingAccounts/{id}
	baID := strings.TrimPrefix(baName, "billingAccounts/")
	g.Resources = append(g.Resources, g.newBillingResource("projects/"+project, "google_billing_project_info", map[string]string{"project": project, "billing_account": baID}))

	if policy, err := cbSvc.BillingAccounts.GetIamPolicy(baName).Context(ctx).Do(); err == nil {
		bindings := []iamBinding{}
		for _, b := range policy.Bindings {
			bindings = append(bindings, iamBinding{Role: b.Role, Members: b.Members})
		}
		g.Resources = append(g.Resources, expandIamPolicy(bindings, baName,
			"google_billing_account_iam_policy", "google_billing_account_iam_binding", "google_billing_account_iam_member",
			map[string]string{"billing_account_id": baID}, g.newBillingResource)...)
	} else {
		log.Println(err)
	}

	if bbSvc != nil {
		err := bbSvc.BillingAccounts.Budgets.List(baName).Pages(ctx, func(page *billingbudgetsv1.GoogleCloudBillingBudgetsV1ListBudgetsResponse) error {
			for _, budget := range page.Budgets {
				g.Resources = append(g.Resources, g.newBillingResource(budget.Name, "google_billing_budget", map[string]string{"billing_account": baID}))
			}
			return nil
		})
		if err != nil {
			log.Println(err)
		}
	}

	err = cbSvc.BillingAccounts.SubAccounts.List(baName).Pages(ctx, func(page *cloudbillingv1.ListBillingAccountsResponse) error {
		for _, sub := range page.BillingAccounts {
			subID := strings.TrimPrefix(sub.Name, "billingAccounts/")
			g.Resources = append(g.Resources, g.newBillingResource(subID, "google_billing_subaccount", map[string]string{"billing_account_id": subID, "master_billing_account": baID}))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	return nil
}
