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
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/certificateregistration/armcertificateregistration"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/servicelinker/armservicelinker"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// WebAppsGenerator imports App Service plans, sites (classified by their
// ARM kind into the azurerm linux/windows web/function app types and Logic
// Apps Standard), slots and site child resources via the track-2
// armappservice SDK.
//
// Skipped azurerm types in this family:
//   - azurerm_api_connection: armappservice v1.0.0 has no Connections client
//     or List operation for Microsoft.Web/connections.
//   - azurerm_app_service_certificate_binding: association glue whose
//     importer ID is a composite "<hostNameBinding id>|<certificate id>";
//     there is no independent ARM resource to enumerate.
//   - azurerm_app_service_managed_certificate: managed certificates are the
//     same Microsoft.Web/certificates ARM objects enumerated as
//     azurerm_app_service_certificate; the SDK list exposes no field
//     distinguishing them.
//   - azurerm_app_service_slot_virtual_network_swift_connection and
//     azurerm_app_service_virtual_network_swift_connection: singleton
//     site config (config/virtualNetwork); the SDK exposes only
//     GetSwiftVirtualNetworkConnection(/Slot), no List.
//   - azurerm_app_service_source_control and
//     azurerm_app_service_source_control_slot: singleton per site/slot; the
//     SDK exposes only GetSourceControl(/Slot), no List, and the importer
//     ID is the site/slot ID itself.
//   - azurerm_function_app_hybrid_connection and
//     azurerm_web_app_hybrid_connection: armappservice v1.0.0 generates
//     ListHybridConnections/ListRelayServiceConnections with single-object
//     (non-collection) response types, so site hybrid connections cannot be
//     enumerated; the only hybrid-connection pager (PlansClient) returns
//     plan-scope IDs, not the per-site importer IDs.
//   - azurerm_function_app_active_slot and azurerm_web_app_active_slot:
//     singleton per app; the importer ID is the app ID itself and the SDK
//     exposes no List operation for an "active slot" resource.
//   - azurerm_function_app_flex_consumption: armappservice v1.0.0
//     (API 2022-03-01) predates Flex Consumption; its Site model has no
//     functionAppConfig field or other signal distinguishing flex apps,
//     which are enumerated as azurerm_linux_function_app.
//   - azurerm_source_control_token: subscription-level singleton
//     (/providers/Microsoft.Web/sourceControls/GitHub); the SDK
//     ListSourceControls operation lists source-control *types*, not
//     user tokens.
type WebAppsGenerator struct {
	AzureService
}

func (g *WebAppsGenerator) simple(id, name, resourceType string) (terraformutils.Resource, bool) {
	if id == "" || name == "" || resourceType == "" {
		return terraformutils.Resource{}, false
	}
	return terraformutils.NewSimpleResource(id, name, resourceType, g.ProviderName, []string{}), true
}

func lastIDSegment(id string) string {
	if idx := strings.LastIndex(id, "/"); idx >= 0 {
		return id[idx+1:]
	}
	return id
}

// webAppKindClass classifies an ARM site kind string.
type webAppKindClass struct {
	category string // "web", "function" or "logic"
	linux    bool
}

// classifySiteKind parses ARM site kinds such as "app,linux",
// "functionapp,linux,container" or "functionapp,workflowapp".
func classifySiteKind(kind string) webAppKindClass {
	k := strings.ToLower(kind)
	cls := webAppKindClass{category: "web", linux: strings.Contains(k, "linux")}
	switch {
	case strings.Contains(k, "workflowapp"):
		cls.category = "logic"
	case strings.Contains(k, "functionapp"):
		cls.category = "function"
	}
	return cls
}

func (c webAppKindClass) resourceType(slot bool) string {
	switch c.category {
	case "logic":
		if slot {
			return ""
		}
		return "azurerm_logic_app_standard"
	case "function":
		if c.linux {
			if slot {
				return "azurerm_linux_function_app_slot"
			}
			return "azurerm_linux_function_app"
		}
		if slot {
			return "azurerm_windows_function_app_slot"
		}
		return "azurerm_windows_function_app"
	default:
		if c.linux {
			if slot {
				return "azurerm_linux_web_app_slot"
			}
			return "azurerm_linux_web_app"
		}
		if slot {
			return "azurerm_windows_web_app_slot"
		}
		return "azurerm_windows_web_app"
	}
}

