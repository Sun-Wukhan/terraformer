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

// ModelArmorGenerator enumerates the Model Armor surface of one
// project over raw REST — the API has no generated client in the
// pinned google.golang.org/api module: prompt/response templates
// and the per-location floor-setting singleton, in the configured
// region and in the "us" and "eu" multi-regions. Every provider ID
// is the resource's full API name.
type ModelArmorGenerator struct {
	GCPService
}

const modelArmorBaseURL = "https://modelarmor.googleapis.com/v1/"

func (g *ModelArmorGenerator) InitResources() error {
	ctx := context.Background()
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		log.Println(err)
		return nil
	}
	g.initResources(ctx, restClient)
	return nil
}

func (g *ModelArmorGenerator) initResources(ctx context.Context, restClient *http.Client) {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	locations := []string{}
	for _, location := range []string{region, "us", "eu"} {
		if location == "" {
			continue
		}
		seen := false
		for _, existing := range locations {
			if existing == location {
				seen = true
			}
		}
		if !seen {
			locations = append(locations, location)
		}
	}
	for _, location := range locations {
		parent := "projects/" + project + "/locations/" + location
		templates, err := restListResources(ctx, restClient, modelArmorBaseURL+parent+"/templates")
		if err != nil {
			log.Println(err)
		} else {
			for _, template := range templates {
				g.Resources = append(g.Resources, g.newModelArmorResource(template.Name, "google_model_armor_template"))
			}
		}
		floorSetting := parent + "/floorSetting"
		if name, err := aiRestSingletonName(ctx, restClient, modelArmorBaseURL+floorSetting); err != nil {
			log.Println(err)
		} else if name != "" {
			g.Resources = append(g.Resources, g.newModelArmorResource(name, "google_model_armor_floorsetting"))
		}
	}
}

func (g *ModelArmorGenerator) newModelArmorResource(id, resourceType string) terraformutils.Resource {
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
