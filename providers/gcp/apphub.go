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
	"net/http"
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	apphub "google.golang.org/api/apphub/v1"
	"google.golang.org/api/compute/v1"
)

const appHubBaseURL = "https://apphub.googleapis.com/v1/"

var appHubAllowEmptyValues = []string{""}

var appHubAdditionalFields = map[string]interface{}{}

type AppHubGenerator struct {
	GCPService
}

// InitResources enumerates App Hub in the configured region:
// applications with their registered services and workloads, service
// project attachments, and the location's application-management
// boundary (a singleton fetched over raw REST; the generated client
// has no boundary surface). Provider IDs are the full resource names
// (the boundary's ID is projects/{p}/locations/{l}/boundary).
func (g *AppHubGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := apphub.NewService(ctx)
	if err != nil {
		return err
	}
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		log.Println(err)
		restClient = nil
	}
	return g.initResources(ctx, svc, restClient)
}

func (g *AppHubGenerator) initResources(ctx context.Context, svc *apphub.APIService, restClient *http.Client) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + region
	if restClient != nil {
		if _, err := restDo(ctx, restClient, http.MethodGet, appHubBaseURL+parent+"/boundary", nil); err == nil {
			g.Resources = append(g.Resources, g.newResource(parent+"/boundary", "google_apphub_boundary"))
		} else {
			log.Println(err)
		}
	}
	if err := svc.Projects.Locations.ServiceProjectAttachments.List(parent).Pages(ctx, func(page *apphub.ListServiceProjectAttachmentsResponse) error {
		for _, att := range page.ServiceProjectAttachments {
			g.Resources = append(g.Resources, g.newResource(att.Name, "google_apphub_service_project_attachment"))
		}
		return nil
	}); err != nil {
		return err
	}
	return svc.Projects.Locations.Applications.List(parent).Pages(ctx, func(page *apphub.ListApplicationsResponse) error {
		for _, app := range page.Applications {
			g.Resources = append(g.Resources, g.newResource(app.Name, "google_apphub_application"))
			if err := svc.Projects.Locations.Applications.Services.List(app.Name).Pages(ctx, func(spage *apphub.ListServicesResponse) error {
				for _, s := range spage.Services {
					g.Resources = append(g.Resources, g.newResource(s.Name, "google_apphub_service"))
				}
				return nil
			}); err != nil {
				log.Println(err)
			}
			if err := svc.Projects.Locations.Applications.Workloads.List(app.Name).Pages(ctx, func(wpage *apphub.ListWorkloadsResponse) error {
				for _, w := range wpage.Workloads {
					g.Resources = append(g.Resources, g.newResource(w.Name, "google_apphub_workload"))
				}
				return nil
			}); err != nil {
				log.Println(err)
			}
		}
		return nil
	})
}

func (g *AppHubGenerator) newResource(id, resourceType string) terraformutils.Resource {
	parts := strings.Split(id, "/")
	return terraformutils.NewResource(id, parts[len(parts)-1], resourceType, g.ProviderName,
		map[string]string{}, appHubAllowEmptyValues, appHubAdditionalFields)
}
