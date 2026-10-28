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

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	managedidentities "google.golang.org/api/managedidentities/v1"
)

// ActiveDirectoryGenerator enumerates Managed Microsoft AD via the
// generated managedidentities v1 client: domains in the global
// location, the domain trusts embedded in each domain resource
// (there is no separate trust collection or list call; trusts are
// attached/detached through the domain), and domain peerings.
type ActiveDirectoryGenerator struct {
	GCPService
}

func (g *ActiveDirectoryGenerator) InitResources() error {
	ctx := context.Background()
	service, err := managedidentities.NewService(ctx)
	if err != nil {
		return err
	}
	g.initResources(ctx, service)
	return nil
}

func (g *ActiveDirectoryGenerator) newResource(id, resourceType string) terraformutils.Resource {
	return terraformutils.NewResource(
		id, opsLastSegment(id), resourceType, g.ProviderName,
		map[string]string{"project": g.GetArgs()["project"].(string)},
		[]string{}, map[string]interface{}{},
	)
}

func (g *ActiveDirectoryGenerator) initResources(ctx context.Context, service *managedidentities.Service) {
	project := g.GetArgs()["project"].(string)
	parent := "projects/" + project + "/locations/global"
	if err := service.Projects.Locations.Global.Domains.List(parent).Pages(ctx, func(page *managedidentities.ListDomainsResponse) error {
		for _, domain := range page.Domains {
			g.Resources = append(g.Resources, g.newResource(domain.Name, "google_active_directory_domain"))
			for _, trust := range domain.Trusts {
				g.Resources = append(g.Resources, g.newResource(
					domain.Name+"/"+trust.TargetDomainName, "google_active_directory_domain_trust"))
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := service.Projects.Locations.Global.Peerings.List(parent).Pages(ctx, func(page *managedidentities.ListPeeringsResponse) error {
		for _, peering := range page.Peerings {
			g.Resources = append(g.Resources, g.newResource(peering.Name, "google_active_directory_peering"))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
}
