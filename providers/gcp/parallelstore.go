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
	parallelstorev1 "google.golang.org/api/parallelstore/v1"
)

var parallelstoreAllowEmptyValues = []string{""}

var parallelstoreAdditionalFields = map[string]interface{}{}

// ParallelstoreGenerator enumerates Parallelstore instances in the
// configured region.
type ParallelstoreGenerator struct {
	GCPService
}

func (g *ParallelstoreGenerator) InitResources() error {
	ctx := context.Background()
	storeService, err := parallelstorev1.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, storeService)
}

func (g *ParallelstoreGenerator) initResources(ctx context.Context, storeService *parallelstorev1.Service) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + region
	baseAttrs := map[string]string{"project": project, "location": region}
	return storeService.Projects.Locations.Instances.List(parent).Pages(ctx, func(page *parallelstorev1.ListInstancesResponse) error {
		for _, instance := range page.Instances {
			parts := strings.Split(instance.Name, "/")
			g.Resources = append(g.Resources, terraformutils.NewResource(
				instance.Name,
				parts[len(parts)-1],
				"google_parallelstore_instance",
				g.ProviderName,
				baseAttrs,
				parallelstoreAllowEmptyValues,
				parallelstoreAdditionalFields,
			))
		}
		return nil
	})
}
