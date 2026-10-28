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

const saasRuntimeBaseURL = "https://saasruntime.googleapis.com/v1/"

type SaasRuntimeGenerator struct {
	GCPService
}

// InitResources enumerates SaaS Runtime in the configured region
// over raw REST (the pinned google.golang.org/api has no
// saasruntime client): SaaS definitions, tenants, releases, rollout
// kinds, unit kinds and units. Provider IDs are the full resource
// names.
//
// Deliberately not generated:
//   - google_saas_runtime_unit_operation: unit operations are
//     transient execution records of actions run against a unit, not
//     durable configuration.
func (g *SaasRuntimeGenerator) InitResources() error {
	ctx := context.Background()
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		log.Println(err)
		return nil
	}
	g.initResources(ctx, restClient)
	return nil
}

var saasRuntimeCollections = []struct {
	segment string
	type_   string
}{
	{"saas", "google_saas_runtime_saas"},
	{"tenants", "google_saas_runtime_tenant"},
	{"releases", "google_saas_runtime_release"},
	{"rolloutKinds", "google_saas_runtime_rollout_kind"},
	{"unitKinds", "google_saas_runtime_unit_kind"},
	{"units", "google_saas_runtime_unit"},
}

func (g *SaasRuntimeGenerator) initResources(ctx context.Context, restClient *http.Client) {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + region
	for _, coll := range saasRuntimeCollections {
		items, err := restListResources(ctx, restClient, saasRuntimeBaseURL+parent+"/"+coll.segment)
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
