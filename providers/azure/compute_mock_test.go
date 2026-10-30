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
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/batch/armbatch"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/devtestlabs/armdevtestlabs"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/hybridcompute/armhybridcompute"
)

func computeStrPtr(s string) *string { return &s }

func TestScaleSetTerraformType(t *testing.T) {
	flexible := armcompute.OrchestrationModeFlexible
	windows := armcompute.OperatingSystemTypesWindows
	linux := armcompute.OperatingSystemTypesLinux

	orchestrated := &armcompute.VirtualMachineScaleSet{
		Properties: &armcompute.VirtualMachineScaleSetProperties{OrchestrationMode: &flexible},
	}
	if got := scaleSetTerraformType(orchestrated); got != "azurerm_orchestrated_virtual_machine_scale_set" {
		t.Errorf("flexible = %q, want azurerm_orchestrated_virtual_machine_scale_set", got)
	}
	winSS := &armcompute.VirtualMachineScaleSet{
		Properties: &armcompute.VirtualMachineScaleSetProperties{
			VirtualMachineProfile: &armcompute.VirtualMachineScaleSetVMProfile{
				StorageProfile: &armcompute.VirtualMachineScaleSetStorageProfile{
					OSDisk: &armcompute.VirtualMachineScaleSetOSDisk{OSType: &windows},
				},
			},
		},
	}
	if got := scaleSetTerraformType(winSS); got != "azurerm_windows_virtual_machine_scale_set" {
		t.Errorf("windows uniform = %q, want azurerm_windows_virtual_machine_scale_set", got)
	}
	linuxSS := &armcompute.VirtualMachineScaleSet{
		Properties: &armcompute.VirtualMachineScaleSetProperties{
			VirtualMachineProfile: &armcompute.VirtualMachineScaleSetVMProfile{
				StorageProfile: &armcompute.VirtualMachineScaleSetStorageProfile{
					OSDisk: &armcompute.VirtualMachineScaleSetOSDisk{OSType: &linux},
				},
			},
		},
	}
	if got := scaleSetTerraformType(linuxSS); got != "azurerm_linux_virtual_machine_scale_set" {
		t.Errorf("linux uniform = %q, want azurerm_linux_virtual_machine_scale_set", got)
	}
}

func TestMaintenanceAssignmentType(t *testing.T) {
	cases := map[string]string{
		"/subscriptions/s/resourceGroups/rg1/providers/Microsoft.Compute/virtualMachines/vm1/providers/Microsoft.Maintenance/configurationAssignments/a1":         "azurerm_maintenance_assignment_virtual_machine",
		"/subscriptions/s/resourceGroups/rg1/providers/Microsoft.Compute/virtualMachineScaleSets/ss1/providers/Microsoft.Maintenance/configurationAssignments/a1": "azurerm_maintenance_assignment_virtual_machine_scale_set",
		"/subscriptions/s/resourceGroups/rg1/providers/Microsoft.Compute/hostGroups/hg1/hosts/h1/providers/Microsoft.Maintenance/configurationAssignments/a1":     "azurerm_maintenance_assignment_dedicated_host",
		"/subscriptions/s/providers/Microsoft.Maintenance/configurationAssignments/a1":                                                                            "azurerm_maintenance_assignment_dynamic_scope",
		"/subscriptions/s/resourceGroups/rg1/providers/Microsoft.Compute/disks/d1/providers/Microsoft.Maintenance/configurationAssignments/a1":                    "",
	}
	for id, want := range cases {
		if got := maintenanceAssignmentType(id); got != want {
			t.Errorf("maintenanceAssignmentType(%q) = %q, want %q", id, got, want)
		}
	}
}

func TestDevTestVMType(t *testing.T) {
	win := &armdevtestlabs.LabVirtualMachine{
		Properties: &armdevtestlabs.LabVirtualMachineProperties{
			GalleryImageReference: &armdevtestlabs.GalleryImageReference{OSType: computeStrPtr("Windows")},
		},
	}
	if got := devTestVMType(win); got != "azurerm_dev_test_windows_virtual_machine" {
		t.Errorf("windows lab VM = %q, want azurerm_dev_test_windows_virtual_machine", got)
	}
	linuxVM := &armdevtestlabs.LabVirtualMachine{
		Properties: &armdevtestlabs.LabVirtualMachineProperties{
			GalleryImageReference: &armdevtestlabs.GalleryImageReference{OSType: computeStrPtr("Linux")},
		},
	}
	if got := devTestVMType(linuxVM); got != "azurerm_dev_test_linux_virtual_machine" {
		t.Errorf("linux lab VM = %q, want azurerm_dev_test_linux_virtual_machine", got)
	}
}

