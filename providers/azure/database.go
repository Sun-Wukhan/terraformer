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
	"fmt"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mariadb/armmariadb"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysql"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresql"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type DatabasesGenerator struct {
	AzureService
}

func (g *DatabasesGenerator) getMariaDBServers() ([]*armmariadb.Server, error) {
	ctx := context.Background()
	client, err := armmariadb.NewServersClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	var servers []*armmariadb.Server
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			servers = append(servers, page.Value...)
		}
	} else {
		pager := client.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			servers = append(servers, page.Value...)
		}
	}

	return servers, nil
}

func (g *DatabasesGenerator) createMariaDBServerResources(servers []*armmariadb.Server) ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource

	for _, server := range servers {
		resources = append(resources, terraformutils.NewResource(
			*server.ID,
			*server.Name,
			"azurerm_mariadb_server",
			g.ProviderName,
			map[string]string{},
			[]string{},
			map[string]interface{}{
				"administrator_login_password": "",
			}))
	}

	return resources, nil
}

func (g *DatabasesGenerator) createMariaDBConfigurationResources(servers []*armmariadb.Server) ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	client, err := armmariadb.NewConfigurationsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	for _, server := range servers {
		id, err := ParseAzureResourceID(*server.ID)
		if err != nil {
			return nil, err
		}
		pager := client.NewListByServerPager(id.ResourceGroup, *server.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			for _, config := range page.Value {
				resources = append(resources, terraformutils.NewSimpleResource(
					*config.ID,
					*config.Name+"-"+*server.Name,
					"azurerm_mariadb_configuration",
					g.ProviderName,
					[]string{"value"}))
			}
		}
	}

	return resources, nil
}

func (g *DatabasesGenerator) createMariaDBDatabaseResources(servers []*armmariadb.Server) ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	client, err := armmariadb.NewDatabasesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	for _, server := range servers {
		id, err := ParseAzureResourceID(*server.ID)
		if err != nil {
			return nil, err
		}
		pager := client.NewListByServerPager(id.ResourceGroup, *server.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			for _, database := range page.Value {
				resources = append(resources, terraformutils.NewSimpleResource(
					*database.ID,
					*database.Name+"-"+*server.Name,
					"azurerm_mariadb_database",
					g.ProviderName,
					[]string{}))
			}
		}
	}

	return resources, nil
}

func (g *DatabasesGenerator) createMariaDBFirewallRuleResources(servers []*armmariadb.Server) ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	client, err := armmariadb.NewFirewallRulesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	for _, server := range servers {
		id, err := ParseAzureResourceID(*server.ID)
		if err != nil {
			return nil, err
		}

		pager := client.NewListByServerPager(id.ResourceGroup, *server.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			for _, rule := range page.Value {
				resources = append(resources, terraformutils.NewSimpleResource(
					*rule.ID,
					*rule.Name,
					"azurerm_mariadb_firewall_rule",
					g.ProviderName,
					[]string{}))
			}
		}
	}

	return resources, nil
}

func (g *DatabasesGenerator) createMariaDBVirtualNetworkRuleResources(servers []*armmariadb.Server) ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	client, err := armmariadb.NewVirtualNetworkRulesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	for _, server := range servers {
		id, err := ParseAzureResourceID(*server.ID)
		if err != nil {
			return nil, err
		}
		pager := client.NewListByServerPager(id.ResourceGroup, *server.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			for _, rule := range page.Value {
				resources = append(resources, terraformutils.NewSimpleResource(
					*rule.ID,
					*rule.Name,
					"azurerm_mariadb_virtual_network_rule",
					g.ProviderName,
					[]string{}))
			}
		}
	}
	return resources, nil
}

func (g *DatabasesGenerator) getMySQLServers() ([]*armmysql.Server, error) {
	ctx := context.Background()
	client, err := armmysql.NewServersClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	var servers []*armmysql.Server
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			servers = append(servers, page.Value...)
		}
	} else {
		pager := client.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			servers = append(servers, page.Value...)
		}
	}

	return servers, nil
}

func (g *DatabasesGenerator) createMySQLServerResources(servers []*armmysql.Server) ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource

	for _, server := range servers {
		resources = append(resources, terraformutils.NewResource(
			*server.ID,
			*server.Name,
			"azurerm_mysql_server",
			g.ProviderName,
			map[string]string{},
			[]string{},
			map[string]interface{}{
				"administrator_login_password": "",
			}))
	}

	return resources, nil
}

