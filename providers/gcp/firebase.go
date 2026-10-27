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
	"encoding/json"
	"log"
	"net/http"
	"strings"

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
)

// FirebaseGenerator enumerates the Firebase ecosystem of one project:
// the Firebase project itself, mobile/web apps, App Check (debug
// tokens, per-app enforcement configs, service configs and resource
// policies), AI Logic (client config, prompt templates and their
// locks), App Hosting (backends, builds, domains, traffic), Data
// Connect services, Realtime Database instances, Hosting sites and
// their channels/versions/releases/custom domains, Firebase Rules
// rulesets and releases, Remote Config, Cloud Storage for Firebase
// buckets, and extension instances.
//
// APIs without a generated client in the pinned google.golang.org/api
// module (AI Logic / firebasevertexai, App Hosting /
// firebaseapphosting, Extensions / firebaseextensions, and the App
// Check reCAPTCHA v3 config) are enumerated over raw REST through
// gcp_rest.go, with endpoints taken from the services' discovery
// documents and the google provider's Magic Modules definitions.
//
// Skipped, with reasons:
//   - google_firebase_app_check_safetynet_config: SafetyNet is
//     deprecated by Google in favour of Play Integrity; the resource
//     is a deprecated duplicate of the Play Integrity enforcement
//     config over the same per-app slot.
//   - google_firebase_android_app / apple / web SHA certificates and
//     app config files are not audit types; only the apps are in
//     scope here.
//   - google_firebase_project_location, google_firebase_google_sign_in,
//     google_firebase_realtime_database_* (rules are a string field on
//     the instance in the current API surface exposed to the provider
//     audit) are not part of this family's audit lines.
type FirebaseGenerator struct {
	GCPService
}

func (g *FirebaseGenerator) InitResources() error {
	ctx := context.Background()
	firebaseService, err := firebasev1beta1.NewService(ctx)
	if err != nil {
		log.Println(err)
		firebaseService = nil
	}
	appCheckService, err := firebaseappcheckv1.NewService(ctx)
	if err != nil {
		log.Println(err)
		appCheckService = nil
	}
	databaseService, err := firebasedatabasev1beta.NewService(ctx)
	if err != nil {
		log.Println(err)
		databaseService = nil
	}
	dataConnectService, err := firebasedataconnectv1.NewService(ctx)
	if err != nil {
		log.Println(err)
		dataConnectService = nil
	}
	hostingService, err := firebasehostingv1beta1.NewService(ctx)
	if err != nil {
		log.Println(err)
		hostingService = nil
	}
	rulesService, err := firebaserulesv1.NewService(ctx)
	if err != nil {
		log.Println(err)
		rulesService = nil
	}
	remoteConfigService, err := firebaseremoteconfigv1.NewService(ctx)
	if err != nil {
		log.Println(err)
		remoteConfigService = nil
	}
	storageService, err := firebasestoragev1beta.NewService(ctx)
	if err != nil {
		log.Println(err)
		storageService = nil
	}
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		log.Println(err)
		restClient = nil
	}
	g.initResources(ctx, firebaseService, appCheckService, databaseService, dataConnectService, hostingService, rulesService, remoteConfigService, storageService, restClient)
	return nil
}

