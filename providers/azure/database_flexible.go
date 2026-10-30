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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysqlflexibleservers"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresqlflexibleservers"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// FlexibleServersGenerator covers the Azure Database flexible servers
// (MySQL + PostgreSQL) depth types via the track-2 flexible-servers SDKs.
//
// Skipped in this family (verified against the latest track-2 modules):
//   - azurerm_mysql_flexible_server_active_directory_administrator and
//     azurerm_postgresql_flexible_server_active_directory_administrator:
//     armmysqlflexibleservers v1.2.0 / armpostgresqlflexibleservers v1.1.0
//     contain no AdministratorsClient (the flexibleServers administrators
//     API is not shipped in the track-2 Go SDK).
//   - azurerm_postgresql_flexible_server_backup: no BackupsClient in
//     armpostgresqlflexibleservers v1.1.0.
//   - azurerm_postgresql_flexible_server_virtual_endpoint: no
//     VirtualEndpointsClient in armpostgresqlflexibleservers v1.1.0, and its
//     provider importer ID is a composite of two server IDs.
type FlexibleServersGenerator struct {
	AzureService
}

func (g *FlexibleServersGenerator) listMySQLFlexibleServers() ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	serversClient, err := armmysqlflexibleservers.NewServersClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	databasesClient, err := armmysqlflexibleservers.NewDatabasesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	configurationsClient, err := armmysqlflexibleservers.NewConfigurationsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	firewallClient, err := armmysqlflexibleservers.NewFirewallRulesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	processServers := func(servers []*armmysqlflexibleservers.Server) error {
		for _, server := range servers {
			resources = append(resources, terraformutils.NewResource(
				*server.ID, *server.Name, "azurerm_mysql_flexible_server", g.ProviderName,
				map[string]string{}, []string{}, map[string]interface{}{"administrator_password": ""}))
			id, err := ParseAzureResourceID(*server.ID)
			if err != nil {
				return err
			}

			databasesPager := databasesClient.NewListByServerPager(id.ResourceGroup, *server.Name, nil)
			for databasesPager.More() {
				page, err := databasesPager.NextPage(ctx)
				if err != nil {
					log.Println(err)
					return err
				}
				for _, database := range page.Value {
					resources = append(resources, terraformutils.NewSimpleResource(
						*database.ID, *database.Name+"-"+*server.Name, "azurerm_mysql_flexible_database", g.ProviderName, []string{}))
				}
			}

			configurationsPager := configurationsClient.NewListByServerPager(id.ResourceGroup, *server.Name, nil)
			for configurationsPager.More() {
				page, err := configurationsPager.NextPage(ctx)
				if err != nil {
					log.Println(err)
					return err
				}
				for _, configuration := range page.Value {
					resources = append(resources, terraformutils.NewSimpleResource(
						*configuration.ID, *configuration.Name+"-"+*server.Name, "azurerm_mysql_flexible_server_configuration", g.ProviderName, []string{"value"}))
				}
			}

			firewallPager := firewallClient.NewListByServerPager(id.ResourceGroup, *server.Name, nil)
			for firewallPager.More() {
				page, err := firewallPager.NextPage(ctx)
				if err != nil {
					log.Println(err)
					return err
				}
				for _, rule := range page.Value {
					resources = append(resources, terraformutils.NewSimpleResource(
						*rule.ID, *rule.Name, "azurerm_mysql_flexible_server_firewall_rule", g.ProviderName, []string{}))
				}
			}
		}
		return nil
	}

	if rg := g.getResourceGroup(); rg != "" {
		pager := serversClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			if err := processServers(page.Value); err != nil {
				return nil, err
			}
		}
	} else {
		pager := serversClient.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			if err := processServers(page.Value); err != nil {
				return nil, err
			}
		}
	}

	return resources, nil
}

func (g *FlexibleServersGenerator) listPostgreSQLFlexibleServers() ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	serversClient, err := armpostgresqlflexibleservers.NewServersClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	databasesClient, err := armpostgresqlflexibleservers.NewDatabasesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	configurationsClient, err := armpostgresqlflexibleservers.NewConfigurationsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	firewallClient, err := armpostgresqlflexibleservers.NewFirewallRulesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	processServers := func(servers []*armpostgresqlflexibleservers.Server) error {
		for _, server := range servers {
			resources = append(resources, terraformutils.NewResource(
				*server.ID, *server.Name, "azurerm_postgresql_flexible_server", g.ProviderName,
				map[string]string{}, []string{}, map[string]interface{}{"administrator_password": ""}))
			id, err := ParseAzureResourceID(*server.ID)
			if err != nil {
				return err
			}

			databasesPager := databasesClient.NewListByServerPager(id.ResourceGroup, *server.Name, nil)
			for databasesPager.More() {
				page, err := databasesPager.NextPage(ctx)
				if err != nil {
					log.Println(err)
					return err
				}
				for _, database := range page.Value {
					resources = append(resources, terraformutils.NewSimpleResource(
						*database.ID, *database.Name+"-"+*server.Name, "azurerm_postgresql_flexible_server_database", g.ProviderName, []string{}))
				}
			}

			configurationsPager := configurationsClient.NewListByServerPager(id.ResourceGroup, *server.Name, nil)
			for configurationsPager.More() {
				page, err := configurationsPager.NextPage(ctx)
				if err != nil {
					log.Println(err)
					return err
				}
				for _, configuration := range page.Value {
					resources = append(resources, terraformutils.NewSimpleResource(
						*configuration.ID, *configuration.Name+"-"+*server.Name, "azurerm_postgresql_flexible_server_configuration", g.ProviderName, []string{"value"}))
				}
			}

			firewallPager := firewallClient.NewListByServerPager(id.ResourceGroup, *server.Name, nil)
			for firewallPager.More() {
				page, err := firewallPager.NextPage(ctx)
				if err != nil {
					log.Println(err)
					return err
				}
				for _, rule := range page.Value {
					resources = append(resources, terraformutils.NewSimpleResource(
						*rule.ID, *rule.Name, "azurerm_postgresql_flexible_server_firewall_rule", g.ProviderName, []string{}))
				}
			}
		}
		return nil
	}

	if rg := g.getResourceGroup(); rg != "" {
		pager := serversClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			if err := processServers(page.Value); err != nil {
				return nil, err
			}
		}
	} else {
		pager := serversClient.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			if err := processServers(page.Value); err != nil {
				return nil, err
			}
		}
	}

	return resources, nil
}

func (g *FlexibleServersGenerator) InitResources() error {
	functions := []func() ([]terraformutils.Resource, error){
		g.listMySQLFlexibleServers,
		g.listPostgreSQLFlexibleServers,
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
