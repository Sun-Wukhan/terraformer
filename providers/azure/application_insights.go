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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/applicationinsights/armapplicationinsights"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// ApplicationInsightsGenerator imports Application Insights components and
// their child resources via the track-2 armapplicationinsights SDK.
//
// Skipped azurerm types in this family:
//   - azurerm_application_insights_standard_web_test: standard and classic
//     web tests share the Microsoft.Insights/webTests ARM type and
//     armapplicationinsights v1.2.0 (API 2020-02-02) exposes no field
//     distinguishing standard tests (its WebTestKind enum is only
//     ping/multistep); web tests are enumerated once as
//     azurerm_application_insights_web_test.
//   - azurerm_application_insights_workbook_template: no workbook-template
//     client exists in armapplicationinsights v1.2.0.
type ApplicationInsightsGenerator struct {
	AzureService
}

type appInsightsComponentRef struct {
	resourceGroup string
	name          string
	id            string
}

func (g *ApplicationInsightsGenerator) simple(id, name, resourceType string) (terraformutils.Resource, bool) {
	if id == "" || name == "" {
		return terraformutils.Resource{}, false
	}
	return terraformutils.NewSimpleResource(id, name, resourceType, g.ProviderName, []string{}), true
}

func (g *ApplicationInsightsGenerator) listComponents(ctx context.Context) ([]appInsightsComponentRef, error) {
	client, err := armapplicationinsights.NewComponentsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	var items []*armapplicationinsights.Component
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return nil, err
			}
			items = append(items, page.Value...)
		}
	} else {
		pager := client.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return nil, err
			}
			items = append(items, page.Value...)
		}
	}
	var refs []appInsightsComponentRef
	for _, item := range items {
		if item.ID == nil || item.Name == nil {
			continue
		}
		parsed, err := ParseAzureResourceID(*item.ID)
		if err != nil {
			continue
		}
		refs = append(refs, appInsightsComponentRef{resourceGroup: parsed.ResourceGroup, name: *item.Name, id: *item.ID})
		if res, ok := g.simple(*item.ID, *item.Name, "azurerm_application_insights"); ok {
			g.Resources = append(g.Resources, res)
		}
	}
	return refs, nil
}

func (g *ApplicationInsightsGenerator) InitResources() error {
	ctx := context.Background()
	refs, err := g.listComponents(ctx)
	if err != nil {
		return err
	}
	if err := g.listAPIKeys(ctx, refs); err != nil {
		return err
	}
	if err := g.listAnalyticsItems(ctx, refs); err != nil {
		return err
	}
	if err := g.listSmartDetectionRules(ctx, refs); err != nil {
		return err
	}
	if err := g.listWebTests(ctx); err != nil {
		return err
	}
	return g.listWorkbooks(ctx)
}

func (g *ApplicationInsightsGenerator) listAPIKeys(ctx context.Context, refs []appInsightsComponentRef) error {
	client, err := armapplicationinsights.NewAPIKeysClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	for _, ref := range refs {
		resources, err := drainPager(ctx, client.NewListPager(ref.resourceGroup, ref.name, nil),
			func(p armapplicationinsights.APIKeysClientListResponse) []*armapplicationinsights.ComponentAPIKey {
				return p.Value
			},
			func(item *armapplicationinsights.ComponentAPIKey) (terraformutils.Resource, bool) {
				if item.ID == nil || item.Name == nil {
					return terraformutils.Resource{}, false
				}
				return g.simple(*item.ID, *item.Name, "azurerm_application_insights_api_key")
			})
		if err != nil {
			return err
		}
		g.Resources = append(g.Resources, resources...)
	}
	return nil
}

