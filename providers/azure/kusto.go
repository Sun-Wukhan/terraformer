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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/kusto/armkusto"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// KustoGenerator covers Azure Data Explorer via armkusto.
//
// Skipped in this family:
//   - azurerm_kusto_cluster_customer_managed_key: a singleton setting on
//     the cluster itself (its provider importer ID is the cluster ID);
//     armkusto exposes no separate CMK resource or List.
type KustoGenerator struct {
	AzureService
}

// kustoDataConnectionType maps the ARM data-connection kind to the
// provider resource type; all four kinds share the same importer ID shape.
func kustoDataConnectionType(kind *armkusto.DataConnectionKind) string {
	if kind == nil {
		return ""
	}
	switch *kind {
	case armkusto.DataConnectionKindCosmosDb:
		return "azurerm_kusto_cosmosdb_data_connection"
	case armkusto.DataConnectionKindEventGrid:
		return "azurerm_kusto_eventgrid_data_connection"
	case armkusto.DataConnectionKindEventHub:
		return "azurerm_kusto_eventhub_data_connection"
	case armkusto.DataConnectionKindIotHub:
		return "azurerm_kusto_iothub_data_connection"
	}
	return ""
}

func (g *KustoGenerator) listClusters() ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	clustersClient, err := armkusto.NewClustersClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	attachedClient, err := armkusto.NewAttachedDatabaseConfigurationsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	clusterPrincipalsClient, err := armkusto.NewClusterPrincipalAssignmentsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	managedPrivateEndpointsClient, err := armkusto.NewManagedPrivateEndpointsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	databasesClient, err := armkusto.NewDatabasesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	databasePrincipalsClient, err := armkusto.NewDatabasePrincipalAssignmentsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	dataConnectionsClient, err := armkusto.NewDataConnectionsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	scriptsClient, err := armkusto.NewScriptsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	add := func(id, name, resourceType string) {
		resources = append(resources, terraformutils.NewSimpleResource(id, name, resourceType, g.ProviderName, []string{}))
	}

	processClusters := func(clusters []*armkusto.Cluster) error {
		for _, cluster := range clusters {
			add(*cluster.ID, *cluster.Name, "azurerm_kusto_cluster")
			id, err := ParseAzureResourceID(*cluster.ID)
			if err != nil {
				return err
			}

			attachedPager := attachedClient.NewListByClusterPager(id.ResourceGroup, *cluster.Name, nil)
			for attachedPager.More() {
				page, err := attachedPager.NextPage(ctx)
				if err != nil {
					log.Println(err)
					return err
				}
				for _, item := range page.Value {
					add(*item.ID, *item.Name, "azurerm_kusto_attached_database_configuration")
				}
			}

			clusterPrincipalsPager := clusterPrincipalsClient.NewListPager(id.ResourceGroup, *cluster.Name, nil)
			for clusterPrincipalsPager.More() {
				page, err := clusterPrincipalsPager.NextPage(ctx)
				if err != nil {
					log.Println(err)
					return err
				}
				for _, item := range page.Value {
					add(*item.ID, *item.Name, "azurerm_kusto_cluster_principal_assignment")
				}
			}

			managedPrivateEndpointsPager := managedPrivateEndpointsClient.NewListPager(id.ResourceGroup, *cluster.Name, nil)
			for managedPrivateEndpointsPager.More() {
				page, err := managedPrivateEndpointsPager.NextPage(ctx)
				if err != nil {
					log.Println(err)
					return err
				}
				for _, item := range page.Value {
					add(*item.ID, *item.Name, "azurerm_kusto_cluster_managed_private_endpoint")
				}
			}

			databasesPager := databasesClient.NewListByClusterPager(id.ResourceGroup, *cluster.Name, nil)
			for databasesPager.More() {
				page, err := databasesPager.NextPage(ctx)
				if err != nil {
					log.Println(err)
					return err
				}
				for _, databaseItem := range page.Value {
					database := databaseItem.GetDatabase()
					if database == nil || database.ID == nil || database.Name == nil {
						continue
					}
					add(*database.ID, *database.Name, "azurerm_kusto_database")

					databasePrincipalsPager := databasePrincipalsClient.NewListPager(id.ResourceGroup, *cluster.Name, *database.Name, nil)
					for databasePrincipalsPager.More() {
						principalsPage, err := databasePrincipalsPager.NextPage(ctx)
						if err != nil {
							log.Println(err)
							return err
						}
						for _, item := range principalsPage.Value {
							add(*item.ID, *item.Name, "azurerm_kusto_database_principal_assignment")
						}
					}

					dataConnectionsPager := dataConnectionsClient.NewListByDatabasePager(id.ResourceGroup, *cluster.Name, *database.Name, nil)
					for dataConnectionsPager.More() {
						connectionsPage, err := dataConnectionsPager.NextPage(ctx)
						if err != nil {
							log.Println(err)
							return err
						}
						for _, item := range connectionsPage.Value {
							connection := item.GetDataConnection()
							if connection == nil || connection.ID == nil || connection.Name == nil {
								continue
							}
							resourceType := kustoDataConnectionType(connection.Kind)
							if resourceType == "" {
								continue
							}
							add(*connection.ID, *connection.Name, resourceType)
						}
					}

					scriptsPager := scriptsClient.NewListByDatabasePager(id.ResourceGroup, *cluster.Name, *database.Name, nil)
					for scriptsPager.More() {
						scriptsPage, err := scriptsPager.NextPage(ctx)
						if err != nil {
							log.Println(err)
							return err
						}
						for _, item := range scriptsPage.Value {
							add(*item.ID, *item.Name, "azurerm_kusto_script")
						}
					}
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
		pager := clustersClient.NewListPager(nil)
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

func (g *KustoGenerator) InitResources() error {
	resources, err := g.listClusters()
	if err != nil {
		return err
	}
	g.Resources = append(g.Resources, resources...)
	return nil
}
