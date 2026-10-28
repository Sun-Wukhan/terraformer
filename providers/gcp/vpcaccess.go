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
	vpcaccessv1 "google.golang.org/api/vpcaccess/v1"
)

var vpcAccessAllowEmptyValues = []string{"tags."}

var vpcAccessAdditionalFields = map[string]interface{}{}

type VpcAccessGenerator struct {
	GCPService
}

// InitResources enumerates the Serverless VPC Access connectors of
// the configured region.
func (g *VpcAccessGenerator) InitResources() error {
	ctx := context.Background()
	service, err := vpcaccessv1.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, service)
}

func (g *VpcAccessGenerator) initResources(ctx context.Context, service *vpcaccessv1.Service) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region)
	locationParent := "projects/" + project + "/locations/" + region.Name
	attrs := map[string]string{"project": project, "region": region.Name}

	err := service.Projects.Locations.Connectors.List(locationParent).Pages(ctx, func(page *vpcaccessv1.ListConnectorsResponse) error {
		for _, connector := range page.Connectors {
			parts := strings.Split(connector.Name, "/")
			g.Resources = append(g.Resources, terraformutils.NewResource(
				connector.Name,
				parts[len(parts)-1],
				"google_vpc_access_connector",
				g.ProviderName,
				attrs,
				vpcAccessAllowEmptyValues,
				vpcAccessAdditionalFields,
			))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	return nil
}
