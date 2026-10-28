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

	beyondcorpv1 "google.golang.org/api/beyondcorp/v1"
	"google.golang.org/api/option"
)

// TestBeyondcorpGenerator feeds canned BeyondCorp responses through
// the generator and verifies the emitted types and IDs: the security
// gateway and its IAM trio, and the gateway's application with its
// IAM trio.
func TestBeyondcorpGenerator(t *testing.T) {
	gateway := "projects/test-project/locations/global/securityGateways/gw-1"
	application := gateway + "/applications/app-1"
	transport := &cannedTransport{routes: []cannedRoute{
		{":getIamPolicy", `{"bindings":[{"role":"roles/beyondcorp.securityGatewayViewer","members":["user:a@example.com"]}]}`},
		{"/securityGateways/gw-1/applications", `{"applications":[{"name":"` + application + `"}]}`},
		{"/locations/global/securityGateways", `{"securityGateways":[{"name":"` + gateway + `"}]}`},
		{"/v1/projects/test-project/locations", `{"locations":[{"name":"projects/test-project/locations/global"}]}`},
	}}

	ctx := context.Background()
	service, err := beyondcorpv1.NewService(ctx,
		option.WithHTTPClient(transport.client()),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatalf("beyondcorp.NewService: %v", err)
	}

	g := &BeyondcorpGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project"})
	if err := g.initResources(ctx, service); err != nil {
		t.Fatalf("initResources: %v", err)
	}

	expected := map[string][]string{
		"google_beyondcorp_security_gateway":                         {gateway},
		"google_beyondcorp_security_gateway_iam_policy":              {gateway},
		"google_beyondcorp_security_gateway_iam_binding":             {gateway},
		"google_beyondcorp_security_gateway_iam_member":              {gateway},
		"google_beyondcorp_security_gateway_application":             {application},
		"google_beyondcorp_security_gateway_application_iam_policy":  {application},
		"google_beyondcorp_security_gateway_application_iam_binding": {application},
		"google_beyondcorp_security_gateway_application_iam_member":  {application},
	}
	assertResourceIDs(t, g.Resources, expected)
}
