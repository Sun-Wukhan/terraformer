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

package aws

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalogappregistry"
)

var servicecatalogappregistryAllowEmptyValues = []string{"tags."}

// ServiceCatalogAppRegistryGenerator covers AppRegistry applications, attribute
// groups and attribute group associations.
// Provider import IDs (verified): application = ID; attribute group = ID;
// association = "<application ID>,<attribute group ID>".
type ServiceCatalogAppRegistryGenerator struct {
	AWSService
}

func (g *ServiceCatalogAppRegistryGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := servicecatalogappregistry.NewFromConfig(config)

	ap := servicecatalogappregistry.NewListApplicationsPaginator(svc, &servicecatalogappregistry.ListApplicationsInput{})
	var appIDs []string
	for ap.HasMorePages() {
		page, err := ap.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, app := range page.Applications {
			id := StringValue(app.Id)
			appIDs = append(appIDs, id)
			name := StringValue(app.Name)
			if name == "" {
				name = id
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, name, "aws_servicecatalogappregistry_application", "aws", servicecatalogappregistryAllowEmptyValues))
		}
	}
	gp := servicecatalogappregistry.NewListAttributeGroupsPaginator(svc, &servicecatalogappregistry.ListAttributeGroupsInput{})
	for gp.HasMorePages() {
		page, err := gp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, grp := range page.AttributeGroups {
			id := StringValue(grp.Id)
			name := StringValue(grp.Name)
			if name == "" {
				name = id
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, name, "aws_servicecatalogappregistry_attribute_group", "aws", servicecatalogappregistryAllowEmptyValues))
		}
	}
	for _, appID := range appIDs {
		out, err := svc.ListAttributeGroupsForApplication(context.TODO(),
			&servicecatalogappregistry.ListAttributeGroupsForApplicationInput{Application: &appID})
		if err != nil {
			continue
		}
		for _, grp := range out.AttributeGroupsDetails {
			id := appID + "," + StringValue(grp.Id)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_servicecatalogappregistry_attribute_group_association", "aws", servicecatalogappregistryAllowEmptyValues))
		}
	}
	return nil
}
