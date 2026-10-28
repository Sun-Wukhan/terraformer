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
	dialogflowv3 "google.golang.org/api/dialogflow/v3"
	dialogflowv3beta1 "google.golang.org/api/dialogflow/v3beta1"
	"google.golang.org/api/option"
)

// DialogflowCxGenerator enumerates the Dialogflow CX surface of one
// project/location: agents with their entity types, environments,
// flows (with pages and versions), generators, intents, test cases,
// webhooks and generative settings, plus the location's security
// settings. Playbooks and tools (with tool versions) are only exposed
// by the v3beta1 API, so they are enumerated through the v3beta1
// client; everything else uses the GA v3 client. Every provider ID
// in this family is the resource's full API name.
type DialogflowCxGenerator struct {
	GCPService
}

func (g *DialogflowCxGenerator) InitResources() error {
	ctx := context.Background()
	location := g.cxLocation()
	endpoint := "https://dialogflow.googleapis.com/"
	if location != "global" {
		endpoint = "https://" + location + "-dialogflow.googleapis.com/"
	}
	service, err := dialogflowv3.NewService(ctx, option.WithEndpoint(endpoint+"v3/"))
	if err != nil {
		return err
	}
	betaService, err := dialogflowv3beta1.NewService(ctx, option.WithEndpoint(endpoint+"v3beta1/"))
	if err != nil {
		return err
	}
	g.initResources(ctx, service, betaService)
	return nil
}

func (g *DialogflowCxGenerator) cxLocation() string {
	region := g.GetArgs()["region"].(compute.Region).Name
	if region == "" {
		return "global"
	}
	return region
}

func (g *DialogflowCxGenerator) initResources(ctx context.Context, service *dialogflowv3.Service, betaService *dialogflowv3beta1.Service) {
	project := g.GetArgs()["project"].(string)
	parent := "projects/" + project + "/locations/" + g.cxLocation()
	g.cxSecuritySettings(ctx, service, parent)
	agents, err := cxListNames(func(pageToken string) ([]string, string, error) {
		resp, err := service.Projects.Locations.Agents.List(parent).PageToken(pageToken).Context(ctx).Do()
		if err != nil {
			return nil, "", err
		}
		names := []string{}
		for _, agent := range resp.Agents {
			names = append(names, agent.Name)
		}
		return names, resp.NextPageToken, nil
	})
	if err != nil {
		log.Println(err)
		return
	}
	for _, agent := range agents {
		g.Resources = append(g.Resources, g.newCxResource(agent, "google_dialogflow_cx_agent"))
		g.cxAgentChildren(ctx, service, betaService, agent)
	}
}

// cxListNames pages a list method that returns resource names.
func cxListNames(list func(pageToken string) ([]string, string, error)) ([]string, error) {
	names := []string{}
	pageToken := ""
	for {
		page, next, err := list(pageToken)
		if err != nil {
			return nil, err
		}
		names = append(names, page...)
		if next == "" {
			return names, nil
		}
		pageToken = next
	}
}

func (g *DialogflowCxGenerator) newCxResource(id, resourceType string) terraformutils.Resource {
	parts := strings.Split(id, "/")
	return terraformutils.NewResource(
		id,
		parts[len(parts)-1],
		resourceType,
		g.ProviderName,
		map[string]string{
			"project":  g.GetArgs()["project"].(string),
			"location": g.cxLocation(),
		},
		[]string{},
		map[string]interface{}{},
	)
}

func (g *DialogflowCxGenerator) appendCxList(names []string, err error, resourceType string) {
	if err != nil {
		log.Println(err)
		return
	}
	for _, name := range names {
		g.Resources = append(g.Resources, g.newCxResource(name, resourceType))
	}
}

func (g *DialogflowCxGenerator) cxSecuritySettings(ctx context.Context, service *dialogflowv3.Service, parent string) {
	names, err := cxListNames(func(pageToken string) ([]string, string, error) {
		resp, err := service.Projects.Locations.SecuritySettings.List(parent).PageToken(pageToken).Context(ctx).Do()
		if err != nil {
			return nil, "", err
		}
		names := []string{}
		for _, setting := range resp.SecuritySettings {
			names = append(names, setting.Name)
		}
		return names, resp.NextPageToken, nil
	})
	g.appendCxList(names, err, "google_dialogflow_cx_security_settings")
}

