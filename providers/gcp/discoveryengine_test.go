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
	"testing"

	discoveryenginev1 "google.golang.org/api/discoveryengine/v1"
	discoveryenginev1alpha "google.golang.org/api/discoveryengine/v1alpha"
	"google.golang.org/api/option"
)

// TestDiscoveryEngineGenerator feeds canned Discovery Engine
// responses (global location populated; us/eu empty) through the
// generator and verifies the emitted types and IDs: a data store
// with schema, control, serving config, sitemap and target site;
// three engines discriminated by solution type (search/chat/
// recommendation) with controls, a serving config, widget configs,
// an assistant and the search engine's combined IAM resource; the
// data connector singleton; CMEK and license configs; a user store;
// and the ACL config singleton (the mock answers every probe with
// 200, so it is emitted for all three locations).
func TestDiscoveryEngineGenerator(t *testing.T) {
	proj := "projects/test-project"
	loc := proj + "/locations/global"
	coll := loc + "/collections/default_collection"
	ds := coll + "/dataStores/ds1"
	search := coll + "/engines/e-search"
	chat := coll + "/engines/e-chat"
	rec := coll + "/engines/e-rec"
	transport := &cannedTransport{routes: []cannedRoute{
		{":getIamPolicy", `{"bindings":[{"role":"roles/discoveryengine.viewer","members":["user:a@example.com"]}]}`},
		// Child collections before their parents' list paths, which
		// are prefixes of them.
		{"/dataStores/ds1/schemas", `{"schemas":[{"name":"` + ds + `/schemas/sc1"}]}`},
		{"/dataStores/ds1/controls", `{"controls":[{"name":"` + ds + `/controls/ctl1"}]}`},
		{"/dataStores/ds1/servingConfigs", `{"servingConfigs":[{"name":"` + ds + `/servingConfigs/svc1"}]}`},
		{"sitemaps:fetch", `{"sitemapsMetadata":[{"sitemap":{"name":"` + ds + `/siteSearchEngine/sitemaps/sm1"}}]}`},
		{"/siteSearchEngine/targetSites", `{"targetSites":[{"name":"` + ds + `/siteSearchEngine/targetSites/ts1"}]}`},
		{"/engines/e-search/controls", `{"controls":[{"name":"` + search + `/controls/ctl-e"}]}`},
		{"/engines/e-search/servingConfigs", `{"servingConfigs":[{"name":"` + search + `/servingConfigs/svc-e"}]}`},
		{"/engines/e-search/assistants", `{"assistants":[{"name":"` + search + `/assistants/a1"}]}`},
		{"/engines/e-chat/assistants", `{"assistants":[]}`},
		{"/engines/e-rec/assistants", `{"assistants":[]}`},
		{"/collections/default_collection/dataStores", `{"dataStores":[{"name":"` + ds + `"}]}`},
		{"/collections/default_collection/engines", `{"engines":[{"name":"` + search + `","solutionType":"SOLUTION_TYPE_SEARCH"},{"name":"` + chat + `","solutionType":"SOLUTION_TYPE_CHAT"},{"name":"` + rec + `","solutionType":"SOLUTION_TYPE_RECOMMENDATION"}]}`},
		{"/locations/global/cmekConfigs", `{"cmekConfigs":[{"name":"` + loc + `/cmekConfigs/ck1"}]}`},
		{"/collections/default_collection/dataConnector", `{"name":"` + coll + `/dataConnector"}`},
		{"/locations/global/licenseConfigs", `{"licenseConfigs":[{"name":"` + loc + `/licenseConfigs/lc1"}]}`},
		{"/locations/global/userStores", `{"userStores":[{"name":"` + loc + `/userStores/us1"}]}`},
	}}

	ctx := context.Background()
	clientOpts := []option.ClientOption{
		option.WithHTTPClient(transport.client()),
		option.WithoutAuthentication(),
	}
	v1Service, err := discoveryenginev1.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("discoveryengine.NewService: %v", err)
	}
	v1alphaService, err := discoveryenginev1alpha.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("discoveryengine/v1alpha.NewService: %v", err)
	}

	g := &DiscoveryEngineGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
	})
	g.initResources(ctx, v1Service, v1alphaService, transport.client())

	expected := map[string][]string{
		"google_discovery_engine_data_store":            {ds},
		"google_discovery_engine_schema":                {ds + "/schemas/sc1"},
		"google_discovery_engine_control":               {ds + "/controls/ctl1", search + "/controls/ctl-e"},
		"google_discovery_engine_serving_config":        {ds + "/servingConfigs/svc1", search + "/servingConfigs/svc-e"},
		"google_discovery_engine_sitemap":               {ds + "/siteSearchEngine/sitemaps/sm1"},
		"google_discovery_engine_target_site":           {ds + "/siteSearchEngine/targetSites/ts1"},
		"google_discovery_engine_search_engine":         {search},
		"google_discovery_engine_chat_engine":           {chat},
		"google_discovery_engine_recommendation_engine": {rec},
		"google_discovery_engine_widget_config": {
			search + "/widgetConfigs/default_search_widget_config",
			chat + "/widgetConfigs/default_search_widget_config",
			rec + "/widgetConfigs/default_search_widget_config",
		},
		"google_discovery_engine_assistant":         {search + "/assistants/a1"},
		"google_discovery_engine_search_engine_iam": {search},
		"google_discovery_engine_cmek_config":       {loc + "/cmekConfigs/ck1"},
		"google_discovery_engine_data_connector":    {coll + "/dataConnector"},
		"google_discovery_engine_license_config":    {loc + "/licenseConfigs/lc1"},
		"google_discovery_engine_user_store":        {loc + "/userStores/us1"},
		"google_discovery_engine_acl_config": {
			proj + "/locations/global/aclConfig",
			proj + "/locations/us/aclConfig",
			proj + "/locations/eu/aclConfig",
		},
	}
	assertResourceIDs(t, g.Resources, expected)
}
