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
	"google.golang.org/api/compute/v1"
)

const privilegedAccessBaseURL = "https://privilegedaccessmanager.googleapis.com/v1/"

type PrivilegedAccessGenerator struct {
	GCPService
}

// InitResources enumerates Privileged Access Manager entitlements in
// the configured region over raw REST (the pinned
// google.golang.org/api has no privilegedaccessmanager client).
// Provider IDs are the full resource names.
//
// Deliberately not generated:
//   - google_privileged_access_manager_settings: PAM settings are a
//     per-location singleton with no list or get surface exposed by
//     any generated client (none exists for this API), so an existing
//     settings object cannot be discovered.
func (g *PrivilegedAccessGenerator) InitResources() error {
	ctx := context.Background()
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		log.Println(err)
		return nil
	}
	g.initResources(ctx, restClient)
	return nil
}

func (g *PrivilegedAccessGenerator) initResources(ctx context.Context, restClient *http.Client) {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + region
	items, err := restListResources(ctx, restClient, privilegedAccessBaseURL+parent+"/entitlements")
	if err != nil {
		log.Println(err)
		return
	}
	for _, item := range items {
		parts := strings.Split(item.Name, "/")
		g.Resources = append(g.Resources, terraformutils.NewResource(
			item.Name, parts[len(parts)-1], "google_privileged_access_manager_entitlement",
			g.ProviderName, map[string]string{}, []string{}, map[string]interface{}{}))
	}
}
