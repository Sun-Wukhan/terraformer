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
	"google.golang.org/api/compute/v1"
	iapv1 "google.golang.org/api/iap/v1"
	"google.golang.org/api/option"
	runv2 "google.golang.org/api/run/v2"
)

// TestIapGenerator feeds canned IAP/AppEngine/Compute/Run responses
// and asserts every IAP trio's parent ID in provider format, plus the
// settings singleton and tunnel dest groups.
func TestIapGenerator(t *testing.T) {
	ctx := context.Background()
	p := "projects/test-project"
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: ":getIamPolicy", body: `{"bindings":[{"role":"roles/iap.httpsResourceAccessor","members":["user:jane@example.com"]}]}`},
		{substr: "iap_settings", body: `{"name":"` + p + `/iap_settings"}`},
		{substr: "/destGroups", body: `{"tunnelDestGroups":[{"name":"` + p + `/locations/us-central1/destGroups/dg1"}]}`},
		{substr: "/versions", body: `{"versions":[{"id":"v1","name":"apps/test-project/services/default/versions/v1"}]}`},
		{substr: "/apps/test-project/services", body: `{"services":[{"id":"default","name":"apps/test-project/services/default"}]}`},
		{substr: "global/backendServices", body: `{"items":[{"name":"bs1"}]}`},
		{substr: "regions/us-central1/backendServices", body: `{"items":[{"name":"rbs1"}]}`},
		{substr: "global/forwardingRules", body: `{"items":[{"name":"fr1"}]}`},
		{substr: "regions/us-central1/forwardingRules", body: `{"items":[{"name":"rfr1"}]}`},
		{substr: "/instances", body: `{"items":[{"name":"inst1"}]}`},
		{substr: "/locations/us-central1/services", body: `{"services":[{"name":"` + p + `/locations/us-central1/services/svc1"}]}`},
	}}
	opts := []option.ClientOption{option.WithHTTPClient(transport.client()), option.WithoutAuthentication()}
	iapSvc, err := iapv1.NewService(ctx, opts...)
	if err != nil {
		t.Fatal(err)
	}
	aeSvc, err := appenginev1.NewService(ctx, opts...)
	if err != nil {
		t.Fatal(err)
	}
	computeSvc, err := compute.NewService(ctx, opts...)
	if err != nil {
		t.Fatal(err)
	}
	runSvc, err := runv2.NewService(ctx, opts...)
	if err != nil {
		t.Fatal(err)
	}
	g := &IapGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1", Zones: []string{"zones/us-central1-a"}},
	})
	if err := g.initResources(ctx, iapSvc, aeSvc, computeSvc, runSvc); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	trio := func(policy, binding, member, id string) map[string][]string {
		return map[string][]string{policy: {id}, binding: {id}, member: {id}}
	}
	expected := map[string][]string{}
	for _, m := range []map[string][]string{
		trio("google_iap_web_iam_policy", "google_iap_web_iam_binding", "google_iap_web_iam_member", p+"/iap_web"),
		trio("google_iap_web_type_compute_iam_policy", "google_iap_web_type_compute_iam_binding", "google_iap_web_type_compute_iam_member", p+"/iap_web/compute"),
		trio("google_iap_web_type_app_engine_iam_policy", "google_iap_web_type_app_engine_iam_binding", "google_iap_web_type_app_engine_iam_member", p+"/iap_web/appengine-test-project"),
		trio("google_iap_tunnel_iam_policy", "google_iap_tunnel_iam_binding", "google_iap_tunnel_iam_member", p+"/iap_tunnel"),
		trio("google_iap_app_engine_service_iam_policy", "google_iap_app_engine_service_iam_binding", "google_iap_app_engine_service_iam_member", p+"/iap_web/appengine-test-project/services/default"),
		trio("google_iap_app_engine_version_iam_policy", "google_iap_app_engine_version_iam_binding", "google_iap_app_engine_version_iam_member", p+"/iap_web/appengine-test-project/services/default/versions/v1"),
		trio("google_iap_web_backend_service_iam_policy", "google_iap_web_backend_service_iam_binding", "google_iap_web_backend_service_iam_member", p+"/iap_web/compute/services/bs1"),
		trio("google_iap_web_region_backend_service_iam_policy", "google_iap_web_region_backend_service_iam_binding", "google_iap_web_region_backend_service_iam_member", p+"/iap_web/compute/regions/us-central1/services/rbs1"),
		trio("google_iap_web_forwarding_rule_service_iam_policy", "google_iap_web_forwarding_rule_service_iam_binding", "google_iap_web_forwarding_rule_service_iam_member", p+"/iap_web/forwarding_rule/services/fr1"),
		trio("google_iap_web_region_forwarding_rule_service_iam_policy", "google_iap_web_region_forwarding_rule_service_iam_binding", "google_iap_web_region_forwarding_rule_service_iam_member", p+"/iap_web/forwarding_rule/regions/us-central1/services/rfr1"),
		trio("google_iap_tunnel_instance_iam_policy", "google_iap_tunnel_instance_iam_binding", "google_iap_tunnel_instance_iam_member", p+"/iap_tunnel/zones/us-central1-a/instances/inst1"),
		trio("google_iap_web_cloud_run_service_iam_policy", "google_iap_web_cloud_run_service_iam_binding", "google_iap_web_cloud_run_service_iam_member", p+"/iap_web/cloud_run-us-central1/services/svc1"),
		trio("google_iap_tunnel_dest_group_iam_policy", "google_iap_tunnel_dest_group_iam_binding", "google_iap_tunnel_dest_group_iam_member", p+"/locations/us-central1/destGroups/dg1"),
	} {
		for k, v := range m {
			expected[k] = v
		}
	}
	expected["google_iap_settings"] = []string{p + "/iap_settings"}
	expected["google_iap_tunnel_dest_group"] = []string{p + "/locations/us-central1/destGroups/dg1"}
	assertResourceIDs(t, g.Resources, expected)
}