const mockBatchAccountsJSON = `{"value":[
	{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.Batch/batchAccounts/ba-mock-1","name":"ba-mock-1","location":"eastus","type":"Microsoft.Batch/batchAccounts","properties":{}}
]}`

func TestMockTrack2BatchAccounts(t *testing.T) {
	tr := &cannedTransport{body: mockBatchAccountsJSON}
	client, err := armbatch.NewAccountClient(mockSubscriptionID, stubCredential{}, mockClientOptions(tr))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	pager := client.NewListPager(nil)
	ctx := context.Background()
	count := 0
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			t.Fatalf("next page: %v", err)
		}
		for _, account := range page.Value {
			if *account.Name != "ba-mock-1" {
				t.Errorf("name = %q, want ba-mock-1", *account.Name)
			}
			count++
		}
	}
	if count != 1 {
		t.Fatalf("got %d accounts, want 1", count)
	}
	if len(tr.urls) == 0 || !strings.Contains(tr.urls[0], "batchAccounts") {
		t.Errorf("expected an ARM batchAccounts call, got %v", tr.urls)
	}
}

const mockArcMachinesJSON = `{"value":[
	{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.HybridCompute/machines/arc-mock-1","name":"arc-mock-1","location":"eastus","type":"Microsoft.HybridCompute/machines","properties":{}}
]}`

func TestMockTrack2ArcMachines(t *testing.T) {
	tr := &cannedTransport{body: mockArcMachinesJSON}
	client, err := armhybridcompute.NewMachinesClient(mockSubscriptionID, stubCredential{}, mockClientOptions(tr))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	g := &ArcGenerator{}
	pager := client.NewListBySubscriptionPager(nil)
	ctx := context.Background()
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			t.Fatalf("next page: %v", err)
		}
		for _, machine := range page.Value {
			g.appendSimple(*machine.ID, *machine.Name, "azurerm_arc_machine")
		}
	}
	if len(g.Resources) != 1 {
		t.Fatalf("got %d resources, want 1", len(g.Resources))
	}
	if g.Resources[0].InstanceInfo.Type != "azurerm_arc_machine" {
		t.Errorf("type = %q, want azurerm_arc_machine", g.Resources[0].InstanceInfo.Type)
	}
	if len(tr.urls) == 0 || !strings.Contains(tr.urls[0], "Microsoft.HybridCompute") {
		t.Errorf("expected a Microsoft.HybridCompute call, got %v", tr.urls)
	}
}

const mockGalleriesJSON = `{"value":[
	{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.Compute/galleries/gal-mock-1","name":"gal-mock-1","location":"eastus","type":"Microsoft.Compute/galleries","properties":{}}
]}`

func TestMockTrack2Galleries(t *testing.T) {
	tr := &cannedTransport{body: mockGalleriesJSON}
	client, err := armcompute.NewGalleriesClient(mockSubscriptionID, stubCredential{}, mockClientOptions(tr))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	g := &ComputeGenerator{}
	pager := client.NewListPager(nil)
	ctx := context.Background()
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			t.Fatalf("next page: %v", err)
		}
		for _, gallery := range page.Value {
			g.appendSimple(*gallery.ID, *gallery.Name, "azurerm_shared_image_gallery")
		}
	}
	if len(g.Resources) != 1 {
		t.Fatalf("got %d resources, want 1", len(g.Resources))
	}
	if g.Resources[0].InstanceInfo.Type != "azurerm_shared_image_gallery" {
		t.Errorf("type = %q, want azurerm_shared_image_gallery", g.Resources[0].InstanceInfo.Type)
	}
	if len(tr.urls) == 0 || !strings.Contains(tr.urls[0], "galleries") {
		t.Errorf("expected an ARM galleries call, got %v", tr.urls)
	}
}
