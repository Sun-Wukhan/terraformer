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
	workflows "google.golang.org/api/workflows/v1"
)

// WorkflowsGenerator enumerates Google Cloud Workflows in the
// configured region via the generated workflows v1 client.
type WorkflowsGenerator struct {
	GCPService
}

func (g *WorkflowsGenerator) InitResources() error {
	ctx := context.Background()
	service, err := workflows.NewService(ctx)
	if err != nil {
		return err
	}
	g.initResources(ctx, service)
	return nil
}

func (g *WorkflowsGenerator) initResources(ctx context.Context, service *workflows.Service) {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + region
	if err := service.Projects.Locations.Workflows.List(parent).Pages(ctx, func(page *workflows.ListWorkflowsResponse) error {
		for _, workflow := range page.Workflows {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				workflow.Name, opsLastSegment(workflow.Name), "google_workflows_workflow", g.ProviderName,
				map[string]string{"project": project, "region": region},
				[]string{}, map[string]interface{}{},
			))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
}
