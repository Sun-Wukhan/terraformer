// Copyright 2020 The Terraformer Authors.
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
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cosmos/armcosmos"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type CosmosDBGenerator struct {
	AzureService
}

func (g *CosmosDBGenerator) listSQLDatabasesAndContainersBehind(resourceGroupName string, accountName string) ([]terraformutils.Resource, []terraformutils.Resource, error) {
	var resourcesDatabase []terraformutils.Resource
	var resourcesContainer []terraformutils.Resource
	ctx := context.Background()
	sqlResourcesClient, err := armcosmos.NewSQLResourcesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, nil, err
	}

	databasesPager := sqlResourcesClient.NewListSQLDatabasesPager(resourceGroupName, accountName, nil)
	for databasesPager.More() {
		databasesPage, err := databasesPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return nil, nil, err
		}
		for _, sqlDatabase := range databasesPage.Value {
			// NOTE:
			// For a similar reason as
			// https://github.com/terraform-providers/terraform-provider-azurerm/issues/7472#issuecomment-650684349
			// The cosmosdb resource format change is NOT yet addressed in terraform provider
			// This line is a workaround to convert to old format, and might be removed if they deprecate the old format
			sqlDatabaseIDInOldFormat := strings.Replace(*sqlDatabase.ID, "sqlDatabases", "databases", 1)
			resourcesDatabase = append(resourcesDatabase, terraformutils.NewSimpleResource(
				sqlDatabaseIDInOldFormat,
				*sqlDatabase.Name,
				"azurerm_cosmosdb_sql_database",
				g.ProviderName,
				[]string{}))

			containersPager := sqlResourcesClient.NewListSQLContainersPager(resourceGroupName, accountName, *sqlDatabase.Name, nil)
			for containersPager.More() {
				containersPage, err := containersPager.NextPage(ctx)
				if err != nil {
					log.Println(err)
					return nil, nil, err
				}
				for _, sqlContainer := range containersPage.Value {
					// NOTE:
					// For a similar reason as
					// https://github.com/terraform-providers/terraform-provider-azurerm/issues/7472#issuecomment-650684349
					// The cosmosdb resource format change is NOT yet addressed in terraform provider
					// This line is a workaround to convert to old format, and might be removed if they deprecate the old format
					sqlContainerIDInOldFormat := strings.Replace(*sqlContainer.ID, "sqlDatabases", "databases", 1)
					resourcesContainer = append(resourcesContainer, terraformutils.NewSimpleResource(
						sqlContainerIDInOldFormat,
						*sqlContainer.Name,
						"azurerm_cosmosdb_sql_container",
						g.ProviderName,
						[]string{}))
				}
			}
		}
	}

	return resourcesDatabase, resourcesContainer, nil
}

func (g *CosmosDBGenerator) listTables(resourceGroupName string, accountName string) ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	tableResourcesClient, err := armcosmos.NewTableResourcesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	pager := tableResourcesClient.NewListTablesPager(resourceGroupName, accountName, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		for _, table := range page.Value {
			resources = append(resources, terraformutils.NewSimpleResource(
				*table.ID,
				*table.Name,
				"azurerm_cosmosdb_table",
				g.ProviderName,
				[]string{}))
		}
	}

	return resources, nil
}

func (g *CosmosDBGenerator) listAndAddForDatabaseAccounts() ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	databaseAccountsClient, err := armcosmos.NewDatabaseAccountsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	processAccounts := func(accounts []*armcosmos.DatabaseAccountGetResults) error {
		for _, account := range accounts {
			resources = append(resources, terraformutils.NewSimpleResource(
				*account.ID,
				*account.Name,
				"azurerm_cosmosdb_account",
				g.ProviderName,
				[]string{}))

			id, err := ParseAzureResourceID(*account.ID)
			if err != nil {
				return err
			}

			tables, err := g.listTables(id.ResourceGroup, *account.Name)
			if err != nil {
				return err
			}
			resources = append(resources, tables...)

			sqlDatabases, sqlContainers, err := g.listSQLDatabasesAndContainersBehind(id.ResourceGroup, *account.Name)
			if err != nil {
				return err
			}
			resources = append(resources, sqlDatabases...)
			resources = append(resources, sqlContainers...)
		}
		return nil
	}

	if rg := g.getResourceGroup(); rg != "" {
		pager := databaseAccountsClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			if err := processAccounts(page.Value); err != nil {
				return nil, err
			}
		}
	} else {
		pager := databaseAccountsClient.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			if err := processAccounts(page.Value); err != nil {
				return nil, err
			}
		}
	}

	return resources, nil
}

func (g *CosmosDBGenerator) InitResources() error {
	functions := []func() ([]terraformutils.Resource, error){
		g.listAndAddForDatabaseAccounts,
	}

	for _, f := range functions {
		resources, err := f()
		if err != nil {
			return err
		}
		g.Resources = append(g.Resources, resources...)
	}

	return nil
}
