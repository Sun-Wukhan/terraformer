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
	storagetransfer "google.golang.org/api/storagetransfer/v1"
)

type StorageTransferGenerator struct {
	GCPService
}

// InitResources enumerates Storage Transfer Service agent pools in
// the configured project. Provider IDs are the full resource names
// (projects/{p}/agentPools/{pool}).
func (g *StorageTransferGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := storagetransfer.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, svc)
}

func (g *StorageTransferGenerator) initResources(ctx context.Context, svc *storagetransfer.Service) error {
	project := g.GetArgs()["project"].(string)
	_ = g.GetArgs()["region"].(compute.Region)
	return svc.Projects.AgentPools.List("projects/"+project).Pages(ctx, func(page *storagetransfer.ListAgentPoolsResponse) error {
		for _, pool := range page.AgentPools {
			parts := strings.Split(pool.Name, "/")
			g.Resources = append(g.Resources, terraformutils.NewResource(
				pool.Name, parts[len(parts)-1], "google_storage_transfer_agent_pool",
				g.ProviderName, map[string]string{}, []string{}, map[string]interface{}{}))
		}
		return nil
	})
}
