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

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/dns/v1"
	"google.golang.org/api/option"
)

// TestCloudDNSDepthGenerator feeds canned Cloud DNS responses through
// the original-estate pass (zone and record set) and the depth pass,
// verifying the emitted types and IDs: the managed zone IAM trio, the
// DNS server policy, the response policy and its rule, and the
// outbound endpoint listed over raw REST.
func TestCloudDNSDepthGenerator(t *testing.T) {
	zoneID := "projects/test-project/managedZones/zone-1"
	transport := &cannedTransport{routes: []cannedRoute{
		{":getIamPolicy", `{"bindings":[{"role":"roles/dns.admin","members":["user:a@example.com"]}]}`},
		{"/responsePolicies/rp-1/rules", `{"responsePolicyRules":[{"ruleName":"rule-1"}]}`},
		{"/responsePolicies", `{"responsePolicies":[{"responsePolicyName":"rp-1"}]}`},
		{"/policies", `{"policies":[{"name":"pol-1"}]}`},
		{"/outboundEndpoints", `{"outboundEndpoints":[{"name":"projects/test-project/locations/us-central1/outboundEndpoints/oe-1"}]}`},
		{"/rrsets", `{"rrsets":[{"name":"www.example.com.","type":"A"}]}`},
		{"/managedZones", `{"managedZones":[{"name":"zone-1","dnsName":"example.com."}]}`},
	}}

	ctx := context.Background()
	clientOpts := []option.ClientOption{
		option.WithHTTPClient(transport.client()),
		option.WithoutAuthentication(),
	}
	svc, err := dns.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("dns.NewService: %v", err)
	}

	g := &CloudDNSGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	g.Resources = g.createZonesResources(ctx, svc, "test-project")
	g.initDnsDepthResources(ctx, svc, transport.client())

	expected := map[string][]string{
		"google_dns_managed_zone":             {"zone-1"},
		"google_dns_record_set":               {"zone-1/www.example.com./A"},
		"google_dns_managed_zone_iam_policy":  {zoneID},
		"google_dns_managed_zone_iam_binding": {zoneID},
		"google_dns_managed_zone_iam_member":  {zoneID},
		"google_dns_policy":                   {"projects/test-project/policies/pol-1"},
		"google_dns_response_policy":          {"projects/test-project/responsePolicies/rp-1"},
		"google_dns_response_policy_rule":     {"projects/test-project/responsePolicies/rp-1/rules/rule-1"},
		"google_dns_outbound_endpoint":        {"projects/test-project/locations/us-central1/outboundEndpoints/oe-1"},
	}
	assertResourceIDs(t, g.Resources, expected)
}