func (g *DatabasesGenerator) createMySQLConfigurationResources(servers []*armmysql.Server) ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	client, err := armmysql.NewConfigurationsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	for _, server := range servers {
		id, err := ParseAzureResourceID(*server.ID)
		if err != nil {
			return nil, err
		}

		pager := client.NewListByServerPager(id.ResourceGroup, *server.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			for _, config := range page.Value {
				resources = append(resources, terraformutils.NewSimpleResource(
					*config.ID,
					*config.Name+"-"+*server.Name,
					"azurerm_mysql_configuration",
					g.ProviderName,
					[]string{"value"}))
			}
		}
	}

	return resources, nil
}

func (g *DatabasesGenerator) createMySQLDatabaseResources(servers []*armmysql.Server) ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	client, err := armmysql.NewDatabasesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	for _, server := range servers {
		id, err := ParseAzureResourceID(*server.ID)
		if err != nil {
			return nil, err
		}
		pager := client.NewListByServerPager(id.ResourceGroup, *server.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			for _, database := range page.Value {
				resources = append(resources, terraformutils.NewSimpleResource(
					*database.ID,
					*database.Name+"-"+*server.Name,
					"azurerm_mysql_database",
					g.ProviderName,
					[]string{}))
			}
		}
	}
	return resources, nil
}

func (g *DatabasesGenerator) createMySQLFirewallRuleResources(servers []*armmysql.Server) ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	client, err := armmysql.NewFirewallRulesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	for _, server := range servers {
		id, err := ParseAzureResourceID(*server.ID)
		if err != nil {
			return nil, err
		}
		pager := client.NewListByServerPager(id.ResourceGroup, *server.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			for _, rule := range page.Value {
				resources = append(resources, terraformutils.NewSimpleResource(
					*rule.ID,
					*rule.Name,
					"azurerm_mysql_firewall_rule",
					g.ProviderName,
					[]string{}))
			}
		}
	}

	return resources, nil
}

func (g *DatabasesGenerator) createMySQLVirtualNetworkRuleResources(servers []*armmysql.Server) ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	client, err := armmysql.NewVirtualNetworkRulesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	for _, server := range servers {
		id, err := ParseAzureResourceID(*server.ID)
		if err != nil {
			return nil, err
		}

		pager := client.NewListByServerPager(id.ResourceGroup, *server.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			for _, rule := range page.Value {
				resources = append(resources, terraformutils.NewSimpleResource(
					*rule.ID,
					*rule.Name,
					"azurerm_mysql_virtual_network_rule",
					g.ProviderName,
					[]string{}))
			}
		}
	}

	return resources, nil
}

func (g *DatabasesGenerator) getPostgreSQLServers() ([]*armpostgresql.Server, error) {
	ctx := context.Background()
	client, err := armpostgresql.NewServersClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	var servers []*armpostgresql.Server
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			servers = append(servers, page.Value...)
		}
	} else {
		pager := client.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			servers = append(servers, page.Value...)
		}
	}

	return servers, nil
}

func (g *DatabasesGenerator) createPostgreSQLServerResources(servers []*armpostgresql.Server) ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource

	for _, server := range servers {
		resources = append(resources, terraformutils.NewResource(
			*server.ID,
			*server.Name,
			"azurerm_postgresql_server",
			g.ProviderName,
			map[string]string{},
			[]string{},
			map[string]interface{}{
				"administrator_login_password": "",
			}))
	}

	return resources, nil
}

func (g *DatabasesGenerator) createPostgreSQLDatabaseResources(servers []*armpostgresql.Server) ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	client, err := armpostgresql.NewDatabasesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	for _, server := range servers {
		id, err := ParseAzureResourceID(*server.ID)
		if err != nil {
			return nil, err
		}
		pager := client.NewListByServerPager(id.ResourceGroup, *server.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			for _, database := range page.Value {
				resources = append(resources, terraformutils.NewSimpleResource(
					*database.ID,
					*database.Name+"-"+*server.Name,
					"azurerm_postgresql_database",
					g.ProviderName,
					[]string{}))
			}
		}
	}
	return resources, nil
}

func (g *DatabasesGenerator) createPostgreSQLConfigurationResources(servers []*armpostgresql.Server) ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	client, err := armpostgresql.NewConfigurationsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	for _, server := range servers {
		id, err := ParseAzureResourceID(*server.ID)
		if err != nil {
			return nil, err
		}
		pager := client.NewListByServerPager(id.ResourceGroup, *server.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			for _, config := range page.Value {
				resources = append(resources, terraformutils.NewSimpleResource(
					*config.ID,
					*config.Name+"-"+*server.Name,
					"azurerm_postgresql_configuration",
					g.ProviderName,
					[]string{"value"}))
			}
		}
	}
	return resources, nil
}

