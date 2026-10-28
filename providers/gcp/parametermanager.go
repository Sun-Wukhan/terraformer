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

const parameterManagerBaseURL = "https://parametermanager.googleapis.com/v1/"

type ParameterManagerGenerator struct {
	GCPService
}

// InitResources enumerates Parameter Manager over raw REST (the
// pinned google.golang.org/api has no parametermanager client):
// global parameters and templates in the global location and
// regional parameters in the configured region, each with their
// versions. Provider IDs are the full resource names.
func (g *ParameterManagerGenerator) InitResources() error {
	ctx := context.Background()
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		log.Println(err)
		return nil
	}
	g.initResources(ctx, restClient)
	return nil
}

func (g *ParameterManagerGenerator) initResources(ctx context.Context, restClient *http.Client) {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	globalParent := "projects/" + project + "/locations/global"
	g.listWithVersions(ctx, restClient, globalParent+"/parameters",
		"google_parameter_manager_parameter", "google_parameter_manager_parameter_version")
	g.list(ctx, restClient, globalParent+"/templates", "google_parameter_manager_template")
	regionalParent := "projects/" + project + "/locations/" + region
	g.listWithVersions(ctx, restClient, regionalParent+"/parameters",
		"google_parameter_manager_regional_parameter", "google_parameter_manager_regional_parameter_version")
}

func (g *ParameterManagerGenerator) listWithVersions(ctx context.Context, restClient *http.Client, collection, paramType, versionType string) {
	params, err := restListResources(ctx, restClient, parameterManagerBaseURL+collection)
	if err != nil {
		log.Println(err)
		return
	}
	for _, param := range params {
		g.appendResource(param.Name, paramType)
		g.list(ctx, restClient, param.Name+"/versions", versionType)
	}
}

func (g *ParameterManagerGenerator) list(ctx context.Context, restClient *http.Client, collection, resourceType string) {
	items, err := restListResources(ctx, restClient, parameterManagerBaseURL+collection)
	if err != nil {
		log.Println(err)
		return
	}
	for _, item := range items {
		g.appendResource(item.Name, resourceType)
	}
}

func (g *ParameterManagerGenerator) appendResource(name, resourceType string) {
	parts := strings.Split(name, "/")
	g.Resources = append(g.Resources, terraformutils.NewResource(
		name, parts[len(parts)-1], resourceType, g.ProviderName,
		map[string]string{}, []string{}, map[string]interface{}{}))
}