type webSiteRef struct {
	resourceGroup string
	name          string
	id            string
	kind          webAppKindClass
}

func (g *WebAppsGenerator) InitResources() error {
	ctx := context.Background()
	if err := g.listPlans(ctx); err != nil {
		return err
	}
	if err := g.listCertificates(ctx); err != nil {
		return err
	}
	if err := g.listCertificateOrders(ctx); err != nil {
		return err
	}
	if err := g.listEnvironments(ctx); err != nil {
		return err
	}
	sites, err := g.listSites(ctx)
	if err != nil {
		return err
	}
	for _, site := range sites {
		if err := g.listSiteChildren(ctx, site); err != nil {
			return err
		}
	}
	return g.listStaticSites(ctx)
}

func (g *WebAppsGenerator) listPlans(ctx context.Context) error {
	client, err := armappservice.NewPlansClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	mapItem := func(item *armappservice.Plan) (terraformutils.Resource, bool) {
		if item.ID == nil || item.Name == nil {
			return terraformutils.Resource{}, false
		}
		return g.simple(*item.ID, *item.Name, "azurerm_service_plan")
	}
	var resources []terraformutils.Resource
	if rg := g.getResourceGroup(); rg != "" {
		resources, err = drainPager(ctx, client.NewListByResourceGroupPager(rg, nil),
			func(p armappservice.PlansClientListByResourceGroupResponse) []*armappservice.Plan { return p.Value }, mapItem)
	} else {
		resources, err = drainPager(ctx, client.NewListPager(nil),
			func(p armappservice.PlansClientListResponse) []*armappservice.Plan { return p.Value }, mapItem)
	}
	if err != nil {
		return err
	}
	g.Resources = append(g.Resources, resources...)
	return nil
}

func (g *WebAppsGenerator) listCertificates(ctx context.Context) error {
	client, err := armappservice.NewCertificatesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	mapItem := func(item *armappservice.AppCertificate) (terraformutils.Resource, bool) {
		if item.ID == nil || item.Name == nil {
			return terraformutils.Resource{}, false
		}
		return g.simple(*item.ID, *item.Name, "azurerm_app_service_certificate")
	}
	var resources []terraformutils.Resource
	if rg := g.getResourceGroup(); rg != "" {
		resources, err = drainPager(ctx, client.NewListByResourceGroupPager(rg, nil),
			func(p armappservice.CertificatesClientListByResourceGroupResponse) []*armappservice.AppCertificate {
				return p.Value
			}, mapItem)
	} else {
		resources, err = drainPager(ctx, client.NewListPager(nil),
			func(p armappservice.CertificatesClientListResponse) []*armappservice.AppCertificate { return p.Value }, mapItem)
	}
	if err != nil {
		return err
	}
	g.Resources = append(g.Resources, resources...)
	return nil
}

func (g *WebAppsGenerator) listCertificateOrders(ctx context.Context) error {
	client, err := armcertificateregistration.NewAppServiceCertificateOrdersClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	mapItem := func(item *armcertificateregistration.AppServiceCertificateOrder) (terraformutils.Resource, bool) {
		if item.ID == nil || item.Name == nil {
			return terraformutils.Resource{}, false
		}
		return g.simple(*item.ID, *item.Name, "azurerm_app_service_certificate_order")
	}
	var resources []terraformutils.Resource
	if rg := g.getResourceGroup(); rg != "" {
		resources, err = drainPager(ctx, client.NewListByResourceGroupPager(rg, nil),
			func(p armcertificateregistration.AppServiceCertificateOrdersClientListByResourceGroupResponse) []*armcertificateregistration.AppServiceCertificateOrder {
				return p.Value
			}, mapItem)
	} else {
		resources, err = drainPager(ctx, client.NewListPager(nil),
			func(p armcertificateregistration.AppServiceCertificateOrdersClientListResponse) []*armcertificateregistration.AppServiceCertificateOrder {
				return p.Value
			}, mapItem)
	}
	if err != nil {
		return err
	}
	g.Resources = append(g.Resources, resources...)
	return nil
}