func (g *FirebaseGenerator) initResources(ctx context.Context, firebaseService *firebasev1beta1.Service, appCheckService *firebaseappcheckv1.Service, databaseService *firebasedatabasev1beta.Service, dataConnectService *firebasedataconnectv1.APIService, hostingService *firebasehostingv1beta1.Service, rulesService *firebaserulesv1.Service, remoteConfigService *firebaseremoteconfigv1.Service, storageService *firebasestoragev1beta.Service, restClient *http.Client) {
	project := g.GetArgs()["project"].(string)
	projectName := "projects/" + project
	region := g.GetArgs()["region"].(compute.Region).Name
	locationName := projectName + "/locations/" + region

	// The Firebase project itself: a project either is a Firebase
	// project (Get succeeds) or is not (404), so Get is the probe.
	if firebaseService != nil {
		if _, err := firebaseService.Projects.Get(projectName).Context(ctx).Do(); err == nil {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				projectName,
				projectName,
				"google_firebase_project",
				g.ProviderName,
				[]string{},
			))
		}
	}

	appNames := g.firebaseApps(ctx, firebaseService, projectName)
	if appCheckService != nil {
		g.firebaseAppCheck(ctx, appCheckService, restClient, projectName, appNames)
	}
	if restClient != nil {
		g.firebaseAILogic(ctx, restClient, projectName)
		g.firebaseAppHosting(ctx, restClient, locationName)
		g.firebaseExtensions(ctx, restClient, projectName)
	}
	if dataConnectService != nil {
		call := dataConnectService.Projects.Locations.Services.List(locationName)
		if err := call.Pages(ctx, func(page *firebasedataconnectv1.ListServicesResponse) error {
			for _, service := range page.Services {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					service.Name,
					service.Name,
					"google_firebase_data_connect_service",
					g.ProviderName,
					[]string{},
				))
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
	}
	if databaseService != nil {
		g.firebaseDatabaseInstances(ctx, databaseService, projectName, region)
	}
	if hostingService != nil {
		g.firebaseHosting(ctx, hostingService, projectName)
	}
	if rulesService != nil {
		call := rulesService.Projects.Rulesets.List(projectName)
		if err := call.Pages(ctx, func(page *firebaserulesv1.ListRulesetsResponse) error {
			for _, ruleset := range page.Rulesets {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					ruleset.Name,
					ruleset.Name,
					"google_firebaserules_ruleset",
					g.ProviderName,
					[]string{},
				))
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
		releaseCall := rulesService.Projects.Releases.List(projectName)
		if err := releaseCall.Pages(ctx, func(page *firebaserulesv1.ListReleasesResponse) error {
			for _, release := range page.Releases {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					release.Name,
					release.Name,
					"google_firebaserules_release",
					g.ProviderName,
					[]string{},
				))
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
	}
	// Remote Config is a per-project singleton: Get is the probe and
	// the provider ID is projects/{project}/remoteConfig.
	if remoteConfigService != nil {
		if _, err := remoteConfigService.Projects.GetRemoteConfig(projectName).Context(ctx).Do(); err == nil {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				projectName+"/remoteConfig",
				projectName+"/remoteConfig",
				"google_firebase_remote_config_remote_config",
				g.ProviderName,
				[]string{},
			))
		}
	}
	if storageService != nil {
		call := storageService.Projects.Buckets.List(projectName)
		if err := call.Pages(ctx, func(page *firebasestoragev1beta.ListBucketsResponse) error {
			for _, bucket := range page.Buckets {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					bucket.Name,
					bucket.Name,
					"google_firebase_storage_bucket",
					g.ProviderName,
					[]string{},
				))
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
		if defaultBucket, err := storageService.Projects.GetDefaultBucket(projectName).Context(ctx).Do(); err == nil {
			id := defaultBucket.Name
			if id == "" {
				id = projectName + "/defaultBucket"
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				id,
				"google_firebase_storage_default_bucket",
				g.ProviderName,
				[]string{},
			))
		}
	}
}

