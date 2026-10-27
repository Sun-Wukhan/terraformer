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

package gcp

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"

	discoveryenginev1 "google.golang.org/api/discoveryengine/v1"
	discoveryenginev1alpha "google.golang.org/api/discoveryengine/v1alpha"
)

// DiscoveryEngineGenerator enumerates the Discovery Engine (Vertex
// AI Search / Gemini Enterprise) estate of one project in the three
// locations the service exists in ("global", "us", "eu"), under the
// default collection: data stores with their schemas, controls,
// serving configs, sitemaps and target sites; engines — discriminated
// by solution type into search, chat and recommendation engines —
// with their controls, serving configs, widget configs, assistants
// and (for search engines) the combined IAM resource; the collection
// data connector singleton; CMEK configs; license configs; user
// stores; and the location ACL config singleton.
//
// Surface choices, all verified against the pinned clients and the
// provider's Magic Modules definitions:
//   - Serving configs and widget configs exist only in the v1alpha
//     client; the data connector Get likewise. Everything else uses
//     the v1 client (sitemaps have no List — the client's
//     SiteSearchEngine Fetch call is the enumeration path).
//   - The widget config is a per-engine singleton whose only
//     accepted ID is "default_search_widget_config", probed with Get.
//   - License configs, user stores, assistants and the ACL config
//     have no generated client surface and run over raw REST through
//     gcp_rest.go (v1 for license configs and assistants, v1alpha
//     for user stores and the ACL config).
//   - The provider models search-engine IAM as the single combined
//     google_discovery_engine_search_engine_iam resource, emitted
//     when a search engine's policy (raw REST :getIamPolicy) has
//     bindings.
type DiscoveryEngineGenerator struct {
	GCPService
}

const (
	discoveryEngineBaseURL      = "https://discoveryengine.googleapis.com/v1/"
	discoveryEngineAlphaBaseURL = "https://discoveryengine.googleapis.com/v1alpha/"
	discoveryEngineCollection   = "default_collection"
	discoveryEngineWidgetConfig = "default_search_widget_config"
)

var discoveryEngineLocations = []string{"global", "us", "eu"}

func (g *DiscoveryEngineGenerator) InitResources() error {
	ctx := context.Background()
	v1Service, err := discoveryenginev1.NewService(ctx)
	if err != nil {
		log.Println(err)
		v1Service = nil
	}
	v1alphaService, err := discoveryenginev1alpha.NewService(ctx)
	if err != nil {
		log.Println(err)
		v1alphaService = nil
	}
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		log.Println(err)
		restClient = nil
	}
	g.initResources(ctx, v1Service, v1alphaService, restClient)
	return nil
}

func (g *DiscoveryEngineGenerator) newDEResource(id, resourceType string) terraformutils.Resource {
	parts := strings.Split(id, "/")
	return terraformutils.NewResource(
		id,
		parts[len(parts)-1],
		resourceType,
		g.ProviderName,
		map[string]string{},
		[]string{},
		map[string]interface{}{},
	)
}

func (g *DiscoveryEngineGenerator) initResources(ctx context.Context, v1Service *discoveryenginev1.Service, v1alphaService *discoveryenginev1alpha.Service, restClient *http.Client) {
	project := g.GetArgs()["project"].(string)
	for _, location := range discoveryEngineLocations {
		locationName := "projects/" + project + "/locations/" + location
		collectionName := locationName + "/collections/" + discoveryEngineCollection
		if v1Service != nil {
			g.discoveryEngineDataStores(ctx, v1Service, v1alphaService, collectionName)
			g.discoveryEngineEngines(ctx, v1Service, v1alphaService, restClient, collectionName)
			g.discoveryEngineCmekConfigs(ctx, v1Service, locationName)
		}
		if v1alphaService != nil {
			if connector, err := v1alphaService.Projects.Locations.Collections.GetDataConnector(collectionName + "/dataConnector").Context(ctx).Do(); err == nil && connector.Name != "" {
				g.Resources = append(g.Resources, g.newDEResource(connector.Name, "google_discovery_engine_data_connector"))
			}
		}
		if restClient != nil {
			g.discoveryEngineREST(ctx, restClient, locationName)
		}
	}
}

