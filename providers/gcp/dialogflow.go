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
	dialogflowv2 "google.golang.org/api/dialogflow/v2"
	dialogflowv2beta1 "google.golang.org/api/dialogflow/v2beta1"
)

// DialogflowGenerator enumerates the Dialogflow (ES / CCAI) surface
// of one project: the per-project agent singleton and its entity
// types, environments, intents, versions and fulfillment, plus the
// per-location conversation profiles, generators, tools, SIP trunks
// and encryption spec singletons. The agent and fulfillment
// singletons are keyed by project in the provider, so their IDs are
// the project ID; every other ID is the resource's full API name.
// The ES agent has no Get method in the generated v2 client and the
// tools collection is missing from both generated clients, so those
// two reads run over raw REST against dialogflow.googleapis.com/v2.
type DialogflowGenerator struct {
	GCPService
}

const dialogflowBaseURL = "https://dialogflow.googleapis.com/v2/"

func (g *DialogflowGenerator) InitResources() error {
	ctx := context.Background()
	service, err := dialogflowv2.NewService(ctx)
	if err != nil {
		return err
	}
	betaService, err := dialogflowv2beta1.NewService(ctx)
	if err != nil {
		return err
	}
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		log.Println(err)
		return nil
	}
	g.initResources(ctx, service, betaService, restClient)
	return nil
}

func (g *DialogflowGenerator) initResources(ctx context.Context, service *dialogflowv2.Service, betaService *dialogflowv2beta1.Service, restClient *http.Client) {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	agentParent := "projects/" + project + "/agent"

	// The ES agent singleton: it exists in every project that has
	// ever used Dialogflow; the provider keys it by project ID.
	if _, err := restDo(ctx, restClient, http.MethodGet, dialogflowBaseURL+"projects/"+project+"/agent", nil); err != nil {
		log.Println(err)
	} else {
		g.Resources = append(g.Resources, g.newDialogflowResource(project, project, "google_dialogflow_agent"))
	}
	if _, err := service.Projects.Agent.GetFulfillment(agentParent).Context(ctx).Do(); err != nil {
		log.Println(err)
	} else {
		g.Resources = append(g.Resources, g.newDialogflowResource(project, project, "google_dialogflow_fulfillment"))
	}

	entityTypes, err := cxListNames(func(pageToken string) ([]string, string, error) {
		resp, err := service.Projects.Agent.EntityTypes.List(agentParent).PageToken(pageToken).Context(ctx).Do()
		if err != nil {
			return nil, "", err
		}
		names := []string{}
		for _, entityType := range resp.EntityTypes {
			names = append(names, entityType.Name)
		}
		return names, resp.NextPageToken, nil
	})
	g.appendDialogflowList(entityTypes, err, "google_dialogflow_entity_type")

	environments, err := cxListNames(func(pageToken string) ([]string, string, error) {
		resp, err := service.Projects.Agent.Environments.List(agentParent).PageToken(pageToken).Context(ctx).Do()
		if err != nil {
			return nil, "", err
		}
		names := []string{}
		for _, environment := range resp.Environments {
			names = append(names, environment.Name)
		}
		return names, resp.NextPageToken, nil
	})
	g.appendDialogflowList(environments, err, "google_dialogflow_environment")

	intents, err := cxListNames(func(pageToken string) ([]string, string, error) {
		resp, err := service.Projects.Agent.Intents.List(agentParent).PageToken(pageToken).Context(ctx).Do()
		if err != nil {
			return nil, "", err
		}
		names := []string{}
		for _, intent := range resp.Intents {
			names = append(names, intent.Name)
		}
		return names, resp.NextPageToken, nil
	})
	g.appendDialogflowList(intents, err, "google_dialogflow_intent")

	versions, err := cxListNames(func(pageToken string) ([]string, string, error) {
		resp, err := service.Projects.Agent.Versions.List(agentParent).PageToken(pageToken).Context(ctx).Do()
		if err != nil {
			return nil, "", err
		}
		names := []string{}
		for _, version := range resp.Versions {
			names = append(names, version.Name)
		}
		return names, resp.NextPageToken, nil
	})
	g.appendDialogflowList(versions, err, "google_dialogflow_version")

	locations := []string{"global"}
	if region != "" && region != "global" {
		locations = append(locations, region)
	}
	for _, location := range locations {
		g.dialogflowLocation(ctx, service, betaService, restClient, "projects/"+project+"/locations/"+location)
	}
}

func (g *DialogflowGenerator) dialogflowLocation(ctx context.Context, service *dialogflowv2.Service, betaService *dialogflowv2beta1.Service, restClient *http.Client, parent string) {
	profiles, err := cxListNames(func(pageToken string) ([]string, string, error) {
		resp, err := service.Projects.Locations.ConversationProfiles.List(parent).PageToken(pageToken).Context(ctx).Do()
		if err != nil {
			return nil, "", err
		}
		names := []string{}
		for _, profile := range resp.ConversationProfiles {
			names = append(names, profile.Name)
		}
		return names, resp.NextPageToken, nil
	})
	g.appendDialogflowList(profiles, err, "google_dialogflow_conversation_profile")

	generators, err := cxListNames(func(pageToken string) ([]string, string, error) {
		resp, err := service.Projects.Locations.Generators.List(parent).PageToken(pageToken).Context(ctx).Do()
		if err != nil {
			return nil, "", err
		}
		names := []string{}
		for _, generator := range resp.Generators {
			names = append(names, generator.Name)
		}
		return names, resp.NextPageToken, nil
	})
	g.appendDialogflowList(generators, err, "google_dialogflow_generator")

	tools, err := restListResources(ctx, restClient, dialogflowBaseURL+parent+"/tools")
	if err != nil {
		log.Println(err)
	} else {
		for _, tool := range tools {
			g.Resources = append(g.Resources, g.newDialogflowResource(tool.Name, tool.Name, "google_dialogflow_tool"))
		}
	}

	sipTrunks, err := cxListNames(func(pageToken string) ([]string, string, error) {
		resp, err := betaService.Projects.Locations.SipTrunks.List(parent).PageToken(pageToken).Context(ctx).Do()
		if err != nil {
			return nil, "", err
		}
		names := []string{}
		for _, trunk := range resp.SipTrunks {
			names = append(names, trunk.Name)
		}
		return names, resp.NextPageToken, nil
	})
	g.appendDialogflowList(sipTrunks, err, "google_dialogflow_sip_trunk")

	// The encryption spec is a per-location singleton; it only exists
	// once CMEK has been initialized for the location.
	specName := parent + "/encryptionSpec"
	if name, err := aiRestSingletonName(ctx, restClient, dialogflowBaseURL+specName); err != nil {
		log.Println(err)
	} else if name != "" {
		g.Resources = append(g.Resources, g.newDialogflowResource(name, name, "google_dialogflow_encryption_spec"))
	}
}

func (g *DialogflowGenerator) newDialogflowResource(id, name, resourceType string) terraformutils.Resource {
	parts := strings.Split(name, "/")
	return terraformutils.NewResource(
		id,
		parts[len(parts)-1],
		resourceType,
		g.ProviderName,
		map[string]string{"project": g.GetArgs()["project"].(string)},
		[]string{},
		map[string]interface{}{},
	)
}

func (g *DialogflowGenerator) appendDialogflowList(names []string, err error, resourceType string) {
	if err != nil {
		log.Println(err)
		return
	}
	for _, name := range names {
		g.Resources = append(g.Resources, g.newDialogflowResource(name, name, resourceType))
	}
}
