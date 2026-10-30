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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storagediscovery/armstoragediscovery"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storagesync/armstoragesync"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// StorageSyncGenerator covers Azure File Sync via armstoragesync: storage
// sync services, sync groups, and cloud/server endpoints.
type StorageSyncGenerator struct {
	AzureService
}

func (g *StorageSyncGenerator) InitResources() error {
	ctx := context.Background()
	sub := g.getSubscriptionID()
	cred := g.getCredential()
	opts := g.armClientOptions()
	var resources []terraformutils.Resource
	add := func(id, name, typ string) {
		if id != "" && name != "" {
			resources = append(resources, terraformutils.NewSimpleResource(id, name, typ, "azurerm", []string{}))
		}
	}
	servicesClient, err := armstoragesync.NewServicesClient(sub, cred, opts)
	if err != nil {
		return err
	}
	groupsClient, err := armstoragesync.NewSyncGroupsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	cloudClient, err := armstoragesync.NewCloudEndpointsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	serverClient, err := armstoragesync.NewServerEndpointsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	type svcRef struct{ rg, name string }
	var services []svcRef
	if rg := g.getResourceGroup(); rg != "" {
		pager := servicesClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_storage_sync")
					services = append(services, svcRef{rg, *item.Name})
				}
			}
		}
	} else {
		pager := servicesClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				if item.ID == nil || item.Name == nil {
					continue
				}
				add(*item.ID, *item.Name, "azurerm_storage_sync")
				if parsed, err := ParseAzureResourceID(*item.ID); err == nil {
					services = append(services, svcRef{parsed.ResourceGroup, *item.Name})
				}
			}
		}
	}
	for _, svc := range services {
		gPager := groupsClient.NewListByStorageSyncServicePager(svc.rg, svc.name, nil)
		for gPager.More() {
			page, err := gPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, grp := range page.Value {
				if grp.ID == nil || grp.Name == nil {
					continue
				}
				add(*grp.ID, *grp.Name, "azurerm_storage_sync_group")
				cPager := cloudClient.NewListBySyncGroupPager(svc.rg, svc.name, *grp.Name, nil)
				for cPager.More() {
					cPage, err := cPager.NextPage(ctx)
					if err != nil {
						log.Println(err)
						break
					}
					for _, item := range cPage.Value {
						if item.ID != nil && item.Name != nil {
							add(*item.ID, *item.Name, "azurerm_storage_sync_cloud_endpoint")
						}
					}
				}
				sPager := serverClient.NewListBySyncGroupPager(svc.rg, svc.name, *grp.Name, nil)
				for sPager.More() {
					sPage, err := sPager.NextPage(ctx)
					if err != nil {
						log.Println(err)
						break
					}
					for _, item := range sPage.Value {
						if item.ID != nil && item.Name != nil {
							add(*item.ID, *item.Name, "azurerm_storage_sync_server_endpoint")
						}
					}
				}
			}
		}
	}
	g.Resources = resources
	return nil
}

// StorageDiscoveryGenerator covers Azure Storage Discovery workspaces via
// armstoragediscovery (Microsoft.StorageDiscovery), importing by ARM ID.
//
// Skipped in this family (see docs/azure.md "Known limits"):
//   - azurerm_qumulo_file_system: no track-2 SDK module is published for
//     Qumulo.Storage (no armqumulo package exists).
type StorageDiscoveryGenerator struct {
	AzureService
}

func (g *StorageDiscoveryGenerator) InitResources() error {
	ctx := context.Background()
	client, err := armstoragediscovery.NewWorkspacesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	var resources []terraformutils.Resource
	add := func(item *armstoragediscovery.Workspace) {
		if item.ID != nil && item.Name != nil {
			resources = append(resources, terraformutils.NewSimpleResource(*item.ID, *item.Name, "azurerm_storage_discovery_workspace", "azurerm", []string{}))
		}
	}
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				add(item)
			}
		}
	} else {
		pager := client.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				add(item)
			}
		}
	}
	g.Resources = resources
	return nil
}
