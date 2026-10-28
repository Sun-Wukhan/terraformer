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

// VectorSearchGenerator enumerates the Vertex AI Vector Search 2.0
// surface of one project/location over raw REST — the API
// (vectorsearch.googleapis.com, v1beta1) has no generated client in
// the pinned google.golang.org/api module: collections, each
// collection's data objects, and each collection's indexes. Every
// provider ID is the resource's full API name.
type VectorSearchGenerator struct {
	GCPService
}

const vectorSearchBaseURL = "https://vectorsearch.googleapis.com/v1beta1/"

func (g *VectorSearchGenerator) InitResources() error {
	ctx := context.Background()
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		log.Println(err)
		return nil
	}
	g.initResources(ctx, restClient)
	return nil
}

func (g *VectorSearchGenerator) initResources(ctx context.Context, restClient *http.Client) {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + region
	collections, err := restListResources(ctx, restClient, vectorSearchBaseURL+parent+"/collections")
	if err != nil {
		log.Println(err)
		return
	}
	for _, collection := range collections {
		g.Resources = append(g.Resources, g.newVectorSearchResource(collection.Name, "google_vector_search_collection"))
		dataObjects, err := restListResources(ctx, restClient, vectorSearchBaseURL+collection.Name+"/dataObjects")
		if err != nil {
			log.Println(err)
		} else {
			for _, dataObject := range dataObjects {
				g.Resources = append(g.Resources, g.newVectorSearchResource(dataObject.Name, "google_vector_search_data_object"))
			}
		}
		indexes, err := restListResources(ctx, restClient, vectorSearchBaseURL+collection.Name+"/indexes")
		if err != nil {
			log.Println(err)
		} else {
			for _, index := range indexes {
				g.Resources = append(g.Resources, g.newVectorSearchResource(index.Name, "google_vector_search_index"))
			}
		}
	}
}

func (g *VectorSearchGenerator) newVectorSearchResource(id, resourceType string) terraformutils.Resource {
	parts := strings.Split(id, "/")
	return terraformutils.NewResource(
		id,
		parts[len(parts)-1],
		resourceType,
		g.ProviderName,
		map[string]string{
			"project": g.GetArgs()["project"].(string),
			"region":  g.GetArgs()["region"].(compute.Region).Name,
		},
		[]string{},
		map[string]interface{}{},
	)
}
