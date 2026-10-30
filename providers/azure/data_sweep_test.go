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
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cosmosforpostgresql/armcosmosforpostgresql"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dataprotection/armdataprotection"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/datashare/armdatashare"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/elastic/armelastic"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/kusto/armkusto"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mongocluster/armmongocluster"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysqlflexibleservers"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/synapse/armsynapse"
)

// Mock enumeration tests for the data-cluster sweep: canned ARM JSON
// through the real track-2 SDK clients/pagers (stubCredential +
// cannedTransport from azure_mock_test.go), plus the discriminator
// mapping helpers factored out of the generators.

const mockKustoClustersJSON = `{"value":[
	{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.Kusto/clusters/kusto-mock-1","name":"kusto-mock-1","location":"eastus","type":"Microsoft.Kusto/clusters","properties":{},"sku":{"name":"Standard_D11_v2","tier":"Standard","capacity":2}}
]}`

func TestMockTrack2KustoClusters(t *testing.T) {
	tr := &cannedTransport{body: mockKustoClustersJSON}
	client, err := armkusto.NewClustersClient(mockSubscriptionID, stubCredential{}, mockClientOptions(tr))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	pager := client.NewListPager(nil)
	ctx := context.Background()
	count := 0
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			t.Fatalf("next page: %v", err)
		}
		for _, cluster := range page.Value {
			if *cluster.Name != "kusto-mock-1" {
				t.Errorf("name = %q, want kusto-mock-1", *cluster.Name)
			}
			count++
		}
	}
	if count != 1 {
		t.Fatalf("got %d clusters, want 1", count)
	}
	if len(tr.urls) == 0 || !strings.Contains(tr.urls[0], "Microsoft.Kusto") {
		t.Errorf("expected a Microsoft.Kusto call, got %v", tr.urls)
	}
}

func TestKustoDataConnectionTypeMapping(t *testing.T) {
	cases := map[armkusto.DataConnectionKind]string{
		armkusto.DataConnectionKindCosmosDb:  "azurerm_kusto_cosmosdb_data_connection",
		armkusto.DataConnectionKindEventGrid: "azurerm_kusto_eventgrid_data_connection",
		armkusto.DataConnectionKindEventHub:  "azurerm_kusto_eventhub_data_connection",
		armkusto.DataConnectionKindIotHub:    "azurerm_kusto_iothub_data_connection",
	}
	for kind, want := range cases {
		k := kind
		if got := kustoDataConnectionType(&k); got != want {
			t.Errorf("kind %q: got %q, want %q", kind, got, want)
		}
	}
	if got := kustoDataConnectionType(nil); got != "" {
		t.Errorf("nil kind: got %q, want empty", got)
	}
}

const mockCosmosPostgreSQLClustersJSON = `{"value":[
	{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.DBforPostgreSQL/serverGroupsv2/pg-mock-1","name":"pg-mock-1","location":"eastus","type":"Microsoft.DBforPostgreSQL/serverGroupsv2","properties":{}}
]}`

func TestMockTrack2CosmosPostgreSQLClusters(t *testing.T) {
	tr := &cannedTransport{body: mockCosmosPostgreSQLClustersJSON}
	client, err := armcosmosforpostgresql.NewClustersClient(mockSubscriptionID, stubCredential{}, mockClientOptions(tr))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	pager := client.NewListPager(nil)
	ctx := context.Background()
	if !pager.More() {
		t.Fatal("expected at least one page")
	}
	page, err := pager.NextPage(ctx)
	if err != nil {
		t.Fatalf("next page: %v", err)
	}
	if len(page.Value) != 1 || *page.Value[0].Name != "pg-mock-1" {
		t.Fatalf("unexpected clusters page: %+v", page.Value)
	}
	if got := cosmosPostgreSQLConfigurationType(armcosmosforpostgresql.ServerRoleCoordinator); got != "azurerm_cosmosdb_postgresql_coordinator_configuration" {
		t.Errorf("coordinator config type = %q", got)
	}
	if got := cosmosPostgreSQLConfigurationType(armcosmosforpostgresql.ServerRoleWorker); got != "azurerm_cosmosdb_postgresql_node_configuration" {
		t.Errorf("worker config type = %q", got)
	}
}

