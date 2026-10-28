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
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	idsv1 "google.golang.org/api/ids/v1"
)

var cloudIdsAllowEmptyValues = []string{"tags."}

var cloudIdsAdditionalFields = map[string]interface{}{}

type CloudIdsGenerator struct {
	GCPService
}

// InitResources enumerates Cloud IDS endpoints in every location the
// API reports for the project.
func (g *CloudIdsGenerator) InitResources() error {
	ctx := context.Background()
	service, err := idsv1.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, service)
}

func (g *CloudIdsGenerator) initResources(ctx context.Context, service *idsv1.Service) error {
	project := g.GetArgs()["project"].(string)
	projectParent := "projects/" + project

	locations, err := listLocationNames(func(pageToken string) ([]string, string, error) {
		resp, err := service.Projects.Locations.List(projectParent).PageToken(pageToken).Context(ctx).Do()
		if err != nil {
			return nil, "", err
		}
		names := []string{}
		for _, location := range resp.Locations {
			names = append(names, location.Name)
		}
		return names, resp.NextPageToken, nil
	})
	if err != nil {
		return err
	}
	for _, locationName := range locations {
		parts := strings.Split(locationName, "/")
		location := parts[len(parts)-1]
		attrs := map[string]string{"project": project, "location": location}
		err := service.Projects.Locations.Endpoints.List(locationName).Pages(ctx, func(page *idsv1.ListEndpointsResponse) error {
			for _, endpoint := range page.Endpoints {
				parts := strings.Split(endpoint.Name, "/")
				g.Resources = append(g.Resources, terraformutils.NewResource(
					endpoint.Name,
					parts[len(parts)-1],
					"google_cloud_ids_endpoint",
					g.ProviderName,
					attrs,
					cloudIdsAllowEmptyValues,
					cloudIdsAdditionalFields,
				))
			}
			return nil
		})
		if err != nil {
			log.Println(err)
		}
	}
	return nil
}
