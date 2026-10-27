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
	"google.golang.org/api/cloudresourcemanager/v1"
	sccv1 "google.golang.org/api/securitycenter/v1"
)

var sccAllowEmptyValues = []string{""}

var sccAdditionalFields = map[string]interface{}{}

const (
	sccV2BaseURL         = "https://securitycenter.googleapis.com/v2/"
	sccManagementBaseURL = "https://securitycentermanagement.googleapis.com/v1/"
	// Security Command Center v2 and Management API resources live in
	// the global location unless data residency pins them elsewhere;
	// the google provider defaults location to "global" as well.
	sccManagementLocation = "global"
)

type SccGenerator struct {
	GCPService
}

// sccScope is one resource-manager level at which Security Command
// Center collections are enumerated: the configured project always,
// plus its ancestor folders and organization when Cloud Resource
// Manager ancestry resolves them.
type sccScope struct {
	kind   string // "organization", "folder" or "project"
	parent string // "organizations/{id}", "folders/{id}" or "projects/{id}"
	id     string
	attrs  map[string]string
}

// InitResources enumerates Security Command Center resources:
//
//   - v1 (generated securitycenter client): organization sources (with
//     the IAM policy, bindings and members expanded from each source's
//     live IAM policy), notification configs, mute configs, BigQuery
//     exports and Security Health Analytics custom modules at project,
//     ancestor-folder and organization scope, and Event Threat
//     Detection custom modules at organization scope. Custom module
//     lists also return modules inherited from ancestors; those are
//     emitted only at their owning scope.
//   - v2 (raw REST; the pinned google.golang.org/api has no v2 client):
//     organization sources (with IAM expansion), mute configs,
//     notification configs and BigQuery exports at all three scopes,
//     in the global location.
//   - The Security Center Management API (raw REST; no generated
//     client exists): Security Health Analytics custom modules at all
//     three scopes and Event Threat Detection custom modules at
//     organization scope, in the global location.
//
// Deliberately not generated:
//   - google_scc_notification_service_account: the "resource" is the
//     SCC service identity for a parent, provisioned through a Service
//     Usage generateServiceIdentity call. There is no listable or
//     gettable object; reading it would mutate state.
//   - google_scc_v2_organization_scc_big_query_exports (plural): a
//     deprecated duplicate of google_scc_v2_organization_scc_big_query_export
//     over the same API objects and ID format (its own provider
//     documentation says deprecated); emitting both would double-manage
//     every export.
func (g *SccGenerator) InitResources() error {
	ctx := context.Background()
	sccService, err := sccv1.NewService(ctx)
	if err != nil {
		return err
	}
	crmService, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		return err
	}
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, sccService, crmService, restClient)
}

func (g *SccGenerator) initResources(ctx context.Context, sccService *sccv1.Service, crmService *cloudresourcemanager.Service, restClient *http.Client) error {
	project := g.GetArgs()["project"].(string)
	scopes := []sccScope{{
		kind:   "project",
		parent: "projects/" + project,
		id:     project,
		attrs:  map[string]string{"project": project},
	}}
	if crmService != nil {
		if ancestry, err := crmService.Projects.GetAncestry(project, &cloudresourcemanager.GetAncestryRequest{}).Context(ctx).Do(); err == nil {
			var folders []sccScope
			var org *sccScope
			for _, ancestor := range ancestry.Ancestor {
				if ancestor.ResourceId == nil {
					continue
				}
				switch ancestor.ResourceId.Type {
				case "folder":
					folders = append(folders, sccScope{
						kind:   "folder",
						parent: "folders/" + ancestor.ResourceId.Id,
						id:     ancestor.ResourceId.Id,
						attrs:  map[string]string{"folder": ancestor.ResourceId.Id},
					})
				case "organization":
					orgScope := sccScope{
						kind:   "organization",
						parent: "organizations/" + ancestor.ResourceId.Id,
						id:     ancestor.ResourceId.Id,
						attrs:  map[string]string{"organization": ancestor.ResourceId.Id},
					}
					org = &orgScope
				}
			}
			scopes = append(scopes, folders...)
			if org != nil {
				scopes = append(scopes, *org)
			}
		} else {
			log.Println(err)
		}
	}
	if len(scopes) == 1 {
		log.Println("scc: could not resolve organization or folders for project " + project + "; enumerating project scope only")
	}

	for _, scope := range scopes {
		if err := g.loadV1Scope(ctx, sccService, scope); err != nil {
			if scope.kind == "project" {
				return err
			}
			log.Println(err)
		}
		// The v2 and Management surfaces are raw REST; a failure in one
		// scope is logged rather than aborting the whole import.
		g.loadV2Scope(ctx, restClient, scope)
		g.loadManagementScope(ctx, restClient, scope)
	}
	return nil
}

