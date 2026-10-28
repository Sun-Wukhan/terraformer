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

	"google.golang.org/api/option"
	privatecav1 "google.golang.org/api/privateca/v1"
)

// TestPrivateCaGenerator feeds canned Certificate Authority Service
// responses through the generator and verifies the emitted types and
// IDs: the CA pool and its IAM trio, the certificate authority and
// certificate inside it, and the certificate template with its IAM
// trio.
func TestPrivateCaGenerator(t *testing.T) {
	pool := "projects/test-project/locations/us-central1/caPools/pool-1"
	template := "projects/test-project/locations/us-central1/certificateTemplates/tmpl-1"
	transport := &cannedTransport{routes: []cannedRoute{
		{"/caPools/pool-1/certificateAuthorities", `{"certificateAuthorities":[{"name":"` + pool + `/certificateAuthorities/ca-1"}]}`},
		{"/caPools/pool-1/certificates", `{"certificates":[{"name":"` + pool + `/certificates/cert-1"}]}`},
		{":getIamPolicy", `{"bindings":[{"role":"roles/privateca.admin","members":["user:a@example.com"]}]}`},
		{"/locations/us-central1/caPools", `{"caPools":[{"name":"` + pool + `"}]}`},
		{"/locations/us-central1/certificateTemplates", `{"certificateTemplates":[{"name":"` + template + `"}]}`},
		{"/v1/projects/test-project/locations", `{"locations":[{"name":"projects/test-project/locations/us-central1"}]}`},
	}}

	ctx := context.Background()
	service, err := privatecav1.NewService(ctx,
		option.WithHTTPClient(transport.client()),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatalf("privateca.NewService: %v", err)
	}

	g := &PrivateCaGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project"})
	if err := g.initResources(ctx, service); err != nil {
		t.Fatalf("initResources: %v", err)
	}

	expected := map[string][]string{
		"google_privateca_ca_pool":                          {pool},
		"google_privateca_ca_pool_iam_policy":               {pool},
		"google_privateca_ca_pool_iam_binding":              {pool},
		"google_privateca_ca_pool_iam_member":               {pool},
		"google_privateca_certificate_authority":            {pool + "/certificateAuthorities/ca-1"},
		"google_privateca_certificate":                      {pool + "/certificates/cert-1"},
		"google_privateca_certificate_template":             {template},
		"google_privateca_certificate_template_iam_policy":  {template},
		"google_privateca_certificate_template_iam_binding": {template},
		"google_privateca_certificate_template_iam_member":  {template},
	}
	assertResourceIDs(t, g.Resources, expected)
}
