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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/batch/armbatch"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// BatchGenerator covers Microsoft.Batch accounts with their applications
// and pools (armbatch).
//
// Skipped in this family:
//   - azurerm_batch_job: jobs live in the Batch data plane; armbatch has
//     no Jobs client (only Account/Application/Pool/Certificate), so there
//     is no ARM List surface for jobs.
type BatchGenerator struct {
	AzureService
}

func (g *BatchGenerator) appendSimple(id, name, resourceType string) {
	if id == "" || name == "" {
		return
	}
	g.Resources = append(g.Resources, terraformutils.NewSimpleResource(id, name, resourceType, "azurerm", []string{}))
}

func (g *BatchGenerator) InitResources() error {
	ctx := context.Background()
	subscriptionID := g.getSubscriptionID()
	cred := g.getCredential()
	opts := g.armClientOptions()
	rg := g.getResourceGroup()

	accountClient, err := armbatch.NewAccountClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	appClient, err := armbatch.NewApplicationClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	poolClient, err := armbatch.NewPoolClient(subscriptionID, cred, opts)
	if err != nil {
		return err
	}
	var accounts []*armbatch.Account
	if rg != "" {
		pager := accountClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			accounts = append(accounts, page.Value...)
		}
	} else {
		pager := accountClient.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			accounts = append(accounts, page.Value...)
		}
	}
	for _, account := range accounts {
		g.appendSimple(*account.ID, *account.Name, "azurerm_batch_account")
		accountRG := resourceGroupFromID(*account.ID)

		pager := appClient.NewListPager(accountRG, *account.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_batch_application")
			}
		}
		ppager := poolClient.NewListByBatchAccountPager(accountRG, *account.Name, nil)
		for ppager.More() {
			page, err := ppager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_batch_pool")
			}
		}
	}
	return nil
}
