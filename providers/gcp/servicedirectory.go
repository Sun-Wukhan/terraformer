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
	svcdirv1 "google.golang.org/api/servicedirectory/v1"
)

var serviceDirectoryAllowEmptyValues = []string{"tags."}

var serviceDirectoryAdditionalFields = map[string]interface{}{}

type ServiceDirectoryGenerator struct {
	GCPService
}

// InitResources enumerates Service Directory resources in every
// location the API reports for the project: namespaces (with the IAM
// policy, bindings and members expanded from each namespace's live
// IAM policy), the services inside each namespace (likewise with
// their IAM trio), and the endpoints inside each service.
func (g *ServiceDirectoryGenerator) InitResources() error {
	ctx := context.Background()
	service, err := svcdirv1.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, service)
}

func (g *ServiceDirectoryGenerator) initResources(ctx context.Context, service *svcdirv1.APIService) error {
	project := g.GetArgs()["project"].(string)
	projectParent := "projects/" + project

	locations, err := listLocationNames(func(pageToken string) ([]string, string, error) {
		resp, err := service.Projects.Locations.List(projectParent).PageToken(pageToken).Context(ctx).Do()
		if err != nil {
			return nil, "", err
		}
		names := []string{}
		for _, location := range resp.Locations {
			names = append(names, location.Name)
		}
		return names, resp.NextPageToken, nil
	})
	if err != nil {
		return err
	}
	for _, locationName := range locations {
		g.loadServiceDirectoryLocation(ctx, service, locationName, project)
	}
	return nil
}

func (g *ServiceDirectoryGenerator) newServiceDirectoryResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
	parts := strings.Split(id, "/")
	return terraformutils.NewResource(
		id,
		parts[len(parts)-1],
		resourceType,
		g.ProviderName,
		attrs,
		serviceDirectoryAllowEmptyValues,
		serviceDirectoryAdditionalFields,
	)
}

func (g *ServiceDirectoryGenerator) loadServiceDirectoryLocation(ctx context.Context, service *svcdirv1.APIService, locationName, project string) {
	parts := strings.Split(locationName, "/")
	location := parts[len(parts)-1]
	attrs := map[string]string{"project": project, "location": location}

	err := service.Projects.Locations.Namespaces.List(locationName).Pages(ctx, func(page *svcdirv1.ListNamespacesResponse) error {
		for _, namespace := range page.Namespaces {
			g.Resources = append(g.Resources, g.newServiceDirectoryResource(namespace.Name, "google_service_directory_namespace", attrs))
			g.loadNamespaceIam(ctx, service, namespace.Name)
			serviceErr := service.Projects.Locations.Namespaces.Services.List(namespace.Name).Pages(ctx, func(page *svcdirv1.ListServicesResponse) error {
				for _, sdService := range page.Services {
					g.Resources = append(g.Resources, g.newServiceDirectoryResource(sdService.Name, "google_service_directory_service", attrs))
					g.loadSdServiceIam(ctx, service, sdService.Name)
					endpointErr := service.Projects.Locations.Namespaces.Services.Endpoints.List(sdService.Name).Pages(ctx, func(page *svcdirv1.ListEndpointsResponse) error {
						for _, endpoint := range page.Endpoints {
							g.Resources = append(g.Resources, g.newServiceDirectoryResource(endpoint.Name, "google_service_directory_endpoint", attrs))
						}
						return nil
					})
					if endpointErr != nil {
						log.Println(endpointErr)
					}
				}
				return nil
			})
			if serviceErr != nil {
				log.Println(serviceErr)
			}
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func sdBindings(policy *svcdirv1.Policy) []iamBinding {
	bindings := []iamBinding{}
	if policy == nil {
		return bindings
	}
	for _, binding := range policy.Bindings {
		bindings = append(bindings, iamBinding{Role: binding.Role, Members: binding.Members})
	}
	return bindings
}

func (g *ServiceDirectoryGenerator) loadNamespaceIam(ctx context.Context, service *svcdirv1.APIService, namespaceName string) {
	policy, err := service.Projects.Locations.Namespaces.GetIamPolicy(namespaceName, &svcdirv1.GetIamPolicyRequest{}).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		return
	}
	g.Resources = append(g.Resources, expandIamPolicy(sdBindings(policy), namespaceName,
		"google_service_directory_namespace_iam_policy", "google_service_directory_namespace_iam_binding", "google_service_directory_namespace_iam_member",
		map[string]string{"namespace": namespaceName},
		func(id, resourceType string, attrs map[string]string) terraformutils.Resource {
			return g.newServiceDirectoryResource(id, resourceType, attrs)
		})...)
}

func (g *ServiceDirectoryGenerator) loadSdServiceIam(ctx context.Context, service *svcdirv1.APIService, serviceName string) {
	policy, err := service.Projects.Locations.Namespaces.Services.GetIamPolicy(serviceName, &svcdirv1.GetIamPolicyRequest{}).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		return
	}
	g.Resources = append(g.Resources, expandIamPolicy(sdBindings(policy), serviceName,
		"google_service_directory_service_iam_policy", "google_service_directory_service_iam_binding", "google_service_directory_service_iam_member",
		map[string]string{"service": serviceName},
		func(id, resourceType string, attrs map[string]string) terraformutils.Resource {
			return g.newServiceDirectoryResource(id, resourceType, attrs)
		})...)
}
