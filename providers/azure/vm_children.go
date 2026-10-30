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
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// VirtualMachineChildGenerator covers the child resources of VMs and VM
// scale sets, enumerated via their parents:
//   - azurerm_virtual_machine_extension (VirtualMachineExtensionsClient.List)
//   - azurerm_virtual_machine_run_command (VirtualMachineRunCommandsClient)
//   - azurerm_virtual_machine_scale_set_extension
//   - azurerm_virtual_machine_data_disk_attachment and
//     azurerm_virtual_machine_implicit_data_disk_from_source: both are
//     entries of the VM's storageProfile.dataDisks (the provider importer
//     ID is {vmID}/dataDisks/{diskName}); entries with createOption=Attach
//     and a managed disk are attachments, entries created FromImage are
//     implicit disks.
//
// Skipped in this family:
//   - azurerm_virtual_machine_packet_capture and
//     azurerm_virtual_machine_scale_set_packet_capture: their importer ID
//     is the Microsoft.Network networkWatchers/packetCaptures ARM ID, the
//     same object already emitted as azurerm_network_packet_capture by the
//     network_watcher service; there is no distinct List surface.
//   - azurerm_virtual_machine_gallery_application_assignment: importer ID
//     is a composite "{vmId}|{galleryApplicationVersionId}" association
//     with no ARM object and no List surface.
//   - azurerm_virtual_machine (legacy): its importer ID is the same
//     Microsoft.Compute/virtualMachines ARM ID already emitted as
//     azurerm_linux_virtual_machine / azurerm_windows_virtual_machine by
//     the virtual_machine service; emitting the legacy type as well
//     would duplicate every VM.
type VirtualMachineChildGenerator struct {
	AzureService
}

func (g *VirtualMachineChildGenerator) appendSimple(id, name, resourceType string) {
	if id == "" || name == "" {
		return
	}
	g.Resources = append(g.Resources, terraformutils.NewSimpleResource(id, name, resourceType, "azurerm", []string{}))
}

func (g *VirtualMachineChildGenerator) InitResources() error {
	ctx := context.Background()
	subscriptionID := g.getSubscriptionID()
	cred := g.getCredential()
	opts := g.armClientOptions()
	rg := g.getResourceGroup()

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

	extClient, err := armcompute.NewVirtualMachineExtensionsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	rcClient, err := armcompute.NewVirtualMachineRunCommandsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	for _, vm := range vms {
		vmRG := resourceGroupFromID(*vm.ID)
		// Extensions
		resp, err := extClient.List(ctx, vmRG, *vm.Name, nil)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, ext := range resp.Value {
			g.appendSimple(*ext.ID, *ext.Name, "azurerm_virtual_machine_extension")
		}
		// Run commands
		pager := rcClient.NewListByVirtualMachinePager(vmRG, *vm.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_virtual_machine_run_command")
			}
		}
		// Data disks: attachments vs implicit disks from source
		if vm.Properties != nil && vm.Properties.StorageProfile != nil {
			for _, disk := range vm.Properties.StorageProfile.DataDisks {
				if disk.Name == nil {
					continue
				}
				diskID := *vm.ID + "/dataDisks/" + *disk.Name
				if disk.CreateOption != nil && *disk.CreateOption == armcompute.DiskCreateOptionTypesFromImage {
					g.appendSimple(diskID, *disk.Name, "azurerm_virtual_machine_implicit_data_disk_from_source")
				} else if disk.ManagedDisk != nil && disk.ManagedDisk.ID != nil {
					g.appendSimple(diskID, *disk.Name, "azurerm_virtual_machine_data_disk_attachment")
				}
			}
		}
	}

	// VMSS extensions
	ssClient, err := armcompute.NewVirtualMachineScaleSetsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	ssextClient, err := armcompute.NewVirtualMachineScaleSetExtensionsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	var scaleSets []*armcompute.VirtualMachineScaleSet
	if rg != "" {
		pager := ssClient.NewListPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			scaleSets = append(scaleSets, page.Value...)
		}
	} else {
		pager := ssClient.NewListAllPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			scaleSets = append(scaleSets, page.Value...)
		}
	}
	for _, ss := range scaleSets {
		ssRG := resourceGroupFromID(*ss.ID)
		pager := ssextClient.NewListPager(ssRG, *ss.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_virtual_machine_scale_set_extension")
			}
		}
	}
	return nil
}
