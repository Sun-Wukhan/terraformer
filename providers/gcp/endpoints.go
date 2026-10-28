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
	svcmgmtv1 "google.golang.org/api/servicemanagement/v1"
)

var endpointsAllowEmptyValues = []string{"tags."}

var endpointsAdditionalFields = map[string]interface{}{}

type EndpointsGenerator struct {
	GCPService
}

// InitResources enumerates Cloud Endpoints services produced by the
// project through the Service Management API (the services list is
// filtered by producer project), with the IAM policy, bindings and
// members expanded from each service's live IAM policy.
//
// Deliberately not generated:
//   - google_endpoints_service_consumers_iam_policy,
//     google_endpoints_service_consumers_iam_binding and
//     google_endpoints_service_consumers_iam_member (the audit's
//     google_endpoints_service_consumers_iam line): the Service
//     Management API exposes only getIamPolicy/setIamPolicy/
//     testIamPermissions on services/{service}/consumers/{consumer};
//     there is no list method for a service's consumers, so the set
//     of consumer projects holding IAM bindings on a service cannot
//     be enumerated.
func (g *EndpointsGenerator) InitResources() error {
	ctx := context.Background()
	service, err := svcmgmtv1.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, service)
}

func (g *EndpointsGenerator) initResources(ctx context.Context, service *svcmgmtv1.APIService) error {
	project := g.GetArgs()["project"].(string)

	err := service.Services.List().ProducerProjectId(project).Pages(ctx, func(page *svcmgmtv1.ListServicesResponse) error {
		for _, managedService := range page.Services {
			if managedService.ProducerProjectId != "" && managedService.ProducerProjectId != project {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewResource(
				managedService.ServiceName,
				managedService.ServiceName,
				"google_endpoints_service",
				g.ProviderName,
				map[string]string{"project": project, "service_name": managedService.ServiceName},
				endpointsAllowEmptyValues,
				endpointsAdditionalFields,
			))
			g.loadEndpointsServiceIam(ctx, service, managedService.ServiceName)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (g *EndpointsGenerator) loadEndpointsServiceIam(ctx context.Context, service *svcmgmtv1.APIService, serviceName string) {
	policy, err := service.Services.GetIamPolicy("services/"+serviceName, &svcmgmtv1.GetIamPolicyRequest{}).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		return
	}
	bindings := []iamBinding{}
	for _, binding := range policy.Bindings {
		bindings = append(bindings, iamBinding{Role: binding.Role, Members: binding.Members})
	}
	g.Resources = append(g.Resources, expandIamPolicy(bindings, serviceName,
		"google_endpoints_service_iam_policy", "google_endpoints_service_iam_binding", "google_endpoints_service_iam_member",
		map[string]string{"service_name": serviceName},
		func(id, resourceType string, attrs map[string]string) terraformutils.Resource {
			return terraformutils.NewResource(id, serviceName, resourceType, g.ProviderName, attrs, endpointsAllowEmptyValues, endpointsAdditionalFields)
		})...)
}
