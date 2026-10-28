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

// DocumentAiWarehouseGenerator enumerates the Document AI Warehouse
// (Content Warehouse) surface of one project/location over raw REST
// — the API has no generated client in the pinned
// google.golang.org/api module: the location singleton (emitted when
// the location has been initialized) and its document schemas. IDs
// are the resources' full API names as returned by the API (which
// carry the project number, as the provider expects).
type DocumentAiWarehouseGenerator struct {
	GCPService
}

const contentWarehouseBaseURL = "https://contentwarehouse.googleapis.com/v1/"

func (g *DocumentAiWarehouseGenerator) InitResources() error {
	ctx := context.Background()
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		log.Println(err)
		return nil
	}
	g.initResources(ctx, restClient)
	return nil
}

func (g *DocumentAiWarehouseGenerator) initResources(ctx context.Context, restClient *http.Client) {
	project := g.GetArgs()["project"].(string)
	location := g.GetArgs()["region"].(compute.Region).Name
	if location == "" {
		location = "us"
	}
	parent := "projects/" + project + "/locations/" + location
	if name, err := aiRestSingletonName(ctx, restClient, contentWarehouseBaseURL+parent); err != nil {
		log.Println(err)
	} else if name != "" {
		g.Resources = append(g.Resources, g.newWarehouseResource(name, "google_document_ai_warehouse_location"))
	}
	schemas, err := restListResources(ctx, restClient, contentWarehouseBaseURL+parent+"/documentSchemas")
	if err != nil {
		log.Println(err)
		return
	}
	for _, schema := range schemas {
		g.Resources = append(g.Resources, g.newWarehouseResource(schema.Name, "google_document_ai_warehouse_document_schema"))
	}
}

func (g *DocumentAiWarehouseGenerator) newWarehouseResource(id, resourceType string) terraformutils.Resource {
	parts := strings.Split(id, "/")
	return terraformutils.NewResource(
		id,
		parts[len(parts)-1],
		resourceType,
		g.ProviderName,
		map[string]string{"project": g.GetArgs()["project"].(string)},
		[]string{},
		map[string]interface{}{},
	)
}
