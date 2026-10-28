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
	looker "google.golang.org/api/looker/v1"
)

type LookerGenerator struct {
	GCPService
}

// InitResources enumerates Looker (Google Cloud core) instances in
// the configured region. Provider IDs are the full resource names.
func (g *LookerGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := looker.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, svc)
}

func (g *LookerGenerator) initResources(ctx context.Context, svc *looker.Service) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + region
	return svc.Projects.Locations.Instances.List(parent).Pages(ctx, func(page *looker.ListInstancesResponse) error {
		for _, inst := range page.Instances {
			parts := strings.Split(inst.Name, "/")
			g.Resources = append(g.Resources, terraformutils.NewResource(
				inst.Name, parts[len(parts)-1], "google_looker_instance",
				g.ProviderName, map[string]string{}, []string{}, map[string]interface{}{}))
		}
		return nil
	})
}
