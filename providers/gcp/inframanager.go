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
	"google.golang.org/api/compute/v1"
	config "google.golang.org/api/config/v1"
)

// InfraManagerGenerator enumerates Infrastructure Manager
// deployments (google_config_deployment) in the configured region
// via the generated config v1 client.
type InfraManagerGenerator struct {
	GCPService
}

func (g *InfraManagerGenerator) InitResources() error {
	ctx := context.Background()
	service, err := config.NewService(ctx)
	if err != nil {
		return err
	}
	g.initResources(ctx, service)
	return nil
}

func (g *InfraManagerGenerator) initResources(ctx context.Context, service *config.Service) {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + region
	if err := service.Projects.Locations.Deployments.List(parent).Pages(ctx, func(page *config.ListDeploymentsResponse) error {
		for _, deployment := range page.Deployments {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				deployment.Name, opsLastSegment(deployment.Name), "google_config_deployment", g.ProviderName,
				map[string]string{"project": project, "location": region},
				[]string{}, map[string]interface{}{},
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
}
