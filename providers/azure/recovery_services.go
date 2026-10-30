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
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservices"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// RecoveryServicesGenerator covers Recovery Services vaults and Azure Backup
// via armrecoveryservices + armrecoveryservicesbackup: vaults, Resource Guard
// associations, backup policies (split by BackupManagementType) and protected
// items (split by workload), plus storage-account backup containers.
type RecoveryServicesGenerator struct {
	AzureService
}

type recoveryVaultRef struct {
	rg   string
	name string
}

func listRecoveryVaults(ctx context.Context, sub string, cred azcore.TokenCredential, opts *arm.ClientOptions, rgFilter string) ([]recoveryVaultRef, []terraformutils.Resource, error) {
	client, err := armrecoveryservices.NewVaultsClient(sub, cred, opts)
	if err != nil {
		return nil, nil, err
	}
	var vaults []recoveryVaultRef
	var resources []terraformutils.Resource
	handle := func(v *armrecoveryservices.Vault, rg string) {
		if v.ID == nil || v.Name == nil {
			return
		}
		resources = append(resources, terraformutils.NewSimpleResource(*v.ID, *v.Name, "azurerm_recovery_services_vault", "azurerm", []string{}))
		vaultRG := rg
		if vaultRG == "" {
			if parsed, err := ParseAzureResourceID(*v.ID); err == nil {
				vaultRG = parsed.ResourceGroup
			}
		}
		vaults = append(vaults, recoveryVaultRef{rg: vaultRG, name: *v.Name})
	}
	if rgFilter != "" {
		pager := client.NewListByResourceGroupPager(rgFilter, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return nil, nil, err
			}
			for _, v := range page.Value {
				handle(v, rgFilter)
			}
		}
	} else {
		pager := client.NewListBySubscriptionIDPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return nil, nil, err
			}
			for _, v := range page.Value {
				handle(v, "")
			}
		}
	}
	return vaults, resources, nil
}

// backupPolicyType maps a backup policy's concrete armrecoveryservicesbackup
// type (or its BackupManagementType) to the azurerm policy resource type.
func backupPolicyType(concreteType, backupManagementType string) string {
	switch backupManagementType {
	case string(armrecoveryservicesbackup.BackupManagementTypeAzureIaasVM):
		return "azurerm_backup_policy_vm"
	case string(armrecoveryservicesbackup.BackupManagementTypeAzureStorage):
		return "azurerm_backup_policy_file_share"
	case string(armrecoveryservicesbackup.BackupManagementTypeAzureWorkload):
		return "azurerm_backup_policy_vm_workload"
	}
	switch concreteType {
	case "*armrecoveryservicesbackup.AzureIaaSVMProtectionPolicy":
		return "azurerm_backup_policy_vm"
	case "*armrecoveryservicesbackup.AzureFileShareProtectionPolicy":
		return "azurerm_backup_policy_file_share"
	case "*armrecoveryservicesbackup.AzureVMWorkloadProtectionPolicy":
		return "azurerm_backup_policy_vm_workload"
	}
	return ""
}

// backupProtectedItemType maps a protected item's BackupManagementType to
// the azurerm protected-item resource type covered by this generator.
func backupProtectedItemType(backupManagementType string) string {
	switch backupManagementType {
	case string(armrecoveryservicesbackup.BackupManagementTypeAzureIaasVM):
		return "azurerm_backup_protected_vm"
	case string(armrecoveryservicesbackup.BackupManagementTypeAzureStorage):
		return "azurerm_backup_protected_file_share"
	}
	return ""
}

func (g *RecoveryServicesGenerator) InitResources() error {
	ctx := context.Background()
	sub := g.getSubscriptionID()
	cred := g.getCredential()
	opts := g.armClientOptions()
	vaults, resources, err := listRecoveryVaults(ctx, sub, cred, opts, g.getResourceGroup())
	if err != nil {
		return err
	}
	add := func(id, name, typ string) {
		if id != "" && name != "" && typ != "" {
			resources = append(resources, terraformutils.NewSimpleResource(id, name, typ, "azurerm", []string{}))
		}
	}

	policiesClient, err := armrecoveryservicesbackup.NewBackupPoliciesClient(sub, cred, opts)
	if err != nil {
		return err
	}
	itemsClient, err := armrecoveryservicesbackup.NewBackupProtectedItemsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	containersClient, err := armrecoveryservicesbackup.NewBackupProtectionContainersClient(sub, cred, opts)
	if err != nil {
		return err
	}
	guardClient, err := armrecoveryservicesbackup.NewResourceGuardProxiesClient(sub, cred, opts)
	if err != nil {
		return err
	}

	for _, vault := range vaults {
		pPager := policiesClient.NewListPager(vault.name, vault.rg, nil)
		for pPager.More() {
			page, err := pPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID == nil || item.Name == nil || item.Properties == nil {
					continue
				}
				base := item.Properties.GetProtectionPolicy()
				mgmt := ""
				if base != nil && base.BackupManagementType != nil {
					mgmt = *base.BackupManagementType
				}
				add(*item.ID, *item.Name, backupPolicyType(fmt.Sprintf("%T", item.Properties), mgmt))
			}
		}
		iPager := itemsClient.NewListPager(vault.name, vault.rg, nil)
		for iPager.More() {
			page, err := iPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID == nil || item.Name == nil || item.Properties == nil {
					continue
				}
				base := item.Properties.GetProtectedItem()
				var mgmt armrecoveryservicesbackup.BackupManagementType
				if base != nil && base.BackupManagementType != nil {
					mgmt = *base.BackupManagementType
				}
				add(*item.ID, *item.Name, backupProtectedItemType(string(mgmt)))
			}
		}
		cPager := containersClient.NewListPager(vault.name, vault.rg, nil)
		for cPager.More() {
			page, err := cPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID == nil || item.Name == nil || item.Properties == nil {
					continue
				}
				base := item.Properties.GetProtectionContainer()
				if base == nil || base.ContainerType == nil {
					continue
				}
				if *base.ContainerType == armrecoveryservicesbackup.ContainerTypeStorageContainer {
					name := *item.Name
					if base.FriendlyName != nil && *base.FriendlyName != "" {
						name = *base.FriendlyName
					}
					add(*item.ID, name, "azurerm_backup_container_storage_account")
				}
			}
		}
		gPager := guardClient.NewGetPager(vault.name, vault.rg, nil)
		for gPager.More() {
			page, err := gPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_recovery_services_vault_resource_guard_association")
				}
			}
		}
	}
	g.Resources = resources
	return nil
}
