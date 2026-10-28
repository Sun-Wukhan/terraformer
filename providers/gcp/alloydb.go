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
	alloydbv1 "google.golang.org/api/alloydb/v1"
	"google.golang.org/api/compute/v1"
)

var alloydbAllowEmptyValues = []string{""}

var alloydbAdditionalFields = map[string]interface{}{}

type AlloydbGenerator struct {
	GCPService
}

// InitResources enumerates AlloyDB resources for the configured
// project and region: backups, clusters, and the instances and users
// of each cluster. Provider IDs are the full API resource names, the
// canonical import format for all of these.
func (g *AlloydbGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := alloydbv1.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, svc)
}

func (g *AlloydbGenerator) initResources(ctx context.Context, svc *alloydbv1.Service) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region)
	locationParent := "projects/" + project + "/locations/" + region.Name
	baseAttrs := map[string]string{"project": project, "location": region.Name}
	err := svc.Projects.Locations.Backups.List(locationParent).Pages(ctx, func(page *alloydbv1.ListBackupsResponse) error {
		for _, backup := range page.Backups {
			g.Resources = append(g.Resources, g.newAlloydbResource(backup.Name, "google_alloydb_backup", baseAttrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = svc.Projects.Locations.Clusters.List(locationParent).Pages(ctx, func(page *alloydbv1.ListClustersResponse) error {
		for _, cluster := range page.Clusters {
			g.Resources = append(g.Resources, g.newAlloydbResource(cluster.Name, "google_alloydb_cluster", baseAttrs))
			g.loadAlloydbChildren(ctx, svc, cluster.Name, baseAttrs)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	return nil
}

func (g *AlloydbGenerator) newAlloydbResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
	if attrs == nil {
		attrs = map[string]string{}
	}
	parts := strings.Split(id, "/")
	return terraformutils.NewResource(
		id,
		parts[len(parts)-1],
		resourceType,
		g.ProviderName,
		attrs,
		alloydbAllowEmptyValues,
		alloydbAdditionalFields,
	)
}

func (g *AlloydbGenerator) loadAlloydbChildren(ctx context.Context, svc *alloydbv1.Service, clusterName string, baseAttrs map[string]string) {
	err := svc.Projects.Locations.Clusters.Instances.List(clusterName).Pages(ctx, func(page *alloydbv1.ListInstancesResponse) error {
		for _, instance := range page.Instances {
			g.Resources = append(g.Resources, g.newAlloydbResource(instance.Name, "google_alloydb_instance", baseAttrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = svc.Projects.Locations.Clusters.Users.List(clusterName).Pages(ctx, func(page *alloydbv1.ListUsersResponse) error {
		for _, user := range page.Users {
			g.Resources = append(g.Resources, g.newAlloydbResource(user.Name, "google_alloydb_user", baseAttrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}
