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

	monitoring "google.golang.org/api/monitoring/v3"
	"google.golang.org/api/option"
)

// TestMonitoringDepthEnumeration asserts services (native + custom),
// SLOs, snoozes, custom-only metric descriptors, dashboards and
// monitored projects are emitted with their provider IDs.
func TestMonitoringDepthEnumeration(t *testing.T) {
	ctx := context.Background()
	proj := "projects/test-project"
	svc := proj + "/services/svc1"
	customSvc := proj + "/services/custom1"
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: "/snoozes", body: `{"snoozes":[{"name":"` + proj + `/snoozes/sn1"}]}`},
		{substr: "/metricDescriptors", body: `{"metricDescriptors":[{"name":"` + proj + `/metricDescriptors/custom.googleapis.com/my_metric","type":"custom.googleapis.com/my_metric"},{"name":"` + proj + `/metricDescriptors/compute.googleapis.com/instance/cpu/utilization","type":"compute.googleapis.com/instance/cpu/utilization"}]}`},
		{substr: "/serviceLevelObjectives", body: `{"serviceLevelObjectives":[{"name":"` + svc + `/serviceLevelObjectives/slo1"}]}`},
		{substr: "/services", body: `{"services":[{"name":"` + svc + `"},{"name":"` + customSvc + `","custom":{}}]}`},
		{substr: "/dashboards", body: `{"dashboards":[{"name":"` + proj + `/dashboards/d1"}]}`},
		{substr: "/metricsScopes", body: `{"monitoredProjects":[{"name":"locations/global/metricsScopes/test-project/projects/other-project"}]}`},
	}}
	service, err := monitoring.NewService(ctx,
		option.WithHTTPClient(transport.client()),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	g := &MonitoringDepthGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project"})
	g.initResources(ctx, service, transport.client())
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_monitoring_snooze":            {proj + "/snoozes/sn1"},
		"google_monitoring_metric_descriptor": {proj + "/metricDescriptors/custom.googleapis.com/my_metric"},
		"google_monitoring_service":           {svc, customSvc},
		"google_monitoring_custom_service":    {customSvc},
		"google_monitoring_slo":               {svc + "/serviceLevelObjectives/slo1"},
		"google_monitoring_dashboard":         {proj + "/dashboards/d1"},
		"google_monitoring_monitored_project": {"locations/global/metricsScopes/test-project/projects/other-project"},
	})
}
