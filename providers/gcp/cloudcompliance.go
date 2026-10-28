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
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/compute/v1"
)

const cloudSecurityComplianceBaseURL = "https://cloudsecuritycompliance.googleapis.com/v1/"

type CloudComplianceGenerator struct {
	GCPService
}

// InitResources enumerates Cloud Security Compliance (Compliance
// Manager) over raw REST (the pinned google.golang.org/api has no
// cloudsecuritycompliance client). Frameworks, cloud controls and
// framework deployments live at organization scope in the global
// location; the organization is resolved from the configured
// project's Cloud Resource Manager ancestry, following the SCC
// generator's pattern. Provider IDs are the full resource names.
//
// Deliberately not generated:
//   - google_cloud_support_support_event_subscription: the Cloud
//     Support API discovery document (v2 and v2beta) exposes only
//     support cases and their attachments/comments — no
//     supportEventSubscriptions collection with a list or get
//     surface — so subscriptions cannot be discovered.
func (g *CloudComplianceGenerator) InitResources() error {
	ctx := context.Background()
	crmService, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		return err
	}
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		log.Println(err)
		return nil
	}
	g.initResources(ctx, crmService, restClient)
	return nil
}

func (g *CloudComplianceGenerator) initResources(ctx context.Context, crmService *cloudresourcemanager.Service, restClient *http.Client) {
	project := g.GetArgs()["project"].(string)
	_ = g.GetArgs()["region"].(compute.Region)
	orgID := ""
	if crmService != nil {
		ancestry, err := crmService.Projects.GetAncestry(project, &cloudresourcemanager.GetAncestryRequest{}).Context(ctx).Do()
		if err != nil {
			log.Println(err)
			return
		}
		for _, ancestor := range ancestry.Ancestor {
			if ancestor.ResourceId != nil && ancestor.ResourceId.Type == "organization" {
				orgID = ancestor.ResourceId.Id
			}
		}
	}
	if orgID == "" {
		log.Println("cloud compliance: no organization ancestor for project", project)
		return
	}
	parent := "organizations/" + orgID + "/locations/global"
	for _, coll := range []struct {
		segment string
		type_   string
	}{
		{"frameworks", "google_cloud_security_compliance_framework"},
		{"cloudControls", "google_cloud_security_compliance_cloud_control"},
		{"frameworkDeployments", "google_cloud_security_compliance_framework_deployment"},
	} {
		items, err := restListResources(ctx, restClient, cloudSecurityComplianceBaseURL+parent+"/"+coll.segment)
		if err != nil {
			log.Println(err)
			continue
		}
		for _, item := range items {
			parts := strings.Split(item.Name, "/")
			g.Resources = append(g.Resources, terraformutils.NewResource(
				item.Name, parts[len(parts)-1], coll.type_,
				g.ProviderName, map[string]string{}, []string{}, map[string]interface{}{}))
		}
	}
}
