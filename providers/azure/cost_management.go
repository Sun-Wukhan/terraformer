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

package azure

import (
	"context"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/consumption/armconsumption"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/costmanagement/armcostmanagement"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// CostManagementGenerator covers consumption budgets and Cost Management
// exports and views at subscription, resource-group and management-group
// scopes.
//
// Skipped in this family:
//   - azurerm_billing_account_cost_management_export: its scope is a billing
//     account (/providers/Microsoft.Billing/billingAccounts/<id>), which
//     cannot be derived from the subscription terraformer operates in.
//   - azurerm_cost_management_scheduled_action and azurerm_cost_anomaly_alert:
//     both are backed by the Microsoft.CostManagement/scheduledActions API,
//     and the track-2 armcostmanagement module (v1.1.1) has no
//     ScheduledActions client.
type CostManagementGenerator struct {
	AzureService
}

func (g *CostManagementGenerator) InitResources() error {
	ctx := context.Background()
	subscriptionID := g.getSubscriptionID()
	cred := g.getCredential()
	opts := g.armClientOptions()
	subscriptionScope := "/subscriptions/" + subscriptionID

	resourceGroups, err := governanceResourceGroupNames(ctx, &g.AzureService)
	if err != nil {
		return err
	}
	groups, err := governanceManagementGroups(ctx, &g.AzureService)
	if err != nil {
		return err
	}

	budgetsClient, err := armconsumption.NewBudgetsClient(cred, opts)
	if err != nil {
		return err
	}
	g.appendBudgets(ctx, budgetsClient, subscriptionScope, "azurerm_consumption_budget_subscription")
	for _, rg := range resourceGroups {
		g.appendBudgets(ctx, budgetsClient, subscriptionScope+"/resourceGroups/"+rg, "azurerm_consumption_budget_resource_group")
	}
	for _, group := range groups {
		g.appendBudgets(ctx, budgetsClient, group.ID, "azurerm_consumption_budget_management_group")
	}

	exportsClient, err := armcostmanagement.NewExportsClient(cred, opts)
	if err != nil {
		return err
	}
	g.appendExports(ctx, exportsClient, subscriptionScope, "azurerm_subscription_cost_management_export")
	for _, rg := range resourceGroups {
		g.appendExports(ctx, exportsClient, subscriptionScope+"/resourceGroups/"+rg, "azurerm_resource_group_cost_management_export")
	}

	viewsClient, err := armcostmanagement.NewViewsClient(cred, opts)
	if err != nil {
		return err
	}
	g.appendViews(ctx, viewsClient, subscriptionScope, "azurerm_subscription_cost_management_view")
	for _, rg := range resourceGroups {
		g.appendViews(ctx, viewsClient, subscriptionScope+"/resourceGroups/"+rg, "azurerm_resource_group_cost_management_view")
	}
	return nil
}

func (g *CostManagementGenerator) appendBudgets(ctx context.Context, client *armconsumption.BudgetsClient, scope string, resourceType string) {
	pager := client.NewListPager(scope, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return
		}
		for _, item := range page.Value {
			g.appendCostResource(item.ID, item.Name, resourceType)
		}
	}
}

func (g *CostManagementGenerator) appendExports(ctx context.Context, client *armcostmanagement.ExportsClient, scope string, resourceType string) {
	resp, err := client.List(ctx, scope, nil)
	if err != nil {
		log.Println(err)
		return
	}
	for _, item := range resp.Value {
		g.appendCostResource(item.ID, item.Name, resourceType)
	}
}

func (g *CostManagementGenerator) appendViews(ctx context.Context, client *armcostmanagement.ViewsClient, scope string, resourceType string) {
	pager := client.NewListByScopePager(scope, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return
		}
		for _, item := range page.Value {
			g.appendCostResource(item.ID, item.Name, resourceType)
		}
	}
}

func (g *CostManagementGenerator) appendCostResource(id *string, name *string, resourceType string) {
	if id == nil || name == nil {
		return
	}
	g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
		*id,
		*name,
		resourceType,
		g.ProviderName,
		[]string{}))
}
