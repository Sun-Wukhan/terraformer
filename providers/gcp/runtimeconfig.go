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

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	runtimeconfig "google.golang.org/api/runtimeconfig/v1beta1"
)

// RuntimeConfigGenerator enumerates Cloud Runtime Configuration via
// the generated runtimeconfig v1beta1 client: configs (with the
// combined google_runtimeconfig_config_iam resource when a config
// carries IAM bindings) and each config's variables.
type RuntimeConfigGenerator struct {
	GCPService
}

func (g *RuntimeConfigGenerator) InitResources() error {
	ctx := context.Background()
	service, err := runtimeconfig.NewService(ctx)
	if err != nil {
		return err
	}
	g.initResources(ctx, service)
	return nil
}

func (g *RuntimeConfigGenerator) newResource(id, resourceType string) terraformutils.Resource {
	return terraformutils.NewResource(
		id, opsLastSegment(id), resourceType, g.ProviderName,
		map[string]string{"project": g.GetArgs()["project"].(string)},
		[]string{}, map[string]interface{}{},
	)
}

func (g *RuntimeConfigGenerator) initResources(ctx context.Context, service *runtimeconfig.Service) {
	project := g.GetArgs()["project"].(string)
	if err := service.Projects.Configs.List("projects/"+project).Pages(ctx, func(page *runtimeconfig.ListConfigsResponse) error {
		for _, cfg := range page.Configs {
			g.Resources = append(g.Resources, g.newResource(cfg.Name, "google_runtimeconfig_config"))
			policy, err := service.Projects.Configs.GetIamPolicy(cfg.Name).Context(ctx).Do()
			if err != nil {
				log.Println(err)
			} else if len(policy.Bindings) > 0 {
				g.Resources = append(g.Resources, g.newResource(cfg.Name, "google_runtimeconfig_config_iam"))
			}
			if err := service.Projects.Configs.Variables.List(cfg.Name).Pages(ctx, func(varPage *runtimeconfig.ListVariablesResponse) error {
				for _, variable := range varPage.Variables {
					g.Resources = append(g.Resources, g.newResource(variable.Name, "google_runtimeconfig_variable"))
				}
				return nil
			}); err != nil {
				log.Println(err)
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
}
