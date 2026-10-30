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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/communication/armcommunication"
)

// CommunicationGenerator covers Azure Communication Services.
// Skipped in this family:
//   - azurerm_email_communication_service, azurerm_email_communication_service_domain
//     and azurerm_email_communication_service_domain_sender_username: no
//     EmailServices client/List exists in the track-2 SDK (armcommunication
//     v1.0.0 covers only CommunicationServices and no armemail module is
//     published for Go).
//   - azurerm_communication_service_email_domain_association: synthetic
//     composite importer ID ("<communicationServiceId>|<emailDomainId>"),
//     with no independent ARM object or List.
type CommunicationGenerator struct {
	AzureService
}

func (g *CommunicationGenerator) InitResources() error {
	ctx := context.Background()
	client, err := armcommunication.NewServiceClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_communication_service")
			}
		}
		return nil
	}
	pager := client.NewListBySubscriptionPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, item := range page.Value {
			g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_communication_service")
		}
	}
	return nil
}
