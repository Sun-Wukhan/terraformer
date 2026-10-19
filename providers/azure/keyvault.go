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
		return err
	}
	g.Resources, err = g.createResources(ctx, vaultsClient)
	return err
}
