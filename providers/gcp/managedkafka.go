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
	managedkafka "google.golang.org/api/managedkafka/v1"
)

var managedKafkaAllowEmptyValues = []string{""}

var managedKafkaAdditionalFields = map[string]interface{}{}

type ManagedKafkaGenerator struct {
	GCPService
}

// InitResources enumerates Google Cloud Managed Service for Apache
// Kafka in the configured region: clusters with their topics and
// ACLs, and connect clusters with their connectors. Provider IDs
// are the full resource names.
func (g *ManagedKafkaGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := managedkafka.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, svc)
}

func (g *ManagedKafkaGenerator) initResources(ctx context.Context, svc *managedkafka.Service) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + region
	if err := svc.Projects.Locations.Clusters.List(parent).Pages(ctx, func(page *managedkafka.ListClustersResponse) error {
		for _, cluster := range page.Clusters {
			g.appendResource(cluster.Name, "google_managed_kafka_cluster")
			if err := svc.Projects.Locations.Clusters.Topics.List(cluster.Name).Pages(ctx, func(tpage *managedkafka.ListTopicsResponse) error {
				for _, topic := range tpage.Topics {
					g.appendResource(topic.Name, "google_managed_kafka_topic")
				}
				return nil
			}); err != nil {
				log.Println(err)
			}
			if err := svc.Projects.Locations.Clusters.Acls.List(cluster.Name).Pages(ctx, func(apage *managedkafka.ListAclsResponse) error {
				for _, acl := range apage.Acls {
					g.appendResource(acl.Name, "google_managed_kafka_acl")
				}
				return nil
			}); err != nil {
				log.Println(err)
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return svc.Projects.Locations.ConnectClusters.List(parent).Pages(ctx, func(page *managedkafka.ListConnectClustersResponse) error {
		for _, cc := range page.ConnectClusters {
			g.appendResource(cc.Name, "google_managed_kafka_connect_cluster")
			if err := svc.Projects.Locations.ConnectClusters.Connectors.List(cc.Name).Pages(ctx, func(cpage *managedkafka.ListConnectorsResponse) error {
				for _, conn := range cpage.Connectors {
					g.appendResource(conn.Name, "google_managed_kafka_connector")
				}
				return nil
			}); err != nil {
				log.Println(err)
			}
		}
		return nil
	})
}

func (g *ManagedKafkaGenerator) appendResource(name, resourceType string) {
	parts := strings.Split(name, "/")
	g.Resources = append(g.Resources, terraformutils.NewResource(
		name, parts[len(parts)-1], resourceType, g.ProviderName,
		map[string]string{}, managedKafkaAllowEmptyValues, managedKafkaAdditionalFields))
}
