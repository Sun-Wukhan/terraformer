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

var edgeNetworkAllowEmptyValues = []string{""}

var edgeNetworkAdditionalFields = map[string]interface{}{}

const edgeNetworkBaseURL = "https://edgenetwork.googleapis.com/v1/"

// EdgeNetworkGenerator enumerates Google Distributed Cloud Edge
// networking (edgenetwork.googleapis.com) in the configured region
// over raw REST — the API has no generated client in the pinned
// google.golang.org/api module: networks with their subnets, and
// interconnect attachments.
type EdgeNetworkGenerator struct {
	GCPService
}

func (g *EdgeNetworkGenerator) InitResources() error {
	ctx := context.Background()
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		return err
	}
	g.initResources(ctx, restClient)
	return nil
}

func (g *EdgeNetworkGenerator) initResources(ctx context.Context, restClient *http.Client) {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + region
	baseAttrs := map[string]string{"project": project, "location": region}

	networks, err := restListResources(ctx, restClient, edgeNetworkBaseURL+parent+"/networks")
	if err != nil {
		log.Println(err)
		return
	}
	for _, network := range networks {
		g.Resources = append(g.Resources, g.newEdgeNetworkResource(network.Name, "google_edgenetwork_network", baseAttrs))
		subnets, err := restListResources(ctx, restClient, edgeNetworkBaseURL+network.Name+"/subnets")
		if err != nil {
			log.Println(err)
			continue
		}
		for _, subnet := range subnets {
			g.Resources = append(g.Resources, g.newEdgeNetworkResource(subnet.Name, "google_edgenetwork_subnet", baseAttrs))
		}
	}
	attachments, err := restListResources(ctx, restClient, edgeNetworkBaseURL+parent+"/interconnectAttachments")
	if err != nil {
		log.Println(err)
		return
	}
	for _, attachment := range attachments {
		g.Resources = append(g.Resources, g.newEdgeNetworkResource(attachment.Name, "google_edgenetwork_interconnect_attachment", baseAttrs))
	}
}

func (g *EdgeNetworkGenerator) newEdgeNetworkResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
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
		edgeNetworkAllowEmptyValues,
		edgeNetworkAdditionalFields,
	)
}
