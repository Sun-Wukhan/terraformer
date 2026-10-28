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
	"google.golang.org/api/compute/v1"
	developerconnect "google.golang.org/api/developerconnect/v1"
)

var developerConnectAllowEmptyValues = []string{""}

var developerConnectAdditionalFields = map[string]interface{}{}

type DeveloperConnectGenerator struct {
	GCPService
}

// InitResources enumerates Developer Connect in the configured
// region: account connectors, connections and the git repository
// links under each connection. Provider IDs are the full resource
// names.
//
// Deliberately not generated:
//   - google_developer_connect_insights_config: the Developer Connect
//     API discovery document exposes no insightsConfigs collection
//     (no list or get surface in the generated client), so insights
//     configs cannot be discovered.
func (g *DeveloperConnectGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := developerconnect.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, svc)
}

func (g *DeveloperConnectGenerator) initResources(ctx context.Context, svc *developerconnect.Service) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + region
	if err := svc.Projects.Locations.AccountConnectors.List(parent).Pages(ctx, func(page *developerconnect.ListAccountConnectorsResponse) error {
		for _, item := range page.AccountConnectors {
			g.appendResource(item.Name, "google_developer_connect_account_connector")
		}
		return nil
	}); err != nil {
		return err
	}
	return svc.Projects.Locations.Connections.List(parent).Pages(ctx, func(page *developerconnect.ListConnectionsResponse) error {
		for _, conn := range page.Connections {
			g.appendResource(conn.Name, "google_developer_connect_connection")
			if err := svc.Projects.Locations.Connections.GitRepositoryLinks.List(conn.Name).Pages(ctx, func(lpage *developerconnect.ListGitRepositoryLinksResponse) error {
				for _, link := range lpage.GitRepositoryLinks {
					g.appendResource(link.Name, "google_developer_connect_git_repository_link")
				}
				return nil
			}); err != nil {
				log.Println(err)
			}
		}
		return nil
	})
}

func (g *DeveloperConnectGenerator) appendResource(name, resourceType string) {
	parts := strings.Split(name, "/")
	g.Resources = append(g.Resources, terraformutils.NewResource(
		name, parts[len(parts)-1], resourceType, g.ProviderName,
		map[string]string{}, developerConnectAllowEmptyValues, developerConnectAdditionalFields))
}
