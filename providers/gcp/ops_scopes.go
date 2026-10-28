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

	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/compute/v1"
)

// opsScopes returns the resource-manager scopes the ops-sweep depth
// generators enumerate: the configured project always, plus its
// ancestor folders and organization when Cloud Resource Manager
// ancestry resolves them (the same discovery path SccGenerator uses).
// It reuses the sccScope type so scope handling stays uniform.
func opsScopes(ctx context.Context, crmService *cloudresourcemanager.Service, project string) []sccScope {
	scopes := []sccScope{{
		kind:   "project",
		parent: "projects/" + project,
		id:     project,
		attrs:  map[string]string{"project": project},
	}}
	if crmService == nil {
		return scopes
	}
	ancestry, err := crmService.Projects.GetAncestry(project, &cloudresourcemanager.GetAncestryRequest{}).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		return scopes
	}
	var folders []sccScope
	var org *sccScope
	for _, ancestor := range ancestry.Ancestor {
		if ancestor.ResourceId == nil {
			continue
		}
		switch ancestor.ResourceId.Type {
		case "folder":
			folders = append(folders, sccScope{
				kind:   "folder",
				parent: "folders/" + ancestor.ResourceId.Id,
				id:     ancestor.ResourceId.Id,
				attrs:  map[string]string{"folder": ancestor.ResourceId.Id},
			})
		case "organization":
			orgScope := sccScope{
				kind:   "organization",
				parent: "organizations/" + ancestor.ResourceId.Id,
				id:     ancestor.ResourceId.Id,
				attrs:  map[string]string{"organization": ancestor.ResourceId.Id},
			}
			org = &orgScope
		}
	}
	scopes = append(scopes, folders...)
	if org != nil {
		scopes = append(scopes, *org)
	}
	return scopes
}

// opsZones returns the zone IDs of the configured region argument.
func opsZones(region compute.Region) []string {
	zones := []string{}
	for _, zone := range region.Zones {
		parts := strings.Split(zone, "/")
		zones = append(zones, parts[len(parts)-1])
	}
	return zones
}

// opsLastSegment returns the final path segment of a resource name.
func opsLastSegment(name string) string {
	parts := strings.Split(name, "/")
	return parts[len(parts)-1]
}
