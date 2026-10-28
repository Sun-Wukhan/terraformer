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

	blockchain "google.golang.org/api/blockchainnodeengine/v1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

func TestBlockchainNodeEngineGeneratorEnumeration(t *testing.T) {
	ctx := context.Background()
	parent := "projects/test-project/locations/us-central1"
	node := parent + "/blockchainNodes/node1"
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: "/blockchainNodes", body: `{"blockchainNodes":[{"name":"` + node + `"}]}`},
	}}
	svc, err := blockchain.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	g := &BlockchainNodeEngineGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project", "region": compute.Region{Name: "us-central1"}})
	if err := g.initResources(ctx, svc); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_blockchain_node_engine_blockchain_nodes": {node},
	})
}