// firebaseApps lists the Android, iOS and web apps of the project and
// returns every app's resource name for the App Check pass.
func (g *FirebaseGenerator) firebaseApps(ctx context.Context, firebaseService *firebasev1beta1.Service, projectName string) []string {
	appNames := []string{}
	if firebaseService == nil {
		return appNames
	}
	androidCall := firebaseService.Projects.AndroidApps.List(projectName)
	if err := androidCall.Pages(ctx, func(page *firebasev1beta1.ListAndroidAppsResponse) error {
		for _, app := range page.Apps {
			appNames = append(appNames, app.Name)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				app.Name,
				app.Name,
				"google_firebase_android_app",
				g.ProviderName,
				[]string{},
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	iosCall := firebaseService.Projects.IosApps.List(projectName)
	if err := iosCall.Pages(ctx, func(page *firebasev1beta1.ListIosAppsResponse) error {
		for _, app := range page.Apps {
			appNames = append(appNames, app.Name)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				app.Name,
				app.Name,
				"google_firebase_apple_app",
				g.ProviderName,
				[]string{},
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	webCall := firebaseService.Projects.WebApps.List(projectName)
	if err := webCall.Pages(ctx, func(page *firebasev1beta1.ListWebAppsResponse) error {
		for _, app := range page.Apps {
			appNames = append(appNames, app.Name)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				app.Name,
				app.Name,
				"google_firebase_web_app",
				g.ProviderName,
				[]string{},
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	return appNames
}

// firebaseAppCheck enumerates App Check debug tokens, the per-app
// enforcement configs (Get-only singletons, emitted when present),
// service configs and their resource policies.
func (g *FirebaseGenerator) firebaseAppCheck(ctx context.Context, appCheckService *firebaseappcheckv1.Service, restClient *http.Client, projectName string, appNames []string) {
	for _, appName := range appNames {
		// App Check addresses apps as projects/{project}/apps/{appId},
		// while the Firebase Management API names them
		// projects/{project}/(android|ios|web)Apps/{appId}.
		appCheckName := projectName + "/apps/" + appName[strings.LastIndex(appName, "/")+1:]
		tokenCall := appCheckService.Projects.Apps.DebugTokens.List(appCheckName)
		if err := tokenCall.Pages(ctx, func(page *firebaseappcheckv1.GoogleFirebaseAppcheckV1ListDebugTokensResponse) error {
			for _, token := range page.DebugTokens {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					token.Name,
					token.Name,
					"google_firebase_app_check_debug_token",
					g.ProviderName,
					[]string{},
				))
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
		// Per-app enforcement configs are singletons addressed by a
		// fixed suffix under the app; each is emitted only when the
		// provider is configured for that app (Get succeeds).
		configTypes := map[string]string{
			"appAttestConfig":           "google_firebase_app_check_app_attest_config",
			"deviceCheckConfig":         "google_firebase_app_check_device_check_config",
			"playIntegrityConfig":       "google_firebase_app_check_play_integrity_config",
			"recaptchaEnterpriseConfig": "google_firebase_app_check_recaptcha_enterprise_config",
		}
		for suffix, resourceType := range configTypes {
			configName := appCheckName + "/" + suffix
			var err error
			switch suffix {
			case "appAttestConfig":
				_, err = appCheckService.Projects.Apps.AppAttestConfig.Get(configName).Context(ctx).Do()
			case "deviceCheckConfig":
				_, err = appCheckService.Projects.Apps.DeviceCheckConfig.Get(configName).Context(ctx).Do()
			case "playIntegrityConfig":
				_, err = appCheckService.Projects.Apps.PlayIntegrityConfig.Get(configName).Context(ctx).Do()
			case "recaptchaEnterpriseConfig":
				_, err = appCheckService.Projects.Apps.RecaptchaEnterpriseConfig.Get(configName).Context(ctx).Do()
			}
			if err == nil {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					configName,
					configName,
					resourceType,
					g.ProviderName,
					[]string{},
				))
			}
		}
		// The reCAPTCHA v3 config has no service in the pinned v1
		// client; probe it over raw REST (firebaseappcheck v1).
		if restClient != nil {
			configName := appCheckName + "/recaptchaV3Config"
			url := "https://firebaseappcheck.googleapis.com/v1/" + configName
			if _, err := restDo(ctx, restClient, http.MethodGet, url, nil); err == nil {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					configName,
					configName,
					"google_firebase_app_check_recaptcha_v3_config",
					g.ProviderName,
					[]string{},
				))
			}
		}
	}
	serviceCall := appCheckService.Projects.Services.List(projectName)
	if err := serviceCall.Pages(ctx, func(page *firebaseappcheckv1.GoogleFirebaseAppcheckV1ListServicesResponse) error {
		for _, service := range page.Services {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				service.Name,
				service.Name,
				"google_firebase_app_check_service_config",
				g.ProviderName,
				[]string{},
			))
			policyCall := appCheckService.Projects.Services.ResourcePolicies.List(service.Name)
			if err := policyCall.Pages(ctx, func(page *firebaseappcheckv1.GoogleFirebaseAppcheckV1ListResourcePoliciesResponse) error {
				for _, policy := range page.ResourcePolicies {
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						policy.Name,
						policy.Name,
						"google_firebase_app_check_resource_policy",
						g.ProviderName,
						[]string{},
					))
				}
				return nil
			}); err != nil {
				log.Println(err)
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
}

// firebaseAILogic enumerates Firebase AI Logic (firebasevertexai
// v1beta, raw REST — no generated client): the per-project client
// config singleton, prompt templates, and a prompt template lock for
// every locked template (the lock shares the template's ID in the
// google provider).
func (g *FirebaseGenerator) firebaseAILogic(ctx context.Context, restClient *http.Client, projectName string) {
	base := "https://firebasevertexai.googleapis.com/v1beta/" + projectName + "/locations/global"
	configName := projectName + "/locations/global/config"
	if _, err := restDo(ctx, restClient, http.MethodGet, "https://firebasevertexai.googleapis.com/v1beta/"+configName, nil); err == nil {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			configName,
			configName,
			"google_firebase_ai_logic_config",
			g.ProviderName,
			[]string{},
		))
	}
	body, err := restDo(ctx, restClient, http.MethodGet, base+"/templates", nil)
	if err != nil {
		log.Println(err)
		return
	}
	var list struct {
		Templates []struct {
			Name   string `json:"name"`
			Locked bool   `json:"locked"`
		} `json:"templates"`
	}
	if err := json.Unmarshal(body, &list); err != nil {
		log.Println(err)
		return
	}
	for _, template := range list.Templates {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			template.Name,
			template.Name,
			"google_firebase_ai_logic_prompt_template",
			g.ProviderName,
			[]string{},
		))
		if template.Locked {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				template.Name,
				template.Name,
				"google_firebase_ai_logic_prompt_template_lock",
				g.ProviderName,
				[]string{},
			))
		}
	}
}

