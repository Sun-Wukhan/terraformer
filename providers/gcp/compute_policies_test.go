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

	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

// crmTestService builds a Cloud Resource Manager client on the canned
// transport, for the Batch 5 generators that resolve ancestry.
func crmTestService(t *testing.T, transport *cannedTransport) *cloudresourcemanager.Service {
	t.Helper()
	service, err := cloudresourcemanager.NewService(context.Background(),
		option.WithHTTPClient(transport.client()),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatalf("cloudresourcemanager.NewService: %v", err)
	}
	return service
}

const testAncestryBody = `{"ancestor":[{"resourceId":{"type":"project","id":"123456789"}},{"resourceId":{"type":"folder","id":"456"}},{"resourceId":{"type":"organization","id":"789"}}]}`

func TestComputePoliciesGenerator(t *testing.T) {
	ctx := context.Background()
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: ":getAncestry", body: testAncestryBody},
		// Raw REST organization security policies (locations-based);
		// the associations route must precede the plain list route.
		{substr: "/locations/global/securityPolicies/listAssociations", body: `{"associations":[{"name":"oassoc1","attachmentTarget":"projects/test-project"}]}`},
		{substr: "/locations/global/securityPolicies", body: `{"items":[{"name":"locations/global/securityPolicies/222","rules":[{"priority":7}]}]}`},
		// Hierarchy firewall policies (client list is locations-based).
		{substr: "/locations/global/firewallPolicies", body: `{"items":[{"name":"locations/global/firewallPolicies/111","shortName":"hier1","rules":[{"priority":10}],"associations":[{"name":"hassoc1","attachmentTarget":"folders/456"}]}]}`},
		// Generated-client collections.
		{substr: "/regions/us-central1/firewallPolicies", body: `{"items":[{"name":"rnfw1","rules":[{"priority":500}],"associations":[{"name":"rassoc1","attachmentTarget":"projects/test-project/regions/us-central1/networks/net1"}]}]}`},
		{substr: "/global/firewallPolicies", body: `{"items":[{"name":"nfw1","rules":[{"priority":1000}],"packetMirroringRules":[{"priority":2000}],"associations":[{"name":"assoc1","attachmentTarget":"projects/test-project/global/networks/net1"}]}]}`},
		{substr: "/regions/us-central1/securityPolicies", body: `{"items":[{"name":"rspol1","rules":[{"priority":100}]}]}`},
		{substr: "/global/securityPolicies", body: `{"items":[{"name":"gspol1","rules":[{"priority":42}]}]}`},
		// Beta region network policies.
		{substr: "/regions/us-central1/networkPolicies", body: `{"items":[{"name":"rnp1","trafficClassificationRules":[{"priority":100}]}]}`},
	}}
	computeService := computeTestService(t, transport)
	crmService := crmTestService(t, transport)

	g := &ComputePoliciesGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	if err := g.initResources(ctx, computeService, crmService, transport.client()); err != nil {
		t.Fatalf("initResources: %v", err)
	}

	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_compute_network_firewall_policy":                           {"nfw1"},
		"google_compute_network_firewall_policy_rule":                      {"projects/test-project/global/firewallPolicies/nfw1/rules/1000"},
		"google_compute_network_firewall_policy_packet_mirroring_rule":     {"projects/test-project/global/firewallPolicies/nfw1/packetMirroringRules/2000"},
		"google_compute_network_firewall_policy_association":               {"projects/test-project/global/firewallPolicies/nfw1/associations/assoc1"},
		"google_compute_region_network_firewall_policy":                    {"rnfw1"},
		"google_compute_region_network_firewall_policy_rule":               {"projects/test-project/regions/us-central1/firewallPolicies/rnfw1/500"},
		"google_compute_region_network_firewall_policy_association":        {"projects/test-project/regions/us-central1/firewallPolicies/rnfw1/associations/rassoc1"},
		"google_compute_region_security_policy":                            {"rspol1"},
		"google_compute_region_security_policy_rule":                       {"projects/test-project/regions/us-central1/securityPolicies/rspol1/priority/100"},
		"google_compute_security_policy_rule":                              {"projects/test-project/global/securityPolicies/gspol1/priority/42"},
		"google_compute_firewall_policy":                                   {"locations/global/firewallPolicies/111"},
		"google_compute_firewall_policy_rule":                              {"locations/global/firewallPolicies/111/rules/10"},
		"google_compute_firewall_policy_association":                       {"locations/global/firewallPolicies/111/associations/hassoc1"},
		"google_compute_organization_security_policy":                      {"locations/global/securityPolicies/222"},
		"google_compute_organization_security_policy_rule":                 {"locations/global/securityPolicies/222/priority/7"},
		"google_compute_organization_security_policy_association":          {"locations/global/securityPolicies/222/association/oassoc1"},
		"google_compute_region_network_policy":                             {"rnp1"},
		"google_compute_region_network_policy_traffic_classification_rule": {"projects/test-project/regions/us-central1/networkPolicies/rnp1/100"},
	})
}
