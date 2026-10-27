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
	netsecv1 "google.golang.org/api/networksecurity/v1"
	netsecbeta "google.golang.org/api/networksecurity/v1beta1"
	"google.golang.org/api/option"
)

// TestNetworkSecurityGenerator feeds canned Network Security (v1 and
// v1beta1) and Cloud Resource Manager responses through the real
// enumeration code and verifies the emitted resource types and IDs across
// project regional, project global, project zonal and organization
// scopes.
func TestNetworkSecurityGenerator(t *testing.T) {
	proj := "projects/test-project/locations/"
	org := "organizations/123456789/locations/"
	transport := &cannedTransport{routes: []cannedRoute{
		{":getAncestry", `{"ancestor":[{"resourceId":{"type":"project","id":"test-project"}},{"resourceId":{"type":"organization","id":"123456789"}}]}`},
		{":getIamPolicy", `{"bindings":[{"role":"roles/networksecurity.viewer","members":["user:a@example.com","user:b@example.com"]}]}`},
		// Organization-scoped collections (must precede the generic
		// location routes, whose substrings also appear in org paths).
		{"/organizations/123456789/locations/us-central1/addressGroups", `{"addressGroups":[{"name":"` + org + `us-central1/addressGroups/org-ag-regional"}]}`},
		{"/organizations/123456789/locations/global/addressGroups", `{"addressGroups":[{"name":"` + org + `global/addressGroups/org-ag-global"}]}`},
		{"/organizations/123456789/locations/global/securityProfiles", `{"securityProfiles":[{"name":"` + org + `global/securityProfiles/sp-1"}]}`},
		{"/organizations/123456789/locations/global/securityProfileGroups", `{"securityProfileGroups":[{"name":"` + org + `global/securityProfileGroups/spg-1"}]}`},
		{"/organizations/123456789/locations/us-central1-a/firewallEndpoints", `{"firewallEndpoints":[{"name":"` + org + `us-central1-a/firewallEndpoints/fwe-1"}]}`},
		// Gateway security policy rules (the rules path also contains
		// the policy list substring, so this route comes first).
		{"/rules", `{"gatewaySecurityPolicyRules":[{"name":"` + proj + `us-central1/gatewaySecurityPolicies/gsp-regional/rules/rule-1"}]}`},
		// Project regional collections.
		{"/locations/us-central1/addressGroups", `{"addressGroups":[{"name":"` + proj + `us-central1/addressGroups/ag-regional"}]}`},
		{"/locations/us-central1/authorizationPolicies", `{"authorizationPolicies":[{"name":"` + proj + `us-central1/authorizationPolicies/authp-regional"}]}`},
		{"/locations/us-central1/authzPolicies", `{"authzPolicies":[{"name":"` + proj + `us-central1/authzPolicies/authz-regional"}]}`},
		{"/locations/us-central1/clientTlsPolicies", `{"clientTlsPolicies":[{"name":"` + proj + `us-central1/clientTlsPolicies/ctp-regional"}]}`},
		{"/locations/us-central1/serverTlsPolicies", `{"serverTlsPolicies":[{"name":"` + proj + `us-central1/serverTlsPolicies/stp-regional"}]}`},
		{"/locations/us-central1/gatewaySecurityPolicies", `{"gatewaySecurityPolicies":[{"name":"` + proj + `us-central1/gatewaySecurityPolicies/gsp-regional"}]}`},
		{"/locations/us-central1/tlsInspectionPolicies", `{"tlsInspectionPolicies":[{"name":"` + proj + `us-central1/tlsInspectionPolicies/tip-regional"}]}`},
		{"/locations/us-central1/urlLists", `{"urlLists":[{"name":"` + proj + `us-central1/urlLists/ul-regional"}]}`},
		{"/locations/us-central1/backendAuthenticationConfigs", `{"backendAuthenticationConfigs":[{"name":"` + proj + `us-central1/backendAuthenticationConfigs/bac-regional"}]}`},
		{"/locations/us-central1/mirroringEndpointGroupAssociations", `{"mirroringEndpointGroupAssociations":[{"name":"` + proj + `us-central1/mirroringEndpointGroupAssociations/mega-1"}]}`},
		{"/locations/us-central1/interceptEndpointGroupAssociations", `{"interceptEndpointGroupAssociations":[{"name":"` + proj + `us-central1/interceptEndpointGroupAssociations/iega-1"}]}`},
		// Project global collections.
		{"/locations/global/addressGroups", `{"addressGroups":[{"name":"` + proj + `global/addressGroups/ag-global"}]}`},
		{"/locations/global/authorizationPolicies", `{"authorizationPolicies":[{"name":"` + proj + `global/authorizationPolicies/authp-global"}]}`},
		{"/locations/global/backendAuthenticationConfigs", `{"backendAuthenticationConfigs":[{"name":"` + proj + `global/backendAuthenticationConfigs/bac-global"}]}`},
		{"/locations/global/mirroringDeploymentGroups", `{"mirroringDeploymentGroups":[{"name":"` + proj + `global/mirroringDeploymentGroups/mdg-1"}]}`},
		{"/locations/global/mirroringEndpointGroups", `{"mirroringEndpointGroups":[{"name":"` + proj + `global/mirroringEndpointGroups/meg-1"}]}`},
		{"/locations/global/interceptDeploymentGroups", `{"interceptDeploymentGroups":[{"name":"` + proj + `global/interceptDeploymentGroups/idg-1"}]}`},
		{"/locations/global/interceptEndpointGroups", `{"interceptEndpointGroups":[{"name":"` + proj + `global/interceptEndpointGroups/ieg-1"}]}`},
		// Project zonal collections.
		{"/locations/us-central1-a/firewallEndpointAssociations", `{"firewallEndpointAssociations":[{"name":"` + proj + `us-central1-a/firewallEndpointAssociations/fea-1"}]}`},
		{"/locations/us-central1-a/mirroringDeployments", `{"mirroringDeployments":[{"name":"` + proj + `us-central1-a/mirroringDeployments/md-1"}]}`},
		{"/locations/us-central1-a/interceptDeployments", `{"interceptDeployments":[{"name":"` + proj + `us-central1-a/interceptDeployments/id-1"}]}`},
	}}

	ctx := context.Background()
	clientOpts := []option.ClientOption{
		option.WithHTTPClient(transport.client()),
		option.WithoutAuthentication(),
	}
	v1Service, err := netsecv1.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("netsecv1.NewService: %v", err)
	}
	betaService, err := netsecbeta.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("netsecbeta.NewService: %v", err)
	}
	crmService, err := cloudresourcemanager.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("cloudresourcemanager.NewService: %v", err)
	}

	g := &NetworkSecurityGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region": compute.Region{
			Name:  "us-central1",
			Zones: []string{"https://www.googleapis.com/compute/v1/projects/test-project/zones/us-central1-a"},
		},
	})
	if err := g.initResources(ctx, v1Service, betaService, crmService); err != nil {
		t.Fatalf("initResources: %v", err)
	}

	regional := proj + "us-central1/"
	global := proj + "global/"
	zonal := proj + "us-central1-a/"
	expected := map[string][]string{
		"google_network_security_address_group": {
			regional + "addressGroups/ag-regional",
			global + "addressGroups/ag-global",
			org + "us-central1/addressGroups/org-ag-regional",
			org + "global/addressGroups/org-ag-global",
		},
		"google_network_security_address_group_iam_policy": {
			regional + "addressGroups/ag-regional",
			global + "addressGroups/ag-global",
		},
		"google_network_security_address_group_iam_binding": {
			regional + "addressGroups/ag-regional",
			global + "addressGroups/ag-global",
		},
		"google_network_security_address_group_iam_member": {
			regional + "addressGroups/ag-regional",
			global + "addressGroups/ag-global",
		},
		"google_network_security_authorization_policy": {
			regional + "authorizationPolicies/authp-regional",
			global + "authorizationPolicies/authp-global",
		},
		"google_network_security_authz_policy":                 {regional + "authzPolicies/authz-regional"},
		"google_network_security_client_tls_policy":            {regional + "clientTlsPolicies/ctp-regional"},
		"google_network_security_server_tls_policy":            {regional + "serverTlsPolicies/stp-regional"},
		"google_network_security_gateway_security_policy":      {regional + "gatewaySecurityPolicies/gsp-regional"},
		"google_network_security_gateway_security_policy_rule": {regional + "gatewaySecurityPolicies/gsp-regional/rules/rule-1"},
		"google_network_security_tls_inspection_policy":        {regional + "tlsInspectionPolicies/tip-regional"},
		"google_network_security_url_lists":                    {regional + "urlLists/ul-regional"},
		"google_network_security_backend_authentication_config": {
			regional + "backendAuthenticationConfigs/bac-regional",
			global + "backendAuthenticationConfigs/bac-global",
		},
		"google_network_security_mirroring_deployment_group":           {global + "mirroringDeploymentGroups/mdg-1"},
		"google_network_security_mirroring_endpoint_group":             {global + "mirroringEndpointGroups/meg-1"},
		"google_network_security_intercept_deployment_group":           {global + "interceptDeploymentGroups/idg-1"},
		"google_network_security_intercept_endpoint_group":             {global + "interceptEndpointGroups/ieg-1"},
		"google_network_security_firewall_endpoint_association":        {zonal + "firewallEndpointAssociations/fea-1"},
		"google_network_security_mirroring_deployment":                 {zonal + "mirroringDeployments/md-1"},
		"google_network_security_intercept_deployment":                 {zonal + "interceptDeployments/id-1"},
		"google_network_security_mirroring_endpoint_group_association": {regional + "mirroringEndpointGroupAssociations/mega-1"},
		"google_network_security_intercept_endpoint_group_association": {regional + "interceptEndpointGroupAssociations/iega-1"},
		"google_network_security_security_profile":                     {org + "global/securityProfiles/sp-1"},
		"google_network_security_security_profile_group":               {org + "global/securityProfileGroups/spg-1"},
		"google_network_security_firewall_endpoint":                    {org + "us-central1-a/firewallEndpoints/fwe-1"},
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
