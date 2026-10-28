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
	"testing"

	"google.golang.org/api/compute/v1"
	managedkafka "google.golang.org/api/managedkafka/v1"
	"google.golang.org/api/option"
)

func TestManagedKafkaGeneratorEnumeration(t *testing.T) {
	ctx := context.Background()
	parent := "projects/test-project/locations/us-central1"
	cluster := parent + "/clusters/cl1"
	cc := parent + "/connectClusters/cc1"
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: "/topics", body: `{"topics":[{"name":"` + cluster + `/topics/topic1"}]}`},
		{substr: "/acls", body: `{"acls":[{"name":"` + cluster + `/acls/acl1"}]}`},
		{substr: "/connectors", body: `{"connectors":[{"name":"` + cc + `/connectors/conn1"}]}`},
		{substr: "/connectClusters", body: `{"connectClusters":[{"name":"` + cc + `"}]}`},
		{substr: "/clusters", body: `{"clusters":[{"name":"` + cluster + `"}]}`},
	}}
	svc, err := managedkafka.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	g := &ManagedKafkaGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project", "region": compute.Region{Name: "us-central1"}})
	if err := g.initResources(ctx, svc); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_managed_kafka_cluster":         {cluster},
		"google_managed_kafka_topic":           {cluster + "/topics/topic1"},
		"google_managed_kafka_acl":             {cluster + "/acls/acl1"},
		"google_managed_kafka_connect_cluster": {cc},
		"google_managed_kafka_connector":       {cc + "/connectors/conn1"},
	})
}