const mockMongoClustersJSON = `{"value":[
	{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.DocumentDB/mongoClusters/mongo-mock-1","name":"mongo-mock-1","location":"eastus","type":"Microsoft.DocumentDB/mongoClusters","properties":{}}
]}`

func TestMockTrack2MongoClusters(t *testing.T) {
	tr := &cannedTransport{body: mockMongoClustersJSON}
	client, err := armmongocluster.NewMongoClustersClient(mockSubscriptionID, stubCredential{}, mockClientOptions(tr))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	pager := client.NewListPager(nil)
	ctx := context.Background()
	if !pager.More() {
		t.Fatal("expected at least one page")
	}
	page, err := pager.NextPage(ctx)
	if err != nil {
		t.Fatalf("next page: %v", err)
	}
	if len(page.Value) != 1 || *page.Value[0].Name != "mongo-mock-1" {
		t.Fatalf("unexpected clusters page: %+v", page.Value)
	}
	if len(tr.urls) == 0 || !strings.Contains(tr.urls[0], "mongoClusters") {
		t.Errorf("expected a mongoClusters call, got %v", tr.urls)
	}
}

const mockBackupVaultsJSON = `{"value":[
	{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.DataProtection/backupVaults/vault-mock-1","name":"vault-mock-1","location":"eastus","type":"Microsoft.DataProtection/backupVaults","properties":{"storageSettings":[{"datastoreType":"VaultStore","type":"LocallyRedundant"}]}}
]}`

func TestMockTrack2DataProtectionVaults(t *testing.T) {
	tr := &cannedTransport{body: mockBackupVaultsJSON}
	client, err := armdataprotection.NewBackupVaultsClient(mockSubscriptionID, stubCredential{}, mockClientOptions(tr))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	pager := client.NewGetInSubscriptionPager(nil)
	ctx := context.Background()
	if !pager.More() {
		t.Fatal("expected at least one page")
	}
	page, err := pager.NextPage(ctx)
	if err != nil {
		t.Fatalf("next page: %v", err)
	}
	if len(page.Value) != 1 || *page.Value[0].Name != "vault-mock-1" {
		t.Fatalf("unexpected vaults page: %+v", page.Value)
	}
}

func TestDataProtectionTypeMappings(t *testing.T) {
	blob := "Microsoft.Storage/storageAccounts/blobServices"
	if got := dataProtectionInstanceType(&blob); got != "azurerm_data_protection_backup_instance_blob_storage" {
		t.Errorf("blob instance type = %q", got)
	}
	adls := "Microsoft.Storage/storageAccounts/adlsBlobServices"
	if got := dataProtectionInstanceType(&adls); got != "azurerm_data_protection_backup_instance_data_lake_storage" {
		t.Errorf("adls instance type = %q", got)
	}
	cosmos := "Microsoft.DocumentDB/databaseAccounts"
	if got := dataProtectionPolicyType([]*string{&cosmos}); got != "azurerm_data_protection_backup_policy_cosmosdb_account" {
		t.Errorf("cosmos policy type = %q", got)
	}
	unknown := "Microsoft.Unknown/things"
	if got := dataProtectionPolicyType([]*string{&unknown}); got != "" {
		t.Errorf("unknown policy type = %q, want empty", got)
	}
}

const mockDataShareAccountsJSON = `{"value":[
	{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.DataShare/accounts/share-mock-1","name":"share-mock-1","location":"eastus","type":"Microsoft.DataShare/accounts","properties":{}}
]}`

func TestMockTrack2DataShareAccounts(t *testing.T) {
	tr := &cannedTransport{body: mockDataShareAccountsJSON}
	client, err := armdatashare.NewAccountsClient(mockSubscriptionID, stubCredential{}, mockClientOptions(tr))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	pager := client.NewListBySubscriptionPager(nil)
	ctx := context.Background()
	if !pager.More() {
		t.Fatal("expected at least one page")
	}
	page, err := pager.NextPage(ctx)
	if err != nil {
		t.Fatalf("next page: %v", err)
	}
	if len(page.Value) != 1 || *page.Value[0].Name != "share-mock-1" {
		t.Fatalf("unexpected accounts page: %+v", page.Value)
	}
	kustoKind := armdatashare.DataSetKindKustoDatabase
	if got := dataShareDatasetType(&kustoKind); got != "azurerm_data_share_dataset_kusto_database" {
		t.Errorf("kusto dataset type = %q", got)
	}
	blobKind := armdatashare.DataSetKindContainer
	if got := dataShareDatasetType(&blobKind); got != "azurerm_data_share_dataset_blob_storage" {
		t.Errorf("blob dataset type = %q", got)
	}
}

