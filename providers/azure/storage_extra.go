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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// StorageExtraGenerator covers the armstorage child resources of a storage
// account: queues, file shares and tables (all import by their ARM ID per the
// azurerm provider docs), encryption scopes, local users, blob inventory
// policies, management policies and container immutability policies.
//
// Skipped in this family (see docs/azure.md "Known limits"):
//   - azurerm_storage_account_customer_managed_key,
//     azurerm_storage_account_network_rules,
//     azurerm_storage_account_queue_properties,
//     azurerm_storage_account_table_properties and
//     azurerm_storage_account_static_website are singleton properties of the
//     storage account / service object (Get only, no separate ARM ID or List).
//   - azurerm_storage_object_replication imports by a composite
//     source;destination policy ID pair, so a single enumerated policy cannot
//     form the provider ID.
//   - azurerm_storage_data_lake_gen2_filesystem,
//     azurerm_storage_data_lake_gen2_path, azurerm_storage_share_directory,
//     azurerm_storage_share_file and azurerm_storage_table_entity are
//     data-plane objects with no ARM List in armstorage.
type StorageExtraGenerator struct {
	AzureService
}

type storageAccountRef struct {
	rg   string
	name string
	id   string
}

func (g *StorageExtraGenerator) listAccounts(ctx context.Context) ([]storageAccountRef, error) {
	client, err := armstorage.NewAccountsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	var accounts []*armstorage.Account
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return nil, err
			}
			accounts = append(accounts, page.Value...)
		}
	} else {
		pager := client.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return nil, err
			}
			accounts = append(accounts, page.Value...)
		}
	}
	refs := make([]storageAccountRef, 0, len(accounts))
	for _, a := range accounts {
		if a.ID == nil || a.Name == nil {
			continue
		}
		parsed, err := ParseAzureResourceID(*a.ID)
		if err != nil {
			continue
		}
		refs = append(refs, storageAccountRef{rg: parsed.ResourceGroup, name: *a.Name, id: *a.ID})
	}
	return refs, nil
}

func (g *StorageExtraGenerator) InitResources() error {
	ctx := context.Background()
	accounts, err := g.listAccounts(ctx)
	if err != nil {
		return err
	}
	var resources []terraformutils.Resource
	add := func(id, name, typ string) {
		if id != "" && name != "" {
			resources = append(resources, terraformutils.NewSimpleResource(id, name, typ, "azurerm", []string{}))
		}
	}

	queueClient, err := armstorage.NewQueueClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	shareClient, err := armstorage.NewFileSharesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	tableClient, err := armstorage.NewTableClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	scopeClient, err := armstorage.NewEncryptionScopesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	inventoryClient, err := armstorage.NewBlobInventoryPoliciesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	localUsersClient, err := armstorage.NewLocalUsersClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	mgmtClient, err := armstorage.NewManagementPoliciesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	containerClient, err := armstorage.NewBlobContainersClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}

	for _, acct := range accounts {
		qPager := queueClient.NewListPager(acct.rg, acct.name, nil)
		for qPager.More() {
			page, err := qPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_storage_queue")
				}
			}
		}
		sPager := shareClient.NewListPager(acct.rg, acct.name, nil)
		for sPager.More() {
			page, err := sPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_storage_share")
				}
			}
		}
		tPager := tableClient.NewListPager(acct.rg, acct.name, nil)
		for tPager.More() {
			page, err := tPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_storage_table")
				}
			}
		}
		ePager := scopeClient.NewListPager(acct.rg, acct.name, nil)
		for ePager.More() {
			page, err := ePager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_storage_encryption_scope")
				}
			}
		}
		iPager := inventoryClient.NewListPager(acct.rg, acct.name, nil)
		for iPager.More() {
			page, err := iPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			// The provider imports a blob inventory policy by the storage
			// account ID itself, not the policy's ARM ID.
			if len(page.Value) > 0 {
				add(acct.id, acct.name, "azurerm_storage_blob_inventory_policy")
			}
		}
		lPager := localUsersClient.NewListPager(acct.rg, acct.name, nil)
		for lPager.More() {
			page, err := lPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_storage_account_local_user")
				}
			}
		}
		if resp, err := mgmtClient.Get(ctx, acct.rg, acct.name, armstorage.ManagementPolicyNameDefault, nil); err == nil && resp.ID != nil && resp.Name != nil {
			add(*resp.ID, acct.name, "azurerm_storage_management_policy")
		}
		cPager := containerClient.NewListPager(acct.rg, acct.name, nil)
		for cPager.More() {
			page, err := cPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.Name == nil {
					continue
				}
				if resp, err := containerClient.GetImmutabilityPolicy(ctx, acct.rg, acct.name, *item.Name, nil); err == nil && resp.ID != nil {
					add(*resp.ID, *item.Name, "azurerm_storage_container_immutability_policy")
				}
			}
		}
	}
	g.Resources = resources
	return nil
}
