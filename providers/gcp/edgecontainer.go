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

var edgeContainerAllowEmptyValues = []string{""}

var edgeContainerAdditionalFields = map[string]interface{}{}

const edgeContainerBaseURL = "https://edgecontainer.googleapis.com/v1/"

// EdgeContainerGenerator enumerates Google Distributed Cloud Edge
// (edgecontainer.googleapis.com) in the configured region over raw
// REST — the API has no generated client in the pinned
// google.golang.org/api module: clusters with their node pools, and
// VPN connections.
type EdgeContainerGenerator struct {
	GCPService
}

func (g *EdgeContainerGenerator) InitResources() error {
	ctx := context.Background()
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		return err
	}
	g.initResources(ctx, restClient)
	return nil
}

func (g *EdgeContainerGenerator) initResources(ctx context.Context, restClient *http.Client) {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + region
	baseAttrs := map[string]string{"project": project, "location": region}

	clusters, err := restListResources(ctx, restClient, edgeContainerBaseURL+parent+"/clusters")
	if err != nil {
		log.Println(err)
		return
	}
	for _, cluster := range clusters {
		g.Resources = append(g.Resources, g.newEdgeContainerResource(cluster.Name, "google_edgecontainer_cluster", baseAttrs))
		pools, err := restListResources(ctx, restClient, edgeContainerBaseURL+cluster.Name+"/nodePools")
		if err != nil {
			log.Println(err)
			continue
		}
		for _, pool := range pools {
			g.Resources = append(g.Resources, g.newEdgeContainerResource(pool.Name, "google_edgecontainer_node_pool", baseAttrs))
		}
	}
	connections, err := restListResources(ctx, restClient, edgeContainerBaseURL+parent+"/vpnConnections")
	if err != nil {
		log.Println(err)
		return
	}
	for _, connection := range connections {
		g.Resources = append(g.Resources, g.newEdgeContainerResource(connection.Name, "google_edgecontainer_vpn_connection", baseAttrs))
	}
}

func (g *EdgeContainerGenerator) newEdgeContainerResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
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
		edgeContainerAllowEmptyValues,
		edgeContainerAdditionalFields,
	)
}