func (g *SccGenerator) newSccResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
	if attrs == nil {
		attrs = map[string]string{}
	}
	parts := strings.Split(id, "/")
	return terraformutils.NewResource(
		id,
		parts[len(parts)-1],
		resourceType,
		g.ProviderName,
		attrs,
		sccAllowEmptyValues,
		sccAdditionalFields,
	)
}

func (g *SccGenerator) loadV1Scope(ctx context.Context, sccService *sccv1.Service, scope sccScope) error {
	if scope.kind == "organization" {
		if err := g.loadV1Sources(ctx, sccService, scope); err != nil {
			return err
		}
		if err := g.loadV1EventThreatDetectionCustomModules(ctx, sccService, scope); err != nil {
			return err
		}
	}
	if err := g.loadV1NotificationConfigs(ctx, sccService, scope); err != nil {
		return err
	}
	if err := g.loadV1MuteConfigs(ctx, sccService, scope); err != nil {
		return err
	}
	if err := g.loadV1BigQueryExports(ctx, sccService, scope); err != nil {
		return err
	}
	return g.loadV1SecurityHealthAnalyticsCustomModules(ctx, sccService, scope)
}

func (g *SccGenerator) loadV1Sources(ctx context.Context, sccService *sccv1.Service, scope sccScope) error {
	return sccService.Organizations.Sources.List(scope.parent).Pages(ctx, func(page *sccv1.ListSourcesResponse) error {
		for _, source := range page.Sources {
			g.Resources = append(g.Resources, g.newSccResource(source.Name, "google_scc_source", scope.attrs))
			g.loadV1SourceIam(ctx, sccService, source.Name)
		}
		return nil
	})
}

func (g *SccGenerator) loadV1SourceIam(ctx context.Context, sccService *sccv1.Service, sourceName string) {
	policy, err := sccService.Organizations.Sources.GetIamPolicy(sourceName, &sccv1.GetIamPolicyRequest{}).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		return
	}
	bindings := []iamBinding{}
	for _, binding := range policy.Bindings {
		bindings = append(bindings, iamBinding{Role: binding.Role, Members: binding.Members})
	}
	g.Resources = append(g.Resources, expandIamPolicy(
		bindings,
		sourceName,
		"google_scc_source_iam_policy",
		"google_scc_source_iam_binding",
		"google_scc_source_iam_member",
		map[string]string{"source": sourceName},
		g.newSccResource,
	)...)
}

func (g *SccGenerator) loadV1NotificationConfigs(ctx context.Context, sccService *sccv1.Service, scope sccScope) error {
	resourceType := map[string]string{
		"organization": "google_scc_notification_config",
		"folder":       "google_scc_folder_notification_config",
		"project":      "google_scc_project_notification_config",
	}[scope.kind]
	emit := func(page *sccv1.ListNotificationConfigsResponse) error {
		for _, config := range page.NotificationConfigs {
			g.Resources = append(g.Resources, g.newSccResource(config.Name, resourceType, scope.attrs))
		}
		return nil
	}
	switch scope.kind {
	case "organization":
		return sccService.Organizations.NotificationConfigs.List(scope.parent).Pages(ctx, emit)
	case "folder":
		return sccService.Folders.NotificationConfigs.List(scope.parent).Pages(ctx, emit)
	default:
		return sccService.Projects.NotificationConfigs.List(scope.parent).Pages(ctx, emit)
	}
}

