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
	certmanagerv1 "google.golang.org/api/certificatemanager/v1"
)

var certificateManagerAllowEmptyValues = []string{"tags."}

var certificateManagerAdditionalFields = map[string]interface{}{}

type CertificateManagerGenerator struct {
	GCPService
}

// InitResources enumerates Certificate Manager resources in every
// location the API reports for the project: certificates, certificate
// issuance configs, certificate maps with their entries, DNS
// authorizations and trust configs.
func (g *CertificateManagerGenerator) InitResources() error {
	ctx := context.Background()
	service, err := certmanagerv1.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, service)
}

func (g *CertificateManagerGenerator) initResources(ctx context.Context, service *certmanagerv1.Service) error {
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
		g.loadCertificateManagerLocation(ctx, service, locationName, project)
	}
	return nil
}

func (g *CertificateManagerGenerator) newCertificateManagerResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
	parts := strings.Split(id, "/")
	return terraformutils.NewResource(
		id,
		parts[len(parts)-1],
		resourceType,
		g.ProviderName,
		attrs,
		certificateManagerAllowEmptyValues,
		certificateManagerAdditionalFields,
	)
}

func (g *CertificateManagerGenerator) loadCertificateManagerLocation(ctx context.Context, service *certmanagerv1.Service, locationName, project string) {
	parts := strings.Split(locationName, "/")
	location := parts[len(parts)-1]
	attrs := map[string]string{"project": project, "location": location}

	err := service.Projects.Locations.Certificates.List(locationName).Pages(ctx, func(page *certmanagerv1.ListCertificatesResponse) error {
		for _, certificate := range page.Certificates {
			g.Resources = append(g.Resources, g.newCertificateManagerResource(certificate.Name, "google_certificate_manager_certificate", attrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = service.Projects.Locations.CertificateIssuanceConfigs.List(locationName).Pages(ctx, func(page *certmanagerv1.ListCertificateIssuanceConfigsResponse) error {
		for _, config := range page.CertificateIssuanceConfigs {
			g.Resources = append(g.Resources, g.newCertificateManagerResource(config.Name, "google_certificate_manager_certificate_issuance_config", attrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = service.Projects.Locations.CertificateMaps.List(locationName).Pages(ctx, func(page *certmanagerv1.ListCertificateMapsResponse) error {
		for _, certificateMap := range page.CertificateMaps {
			g.Resources = append(g.Resources, g.newCertificateManagerResource(certificateMap.Name, "google_certificate_manager_certificate_map", attrs))
			entryErr := service.Projects.Locations.CertificateMaps.CertificateMapEntries.List(certificateMap.Name).Pages(ctx, func(page *certmanagerv1.ListCertificateMapEntriesResponse) error {
				for _, entry := range page.CertificateMapEntries {
					g.Resources = append(g.Resources, g.newCertificateManagerResource(entry.Name, "google_certificate_manager_certificate_map_entry", attrs))
				}
				return nil
			})
			if entryErr != nil {
				log.Println(entryErr)
			}
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = service.Projects.Locations.DnsAuthorizations.List(locationName).Pages(ctx, func(page *certmanagerv1.ListDnsAuthorizationsResponse) error {
		for _, authorization := range page.DnsAuthorizations {
			g.Resources = append(g.Resources, g.newCertificateManagerResource(authorization.Name, "google_certificate_manager_dns_authorization", attrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = service.Projects.Locations.TrustConfigs.List(locationName).Pages(ctx, func(page *certmanagerv1.ListTrustConfigsResponse) error {
		for _, config := range page.TrustConfigs {
			g.Resources = append(g.Resources, g.newCertificateManagerResource(config.Name, "google_certificate_manager_trust_config", attrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}
