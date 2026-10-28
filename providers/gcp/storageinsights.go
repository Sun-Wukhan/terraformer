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

const storageInsightsBaseURL = "https://storageinsights.googleapis.com/v1/"

type StorageInsightsGenerator struct {
	GCPService
}

// InitResources enumerates Storage Insights dataset configs and
// report configs in the configured region over raw REST (the pinned
// google.golang.org/api has no storageinsights client). Provider
// IDs are the full resource names.
func (g *StorageInsightsGenerator) InitResources() error {
	ctx := context.Background()
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		log.Println(err)
		return nil
	}
	g.initResources(ctx, restClient)
	return nil
}

func (g *StorageInsightsGenerator) initResources(ctx context.Context, restClient *http.Client) {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + region
	for _, coll := range []struct {
		segment string
		type_   string
	}{
		{"datasetConfigs", "google_storage_insights_dataset_config"},
		{"reportConfigs", "google_storage_insights_report_config"},
	} {
		items, err := restListResources(ctx, restClient, storageInsightsBaseURL+parent+"/"+coll.segment)
		if err != nil {
			log.Println(err)
			continue
		}
		for _, item := range items {
			parts := strings.Split(item.Name, "/")
			g.Resources = append(g.Resources, terraformutils.NewResource(
				item.Name, parts[len(parts)-1], coll.type_,
				g.ProviderName, map[string]string{}, []string{}, map[string]interface{}{}))
		}
	}
}
