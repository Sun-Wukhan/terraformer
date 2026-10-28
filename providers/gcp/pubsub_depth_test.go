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
	"google.golang.org/api/option"
	pubsub "google.golang.org/api/pubsub/v1"
	pubsublite "google.golang.org/api/pubsublite/v1"
)

// TestPubsubDepthEnumeration asserts schemas with their combined IAM
// resource, topic/subscription combined IAM resources, and the zonal
// Pub/Sub Lite collections.
func TestPubsubDepthEnumeration(t *testing.T) {
	ctx := context.Background()
	proj := "projects/test-project"
	lite := proj + "/locations/us-central1-a"
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: ":getIamPolicy", body: `{"bindings":[{"role":"roles/pubsub.publisher","members":["user:a@example.com"]}]}`},
		{substr: "locations/us-central1-a/topics", body: `{"topics":[{"name":"` + lite + `/topics/t1"}]}`},
		{substr: "locations/us-central1-a/subscriptions", body: `{"subscriptions":[{"name":"` + lite + `/subscriptions/sub1"}]}`},
		{substr: "/schemas", body: `{"schemas":[{"name":"` + proj + `/schemas/s1"}]}`},
		{substr: "/topics", body: `{"topics":[{"name":"` + proj + `/topics/t1"}]}`},
		{substr: "/subscriptions", body: `{"subscriptions":[{"name":"` + proj + `/subscriptions/sub1"}]}`},
		{substr: "/reservations", body: `{"reservations":[{"name":"` + lite + `/reservations/r1"}]}`},
	}}
	service, err := pubsub.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("pubsub NewService: %v", err)
	}
	liteService, err := pubsublite.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("pubsublite NewService: %v", err)
	}
	g := &PubsubDepthGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region": compute.Region{Name: "us-central1", Zones: []string{
			"https://www.googleapis.com/compute/v1/projects/test-project/zones/us-central1-a",
		}},
	})
	g.initResources(ctx, service, liteService)
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_pubsub_schema":            {proj + "/schemas/s1"},
		"google_pubsub_schema_iam":        {proj + "/schemas/s1"},
		"google_pubsub_topic_iam":         {proj + "/topics/t1"},
		"google_pubsub_subscription_iam":  {proj + "/subscriptions/sub1"},
		"google_pubsub_lite_topic":        {lite + "/topics/t1"},
		"google_pubsub_lite_subscription": {lite + "/subscriptions/sub1"},
		"google_pubsub_lite_reservation":  {lite + "/reservations/r1"},
	})
}
