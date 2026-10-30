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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appconfiguration/armappconfiguration"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// AppConfigurationGenerator imports App Configuration stores via the
// track-2 armappconfiguration SDK.
//
// Skipped azurerm types in this family:
//   - azurerm_app_configuration_feature, azurerm_app_configuration_key:
//     data-plane objects; the provider importer ID is the store's data-plane
//     endpoint URL (https://<store>.azconfig.io/kv/<key>?label=<label>), not
//     an ARM resource ID, so they cannot be enumerated as ARM resources.
type AppConfigurationGenerator struct {
	AzureService
}

func (g *AppConfigurationGenerator) InitResources() error {
	ctx := context.Background()
	client, err := armappconfiguration.NewConfigurationStoresClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	mapItem := func(item *armappconfiguration.ConfigurationStore) (terraformutils.Resource, bool) {
		if item.ID == nil || item.Name == nil {
			return terraformutils.Resource{}, false
		}
		return terraformutils.NewSimpleResource(*item.ID, *item.Name, "azurerm_app_configuration", g.ProviderName, []string{}), true
	}
	var resources []terraformutils.Resource
	if rg := g.getResourceGroup(); rg != "" {
		resources, err = drainPager(ctx, client.NewListByResourceGroupPager(rg, nil),
			func(page armappconfiguration.ConfigurationStoresClientListByResourceGroupResponse) []*armappconfiguration.ConfigurationStore {
				return page.Value
			}, mapItem)
	} else {
		resources, err = drainPager(ctx, client.NewListPager(nil),
			func(page armappconfiguration.ConfigurationStoresClientListResponse) []*armappconfiguration.ConfigurationStore {
				return page.Value
			}, mapItem)
	}
	if err != nil {
		return err
	}
	g.Resources = append(g.Resources, resources...)
	return nil
}
