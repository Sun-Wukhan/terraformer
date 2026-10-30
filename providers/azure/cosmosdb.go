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

// Depth enumeration added by the Azure full-coverage sweep (data cluster).
// Skipped in this family:
//   - azurerm_cosmosdb_mongo_role_definition / azurerm_cosmosdb_mongo_user_definition:
//     armcosmos v1.0.0 MongoDBResourcesClient exposes no role/user definition
//     operations at all; those definitions live behind the Cosmos MongoDB
//     data-plane endpoint, so there is no ARM List to enumerate.
//   - azurerm_cosmosdb_sql_dedicated_gateway: armcosmos v1.0.0 has no
//     ServicesClient; the account services endpoint (services/SqlDedicatedGateway)
//     is Get-only in the track-2 SDK, no List.

// listDepthChildrenBehind enumerates the per-account depth children:
// Cassandra keyspaces/tables, Gremlin databases/graphs, MongoDB
// databases/collections, SQL role definitions/assignments and the SQL
// container children (stored procedures, user-defined functions, triggers).
// All of these are ARM-enumerable via armcosmos resource clients; the
// stored-procedure/function/trigger importer IDs are the ARM
// sqlDatabases/.../containers/... IDs returned by ListSQL* pagers.
func (g *CosmosDBGenerator) listDepthChildrenBehind(resourceGroupName string, accountName string) ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()

	add := func(id, name, resourceType string) {
		resources = append(resources, terraformutils.NewSimpleResource(id, name, resourceType, g.ProviderName, []string{}))
	}

	cassandraClient, err := armcosmos.NewCassandraResourcesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	keyspacesPager := cassandraClient.NewListCassandraKeyspacesPager(resourceGroupName, accountName, nil)
	for keyspacesPager.More() {
		page, err := keyspacesPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		for _, keyspace := range page.Value {
			add(*keyspace.ID, *keyspace.Name, "azurerm_cosmosdb_cassandra_keyspace")
			tablesPager := cassandraClient.NewListCassandraTablesPager(resourceGroupName, accountName, *keyspace.Name, nil)
			for tablesPager.More() {
				tablesPage, err := tablesPager.NextPage(ctx)
				if err != nil {
					log.Println(err)
					return nil, err
				}
				for _, table := range tablesPage.Value {
					add(*table.ID, *table.Name, "azurerm_cosmosdb_cassandra_table")
				}
			}
		}
	}

	gremlinClient, err := armcosmos.NewGremlinResourcesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	gremlinPager := gremlinClient.NewListGremlinDatabasesPager(resourceGroupName, accountName, nil)
	for gremlinPager.More() {
		page, err := gremlinPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		for _, db := range page.Value {
			add(*db.ID, *db.Name, "azurerm_cosmosdb_gremlin_database")
			graphsPager := gremlinClient.NewListGremlinGraphsPager(resourceGroupName, accountName, *db.Name, nil)
			for graphsPager.More() {
				graphsPage, err := graphsPager.NextPage(ctx)
				if err != nil {
					log.Println(err)
					return nil, err
				}
				for _, graph := range graphsPage.Value {
					add(*graph.ID, *graph.Name, "azurerm_cosmosdb_gremlin_graph")
				}
			}
		}
	}

	mongoClient, err := armcosmos.NewMongoDBResourcesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	mongoPager := mongoClient.NewListMongoDBDatabasesPager(resourceGroupName, accountName, nil)
	for mongoPager.More() {
		page, err := mongoPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		for _, db := range page.Value {
			add(*db.ID, *db.Name, "azurerm_cosmosdb_mongo_database")
			collectionsPager := mongoClient.NewListMongoDBCollectionsPager(resourceGroupName, accountName, *db.Name, nil)
			for collectionsPager.More() {
				collectionsPage, err := collectionsPager.NextPage(ctx)
				if err != nil {
					log.Println(err)
					return nil, err
				}
				for _, collection := range collectionsPage.Value {
					add(*collection.ID, *collection.Name, "azurerm_cosmosdb_mongo_collection")
				}
			}
		}
	}

	sqlClient, err := armcosmos.NewSQLResourcesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	roleDefinitionsPager := sqlClient.NewListSQLRoleDefinitionsPager(resourceGroupName, accountName, nil)
	for roleDefinitionsPager.More() {
		page, err := roleDefinitionsPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		for _, item := range page.Value {
			add(*item.ID, *item.Name, "azurerm_cosmosdb_sql_role_definition")
		}
	}
	roleAssignmentsPager := sqlClient.NewListSQLRoleAssignmentsPager(resourceGroupName, accountName, nil)
	for roleAssignmentsPager.More() {
		page, err := roleAssignmentsPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		for _, item := range page.Value {
			add(*item.ID, *item.Name, "azurerm_cosmosdb_sql_role_assignment")
		}
	}

	// SQL container children, chained account -> database -> container using
	// the raw ARM IDs (the provider importer for these types uses the
	// sqlDatabases form, unlike sql_database/sql_container above).
	databasesPager := sqlClient.NewListSQLDatabasesPager(resourceGroupName, accountName, nil)
	for databasesPager.More() {
		databasesPage, err := databasesPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		for _, db := range databasesPage.Value {
			containersPager := sqlClient.NewListSQLContainersPager(resourceGroupName, accountName, *db.Name, nil)
			for containersPager.More() {
				containersPage, err := containersPager.NextPage(ctx)
				if err != nil {
					log.Println(err)
					return nil, err
				}
				for _, container := range containersPage.Value {
					proceduresPager := sqlClient.NewListSQLStoredProceduresPager(resourceGroupName, accountName, *db.Name, *container.Name, nil)
					for proceduresPager.More() {
						proceduresPage, err := proceduresPager.NextPage(ctx)
						if err != nil {
							log.Println(err)
							return nil, err
						}
						for _, item := range proceduresPage.Value {
							add(*item.ID, *item.Name, "azurerm_cosmosdb_sql_stored_procedure")
						}
					}
					functionsPager := sqlClient.NewListSQLUserDefinedFunctionsPager(resourceGroupName, accountName, *db.Name, *container.Name, nil)
					for functionsPager.More() {
						functionsPage, err := functionsPager.NextPage(ctx)
						if err != nil {
							log.Println(err)
							return nil, err
						}
						for _, item := range functionsPage.Value {
							add(*item.ID, *item.Name, "azurerm_cosmosdb_sql_function")
						}
					}
					triggersPager := sqlClient.NewListSQLTriggersPager(resourceGroupName, accountName, *db.Name, *container.Name, nil)
					for triggersPager.More() {
						triggersPage, err := triggersPager.NextPage(ctx)
						if err != nil {
							log.Println(err)
							return nil, err
						}
						for _, item := range triggersPage.Value {
							add(*item.ID, *item.Name, "azurerm_cosmosdb_sql_trigger")
						}
					}
				}
			}
		}
	}

	return resources, nil
}

