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
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"google.golang.org/api/compute/v1"
	connectors "google.golang.org/api/connectors/v1"
)

var connectorsAllowEmptyValues = []string{""}

var connectorsAdditionalFields = map[string]interface{}{}

type IntegrationConnectorsGenerator struct {
	GCPService
}

// InitResources enumerates Integration Connectors: connections and
// endpoint attachments in the configured region, and managed zones
// in the global location. Provider IDs are the full resource names.
func (g *IntegrationConnectorsGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := connectors.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, svc)
}

func (g *IntegrationConnectorsGenerator) initResources(ctx context.Context, svc *connectors.Service) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + region
	if err := svc.Projects.Locations.Connections.List(parent).Pages(ctx, func(page *connectors.ListConnectionsResponse) error {
		for _, item := range page.Connections {
			g.appendResource(item.Name, "google_integration_connectors_connection")
		}
		return nil
	}); err != nil {
		return err
	}
	if err := svc.Projects.Locations.EndpointAttachments.List(parent).Pages(ctx, func(page *connectors.ListEndpointAttachmentsResponse) error {
		for _, item := range page.EndpointAttachments {
			g.appendResource(item.Name, "google_integration_connectors_endpoint_attachment")
		}
		return nil
	}); err != nil {
		return err
	}
	globalParent := "projects/" + project + "/locations/global"
	return svc.Projects.Locations.Global.ManagedZones.List(globalParent).Pages(ctx, func(page *connectors.ListManagedZonesResponse) error {
		for _, item := range page.ManagedZones {
			g.appendResource(item.Name, "google_integration_connectors_managed_zone")
		}
		return nil
	})
}

func (g *IntegrationConnectorsGenerator) appendResource(name, resourceType string) {
	parts := strings.Split(name, "/")
	g.Resources = append(g.Resources, terraformutils.NewResource(
		name, parts[len(parts)-1], resourceType, g.ProviderName,
		map[string]string{}, connectorsAllowEmptyValues, connectorsAdditionalFields))
}
