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
	idsv1 "google.golang.org/api/ids/v1"
	"google.golang.org/api/option"
	recaptchav1 "google.golang.org/api/recaptchaenterprise/v1"
	vpcaccessv1 "google.golang.org/api/vpcaccess/v1"
)

// TestCloudIdsGenerator verifies the Cloud IDS endpoint emitted from
// the canned locations + endpoints lists.
func TestCloudIdsGenerator(t *testing.T) {
	endpoint := "projects/test-project/locations/us-central1/endpoints/ids-1"
	transport := &cannedTransport{routes: []cannedRoute{
		{"/locations/us-central1/endpoints", `{"endpoints":[{"name":"` + endpoint + `"}]}`},
		{"/v1/projects/test-project/locations", `{"locations":[{"name":"projects/test-project/locations/us-central1"}]}`},
	}}
	ctx := context.Background()
	service, err := idsv1.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("ids.NewService: %v", err)
	}
	g := &CloudIdsGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project"})
	if err := g.initResources(ctx, service); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_cloud_ids_endpoint": {endpoint},
	})
}

// TestRecaptchaEnterpriseGenerator verifies the reCAPTCHA Enterprise
// key emitted from the canned keys list.
func TestRecaptchaEnterpriseGenerator(t *testing.T) {
	key := "projects/test-project/keys/6Lc1234567890"
	transport := &cannedTransport{routes: []cannedRoute{
		{"/keys", `{"keys":[{"name":"` + key + `"}]}`},
	}}
	ctx := context.Background()
	service, err := recaptchav1.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("recaptchaenterprise.NewService: %v", err)
	}
	g := &RecaptchaEnterpriseGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project"})
	if err := g.initResources(ctx, service); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_recaptcha_enterprise_key": {key},
	})
}

// TestVpcAccessGenerator verifies the Serverless VPC Access connector
// emitted from the canned connectors list at the configured region.
func TestVpcAccessGenerator(t *testing.T) {
	connector := "projects/test-project/locations/us-central1/connectors/conn-1"
	transport := &cannedTransport{routes: []cannedRoute{
		{"/locations/us-central1/connectors", `{"connectors":[{"name":"` + connector + `"}]}`},
	}}
	ctx := context.Background()
	service, err := vpcaccessv1.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("vpcaccess.NewService: %v", err)
	}
	g := &VpcAccessGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	if err := g.initResources(ctx, service); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_vpc_access_connector": {connector},
	})
}