func (g *DatabasesGenerator) createPostgreSQLFirewallRuleResources(servers []*armpostgresql.Server) ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	client, err := armpostgresql.NewFirewallRulesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	for _, server := range servers {
		id, err := ParseAzureResourceID(*server.ID)
		if err != nil {
			return nil, err
		}
		pager := client.NewListByServerPager(id.ResourceGroup, *server.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			for _, rule := range page.Value {
				resources = append(resources, terraformutils.NewSimpleResource(
					*rule.ID,
					*rule.Name,
					"azurerm_postgresql_firewall_rule",
					g.ProviderName,
					[]string{}))
			}
		}
	}
	return resources, nil
}

func (g *DatabasesGenerator) createPostgreSQLVirtualNetworkRuleResources(servers []*armpostgresql.Server) ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	client, err := armpostgresql.NewVirtualNetworkRulesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	for _, server := range servers {
		id, err := ParseAzureResourceID(*server.ID)
		if err != nil {
			return nil, err
		}
		pager := client.NewListByServerPager(id.ResourceGroup, *server.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			for _, rule := range page.Value {
				resources = append(resources, terraformutils.NewSimpleResource(
					*rule.ID,
					*rule.Name,
					"azurerm_postgresql_virtual_network_rule",
					g.ProviderName,
					[]string{}))
			}
		}
	}
	return resources, nil
}

func (g *DatabasesGenerator) getSQLServers() ([]*armsql.Server, error) {
	var servers []*armsql.Server
	ctx := context.Background()
	client, err := armsql.NewServersClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			servers = append(servers, page.Value...)
		}
	} else {
		pager := client.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			servers = append(servers, page.Value...)
		}
	}

	return servers, nil
}

func (g *DatabasesGenerator) createSQLServerResources(servers []*armsql.Server) ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource

	for _, server := range servers {
		resources = append(resources, terraformutils.NewResource(
			*server.ID,
			*server.Name,
			"azurerm_mssql_server",
			g.ProviderName,
			map[string]string{},
			[]string{},
			map[string]interface{}{
				"administrator_login_password": "",
			}))
	}

	return resources, nil
}

func (g *DatabasesGenerator) createSQLDatabaseResources(servers []*armsql.Server) ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	client, err := armsql.NewDatabasesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	for _, server := range servers {
		id, err := ParseAzureResourceID(*server.ID)
		if err != nil {
			return nil, err
		}
		pager := client.NewListByServerPager(id.ResourceGroup, *server.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			for _, database := range page.Value {
				resources = append(resources, terraformutils.NewSimpleResource(
					*database.ID,
					*database.Name+"-"+*server.Name,
					"azurerm_mssql_database",
					g.ProviderName,
					[]string{}))
			}
		}
	}
	return resources, nil
}

func (g *DatabasesGenerator) createSQLFirewallRuleResources(servers []*armsql.Server) ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	client, err := armsql.NewFirewallRulesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	for _, server := range servers {
		id, err := ParseAzureResourceID(*server.ID)
		if err != nil {
			return nil, err
		}
		pager := client.NewListByServerPager(id.ResourceGroup, *server.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			for _, rule := range page.Value {
				resources = append(resources, terraformutils.NewSimpleResource(
					*rule.ID,
					*rule.Name,
					"azurerm_mssql_firewall_rule",
					g.ProviderName,
					[]string{}))
			}
		}
	}
	return resources, nil
}

func (g *DatabasesGenerator) createSQLVirtualNetworkRuleResources(servers []*armsql.Server) ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	client, err := armsql.NewVirtualNetworkRulesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	for _, server := range servers {
		id, err := ParseAzureResourceID(*server.ID)
		if err != nil {
			return nil, err
		}
		pager := client.NewListByServerPager(id.ResourceGroup, *server.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			for _, rule := range page.Value {
				resources = append(resources, terraformutils.NewSimpleResource(
					*rule.ID,
					*rule.Name,
					"azurerm_sql_virtual_network_rule",
					g.ProviderName,
					[]string{}))
			}
		}
	}
	return resources, nil
}

func (g *DatabasesGenerator) createSQLElasticPoolResources(servers []*armsql.Server) ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	client, err := armsql.NewElasticPoolsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	for _, server := range servers {
		id, err := ParseAzureResourceID(*server.ID)
		if err != nil {
			return nil, err
		}
		pager := client.NewListByServerPager(id.ResourceGroup, *server.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			for _, pool := range page.Value {
				resources = append(resources, terraformutils.NewSimpleResource(
					*pool.ID,
					*pool.Name,
					"azurerm_sql_elasticpool",
					g.ProviderName,
					[]string{}))
			}
		}
	}
	return resources, nil
}

