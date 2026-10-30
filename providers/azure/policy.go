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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/policyinsights/armpolicyinsights"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armpolicy"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// PolicyGenerator covers Azure Policy definitions, assignments and
// remediations at subscription, resource-group and management-group scopes.
//
// Skipped in this family:
//   - azurerm_resource_policy_assignment: assignments at an individual
//     resource scope can only be listed per target resource
//     (AssignmentsClient.NewListForResourcePager needs the resource's group,
//     provider, type and name), so they cannot be enumerated generically.
//   - azurerm_subscription_policy_exemption, azurerm_resource_group_policy_exemption,
//     azurerm_management_group_policy_exemption, azurerm_resource_policy_exemption:
//     the track-2 armpolicy module (v1.0.0, latest) has no PolicyExemptions
//     client at all.
//   - azurerm_resource_policy_remediation: resource-scope remediations have
//     the same per-resource enumeration problem as resource assignments.
type PolicyGenerator struct {
	AzureService
}

func (g *PolicyGenerator) InitResources() error {
	ctx := context.Background()
	subscriptionID := g.getSubscriptionID()
	cred := g.getCredential()
	opts := g.armClientOptions()

	definitionsClient, err := armpolicy.NewDefinitionsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	pager := definitionsClient.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			break
		}
		for _, item := range page.Value {
			g.appendPolicyResource(item.ID, item.Name, "azurerm_policy_definition")
		}
	}

	setDefinitionsClient, err := armpolicy.NewSetDefinitionsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	setPager := setDefinitionsClient.NewListPager(nil)
	for setPager.More() {
		page, err := setPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			break
		}
		for _, item := range page.Value {
			g.appendPolicyResource(item.ID, item.Name, "azurerm_policy_set_definition")
		}
	}

	assignmentsClient, err := armpolicy.NewAssignmentsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	assignPager := assignmentsClient.NewListPager(nil)
	for assignPager.More() {
		page, err := assignPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			break
		}
		for _, item := range page.Value {
			g.appendPolicyResource(item.ID, item.Name, "azurerm_subscription_policy_assignment")
		}
	}

	resourceGroups, err := governanceResourceGroupNames(ctx, &g.AzureService)
	if err != nil {
		return err
	}
	for _, rg := range resourceGroups {
		rgPager := assignmentsClient.NewListForResourceGroupPager(rg, nil)
		for rgPager.More() {
			page, err := rgPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				g.appendPolicyResource(item.ID, item.Name, "azurerm_resource_group_policy_assignment")
			}
		}
	}

	remediationsClient, err := armpolicyinsights.NewRemediationsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	remPager := remediationsClient.NewListForSubscriptionPager(nil)
	for remPager.More() {
		page, err := remPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			break
		}
		for _, item := range page.Value {
			g.appendPolicyResource(item.ID, item.Name, "azurerm_subscription_policy_remediation")
		}
	}
	for _, rg := range resourceGroups {
		rgPager := remediationsClient.NewListForResourceGroupPager(rg, nil)
		for rgPager.More() {
			page, err := rgPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				g.appendPolicyResource(item.ID, item.Name, "azurerm_resource_group_policy_remediation")
			}
		}
	}

	groups, err := governanceManagementGroups(ctx, &g.AzureService)
	if err != nil {
		return err
	}
	for _, group := range groups {
		mgAssignPager := assignmentsClient.NewListForManagementGroupPager(group.Name, nil)
		for mgAssignPager.More() {
			page, err := mgAssignPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				g.appendPolicyResource(item.ID, item.Name, "azurerm_management_group_policy_assignment")
			}
		}
		mgSetPager := setDefinitionsClient.NewListByManagementGroupPager(group.Name, nil)
		for mgSetPager.More() {
			page, err := mgSetPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				g.appendPolicyResource(item.ID, item.Name, "azurerm_management_group_policy_set_definition")
			}
		}
		mgRemPager := remediationsClient.NewListForManagementGroupPager(group.Name, nil)
		for mgRemPager.More() {
			page, err := mgRemPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				g.appendPolicyResource(item.ID, item.Name, "azurerm_management_group_policy_remediation")
			}
		}
	}
	return nil
}

func (g *PolicyGenerator) appendPolicyResource(id *string, name *string, resourceType string) {
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
