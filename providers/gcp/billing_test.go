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
	"testing"

	billingbudgetsv1 "google.golang.org/api/billingbudgets/v1"
	cloudbillingv1 "google.golang.org/api/cloudbilling/v1"
	"google.golang.org/api/option"
)

// TestBillingGenerator asserts the billing account IAM trio, budget,
// subaccount and project billing info IDs from canned responses.
func TestBillingGenerator(t *testing.T) {
	ctx := context.Background()
	ba := "billingAccounts/01AB23-CD45EF"
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: ":getIamPolicy", body: `{"bindings":[{"role":"roles/billing.admin","members":["user:jane@example.com"]}]}`},
		{substr: "/billingInfo", body: `{"name":"projects/test-project/billingInfo","billingAccountName":"` + ba + `","billingEnabled":true}`},
		{substr: "/budgets", body: `{"budgets":[{"name":"` + ba + `/budgets/budget1"}]}`},
		{substr: "subAccounts", body: `{"billingAccounts":[{"name":"billingAccounts/02XY34-ZW56AB"}]}`},
		{substr: "/v1/billingAccounts", body: `{}`},
	}}
	opts := []option.ClientOption{option.WithHTTPClient(transport.client()), option.WithoutAuthentication()}
	cbSvc, err := cloudbillingv1.NewService(ctx, opts...)
	if err != nil {
		t.Fatal(err)
	}
	bbSvc, err := billingbudgetsv1.NewService(ctx, opts...)
	if err != nil {
		t.Fatal(err)
	}
	g := &BillingGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project"})
	if err := g.initResources(ctx, cbSvc, bbSvc); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_billing_project_info":        {"projects/test-project"},
		"google_billing_account_iam_policy":  {ba},
		"google_billing_account_iam_binding": {ba},
		"google_billing_account_iam_member":  {ba},
		"google_billing_budget":              {ba + "/budgets/budget1"},
		"google_billing_subaccount":          {"02XY34-ZW56AB"},
	})
}
