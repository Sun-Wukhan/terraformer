// Copyright 2021 The Terraformer Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package azure

import (
	"context"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/synapse/armsynapse"
	"github.com/Azure/azure-sdk-for-go/services/preview/synapse/2019-06-01-preview/managedvirtualnetwork"
	// "github.com/Azure/azure-sdk-for-go/services/preview/synapse/2020-08-01-preview/accesscontrol"
)

type SynapseGenerator struct {
	AzureService
}

func (az *SynapseGenerator) listWorkspaces() ([]armsynapse.Workspace, error) {
	client, err := armsynapse.NewWorkspacesClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	var resources []armsynapse.Workspace
	if rg := az.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return resources, err
			}
			for _, workspace := range page.Value {
				resources = append(resources, *workspace)
			}
		}
	} else {
		pager := client.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return resources, err
			}
			for _, workspace := range page.Value {
				resources = append(resources, *workspace)
			}
		}
	}
	return resources, nil
}

func (az *SynapseGenerator) appendWorkspace(workspace *armsynapse.Workspace) {
	az.AppendSimpleResource(*workspace.ID, *workspace.Name, "azurerm_synapse_workspace")
}

func (az *SynapseGenerator) appendSQLPools(workspace *armsynapse.Workspace, workspaceRg *ResourceID) error {
	client, err := armsynapse.NewSQLPoolsClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return err
	}
	ctx := context.Background()
	pager := client.NewListByWorkspacePager(workspaceRg.ResourceGroup, *workspace.Name, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			az.AppendSimpleResource(*item.ID, *item.Name, "azurerm_synapse_sql_pool")
		}
	}
	return nil
}

func (az *SynapseGenerator) appendSparkPools(workspace *armsynapse.Workspace, workspaceRg *ResourceID) error {
	client, err := armsynapse.NewBigDataPoolsClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return err
	}
	ctx := context.Background()
	pager := client.NewListByWorkspacePager(workspaceRg.ResourceGroup, *workspace.Name, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			az.AppendSimpleResource(*item.ID, *item.Name, "azurerm_synapse_spark_pool")
		}
	}
	return nil
}

func (az *SynapseGenerator) appendFirewallRule(workspace *armsynapse.Workspace, workspaceRg *ResourceID) error {
	client, err := armsynapse.NewIPFirewallRulesClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return err
	}
	ctx := context.Background()
	pager := client.NewListByWorkspacePager(workspaceRg.ResourceGroup, *workspace.Name, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			az.AppendSimpleResource(*item.ID, *item.Name, "azurerm_synapse_firewall_rule")
		}
	}
	return nil
}

func (az *SynapseGenerator) appendManagedPrivateEndpoint(workspace *armsynapse.Workspace) error {
	// armsynapse@v0.8.0 has no ManagedPrivateEndpointsClient; keep this one
	// listing on track-1 (the workspace check uses the track-2 Properties).
	if workspace.Properties == nil || workspace.Properties.ManagedVirtualNetwork == nil {
		return nil
	}
	virtualNetworkName := *workspace.Properties.ManagedVirtualNetwork
	if virtualNetworkName == "" || virtualNetworkName == "default" {
		return nil
	}
	subscriptionID, _, authorizer, _ := az.getClientArgs()
	// ManagedPrivateEndpointsClient does not have a ...WithBaseURI function, why is this different?
	client := managedvirtualnetwork.NewManagedPrivateEndpointsClient(subscriptionID)
	client.Authorizer = authorizer
	ctx := context.Background()
	iterator, err := client.ListComplete(ctx, virtualNetworkName)
	if err != nil {
		return err
	}
	for iterator.NotDone() {
		item := iterator.Value()
		az.AppendSimpleResource(*item.ID, *item.Name, "azurerm_synapse_managed_private_endpoint")
		if err := iterator.NextWithContext(ctx); err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}

func (az *SynapseGenerator) listPrivateLinkHubs() ([]armsynapse.PrivateLinkHub, error) {
	client, err := armsynapse.NewPrivateLinkHubsClient(az.getSubscriptionID(), az.getCredential(), az.armClientOptions())
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	var resources []armsynapse.PrivateLinkHub
	if rg := az.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return resources, err
			}
			for _, hub := range page.Value {
				resources = append(resources, *hub)
			}
		}
	} else {
		pager := client.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return resources, err
			}
			for _, hub := range page.Value {
				resources = append(resources, *hub)
			}
		}
	}
	return resources, nil
}

func (az *SynapseGenerator) appendtPrivateLinkHubs(workspace *armsynapse.PrivateLinkHub) {
	az.AppendSimpleResource(*workspace.ID, *workspace.Name, "azurerm_synapse_private_link_hub")
}

func (az *SynapseGenerator) InitResources() error {

	workspaces, err := az.listWorkspaces()
	if err != nil {
		return err
	}
	for _, workspace := range workspaces {
		az.appendWorkspace(&workspace)
		workspaceRg, err := ParseAzureResourceID(*workspace.ID)
		if err != nil {
			return err
		}
		err = az.appendSQLPools(&workspace, workspaceRg)
		if err != nil {
			return err
		}
		err = az.appendSparkPools(&workspace, workspaceRg)
		if err != nil {
			return err
		}
		err = az.appendFirewallRule(&workspace, workspaceRg)
		if err != nil {
			return err
		}
		err = az.appendManagedPrivateEndpoint(&workspace)
		if err != nil {
			return err
		}
	}

	hubs, err := az.listPrivateLinkHubs()
	if err == nil {
		for _, hub := range hubs {
			az.appendtPrivateLinkHubs(&hub)
		}
	}
	return nil
}
