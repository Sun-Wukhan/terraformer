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
	eventarc "google.golang.org/api/eventarc/v1"
)

// EventarcGenerator enumerates Eventarc in the configured region via
// the generated eventarc v1 client: channels, enrollments, Google API
// sources, message buses, pipelines (with the combined
// google_eventarc_pipeline_iam resource when a pipeline carries IAM
// bindings), triggers, and the per-location Google channel config
// singleton (emitted when its get succeeds).
type EventarcGenerator struct {
	GCPService
}

func (g *EventarcGenerator) InitResources() error {
	ctx := context.Background()
	service, err := eventarc.NewService(ctx)
	if err != nil {
		return err
	}
	g.initResources(ctx, service)
	return nil
}

func (g *EventarcGenerator) newResource(id, resourceType string) terraformutils.Resource {
	return terraformutils.NewResource(
		id, opsLastSegment(id), resourceType, g.ProviderName,
		map[string]string{"project": g.GetArgs()["project"].(string)},
		[]string{}, map[string]interface{}{},
	)
}

func (g *EventarcGenerator) initResources(ctx context.Context, service *eventarc.Service) {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + region

	if err := service.Projects.Locations.Channels.List(parent).Pages(ctx, func(page *eventarc.ListChannelsResponse) error {
		for _, channel := range page.Channels {
			g.Resources = append(g.Resources, g.newResource(channel.Name, "google_eventarc_channel"))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := service.Projects.Locations.Enrollments.List(parent).Pages(ctx, func(page *eventarc.ListEnrollmentsResponse) error {
		for _, enrollment := range page.Enrollments {
			g.Resources = append(g.Resources, g.newResource(enrollment.Name, "google_eventarc_enrollment"))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := service.Projects.Locations.GoogleApiSources.List(parent).Pages(ctx, func(page *eventarc.ListGoogleApiSourcesResponse) error {
		for _, source := range page.GoogleApiSources {
			g.Resources = append(g.Resources, g.newResource(source.Name, "google_eventarc_google_api_source"))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := service.Projects.Locations.MessageBuses.List(parent).Pages(ctx, func(page *eventarc.ListMessageBusesResponse) error {
		for _, bus := range page.MessageBuses {
			g.Resources = append(g.Resources, g.newResource(bus.Name, "google_eventarc_message_bus"))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := service.Projects.Locations.Pipelines.List(parent).Pages(ctx, func(page *eventarc.ListPipelinesResponse) error {
		for _, pipeline := range page.Pipelines {
			g.Resources = append(g.Resources, g.newResource(pipeline.Name, "google_eventarc_pipeline"))
			policy, err := service.Projects.Locations.Pipelines.GetIamPolicy(pipeline.Name).Context(ctx).Do()
			if err != nil {
				log.Println(err)
				continue
			}
			if len(policy.Bindings) > 0 {
				// The provider models pipeline IAM as the single
				// combined _iam resource.
				g.Resources = append(g.Resources, g.newResource(pipeline.Name, "google_eventarc_pipeline_iam"))
			}
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := service.Projects.Locations.Triggers.List(parent).Pages(ctx, func(page *eventarc.ListTriggersResponse) error {
		for _, trigger := range page.Triggers {
			g.Resources = append(g.Resources, g.newResource(trigger.Name, "google_eventarc_trigger"))
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	// The Google channel config is a per-location singleton; its
	// provider ID is the location itself.
	if _, err := service.Projects.Locations.GetGoogleChannelConfig(parent).Context(ctx).Do(); err == nil {
		g.Resources = append(g.Resources, g.newResource(parent, "google_eventarc_google_channel_config"))
	} else {
		log.Println(err)
	}
}
