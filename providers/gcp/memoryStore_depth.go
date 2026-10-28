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
	"google.golang.org/api/redis/v1"
)

const redisBaseURL = "https://redis.googleapis.com/v1/"

// This file extends MemoryStoreGenerator (memoryStore.go, which
// enumerates Memorystore for Redis instances) with Redis clusters and
// cluster ACL policies. Clusters come from the generated redis v1
// client; ACL policies have no surface in the pinned client, so they
// are enumerated through raw REST (projects.locations.aclPolicies
// .list in the Redis discovery document). Provider IDs are the full
// resource names for both.
//
// Deliberately not generated:
//   - google_redis_cluster_user_created_connections: user-created
//     connections are entries of the cluster object's own PSC
//     connection configuration, not a separate API collection — the
//     provider resource is an overlay on google_redis_cluster
//     carrying the cluster's own ID, so emitting it would
//     double-manage the cluster's connection spec.
func (g *MemoryStoreGenerator) initDepthResources(ctx context.Context, redisService *redis.Service, restClient *http.Client) {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region)
	locationParent := "projects/" + project + "/locations/" + region.Name
	baseAttrs := map[string]string{"project": project, "region": region.Name}
	if redisService != nil {
		err := redisService.Projects.Locations.Clusters.List(locationParent).Pages(ctx, func(page *redis.ListClustersResponse) error {
			for _, cluster := range page.Clusters {
				g.Resources = append(g.Resources, g.newRedisDepthResource(cluster.Name, "google_redis_cluster", baseAttrs))
			}
			return nil
		})
		if err != nil {
			log.Println(err)
		}
	}
	if restClient != nil {
		items, err := restListResources(ctx, restClient, redisBaseURL+locationParent+"/aclPolicies")
		if err != nil {
			log.Println(err)
		} else {
			for _, item := range items {
				g.Resources = append(g.Resources, g.newRedisDepthResource(item.Name, "google_redis_cluster_acl_policy", baseAttrs))
			}
		}
	}
}

func (g *MemoryStoreGenerator) newRedisDepthResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
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
		redisAllowEmptyValues,
		redisAdditionalFields,
	)
}
