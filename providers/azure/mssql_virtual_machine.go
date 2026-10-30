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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sqlvirtualmachine/armsqlvirtualmachine"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// MSSQLVirtualMachineGenerator covers SQL Server on Azure VMs
// (Microsoft.SqlVirtualMachine) via armsqlvirtualmachine.
type MSSQLVirtualMachineGenerator struct {
	AzureService
}

func (g *MSSQLVirtualMachineGenerator) listResources() ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	machinesClient, err := armsqlvirtualmachine.NewSQLVirtualMachinesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	groupsClient, err := armsqlvirtualmachine.NewGroupsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	listenersClient, err := armsqlvirtualmachine.NewAvailabilityGroupListenersClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	processGroups := func(groups []*armsqlvirtualmachine.Group) error {
		for _, group := range groups {
			resources = append(resources, terraformutils.NewSimpleResource(
				*group.ID, *group.Name, "azurerm_mssql_virtual_machine_group", g.ProviderName, []string{}))
			id, err := ParseAzureResourceID(*group.ID)
			if err != nil {
				return err
			}
			pager := listenersClient.NewListByGroupPager(id.ResourceGroup, *group.Name, nil)
			for pager.More() {
				page, err := pager.NextPage(ctx)
				if err != nil {
					log.Println(err)
					return err
				}
				for _, listener := range page.Value {
					resources = append(resources, terraformutils.NewSimpleResource(
						*listener.ID, *listener.Name, "azurerm_mssql_virtual_machine_availability_group_listener", g.ProviderName, []string{}))
				}
			}
		}
		return nil
	}

	if rg := g.getResourceGroup(); rg != "" {
		machinesPager := machinesClient.NewListByResourceGroupPager(rg, nil)
		for machinesPager.More() {
			page, err := machinesPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			for _, machine := range page.Value {
				resources = append(resources, terraformutils.NewSimpleResource(
					*machine.ID, *machine.Name, "azurerm_mssql_virtual_machine", g.ProviderName, []string{}))
			}
		}
		groupsPager := groupsClient.NewListByResourceGroupPager(rg, nil)
		for groupsPager.More() {
			page, err := groupsPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			if err := processGroups(page.Value); err != nil {
				return nil, err
			}
		}
	} else {
		machinesPager := machinesClient.NewListPager(nil)
		for machinesPager.More() {
			page, err := machinesPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			for _, machine := range page.Value {
				resources = append(resources, terraformutils.NewSimpleResource(
					*machine.ID, *machine.Name, "azurerm_mssql_virtual_machine", g.ProviderName, []string{}))
			}
		}
		groupsPager := groupsClient.NewListPager(nil)
		for groupsPager.More() {
			page, err := groupsPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			if err := processGroups(page.Value); err != nil {
				return nil, err
			}
		}
	}

	return resources, nil
}

func (g *MSSQLVirtualMachineGenerator) InitResources() error {
	resources, err := g.listResources()
	if err != nil {
		return err
	}
	g.Resources = append(g.Resources, resources...)
	return nil
}
