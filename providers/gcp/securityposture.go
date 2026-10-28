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
	"google.golang.org/api/cloudresourcemanager/v1"
	securityposturev1 "google.golang.org/api/securityposture/v1"
)

var securityPostureAllowEmptyValues = []string{"tags."}

var securityPostureAdditionalFields = map[string]interface{}{}

type SecurityPostureGenerator struct {
	GCPService
}

// InitResources enumerates Security Posture resources of the
// organization the project belongs to (resolved through Cloud
// Resource Manager ancestry): postures and posture deployments at
// the global location.
func (g *SecurityPostureGenerator) InitResources() error {
	ctx := context.Background()
	service, err := securityposturev1.NewService(ctx)
	if err != nil {
		return err
	}
	crmService, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		log.Println(err)
		crmService = nil
	}
	return g.initResources(ctx, service, crmService)
}

func (g *SecurityPostureGenerator) initResources(ctx context.Context, service *securityposturev1.Service, crmService *cloudresourcemanager.Service) error {
	project := g.GetArgs()["project"].(string)
	orgID := orgIDFromAncestry(ctx, crmService, project)
	if orgID == "" {
		log.Println("securityposture: could not resolve organization for project " + project + "; skipping organization-scoped collections")
		return nil
	}
	orgParent := "organizations/" + orgID + "/locations/global"
	attrs := map[string]string{"organization": orgID, "location": "global"}

	err := service.Organizations.Locations.Postures.List(orgParent).Pages(ctx, func(page *securityposturev1.ListPosturesResponse) error {
		for _, posture := range page.Postures {
			g.Resources = append(g.Resources, g.newSecurityPostureResource(posture.Name, "google_securityposture_posture", attrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = service.Organizations.Locations.PostureDeployments.List(orgParent).Pages(ctx, func(page *securityposturev1.ListPostureDeploymentsResponse) error {
		for _, deployment := range page.PostureDeployments {
			g.Resources = append(g.Resources, g.newSecurityPostureResource(deployment.Name, "google_securityposture_posture_deployment", attrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	return nil
}

func (g *SecurityPostureGenerator) newSecurityPostureResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
	parts := strings.Split(id, "/")
	return terraformutils.NewResource(
		id,
		parts[len(parts)-1],
		resourceType,
		g.ProviderName,
		attrs,
		securityPostureAllowEmptyValues,
		securityPostureAdditionalFields,
	)
}
