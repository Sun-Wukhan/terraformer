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

	certmanagerv1 "google.golang.org/api/certificatemanager/v1"
	"google.golang.org/api/option"
)

// TestCertificateManagerGenerator feeds canned Certificate Manager
// responses through the generator and verifies the emitted types and
// IDs for all six resource types at the reported location.
func TestCertificateManagerGenerator(t *testing.T) {
	loc := "projects/test-project/locations/global/"
	certMap := loc + "certificateMaps/map-1"
	transport := &cannedTransport{routes: []cannedRoute{
		{"/certificateMaps/map-1/certificateMapEntries", `{"certificateMapEntries":[{"name":"` + certMap + `/certificateMapEntries/entry-1"}]}`},
		{"/locations/global/certificates", `{"certificates":[{"name":"` + loc + `certificates/cert-1"}]}`},
		{"/locations/global/certificateIssuanceConfigs", `{"certificateIssuanceConfigs":[{"name":"` + loc + `certificateIssuanceConfigs/cic-1"}]}`},
		{"/locations/global/certificateMaps", `{"certificateMaps":[{"name":"` + certMap + `"}]}`},
		{"/locations/global/dnsAuthorizations", `{"dnsAuthorizations":[{"name":"` + loc + `dnsAuthorizations/dnsauth-1"}]}`},
		{"/locations/global/trustConfigs", `{"trustConfigs":[{"name":"` + loc + `trustConfigs/tc-1"}]}`},
		{"/v1/projects/test-project/locations", `{"locations":[{"name":"projects/test-project/locations/global"}]}`},
	}}

	ctx := context.Background()
	service, err := certmanagerv1.NewService(ctx,
		option.WithHTTPClient(transport.client()),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatalf("certificatemanager.NewService: %v", err)
	}

	g := &CertificateManagerGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project"})
	if err := g.initResources(ctx, service); err != nil {
		t.Fatalf("initResources: %v", err)
	}

	expected := map[string][]string{
		"google_certificate_manager_certificate":                 {loc + "certificates/cert-1"},
		"google_certificate_manager_certificate_issuance_config": {loc + "certificateIssuanceConfigs/cic-1"},
		"google_certificate_manager_certificate_map":             {certMap},
		"google_certificate_manager_certificate_map_entry":       {certMap + "/certificateMapEntries/entry-1"},
		"google_certificate_manager_dns_authorization":           {loc + "dnsAuthorizations/dnsauth-1"},
		"google_certificate_manager_trust_config":                {loc + "trustConfigs/tc-1"},
	}
	assertResourceIDs(t, g.Resources, expected)
}