// firebaseAppHosting enumerates Firebase App Hosting (raw REST,
// firebaseapphosting v1beta — no generated client) in the configured
// region: backends, their builds, their domains (the API's type field
// discriminates the provider's default-domain and custom-domain
// resources, which share one ID format), and the per-backend traffic
// singleton.
func (g *FirebaseGenerator) firebaseAppHosting(ctx context.Context, restClient *http.Client, locationName string) {
	backends, err := restListResources(ctx, restClient, "https://firebaseapphosting.googleapis.com/v1beta/"+locationName+"/backends")
	if err != nil {
		log.Println(err)
		return
	}
	for _, backend := range backends {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			backend.Name,
			backend.Name,
			"google_firebase_app_hosting_backend",
			g.ProviderName,
			[]string{},
		))
		builds, err := restListResources(ctx, restClient, "https://firebaseapphosting.googleapis.com/v1beta/"+backend.Name+"/builds")
		if err != nil {
			log.Println(err)
		}
		for _, build := range builds {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				build.Name,
				build.Name,
				"google_firebase_app_hosting_build",
				g.ProviderName,
				[]string{},
			))
		}
		g.firebaseAppHostingDomains(ctx, restClient, backend.Name)
		trafficName := backend.Name + "/traffic"
		if _, err := restDo(ctx, restClient, http.MethodGet, "https://firebaseapphosting.googleapis.com/v1beta/"+trafficName, nil); err == nil {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				trafficName,
				trafficName,
				"google_firebase_app_hosting_traffic",
				g.ProviderName,
				[]string{},
			))
		}
	}
}

