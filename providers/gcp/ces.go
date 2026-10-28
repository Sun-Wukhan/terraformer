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
)

// CesGenerator enumerates the CX Agent Studio (Gemini Enterprise
// for Customer Experience, ces.googleapis.com) surface of one
// project/location over raw REST — the API has no generated client
// in the pinned google.golang.org/api module, and its discovery
// document is v1beta: the per-location security settings singleton,
// apps, and each app's agents, deployments, evaluations, examples,
// guardrails, tools, toolsets and versions. An app whose rootAgent
// is set also yields its root-agent association, which the provider
// keys by the app's own name. Every other provider ID is the
// resource's full API name.
type CesGenerator struct {
	GCPService
}

const cesBaseURL = "https://ces.googleapis.com/v1beta/"

// cesAppCollections maps each per-app collection segment onto its
// provider resource type.
var cesAppCollections = []struct {
	segment      string
	resourceType string
}{
	{"agents", "google_ces_agent"},
	{"deployments", "google_ces_deployment"},
	{"evaluations", "google_ces_evaluation"},
	{"examples", "google_ces_example"},
	{"guardrails", "google_ces_guardrail"},
	{"tools", "google_ces_tool"},
	{"toolsets", "google_ces_toolset"},
	{"versions", "google_ces_app_version"},
}

func (g *CesGenerator) InitResources() error {
	ctx := context.Background()
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		log.Println(err)
		return nil
	}
	g.initResources(ctx, restClient)
	return nil
}

func (g *CesGenerator) cesLocation() string {
	region := g.GetArgs()["region"].(compute.Region).Name
	if region == "" {
		return "us"
	}
	return region
}

type cesApp struct {
	Name      string `json:"name"`
	RootAgent string `json:"rootAgent"`
}

func (g *CesGenerator) initResources(ctx context.Context, restClient *http.Client) {
	project := g.GetArgs()["project"].(string)
	parent := "projects/" + project + "/locations/" + g.cesLocation()

	securitySettings := parent + "/securitySettings"
	if name, err := aiRestSingletonName(ctx, restClient, cesBaseURL+securitySettings); err != nil {
		log.Println(err)
	} else if name != "" {
		g.Resources = append(g.Resources, g.newCesResource(name, "google_ces_security_settings"))
	}

	apps, err := g.cesListApps(ctx, restClient, cesBaseURL+parent+"/apps")
	if err != nil {
		log.Println(err)
		return
	}
	for _, app := range apps {
		g.Resources = append(g.Resources, g.newCesResource(app.Name, "google_ces_app"))
		if app.RootAgent != "" {
			g.Resources = append(g.Resources, g.newCesResource(app.Name, "google_ces_app_root_agent_association"))
		}
		for _, collection := range cesAppCollections {
			items, err := restListResources(ctx, restClient, cesBaseURL+app.Name+"/"+collection.segment)
			if err != nil {
				log.Println(err)
				continue
			}
			for _, item := range items {
				g.Resources = append(g.Resources, g.newCesResource(item.Name, collection.resourceType))
			}
		}
	}
}

// cesListApps pages the apps collection, keeping the rootAgent
// field that restListResources discards.
func (g *CesGenerator) cesListApps(ctx context.Context, restClient *http.Client, url string) ([]cesApp, error) {
	apps := []cesApp{}
	pageToken := ""
	for {
		pageURL := url
		if pageToken != "" {
			pageURL += "?pageToken=" + pageToken
		}
		body, err := restDo(ctx, restClient, http.MethodGet, pageURL, nil)
		if err != nil {
			return nil, err
		}
		var page struct {
			Apps          []cesApp `json:"apps"`
			NextPageToken string   `json:"nextPageToken"`
		}
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		apps = append(apps, page.Apps...)
		if page.NextPageToken == "" {
			return apps, nil
		}
		pageToken = page.NextPageToken
	}
}

func (g *CesGenerator) newCesResource(id, resourceType string) terraformutils.Resource {
	parts := strings.Split(id, "/")
	return terraformutils.NewResource(
		id,
		parts[len(parts)-1],
		resourceType,
		g.ProviderName,
		map[string]string{
			"project":  g.GetArgs()["project"].(string),
			"location": g.cesLocation(),
		},
		[]string{},
		map[string]interface{}{},
	)
}
