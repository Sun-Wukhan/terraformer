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

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	cloudasset "google.golang.org/api/cloudasset/v1"
	"google.golang.org/api/cloudresourcemanager/v1"
)

// CloudAssetGenerator enumerates Cloud Asset Inventory feeds at
// project scope and, through CRM ancestry, at folder and
// organization scope, via the generated cloudasset v1 client.
type CloudAssetGenerator struct {
	GCPService
}

func (g *CloudAssetGenerator) InitResources() error {
	ctx := context.Background()
	service, err := cloudasset.NewService(ctx)
	if err != nil {
		return err
	}
	crmService, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		log.Println(err)
	}
	g.initResources(ctx, service, crmService)
	return nil
}

func (g *CloudAssetGenerator) initResources(ctx context.Context, service *cloudasset.Service, crmService *cloudresourcemanager.Service) {
	project := g.GetArgs()["project"].(string)
	feedTypes := map[string]string{
		"project":      "google_cloud_asset_project_feed",
		"folder":       "google_cloud_asset_folder_feed",
		"organization": "google_cloud_asset_organization_feed",
	}
	for _, scope := range opsScopes(ctx, crmService, project) {
		// The feeds list call has no paging surface in the generated
		// client (its response carries no page token).
		page, err := service.Feeds.List(scope.parent).Context(ctx).Do()
		if err != nil {
			log.Println(err)
			continue
		}
		for _, feed := range page.Feeds {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				feed.Name, opsLastSegment(feed.Name), feedTypes[scope.kind], g.ProviderName,
				map[string]string{}, []string{}, map[string]interface{}{},
			))
		}
	}
}
