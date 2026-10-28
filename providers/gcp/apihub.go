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
	"google.golang.org/api/compute/v1"
)

const apiHubBaseURL = "https://apihub.googleapis.com/v1/"

type ApiHubGenerator struct {
	GCPService
}

// InitResources enumerates API Hub in the configured region over raw
// REST (the pinned google.golang.org/api has no apihub client): the
// API Hub instance, curations, host project registrations, plugins
// and their plugin instances, and runtime project attachments.
// Provider IDs are the full resource names.
func (g *ApiHubGenerator) InitResources() error {
	ctx := context.Background()
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		log.Println(err)
		return nil
	}
	g.initResources(ctx, restClient)
	return nil
}

func (g *ApiHubGenerator) initResources(ctx context.Context, restClient *http.Client) {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + region
	g.list(ctx, restClient, parent+"/apiHubInstances", "google_apihub_api_hub_instance")
	g.list(ctx, restClient, parent+"/curations", "google_apihub_curation")
	g.list(ctx, restClient, parent+"/hostProjectRegistrations", "google_apihub_host_project_registration")
	g.list(ctx, restClient, parent+"/runtimeProjectAttachments", "google_apihub_runtime_project_attachment")
	plugins, err := restListResources(ctx, restClient, apiHubBaseURL+parent+"/plugins")
	if err != nil {
		log.Println(err)
		return
	}
	for _, plugin := range plugins {
		g.appendResource(plugin.Name, "google_apihub_plugin")
		g.list(ctx, restClient, plugin.Name+"/instances", "google_apihub_plugin_instance")
	}
}

func (g *ApiHubGenerator) list(ctx context.Context, restClient *http.Client, collection, resourceType string) {
	items, err := restListResources(ctx, restClient, apiHubBaseURL+collection)
	if err != nil {
		log.Println(err)
		return
	}
	for _, item := range items {
		g.appendResource(item.Name, resourceType)
	}
}

func (g *ApiHubGenerator) appendResource(name, resourceType string) {
	parts := strings.Split(name, "/")
	g.Resources = append(g.Resources, terraformutils.NewResource(
		name, parts[len(parts)-1], resourceType, g.ProviderName,
		map[string]string{}, []string{}, map[string]interface{}{}))
}