// discoveryEngineDataStores lists one collection's data stores and
// their child collections.
func (g *DiscoveryEngineGenerator) discoveryEngineDataStores(ctx context.Context, v1Service *discoveryenginev1.Service, v1alphaService *discoveryenginev1alpha.Service, collectionName string) {
	call := v1Service.Projects.Locations.Collections.DataStores.List(collectionName)
	if err := call.Pages(ctx, func(page *discoveryenginev1.GoogleCloudDiscoveryengineV1ListDataStoresResponse) error {
		for _, dataStore := range page.DataStores {
			g.Resources = append(g.Resources, g.newDEResource(dataStore.Name, "google_discovery_engine_data_store"))
			g.discoveryEngineDataStoreChildren(ctx, v1Service, v1alphaService, dataStore.Name)
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
}

func (g *DiscoveryEngineGenerator) discoveryEngineDataStoreChildren(ctx context.Context, v1Service *discoveryenginev1.Service, v1alphaService *discoveryenginev1alpha.Service, dataStoreName string) {
	schemasCall := v1Service.Projects.Locations.Collections.DataStores.Schemas.List(dataStoreName)
	if err := schemasCall.Pages(ctx, func(page *discoveryenginev1.GoogleCloudDiscoveryengineV1ListSchemasResponse) error {
		for _, schema := range page.Schemas {
			g.Resources = append(g.Resources, g.newDEResource(schema.Name, "google_discovery_engine_schema"))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	controlsCall := v1Service.Projects.Locations.Collections.DataStores.Controls.List(dataStoreName)
	if err := controlsCall.Pages(ctx, func(page *discoveryenginev1.GoogleCloudDiscoveryengineV1ListControlsResponse) error {
		for _, control := range page.Controls {
			g.Resources = append(g.Resources, g.newDEResource(control.Name, "google_discovery_engine_control"))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	siteSearchEngine := dataStoreName + "/siteSearchEngine"
	targetSitesCall := v1Service.Projects.Locations.Collections.DataStores.SiteSearchEngine.TargetSites.List(siteSearchEngine)
	if err := targetSitesCall.Pages(ctx, func(page *discoveryenginev1.GoogleCloudDiscoveryengineV1ListTargetSitesResponse) error {
		for _, site := range page.TargetSites {
			g.Resources = append(g.Resources, g.newDEResource(site.Name, "google_discovery_engine_target_site"))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if response, err := v1Service.Projects.Locations.Collections.DataStores.SiteSearchEngine.Sitemaps.Fetch(siteSearchEngine).Context(ctx).Do(); err == nil {
		for _, metadata := range response.SitemapsMetadata {
			if metadata.Sitemap != nil {
				g.Resources = append(g.Resources, g.newDEResource(metadata.Sitemap.Name, "google_discovery_engine_sitemap"))
			}
		}
	} else {
		log.Println(err)
	}
	if v1alphaService != nil {
		servingCall := v1alphaService.Projects.Locations.Collections.DataStores.ServingConfigs.List(dataStoreName)
		if err := servingCall.Pages(ctx, func(page *discoveryenginev1alpha.GoogleCloudDiscoveryengineV1alphaListServingConfigsResponse) error {
			for _, config := range page.ServingConfigs {
				g.Resources = append(g.Resources, g.newDEResource(config.Name, "google_discovery_engine_serving_config"))
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
	}
}

// discoveryEngineEngines lists one collection's engines, maps each
// onto its solution-type resource, and enumerates engine children.
func (g *DiscoveryEngineGenerator) discoveryEngineEngines(ctx context.Context, v1Service *discoveryenginev1.Service, v1alphaService *discoveryenginev1alpha.Service, restClient *http.Client, collectionName string) {
	call := v1Service.Projects.Locations.Collections.Engines.List(collectionName)
	if err := call.Pages(ctx, func(page *discoveryenginev1.GoogleCloudDiscoveryengineV1ListEnginesResponse) error {
		for _, engine := range page.Engines {
			resourceType := ""
			switch engine.SolutionType {
			case "SOLUTION_TYPE_SEARCH":
				resourceType = "google_discovery_engine_search_engine"
			case "SOLUTION_TYPE_CHAT":
				resourceType = "google_discovery_engine_chat_engine"
			case "SOLUTION_TYPE_RECOMMENDATION":
				resourceType = "google_discovery_engine_recommendation_engine"
			default:
				log.Println("discoveryengine: unmapped engine solution type", engine.SolutionType, engine.Name)
				continue
			}
			g.Resources = append(g.Resources, g.newDEResource(engine.Name, resourceType))
			g.discoveryEngineEngineChildren(ctx, v1Service, v1alphaService, restClient, engine.Name, resourceType)
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
}

func (g *DiscoveryEngineGenerator) discoveryEngineEngineChildren(ctx context.Context, v1Service *discoveryenginev1.Service, v1alphaService *discoveryenginev1alpha.Service, restClient *http.Client, engineName, engineType string) {
	controlsCall := v1Service.Projects.Locations.Collections.Engines.Controls.List(engineName)
	if err := controlsCall.Pages(ctx, func(page *discoveryenginev1.GoogleCloudDiscoveryengineV1ListControlsResponse) error {
		for _, control := range page.Controls {
			g.Resources = append(g.Resources, g.newDEResource(control.Name, "google_discovery_engine_control"))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if v1alphaService != nil {
		servingCall := v1alphaService.Projects.Locations.Collections.Engines.ServingConfigs.List(engineName)
		if err := servingCall.Pages(ctx, func(page *discoveryenginev1alpha.GoogleCloudDiscoveryengineV1alphaListServingConfigsResponse) error {
			for _, config := range page.ServingConfigs {
				g.Resources = append(g.Resources, g.newDEResource(config.Name, "google_discovery_engine_serving_config"))
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
		widgetName := engineName + "/widgetConfigs/" + discoveryEngineWidgetConfig
		if _, err := v1alphaService.Projects.Locations.Collections.Engines.WidgetConfigs.Get(widgetName).Context(ctx).Do(); err == nil {
			g.Resources = append(g.Resources, g.newDEResource(widgetName, "google_discovery_engine_widget_config"))
		}
	}
	if restClient != nil {
		assistants, err := restListResources(ctx, restClient, discoveryEngineBaseURL+engineName+"/assistants")
		if err != nil {
			log.Println(err)
		}
		for _, assistant := range assistants {
			g.Resources = append(g.Resources, g.newDEResource(assistant.Name, "google_discovery_engine_assistant"))
		}
		if engineType == "google_discovery_engine_search_engine" {
			bindings, err := restGetIamPolicy(ctx, restClient, discoveryEngineBaseURL+engineName+":getIamPolicy")
			if err != nil {
				log.Println(err)
			} else if len(bindings) > 0 {
				g.Resources = append(g.Resources, g.newDEResource(engineName, "google_discovery_engine_search_engine_iam"))
			}
		}
	}
}

// discoveryEngineCmekConfigs lists one location's CMEK configs.
// The v1 list call has no pagination support in the client (the
// response carries no page token), so it is a single Do.
func (g *DiscoveryEngineGenerator) discoveryEngineCmekConfigs(ctx context.Context, v1Service *discoveryenginev1.Service, locationName string) {
	response, err := v1Service.Projects.Locations.CmekConfigs.List(locationName).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		return
	}
	for _, config := range response.CmekConfigs {
		g.Resources = append(g.Resources, g.newDEResource(config.Name, "google_discovery_engine_cmek_config"))
	}
}

// discoveryEngineREST covers the raw-REST-only surface of one
// location: license configs (v1), user stores (v1alpha) and the ACL
// config singleton (v1alpha, emitted when present).
func (g *DiscoveryEngineGenerator) discoveryEngineREST(ctx context.Context, restClient *http.Client, locationName string) {
	licenseConfigs, err := restListResources(ctx, restClient, discoveryEngineBaseURL+locationName+"/licenseConfigs")
	if err != nil {
		log.Println(err)
	}
	for _, config := range licenseConfigs {
		g.Resources = append(g.Resources, g.newDEResource(config.Name, "google_discovery_engine_license_config"))
	}
	userStores, err := restListResources(ctx, restClient, discoveryEngineAlphaBaseURL+locationName+"/userStores")
	if err != nil {
		log.Println(err)
	}
	for _, store := range userStores {
		g.Resources = append(g.Resources, g.newDEResource(store.Name, "google_discovery_engine_user_store"))
	}
	aclConfigName := locationName + "/aclConfig"
	if _, err := restDo(ctx, restClient, http.MethodGet, discoveryEngineAlphaBaseURL+aclConfigName, nil); err == nil {
		g.Resources = append(g.Resources, g.newDEResource(aclConfigName, "google_discovery_engine_acl_config"))
	}
}
