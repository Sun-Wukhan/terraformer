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

package aws

// ElastiCache depth coverage (AWS sweep, database cluster).

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
)

func (g *ElastiCacheGenerator) loadElastiCacheDepth(svc *elasticache.Client) error {
	// aws_elasticache_global_replication_group, ID = global replication group ID.
	gp := elasticache.NewDescribeGlobalReplicationGroupsPaginator(svc, &elasticache.DescribeGlobalReplicationGroupsInput{})
	for gp.HasMorePages() {
		page, err := gp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, grg := range page.GlobalReplicationGroups {
			name := StringValue(grg.GlobalReplicationGroupId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_elasticache_global_replication_group", "aws", elastiCacheAllowEmptyValues))
		}
	}

	// aws_elasticache_reserved_cache_node, ID = reservation ID.
	rp := elasticache.NewDescribeReservedCacheNodesPaginator(svc, &elasticache.DescribeReservedCacheNodesInput{})
	for rp.HasMorePages() {
		page, err := rp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, node := range page.ReservedCacheNodes {
			name := StringValue(node.ReservedCacheNodeId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_elasticache_reserved_cache_node", "aws", elastiCacheAllowEmptyValues))
		}
	}

	// aws_elasticache_serverless_cache, ID = serverless cache name.
	sp := elasticache.NewDescribeServerlessCachesPaginator(svc, &elasticache.DescribeServerlessCachesInput{})
	for sp.HasMorePages() {
		page, err := sp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, cache := range page.ServerlessCaches {
			name := StringValue(cache.ServerlessCacheName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_elasticache_serverless_cache", "aws", elastiCacheAllowEmptyValues))
		}
	}

	// aws_elasticache_user, ID = user ID.
	up := elasticache.NewDescribeUsersPaginator(svc, &elasticache.DescribeUsersInput{})
	for up.HasMorePages() {
		page, err := up.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, user := range page.Users {
			name := StringValue(user.UserId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_elasticache_user", "aws", elastiCacheAllowEmptyValues))
		}
	}

	// aws_elasticache_user_group, ID = user group ID; plus
	// aws_elasticache_user_group_association, ID = "<user group ID>,<user ID>".
	ugp := elasticache.NewDescribeUserGroupsPaginator(svc, &elasticache.DescribeUserGroupsInput{})
	for ugp.HasMorePages() {
		page, err := ugp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, group := range page.UserGroups {
			groupID := StringValue(group.UserGroupId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				groupID, groupID, "aws_elasticache_user_group", "aws", elastiCacheAllowEmptyValues))
			for _, userID := range group.UserIds {
				id := groupID + "," + userID
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, groupID+"_"+userID, "aws_elasticache_user_group_association", "aws", elastiCacheAllowEmptyValues))
			}
		}
	}

	return nil
}
