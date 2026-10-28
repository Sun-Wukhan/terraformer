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
	serviceusagev1 "google.golang.org/api/serviceusage/v1"
)

// initProjectDepthResources extends the project generator with the
// project's enabled services (google_project_service, whose provider
// ID is "{project}/{service}") and its usage export bucket singleton
// (google_project_usage_export_bucket, whose provider ID is the
// project ID), read from the Compute project resource.
//
// Deliberately not generated:
//   - google_project_default_service_accounts: an action-only helper
//     (deprivilege/delete/restore the default service accounts); it
//     has no underlying API object and the provider documents no
//     import for it.
//   - google_project_iam_member_remove: an action-only helper that
//     removes a member on apply; there is no API object to import.
//   - google_project_service_identity: see iam_org.go — a service
//     identity is provisioned through generateServiceIdentity and has
//     no listable/gettable object.
func (g *ProjectGenerator) InitResources() error {
	g.Resources = append(g.Resources, terraformutils.NewResource(
		g.GetArgs()["project"].(string),
		g.GetArgs()["project"].(string),
		"google_project",
		g.ProviderName,
		map[string]string{
			"auto_create_network": "true",
		},
		projectAllowEmptyValues,
		projectAdditionalFields,
	))
	ctx := context.Background()
	suSvc, err := serviceusagev1.NewService(ctx)
	if err != nil {
		log.Println(err)
		return nil
	}
	computeSvc, err := compute.NewService(ctx)
	if err != nil {
		log.Println(err)
		return nil
	}
	g.initProjectDepthResources(ctx, suSvc, computeSvc)
	return nil
}

func (g *ProjectGenerator) initProjectDepthResources(ctx context.Context, suSvc *serviceusagev1.Service, computeSvc *compute.Service) {
	project := g.GetArgs()["project"].(string)
	if suSvc != nil {
		err := suSvc.Services.List("projects/"+project).Filter("state:ENABLED").Pages(ctx, func(page *serviceusagev1.ListServicesResponse) error {
			for _, svc := range page.Services {
				short := svc.Name[strings.LastIndex(svc.Name, "/")+1:]
				id := project + "/" + short
				g.Resources = append(g.Resources, terraformutils.NewResource(
					id, short, "google_project_service", g.ProviderName,
					map[string]string{"project": project, "service": short},
					projectAllowEmptyValues, projectAdditionalFields))
			}
			return nil
		})
		if err != nil {
			log.Println(err)
		}
	}
	if computeSvc != nil {
		proj, err := computeSvc.Projects.Get(project).Context(ctx).Do()
		if err != nil {
			log.Println(err)
			return
		}
		if proj.UsageExportLocation != nil && proj.UsageExportLocation.BucketName != "" {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				project, project, "google_project_usage_export_bucket", g.ProviderName,
				map[string]string{"project": project, "bucket_name": proj.UsageExportLocation.BucketName},
				projectAllowEmptyValues, projectAdditionalFields))
		}
	}
}
