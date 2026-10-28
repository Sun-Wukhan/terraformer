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
	privatecav1 "google.golang.org/api/privateca/v1"
)

var privatecaAllowEmptyValues = []string{"tags."}

var privatecaAdditionalFields = map[string]interface{}{}

type PrivateCaGenerator struct {
	GCPService
}

// InitResources enumerates Certificate Authority Service resources in
// every location the API reports for the project: CA pools (with the
// IAM policy, bindings and members expanded from each pool's live IAM
// policy), the certificate authorities and certificates inside each
// pool, and certificate templates (likewise with their IAM trio).
func (g *PrivateCaGenerator) InitResources() error {
	ctx := context.Background()
	service, err := privatecav1.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, service)
}

func (g *PrivateCaGenerator) initResources(ctx context.Context, service *privatecav1.Service) error {
	project := g.GetArgs()["project"].(string)
	projectParent := "projects/" + project

	locations, err := listLocationNames(func(pageToken string) ([]string, string, error) {
		resp, err := service.Projects.Locations.List(projectParent).PageToken(pageToken).Context(ctx).Do()
		if err != nil {
			return nil, "", err
		}
		names := []string{}
		for _, location := range resp.Locations {
			names = append(names, location.Name)
		}
		return names, resp.NextPageToken, nil
	})
	if err != nil {
		return err
	}
	for _, locationName := range locations {
		g.loadPrivateCaLocation(ctx, service, locationName, project)
	}
	return nil
}

// listLocationNames pages a projects.locations list endpoint of one of
// the location-scoped APIs and returns every location resource name.
func listLocationNames(listPage func(pageToken string) ([]string, string, error)) ([]string, error) {
	names := []string{}
	pageToken := ""
	for {
		page, nextPageToken, err := listPage(pageToken)
		if err != nil {
			return nil, err
		}
		names = append(names, page...)
		if nextPageToken == "" {
			return names, nil
		}
		pageToken = nextPageToken
	}
}

func (g *PrivateCaGenerator) newPrivateCaResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
	parts := strings.Split(id, "/")
	return terraformutils.NewResource(
		id,
		parts[len(parts)-1],
		resourceType,
		g.ProviderName,
		attrs,
		privatecaAllowEmptyValues,
		privatecaAdditionalFields,
	)
}

func (g *PrivateCaGenerator) loadPrivateCaLocation(ctx context.Context, service *privatecav1.Service, locationName, project string) {
	parts := strings.Split(locationName, "/")
	location := parts[len(parts)-1]
	attrs := map[string]string{"project": project, "location": location}

	err := service.Projects.Locations.CaPools.List(locationName).Pages(ctx, func(page *privatecav1.ListCaPoolsResponse) error {
		for _, pool := range page.CaPools {
			g.Resources = append(g.Resources, g.newPrivateCaResource(pool.Name, "google_privateca_ca_pool", attrs))
			g.loadCaPoolIam(ctx, service, pool.Name)
			authorityErr := service.Projects.Locations.CaPools.CertificateAuthorities.List(pool.Name).Pages(ctx, func(page *privatecav1.ListCertificateAuthoritiesResponse) error {
				for _, authority := range page.CertificateAuthorities {
					g.Resources = append(g.Resources, g.newPrivateCaResource(authority.Name, "google_privateca_certificate_authority", attrs))
				}
				return nil
			})
			if authorityErr != nil {
				log.Println(authorityErr)
			}
			certificateErr := service.Projects.Locations.CaPools.Certificates.List(pool.Name).Pages(ctx, func(page *privatecav1.ListCertificatesResponse) error {
				for _, certificate := range page.Certificates {
					g.Resources = append(g.Resources, g.newPrivateCaResource(certificate.Name, "google_privateca_certificate", attrs))
				}
				return nil
			})
			if certificateErr != nil {
				log.Println(certificateErr)
			}
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = service.Projects.Locations.CertificateTemplates.List(locationName).Pages(ctx, func(page *privatecav1.ListCertificateTemplatesResponse) error {
		for _, template := range page.CertificateTemplates {
			g.Resources = append(g.Resources, g.newPrivateCaResource(template.Name, "google_privateca_certificate_template", attrs))
			g.loadCertificateTemplateIam(ctx, service, template.Name)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (g *PrivateCaGenerator) loadCaPoolIam(ctx context.Context, service *privatecav1.Service, poolName string) {
	policy, err := service.Projects.Locations.CaPools.GetIamPolicy(poolName).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		return
	}
	bindings := []iamBinding{}
	for _, binding := range policy.Bindings {
		bindings = append(bindings, iamBinding{Role: binding.Role, Members: binding.Members})
	}
	g.Resources = append(g.Resources, expandIamPolicy(bindings, poolName,
		"google_privateca_ca_pool_iam_policy", "google_privateca_ca_pool_iam_binding", "google_privateca_ca_pool_iam_member",
		map[string]string{"ca_pool": poolName},
		func(id, resourceType string, attrs map[string]string) terraformutils.Resource {
			return g.newPrivateCaResource(id, resourceType, attrs)
		})...)
}

func (g *PrivateCaGenerator) loadCertificateTemplateIam(ctx context.Context, service *privatecav1.Service, templateName string) {
	policy, err := service.Projects.Locations.CertificateTemplates.GetIamPolicy(templateName).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		return
	}
	bindings := []iamBinding{}
	for _, binding := range policy.Bindings {
		bindings = append(bindings, iamBinding{Role: binding.Role, Members: binding.Members})
	}
	g.Resources = append(g.Resources, expandIamPolicy(bindings, templateName,
		"google_privateca_certificate_template_iam_policy", "google_privateca_certificate_template_iam_binding", "google_privateca_certificate_template_iam_member",
		map[string]string{"certificate_template": templateName},
		func(id, resourceType string, attrs map[string]string) terraformutils.Resource {
			return g.newPrivateCaResource(id, resourceType, attrs)
		})...)
}
