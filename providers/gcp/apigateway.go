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
	apigateway "google.golang.org/api/apigateway/v1"
	"google.golang.org/api/compute/v1"
)

var apiGatewayAllowEmptyValues = []string{""}

var apiGatewayAdditionalFields = map[string]interface{}{}

type ApiGatewayGenerator struct {
	GCPService
}

// InitResources enumerates API Gateway: APIs in the global location
// (with their configs), and gateways in the configured region. Each
// API, config and gateway also gets its IAM policy, bindings and
// members expanded from its live IAM policy. Provider IDs are the
// full resource names (projects/{p}/locations/global/apis/{a},
// .../apis/{a}/configs/{c}, projects/{p}/locations/{l}/gateways/{g}).
func (g *ApiGatewayGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := apigateway.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, svc)
}

func (g *ApiGatewayGenerator) initResources(ctx context.Context, svc *apigateway.Service) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	globalParent := "projects/" + project + "/locations/global"
	if err := svc.Projects.Locations.Apis.List(globalParent).Pages(ctx, func(page *apigateway.ApigatewayListApisResponse) error {
		for _, api := range page.Apis {
			g.emitWithIam(api.Name, "google_api_gateway_api", func() (*apigateway.ApigatewayPolicy, error) {
				return svc.Projects.Locations.Apis.GetIamPolicy(api.Name).Context(ctx).Do()
			})
			if err := svc.Projects.Locations.Apis.Configs.List(api.Name).Pages(ctx, func(cpage *apigateway.ApigatewayListApiConfigsResponse) error {
				for _, cfg := range cpage.ApiConfigs {
					g.emitWithIam(cfg.Name, "google_api_gateway_api_config", func() (*apigateway.ApigatewayPolicy, error) {
						return svc.Projects.Locations.Apis.Configs.GetIamPolicy(cfg.Name).Context(ctx).Do()
					})
				}
				return nil
			}); err != nil {
				log.Println(err)
			}
		}
		return nil
	}); err != nil {
		return err
	}
	regionParent := "projects/" + project + "/locations/" + region
	return svc.Projects.Locations.Gateways.List(regionParent).Pages(ctx, func(page *apigateway.ApigatewayListGatewaysResponse) error {
		for _, gw := range page.Gateways {
			g.emitWithIam(gw.Name, "google_api_gateway_gateway", func() (*apigateway.ApigatewayPolicy, error) {
				return svc.Projects.Locations.Gateways.GetIamPolicy(gw.Name).Context(ctx).Do()
			})
		}
		return nil
	})
}

func (g *ApiGatewayGenerator) newResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
	parts := strings.Split(id, "/")
	if attrs == nil {
		attrs = map[string]string{}
	}
	return terraformutils.NewResource(id, parts[len(parts)-1], resourceType, g.ProviderName, attrs,
		apiGatewayAllowEmptyValues, apiGatewayAdditionalFields)
}

func (g *ApiGatewayGenerator) emitWithIam(name, resourceType string, getPolicy func() (*apigateway.ApigatewayPolicy, error)) {
	g.Resources = append(g.Resources, g.newResource(name, resourceType, nil))
	policy, err := getPolicy()
	if err != nil {
		log.Println(err)
		return
	}
	bindings := []iamBinding{}
	for _, b := range policy.Bindings {
		bindings = append(bindings, iamBinding{Role: b.Role, Members: b.Members})
	}
	g.Resources = append(g.Resources, expandIamPolicy(bindings, name,
		resourceType+"_iam_policy", resourceType+"_iam_binding", resourceType+"_iam_member",
		map[string]string{}, g.newResource)...)
}
