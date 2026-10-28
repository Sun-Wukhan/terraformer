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
)

func TestComputeDepthGenerator(t *testing.T) {
	ctx := context.Background()
	transport := &cannedTransport{routes: []cannedRoute{
		// Generated-client collections.
		{substr: "/global/interconnects", body: `{"items":[{"name":"ic1"}]}`},
		{substr: "/global/targetGrpcProxies", body: `{"items":[{"name":"grpc1"}]}`},
		{substr: "/global/publicAdvertisedPrefixes", body: `{"items":[{"name":"pap1"}]}`},
		{substr: "/regions/us-central1/vpnGateways", body: `{"items":[{"name":"ha1"}]}`},
		{substr: "/regions/us-central1/commitments", body: `{"items":[{"name":"c1"}]}`},
		{substr: "/regions/us-central1/instanceTemplates", body: `{"items":[{"name":"rit1"}]}`},
		{substr: "/regions/us-central1/sslPolicies", body: `{"items":[{"name":"rsp1"}]}`},
		{substr: "/regions/us-central1/targetTcpProxies", body: `{"items":[{"name":"ttp1"}]}`},
		{substr: "/regions/us-central1/serviceAttachments", body: `{"items":[{"name":"sa1"}]}`},
		{substr: "/regions/us-central1/networkAttachments", body: `{"items":[{"name":"na1"}]}`},
		{substr: "/regions/us-central1/publicDelegatedPrefixes", body: `{"items":[{"name":"pdp1"}]}`},
		{substr: "/aggregated/networkEdgeSecurityServices", body: `{"items":{"regions/us-central1":{"networkEdgeSecurityServices":[{"name":"ness1"}]},"regions/europe-west1":{"networkEdgeSecurityServices":[{"name":"ness-eu"}]}}}`},
		{substr: "/regions/us-central1/networkEndpointGroups", body: `{"items":[{"name":"rneg1"}]}`},
		{substr: "/regions/us-central1/instantSnapshots", body: `{"items":[{"name":"ris1"}]}`},
		{substr: "/zones/us-central1-a/storagePools", body: `{"items":[{"name":"sp1"}]}`},
		{substr: "/zones/us-central1-a/instantSnapshots", body: `{"items":[{"name":"is1"}]}`},
		// Raw REST collections (no generated client service).
		{substr: "/global/interconnectGroups", body: `{"items":[{"name":"ig1"}]}`},
		{substr: "/global/interconnectAttachmentGroups", body: `{"items":[{"name":"iag1"}]}`},
		{substr: "/global/vmExtensionPolicies", body: `{"items":[{"name":"vme1"}]}`},
		{substr: "/global/rolloutPlans", body: `{"items":[{"name":"rp1"}]}`},
		{substr: "/global/previewFeatures", body: `{"items":[{"name":"projects/test-project/global/previewFeatures/pf-on","activationStatus":"ENABLED"},{"name":"projects/test-project/global/previewFeatures/pf-off","activationStatus":"ACTIVATION_STATE_UNSPECIFIED"}]}`},
		{substr: "/crossSiteNetworks/csn1/wireGroups", body: `{"items":[{"name":"wg1"}]}`},
		{substr: "/global/crossSiteNetworks", body: `{"items":[{"name":"csn1"}]}`},
		{substr: "/regions/us-central1/backendBuckets", body: `{"items":[{"name":"rbb1"}]}`},
		{substr: "/regions/us-central1/healthAggregationPolicies", body: `{"items":[{"name":"hap1"}]}`},
		{substr: "/regions/us-central1/healthSources", body: `{"items":[{"name":"hs1"}]}`},
		{substr: "/regions/us-central1/compositeHealthChecks", body: `{"items":[{"name":"chc1"}]}`},
		{substr: "/zones/us-central1-a/futureReservations", body: `{"items":[{"name":"fr1"}]}`},
		{substr: "/zones/us-central1-a/vmExtensionPolicies", body: `{"items":[{"name":"zvme1"}]}`},
	}}
	service := computeTestService(t, transport)

	g := &ComputeDepthGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region": compute.Region{
			Name:  "us-central1",
			Zones: []string{"https://www.googleapis.com/compute/v1/projects/test-project/zones/us-central1-a"},
		},
	})
	if err := g.initResources(ctx, service, transport.client()); err != nil {
		t.Fatalf("initResources: %v", err)
	}

	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_compute_interconnect":                     {"ic1"},
		"google_compute_target_grpc_proxy":                {"grpc1"},
		"google_compute_public_advertised_prefix":         {"projects/test-project/global/publicAdvertisedPrefixes/pap1"},
		"google_compute_ha_vpn_gateway":                   {"ha1"},
		"google_compute_region_commitment":                {"c1"},
		"google_compute_region_instance_template":         {"rit1"},
		"google_compute_region_ssl_policy":                {"rsp1"},
		"google_compute_region_target_tcp_proxy":          {"ttp1"},
		"google_compute_service_attachment":               {"sa1"},
		"google_compute_network_attachment":               {"na1"},
		"google_compute_public_delegated_prefix":          {"pdp1"},
		"google_compute_network_edge_security_service":    {"ness1"},
		"google_compute_region_network_endpoint_group":    {"rneg1"},
		"google_compute_region_instant_snapshot":          {"ris1"},
		"google_compute_storage_pool":                     {"us-central1-a/sp1"},
		"google_compute_instant_snapshot":                 {"us-central1-a/is1"},
		"google_compute_interconnect_group":               {"ig1"},
		"google_compute_interconnect_attachment_group":    {"iag1"},
		"google_compute_global_vm_extension_policy":       {"vme1"},
		"google_compute_rollout_plan":                     {"rp1"},
		"google_compute_preview_feature":                  {"pf-on"},
		"google_compute_cross_site_network":               {"csn1"},
		"google_compute_wire_group":                       {"csn1/wg1"},
		"google_compute_region_backend_bucket":            {"rbb1"},
		"google_compute_region_health_aggregation_policy": {"hap1"},
		"google_compute_region_health_source":             {"hs1"},
		"google_compute_region_composite_health_check":    {"chc1"},
		"google_compute_future_reservation":               {"us-central1-a/fr1"},
		"google_compute_zone_vm_extension_policy":         {"us-central1-a/zvme1"},
	})
}

// TestSslCertificatesTypeDiscrimination pins the Batch 5 correction to
// sslCertificates_gen.go: the collection holds both managed and
// self-managed certificates and each must be emitted under the
// provider resource matching its type field.
func TestSslCertificatesTypeDiscrimination(t *testing.T) {
	ctx := context.Background()
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: "/global/sslCertificates", body: `{"items":[{"name":"cert-self","type":"SELF_MANAGED"},{"name":"cert-managed","type":"MANAGED"}]}`},
	}}
	service := computeTestService(t, transport)

	g := &SslCertificatesGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	resources := g.createResources(ctx, service.SslCertificates.List("test-project"))
	assertResourceIDs(t, resources, map[string][]string{
		"google_compute_ssl_certificate":         {"cert-self"},
		"google_compute_managed_ssl_certificate": {"cert-managed"},
	})
}
