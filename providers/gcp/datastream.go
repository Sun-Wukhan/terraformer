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
	datastreamv1 "google.golang.org/api/datastream/v1"
)

var datastreamAllowEmptyValues = []string{""}

var datastreamAdditionalFields = map[string]interface{}{}

type DatastreamGenerator struct {
	GCPService
}

// InitResources enumerates Datastream resources for the configured
// project and region: connection profiles, private connections and
// streams. Provider IDs are the full API resource names, the
// canonical import format for all of these.
func (g *DatastreamGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := datastreamv1.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, svc)
}

func (g *DatastreamGenerator) initResources(ctx context.Context, svc *datastreamv1.Service) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region)
	locationParent := "projects/" + project + "/locations/" + region.Name
	baseAttrs := map[string]string{"project": project, "location": region.Name}
	err := svc.Projects.Locations.ConnectionProfiles.List(locationParent).Pages(ctx, func(page *datastreamv1.ListConnectionProfilesResponse) error {
		for _, profile := range page.ConnectionProfiles {
			g.Resources = append(g.Resources, g.newDatastreamResource(profile.Name, "google_datastream_connection_profile", baseAttrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = svc.Projects.Locations.PrivateConnections.List(locationParent).Pages(ctx, func(page *datastreamv1.ListPrivateConnectionsResponse) error {
		for _, connection := range page.PrivateConnections {
			g.Resources = append(g.Resources, g.newDatastreamResource(connection.Name, "google_datastream_private_connection", baseAttrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = svc.Projects.Locations.Streams.List(locationParent).Pages(ctx, func(page *datastreamv1.ListStreamsResponse) error {
		for _, stream := range page.Streams {
			g.Resources = append(g.Resources, g.newDatastreamResource(stream.Name, "google_datastream_stream", baseAttrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	return nil
}

func (g *DatastreamGenerator) newDatastreamResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
	if attrs == nil {
		attrs = map[string]string{}
	}
	parts := strings.Split(id, "/")
	return terraformutils.NewResource(
		id,
		parts[len(parts)-1],
		resourceType,
		g.ProviderName,
		attrs,
		datastreamAllowEmptyValues,
		datastreamAdditionalFields,
	)
}
