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
	blockchain "google.golang.org/api/blockchainnodeengine/v1"
	"google.golang.org/api/compute/v1"
)

type BlockchainNodeEngineGenerator struct {
	GCPService
}

// InitResources enumerates Blockchain Node Engine nodes in the
// configured region. Provider IDs are the full resource names.
func (g *BlockchainNodeEngineGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := blockchain.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, svc)
}

func (g *BlockchainNodeEngineGenerator) initResources(ctx context.Context, svc *blockchain.Service) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + region
	return svc.Projects.Locations.BlockchainNodes.List(parent).Pages(ctx, func(page *blockchain.ListBlockchainNodesResponse) error {
		for _, n := range page.BlockchainNodes {
			parts := strings.Split(n.Name, "/")
			g.Resources = append(g.Resources, terraformutils.NewResource(
				n.Name, parts[len(parts)-1], "google_blockchain_node_engine_blockchain_nodes",
				g.ProviderName, map[string]string{}, []string{}, map[string]interface{}{}))
		}
		return nil
	})
}
