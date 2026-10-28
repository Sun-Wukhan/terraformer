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
	"net/http"
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	monitoring "google.golang.org/api/monitoring/v3"
)

// Monitoring depth enumeration (ops sweep), a sibling service to the
// original MonitoringGenerator (which uses the GAPIC clients for alert
// policies, groups, notification channels and uptime checks). This
// service covers the remaining Monitoring surface:
//
//   - services and service-level objectives, and custom services
//     (services whose API representation carries a custom definition),
//     via the generated monitoring v3 client;
//   - snoozes, via the generated v3 client;
//   - custom metric descriptors only (type prefix
//     custom.googleapis.com/): system and agent descriptors are
//     Google-managed and not importable configuration;
//   - dashboards and monitored projects over raw REST: dashboards live
//     in the monitoring v1 API, which has no generated client in the
//     pinned module, and monitored projects are listed through the
//     v1 metricsScopes collection of the scoping project.
type MonitoringDepthGenerator struct {
	GCPService
}

const monitoringV1BaseURL = "https://monitoring.googleapis.com/v1/"

func (g *MonitoringDepthGenerator) InitResources() error {
	ctx := context.Background()
	service, err := monitoring.NewService(ctx)
	if err != nil {
		return err
	}
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		log.Println(err)
	}
	g.initResources(ctx, service, restClient)
	return nil
}

func (g *MonitoringDepthGenerator) newResource(id, resourceType string) terraformutils.Resource {
	return terraformutils.NewResource(
		id, opsLastSegment(id), resourceType, g.ProviderName,
		map[string]string{"project": g.GetArgs()["project"].(string)},
		[]string{}, map[string]interface{}{},
	)
}

func (g *MonitoringDepthGenerator) initResources(ctx context.Context, service *monitoring.Service, restClient *http.Client) {
	project := g.GetArgs()["project"].(string)
	parent := "projects/" + project

	if err := service.Projects.Snoozes.List(parent).Pages(ctx, func(page *monitoring.ListSnoozesResponse) error {
		for _, snooze := range page.Snoozes {
			g.Resources = append(g.Resources, g.newResource(snooze.Name, "google_monitoring_snooze"))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}

	if err := service.Projects.MetricDescriptors.List(parent).Pages(ctx, func(page *monitoring.ListMetricDescriptorsResponse) error {
		for _, descriptor := range page.MetricDescriptors {
			if !strings.HasPrefix(descriptor.Type, "custom.googleapis.com/") {
				continue
			}
			g.Resources = append(g.Resources, g.newResource(descriptor.Name, "google_monitoring_metric_descriptor"))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}

	if err := service.Services.List(parent).Pages(ctx, func(page *monitoring.ListServicesResponse) error {
		for _, svc := range page.Services {
			g.Resources = append(g.Resources, g.newResource(svc.Name, "google_monitoring_service"))
			if svc.Custom != nil {
				g.Resources = append(g.Resources, g.newResource(svc.Name, "google_monitoring_custom_service"))
			}
			if err := service.Services.ServiceLevelObjectives.List(svc.Name).Pages(ctx, func(sloPage *monitoring.ListServiceLevelObjectivesResponse) error {
				for _, slo := range sloPage.ServiceLevelObjectives {
					g.Resources = append(g.Resources, g.newResource(slo.Name, "google_monitoring_slo"))
				}
				return nil
			}); err != nil {
				log.Println(err)
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}

	if restClient != nil {
		dashboards, err := restListResources(ctx, restClient, monitoringV1BaseURL+parent+"/dashboards")
		if err != nil {
			log.Println(err)
		}
		for _, dashboard := range dashboards {
			g.Resources = append(g.Resources, g.newResource(dashboard.Name, "google_monitoring_dashboard"))
		}
		monitored, err := restListResources(ctx, restClient, monitoringV1BaseURL+"locations/global/metricsScopes/"+project+"/projects")
		if err != nil {
			log.Println(err)
		}
		for _, mp := range monitored {
			// The provider ID of a monitored project is its full
			// metrics-scope resource name.
			g.Resources = append(g.Resources, g.newResource(mp.Name, "google_monitoring_monitored_project"))
		}
	}
}
