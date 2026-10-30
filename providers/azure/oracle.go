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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/oracledatabase/armoracledatabase"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// OracleGenerator covers Oracle.Database resources.
//
// Skipped in this service (see ledger):
//   - azurerm_oracle_resource_anchor: armoracledatabase v1.1.0 has no
//     ResourceAnchors client / List operation.
type OracleGenerator struct {
	AzureService
}

// classifyOracleAutonomousDatabase maps the polymorphic ARM properties
// type of an AutonomousDatabase (all four Terraform types share the
// Oracle.Database/autonomousDatabases ARM resource) to the Terraform
// type the azurerm provider uses for that creation mode.
func classifyOracleAutonomousDatabase(props armoracledatabase.AutonomousDatabaseBasePropertiesClassification) string {
	switch props.(type) {
	case *armoracledatabase.AutonomousDatabaseCloneProperties:
		return "azurerm_oracle_autonomous_database_clone_from_database"
	case *armoracledatabase.AutonomousDatabaseFromBackupTimestampProperties:
		return "azurerm_oracle_autonomous_database_clone_from_backup"
	case *armoracledatabase.AutonomousDatabaseCrossRegionDisasterRecoveryProperties:
		return "azurerm_oracle_autonomous_database_cross_region_disaster_recovery"
	default:
		return "azurerm_oracle_autonomous_database"
	}
}

func (g *OracleGenerator) listByRGOrSubscription(listRG func(rg string) error, listSub func() error) error {
	if rg := g.getResourceGroup(); rg != "" {
		return listRG(rg)
	}
	return listSub()
}

func (g *OracleGenerator) InitResources() error {
	ctx := context.Background()
	sub, cred, opts := g.getSubscriptionID(), g.getCredential(), g.armClientOptions()

	// Autonomous databases (+ backups per database).
	if client, err := armoracledatabase.NewAutonomousDatabasesClient(sub, cred, opts); err == nil {
		backupClient, berr := armoracledatabase.NewAutonomousDatabaseBackupsClient(sub, cred, opts)
		appendDB := func(db *armoracledatabase.AutonomousDatabase) {
			if db.ID == nil || db.Name == nil {
				return
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(*db.ID, *db.Name, classifyOracleAutonomousDatabase(db.Properties), g.ProviderName, []string{}))
			if berr != nil {
				return
			}
			parsed, err := ParseAzureResourceID(*db.ID)
			if err != nil {
				return
			}
			pager := backupClient.NewListByAutonomousDatabasePager(parsed.ResourceGroup, *db.Name, nil)
			for pager.More() {
				page, err := pager.NextPage(ctx)
				if err != nil {
					log.Println(err)
					break
				}
				for _, b := range page.Value {
					if b.ID != nil && b.Name != nil {
						g.Resources = append(g.Resources, terraformutils.NewSimpleResource(*b.ID, *b.Name, "azurerm_oracle_autonomous_database_backup", g.ProviderName, []string{}))
					}
				}
			}
		}
		err := g.listByRGOrSubscription(func(rg string) error {
			pager := client.NewListByResourceGroupPager(rg, nil)
			for pager.More() {
				page, err := pager.NextPage(ctx)
				if err != nil {
					return err
				}
				for _, db := range page.Value {
					appendDB(db)
				}
			}
			return nil
		}, func() error {
			pager := client.NewListBySubscriptionPager(nil)
			for pager.More() {
				page, err := pager.NextPage(ctx)
				if err != nil {
					return err
				}
				for _, db := range page.Value {
					appendDB(db)
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	} else {
		return err
	}

	// Cloud VM clusters.
	if client, err := armoracledatabase.NewCloudVMClustersClient(sub, cred, opts); err == nil {
		appendItem := func(item *armoracledatabase.CloudVMCluster) {
			if item.ID != nil && item.Name != nil {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(*item.ID, *item.Name, "azurerm_oracle_cloud_vm_cluster", g.ProviderName, []string{}))
			}
		}
		err := g.listByRGOrSubscription(func(rg string) error {
			pager := client.NewListByResourceGroupPager(rg, nil)
			for pager.More() {
				page, err := pager.NextPage(ctx)
				if err != nil {
					return err
				}
				for _, item := range page.Value {
					appendItem(item)
				}
			}
			return nil
		}, func() error {
			pager := client.NewListBySubscriptionPager(nil)
			for pager.More() {
				page, err := pager.NextPage(ctx)
				if err != nil {
					return err
				}
				for _, item := range page.Value {
					appendItem(item)
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	// Cloud Exadata infrastructures.
	if client, err := armoracledatabase.NewCloudExadataInfrastructuresClient(sub, cred, opts); err == nil {
		appendItem := func(item *armoracledatabase.CloudExadataInfrastructure) {
			if item.ID != nil && item.Name != nil {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(*item.ID, *item.Name, "azurerm_oracle_exadata_infrastructure", g.ProviderName, []string{}))
			}
		}
		err := g.listByRGOrSubscription(func(rg string) error {
			pager := client.NewListByResourceGroupPager(rg, nil)
			for pager.More() {
				page, err := pager.NextPage(ctx)
				if err != nil {
					return err
				}
				for _, item := range page.Value {
					appendItem(item)
				}
			}
			return nil
		}, func() error {
			pager := client.NewListBySubscriptionPager(nil)
			for pager.More() {
				page, err := pager.NextPage(ctx)
				if err != nil {
					return err
				}
				for _, item := range page.Value {
					appendItem(item)
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	// Exascale DB storage vaults.
	if client, err := armoracledatabase.NewExascaleDbStorageVaultsClient(sub, cred, opts); err == nil {
		appendItem := func(item *armoracledatabase.ExascaleDbStorageVault) {
			if item.ID != nil && item.Name != nil {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(*item.ID, *item.Name, "azurerm_oracle_exascale_database_storage_vault", g.ProviderName, []string{}))
			}
		}
		err := g.listByRGOrSubscription(func(rg string) error {
			pager := client.NewListByResourceGroupPager(rg, nil)
			for pager.More() {
				page, err := pager.NextPage(ctx)
				if err != nil {
					return err
				}
				for _, item := range page.Value {
					appendItem(item)
				}
			}
			return nil
		}, func() error {
			pager := client.NewListBySubscriptionPager(nil)
			for pager.More() {
				page, err := pager.NextPage(ctx)
				if err != nil {
					return err
				}
				for _, item := range page.Value {
					appendItem(item)
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}
