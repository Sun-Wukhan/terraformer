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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/msi/armmsi"
)

// FederatedIdentityCredentialGenerator enumerates user-assigned managed
// identities and, under each, their federated identity credentials
// (Microsoft.ManagedIdentity/userAssignedIdentities/federatedIdentityCredentials).
type FederatedIdentityCredentialGenerator struct {
	AzureService
}

func (g *FederatedIdentityCredentialGenerator) InitResources() error {
	ctx := context.Background()
	uaiClient, err := armmsi.NewUserAssignedIdentitiesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	ficClient, err := armmsi.NewFederatedIdentityCredentialsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}

	var identities []*armmsi.Identity
	if rg := g.getResourceGroup(); rg != "" {
		pager := uaiClient.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			identities = append(identities, page.Value...)
		}
	} else {
		pager := uaiClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			identities = append(identities, page.Value...)
		}
	}

	for _, identity := range identities {
		parsed, err := ParseAzureResourceID(*identity.ID)
		if err != nil {
			return err
		}
		pager := ficClient.NewListPager(parsed.ResourceGroup, *identity.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_federated_identity_credential")
			}
		}
	}
	return nil
}