func (g *DatabasesGenerator) createSQLFailoverResources(servers []*armsql.Server) ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	client, err := armsql.NewFailoverGroupsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	for _, server := range servers {
		id, err := ParseAzureResourceID(*server.ID)
		if err != nil {
			return nil, err
		}

		pager := client.NewListByServerPager(id.ResourceGroup, *server.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			for _, failoverGroup := range page.Value {
				resources = append(resources, terraformutils.NewSimpleResource(
					*failoverGroup.ID,
					*failoverGroup.Name,
					"azurerm_sql_failover_group",
					g.ProviderName,
					[]string{}))
			}
		}
	}
	return resources, nil
}

func (g *DatabasesGenerator) createSQLADAdministratorResources(servers []*armsql.Server) ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	client, err := armsql.NewServerAzureADAdministratorsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	for _, server := range servers {
		id, err := ParseAzureResourceID(*server.ID)
		if err != nil {
			return nil, err
		}

		pager := client.NewListByServerPager(id.ResourceGroup, *server.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			for _, administrator := range page.Value {
				resources = append(resources, terraformutils.NewSimpleResource(
					*administrator.ID,
					*administrator.Name,
					"azurerm_sql_active_directory_administrator",
					g.ProviderName,
					[]string{}))
			}
		}
	}
	return resources, nil
}

func (g *DatabasesGenerator) InitResources() error {
	mariadbServers, err := g.getMariaDBServers()
	if err != nil {
		return err
	}

	mysqlServers, err := g.getMySQLServers()
	if err != nil {
		return err
	}

	postgresqlServers, err := g.getPostgreSQLServers()
	if err != nil {
		return err
	}

	sqlServers, err := g.getSQLServers()
	if err != nil {
		return err
	}

	mariadbFunctions := []func([]*armmariadb.Server) ([]terraformutils.Resource, error){
		g.createMariaDBServerResources,
		g.createMariaDBDatabaseResources,
		g.createMariaDBConfigurationResources,
		g.createMariaDBFirewallRuleResources,
		g.createMariaDBVirtualNetworkRuleResources,
	}

	mysqlFunctions := []func([]*armmysql.Server) ([]terraformutils.Resource, error){
		g.createMySQLServerResources,
		g.createMySQLDatabaseResources,
		g.createMySQLConfigurationResources,
		g.createMySQLFirewallRuleResources,
		g.createMySQLVirtualNetworkRuleResources,
	}

	postgresqlFunctions := []func([]*armpostgresql.Server) ([]terraformutils.Resource, error){
		g.createPostgreSQLServerResources,
		g.createPostgreSQLDatabaseResources,
		g.createPostgreSQLConfigurationResources,
		g.createPostgreSQLFirewallRuleResources,
		g.createPostgreSQLVirtualNetworkRuleResources,
	}

	sqlFunctions := []func([]*armsql.Server) ([]terraformutils.Resource, error){
		g.createSQLServerResources,
		g.createSQLDatabaseResources,
		g.createSQLADAdministratorResources,
		g.createSQLElasticPoolResources,
		g.createSQLFailoverResources,
		g.createSQLFirewallRuleResources,
		g.createSQLVirtualNetworkRuleResources,
	}

	for _, f := range mariadbFunctions {
		resources, err := f(mariadbServers)
		if err != nil {
			return err
		}
		g.Resources = append(g.Resources, resources...)
	}

	for _, f := range mysqlFunctions {
		resources, err := f(mysqlServers)
		if err != nil {
			return err
		}
		g.Resources = append(g.Resources, resources...)
	}

	for _, f := range postgresqlFunctions {
		resources, err := f(postgresqlServers)
		if err != nil {
			return err
		}
		g.Resources = append(g.Resources, resources...)
	}

	for _, f := range sqlFunctions {
		resources, err := f(sqlServers)
		if err != nil {
			return err
		}
		g.Resources = append(g.Resources, resources...)
	}

	return nil
}

func (g *DatabasesGenerator) PostConvertHook() error {
	dbEngines := []string{
		"mariadb",
		"mysql",
		"postgresql",
		"sql",
	}

	for _, engineName := range dbEngines {
		for _, resource := range g.Resources {
			dbServerResourceType := fmt.Sprintf("azurerm_%s_server", engineName)
			if resource.InstanceInfo.Type == dbServerResourceType {
				dbName := resource.Item["name"]
				for rIdx, r := range g.Resources {
					if r.InstanceInfo.Type != dbServerResourceType &&
						strings.Contains(r.InstanceInfo.Type, engineName) &&
						r.Item["server_name"] == dbName {
						g.Resources[rIdx].Item["server_name"] = fmt.Sprintf("${%s.%s}", resource.InstanceInfo.Id, "name")
					}
				}
			}
		}
	}

	return nil
}