// firebaseAppHostingDomains lists one backend's domains and maps the
// API type onto the two provider resources that share the collection.
func (g *FirebaseGenerator) firebaseAppHostingDomains(ctx context.Context, restClient *http.Client, backendName string) {
	body, err := restDo(ctx, restClient, http.MethodGet, "https://firebaseapphosting.googleapis.com/v1beta/"+backendName+"/domains", nil)
	if err != nil {
		log.Println(err)
		return
	}
	var list struct {
		Domains []struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"domains"`
	}
	if err := json.Unmarshal(body, &list); err != nil {
		log.Println(err)
		return
	}
	for _, domain := range list.Domains {
		resourceType := "google_firebase_app_hosting_domain"
		if strings.HasSuffix(domain.Type, "DEFAULT") {
			resourceType = "google_firebase_app_hosting_default_domain"
		}
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			domain.Name,
			domain.Name,
			resourceType,
			g.ProviderName,
			[]string{},
		))
	}
}

// firebaseDatabaseInstances lists Realtime Database instances. The
// configured region is tried first, then the three regions the
// service exists in; instances are deduplicated by name.
func (g *FirebaseGenerator) firebaseDatabaseInstances(ctx context.Context, databaseService *firebasedatabasev1beta.Service, projectName, region string) {
	seen := map[string]bool{}
	locations := []string{region, "us-central1", "europe-west1", "asia-southeast1"}
	for _, location := range locations {
		parent := projectName + "/locations/" + location
		call := databaseService.Projects.Locations.Instances.List(parent)
		if err := call.Pages(ctx, func(page *firebasedatabasev1beta.ListDatabaseInstancesResponse) error {
			for _, instance := range page.Instances {
				if seen[instance.Name] {
					continue
				}
				seen[instance.Name] = true
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					instance.Name,
					instance.Name,
					"google_firebase_database_instance",
					g.ProviderName,
					[]string{},
				))
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
	}
}

// firebaseHosting enumerates Hosting sites with their channels,
// channel releases, versions and custom domains.
func (g *FirebaseGenerator) firebaseHosting(ctx context.Context, hostingService *firebasehostingv1beta1.Service, projectName string) {
	siteCall := hostingService.Projects.Sites.List(projectName)
	if err := siteCall.Pages(ctx, func(page *firebasehostingv1beta1.ListSitesResponse) error {
		for _, site := range page.Sites {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				site.Name,
				site.Name,
				"google_firebase_hosting_site",
				g.ProviderName,
				[]string{},
			))
			channelCall := hostingService.Projects.Sites.Channels.List(site.Name)
			if err := channelCall.Pages(ctx, func(page *firebasehostingv1beta1.ListChannelsResponse) error {
				for _, channel := range page.Channels {
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						channel.Name,
						channel.Name,
						"google_firebase_hosting_channel",
						g.ProviderName,
						[]string{},
					))
					releaseCall := hostingService.Projects.Sites.Channels.Releases.List(channel.Name)
					if err := releaseCall.Pages(ctx, func(page *firebasehostingv1beta1.ListReleasesResponse) error {
						for _, release := range page.Releases {
							g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
								release.Name,
								release.Name,
								"google_firebase_hosting_release",
								g.ProviderName,
								[]string{},
							))
						}
						return nil
					}); err != nil {
						log.Println(err)
					}
				}
				return nil
			}); err != nil {
				log.Println(err)
			}
			versionCall := hostingService.Projects.Sites.Versions.List(site.Name)
			if err := versionCall.Pages(ctx, func(page *firebasehostingv1beta1.ListVersionsResponse) error {
				for _, version := range page.Versions {
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						version.Name,
						version.Name,
						"google_firebase_hosting_version",
						g.ProviderName,
						[]string{},
					))
				}
				return nil
			}); err != nil {
				log.Println(err)
			}
			domainCall := hostingService.Projects.Sites.CustomDomains.List(site.Name)
			if err := domainCall.Pages(ctx, func(page *firebasehostingv1beta1.ListCustomDomainsResponse) error {
				for _, domain := range page.CustomDomains {
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						domain.Name,
						domain.Name,
						"google_firebase_hosting_custom_domain",
						g.ProviderName,
						[]string{},
					))
				}
				return nil
			}); err != nil {
				log.Println(err)
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
}

// firebaseExtensions enumerates extension instances over raw REST
// (firebaseextensions v1beta — no generated client).
func (g *FirebaseGenerator) firebaseExtensions(ctx context.Context, restClient *http.Client, projectName string) {
	instances, err := restListResources(ctx, restClient, "https://firebaseextensions.googleapis.com/v1beta/"+projectName+"/instances")
	if err != nil {
		log.Println(err)
		return
	}
	for _, instance := range instances {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			instance.Name,
			instance.Name,
			"google_firebase_extensions_instance",
			g.ProviderName,
			[]string{},
		))
	}
}
