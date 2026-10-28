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

var multicloudAllowEmptyValues = []string{""}

var multicloudAdditionalFields = map[string]interface{}{}

const multicloudBaseURL = "https://gkemulticloud.googleapis.com/v1/"

// ContainerMulticloudGenerator enumerates GKE multicloud clusters
// in the configured region over raw REST: the GKE multicloud API
// (gkemulticloud.googleapis.com) has no generated client in the
// pinned google.golang.org/api module. It covers Azure clients,
// Azure clusters with their node pools, and AWS clusters with
// their node pools.
type ContainerMulticloudGenerator struct {
	GCPService
}

func (g *ContainerMulticloudGenerator) InitResources() error {
	ctx := context.Background()
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		return err
	}
	g.initResources(ctx, restClient)
	return nil
}

func (g *ContainerMulticloudGenerator) initResources(ctx context.Context, restClient *http.Client) {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + region
	baseAttrs := map[string]string{"project": project, "location": region}

	g.loadMulticloudList(ctx, restClient, parent+"/azureClients", "google_container_azure_client", baseAttrs)
	for _, cluster := range g.loadMulticloudList(ctx, restClient, parent+"/azureClusters", "google_container_azure_cluster", baseAttrs) {
		g.loadMulticloudList(ctx, restClient, cluster+"/azureNodePools", "google_container_azure_node_pool", baseAttrs)
	}
	for _, cluster := range g.loadMulticloudList(ctx, restClient, parent+"/awsClusters", "google_container_aws_cluster", baseAttrs) {
		g.loadMulticloudList(ctx, restClient, cluster+"/awsNodePools", "google_container_aws_node_pool", baseAttrs)
	}
}

// loadMulticloudList lists one multicloud collection, emits each
// item as a resource of resourceType, and returns the item names
// (for child enumeration). Failures are logged, not returned: one
// cloud's API surface being unavailable must not lose the other.
func (g *ContainerMulticloudGenerator) loadMulticloudList(ctx context.Context, restClient *http.Client, collectionPath, resourceType string, baseAttrs map[string]string) []string {
	items, err := restListResources(ctx, restClient, multicloudBaseURL+collectionPath)
	if err != nil {
		log.Println(err)
		return nil
	}
	names := []string{}
	for _, item := range items {
		g.Resources = append(g.Resources, g.newMulticloudResource(item.Name, resourceType, baseAttrs))
		names = append(names, item.Name)
	}
	return names
}

func (g *ContainerMulticloudGenerator) newMulticloudResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
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
		multicloudAllowEmptyValues,
		multicloudAdditionalFields,
	)
}
