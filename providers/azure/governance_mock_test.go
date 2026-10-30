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
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/consumption/armconsumption"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/managementgroups/armmanagementgroups"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/msi/armmsi"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armpolicy"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/workloads/armworkloads"
)

const mockPolicyAssignmentsJSON = `{"value":[
	{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/providers/Microsoft.Authorization/policyAssignments/pa-mock-1","name":"pa-mock-1","type":"Microsoft.Authorization/policyAssignments","properties":{"policyDefinitionId":"/subscriptions/00000000-0000-0000-0000-000000000000/providers/Microsoft.Authorization/policyDefinitions/pd-1"}}
]}`

func TestMockTrack2PolicyAssignments(t *testing.T) {
	tr := &cannedTransport{body: mockPolicyAssignmentsJSON}
	client, err := armpolicy.NewAssignmentsClient(mockSubscriptionID, stubCredential{}, mockClientOptions(tr))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	g := &PolicyGenerator{}
	pager := client.NewListPager(nil)
	ctx := context.Background()
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			t.Fatalf("next page: %v", err)
		}
		for _, item := range page.Value {
			g.appendPolicyResource(item.ID, item.Name, "azurerm_subscription_policy_assignment")
		}
	}
	if len(g.Resources) != 1 {
		t.Fatalf("got %d resources, want 1", len(g.Resources))
	}
	if g.Resources[0].ResourceName != "tfer--pa-mock-1" {
		t.Errorf("name = %q, want tfer--pa-mock-1", g.Resources[0].ResourceName)
	}
	if g.Resources[0].InstanceInfo.Type != "azurerm_subscription_policy_assignment" {
		t.Errorf("type = %q, want azurerm_subscription_policy_assignment", g.Resources[0].InstanceInfo.Type)
	}
	if len(tr.urls) == 0 || !strings.Contains(tr.urls[0], "policyAssignments") {
		t.Errorf("expected an ARM policyAssignments call, got %v", tr.urls)
	}
}

const mockManagementGroupsJSON = `{"value":[
	{"id":"/providers/Microsoft.Management/managementGroups/mg-mock-1","name":"mg-mock-1","type":"Microsoft.Management/managementGroups","properties":{"displayName":"Mock MG"}}
]}`

func TestMockTrack2ManagementGroups(t *testing.T) {
	tr := &cannedTransport{body: mockManagementGroupsJSON}
	client, err := armmanagementgroups.NewClient(stubCredential{}, mockClientOptions(tr))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	pager := client.NewListPager(nil)
	ctx := context.Background()
	var names []string
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			t.Fatalf("next page: %v", err)
		}
		for _, item := range page.Value {
			names = append(names, *item.Name)
			if *item.ID != "/providers/Microsoft.Management/managementGroups/mg-mock-1" {
				t.Errorf("id = %q, want the management group ARM ID", *item.ID)
			}
		}
	}
	if len(names) != 1 || names[0] != "mg-mock-1" {
		t.Fatalf("got %v, want [mg-mock-1]", names)
	}
	if len(tr.urls) == 0 || !strings.Contains(tr.urls[0], "managementGroups") {
		t.Errorf("expected an ARM managementGroups call, got %v", tr.urls)
	}
}

const mockDeploymentsJSON = `{"value":[
	{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/providers/Microsoft.Resources/deployments/dep-mock-1","name":"dep-mock-1","type":"Microsoft.Resources/deployments","properties":{"provisioningState":"Succeeded"}}
]}`

func TestMockTrack2SubscriptionDeployments(t *testing.T) {
	tr := &cannedTransport{body: mockDeploymentsJSON}
	client, err := armresources.NewDeploymentsClient(mockSubscriptionID, stubCredential{}, mockClientOptions(tr))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	g := &TemplateDeploymentGenerator{}
	appendDeploymentPager(context.Background(), g, client.NewListAtSubscriptionScopePager(nil),
		func(page armresources.DeploymentsClientListAtSubscriptionScopeResponse) []*armresources.DeploymentExtended {
			return page.Value
		},
		"azurerm_subscription_template_deployment")
	if len(g.Resources) != 1 {
		t.Fatalf("got %d resources, want 1", len(g.Resources))
	}
	if g.Resources[0].InstanceInfo.Type != "azurerm_subscription_template_deployment" {
		t.Errorf("type = %q, want azurerm_subscription_template_deployment", g.Resources[0].InstanceInfo.Type)
	}
	if len(tr.urls) == 0 || !strings.Contains(tr.urls[0], "deployments") {
		t.Errorf("expected an ARM deployments call, got %v", tr.urls)
	}
}

