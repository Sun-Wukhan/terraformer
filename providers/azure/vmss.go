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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/standbypool/armstandbypool"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// ModernScaleSetGenerator emits the current VMSS resource types, which all
// share the Microsoft.Compute/virtualMachineScaleSets ARM object (and the
// importer ID of the legacy azurerm_virtual_machine_scale_set emitted by
// the scaleset service): Flexible orchestration sets are
// azurerm_orchestrated_virtual_machine_scale_set, Uniform sets are split
// into linux/windows by the OS disk type. It also emits
// azurerm_virtual_machine_scale_set_standby_pool (Microsoft.StandbyPool).
type ModernScaleSetGenerator struct {
	AzureService
}

func scaleSetTerraformType(ss *armcompute.VirtualMachineScaleSet) string {
	if ss.Properties != nil && ss.Properties.OrchestrationMode != nil &&
		*ss.Properties.OrchestrationMode == armcompute.OrchestrationModeFlexible {
		return "azurerm_orchestrated_virtual_machine_scale_set"
	}
	if ss.Properties != nil && ss.Properties.VirtualMachineProfile != nil &&
		ss.Properties.VirtualMachineProfile.StorageProfile != nil &&
		ss.Properties.VirtualMachineProfile.StorageProfile.OSDisk != nil &&
		ss.Properties.VirtualMachineProfile.StorageProfile.OSDisk.OSType != nil &&
		*ss.Properties.VirtualMachineProfile.StorageProfile.OSDisk.OSType == armcompute.OperatingSystemTypesWindows {
		return "azurerm_windows_virtual_machine_scale_set"
	}
	return "azurerm_linux_virtual_machine_scale_set"
}

func (g *ModernScaleSetGenerator) appendResources(scaleSets []*armcompute.VirtualMachineScaleSet) {
	for _, ss := range scaleSets {
		if ss.ID == nil || ss.Name == nil {
			continue
		}
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			*ss.ID, *ss.Name, scaleSetTerraformType(ss), "azurerm", []string{}))
	}
}

func (g *ModernScaleSetGenerator) InitResources() error {
	ctx := context.Background()
	subscriptionID := g.getSubscriptionID()
	cred := g.getCredential()
	opts := g.armClientOptions()
	rg := g.getResourceGroup()

	client, err := armcompute.NewVirtualMachineScaleSetsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	if rg != "" {
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

	// Standby pools for VMSS
	poolClient, err := armstandbypool.NewStandbyVirtualMachinePoolsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	appendPool := func(items []*armstandbypool.StandbyVirtualMachinePoolResource) {
		for _, item := range items {
			if item.ID == nil || item.Name == nil {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*item.ID, *item.Name, "azurerm_virtual_machine_scale_set_standby_pool", "azurerm", []string{}))
		}
	}
	if rg != "" {
		pager := poolClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			appendPool(page.Value)
		}
	} else {
		pager := poolClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			appendPool(page.Value)
		}
	}
	return nil
}
