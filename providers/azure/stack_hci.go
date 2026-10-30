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

	armazurestackhci "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/azurestackhci/armazurestackhci/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/azurestackhci/armazurestackhcivm"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// StackHCIGenerator covers Azure Stack HCI / Azure Local: clusters with
// their deployment settings and arc-setting extensions
// (armazurestackhci/v2), plus the HCI VM resources from
// armazurestackhcivm (logical networks, marketplace gallery images,
// network interfaces, storage paths = storageContainers, virtual hard
// disks).
type StackHCIGenerator struct {
	AzureService
}

func (g *StackHCIGenerator) appendSimple(id, name, resourceType string) {
	if id == "" || name == "" {
		return
	}
	g.Resources = append(g.Resources, terraformutils.NewSimpleResource(id, name, resourceType, "azurerm", []string{}))
}

func (g *StackHCIGenerator) InitResources() error {
	ctx := context.Background()
	subscriptionID := g.getSubscriptionID()
	cred := g.getCredential()
	opts := g.armClientOptions()
	rg := g.getResourceGroup()

	clusterClient, err := armazurestackhci.NewClustersClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	deploymentClient, err := armazurestackhci.NewDeploymentSettingsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	extClient, err := armazurestackhci.NewExtensionsClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	var clusters []*armazurestackhci.Cluster
	if rg != "" {
		pager := clusterClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			clusters = append(clusters, page.Value...)
		}
	} else {
		pager := clusterClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			clusters = append(clusters, page.Value...)
		}
	}
	for _, cluster := range clusters {
		g.appendSimple(*cluster.ID, *cluster.Name, "azurerm_stack_hci_cluster")
		clusterRG := resourceGroupFromID(*cluster.ID)
		dpager := deploymentClient.NewListByClustersPager(clusterRG, *cluster.Name, nil)
		for dpager.More() {
			page, err := dpager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_stack_hci_deployment_setting")
			}
		}
		epager := extClient.NewListByArcSettingPager(clusterRG, *cluster.Name, "default", nil)
		for epager.More() {
			page, err := epager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_stack_hci_extension")
			}
		}
	}

	// HCI VM resources (top-level in the resource group / subscription)
	lnClient, err := armazurestackhcivm.NewLogicalNetworksClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	if rg != "" {
		pager := lnClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_stack_hci_logical_network")
			}
		}
	} else {
		pager := lnClient.NewListAllPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_stack_hci_logical_network")
			}
		}
	}

	mgiClient, err := armazurestackhcivm.NewMarketplaceGalleryImagesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	if rg != "" {
		pager := mgiClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_stack_hci_marketplace_gallery_image")
			}
		}
	} else {
		pager := mgiClient.NewListAllPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_stack_hci_marketplace_gallery_image")
			}
		}
	}

	niClient, err := armazurestackhcivm.NewNetworkInterfacesClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	if rg != "" {
		pager := niClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_stack_hci_network_interface")
			}
		}
	} else {
		pager := niClient.NewListAllPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_stack_hci_network_interface")
			}
		}
	}

	scClient, err := armazurestackhcivm.NewStorageContainersClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	if rg != "" {
		pager := scClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_stack_hci_storage_path")
			}
		}
	} else {
		pager := scClient.NewListAllPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_stack_hci_storage_path")
			}
		}
	}

	vhdClient, err := armazurestackhcivm.NewVirtualHardDisksClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	if rg != "" {
		pager := vhdClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_stack_hci_virtual_hard_disk")
			}
		}
	} else {
		pager := vhdClient.NewListAllPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_stack_hci_virtual_hard_disk")
			}
		}
	}
	return nil
}
