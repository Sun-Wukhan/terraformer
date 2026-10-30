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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cosmosforpostgresql/armcosmosforpostgresql"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// CosmosPostgreSQLGenerator covers Azure Cosmos DB for PostgreSQL
// (Microsoft.DBforPostgreSQL/serverGroupsv2) via armcosmosforpostgresql.
type CosmosPostgreSQLGenerator struct {
	AzureService
}

// cosmosPostgreSQLConfigurationType maps a server role group to the
// provider resource type for a cluster configuration: the provider splits
// the generic ARM configurations list into coordinator and node resources,
// whose importer IDs use the coordinatorConfigurations / nodeConfigurations
// segments.
func cosmosPostgreSQLConfigurationType(role armcosmosforpostgresql.ServerRole) string {
	switch role {
	case armcosmosforpostgresql.ServerRoleCoordinator:
		return "azurerm_cosmosdb_postgresql_coordinator_configuration"
	case armcosmosforpostgresql.ServerRoleWorker:
		return "azurerm_cosmosdb_postgresql_node_configuration"
	}
	return ""
}

func (g *CosmosPostgreSQLGenerator) listClusters() ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	clustersClient, err := armcosmosforpostgresql.NewClustersClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	firewallClient, err := armcosmosforpostgresql.NewFirewallRulesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	rolesClient, err := armcosmosforpostgresql.NewRolesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	configurationsClient, err := armcosmosforpostgresql.NewConfigurationsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	processClusters := func(clusters []*armcosmosforpostgresql.Cluster) error {
		for _, cluster := range clusters {
			resources = append(resources, terraformutils.NewSimpleResource(
				*cluster.ID, *cluster.Name, "azurerm_cosmosdb_postgresql_cluster", g.ProviderName, []string{}))
			id, err := ParseAzureResourceID(*cluster.ID)
			if err != nil {
				return err
			}

			firewallPager := firewallClient.NewListByClusterPager(id.ResourceGroup, *cluster.Name, nil)
			for firewallPager.More() {
				page, err := firewallPager.NextPage(ctx)
				if err != nil {
					log.Println(err)
					return err
				}
				for _, rule := range page.Value {
					resources = append(resources, terraformutils.NewSimpleResource(
						*rule.ID, *rule.Name, "azurerm_cosmosdb_postgresql_firewall_rule", g.ProviderName, []string{}))
				}
			}

			rolesPager := rolesClient.NewListByClusterPager(id.ResourceGroup, *cluster.Name, nil)
			for rolesPager.More() {
				page, err := rolesPager.NextPage(ctx)
				if err != nil {
					log.Println(err)
					return err
				}
				for _, role := range page.Value {
					resources = append(resources, terraformutils.NewSimpleResource(
						*role.ID, *role.Name, "azurerm_cosmosdb_postgresql_role", g.ProviderName, []string{}))
				}
			}

			configurationsPager := configurationsClient.NewListByClusterPager(id.ResourceGroup, *cluster.Name, nil)
			for configurationsPager.More() {
				page, err := configurationsPager.NextPage(ctx)
				if err != nil {
					log.Println(err)
					return err
				}
				for _, configuration := range page.Value {
					if configuration.Properties == nil {
						continue
					}
					// One ARM configuration carries one value per server
					// role group; the provider models each role group as a
					// separate resource with its own ID segment.
					for _, roleGroup := range configuration.Properties.ServerRoleGroupConfigurations {
						if roleGroup.Role == nil {
							continue
						}
						resourceType := cosmosPostgreSQLConfigurationType(*roleGroup.Role)
						if resourceType == "" {
							continue
						}
						segment := "nodeConfigurations"
						if *roleGroup.Role == armcosmosforpostgresql.ServerRoleCoordinator {
							segment = "coordinatorConfigurations"
						}
						configurationID := *cluster.ID + "/" + segment + "/" + *configuration.Name
						resources = append(resources, terraformutils.NewSimpleResource(
							configurationID, *configuration.Name+"-"+*cluster.Name, resourceType, g.ProviderName, []string{}))
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

func (g *CosmosPostgreSQLGenerator) InitResources() error {
	resources, err := g.listClusters()
	if err != nil {
		return err
	}
	g.Resources = append(g.Resources, resources...)
	return nil
}
