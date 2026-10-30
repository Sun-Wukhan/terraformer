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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/automanage/armautomanage"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/hybridcompute/armhybridcompute"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// AutomanageGenerator covers Microsoft.AutoManage configuration profiles
// and their assignments on Azure VMs and Arc machines, enumerated via the
// armautomanage ConfigurationProfileAssignmentsClient per-parent pagers.
type AutomanageGenerator struct {
	AzureService
}

func (g *AutomanageGenerator) appendSimple(id, name, resourceType string) {
	if id == "" || name == "" {
		return
	}
	g.Resources = append(g.Resources, terraformutils.NewSimpleResource(id, name, resourceType, "azurerm", []string{}))
}

func (g *AutomanageGenerator) InitResources() error {
	ctx := context.Background()
	subscriptionID := g.getSubscriptionID()
	cred := g.getCredential()
	opts := g.armClientOptions()
	rg := g.getResourceGroup()

	// Configuration profiles
	profileClient, err := armautomanage.NewConfigurationProfilesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	if rg != "" {
		pager := profileClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_automanage_configuration")
			}
		}
	} else {
		pager := profileClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_automanage_configuration")
			}
		}
	}

	assignmentClient, err := armautomanage.NewConfigurationProfileAssignmentsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}

	// Assignments on Azure VMs
	vmClient, err := armcompute.NewVirtualMachinesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	var vms []*armcompute.VirtualMachine
	if rg != "" {
		pager := vmClient.NewListPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			vms = append(vms, page.Value...)
		}
	} else {
		pager := vmClient.NewListAllPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			vms = append(vms, page.Value...)
		}
	}
	for _, vm := range vms {
		vmRG := resourceGroupFromID(*vm.ID)
		pager := assignmentClient.NewListByVirtualMachinesPager(vmRG, *vm.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_virtual_machine_automanage_configuration_assignment")
			}
		}
	}

	// Assignments on Arc machines
	machineClient, err := armhybridcompute.NewMachinesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	var machines []*armhybridcompute.Machine
	if rg != "" {
		pager := machineClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			machines = append(machines, page.Value...)
		}
	} else {
		pager := machineClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			machines = append(machines, page.Value...)
		}
	}
	for _, machine := range machines {
		machineRG := resourceGroupFromID(*machine.ID)
		pager := assignmentClient.NewListByMachineNamePager(machineRG, *machine.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_arc_machine_automanage_configuration_assignment")
			}
		}
	}
	return nil
}
