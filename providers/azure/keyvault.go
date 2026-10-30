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
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package azure

import (
	"context"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// KeyVaultGenerator covers Key Vault vaults and Managed HSMs.
// The generator also emits azurerm_key_vault_managed_hardware_security_module
// from the ARM ManagedHsms List (see InitResources).
// Skipped in this family:
//   - azurerm_key_vault_access_policy: an embedded property
//     (properties.accessPolicies) of the vault itself, imported by a
//     synthetic composite ID, with no independent ARM object or List.
//   - azurerm_key_vault_certificate, azurerm_key_vault_certificate_contacts,
//     azurerm_key_vault_certificate_issuer, azurerm_key_vault_key and
//     azurerm_key_vault_secret: data-plane resource; enumeration requires
//     per-vault data-plane clients, outside the ARM generator pattern.
//     (Their provider importer IDs are per-vault data-plane URLs, e.g.
//     https://<vault>.vault.azure.net/keys/<name>/<version>, not ARM IDs.)
//   - azurerm_key_vault_managed_hardware_security_module_key,
//     azurerm_key_vault_managed_hardware_security_module_key_rotation_policy,
//     azurerm_key_vault_managed_hardware_security_module_role_assignment,
//     azurerm_key_vault_managed_hardware_security_module_role_definition,
//     azurerm_key_vault_managed_storage_account and
//     azurerm_key_vault_managed_storage_account_sas_token_definition:
//     data-plane resource; enumeration requires per-vault data-plane
//     clients, outside the ARM generator pattern. (Their provider
//     importer IDs are per-HSM/vault data-plane URLs under
//     https://<name>.managedhsm.azure.net/ or the vault URL.)
type KeyVaultGenerator struct {
	AzureService
}

func (g KeyVaultGenerator) createResources(ctx context.Context, client *armkeyvault.VaultsClient) ([]terraformutils.Resource, error) {
	pager := client.NewListPager(nil)
	var resources []terraformutils.Resource
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return resources, err
		}
		for _, vault := range page.Value {
			resources = append(resources, terraformutils.NewSimpleResource(
				*vault.ID,
				*vault.Name,
				"azurerm_key_vault",
				"azurerm",
				[]string{}))
		}
	}
	return resources, nil
}

func (g KeyVaultGenerator) createResourcesByResourceGroup(ctx context.Context, rg string, client *armkeyvault.VaultsClient) ([]terraformutils.Resource, error) {
	pager := client.NewListByResourceGroupPager(rg, nil)
	var resources []terraformutils.Resource
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return resources, err
		}
		for _, vault := range page.Value {
			resources = append(resources, terraformutils.NewSimpleResource(
				*vault.ID,
				*vault.Name,
				"azurerm_key_vault",
				"azurerm",
				[]string{}))
		}
	}
	return resources, nil
}

func (g *KeyVaultGenerator) InitResources() error {
	ctx := context.Background()
	vaultsClient, err := armkeyvault.NewVaultsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}

	if rg := g.getResourceGroup(); rg != "" {
		g.Resources, err = g.createResourcesByResourceGroup(ctx, rg, vaultsClient)
		if err != nil {
			return err
		}
	} else {
		g.Resources, err = g.createResources(ctx, vaultsClient)
		if err != nil {
			return err
		}
	}

	mhsmClient, err := armkeyvault.NewManagedHsmsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	if rg := g.getResourceGroup(); rg != "" {
		pager := mhsmClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, mhsm := range page.Value {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					*mhsm.ID,
					*mhsm.Name,
					"azurerm_key_vault_managed_hardware_security_module",
					g.ProviderName,
					[]string{}))
			}
		}
		return nil
	}
	pager := mhsmClient.NewListBySubscriptionPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, mhsm := range page.Value {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*mhsm.ID,
				*mhsm.Name,
				"azurerm_key_vault_managed_hardware_security_module",
				g.ProviderName,
				[]string{}))
		}
	}
	return nil
}