func (g *WebAppsGenerator) listEnvironments(ctx context.Context) error {
	client, err := armappservice.NewEnvironmentsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	mapItem := func(item *armappservice.EnvironmentResource) (terraformutils.Resource, bool) {
		if item.ID == nil || item.Name == nil || item.Kind == nil {
			return terraformutils.Resource{}, false
		}
		// Only ASEv3 maps to azurerm_app_service_environment_v3;
		// v1/v2 environments use the legacy (non-v3) resource types.
		if !strings.EqualFold(*item.Kind, "ASEV3") {
			return terraformutils.Resource{}, false
		}
		return g.simple(*item.ID, *item.Name, "azurerm_app_service_environment_v3")
	}
	var resources []terraformutils.Resource
	if rg := g.getResourceGroup(); rg != "" {
		resources, err = drainPager(ctx, client.NewListByResourceGroupPager(rg, nil),
			func(p armappservice.EnvironmentsClientListByResourceGroupResponse) []*armappservice.EnvironmentResource {
				return p.Value
			}, mapItem)
	} else {
		resources, err = drainPager(ctx, client.NewListPager(nil),
			func(p armappservice.EnvironmentsClientListResponse) []*armappservice.EnvironmentResource {
				return p.Value
			}, mapItem)
	}
	if err != nil {
		return err
	}
	g.Resources = append(g.Resources, resources...)
	return nil
}

func (g *WebAppsGenerator) listSites(ctx context.Context) ([]webSiteRef, error) {
	client, err := armappservice.NewWebAppsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	var sites []*armappservice.Site
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return nil, err
			}
			sites = append(sites, page.Value...)
		}
	} else {
		pager := client.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return nil, err
			}
			sites = append(sites, page.Value...)
		}
	}
	var refs []webSiteRef
	for _, site := range sites {
		if site.ID == nil || site.Name == nil {
			continue
		}
		parsed, err := ParseAzureResourceID(*site.ID)
		if err != nil {
			continue
		}
		kind := ""
		if site.Kind != nil {
			kind = *site.Kind
		}
		cls := classifySiteKind(kind)
		refs = append(refs, webSiteRef{resourceGroup: parsed.ResourceGroup, name: *site.Name, id: *site.ID, kind: cls})
		if res, ok := g.simple(*site.ID, *site.Name, cls.resourceType(false)); ok {
			g.Resources = append(g.Resources, res)
		}
	}
	return refs, nil
}

func (g *WebAppsGenerator) listSiteChildren(ctx context.Context, site webSiteRef) error {
	client, err := armappservice.NewWebAppsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}

	// Hostname bindings of the site itself.
	bindings, err := drainPager(ctx, client.NewListHostNameBindingsPager(site.resourceGroup, site.name, nil),
		func(p armappservice.WebAppsClientListHostNameBindingsResponse) []*armappservice.HostNameBinding {
			return p.Value
		},
		func(item *armappservice.HostNameBinding) (terraformutils.Resource, bool) {
			if item.ID == nil || item.Name == nil {
				return terraformutils.Resource{}, false
			}
			return g.simple(*item.ID, *item.Name, "azurerm_app_service_custom_hostname_binding")
		})
	if err != nil {
		return err
	}
	g.Resources = append(g.Resources, bindings...)

	// Public certificates of the site.
	publicCerts, err := drainPager(ctx, client.NewListPublicCertificatesPager(site.resourceGroup, site.name, nil),
		func(p armappservice.WebAppsClientListPublicCertificatesResponse) []*armappservice.PublicCertificate {
			return p.Value
		},
		func(item *armappservice.PublicCertificate) (terraformutils.Resource, bool) {
			if item.ID == nil || item.Name == nil {
				return terraformutils.Resource{}, false
			}
			return g.simple(*item.ID, *item.Name, "azurerm_app_service_public_certificate")
		})
	if err != nil {
		return err
	}
	g.Resources = append(g.Resources, publicCerts...)

	// Functions of function apps.
	if site.kind.category == "function" {
		functions, err := drainPager(ctx, client.NewListFunctionsPager(site.resourceGroup, site.name, nil),
			func(p armappservice.WebAppsClientListFunctionsResponse) []*armappservice.FunctionEnvelope {
				return p.Value
			},
			func(item *armappservice.FunctionEnvelope) (terraformutils.Resource, bool) {
				if item.ID == nil || item.Name == nil {
					return terraformutils.Resource{}, false
				}
				return g.simple(*item.ID, lastIDSegment(*item.ID), "azurerm_function_app_function")
			})
		if err != nil {
			return err
		}
		g.Resources = append(g.Resources, functions...)
	}

	// ServiceLinker connections of web/function apps.
	if site.kind.category == "web" || site.kind.category == "function" {
		if err := g.listSiteConnections(ctx, site); err != nil {
			return err
		}
	}

	// Slots and their hostname bindings.
	pager := client.NewListSlotsPager(site.resourceGroup, site.name, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, slot := range page.Value {
			if slot.ID == nil || slot.Name == nil {
				continue
			}
			slotName := lastIDSegment(*slot.ID)
			kind := ""
			if slot.Kind != nil {
				kind = *slot.Kind
			}
			cls := classifySiteKind(kind)
			if res, ok := g.simple(*slot.ID, slotName, cls.resourceType(true)); ok {
				g.Resources = append(g.Resources, res)
			}
			slotBindings, err := drainPager(ctx, client.NewListHostNameBindingsSlotPager(site.resourceGroup, site.name, slotName, nil),
				func(p armappservice.WebAppsClientListHostNameBindingsSlotResponse) []*armappservice.HostNameBinding {
					return p.Value
				},
				func(item *armappservice.HostNameBinding) (terraformutils.Resource, bool) {
					if item.ID == nil || item.Name == nil {
						return terraformutils.Resource{}, false
					}
					return g.simple(*item.ID, *item.Name, "azurerm_app_service_slot_custom_hostname_binding")
				})
			if err != nil {
				return err
			}
			g.Resources = append(g.Resources, slotBindings...)
		}
	}
	return nil
}

