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
	cloudbuildv1 "google.golang.org/api/cloudbuild/v1"
	cloudbuildv2 "google.golang.org/api/cloudbuild/v2"
	"google.golang.org/api/compute/v1"
)

// Cloud Build depth enumeration (ops sweep), a sibling service to the
// original CloudBuildGenerator (build triggers): worker pools and
// Bitbucket Server configs via the generated cloudbuild v1 REST
// client, and Cloud Build v2 connections (with the combined
// google_cloudbuildv2_connection_iam resource when a connection
// carries IAM bindings) and their repositories via the generated
// cloudbuild v2 client, all in the configured region.
type CloudBuildDepthGenerator struct {
	GCPService
}

func (g *CloudBuildDepthGenerator) InitResources() error {
	ctx := context.Background()
	v1Service, err := cloudbuildv1.NewService(ctx)
	if err != nil {
		return err
	}
	v2Service, err := cloudbuildv2.NewService(ctx)
	if err != nil {
		return err
	}
	g.initResources(ctx, v1Service, v2Service)
	return nil
}

func (g *CloudBuildDepthGenerator) newResource(id, resourceType string) terraformutils.Resource {
	return terraformutils.NewResource(
		id, opsLastSegment(id), resourceType, g.ProviderName,
		map[string]string{"project": g.GetArgs()["project"].(string)},
		[]string{}, map[string]interface{}{},
	)
}

func (g *CloudBuildDepthGenerator) initResources(ctx context.Context, v1Service *cloudbuildv1.Service, v2Service *cloudbuildv2.Service) {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + region

	if err := v1Service.Projects.Locations.WorkerPools.List(parent).Pages(ctx, func(page *cloudbuildv1.ListWorkerPoolsResponse) error {
		for _, pool := range page.WorkerPools {
			g.Resources = append(g.Resources, g.newResource(pool.Name, "google_cloudbuild_worker_pool"))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := v1Service.Projects.Locations.BitbucketServerConfigs.List(parent).Pages(ctx, func(page *cloudbuildv1.ListBitbucketServerConfigsResponse) error {
		for _, cfg := range page.BitbucketServerConfigs {
			g.Resources = append(g.Resources, g.newResource(cfg.Name, "google_cloudbuild_bitbucket_server_config"))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := v2Service.Projects.Locations.Connections.List(parent).Pages(ctx, func(page *cloudbuildv2.ListConnectionsResponse) error {
		for _, conn := range page.Connections {
			g.Resources = append(g.Resources, g.newResource(conn.Name, "google_cloudbuildv2_connection"))
			policy, err := v2Service.Projects.Locations.Connections.GetIamPolicy(conn.Name).Context(ctx).Do()
			if err != nil {
				log.Println(err)
			} else if len(policy.Bindings) > 0 {
				g.Resources = append(g.Resources, g.newResource(conn.Name, "google_cloudbuildv2_connection_iam"))
			}
			if err := v2Service.Projects.Locations.Connections.Repositories.List(conn.Name).Pages(ctx, func(repoPage *cloudbuildv2.ListRepositoriesResponse) error {
				for _, repo := range repoPage.Repositories {
					g.Resources = append(g.Resources, g.newResource(repo.Name, "google_cloudbuildv2_repository"))
				}
				return nil
			}); err != nil {
				log.Println(err)
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
}
