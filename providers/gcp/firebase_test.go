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
	"sort"
	"testing"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"

	"google.golang.org/api/compute/v1"
	firebasev1beta1 "google.golang.org/api/firebase/v1beta1"
	firebaseappcheckv1 "google.golang.org/api/firebaseappcheck/v1"
	firebasedatabasev1beta "google.golang.org/api/firebasedatabase/v1beta"
	firebasedataconnectv1 "google.golang.org/api/firebasedataconnect/v1"
	firebasehostingv1beta1 "google.golang.org/api/firebasehosting/v1beta1"
	firebaseremoteconfigv1 "google.golang.org/api/firebaseremoteconfig/v1"
	firebaserulesv1 "google.golang.org/api/firebaserules/v1"
	firebasestoragev1beta "google.golang.org/api/firebasestorage/v1beta"
	"google.golang.org/api/option"
)

// assertResourceIDs compares the multiset of emitted (type, ID) pairs
// against the expected map of type -> IDs.
func assertResourceIDs(t *testing.T, resources []terraformutils.Resource, expected map[string][]string) {
	t.Helper()
	got := map[string]map[string]bool{}
	for _, resource := range resources {
		if got[resource.InstanceInfo.Type] == nil {
			got[resource.InstanceInfo.Type] = map[string]bool{}
		}
		got[resource.InstanceInfo.Type][resource.InstanceState.ID] = true
	}
	if len(got) != len(expected) {
		t.Errorf("resource types = %d (%v), want %d", len(got), keysOf(got), len(expected))
	}
	for resourceType, ids := range expected {
		gotIDs := []string{}
		for id := range got[resourceType] {
			gotIDs = append(gotIDs, id)
		}
		sort.Strings(gotIDs)
		wantIDs := append([]string{}, ids...)
		sort.Strings(wantIDs)
		if len(gotIDs) != len(wantIDs) {
			t.Errorf("%s IDs = %v, want %v", resourceType, gotIDs, wantIDs)
			continue
		}
		for i := range wantIDs {
			if gotIDs[i] != wantIDs[i] {
				t.Errorf("%s IDs = %v, want %v", resourceType, gotIDs, wantIDs)
				break
			}
		}
	}
}

