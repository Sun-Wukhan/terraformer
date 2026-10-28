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
	domains "google.golang.org/api/domains/v1"
)

type CloudDomainsGenerator struct {
	GCPService
}

// InitResources enumerates Cloud Domains registrations in the
// global location. Provider IDs are the full resource names
// (projects/{p}/locations/global/registrations/{domain}).
func (g *CloudDomainsGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := domains.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, svc)
}

func (g *CloudDomainsGenerator) initResources(ctx context.Context, svc *domains.Service) error {
	project := g.GetArgs()["project"].(string)
	_ = g.GetArgs()["region"].(compute.Region)
	parent := "projects/" + project + "/locations/global"
	return svc.Projects.Locations.Registrations.List(parent).Pages(ctx, func(page *domains.ListRegistrationsResponse) error {
		for _, reg := range page.Registrations {
			parts := strings.Split(reg.Name, "/")
			g.Resources = append(g.Resources, terraformutils.NewResource(
				reg.Name, parts[len(parts)-1], "google_clouddomains_registration",
				g.ProviderName, map[string]string{}, []string{}, map[string]interface{}{}))
		}
		return nil
	})
}
