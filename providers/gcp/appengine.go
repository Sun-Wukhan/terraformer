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
	"strconv"
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	appenginev1 "google.golang.org/api/appengine/v1"
)

var appEngineAllowEmptyValues = []string{""}

var appEngineAdditionalFields = map[string]interface{}{}

// AppEngineGenerator enumerates the App Engine application of the
// configured project (at most one exists per project): the
// application itself and its URL dispatch rules singleton, domain
// mappings, firewall ingress rules, and per service the network
// settings and traffic split singletons plus the service's versions
// — split into standard and flexible by each version's env, the
// same discrimination the provider applies.
type AppEngineGenerator struct {
	GCPService
}

func (g *AppEngineGenerator) InitResources() error {
	ctx := context.Background()
	aeService, err := appenginev1.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, aeService)
}

func (g *AppEngineGenerator) initResources(ctx context.Context, aeService *appenginev1.APIService) error {
	project := g.GetArgs()["project"].(string)
	appName := "apps/" + project
	baseAttrs := map[string]string{"project": project}

	app, err := aeService.Apps.Get(project).Context(ctx).Do()
	if err != nil {
		// A project without an App Engine application has nothing
		// to enumerate in this family.
		log.Println(err)
		return nil
	}
	g.Resources = append(g.Resources, g.newAppEngineResource(project, "google_app_engine_application", baseAttrs))
	if app != nil && len(app.DispatchRules) > 0 {
		g.Resources = append(g.Resources, g.newAppEngineResource(project, "google_app_engine_application_url_dispatch_rules", baseAttrs))
	}

	err = aeService.Apps.DomainMappings.List(project).Pages(ctx, func(page *appenginev1.ListDomainMappingsResponse) error {
		for _, mapping := range page.DomainMappings {
			g.Resources = append(g.Resources, g.newAppEngineResource(mapping.Name, "google_app_engine_domain_mapping", baseAttrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = aeService.Apps.Firewall.IngressRules.List(project).Pages(ctx, func(page *appenginev1.ListIngressRulesResponse) error {
		for _, rule := range page.IngressRules {
			id := appName + "/firewall/ingressRules/" + strconv.FormatInt(rule.Priority, 10)
			g.Resources = append(g.Resources, g.newAppEngineResource(id, "google_app_engine_firewall_rule", baseAttrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	return aeService.Apps.Services.List(project).Pages(ctx, func(page *appenginev1.ListServicesResponse) error {
		for _, service := range page.Services {
			serviceName := appName + "/services/" + service.Id
			if service.NetworkSettings != nil {
				g.Resources = append(g.Resources, g.newAppEngineResource(serviceName, "google_app_engine_service_network_settings", baseAttrs))
			}
			if service.Split != nil && len(service.Split.Allocations) > 0 {
				g.Resources = append(g.Resources, g.newAppEngineResource(serviceName, "google_app_engine_service_split_traffic", baseAttrs))
			}
			if err := g.loadAppEngineVersions(ctx, aeService, project, service.Id, baseAttrs); err != nil {
				return err
			}
		}
		return nil
	})
}

func (g *AppEngineGenerator) newAppEngineResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
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
		appEngineAllowEmptyValues,
		appEngineAdditionalFields,
	)
}

func (g *AppEngineGenerator) loadAppEngineVersions(ctx context.Context, aeService *appenginev1.APIService, project, serviceID string, baseAttrs map[string]string) error {
	serviceName := "apps/" + project + "/services/" + serviceID
	return aeService.Apps.Services.Versions.List(project, serviceID).Pages(ctx, func(page *appenginev1.ListVersionsResponse) error {
		for _, version := range page.Versions {
			id := serviceName + "/versions/" + version.Id
			if version.Env == "flexible" {
				g.Resources = append(g.Resources, g.newAppEngineResource(id, "google_app_engine_flexible_app_version", baseAttrs))
			} else {
				g.Resources = append(g.Resources, g.newAppEngineResource(id, "google_app_engine_standard_app_version", baseAttrs))
			}
		}
		return nil
	})
}
