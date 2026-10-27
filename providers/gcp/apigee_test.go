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
	"reflect"
	"sort"
	"testing"

	"google.golang.org/api/apigee/v1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

// TestApigeeGenerator feeds canned Apigee API responses through the real
// enumeration code and verifies the emitted resource types and IDs,
// including the org-scoped and environment-scoped composite ID forms.
func TestApigeeGenerator(t *testing.T) {
	transport := &cannedTransport{routes: []cannedRoute{
		{"/controlPlaneAccess", `{"synchronizerGroups":[{"name":"sync-group"}]}`},
		{":getSyncAuthorization", `{"identities":["serviceAccount:sa@test-project.iam.gserviceaccount.com"]}`},
		{"/addonsConfig", `{"apiSecurityConfig":{"enabled":true}}`},
		{"/debugmask", `{"requestJSONPaths":["/password"]}`},
		{":getIamPolicy", `{"bindings":[{"role":"roles/apigee.environmentAdmin","members":["user:admin@example.com","group:devs@example.com"]}]}`},
		{"/flowhooks/PreProxyFlowHook", `{"name":"PreProxyFlowHook","sharedFlow":"sf-1"}`},
		{"/securityActions", `{"securityActions":[{"name":"organizations/test-project/environments/dev/securityActions/block-bots"}]}`},
		{"developers/dev@example.com/apps", `{"app":[{"name":"app-1"}]}`},
		{"/developers", `{"developer":[{"email":"dev@example.com"}]}`},
		{"/sharedflows/sf-1/deployments", `{"deployments":[{"environment":"dev","revision":"2"}]}`},
		{"/deployments", `{"deployments":[{"apiProxy":"hello-proxy","environment":"dev","revision":"3"}]}`},
		{"/apis", `{"proxies":[{"name":"hello-proxy"}]}`},
		{"/apiproducts", `{"apiProduct":[{"name":"product-1"}]}`},
		{"/appgroups", `{"appGroups":[{"name":"group-1"}]}`},
		{"/datacollectors", `{"dataCollectors":[{"name":"dc-1"}]}`},
		{"/analytics/datastores", `{"datastores":[{"displayName":"ds-1","self":"https://apigee.googleapis.com/v1/organizations/test-project/analytics/datastores/9d2c-uuid"}]}`},
		{"/dnsZones", `{"dnsZones":[{"name":"zone-1"}]}`},
		{"/endpointAttachments", `{"endpointAttachments":[{"name":"ea-1"}]}`},
		{"/envgroups/eg-1/attachments", `{"environmentGroupAttachments":[{"name":"organizations/test-project/envgroups/eg-1/attachments/att-1"}]}`},
		{"/envgroups", `{"environmentGroups":[{"name":"eg-1"}]}`},
		{"/instances/inst-1/attachments", `{"attachments":[{"name":"organizations/test-project/instances/inst-1/attachments/att-1"}]}`},
		{"/instances/inst-1/natAddresses", `{"natAddresses":[{"name":"nat-1","ipAddress":"1.2.3.4"}]}`},
		{"/instances", `{"instances":[{"name":"inst-1"}]}`},
		{"/securityMonitoringConditions", `{"securityMonitoringConditions":[{"name":"smc-1"}]}`},
		{"/securityProfiles", `{"securityProfiles":[{"name":"organizations/test-project/securityProfiles/sp-1"}]}`},
		{"/sharedflows", `{"sharedFlows":[{"name":"sf-1"}]}`},
		{"/spaces", `{"spaces":[{"name":"space-1"}]}`},
		{"/v1/organizations/test-project", `{"name":"test-project","projectId":"test-project","environments":["dev"],"addonsConfig":{"monetizationConfig":{"enabled":true}}}`},
	}}

	ctx := context.Background()
	service, err := apigee.NewService(ctx,
		option.WithHTTPClient(transport.client()),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatalf("apigee.NewService: %v", err)
	}

	g := &ApigeeGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	if err := g.initResources(ctx, service); err != nil {
		t.Fatalf("initResources: %v", err)
	}

	org := "organizations/test-project"
	env := org + "/environments/dev"
	expected := map[string][]string{
		"google_apigee_organization":                        {org},
		"google_apigee_addons_config":                       {org},
		"google_apigee_control_plane_access":                {org},
		"google_apigee_sync_authorization":                  {org},
		"google_apigee_environment":                         {env},
		"google_apigee_environment_addons_config":           {env},
		"google_apigee_environment_debugmask":               {env},
		"google_apigee_environment_iam_policy":              {env},
		"google_apigee_environment_iam_binding":             {env},
		"google_apigee_environment_iam_member":              {env, env},
		"google_apigee_flowhook":                            {env + "/flowhooks/PreProxyFlowHook"},
		"google_apigee_security_action":                     {env + "/securityActions/block-bots"},
		"google_apigee_api":                                 {org + "/apis/hello-proxy"},
		"google_apigee_api_deployment":                      {org + "/environments/dev/apis/hello-proxy/revisions/3/deployments"},
		"google_apigee_environment_api_revision_deployment": {org + "/environments/dev/apis/hello-proxy/revisions/3"},
		"google_apigee_api_product":                         {org + "/apiproducts/product-1"},
		"google_apigee_app_group":                           {org + "/appgroups/group-1"},
		"google_apigee_data_collector":                      {org + "/datacollectors/dc-1"},
		"google_apigee_datastore":                           {org + "/analytics/datastores/9d2c-uuid"},
		"google_apigee_developer":                           {org + "/developers/dev@example.com"},
		"google_apigee_developer_app":                       {org + "/developers/dev@example.com/apps/app-1"},
		"google_apigee_dns_zone":                            {org + "/dnsZones/zone-1"},
		"google_apigee_endpoint_attachment":                 {org + "/endpointAttachments/ea-1"},
		"google_apigee_envgroup":                            {org + "/envgroups/eg-1"},
		"google_apigee_envgroup_attachment":                 {org + "/envgroups/eg-1/attachments/att-1"},
		"google_apigee_instance":                            {org + "/instances/inst-1"},
		"google_apigee_instance_attachment":                 {org + "/instances/inst-1/attachments/att-1"},
		"google_apigee_nat_address":                         {org + "/instances/inst-1/natAddresses/nat-1"},
		"google_apigee_security_monitoring_condition":       {org + "/securityMonitoringConditions/smc-1"},
		"google_apigee_security_profile_v2":                 {org + "/securityProfiles/sp-1"},
		"google_apigee_sharedflow":                          {org + "/sharedflows/sf-1"},
		"google_apigee_sharedflow_deployment":               {org + "/environments/dev/sharedflows/sf-1/revisions/2/deployments"},
		"google_apigee_space":                               {org + "/spaces/space-1"},
	}

	got := map[string][]string{}
	total := 0
	for _, resource := range g.Resources {
		got[resource.InstanceInfo.Type] = append(got[resource.InstanceInfo.Type], resource.InstanceState.ID)
		total++
	}
	expectedTotal := 0
	for _, ids := range expected {
		expectedTotal += len(ids)
	}
	if total != expectedTotal {
		t.Errorf("resource count = %d, want %d; got %v", total, expectedTotal, got)
	}
	for resourceType, ids := range expected {
		gotIDs := append([]string{}, got[resourceType]...)
		wantIDs := append([]string{}, ids...)
		sort.Strings(gotIDs)
		sort.Strings(wantIDs)
		if !reflect.DeepEqual(gotIDs, wantIDs) {
			t.Errorf("%s IDs = %v, want %v", resourceType, gotIDs, wantIDs)
		}
	}
}
