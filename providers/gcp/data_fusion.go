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
	datafusionv1 "google.golang.org/api/datafusion/v1"
)

var dataFusionAllowEmptyValues = []string{""}

var dataFusionAdditionalFields = map[string]interface{}{}

type DataFusionGenerator struct {
	GCPService
}

// InitResources enumerates Cloud Data Fusion instances for the
// configured project and region, expanding each instance's IAM trio.
// The provider ID is the full API resource name, its canonical
// import format (likewise for its IAM resources).
func (g *DataFusionGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := datafusionv1.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, svc)
}

func (g *DataFusionGenerator) initResources(ctx context.Context, svc *datafusionv1.Service) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region)
	locationParent := "projects/" + project + "/locations/" + region.Name
	baseAttrs := map[string]string{"project": project, "location": region.Name}
	err := svc.Projects.Locations.Instances.List(locationParent).Pages(ctx, func(page *datafusionv1.ListInstancesResponse) error {
		for _, instance := range page.Instances {
			g.Resources = append(g.Resources, g.newDataFusionResource(instance.Name, "google_data_fusion_instance", baseAttrs))
			policy, err := svc.Projects.Locations.Instances.GetIamPolicy(instance.Name).Context(ctx).Do()
			if err != nil {
				log.Println(err)
				continue
			}
			bindings := []iamBinding{}
			for _, binding := range policy.Bindings {
				bindings = append(bindings, iamBinding{Role: binding.Role, Members: binding.Members})
			}
			g.Resources = append(g.Resources, expandIamPolicy(
				bindings,
				instance.Name,
				"google_data_fusion_instance_iam_policy",
				"google_data_fusion_instance_iam_binding",
				"google_data_fusion_instance_iam_member",
				baseAttrs,
				g.newDataFusionResource,
			)...)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	return nil
}

func (g *DataFusionGenerator) newDataFusionResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
	if attrs == nil {
		attrs = map[string]string{}
	}
	parts := strings.Split(id, "/")
	return terraformutils.NewResource(
		id,
		parts[len(parts)-1],
		resourceType,
		g.ProviderName,
		attrs,
		dataFusionAllowEmptyValues,
		dataFusionAdditionalFields,
	)
}
