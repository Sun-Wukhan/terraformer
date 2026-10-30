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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/desktopvirtualization/armdesktopvirtualization"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// VirtualDesktopGenerator covers Microsoft.DesktopVirtualization: host
// pools, application groups with their applications, workspaces and
// scaling plans.
//
// Skipped in this family:
//   - azurerm_virtual_desktop_host_pool_registration_info: singleton
//     registrationInfo/default child of a host pool; the SDK exposes only
//     Get/Update, no List.
//   - azurerm_virtual_desktop_scaling_plan_host_pool_association and
//     azurerm_virtual_desktop_workspace_application_group_association:
//     composite "{planId}|{hostPoolId}" / "{workspaceId}|{appGroupId}"
//     association IDs with no ARM object and no List surface.
type VirtualDesktopGenerator struct {
	AzureService
}

func (g *VirtualDesktopGenerator) appendSimple(id, name, resourceType string) {
	if id == "" || name == "" {
		return
	}
	g.Resources = append(g.Resources, terraformutils.NewSimpleResource(id, name, resourceType, "azurerm", []string{}))
}

func (g *VirtualDesktopGenerator) InitResources() error {
	ctx := context.Background()
	subscriptionID := g.getSubscriptionID()
	cred := g.getCredential()
	opts := g.armClientOptions()
	rg := g.getResourceGroup()

	// Host pools
	hpClient, err := armdesktopvirtualization.NewHostPoolsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	if rg != "" {
		pager := hpClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_virtual_desktop_host_pool")
			}
		}
	} else {
		pager := hpClient.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_virtual_desktop_host_pool")
			}
		}
	}

	// Application groups + applications
	agClient, err := armdesktopvirtualization.NewApplicationGroupsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	appClient, err := armdesktopvirtualization.NewApplicationsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	var groups []*armdesktopvirtualization.ApplicationGroup
	if rg != "" {
		pager := agClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			groups = append(groups, page.Value...)
		}
	} else {
		pager := agClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			groups = append(groups, page.Value...)
		}
	}
	for _, group := range groups {
		g.appendSimple(*group.ID, *group.Name, "azurerm_virtual_desktop_application_group")
		groupRG := resourceGroupFromID(*group.ID)
		pager := appClient.NewListPager(groupRG, *group.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_virtual_desktop_application")
			}
		}
	}

	// Workspaces
	wsClient, err := armdesktopvirtualization.NewWorkspacesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	if rg != "" {
		pager := wsClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_virtual_desktop_workspace")
			}
		}
	} else {
		pager := wsClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_virtual_desktop_workspace")
			}
		}
	}

	// Scaling plans
	spClient, err := armdesktopvirtualization.NewScalingPlansClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	if rg != "" {
		pager := spClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_virtual_desktop_scaling_plan")
			}
		}
	} else {
		pager := spClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_virtual_desktop_scaling_plan")
			}
		}
	}
	return nil
}
