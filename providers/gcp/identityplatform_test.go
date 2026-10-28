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
	identitytoolkit "google.golang.org/api/identitytoolkit/v2"
	"google.golang.org/api/option"
)

func TestIdentityPlatformGeneratorEnumeration(t *testing.T) {
	ctx := context.Background()
	project := "projects/test-project"
	tenant := project + "/tenants/tenant1"
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: "/tenants/tenant1/defaultSupportedIdpConfigs", body: `{"defaultSupportedIdpConfigs":[{"name":"` + tenant + `/defaultSupportedIdpConfigs/google.com"}]}`},
		{substr: "/tenants/tenant1/inboundSamlConfigs", body: `{"inboundSamlConfigs":[{"name":"` + tenant + `/inboundSamlConfigs/saml1"}]}`},
		{substr: "/tenants/tenant1/oauthIdpConfigs", body: `{"oauthIdpConfigs":[{"name":"` + tenant + `/oauthIdpConfigs/oauth1"}]}`},
		{substr: "/tenants", body: `{"tenants":[{"name":"` + tenant + `"}]}`},
		{substr: "/defaultSupportedIdpConfigs", body: `{"defaultSupportedIdpConfigs":[{"name":"` + project + `/defaultSupportedIdpConfigs/google.com"}]}`},
		{substr: "/inboundSamlConfigs", body: `{"inboundSamlConfigs":[{"name":"` + project + `/inboundSamlConfigs/saml1"}]}`},
		{substr: "/oauthIdpConfigs", body: `{"oauthIdpConfigs":[{"name":"` + project + `/oauthIdpConfigs/oauth1"}]}`},
		{substr: ":getConfig", body: `{}`},
	}}
	svc, err := identitytoolkit.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	g := &IdentityPlatformGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project", "region": compute.Region{Name: "us-central1"}})
	if err := g.initResources(ctx, svc); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_identity_platform_config":                              {project},
		"google_identity_platform_default_supported_idp_config":        {project + "/defaultSupportedIdpConfigs/google.com"},
		"google_identity_platform_inbound_saml_config":                 {project + "/inboundSamlConfigs/saml1"},
		"google_identity_platform_oauth_idp_config":                    {project + "/oauthIdpConfigs/oauth1"},
		"google_identity_platform_tenant":                              {tenant},
		"google_identity_platform_tenant_default_supported_idp_config": {tenant + "/defaultSupportedIdpConfigs/google.com"},
		"google_identity_platform_tenant_inbound_saml_config":          {tenant + "/inboundSamlConfigs/saml1"},
		"google_identity_platform_tenant_oauth_idp_config":             {tenant + "/oauthIdpConfigs/oauth1"},
	})
}
