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
	"testing"

	apigateway "google.golang.org/api/apigateway/v1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

func TestApiGatewayGeneratorEnumeration(t *testing.T) {
	ctx := context.Background()
	apiName := "projects/test-project/locations/global/apis/my-api"
	cfgName := apiName + "/configs/my-config"
	gwName := "projects/test-project/locations/us-central1/gateways/my-gw"
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: ":getIamPolicy", body: `{"bindings":[{"role":"roles/apigateway.viewer","members":["user:a@example.com"]}]}`},
		{substr: "/configs", body: `{"apiConfigs":[{"name":"` + cfgName + `"}]}`},
		{substr: "/gateways", body: `{"gateways":[{"name":"` + gwName + `"}]}`},
		{substr: "/apis", body: `{"apis":[{"name":"` + apiName + `"}]}`},
	}}
	svc, err := apigateway.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	g := &ApiGatewayGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project", "region": compute.Region{Name: "us-central1"}})
	if err := g.initResources(ctx, svc); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_api_gateway_api":                    {apiName},
		"google_api_gateway_api_iam_policy":         {apiName},
		"google_api_gateway_api_iam_binding":        {apiName},
		"google_api_gateway_api_iam_member":         {apiName},
		"google_api_gateway_api_config":             {cfgName},
		"google_api_gateway_api_config_iam_policy":  {cfgName},
		"google_api_gateway_api_config_iam_binding": {cfgName},
		"google_api_gateway_api_config_iam_member":  {cfgName},
		"google_api_gateway_gateway":                {gwName},
		"google_api_gateway_gateway_iam_policy":     {gwName},
		"google_api_gateway_gateway_iam_binding":    {gwName},
		"google_api_gateway_gateway_iam_member":     {gwName},
	})
}