func (g *WebAppsGenerator) listSiteConnections(ctx context.Context, site webSiteRef) error {
	client, err := armservicelinker.NewLinkerClient(g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	resourceType := "azurerm_app_service_connection"
	if site.kind.category == "function" {
		resourceType = "azurerm_function_app_connection"
	}
	resources, err := drainPager(ctx, client.NewListPager(site.id, nil),
		func(p armservicelinker.LinkerClientListResponse) []*armservicelinker.LinkerResource { return p.Value },
		func(item *armservicelinker.LinkerResource) (terraformutils.Resource, bool) {
			if item.ID == nil || item.Name == nil {
				return terraformutils.Resource{}, false
			}
			return g.simple(*item.ID, *item.Name, resourceType)
		})
	if err != nil {
		return err
	}
	g.Resources = append(g.Resources, resources...)
	return nil
}

func (g *WebAppsGenerator) listStaticSites(ctx context.Context) error {
	client, err := armappservice.NewStaticSitesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	var sites []*armappservice.StaticSiteARMResource
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewGetStaticSitesByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			sites = append(sites, page.Value...)
		}
	} else {
		pager := client.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			sites = append(sites, page.Value...)
		}
	}
	for _, site := range sites {
		if site.ID == nil || site.Name == nil {
			continue
		}
		if res, ok := g.simple(*site.ID, *site.Name, "azurerm_static_web_app"); ok {
			g.Resources = append(g.Resources, res)
		}
		parsed, err := ParseAzureResourceID(*site.ID)
		if err != nil {
			continue
		}
		domains, err := drainPager(ctx, client.NewListStaticSiteCustomDomainsPager(parsed.ResourceGroup, *site.Name, nil),
			func(p armappservice.StaticSitesClientListStaticSiteCustomDomainsResponse) []*armappservice.StaticSiteCustomDomainOverviewARMResource {
				return p.Value
			},
			func(item *armappservice.StaticSiteCustomDomainOverviewARMResource) (terraformutils.Resource, bool) {
				if item.ID == nil || item.Name == nil {
					return terraformutils.Resource{}, false
				}
				return g.simple(*item.ID, *item.Name, "azurerm_static_web_app_custom_domain")
			})
		if err != nil {
			return err
		}
		g.Resources = append(g.Resources, domains...)
		registrations, err := drainPager(ctx, client.NewGetUserProvidedFunctionAppsForStaticSitePager(parsed.ResourceGroup, *site.Name, nil),
			func(p armappservice.StaticSitesClientGetUserProvidedFunctionAppsForStaticSiteResponse) []*armappservice.StaticSiteUserProvidedFunctionAppARMResource {
				return p.Value
			},
			func(item *armappservice.StaticSiteUserProvidedFunctionAppARMResource) (terraformutils.Resource, bool) {
				if item.ID == nil || item.Name == nil {
					return terraformutils.Resource{}, false
				}
				return g.simple(*item.ID, *item.Name, "azurerm_static_web_app_function_app_registration")
			})
		if err != nil {
			return err
		}
		g.Resources = append(g.Resources, registrations...)
	}
	return nil
}
