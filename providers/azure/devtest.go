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
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/devtestlabs/armdevtestlabs"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// DevTestGenerator covers Microsoft.DevTestLab: labs with their virtual
// machines (split into the linux/windows provider types by the gallery
// image OS type), policies (under the "default" policy set), schedules
// and virtual networks, plus the subscription/RG-scoped global VM
// shutdown schedules.
type DevTestGenerator struct {
	AzureService
}

func (g *DevTestGenerator) appendSimple(id, name, resourceType string) {
	if id == "" || name == "" {
		return
	}
	g.Resources = append(g.Resources, terraformutils.NewSimpleResource(id, name, resourceType, "azurerm", []string{}))
}

func devTestVMType(vm *armdevtestlabs.LabVirtualMachine) string {
	if vm.Properties != nil && vm.Properties.GalleryImageReference != nil && vm.Properties.GalleryImageReference.OSType != nil &&
		strings.EqualFold(*vm.Properties.GalleryImageReference.OSType, "Windows") {
		return "azurerm_dev_test_windows_virtual_machine"
	}
	return "azurerm_dev_test_linux_virtual_machine"
}

func (g *DevTestGenerator) InitResources() error {
	ctx := context.Background()
	subscriptionID := g.getSubscriptionID()
	cred := g.getCredential()
	opts := g.armClientOptions()
	rg := g.getResourceGroup()

	// Global VM shutdown schedules
	gsClient, err := armdevtestlabs.NewGlobalSchedulesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	if rg != "" {
		pager := gsClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_dev_test_global_vm_shutdown_schedule")
			}
		}
	} else {
		pager := gsClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_dev_test_global_vm_shutdown_schedule")
			}
		}
	}

	// Labs + children
	labClient, err := armdevtestlabs.NewLabsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	vmClient, err := armdevtestlabs.NewVirtualMachinesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	policyClient, err := armdevtestlabs.NewPoliciesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	scheduleClient, err := armdevtestlabs.NewSchedulesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	vnClient, err := armdevtestlabs.NewVirtualNetworksClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	var labs []*armdevtestlabs.Lab
	if rg != "" {
		pager := labClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			labs = append(labs, page.Value...)
		}
	} else {
		pager := labClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			labs = append(labs, page.Value...)
		}
	}
	for _, lab := range labs {
		g.appendSimple(*lab.ID, *lab.Name, "azurerm_dev_test_lab")
		labRG := resourceGroupFromID(*lab.ID)

		pager := vmClient.NewListPager(labRG, *lab.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, devTestVMType(item))
			}
		}
		ppager := policyClient.NewListPager(labRG, *lab.Name, "default", nil)
		for ppager.More() {
			page, err := ppager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_dev_test_policy")
			}
		}
		spager := scheduleClient.NewListPager(labRG, *lab.Name, nil)
		for spager.More() {
			page, err := spager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_dev_test_schedule")
			}
		}
		vpager := vnClient.NewListPager(labRG, *lab.Name, nil)
		for vpager.More() {
			page, err := vpager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_dev_test_virtual_network")
			}
		}
	}
	return nil
}
