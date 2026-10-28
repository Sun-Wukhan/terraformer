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
	"google.golang.org/api/compute/v1"
	pubsub "google.golang.org/api/pubsub/v1"
	pubsublite "google.golang.org/api/pubsublite/v1"
)

// Pubsub depth enumeration (ops sweep), a sibling service to the
// original PubsubGenerator (topics and subscriptions):
//
//   - schemas, via the generated pubsub v1 client, with the combined
//     google_pubsub_schema_iam resource when a schema carries IAM
//     bindings; the combined google_pubsub_topic_iam and
//     google_pubsub_subscription_iam resources are likewise emitted
//     for topics and subscriptions carrying bindings (the provider
//     models Pub/Sub IAM through these combined _iam resources);
//   - Pub/Sub Lite topics, subscriptions and reservations, via the
//     generated pubsublite v1 admin client, in every zone of the
//     configured region (Lite collections are zonal).
type PubsubDepthGenerator struct {
	GCPService
}

func (g *PubsubDepthGenerator) InitResources() error {
	ctx := context.Background()
	service, err := pubsub.NewService(ctx)
	if err != nil {
		return err
	}
	liteService, err := pubsublite.NewService(ctx)
	if err != nil {
		log.Println(err)
	}
	g.initResources(ctx, service, liteService)
	return nil
}

func (g *PubsubDepthGenerator) newResource(id, resourceType string) terraformutils.Resource {
	return terraformutils.NewResource(
		id, opsLastSegment(id), resourceType, g.ProviderName,
		map[string]string{"project": g.GetArgs()["project"].(string)},
		[]string{}, map[string]interface{}{},
	)
}

func (g *PubsubDepthGenerator) initResources(ctx context.Context, service *pubsub.Service, liteService *pubsublite.Service) {
	project := g.GetArgs()["project"].(string)
	parent := "projects/" + project

	if err := service.Projects.Schemas.List(parent).Pages(ctx, func(page *pubsub.ListSchemasResponse) error {
		for _, schema := range page.Schemas {
			g.Resources = append(g.Resources, g.newResource(schema.Name, "google_pubsub_schema"))
			policy, err := service.Projects.Schemas.GetIamPolicy(schema.Name).Context(ctx).Do()
			if err != nil {
				log.Println(err)
				continue
			}
			if len(policy.Bindings) > 0 {
				g.Resources = append(g.Resources, g.newResource(schema.Name, "google_pubsub_schema_iam"))
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := service.Projects.Topics.List(parent).Pages(ctx, func(page *pubsub.ListTopicsResponse) error {
		for _, topic := range page.Topics {
			policy, err := service.Projects.Topics.GetIamPolicy(topic.Name).Context(ctx).Do()
			if err != nil {
				log.Println(err)
				continue
			}
			if len(policy.Bindings) > 0 {
				g.Resources = append(g.Resources, g.newResource(topic.Name, "google_pubsub_topic_iam"))
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := service.Projects.Subscriptions.List(parent).Pages(ctx, func(page *pubsub.ListSubscriptionsResponse) error {
		for _, subscription := range page.Subscriptions {
			policy, err := service.Projects.Subscriptions.GetIamPolicy(subscription.Name).Context(ctx).Do()
			if err != nil {
				log.Println(err)
				continue
			}
			if len(policy.Bindings) > 0 {
				g.Resources = append(g.Resources, g.newResource(subscription.Name, "google_pubsub_subscription_iam"))
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}

	if liteService == nil {
		return
	}
	region := g.GetArgs()["region"].(compute.Region)
	for _, zone := range opsZones(region) {
		liteParent := parent + "/locations/" + zone
		if err := liteService.Admin.Projects.Locations.Topics.List(liteParent).Pages(ctx, func(page *pubsublite.ListTopicsResponse) error {
			for _, topic := range page.Topics {
				g.Resources = append(g.Resources, g.newResource(topic.Name, "google_pubsub_lite_topic"))
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
		if err := liteService.Admin.Projects.Locations.Subscriptions.List(liteParent).Pages(ctx, func(page *pubsublite.ListSubscriptionsResponse) error {
			for _, subscription := range page.Subscriptions {
				g.Resources = append(g.Resources, g.newResource(subscription.Name, "google_pubsub_lite_subscription"))
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
		if err := liteService.Admin.Projects.Locations.Reservations.List(liteParent).Pages(ctx, func(page *pubsublite.ListReservationsResponse) error {
			for _, reservation := range page.Reservations {
				g.Resources = append(g.Resources, g.newResource(reservation.Name, "google_pubsub_lite_reservation"))
			}
			return nil
		}); err != nil {
			log.Println(err)
		}
	}
}