func (g *SccGenerator) loadV1MuteConfigs(ctx context.Context, sccService *sccv1.Service, scope sccScope) error {
	// google_scc_mute_config is a single provider resource whose parent
	// may be an organization, folder or project.
	emit := func(page *sccv1.ListMuteConfigsResponse) error {
		for _, config := range page.MuteConfigs {
			g.Resources = append(g.Resources, g.newSccResource(config.Name, "google_scc_mute_config", scope.attrs))
		}
		return nil
	}
	switch scope.kind {
	case "organization":
		return sccService.Organizations.MuteConfigs.List(scope.parent).Pages(ctx, emit)
	case "folder":
		return sccService.Folders.MuteConfigs.List(scope.parent).Pages(ctx, emit)
	default:
		return sccService.Projects.MuteConfigs.List(scope.parent).Pages(ctx, emit)
	}
}

func (g *SccGenerator) loadV1BigQueryExports(ctx context.Context, sccService *sccv1.Service, scope sccScope) error {
	resourceType := map[string]string{
		"organization": "google_scc_organization_scc_big_query_export",
		"folder":       "google_scc_folder_scc_big_query_export",
		"project":      "google_scc_project_scc_big_query_export",
	}[scope.kind]
	emit := func(page *sccv1.ListBigQueryExportsResponse) error {
		for _, export := range page.BigQueryExports {
			g.Resources = append(g.Resources, g.newSccResource(export.Name, resourceType, scope.attrs))
		}
		return nil
	}
	switch scope.kind {
	case "organization":
		return sccService.Organizations.BigQueryExports.List(scope.parent).Pages(ctx, emit)
	case "folder":
		return sccService.Folders.BigQueryExports.List(scope.parent).Pages(ctx, emit)
	default:
		return sccService.Projects.BigQueryExports.List(scope.parent).Pages(ctx, emit)
	}
}

func (g *SccGenerator) loadV1SecurityHealthAnalyticsCustomModules(ctx context.Context, sccService *sccv1.Service, scope sccScope) error {
	resourceType := map[string]string{
		"organization": "google_scc_organization_custom_module",
		"folder":       "google_scc_folder_custom_module",
		"project":      "google_scc_project_custom_module",
	}[scope.kind]
	parent := scope.parent + "/securityHealthAnalyticsSettings"
	emit := func(page *sccv1.ListSecurityHealthAnalyticsCustomModulesResponse) error {
		for _, module := range page.SecurityHealthAnalyticsCustomModules {
			// Lists include modules inherited from ancestors; those
			// are owned (and emitted) at the ancestor scope.
			if scope.kind != "organization" && module.AncestorModule != "" {
				continue
			}
			g.Resources = append(g.Resources, g.newSccResource(module.Name, resourceType, scope.attrs))
		}
		return nil
	}
	switch scope.kind {
	case "organization":
		return sccService.Organizations.SecurityHealthAnalyticsSettings.CustomModules.List(parent).Pages(ctx, emit)
	case "folder":
		return sccService.Folders.SecurityHealthAnalyticsSettings.CustomModules.List(parent).Pages(ctx, emit)
	default:
		return sccService.Projects.SecurityHealthAnalyticsSettings.CustomModules.List(parent).Pages(ctx, emit)
	}
}

func (g *SccGenerator) loadV1EventThreatDetectionCustomModules(ctx context.Context, sccService *sccv1.Service, scope sccScope) error {
	// Event Threat Detection custom modules can be created at the
	// organization level only, matching the single provider resource.
	parent := scope.parent + "/eventThreatDetectionSettings"
	return sccService.Organizations.EventThreatDetectionSettings.CustomModules.List(parent).Pages(ctx, func(page *sccv1.ListEventThreatDetectionCustomModulesResponse) error {
		for _, module := range page.EventThreatDetectionCustomModules {
			g.Resources = append(g.Resources, g.newSccResource(module.Name, "google_scc_event_threat_detection_custom_module", scope.attrs))
		}
		return nil
	})
}

