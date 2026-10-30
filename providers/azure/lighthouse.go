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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/managedservices/armmanagedservices"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// LighthouseGenerator covers Azure Lighthouse (Microsoft.ManagedServices)
// registration definitions and assignments at the subscription scope, which
// matches both azurerm importer ID shapes.
type LighthouseGenerator struct {
	AzureService
}

func (g *LighthouseGenerator) InitResources() error {
	ctx := context.Background()
	scope := "/subscriptions/" + g.getSubscriptionID()

	definitionsClient, err := armmanagedservices.NewRegistrationDefinitionsClient(g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	defPager := definitionsClient.NewListPager(scope, nil)
	for defPager.More() {
		page, err := defPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			break
		}
		for _, item := range page.Value {
			g.appendLighthouseResource(item.ID, item.Name, "azurerm_lighthouse_definition")
		}
	}

	assignmentsClient, err := armmanagedservices.NewRegistrationAssignmentsClient(g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	assignPager := assignmentsClient.NewListPager(scope, nil)
	for assignPager.More() {
		page, err := assignPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			break
		}
		for _, item := range page.Value {
			g.appendLighthouseResource(item.ID, item.Name, "azurerm_lighthouse_assignment")
		}
	}
	return nil
}

func (g *LighthouseGenerator) appendLighthouseResource(id *string, name *string, resourceType string) {
	if id == nil || name == nil {
		return
	}
	g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
		*id,
		*name,
		resourceType,
		g.ProviderName,
		[]string{}))
}