const mockElasticMonitorsJSON = `{"value":[
	{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.Elastic/monitors/elastic-mock-1","name":"elastic-mock-1","location":"eastus","type":"Microsoft.Elastic/monitors","properties":{}}
]}`

func TestMockTrack2ElasticMonitors(t *testing.T) {
	tr := &cannedTransport{body: mockElasticMonitorsJSON}
	client, err := armelastic.NewMonitorsClient(mockSubscriptionID, stubCredential{}, mockClientOptions(tr))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	pager := client.NewListPager(nil)
	ctx := context.Background()
	if !pager.More() {
		t.Fatal("expected at least one page")
	}
	page, err := pager.NextPage(ctx)
	if err != nil {
		t.Fatalf("next page: %v", err)
	}
	if len(page.Value) != 1 || *page.Value[0].Name != "elastic-mock-1" {
		t.Fatalf("unexpected monitors page: %+v", page.Value)
	}
}

const mockMySQLFlexibleServersJSON = `{"value":[
	{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-mock-1/providers/Microsoft.DBforMySQL/flexibleServers/mysql-mock-1","name":"mysql-mock-1","location":"eastus","type":"Microsoft.DBforMySQL/flexibleServers","properties":{},"sku":{"name":"Standard_B1ms","tier":"Burstable"}}
]}`

func TestMockTrack2MySQLFlexibleServers(t *testing.T) {
	tr := &cannedTransport{body: mockMySQLFlexibleServersJSON}
	client, err := armmysqlflexibleservers.NewServersClient(mockSubscriptionID, stubCredential{}, mockClientOptions(tr))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	pager := client.NewListPager(nil)
	ctx := context.Background()
	if !pager.More() {
		t.Fatal("expected at least one page")
	}
	page, err := pager.NextPage(ctx)
	if err != nil {
		t.Fatalf("next page: %v", err)
	}
	if len(page.Value) != 1 || *page.Value[0].Name != "mysql-mock-1" {
		t.Fatalf("unexpected servers page: %+v", page.Value)
	}
}

func TestDataFactorySweepMappings(t *testing.T) {
	cases := map[string]string{
		"AzureSqlTable":       "azurerm_data_factory_dataset_azure_sql_table",
		"CosmosDbMongoDbApi":  "azurerm_data_factory_linked_service_cosmosdb_mongoapi",
		"Odbc":                "azurerm_data_factory_linked_service_odbc",
		"CustomEventsTrigger": "azurerm_data_factory_trigger_custom_event",
	}
	for azureType, want := range cases {
		if got := getResourceTypeFrom(azureType); got != want {
			t.Errorf("ADF type %q: got %q, want %q", azureType, got, want)
		}
	}
	// SqlManagedInstance has no discriminator in armdatafactory v1.3.0 and
	// must stay unmapped (documented skip).
	if got := getResourceTypeFrom("SqlManagedInstance"); got != "" {
		t.Errorf("SqlManagedInstance unexpectedly mapped to %q", got)
	}
}

func TestSynapseIntegrationRuntimeTypeMapping(t *testing.T) {
	managed := armsynapse.IntegrationRuntimeTypeManaged
	selfHosted := armsynapse.IntegrationRuntimeTypeSelfHosted
	if got := synapseIntegrationRuntimeType(&armsynapse.ManagedIntegrationRuntime{Type: &managed}); got != "azurerm_synapse_integration_runtime_azure" {
		t.Errorf("managed IR type = %q", got)
	}
	if got := synapseIntegrationRuntimeType(&armsynapse.SelfHostedIntegrationRuntime{Type: &selfHosted}); got != "azurerm_synapse_integration_runtime_self_hosted" {
		t.Errorf("self-hosted IR type = %q", got)
	}
	if got := synapseIntegrationRuntimeType(nil); got != "" {
		t.Errorf("nil IR type = %q, want empty", got)
	}
}
