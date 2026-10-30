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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// RBACGenerator covers role assignments and definitions, role management
// policies and PIM (privileged identity management) schedules at the
// subscription scope, plus marketplace role assignments at the
// /providers/Microsoft.Marketplace scope.
//
// Several azurerm RBAC importers use composite IDs rather than the ARM ID:
//   - azurerm_role_definition: "<role definition ARM ID>|<scope>"
//   - azurerm_role_management_policy: "<role definition ARM ID>|<scope>"
//   - azurerm_pim_eligible_role_assignment / azurerm_pim_active_role_assignment:
//     "<scope>|<role definition ARM ID>|<principal ID>"
//
// Those composites are constructed from the schedule/assignment properties
// returned by the list calls.
type RBACGenerator struct {
	AzureService
}

func (g *RBACGenerator) InitResources() error {
	ctx := context.Background()
	subscriptionID := g.getSubscriptionID()
	cred := g.getCredential()
	opts := g.armClientOptions()
	scope := "/subscriptions/" + subscriptionID

	assignmentsClient, err := armauthorization.NewRoleAssignmentsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	pager := assignmentsClient.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			break
		}
		for _, item := range page.Value {
			g.appendRBACResource(item.ID, item.Name, "azurerm_role_assignment")
		}
	}

	marketplacePager := assignmentsClient.NewListForScopePager("/providers/Microsoft.Marketplace", nil)
	for marketplacePager.More() {
		page, err := marketplacePager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			break
		}
		for _, item := range page.Value {
			g.appendRBACResource(item.ID, item.Name, "azurerm_marketplace_role_assignment")
		}
	}

	definitionsClient, err := armauthorization.NewRoleDefinitionsClient(cred, opts)
	if err != nil {
		return err
	}
	defPager := definitionsClient.NewListPager(scope, nil)
	for defPager.More() {
		page, err := defPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			break
		}
		for _, item := range page.Value {
			if item.ID == nil || item.Name == nil {
				continue
			}
			g.appendRBACResource(ptrString(*item.ID+"|"+scope), item.Name, "azurerm_role_definition")
		}
	}

	policyAssignmentsClient, err := armauthorization.NewRoleManagementPolicyAssignmentsClient(cred, opts)
	if err != nil {
		return err
	}
	policyPager := policyAssignmentsClient.NewListForScopePager(scope, nil)
	for policyPager.More() {
		page, err := policyPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			break
		}
		for _, item := range page.Value {
			if item.Name == nil || item.Properties == nil || item.Properties.RoleDefinitionID == nil || item.Properties.Scope == nil {
				continue
			}
			id := *item.Properties.RoleDefinitionID + "|" + *item.Properties.Scope
			g.appendRBACResource(&id, item.Name, "azurerm_role_management_policy")
		}
	}

	eligibleClient, err := armauthorization.NewRoleEligibilitySchedulesClient(cred, opts)
	if err != nil {
		return err
	}
	eligiblePager := eligibleClient.NewListForScopePager(scope, nil)
	for eligiblePager.More() {
		page, err := eligiblePager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			break
		}
		for _, item := range page.Value {
			if item.Name == nil || item.Properties == nil || item.Properties.Scope == nil || item.Properties.RoleDefinitionID == nil || item.Properties.PrincipalID == nil {
				continue
			}
			id := *item.Properties.Scope + "|" + *item.Properties.RoleDefinitionID + "|" + *item.Properties.PrincipalID
			g.appendRBACResource(&id, item.Name, "azurerm_pim_eligible_role_assignment")
		}
	}

	activeClient, err := armauthorization.NewRoleAssignmentSchedulesClient(cred, opts)
	if err != nil {
		return err
	}
	activePager := activeClient.NewListForScopePager(scope, nil)
	for activePager.More() {
		page, err := activePager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			break
		}
		for _, item := range page.Value {
			if item.Name == nil || item.Properties == nil || item.Properties.Scope == nil || item.Properties.RoleDefinitionID == nil || item.Properties.PrincipalID == nil {
				continue
			}
			id := *item.Properties.Scope + "|" + *item.Properties.RoleDefinitionID + "|" + *item.Properties.PrincipalID
			g.appendRBACResource(&id, item.Name, "azurerm_pim_active_role_assignment")
		}
	}
	return nil
}

func (g *RBACGenerator) appendRBACResource(id *string, name *string, resourceType string) {
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

func ptrString(s string) *string {
	return &s
}
