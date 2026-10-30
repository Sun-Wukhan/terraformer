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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/chaos/armchaos"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// ChaosStudioGenerator covers Azure Chaos Studio via armchaos: experiments.
//
// Skipped in this family (see docs/azure.md "Known limits"):
//   - azurerm_chaos_studio_target and azurerm_chaos_studio_capability are
//     extension resources on an arbitrary target resource; armchaos lists
//     them only per parent resource (resource group + parent provider
//     namespace/type/name), with no subscription or resource-group List.
type ChaosStudioGenerator struct {
	AzureService
}

func (g *ChaosStudioGenerator) InitResources() error {
	ctx := context.Background()
	client, err := armchaos.NewExperimentsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	var resources []terraformutils.Resource
	add := func(item *armchaos.Experiment) {
		if item.ID != nil && item.Name != nil {
			resources = append(resources, terraformutils.NewSimpleResource(*item.ID, *item.Name, "azurerm_chaos_studio_experiment", "azurerm", []string{}))
		}
	}
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				add(item)
			}
		}
	} else {
		pager := client.NewListAllPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				add(item)
			}
		}
	}
	g.Resources = resources
	return nil
}
