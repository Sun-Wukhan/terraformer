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

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/compute/v1"
)

// ObservabilityGenerator enumerates the Google Cloud Observability
// API (observability.googleapis.com/v1) over raw REST: the pinned
// generated client only carries the legacy Scope surface, while
// buckets, links, trace scopes and the per-parent settings singletons
// exist only in the REST discovery document. Collections are listed
// in the configured region and in the global location; settings are
// probed with a get at project scope and at every ancestor folder
// and organization scope resolved through CRM ancestry.
type ObservabilityGenerator struct {
	GCPService
}

const observabilityBaseURL = "https://observability.googleapis.com/v1/"

func (g *ObservabilityGenerator) InitResources() error {
	ctx := context.Background()
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		return err
	}
	crmService, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		log.Println(err)
	}
	g.initResources(ctx, restClient, crmService)
	return nil
}

func (g *ObservabilityGenerator) newResource(id, resourceType string) terraformutils.Resource {
	return terraformutils.NewResource(
		id, opsLastSegment(id), resourceType, g.ProviderName,
		map[string]string{}, []string{}, map[string]interface{}{},
	)
}

func (g *ObservabilityGenerator) initResources(ctx context.Context, restClient *http.Client, crmService *cloudresourcemanager.Service) {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	for _, location := range []string{region, "global"} {
		parent := "projects/" + project + "/locations/" + location
		buckets, err := restListResources(ctx, restClient, observabilityBaseURL+parent+"/buckets")
		if err != nil {
			log.Println(err)
			continue
		}
		for _, bucket := range buckets {
			g.Resources = append(g.Resources, g.newResource(bucket.Name, "google_observability_bucket"))
			links, err := restListResources(ctx, restClient, observabilityBaseURL+bucket.Name+"/links")
			if err != nil {
				log.Println(err)
				continue
			}
			for _, link := range links {
				g.Resources = append(g.Resources, g.newResource(link.Name, "google_observability_link"))
			}
		}
		scopes, err := restListResources(ctx, restClient, observabilityBaseURL+parent+"/traceScopes")
		if err != nil {
			log.Println(err)
			continue
		}
		for _, scope := range scopes {
			g.Resources = append(g.Resources, g.newResource(scope.Name, "google_observability_trace_scope"))
		}
	}
	for _, scope := range opsScopes(ctx, crmService, project) {
		settingTypes := map[string]string{
			"project":      "google_observability_project_settings",
			"folder":       "google_observability_folder_settings",
			"organization": "google_observability_organization_settings",
		}
		if _, err := restDo(ctx, restClient, http.MethodGet, observabilityBaseURL+scope.parent+"/settings", nil); err == nil {
			g.Resources = append(g.Resources, g.newResource(scope.parent, settingTypes[scope.kind]))
		}
	}
}
