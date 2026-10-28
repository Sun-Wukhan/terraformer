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
	beyondcorpv1 "google.golang.org/api/beyondcorp/v1"
)

var beyondcorpAllowEmptyValues = []string{"tags."}

var beyondcorpAdditionalFields = map[string]interface{}{}

type BeyondcorpGenerator struct {
	GCPService
}

// InitResources enumerates BeyondCorp resources in every location the
// API reports for the project: security gateways (with the IAM policy,
// bindings and members expanded from each gateway's live IAM policy)
// and the applications of each gateway (likewise with their IAM trio).
func (g *BeyondcorpGenerator) InitResources() error {
	ctx := context.Background()
	service, err := beyondcorpv1.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, service)
}

func (g *BeyondcorpGenerator) initResources(ctx context.Context, service *beyondcorpv1.Service) error {
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
		g.loadBeyondcorpLocation(ctx, service, locationName, project)
	}
	return nil
}

func (g *BeyondcorpGenerator) newBeyondcorpResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
	parts := strings.Split(id, "/")
	return terraformutils.NewResource(
		id,
		parts[len(parts)-1],
		resourceType,
		g.ProviderName,
		attrs,
		beyondcorpAllowEmptyValues,
		beyondcorpAdditionalFields,
	)
}

func (g *BeyondcorpGenerator) loadBeyondcorpLocation(ctx context.Context, service *beyondcorpv1.Service, locationName, project string) {
	parts := strings.Split(locationName, "/")
	location := parts[len(parts)-1]
	attrs := map[string]string{"project": project, "location": location}

	err := service.Projects.Locations.SecurityGateways.List(locationName).Pages(ctx, func(page *beyondcorpv1.GoogleCloudBeyondcorpSecuritygatewaysV1ListSecurityGatewaysResponse) error {
		for _, gateway := range page.SecurityGateways {
			g.Resources = append(g.Resources, g.newBeyondcorpResource(gateway.Name, "google_beyondcorp_security_gateway", attrs))
			g.loadSecurityGatewayIam(ctx, service, gateway.Name)
			appErr := service.Projects.Locations.SecurityGateways.Applications.List(gateway.Name).Pages(ctx, func(page *beyondcorpv1.GoogleCloudBeyondcorpSecuritygatewaysV1ListApplicationsResponse) error {
				for _, application := range page.Applications {
					g.Resources = append(g.Resources, g.newBeyondcorpResource(application.Name, "google_beyondcorp_security_gateway_application", attrs))
					g.loadSecurityGatewayApplicationIam(ctx, service, application.Name)
				}
				return nil
			})
			if appErr != nil {
				log.Println(appErr)
			}
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func beyondcorpBindings(policy *beyondcorpv1.GoogleIamV1Policy) []iamBinding {
	bindings := []iamBinding{}
	if policy == nil {
		return bindings
	}
	for _, binding := range policy.Bindings {
		bindings = append(bindings, iamBinding{Role: binding.Role, Members: binding.Members})
	}
	return bindings
}

func (g *BeyondcorpGenerator) loadSecurityGatewayIam(ctx context.Context, service *beyondcorpv1.Service, gatewayName string) {
	policy, err := service.Projects.Locations.SecurityGateways.GetIamPolicy(gatewayName).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		return
	}
	g.Resources = append(g.Resources, expandIamPolicy(beyondcorpBindings(policy), gatewayName,
		"google_beyondcorp_security_gateway_iam_policy", "google_beyondcorp_security_gateway_iam_binding", "google_beyondcorp_security_gateway_iam_member",
		map[string]string{"security_gateway_id": gatewayName},
		func(id, resourceType string, attrs map[string]string) terraformutils.Resource {
			return g.newBeyondcorpResource(id, resourceType, attrs)
		})...)
}

func (g *BeyondcorpGenerator) loadSecurityGatewayApplicationIam(ctx context.Context, service *beyondcorpv1.Service, applicationName string) {
	policy, err := service.Projects.Locations.SecurityGateways.Applications.GetIamPolicy(applicationName).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		return
	}
	g.Resources = append(g.Resources, expandIamPolicy(beyondcorpBindings(policy), applicationName,
		"google_beyondcorp_security_gateway_application_iam_policy", "google_beyondcorp_security_gateway_application_iam_binding", "google_beyondcorp_security_gateway_application_iam_member",
		map[string]string{"application_id": applicationName},
		func(id, resourceType string, attrs map[string]string) terraformutils.Resource {
			return g.newBeyondcorpResource(id, resourceType, attrs)
		})...)
}