const mockBudgetsJSON = `{"value":[
	{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/providers/Microsoft.Consumption/budgets/budget-mock-1","name":"budget-mock-1","type":"Microsoft.Consumption/budgets","properties":{"category":"Cost","amount":100,"timeGrain":"Monthly","timePeriod":{"startDate":"2026-01-01T00:00:00Z"}}}
]}`

func TestMockTrack2ConsumptionBudgets(t *testing.T) {
	tr := &cannedTransport{body: mockBudgetsJSON}
	client, err := armconsumption.NewBudgetsClient(stubCredential{}, mockClientOptions(tr))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	g := &CostManagementGenerator{}
	g.appendBudgets(context.Background(), client, "/subscriptions/"+mockSubscriptionID, "azurerm_consumption_budget_subscription")
	if len(g.Resources) != 1 {
		t.Fatalf("got %d resources, want 1", len(g.Resources))
	}
	if g.Resources[0].ResourceName != "tfer--budget-mock-1" {
		t.Errorf("name = %q, want tfer--budget-mock-1", g.Resources[0].ResourceName)
	}
	if len(tr.urls) == 0 || !strings.Contains(tr.urls[0], "budgets") {
		t.Errorf("expected an ARM budgets call, got %v", tr.urls)
	}
}

const mockIdentitiesJSON = `{"value":[
	{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.ManagedIdentity/userAssignedIdentities/id-mock-1","name":"id-mock-1","location":"eastus","type":"Microsoft.ManagedIdentity/userAssignedIdentities","properties":{"clientId":"11111111-1111-1111-1111-111111111111","principalId":"22222222-2222-2222-2222-222222222222","tenantId":"33333333-3333-3333-3333-333333333333"}}
]}`

func TestMockTrack2UserAssignedIdentities(t *testing.T) {
	tr := &cannedTransport{body: mockIdentitiesJSON}
	client, err := armmsi.NewUserAssignedIdentitiesClient(mockSubscriptionID, stubCredential{}, mockClientOptions(tr))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	g := &PlatformMiscGenerator{}
	pager := client.NewListBySubscriptionPager(nil)
	ctx := context.Background()
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			t.Fatalf("next page: %v", err)
		}
		for _, item := range page.Value {
			g.add(item.ID, item.Name, "azurerm_user_assigned_identity")
		}
	}
	if len(g.Resources) != 1 {
		t.Fatalf("got %d resources, want 1", len(g.Resources))
	}
	if g.Resources[0].InstanceInfo.Type != "azurerm_user_assigned_identity" {
		t.Errorf("type = %q, want azurerm_user_assigned_identity", g.Resources[0].InstanceInfo.Type)
	}
	if len(tr.urls) == 0 || !strings.Contains(tr.urls[0], "userAssignedIdentities") {
		t.Errorf("expected an ARM userAssignedIdentities call, got %v", tr.urls)
	}
}

func TestSAPVirtualInstanceType(t *testing.T) {
	cases := []struct {
		name  string
		props *armworkloads.SAPVirtualInstanceProperties
		want  string
	}{
		{"nil properties", nil, ""},
		{"discovery", &armworkloads.SAPVirtualInstanceProperties{
			Configuration: &armworkloads.DiscoveryConfiguration{},
		}, "azurerm_workloads_sap_discovery_virtual_instance"},
		{"single server", &armworkloads.SAPVirtualInstanceProperties{
			Configuration: &armworkloads.DeploymentConfiguration{
				InfrastructureConfiguration: &armworkloads.SingleServerConfiguration{},
			},
		}, "azurerm_workloads_sap_single_node_virtual_instance"},
		{"three tier", &armworkloads.SAPVirtualInstanceProperties{
			Configuration: &armworkloads.DeploymentWithOSConfiguration{
				InfrastructureConfiguration: &armworkloads.ThreeTierConfiguration{},
			},
		}, "azurerm_workloads_sap_three_tier_virtual_instance"},
	}
	for _, c := range cases {
		if got := sapVirtualInstanceType(c.props); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}
