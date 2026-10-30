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
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicessiterecovery"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// SiteRecoveryGenerator covers Azure Site Recovery via
// armrecoveryservicessiterecovery, whose clients are scoped to a single
// vault: fabrics, protection containers and their mappings, network
// mappings, replication policies, recovery plans and replicated items
// (split into Azure-to-Azure/Hyper-V and VMware by provider-specific type).
//
// Skipped in this family (see docs/azure.md "Known limits"):
//   - azurerm_site_recovery_services_vault_hyperv_site and
//     azurerm_site_recovery_hyperv_network_mapping are the same ARM fabric /
//     network-mapping objects already imported as
//     azurerm_site_recovery_fabric / azurerm_site_recovery_network_mapping;
//     armrecoveryservicessiterecovery exposes no distinct type or List.
//   - azurerm_site_recovery_hyperv_replication_policy_association and
//     azurerm_site_recovery_vmware_replication_policy_association are
//     association glue between a container mapping and a policy (no
//     independent ARM resource/List).
type SiteRecoveryGenerator struct {
	AzureService
}

// siteRecoveryPolicyType maps a replication policy's provider-specific input
// type name to the azurerm policy resource type.
func siteRecoveryPolicyType(providerInputType string) string {
	switch {
	case strings.Contains(providerInputType, "VMwareCbt") || strings.Contains(providerInputType, "InMage"):
		return "azurerm_site_recovery_vmware_replication_policy"
	case strings.Contains(providerInputType, "HyperV"):
		return "azurerm_site_recovery_hyperv_replication_policy"
	default:
		return "azurerm_site_recovery_replication_policy"
	}
}

// siteRecoveryProtectedItemType maps a replicated item's provider-specific
// details type name to the azurerm replicated-item resource type.
func siteRecoveryProtectedItemType(providerDetailsType string) string {
	if strings.Contains(providerDetailsType, "VMwareCbt") || strings.Contains(providerDetailsType, "InMage") {
		return "azurerm_site_recovery_vmware_replicated_vm"
	}
	return "azurerm_site_recovery_replicated_vm"
}

func (g *SiteRecoveryGenerator) InitResources() error {
	ctx := context.Background()
	sub := g.getSubscriptionID()
	cred := g.getCredential()
	opts := g.armClientOptions()
	vaults, _, err := listRecoveryVaults(ctx, sub, cred, opts, g.getResourceGroup())
	if err != nil {
		return err
	}
	var resources []terraformutils.Resource
	add := func(id, name, typ string) {
		if id != "" && name != "" {
			resources = append(resources, terraformutils.NewSimpleResource(id, name, typ, "azurerm", []string{}))
		}
	}

	for _, vault := range vaults {
		fabricsClient, err := armrecoveryservicessiterecovery.NewReplicationFabricsClient(vault.name, vault.rg, sub, cred, opts)
		if err != nil {
			return err
		}
		fPager := fabricsClient.NewListPager(nil)
		for fPager.More() {
			page, err := fPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_site_recovery_fabric")
				}
			}
		}

		containersClient, err := armrecoveryservicessiterecovery.NewReplicationProtectionContainersClient(vault.name, vault.rg, sub, cred, opts)
		if err != nil {
			return err
		}
		cPager := containersClient.NewListPager(nil)
		for cPager.More() {
			page, err := cPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_site_recovery_protection_container")
				}
			}
		}

		mappingsClient, err := armrecoveryservicessiterecovery.NewReplicationProtectionContainerMappingsClient(vault.name, vault.rg, sub, cred, opts)
		if err != nil {
			return err
		}
		mPager := mappingsClient.NewListPager(nil)
		for mPager.More() {
			page, err := mPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_site_recovery_protection_container_mapping")
				}
			}
		}

		networkMappingsClient, err := armrecoveryservicessiterecovery.NewReplicationNetworkMappingsClient(vault.name, vault.rg, sub, cred, opts)
		if err != nil {
			return err
		}
		nPager := networkMappingsClient.NewListPager(nil)
		for nPager.More() {
			page, err := nPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_site_recovery_network_mapping")
				}
			}
		}

		policiesClient, err := armrecoveryservicessiterecovery.NewReplicationPoliciesClient(vault.name, vault.rg, sub, cred, opts)
		if err != nil {
			return err
		}
		pPager := policiesClient.NewListPager(nil)
		for pPager.More() {
			page, err := pPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID == nil || item.Name == nil {
					continue
				}
				inputType := ""
				if item.Properties != nil && item.Properties.ProviderSpecificDetails != nil {
					inputType = typeName(item.Properties.ProviderSpecificDetails)
				}
				add(*item.ID, *item.Name, siteRecoveryPolicyType(inputType))
			}
		}

		plansClient, err := armrecoveryservicessiterecovery.NewReplicationRecoveryPlansClient(vault.name, vault.rg, sub, cred, opts)
		if err != nil {
			return err
		}
		rPager := plansClient.NewListPager(nil)
		for rPager.More() {
			page, err := rPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_site_recovery_replication_recovery_plan")
				}
			}
		}

		itemsClient, err := armrecoveryservicessiterecovery.NewReplicationProtectedItemsClient(vault.name, vault.rg, sub, cred, opts)
		if err != nil {
			return err
		}
		iPager := itemsClient.NewListPager(nil)
		for iPager.More() {
			page, err := iPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID == nil || item.Name == nil {
					continue
				}
				detailsType := ""
				if item.Properties != nil && item.Properties.ProviderSpecificDetails != nil {
					detailsType = typeName(item.Properties.ProviderSpecificDetails)
				}
				add(*item.ID, *item.Name, siteRecoveryProtectedItemType(detailsType))
			}
		}
	}
	g.Resources = resources
	return nil
}

// typeName returns the %T rendering of a polymorphic SDK value.
func typeName(v interface{}) string {
	return fmt.Sprintf("%T", v)
}