func (g *DialogflowCxGenerator) cxAgentChildren(ctx context.Context, service *dialogflowv3.Service, betaService *dialogflowv3beta1.Service, agent string) {
	// Generative settings are a singleton under the agent.
	if settings, err := service.Projects.Locations.Agents.GetGenerativeSettings(agent).Context(ctx).Do(); err != nil {
		log.Println(err)
	} else {
		name := settings.Name
		if name == "" {
			name = agent + "/generativeSettings"
		}
		g.Resources = append(g.Resources, g.newCxResource(name, "google_dialogflow_cx_generative_settings"))
	}

	entityTypes, err := cxListNames(func(pageToken string) ([]string, string, error) {
		resp, err := service.Projects.Locations.Agents.EntityTypes.List(agent).PageToken(pageToken).Context(ctx).Do()
		if err != nil {
			return nil, "", err
		}
		names := []string{}
		for _, entityType := range resp.EntityTypes {
			names = append(names, entityType.Name)
		}
		return names, resp.NextPageToken, nil
	})
	g.appendCxList(entityTypes, err, "google_dialogflow_cx_entity_type")

	environments, err := cxListNames(func(pageToken string) ([]string, string, error) {
		resp, err := service.Projects.Locations.Agents.Environments.List(agent).PageToken(pageToken).Context(ctx).Do()
		if err != nil {
			return nil, "", err
		}
		names := []string{}
		for _, environment := range resp.Environments {
			names = append(names, environment.Name)
		}
		return names, resp.NextPageToken, nil
	})
	g.appendCxList(environments, err, "google_dialogflow_cx_environment")

	flows, err := cxListNames(func(pageToken string) ([]string, string, error) {
		resp, err := service.Projects.Locations.Agents.Flows.List(agent).PageToken(pageToken).Context(ctx).Do()
		if err != nil {
			return nil, "", err
		}
		names := []string{}
		for _, flow := range resp.Flows {
			names = append(names, flow.Name)
		}
		return names, resp.NextPageToken, nil
	})
	if err != nil {
		log.Println(err)
	} else {
		for _, flow := range flows {
			g.Resources = append(g.Resources, g.newCxResource(flow, "google_dialogflow_cx_flow"))
			g.cxFlowChildren(ctx, service, flow)
		}
	}

	generators, err := cxListNames(func(pageToken string) ([]string, string, error) {
		resp, err := service.Projects.Locations.Agents.Generators.List(agent).PageToken(pageToken).Context(ctx).Do()
		if err != nil {
			return nil, "", err
		}
		names := []string{}
		for _, generator := range resp.Generators {
			names = append(names, generator.Name)
		}
		return names, resp.NextPageToken, nil
	})
	g.appendCxList(generators, err, "google_dialogflow_cx_generator")

	intents, err := cxListNames(func(pageToken string) ([]string, string, error) {
		resp, err := service.Projects.Locations.Agents.Intents.List(agent).PageToken(pageToken).Context(ctx).Do()
		if err != nil {
			return nil, "", err
		}
		names := []string{}
		for _, intent := range resp.Intents {
			names = append(names, intent.Name)
		}
		return names, resp.NextPageToken, nil
	})
	g.appendCxList(intents, err, "google_dialogflow_cx_intent")

	testCases, err := cxListNames(func(pageToken string) ([]string, string, error) {
		resp, err := service.Projects.Locations.Agents.TestCases.List(agent).PageToken(pageToken).Context(ctx).Do()
		if err != nil {
			return nil, "", err
		}
		names := []string{}
		for _, testCase := range resp.TestCases {
			names = append(names, testCase.Name)
		}
		return names, resp.NextPageToken, nil
	})
	g.appendCxList(testCases, err, "google_dialogflow_cx_test_case")

	webhooks, err := cxListNames(func(pageToken string) ([]string, string, error) {
		resp, err := service.Projects.Locations.Agents.Webhooks.List(agent).PageToken(pageToken).Context(ctx).Do()
		if err != nil {
			return nil, "", err
		}
		names := []string{}
		for _, webhook := range resp.Webhooks {
			names = append(names, webhook.Name)
		}
		return names, resp.NextPageToken, nil
	})
	g.appendCxList(webhooks, err, "google_dialogflow_cx_webhook")

	playbooks, err := cxListNames(func(pageToken string) ([]string, string, error) {
		resp, err := betaService.Projects.Locations.Agents.Playbooks.List(agent).PageToken(pageToken).Context(ctx).Do()
		if err != nil {
			return nil, "", err
		}
		names := []string{}
		for _, playbook := range resp.Playbooks {
			names = append(names, playbook.Name)
		}
		return names, resp.NextPageToken, nil
	})
	g.appendCxList(playbooks, err, "google_dialogflow_cx_playbook")

	tools, err := cxListNames(func(pageToken string) ([]string, string, error) {
		resp, err := betaService.Projects.Locations.Agents.Tools.List(agent).PageToken(pageToken).Context(ctx).Do()
		if err != nil {
			return nil, "", err
		}
		names := []string{}
		for _, tool := range resp.Tools {
			names = append(names, tool.Name)
		}
		return names, resp.NextPageToken, nil
	})
	if err != nil {
		log.Println(err)
		return
	}
	for _, tool := range tools {
		g.Resources = append(g.Resources, g.newCxResource(tool, "google_dialogflow_cx_tool"))
		versions, err := cxListNames(func(pageToken string) ([]string, string, error) {
			resp, err := betaService.Projects.Locations.Agents.Tools.Versions.List(tool).PageToken(pageToken).Context(ctx).Do()
			if err != nil {
				return nil, "", err
			}
			names := []string{}
			for _, version := range resp.ToolVersions {
				names = append(names, version.Name)
			}
			return names, resp.NextPageToken, nil
		})
		g.appendCxList(versions, err, "google_dialogflow_cx_tool_version")
	}
}

func (g *DialogflowCxGenerator) cxFlowChildren(ctx context.Context, service *dialogflowv3.Service, flow string) {
	pages, err := cxListNames(func(pageToken string) ([]string, string, error) {
		resp, err := service.Projects.Locations.Agents.Flows.Pages.List(flow).PageToken(pageToken).Context(ctx).Do()
		if err != nil {
			return nil, "", err
		}
		names := []string{}
		for _, page := range resp.Pages {
			names = append(names, page.Name)
		}
		return names, resp.NextPageToken, nil
	})
	g.appendCxList(pages, err, "google_dialogflow_cx_page")

	versions, err := cxListNames(func(pageToken string) ([]string, string, error) {
		resp, err := service.Projects.Locations.Agents.Flows.Versions.List(flow).PageToken(pageToken).Context(ctx).Do()
		if err != nil {
			return nil, "", err
		}
		names := []string{}
		for _, version := range resp.Versions {
			names = append(names, version.Name)
		}
		return names, resp.NextPageToken, nil
	})
	g.appendCxList(versions, err, "google_dialogflow_cx_version")
}