// listCassandraClusters enumerates the managed Cassandra clusters
// (Microsoft.DocumentDB/cassandraClusters) and their data centers. These
// are a separate top-level resource from database accounts.
func (g *CosmosDBGenerator) listCassandraClusters() ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	clustersClient, err := armcosmos.NewCassandraClustersClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	dataCentersClient, err := armcosmos.NewCassandraDataCentersClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	processClusters := func(clusters []*armcosmos.ClusterResource) error {
		for _, cluster := range clusters {
			resources = append(resources, terraformutils.NewSimpleResource(
				*cluster.ID, *cluster.Name, "azurerm_cosmosdb_cassandra_cluster", g.ProviderName, []string{}))
			id, err := ParseAzureResourceID(*cluster.ID)
			if err != nil {
				return err
			}
			pager := dataCentersClient.NewListPager(id.ResourceGroup, *cluster.Name, nil)
			for pager.More() {
				page, err := pager.NextPage(ctx)
				if err != nil {
					log.Println(err)
					return err
				}
				for _, dc := range page.Value {
					resources = append(resources, terraformutils.NewSimpleResource(
						*dc.ID, *dc.Name, "azurerm_cosmosdb_cassandra_datacenter", g.ProviderName, []string{}))
				}
			}
		}
		return nil
	}

	if rg := g.getResourceGroup(); rg != "" {
		pager := clustersClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			if err := processClusters(page.Value); err != nil {
				return nil, err
			}
		}
	} else {
		pager := clustersClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			if err := processClusters(page.Value); err != nil {
				return nil, err
			}
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

			depthChildren, err := g.listDepthChildrenBehind(id.ResourceGroup, *account.Name)
			if err != nil {
				return err
			}
			resources = append(resources, depthChildren...)
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
		g.listCassandraClusters,
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
