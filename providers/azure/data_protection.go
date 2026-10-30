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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dataprotection/armdataprotection"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// DataProtectionGenerator covers Azure Backup vaults (Microsoft.DataProtection)
// via armdataprotection. Backup instances and policies are a single ARM
// resource each; the provider splits them into per-datasource resources,
// discriminated by the datasource type below (verified against the
// provider's own create payloads).
//
// Skipped in this family:
//   - azurerm_data_protection_backup_vault_customer_managed_key: a
//     singleton setting on the vault itself (its provider importer ID is
//     the vault ID); armdataprotection exposes no separate CMK resource.
type DataProtectionGenerator struct {
	AzureService
}

// Only the datasource types that have a backup-instance resource in this
// sweep are mapped; any other datasource type is left unmapped.
var dataProtectionDatasourceTypes = map[string]string{
	"Microsoft.Storage/storageAccounts/blobServices":     "azurerm_data_protection_backup_instance_blob_storage",
	"Microsoft.Storage/storageAccounts/adlsBlobServices": "azurerm_data_protection_backup_instance_data_lake_storage",
	"Microsoft.Compute/disks":                            "azurerm_data_protection_backup_instance_disk",
	"Microsoft.ContainerService/managedClusters":         "azurerm_data_protection_backup_instance_kubernetes_cluster",
	"Microsoft.DBforMySQL/flexibleServers":               "azurerm_data_protection_backup_instance_mysql_flexible_server",
	"Microsoft.DBforPostgreSQL/flexibleServers":          "azurerm_data_protection_backup_instance_postgresql_flexible_server",
}

// dataProtectionInstanceType maps a backup instance's datasource type to
// the provider resource type, or "" when this sweep has no resource for it.
func dataProtectionInstanceType(datasourceType *string) string {
	if datasourceType == nil {
		return ""
	}
	return dataProtectionDatasourceTypes[*datasourceType]
}

// dataProtectionPolicyType maps a backup policy's datasource types to the
// provider resource type. Cosmos DB account policies are the
// databaseAccounts entry, which has no backup-instance resource in this
// cluster (the account instance type is not in the sweep list).
func dataProtectionPolicyType(datasourceTypes []*string) string {
	for _, datasourceType := range datasourceTypes {
		if datasourceType == nil {
			continue
		}
		switch *datasourceType {
		case "Microsoft.Storage/storageAccounts/blobServices":
			return "azurerm_data_protection_backup_policy_blob_storage"
		case "Microsoft.Storage/storageAccounts/adlsBlobServices":
			return "azurerm_data_protection_backup_policy_data_lake_storage"
		case "Microsoft.Compute/disks":
			return "azurerm_data_protection_backup_policy_disk"
		case "Microsoft.ContainerService/managedClusters":
			return "azurerm_data_protection_backup_policy_kubernetes_cluster"
		case "Microsoft.DBforMySQL/flexibleServers":
			return "azurerm_data_protection_backup_policy_mysql_flexible_server"
		case "Microsoft.DBforPostgreSQL/flexibleServers":
			return "azurerm_data_protection_backup_policy_postgresql_flexible_server"
		case "Microsoft.DocumentDB/databaseAccounts":
			return "azurerm_data_protection_backup_policy_cosmosdb_account"
		}
	}
	return ""
}

func (g *DataProtectionGenerator) listVaultsAndChildren() ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	vaultsClient, err := armdataprotection.NewBackupVaultsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	instancesClient, err := armdataprotection.NewBackupInstancesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	policiesClient, err := armdataprotection.NewBackupPoliciesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}

	processVaults := func(vaults []*armdataprotection.BackupVaultResource) error {
		for _, vault := range vaults {
			resources = append(resources, terraformutils.NewSimpleResource(
				*vault.ID, *vault.Name, "azurerm_data_protection_backup_vault", g.ProviderName, []string{}))
			id, err := ParseAzureResourceID(*vault.ID)
			if err != nil {
				return err
			}

			instancesPager := instancesClient.NewListPager(id.ResourceGroup, *vault.Name, nil)
			for instancesPager.More() {
				page, err := instancesPager.NextPage(ctx)
				if err != nil {
					log.Println(err)
					return err
				}
				for _, instance := range page.Value {
					if instance.Properties == nil || instance.Properties.DataSourceInfo == nil {
						continue
					}
					resourceType := dataProtectionInstanceType(instance.Properties.DataSourceInfo.DatasourceType)
					if resourceType == "" {
						continue
					}
					resources = append(resources, terraformutils.NewSimpleResource(
						*instance.ID, *instance.Name, resourceType, g.ProviderName, []string{}))
				}
			}

			policiesPager := policiesClient.NewListPager(id.ResourceGroup, *vault.Name, nil)
			for policiesPager.More() {
				page, err := policiesPager.NextPage(ctx)
				if err != nil {
					log.Println(err)
					return err
				}
				for _, policy := range page.Value {
					policyProperties := policy.Properties.GetBaseBackupPolicy()
					if policyProperties == nil {
						continue
					}
					resourceType := dataProtectionPolicyType(policyProperties.DatasourceTypes)
					if resourceType == "" {
						continue
					}
					resources = append(resources, terraformutils.NewSimpleResource(
						*policy.ID, *policy.Name, resourceType, g.ProviderName, []string{}))
				}
			}
		}
		return nil
	}

	if rg := g.getResourceGroup(); rg != "" {
		pager := vaultsClient.NewGetInResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			if err := processVaults(page.Value); err != nil {
				return nil, err
			}
		}
	} else {
		pager := vaultsClient.NewGetInSubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			if err := processVaults(page.Value); err != nil {
				return nil, err
			}
		}
	}

	return resources, nil
}

func (g *DataProtectionGenerator) listResourceGuards() ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()
	client, err := armdataprotection.NewResourceGuardsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	process := func(guards []*armdataprotection.ResourceGuardResource) {
		for _, guard := range guards {
			resources = append(resources, terraformutils.NewSimpleResource(
				*guard.ID, *guard.Name, "azurerm_data_protection_resource_guard", g.ProviderName, []string{}))
		}
	}
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewGetResourcesInResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			process(page.Value)
		}
	} else {
		pager := client.NewGetResourcesInSubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			process(page.Value)
		}
	}
	return resources, nil
}

func (g *DataProtectionGenerator) InitResources() error {
	functions := []func() ([]terraformutils.Resource, error){
		g.listVaultsAndChildren,
		g.listResourceGuards,
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
