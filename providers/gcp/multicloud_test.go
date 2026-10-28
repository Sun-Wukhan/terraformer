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
	"testing"

	"google.golang.org/api/compute/v1"
)

// TestContainerMulticloudGenerator feeds canned GKE multicloud
// raw REST responses and verifies Azure/AWS clusters, clients and
// node pools.
func TestContainerMulticloudGenerator(t *testing.T) {
	loc := "projects/test-project/locations/us-central1"
	az := loc + "/azureClusters/az1"
	aws := loc + "/awsClusters/aws1"
	transport := &cannedTransport{routes: []cannedRoute{
		{"/azureClusters/az1/azureNodePools", `{"azureNodePools":[{"name":"` + az + `/azureNodePools/np1"}]}`},
		{"/awsClusters/aws1/awsNodePools", `{"awsNodePools":[{"name":"` + aws + `/awsNodePools/np1"}]}`},
		{"/azureClients", `{"azureClients":[{"name":"` + loc + `/azureClients/client1"}]}`},
		{"/azureClusters", `{"azureClusters":[{"name":"` + az + `"}]}`},
		{"/awsClusters", `{"awsClusters":[{"name":"` + aws + `"}]}`},
	}}
	g := &ContainerMulticloudGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project", "region": compute.Region{Name: "us-central1"}})
	g.initResources(context.Background(), transport.client())
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_container_azure_client":    {loc + "/azureClients/client1"},
		"google_container_azure_cluster":   {az},
		"google_container_azure_node_pool": {az + "/azureNodePools/np1"},
		"google_container_aws_cluster":     {aws},
		"google_container_aws_node_pool":   {aws + "/awsNodePools/np1"},
	})
}
