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

	appenginev1 "google.golang.org/api/appengine/v1"
	"google.golang.org/api/option"
)

// TestAppEngineGenerator feeds canned App Engine responses and
// verifies the application, dispatch rules, domain mappings,
// firewall rules, service singletons and the standard/flexible
// version split.
func TestAppEngineGenerator(t *testing.T) {
	app := "apps/test-project"
	service := app + "/services/default"
	transport := &cannedTransport{routes: []cannedRoute{
		{"/services/default/versions", `{"versions":[{"id":"v1","env":"standard"},{"id":"v2","env":"flexible"}]}`},
		{"/domainMappings", `{"domainMappings":[{"name":"` + app + `/domainMappings/example.com","id":"example.com"}]}`},
		{"/firewall/ingressRules", `{"ingressRules":[{"priority":1000,"action":"ALLOW","sourceRange":"*"}]}`},
		{"/services", `{"services":[{"id":"default","networkSettings":{"ingressTrafficAllowed":"INGRESS_TRAFFIC_ALLOWED_ALL"},"split":{"allocations":{"v1":1}}}]}`},
		{"/v1/apps/test-project", `{"name":"` + app + `","id":"test-project","dispatchRules":[{"domain":"*","path":"/*","service":"default"}]}`},
	}}
	ctx := context.Background()
	clientOpts := []option.ClientOption{option.WithHTTPClient(transport.client()), option.WithoutAuthentication()}
	aeService, err := appenginev1.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("appengine.NewService: %v", err)
	}
	g := &AppEngineGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project"})
	if err := g.initResources(ctx, aeService); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_app_engine_application":                    {"test-project"},
		"google_app_engine_application_url_dispatch_rules": {"test-project"},
		"google_app_engine_domain_mapping":                 {app + "/domainMappings/example.com"},
		"google_app_engine_firewall_rule":                  {app + "/firewall/ingressRules/1000"},
		"google_app_engine_service_network_settings":       {service},
		"google_app_engine_service_split_traffic":          {service},
		"google_app_engine_standard_app_version":           {service + "/versions/v1"},
		"google_app_engine_flexible_app_version":           {service + "/versions/v2"},
	})
}
