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
	integrations "google.golang.org/api/integrations/v1"
)

var integrationsAllowEmptyValues = []string{""}

var integrationsAdditionalFields = map[string]interface{}{}

type IntegrationsGenerator struct {
	GCPService
}

// InitResources enumerates Application Integration auth configs in
// the configured region. Provider IDs are the full resource names.
//
// Deliberately not generated:
//   - google_integrations_client: the Application Integration API
//     exposes only provision and deprovision methods on the
//     per-location client singleton (no get or list in the generated
//     client or the discovery document), so an existing client
//     registration cannot be discovered or read back.
func (g *IntegrationsGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := integrations.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, svc)
}

func (g *IntegrationsGenerator) initResources(ctx context.Context, svc *integrations.Service) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + region
	return svc.Projects.Locations.AuthConfigs.List(parent).Pages(ctx, func(page *integrations.GoogleCloudIntegrationsV1alphaListAuthConfigsResponse) error {
		for _, item := range page.AuthConfigs {
			parts := strings.Split(item.Name, "/")
			g.Resources = append(g.Resources, terraformutils.NewResource(
				item.Name, parts[len(parts)-1], "google_integrations_auth_config",
				g.ProviderName, map[string]string{}, integrationsAllowEmptyValues, integrationsAdditionalFields))
		}
		return nil
	})
}
