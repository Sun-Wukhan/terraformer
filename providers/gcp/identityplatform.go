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
	"log"
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"google.golang.org/api/compute/v1"
	identitytoolkit "google.golang.org/api/identitytoolkit/v2"
)

var identityPlatformAllowEmptyValues = []string{""}

var identityPlatformAdditionalFields = map[string]interface{}{}

type IdentityPlatformGenerator struct {
	GCPService
}

// InitResources enumerates Identity Platform for the configured
// project: the project config singleton (provider ID projects/{p}),
// the default supported, inbound SAML and OAuth IdP configs, the
// tenants, and each tenant's own three IdP config collections.
// Provider IDs are the full resource names returned by the API.
func (g *IdentityPlatformGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := identitytoolkit.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, svc)
}

func (g *IdentityPlatformGenerator) initResources(ctx context.Context, svc *identitytoolkit.Service) error {
	project := g.GetArgs()["project"].(string)
	_ = g.GetArgs()["region"].(compute.Region)
	parent := "projects/" + project
	if _, err := svc.Projects.GetConfig(parent).Context(ctx).Do(); err == nil {
		g.appendResource(parent, "google_identity_platform_config")
	} else {
		log.Println(err)
	}
	g.listProjectConfigs(ctx, svc, parent)
	return svc.Projects.Tenants.List(parent).Pages(ctx, func(page *identitytoolkit.GoogleCloudIdentitytoolkitAdminV2ListTenantsResponse) error {
		for _, tenant := range page.Tenants {
			g.appendResource(tenant.Name, "google_identity_platform_tenant")
			g.listTenantConfigs(ctx, svc, tenant.Name)
		}
		return nil
	})
}

func (g *IdentityPlatformGenerator) listProjectConfigs(ctx context.Context, svc *identitytoolkit.Service, parent string) {
	if err := svc.Projects.DefaultSupportedIdpConfigs.List(parent).Pages(ctx, func(page *identitytoolkit.GoogleCloudIdentitytoolkitAdminV2ListDefaultSupportedIdpConfigsResponse) error {
		for _, item := range page.DefaultSupportedIdpConfigs {
			g.appendResource(item.Name, "google_identity_platform_default_supported_idp_config")
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := svc.Projects.InboundSamlConfigs.List(parent).Pages(ctx, func(page *identitytoolkit.GoogleCloudIdentitytoolkitAdminV2ListInboundSamlConfigsResponse) error {
		for _, item := range page.InboundSamlConfigs {
			g.appendResource(item.Name, "google_identity_platform_inbound_saml_config")
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := svc.Projects.OauthIdpConfigs.List(parent).Pages(ctx, func(page *identitytoolkit.GoogleCloudIdentitytoolkitAdminV2ListOAuthIdpConfigsResponse) error {
		for _, item := range page.OauthIdpConfigs {
			g.appendResource(item.Name, "google_identity_platform_oauth_idp_config")
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
}

func (g *IdentityPlatformGenerator) listTenantConfigs(ctx context.Context, svc *identitytoolkit.Service, tenant string) {
	if err := svc.Projects.Tenants.DefaultSupportedIdpConfigs.List(tenant).Pages(ctx, func(page *identitytoolkit.GoogleCloudIdentitytoolkitAdminV2ListDefaultSupportedIdpConfigsResponse) error {
		for _, item := range page.DefaultSupportedIdpConfigs {
			g.appendResource(item.Name, "google_identity_platform_tenant_default_supported_idp_config")
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := svc.Projects.Tenants.InboundSamlConfigs.List(tenant).Pages(ctx, func(page *identitytoolkit.GoogleCloudIdentitytoolkitAdminV2ListInboundSamlConfigsResponse) error {
		for _, item := range page.InboundSamlConfigs {
			g.appendResource(item.Name, "google_identity_platform_tenant_inbound_saml_config")
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := svc.Projects.Tenants.OauthIdpConfigs.List(tenant).Pages(ctx, func(page *identitytoolkit.GoogleCloudIdentitytoolkitAdminV2ListOAuthIdpConfigsResponse) error {
		for _, item := range page.OauthIdpConfigs {
			g.appendResource(item.Name, "google_identity_platform_tenant_oauth_idp_config")
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
}

func (g *IdentityPlatformGenerator) appendResource(name, resourceType string) {
	parts := strings.Split(name, "/")
	g.Resources = append(g.Resources, terraformutils.NewResource(
		name, parts[len(parts)-1], resourceType, g.ProviderName,
		map[string]string{}, identityPlatformAllowEmptyValues, identityPlatformAdditionalFields))
}
