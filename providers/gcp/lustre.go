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

var lustreAllowEmptyValues = []string{""}

var lustreAdditionalFields = map[string]interface{}{}

const lustreBaseURL = "https://lustre.googleapis.com/v1/"

// LustreGenerator enumerates Google Cloud Managed Lustre instances
// in the configured region over raw REST — the Lustre API
// (lustre.googleapis.com) has no generated client in the pinned
// google.golang.org/api module.
type LustreGenerator struct {
	GCPService
}

func (g *LustreGenerator) InitResources() error {
	ctx := context.Background()
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		return err
	}
	g.initResources(ctx, restClient)
	return nil
}

func (g *LustreGenerator) initResources(ctx context.Context, restClient *http.Client) {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + region
	baseAttrs := map[string]string{"project": project, "location": region}
	instances, err := restListResources(ctx, restClient, lustreBaseURL+parent+"/instances")
	if err != nil {
		log.Println(err)
		return
	}
	for _, instance := range instances {
		parts := strings.Split(instance.Name, "/")
		g.Resources = append(g.Resources, terraformutils.NewResource(
			instance.Name,
			parts[len(parts)-1],
			"google_lustre_instance",
			g.ProviderName,
			baseAttrs,
			lustreAllowEmptyValues,
			lustreAdditionalFields,
		))
	}
}
