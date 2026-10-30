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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/hybridcompute/armhybridcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/scvmm/armscvmm"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// SCVMMGenerator covers System Center Virtual Machine Manager
// (Microsoft.ScVmm): availability sets, clouds, VMM servers, VM templates
// and virtual networks are top-level; VM instances and their guest agents
// are scoped under Arc (HybridCompute) machines and are enumerated per
// machine via armscvmm's resource-URI pagers.
type SCVMMGenerator struct {
	AzureService
}

func (g *SCVMMGenerator) appendSimple(id, name, resourceType string) {
	if id == "" || name == "" {
		return
	}
	g.Resources = append(g.Resources, terraformutils.NewSimpleResource(id, name, resourceType, "azurerm", []string{}))
}

func (g *SCVMMGenerator) InitResources() error {
	ctx := context.Background()
	subscriptionID := g.getSubscriptionID()
	cred := g.getCredential()
	opts := g.armClientOptions()
	rg := g.getResourceGroup()

	asClient, err := armscvmm.NewAvailabilitySetsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	if rg != "" {
		pager := asClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_system_center_virtual_machine_manager_availability_set")
			}
		}
	} else {
		pager := asClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_system_center_virtual_machine_manager_availability_set")
			}
		}
	}

	cloudClient, err := armscvmm.NewCloudsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	if rg != "" {
		pager := cloudClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_system_center_virtual_machine_manager_cloud")
			}
		}
	} else {
		pager := cloudClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_system_center_virtual_machine_manager_cloud")
			}
		}
	}

	serverClient, err := armscvmm.NewVmmServersClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	if rg != "" {
		pager := serverClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_system_center_virtual_machine_manager_server")
			}
		}
	} else {
		pager := serverClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_system_center_virtual_machine_manager_server")
			}
		}
	}

	templateClient, err := armscvmm.NewVirtualMachineTemplatesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	if rg != "" {
		pager := templateClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_system_center_virtual_machine_manager_virtual_machine_template")
			}
		}
	} else {
		pager := templateClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_system_center_virtual_machine_manager_virtual_machine_template")
			}
		}
	}

	vnClient, err := armscvmm.NewVirtualNetworksClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	if rg != "" {
		pager := vnClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_system_center_virtual_machine_manager_virtual_network")
			}
		}
	} else {
		pager := vnClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_system_center_virtual_machine_manager_virtual_network")
			}
		}
	}

	// VM instances + guest agents, scoped under Arc machines
	machineClient, err := armhybridcompute.NewMachinesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	instanceClient, err := armscvmm.NewVirtualMachineInstancesClient(cred, opts)
	if err != nil {
		return err
	}
	guestClient, err := armscvmm.NewGuestAgentsClient(cred, opts)
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
		pager := instanceClient.NewListPager(*machine.ID, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				// Most Arc machines have no SCVMM instance; a failed
				// scoped list must not abort the whole service.
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_system_center_virtual_machine_manager_virtual_machine_instance")
				gpager := guestClient.NewListByVirtualMachineInstancePager(*item.ID, nil)
				for gpager.More() {
					gpage, err := gpager.NextPage(ctx)
					if err != nil {
						log.Println(err)
						break
					}
					for _, agent := range gpage.Value {
						g.appendSimple(*agent.ID, *agent.Name, "azurerm_system_center_virtual_machine_manager_virtual_machine_instance_guest_agent")
					}
				}
			}
		}
	}
	return nil
}
