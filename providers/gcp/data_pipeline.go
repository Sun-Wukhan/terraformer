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
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"google.golang.org/api/compute/v1"
	datapipelinesv1 "google.golang.org/api/datapipelines/v1"
)

var dataPipelineAllowEmptyValues = []string{""}

var dataPipelineAdditionalFields = map[string]interface{}{}

type DataPipelineGenerator struct {
	GCPService
}

// InitResources enumerates Data Pipelines pipelines for the
// configured project and region. The provider ID is the full API
// resource name, its canonical import format.
func (g *DataPipelineGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := datapipelinesv1.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, svc)
}

func (g *DataPipelineGenerator) initResources(ctx context.Context, svc *datapipelinesv1.Service) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region)
	locationParent := "projects/" + project + "/locations/" + region.Name
	baseAttrs := map[string]string{"project": project, "location": region.Name}
	err := svc.Projects.Locations.Pipelines.List(locationParent).Pages(ctx, func(page *datapipelinesv1.GoogleCloudDatapipelinesV1ListPipelinesResponse) error {
		for _, pipeline := range page.Pipelines {
			parts := strings.Split(pipeline.Name, "/")
			g.Resources = append(g.Resources, terraformutils.NewResource(
				pipeline.Name,
				parts[len(parts)-1],
				"google_data_pipeline_pipeline",
				g.ProviderName,
				baseAttrs,
				dataPipelineAllowEmptyValues,
				dataPipelineAdditionalFields,
			))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	return nil
}
