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
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"google.golang.org/api/compute/v1"
	dlp "google.golang.org/api/dlp/v2"
)

var dlpAllowEmptyValues = []string{""}

var dlpAdditionalFields = map[string]interface{}{}

type DlpGenerator struct {
	GCPService
}

// InitResources enumerates Cloud Data Loss Prevention in the
// configured project and region: deidentify templates, discovery
// configs, inspect templates, job triggers and stored info types.
// Provider IDs are the full resource names.
//
// Deliberately not generated:
//   - google_data_loss_prevention_content_policy: the DLP API
//     discovery document exposes no contentPolicy collection under
//     projects or organizations (no list or get surface in the
//     generated client), so content policies cannot be discovered.
func (g *DlpGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := dlp.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, svc)
}

func (g *DlpGenerator) initResources(ctx context.Context, svc *dlp.Service) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + region
	if err := svc.Projects.Locations.DeidentifyTemplates.List(parent).Pages(ctx, func(page *dlp.GooglePrivacyDlpV2ListDeidentifyTemplatesResponse) error {
		for _, item := range page.DeidentifyTemplates {
			g.appendResource(item.Name, "google_data_loss_prevention_deidentify_template")
		}
		return nil
	}); err != nil {
		return err
	}
	if err := svc.Projects.Locations.DiscoveryConfigs.List(parent).Pages(ctx, func(page *dlp.GooglePrivacyDlpV2ListDiscoveryConfigsResponse) error {
		for _, item := range page.DiscoveryConfigs {
			g.appendResource(item.Name, "google_data_loss_prevention_discovery_config")
		}
		return nil
	}); err != nil {
		return err
	}
	if err := svc.Projects.Locations.InspectTemplates.List(parent).Pages(ctx, func(page *dlp.GooglePrivacyDlpV2ListInspectTemplatesResponse) error {
		for _, item := range page.InspectTemplates {
			g.appendResource(item.Name, "google_data_loss_prevention_inspect_template")
		}
		return nil
	}); err != nil {
		return err
	}
	if err := svc.Projects.Locations.JobTriggers.List(parent).Pages(ctx, func(page *dlp.GooglePrivacyDlpV2ListJobTriggersResponse) error {
		for _, item := range page.JobTriggers {
			g.appendResource(item.Name, "google_data_loss_prevention_job_trigger")
		}
		return nil
	}); err != nil {
		return err
	}
	return svc.Projects.Locations.StoredInfoTypes.List(parent).Pages(ctx, func(page *dlp.GooglePrivacyDlpV2ListStoredInfoTypesResponse) error {
		for _, item := range page.StoredInfoTypes {
			g.appendResource(item.Name, "google_data_loss_prevention_stored_info_type")
		}
		return nil
	})
}

func (g *DlpGenerator) appendResource(name, resourceType string) {
	parts := strings.Split(name, "/")
	g.Resources = append(g.Resources, terraformutils.NewResource(
		name, parts[len(parts)-1], resourceType, g.ProviderName,
		map[string]string{}, dlpAllowEmptyValues, dlpAdditionalFields))
}