func (g *ApplicationInsightsGenerator) listAnalyticsItems(ctx context.Context, refs []appInsightsComponentRef) error {
	client, err := armapplicationinsights.NewAnalyticsItemsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	for _, ref := range refs {
		resp, err := client.List(ctx, ref.resourceGroup, ref.name, armapplicationinsights.ItemScopePathAnalyticsItems, nil)
		if err != nil {
			return err
		}
		for _, item := range resp.ComponentAnalyticsItemArray {
			if item.ID == nil {
				continue
			}
			name := *item.ID
			if item.Name != nil && *item.Name != "" {
				name = *item.Name
			}
			// The item ID returned by the API is the item GUID; the
			// provider importer ID nests it under the component.
			id := ref.id + "/analyticsItems/" + *item.ID
			if res, ok := g.simple(id, name, "azurerm_application_insights_analytics_item"); ok {
				g.Resources = append(g.Resources, res)
			}
		}
	}
	return nil
}

func (g *ApplicationInsightsGenerator) listSmartDetectionRules(ctx context.Context, refs []appInsightsComponentRef) error {
	client, err := armapplicationinsights.NewProactiveDetectionConfigurationsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	for _, ref := range refs {
		resp, err := client.List(ctx, ref.resourceGroup, ref.name, nil)
		if err != nil {
			return err
		}
		for _, item := range resp.ComponentProactiveDetectionConfigurationArray {
			if item.Name == nil {
				continue
			}
			// Proactive detection configs have no ARM ID of their own; the
			// provider importer ID nests the rule name under the component.
			id := ref.id + "/proactiveDetectionConfigs/" + *item.Name
			if res, ok := g.simple(id, *item.Name, "azurerm_application_insights_smart_detection_rule"); ok {
				g.Resources = append(g.Resources, res)
			}
		}
	}
	return nil
}

func (g *ApplicationInsightsGenerator) listWebTests(ctx context.Context) error {
	client, err := armapplicationinsights.NewWebTestsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	mapItem := func(item *armapplicationinsights.WebTest) (terraformutils.Resource, bool) {
		if item.ID == nil || item.Name == nil {
			return terraformutils.Resource{}, false
		}
		return g.simple(*item.ID, *item.Name, "azurerm_application_insights_web_test")
	}
	var resources []terraformutils.Resource
	if rg := g.getResourceGroup(); rg != "" {
		resources, err = drainPager(ctx, client.NewListByResourceGroupPager(rg, nil),
			func(p armapplicationinsights.WebTestsClientListByResourceGroupResponse) []*armapplicationinsights.WebTest {
				return p.Value
			}, mapItem)
	} else {
		resources, err = drainPager(ctx, client.NewListPager(nil),
			func(p armapplicationinsights.WebTestsClientListResponse) []*armapplicationinsights.WebTest {
				return p.Value
			}, mapItem)
	}
	if err != nil {
		return err
	}
	g.Resources = append(g.Resources, resources...)
	return nil
}

func (g *ApplicationInsightsGenerator) listWorkbooks(ctx context.Context) error {
	client, err := armapplicationinsights.NewWorkbooksClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	resourceGroups := []string{g.getResourceGroup()}
	if resourceGroups[0] == "" {
		// WorkbooksClient has no subscription-scope List; enumerate the
		// resource groups and list per group.
		rgClient, err := armresources.NewResourceGroupsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
		if err != nil {
			return err
		}
		resourceGroups = nil
		pager := rgClient.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, rg := range page.Value {
				if rg.Name != nil {
					resourceGroups = append(resourceGroups, *rg.Name)
				}
			}
		}
	}
	for _, rg := range resourceGroups {
		resources, err := drainPager(ctx, client.NewListByResourceGroupPager(rg, armapplicationinsights.CategoryTypeWorkbook, nil),
			func(p armapplicationinsights.WorkbooksClientListByResourceGroupResponse) []*armapplicationinsights.Workbook {
				return p.Value
			},
			func(item *armapplicationinsights.Workbook) (terraformutils.Resource, bool) {
				if item.ID == nil || item.Name == nil {
					return terraformutils.Resource{}, false
				}
				return g.simple(*item.ID, *item.Name, "azurerm_application_insights_workbook")
			})
		if err != nil {
			return err
		}
		g.Resources = append(g.Resources, resources...)
	}
	return nil
}
