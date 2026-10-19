// Copyright 2019 The Terraformer Authors.
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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type VirtualMachineGenerator struct {
	AzureService
}

func (g *VirtualMachineGenerator) appendResources(vms []*armcompute.VirtualMachine) {
	for _, vm := range vms {
		var newResource terraformutils.Resource
		if vm.Properties.OSProfile == nil {
			if *vm.Properties.StorageProfile.OSDisk.OSType == armcompute.OperatingSystemTypesWindows {
				newResource = terraformutils.NewSimpleResource(
					*vm.ID,
					*vm.Name,
					"azurerm_windows_virtual_machine",
					"azurerm",
					[]string{})
			} else {
				newResource = terraformutils.NewSimpleResource(
					*vm.ID,
					*vm.Name,
					"azurerm_linux_virtual_machine",
					"azurerm",
					[]string{})
			}
		} else {
			if vm.Properties.OSProfile.WindowsConfiguration != nil {
				newResource = terraformutils.NewSimpleResource(
					*vm.ID,
					*vm.Name,
					"azurerm_windows_virtual_machine",
					"azurerm",
					[]string{})
			} else {
				newResource = terraformutils.NewSimpleResource(
					*vm.ID,
					*vm.Name,
					"azurerm_linux_virtual_machine",
					"azurerm",
					[]string{})
			}
		}

		g.Resources = append(g.Resources, newResource)
	}
}

func (g *VirtualMachineGenerator) InitResources() error {
	ctx := context.Background()
	client, err := armcompute.NewVirtualMachinesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}

	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			g.appendResources(page.Value)
		}
	} else {
		pager := client.NewListAllPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			g.appendResources(page.Value)
		}
	}
	return nil
}