func keysOf(m map[string]map[string]bool) []string {
	keys := []string{}
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// TestFirebaseGenerator feeds canned Firebase API responses through
// the generator and verifies the emitted resource types and IDs: the
// project, the three app platforms, App Check debug tokens, per-app
// enforcement configs (addressed under apps/{appId}, converted from
// the Management API app names), service config + resource policy,
// AI Logic config/templates/lock, App Hosting backend/build/domains
// (DEFAULT vs CUSTOM discrimination) and traffic, Data Connect,
// database instance, Hosting site tree, Rules, Remote Config,
// Storage for Firebase buckets and extension instances.
func TestFirebaseGenerator(t *testing.T) {
	proj := "projects/test-project"
	loc := proj + "/locations/us-central1"
	androidApp := proj + "/androidApps/app-a"
	iosApp := proj + "/iosApps/app-i"
	webApp := proj + "/webApps/app-w"
	acAndroid := proj + "/apps/app-a"
	acIos := proj + "/apps/app-i"
	acWeb := proj + "/apps/app-w"
	site := proj + "/sites/site1"
	channel := site + "/channels/live"
	backend := loc + "/backends/be1"
	transport := &cannedTransport{routes: []cannedRoute{
		// Hosting tree first: its paths share the /v1beta1/projects
		// prefix with the Firebase Management routes below.
		{"/channels/live/releases", `{"releases":[{"name":"` + channel + `/releases/r1"}]}`},
		{"/sites/site1/channels", `{"channels":[{"name":"` + channel + `"}]}`},
		{"/sites/site1/versions", `{"versions":[{"name":"` + site + `/versions/v1"}]}`},
		{"/sites/site1/customDomains", `{"customDomains":[{"name":"` + site + `/customDomains/example.com"}]}`},
		{"/v1beta1/projects/test-project/sites", `{"sites":[{"name":"` + site + `"}]}`},
		{"/androidApps", `{"apps":[{"name":"` + androidApp + `"}]}`},
		{"/iosApps", `{"apps":[{"name":"` + iosApp + `"}]}`},
		{"/webApps", `{"apps":[{"name":"` + webApp + `"}]}`},
		// App Check per-app debug tokens.
		{"apps/app-a/debugTokens", `{"debugTokens":[{"name":"` + acAndroid + `/debugTokens/tok-a"}]}`},
		{"apps/app-i/debugTokens", `{"debugTokens":[{"name":"` + acIos + `/debugTokens/tok-i"}]}`},
		{"apps/app-w/debugTokens", `{"debugTokens":[{"name":"` + acWeb + `/debugTokens/tok-w"}]}`},
		{"/services/cloud.firestore/resourcePolicies", `{"resourcePolicies":[{"name":"` + proj + `/services/cloud.firestore/resourcePolicies/rp1"}]}`},
		{"/v1/projects/test-project/services", `{"services":[{"name":"` + proj + `/services/cloud.firestore"}]}`},
		// AI Logic (raw REST).
		{"/locations/global/templates", `{"templates":[{"name":"` + proj + `/locations/global/templates/t1","locked":true},{"name":"` + proj + `/locations/global/templates/t2"}]}`},
		// App Hosting (raw REST).
		{"/backends/be1/builds", `{"builds":[{"name":"` + backend + `/builds/b1"}]}`},
		{"/backends/be1/domains", `{"domains":[{"name":"` + backend + `/domains/d1","type":"DEFAULT"},{"name":"` + backend + `/domains/d2","type":"CUSTOM"}]}`},
		{"/locations/us-central1/backends", `{"backends":[{"name":"` + backend + `"}]}`},
		// Data Connect and Realtime Database share the
		// locations/us-central1 segment.
		{"/locations/us-central1/services", `{"services":[{"name":"` + loc + `/services/dc1"}]}`},
		{"/locations/us-central1/instances", `{"instances":[{"name":"` + loc + `/instances/db1"}]}`},
		// Rules.
		{"/v1/projects/test-project/rulesets", `{"rulesets":[{"name":"` + proj + `/rulesets/rs1"}]}`},
		{"/v1/projects/test-project/releases", `{"releases":[{"name":"` + proj + `/releases/rel1"}]}`},
		// Storage for Firebase.
		{"/v1beta/projects/test-project/buckets", `{"buckets":[{"name":"` + proj + `/buckets/bkt1"}]}`},
		{"/defaultBucket", `{"name":"` + proj + `/defaultBucket"}`},
		// Extensions (raw REST).
		{"/v1beta/projects/test-project/instances", `{"instances":[{"name":"` + proj + `/instances/ext1"}]}`},
	}}

	ctx := context.Background()
	clientOpts := []option.ClientOption{
		option.WithHTTPClient(transport.client()),
		option.WithoutAuthentication(),
	}
	firebaseService, err := firebasev1beta1.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("firebase.NewService: %v", err)
	}
	appCheckService, err := firebaseappcheckv1.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("firebaseappcheck.NewService: %v", err)
	}
	databaseService, err := firebasedatabasev1beta.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("firebasedatabase.NewService: %v", err)
	}
	dataConnectService, err := firebasedataconnectv1.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("firebasedataconnect.NewService: %v", err)
	}
	hostingService, err := firebasehostingv1beta1.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("firebasehosting.NewService: %v", err)
	}
	rulesService, err := firebaserulesv1.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("firebaserules.NewService: %v", err)
	}
	remoteConfigService, err := firebaseremoteconfigv1.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("firebaseremoteconfig.NewService: %v", err)
	}
	storageService, err := firebasestoragev1beta.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("firebasestorage.NewService: %v", err)
	}

	g := &FirebaseGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	g.initResources(ctx, firebaseService, appCheckService, databaseService, dataConnectService, hostingService, rulesService, remoteConfigService, storageService, transport.client())

	expected := map[string][]string{
		"google_firebase_project":     {proj},
		"google_firebase_android_app": {androidApp},
		"google_firebase_apple_app":   {iosApp},
		"google_firebase_web_app":     {webApp},
		"google_firebase_app_check_debug_token": {
			acAndroid + "/debugTokens/tok-a",
			acIos + "/debugTokens/tok-i",
			acWeb + "/debugTokens/tok-w",
		},
		"google_firebase_app_check_service_config":      {proj + "/services/cloud.firestore"},
		"google_firebase_app_check_resource_policy":     {proj + "/services/cloud.firestore/resourcePolicies/rp1"},
		"google_firebase_ai_logic_config":               {proj + "/locations/global/config"},
		"google_firebase_ai_logic_prompt_template":      {proj + "/locations/global/templates/t1", proj + "/locations/global/templates/t2"},
		"google_firebase_ai_logic_prompt_template_lock": {proj + "/locations/global/templates/t1"},
		"google_firebase_app_hosting_backend":           {backend},
		"google_firebase_app_hosting_build":             {backend + "/builds/b1"},
		"google_firebase_app_hosting_default_domain":    {backend + "/domains/d1"},
		"google_firebase_app_hosting_domain":            {backend + "/domains/d2"},
		"google_firebase_app_hosting_traffic":           {backend + "/traffic"},
		"google_firebase_data_connect_service":          {loc + "/services/dc1"},
		"google_firebase_database_instance":             {loc + "/instances/db1"},
		"google_firebase_hosting_site":                  {site},
		"google_firebase_hosting_channel":               {channel},
		"google_firebase_hosting_release":               {channel + "/releases/r1"},
		"google_firebase_hosting_version":               {site + "/versions/v1"},
		"google_firebase_hosting_custom_domain":         {site + "/customDomains/example.com"},
		"google_firebaserules_ruleset":                  {proj + "/rulesets/rs1"},
		"google_firebaserules_release":                  {proj + "/releases/rel1"},
		"google_firebase_remote_config_remote_config":   {proj + "/remoteConfig"},
		"google_firebase_storage_bucket":                {proj + "/buckets/bkt1"},
		"google_firebase_storage_default_bucket":        {proj + "/defaultBucket"},
		"google_firebase_extensions_instance":           {proj + "/instances/ext1"},
	}
	// Every app gets all five App Check enforcement configs: the mock
	// transport answers every Get with 200, so each singleton is
	// "configured" for each app.
	for _, acApp := range []string{acAndroid, acIos, acWeb} {
		expected["google_firebase_app_check_app_attest_config"] = append(expected["google_firebase_app_check_app_attest_config"], acApp+"/appAttestConfig")
		expected["google_firebase_app_check_device_check_config"] = append(expected["google_firebase_app_check_device_check_config"], acApp+"/deviceCheckConfig")
		expected["google_firebase_app_check_play_integrity_config"] = append(expected["google_firebase_app_check_play_integrity_config"], acApp+"/playIntegrityConfig")
		expected["google_firebase_app_check_recaptcha_enterprise_config"] = append(expected["google_firebase_app_check_recaptcha_enterprise_config"], acApp+"/recaptchaEnterpriseConfig")
		expected["google_firebase_app_check_recaptcha_v3_config"] = append(expected["google_firebase_app_check_recaptcha_v3_config"], acApp+"/recaptchaV3Config")
	}
	assertResourceIDs(t, g.Resources, expected)
}
