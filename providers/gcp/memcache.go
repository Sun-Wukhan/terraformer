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
	memcachev1 "google.golang.org/api/memcache/v1"
)

var memcacheAllowEmptyValues = []string{""}

var memcacheAdditionalFields = map[string]interface{}{}

type MemcacheGenerator struct {
	GCPService
}

// InitResources enumerates Memorystore for Memcached instances for
// the configured project and region. The provider ID is the full API
// resource name, its canonical import format.
func (g *MemcacheGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := memcachev1.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, svc)
}

func (g *MemcacheGenerator) initResources(ctx context.Context, svc *memcachev1.Service) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region)
	locationParent := "projects/" + project + "/locations/" + region.Name
	baseAttrs := map[string]string{"project": project, "region": region.Name}
	err := svc.Projects.Locations.Instances.List(locationParent).Pages(ctx, func(page *memcachev1.ListInstancesResponse) error {
		for _, instance := range page.Instances {
			parts := strings.Split(instance.Name, "/")
			g.Resources = append(g.Resources, terraformutils.NewResource(
				instance.Name,
				parts[len(parts)-1],
				"google_memcache_instance",
				g.ProviderName,
				baseAttrs,
				memcacheAllowEmptyValues,
				memcacheAdditionalFields,
			))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	return nil
}