// loadV2Scope enumerates the v2 collections of one scope through raw
// REST. Failures are logged: the v2 surface is best-effort on top of
// the v1 enumeration.
func (g *SccGenerator) loadV2Scope(ctx context.Context, restClient *http.Client, scope sccScope) {
	locationParent := sccV2BaseURL + scope.parent + "/locations/" + sccManagementLocation
	collections := []struct {
		url          string
		resourceType string
	}{
		{locationParent + "/muteConfigs", map[string]string{
			"organization": "google_scc_v2_organization_mute_config",
			"folder":       "google_scc_v2_folder_mute_config",
			"project":      "google_scc_v2_project_mute_config",
		}[scope.kind]},
		{locationParent + "/notificationConfigs", map[string]string{
			"organization": "google_scc_v2_organization_notification_config",
			"folder":       "google_scc_v2_folder_notification_config",
			"project":      "google_scc_v2_project_notification_config",
		}[scope.kind]},
		{locationParent + "/bigQueryExports", map[string]string{
			"organization": "google_scc_v2_organization_scc_big_query_export",
			"folder":       "google_scc_v2_folder_scc_big_query_export",
			"project":      "google_scc_v2_project_scc_big_query_export",
		}[scope.kind]},
	}
	for _, collection := range collections {
		items, err := restListResources(ctx, restClient, collection.url)
		if err != nil {
			log.Println(err)
			continue
		}
		for _, item := range items {
			g.Resources = append(g.Resources, g.newSccResource(item.Name, collection.resourceType, scope.attrs))
		}
	}
	if scope.kind != "organization" {
		return
	}
	sources, err := restListResources(ctx, restClient, sccV2BaseURL+scope.parent+"/sources")
	if err != nil {
		log.Println(err)
		return
	}
	for _, source := range sources {
		g.Resources = append(g.Resources, g.newSccResource(source.Name, "google_scc_v2_organization_source", scope.attrs))
		bindings, err := restGetIamPolicy(ctx, restClient, sccV2BaseURL+source.Name+":getIamPolicy")
		if err != nil {
			log.Println(err)
			continue
		}
		g.Resources = append(g.Resources, expandIamPolicy(
			bindings,
			source.Name,
			"google_scc_v2_organization_source_iam_policy",
			"google_scc_v2_organization_source_iam_binding",
			"google_scc_v2_organization_source_iam_member",
			map[string]string{"source": source.Name},
			g.newSccResource,
		)...)
	}
}

// loadManagementScope enumerates the Security Center Management API
// custom modules of one scope through raw REST, in the global
// location. Failures are logged, as with the v2 surface.
func (g *SccGenerator) loadManagementScope(ctx context.Context, restClient *http.Client, scope sccScope) {
	locationParent := sccManagementBaseURL + scope.parent + "/locations/" + sccManagementLocation
	shaType := map[string]string{
		"organization": "google_scc_management_organization_security_health_analytics_custom_module",
		"folder":       "google_scc_management_folder_security_health_analytics_custom_module",
		"project":      "google_scc_management_project_security_health_analytics_custom_module",
	}[scope.kind]
	modules, err := restListResources(ctx, restClient, locationParent+"/securityHealthAnalyticsCustomModules")
	if err != nil {
		log.Println(err)
	} else {
		for _, module := range modules {
			if scope.kind != "organization" && module.AncestorModule != "" {
				continue
			}
			g.Resources = append(g.Resources, g.newSccResource(module.Name, shaType, scope.attrs))
		}
	}
	if scope.kind != "organization" {
		return
	}
	modules, err = restListResources(ctx, restClient, locationParent+"/eventThreatDetectionCustomModules")
	if err != nil {
		log.Println(err)
		return
	}
	for _, module := range modules {
		g.Resources = append(g.Resources, g.newSccResource(module.Name, "google_scc_management_organization_event_threat_detection_custom_module", scope.attrs))
	}
}
