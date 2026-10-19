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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type StorageAccountGenerator struct {
	AzureService
}

func appendStorageAccounts(resources []terraformutils.Resource, accounts []*armstorage.Account) []terraformutils.Resource {
	for _, account := range accounts {
		resources = append(resources, terraformutils.NewSimpleResource(
			*account.ID,
			*account.Name,
			"azurerm_storage_account",
			"azurerm",
			[]string{}))
	}
	return resources
}

func (g *StorageAccountGenerator) InitResources() error {
	ctx := context.Background()
	client, err := armstorage.NewAccountsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	var resources []terraformutils.Resource
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			resources = appendStorageAccounts(resources, page.Value)
		}
	} else {
		pager := client.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			resources = appendStorageAccounts(resources, page.Value)
		}
	}
	g.Resources = resources
	return nil
}
