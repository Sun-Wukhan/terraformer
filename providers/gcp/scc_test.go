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
	"sort"
	"testing"

	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
	sccv1 "google.golang.org/api/securitycenter/v1"
)

// TestSccGenerator feeds canned Security Command Center v1, v2 and
// Management API responses (plus Cloud Resource Manager ancestry)
// through the real enumeration code and verifies the emitted resource
// types and IDs across project, folder and organization scopes,
// including the inherited-custom-module filtering and the IAM
// expansion of v1 and v2 sources.
func TestSccGenerator(t *testing.T) {
	org := "organizations/123456789"
	folder := "folders/456"
	proj := "projects/test-project"
	transport := &cannedTransport{routes: []cannedRoute{
		{":getAncestry", `{"ancestor":[{"resourceId":{"type":"project","id":"test-project"}},{"resourceId":{"type":"folder","id":"456"}},{"resourceId":{"type":"organization","id":"123456789"}}]}`},
		{":getIamPolicy", `{"bindings":[{"role":"roles/securitycenter.admin","members":["user:a@example.com"]}]}`},
		// v1 organization scope.
		{"/v1/organizations/123456789/sources", `{"sources":[{"name":"` + org + `/sources/src1"}]}`},
		{"/v1/organizations/123456789/notificationConfigs", `{"notificationConfigs":[{"name":"` + org + `/notificationConfigs/nc-org"}]}`},
		{"/v1/organizations/123456789/muteConfigs", `{"muteConfigs":[{"name":"` + org + `/muteConfigs/mute-org"}]}`},
		{"/v1/organizations/123456789/bigQueryExports", `{"bigQueryExports":[{"name":"` + org + `/bigQueryExports/exp-org"}]}`},
		{"/v1/organizations/123456789/securityHealthAnalyticsSettings/customModules", `{"securityHealthAnalyticsCustomModules":[{"name":"` + org + `/securityHealthAnalyticsSettings/customModules/sha-org"}]}`},
		{"/v1/organizations/123456789/eventThreatDetectionSettings/customModules", `{"eventThreatDetectionCustomModules":[{"name":"` + org + `/eventThreatDetectionSettings/customModules/etd-org"}]}`},
		// v1 folder scope.
		{"/v1/folders/456/notificationConfigs", `{"notificationConfigs":[{"name":"` + folder + `/notificationConfigs/nc-folder"}]}`},
		{"/v1/folders/456/muteConfigs", `{"muteConfigs":[{"name":"` + folder + `/muteConfigs/mute-folder"}]}`},
		{"/v1/folders/456/bigQueryExports", `{"bigQueryExports":[{"name":"` + folder + `/bigQueryExports/exp-folder"}]}`},
		{"/v1/folders/456/securityHealthAnalyticsSettings/customModules", `{"securityHealthAnalyticsCustomModules":[{"name":"` + folder + `/securityHealthAnalyticsSettings/customModules/sha-folder"}]}`},
		// v1 project scope: the second SHA module is inherited from
		// the organization and must not be emitted at project scope.
		{"/v1/projects/test-project/notificationConfigs", `{"notificationConfigs":[{"name":"` + proj + `/notificationConfigs/nc-proj"}]}`},
		{"/v1/projects/test-project/muteConfigs", `{"muteConfigs":[{"name":"` + proj + `/muteConfigs/mute-proj"}]}`},
		{"/v1/projects/test-project/bigQueryExports", `{"bigQueryExports":[{"name":"` + proj + `/bigQueryExports/exp-proj"}]}`},
		{"/v1/projects/test-project/securityHealthAnalyticsSettings/customModules", `{"securityHealthAnalyticsCustomModules":[{"name":"` + proj + `/securityHealthAnalyticsSettings/customModules/sha-proj"},{"name":"` + proj + `/securityHealthAnalyticsSettings/customModules/sha-org","ancestorModule":"` + org + `/securityHealthAnalyticsSettings/customModules/sha-org"}]}`},
		// v2 (raw REST), global location.
		{"/v2/organizations/123456789/sources", `{"sources":[{"name":"` + org + `/sources/v2src"}]}`},
		{"/v2/organizations/123456789/locations/global/muteConfigs", `{"muteConfigs":[{"name":"` + org + `/locations/global/muteConfigs/v2mute-org"}]}`},
		{"/v2/organizations/123456789/locations/global/notificationConfigs", `{"notificationConfigs":[{"name":"` + org + `/locations/global/notificationConfigs/v2nc-org"}]}`},
		{"/v2/organizations/123456789/locations/global/bigQueryExports", `{"bigQueryExports":[{"name":"` + org + `/locations/global/bigQueryExports/v2exp-org"}]}`},
		{"/v2/folders/456/locations/global/muteConfigs", `{"muteConfigs":[{"name":"` + folder + `/locations/global/muteConfigs/v2mute-folder"}]}`},
		{"/v2/folders/456/locations/global/notificationConfigs", `{"notificationConfigs":[{"name":"` + folder + `/locations/global/notificationConfigs/v2nc-folder"}]}`},
		{"/v2/folders/456/locations/global/bigQueryExports", `{"bigQueryExports":[{"name":"` + folder + `/locations/global/bigQueryExports/v2exp-folder"}]}`},
		{"/v2/projects/test-project/locations/global/muteConfigs", `{"muteConfigs":[{"name":"` + proj + `/locations/global/muteConfigs/v2mute-proj"}]}`},
		{"/v2/projects/test-project/locations/global/notificationConfigs", `{"notificationConfigs":[{"name":"` + proj + `/locations/global/notificationConfigs/v2nc-proj"}]}`},
		{"/v2/projects/test-project/locations/global/bigQueryExports", `{"bigQueryExports":[{"name":"` + proj + `/locations/global/bigQueryExports/v2exp-proj"}]}`},
		// Management API (raw REST), global location. The project's
		// second module is inherited and must be filtered out.
		{"/v1/organizations/123456789/locations/global/securityHealthAnalyticsCustomModules", `{"securityHealthAnalyticsCustomModules":[{"name":"` + org + `/locations/global/securityHealthAnalyticsCustomModules/mgmt-sha-org"}]}`},
		{"/v1/organizations/123456789/locations/global/eventThreatDetectionCustomModules", `{"eventThreatDetectionCustomModules":[{"name":"` + org + `/locations/global/eventThreatDetectionCustomModules/mgmt-etd-org"}]}`},
		{"/v1/folders/456/locations/global/securityHealthAnalyticsCustomModules", `{"securityHealthAnalyticsCustomModules":[{"name":"` + folder + `/locations/global/securityHealthAnalyticsCustomModules/mgmt-sha-folder"}]}`},
		{"/v1/projects/test-project/locations/global/securityHealthAnalyticsCustomModules", `{"securityHealthAnalyticsCustomModules":[{"name":"` + proj + `/locations/global/securityHealthAnalyticsCustomModules/mgmt-sha-proj"},{"name":"` + proj + `/locations/global/securityHealthAnalyticsCustomModules/mgmt-sha-org","ancestorModule":"` + org + `/locations/global/securityHealthAnalyticsCustomModules/mgmt-sha-org"}]}`},
	}}

	ctx := context.Background()
	clientOpts := []option.ClientOption{
		option.WithHTTPClient(transport.client()),
		option.WithoutAuthentication(),
	}
	sccService, err := sccv1.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("sccv1.NewService: %v", err)
	}
	crmService, err := cloudresourcemanager.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("cloudresourcemanager.NewService: %v", err)
	}

	g := &SccGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	if err := g.initResources(ctx, sccService, crmService, transport.client()); err != nil {
		t.Fatalf("initResources: %v", err)
	}

	global := "/locations/global"
	expected := map[string][]string{
		"google_scc_source":                      {org + "/sources/src1"},
		"google_scc_source_iam_policy":           {org + "/sources/src1"},
		"google_scc_source_iam_binding":          {org + "/sources/src1"},
		"google_scc_source_iam_member":           {org + "/sources/src1"},
		"google_scc_notification_config":         {org + "/notificationConfigs/nc-org"},
		"google_scc_folder_notification_config":  {folder + "/notificationConfigs/nc-folder"},
		"google_scc_project_notification_config": {proj + "/notificationConfigs/nc-proj"},
		"google_scc_mute_config": {
			org + "/muteConfigs/mute-org",
			folder + "/muteConfigs/mute-folder",
			proj + "/muteConfigs/mute-proj",
		},
		"google_scc_organization_scc_big_query_export":                               {org + "/bigQueryExports/exp-org"},
		"google_scc_folder_scc_big_query_export":                                     {folder + "/bigQueryExports/exp-folder"},
		"google_scc_project_scc_big_query_export":                                    {proj + "/bigQueryExports/exp-proj"},
		"google_scc_organization_custom_module":                                      {org + "/securityHealthAnalyticsSettings/customModules/sha-org"},
		"google_scc_folder_custom_module":                                            {folder + "/securityHealthAnalyticsSettings/customModules/sha-folder"},
		"google_scc_project_custom_module":                                           {proj + "/securityHealthAnalyticsSettings/customModules/sha-proj"},
		"google_scc_event_threat_detection_custom_module":                            {org + "/eventThreatDetectionSettings/customModules/etd-org"},
		"google_scc_v2_organization_source":                                          {org + "/sources/v2src"},
		"google_scc_v2_organization_source_iam_policy":                               {org + "/sources/v2src"},
		"google_scc_v2_organization_source_iam_binding":                              {org + "/sources/v2src"},
		"google_scc_v2_organization_source_iam_member":                               {org + "/sources/v2src"},
		"google_scc_v2_organization_mute_config":                                     {org + global + "/muteConfigs/v2mute-org"},
		"google_scc_v2_folder_mute_config":                                           {folder + global + "/muteConfigs/v2mute-folder"},
		"google_scc_v2_project_mute_config":                                          {proj + global + "/muteConfigs/v2mute-proj"},
		"google_scc_v2_organization_notification_config":                             {org + global + "/notificationConfigs/v2nc-org"},
		"google_scc_v2_folder_notification_config":                                   {folder + global + "/notificationConfigs/v2nc-folder"},
		"google_scc_v2_project_notification_config":                                  {proj + global + "/notificationConfigs/v2nc-proj"},
		"google_scc_v2_organization_scc_big_query_export":                            {org + global + "/bigQueryExports/v2exp-org"},
		"google_scc_v2_folder_scc_big_query_export":                                  {folder + global + "/bigQueryExports/v2exp-folder"},
		"google_scc_v2_project_scc_big_query_export":                                 {proj + global + "/bigQueryExports/v2exp-proj"},
		"google_scc_management_organization_security_health_analytics_custom_module": {org + global + "/securityHealthAnalyticsCustomModules/mgmt-sha-org"},
		"google_scc_management_folder_security_health_analytics_custom_module":       {folder + global + "/securityHealthAnalyticsCustomModules/mgmt-sha-folder"},
		"google_scc_management_project_security_health_analytics_custom_module":      {proj + global + "/securityHealthAnalyticsCustomModules/mgmt-sha-proj"},
		"google_scc_management_organization_event_threat_detection_custom_module":    {org + global + "/eventThreatDetectionCustomModules/mgmt-etd-org"},
	}

	got := map[string]map[string]bool{}
	for _, resource := range g.Resources {
		if got[resource.InstanceInfo.Type] == nil {
			got[resource.InstanceInfo.Type] = map[string]bool{}
		}
		got[resource.InstanceInfo.Type][resource.InstanceState.ID] = true
	}
	if len(got) != len(expected) {
		t.Errorf("resource types = %d, want %d", len(got), len(expected))
	}
	for resourceType, ids := range expected {
		gotIDs := []string{}
		for id := range got[resourceType] {
			gotIDs = append(gotIDs, id)
		}
		sort.Strings(gotIDs)
		wantIDs := append([]string{}, ids...)
		sort.Strings(wantIDs)
		if len(gotIDs) != len(wantIDs) {
			t.Errorf("%s IDs = %v, want %v", resourceType, gotIDs, wantIDs)
			continue
		}
		for i := range wantIDs {
			if gotIDs[i] != wantIDs[i] {
				t.Errorf("%s IDs = %v, want %v", resourceType, gotIDs, wantIDs)
				break
			}
		}
	}
}
