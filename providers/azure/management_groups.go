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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/managementgroups/armmanagementgroups"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// governanceResourceGroupNames returns the resource groups in scope for
// governance generators: the single resource group terraformer was pointed
// at, or every resource group in the subscription.
func governanceResourceGroupNames(ctx context.Context, g *AzureService) ([]string, error) {
	if rg := g.getResourceGroup(); rg != "" {
		return []string{rg}, nil
	}
	client, err := armresources.NewResourceGroupsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	var names []string
	pager := client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return names, err
		}
		for _, group := range page.Value {
			if group.Name != nil {
				names = append(names, *group.Name)
			}
		}
	}
	return names, nil
}

// governanceManagementGroup is one enumerated management group: ID is the
// full ARM ID (/providers/Microsoft.Management/managementGroups/<name>) and
// Name is the management group name used by the MG-scoped list calls.
type governanceManagementGroup struct {
	ID   string
	Name string
}

func governanceManagementGroups(ctx context.Context, g *AzureService) ([]governanceManagementGroup, error) {
	client, err := armmanagementgroups.NewClient(g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	var groups []governanceManagementGroup
	pager := client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return groups, err
		}
		for _, group := range page.Value {
			if group.ID == nil || group.Name == nil {
				continue
			}
			groups = append(groups, governanceManagementGroup{ID: *group.ID, Name: *group.Name})
		}
	}
	return groups, nil
}

type ManagementGroupGenerator struct {
	AzureService
}

// InitResources enumerates management groups and, for each one, the
// subscriptions associated with it. The association importer ID is
// /providers/Microsoft.Management/managementGroups/<mg>/subscriptions/<sub>,
// which is exactly the ID shape constructed here.
func (g *ManagementGroupGenerator) InitResources() error {
	ctx := context.Background()
	groups, err := governanceManagementGroups(ctx, &g.AzureService)
	if err != nil {
		return err
	}
	for _, group := range groups {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			group.ID,
			group.Name,
			"azurerm_management_group",
			g.ProviderName,
			[]string{}))
	}

	subsClient, err := armmanagementgroups.NewManagementGroupSubscriptionsClient(g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	for _, group := range groups {
		pager := subsClient.NewGetSubscriptionsUnderManagementGroupPager(group.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, sub := range page.Value {
				if sub.ID == nil || sub.Name == nil {
					continue
				}
				// sub.ID is already the association importer ID:
				// /providers/Microsoft.Management/managementGroups/<mg>/subscriptions/<sub>
				// and sub.Name is the subscription ID.
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					*sub.ID,
					*sub.Name,
					"azurerm_management_group_subscription_association",
					g.ProviderName,
					[]string{}))
			}
		}
	}
	return nil
}
