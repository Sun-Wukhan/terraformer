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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cognitiveservices/armcognitiveservices"
)

// CognitiveGenerator covers Cognitive Services accounts and their children.
// Skipped in this family:
//   - azurerm_cognitive_account_customer_managed_key: an embedded property
//     (properties.encryption.keySource/keyVaultProperties) of the account
//     itself, with no independent ARM object or List.
type CognitiveGenerator struct {
	AzureService
}

// cognitiveConnectionType maps the ARM connection authType discriminator to
// the azurerm resource type the provider splits connections into. It returns
// "" for auth types with no dedicated azurerm resource.
func cognitiveConnectionType(authType string) string {
	switch authType {
	case "AccountKey":
		return "azurerm_cognitive_account_connection_account_key"
	case "ManagedIdentity":
		return "azurerm_cognitive_account_connection_account_managed_identity"
	case "ApiKey":
		return "azurerm_cognitive_account_connection_api_key"
	case "CustomKeys":
		return "azurerm_cognitive_account_connection_custom_keys"
	case "AAD":
		return "azurerm_cognitive_account_connection_entra_id"
	}
	return ""
}

func (g *CognitiveGenerator) listAccounts(ctx context.Context) ([]*armcognitiveservices.Account, error) {
	client, err := armcognitiveservices.NewAccountsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	var accounts []*armcognitiveservices.Account
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return accounts, err
			}
			accounts = append(accounts, page.Value...)
		}
		return accounts, nil
	}
	pager := client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return accounts, err
		}
		accounts = append(accounts, page.Value...)
	}
	return accounts, nil
}

func (g *CognitiveGenerator) appendAccountChildren(ctx context.Context, rg, accountName string) error {
	deployClient, err := armcognitiveservices.NewDeploymentsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	pager := deployClient.NewListPager(rg, accountName, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, item := range page.Value {
			g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_cognitive_deployment")
		}
	}

	blocklistClient, err := armcognitiveservices.NewRaiBlocklistsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	blPager := blocklistClient.NewListPager(rg, accountName, nil)
	for blPager.More() {
		page, err := blPager.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, item := range page.Value {
			g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_cognitive_account_rai_blocklist")
		}
	}

	policyClient, err := armcognitiveservices.NewRaiPoliciesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	polPager := policyClient.NewListPager(rg, accountName, nil)
	for polPager.More() {
		page, err := polPager.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, item := range page.Value {
			g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_cognitive_account_rai_policy")
		}
	}

	projectClient, err := armcognitiveservices.NewProjectsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	projPager := projectClient.NewListPager(rg, accountName, nil)
	for projPager.More() {
		page, err := projPager.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, item := range page.Value {
			g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_cognitive_account_project")
		}
	}

	connClient, err := armcognitiveservices.NewAccountConnectionsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	connPager := connClient.NewListPager(rg, accountName, nil)
	for connPager.More() {
		page, err := connPager.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, item := range page.Value {
			if item.Properties == nil {
				continue
			}
			props := item.Properties.GetConnectionPropertiesV2()
			if props == nil || props.AuthType == nil {
				continue
			}
			if tfType := cognitiveConnectionType(string(*props.AuthType)); tfType != "" {
				g.AppendSimpleResource(*item.ID, *item.Name, tfType)
			}
		}
	}
	return nil
}

func (g *CognitiveGenerator) InitResources() error {
	ctx := context.Background()
	accounts, err := g.listAccounts(ctx)
	if err != nil {
		return err
	}
	for _, account := range accounts {
		g.AppendSimpleResource(*account.ID, *account.Name, "azurerm_cognitive_account")
		parsed, err := ParseAzureResourceID(*account.ID)
		if err != nil {
			return err
		}
		if err := g.appendAccountChildren(ctx, parsed.ResourceGroup, *account.Name); err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}

// AIFoundryGenerator covers Microsoft Foundry (AIServices) accounts and
// their projects. An azurerm_ai_foundry is a Cognitive Services account
// whose kind is "AIServices"; its projects share the accounts/projects ARM
// collection with azurerm_ai_foundry_project as the provider type.
type AIFoundryGenerator struct {
	AzureService
}

func (g *AIFoundryGenerator) InitResources() error {
	ctx := context.Background()
	cog := &CognitiveGenerator{}
	cog.AzureService = g.AzureService
	accounts, err := cog.listAccounts(ctx)
	if err != nil {
		return err
	}
	projectClient, err := armcognitiveservices.NewProjectsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	for _, account := range accounts {
		if account.Kind == nil || *account.Kind != "AIServices" {
			continue
		}
		g.AppendSimpleResource(*account.ID, *account.Name, "azurerm_ai_foundry")
		parsed, err := ParseAzureResourceID(*account.ID)
		if err != nil {
			return err
		}
		pager := projectClient.NewListPager(parsed.ResourceGroup, *account.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_ai_foundry_project")
			}
		}
	}
	return nil
}
