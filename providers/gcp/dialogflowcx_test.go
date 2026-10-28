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
	dialogflowv3 "google.golang.org/api/dialogflow/v3"
	dialogflowv3beta1 "google.golang.org/api/dialogflow/v3beta1"
	"google.golang.org/api/option"
)

// TestDialogflowCxGenerator feeds canned Dialogflow CX v3/v3beta1
// responses and verifies every resource type of the family: the
// agent tree (entity types, environments, flows with pages and
// versions, generators, intents, test cases, webhooks, playbooks,
// tools with tool versions, the generative-settings singleton) and
// the location-level security settings.
func TestDialogflowCxGenerator(t *testing.T) {
	loc := "projects/test-project/locations/us-central1"
	agent := loc + "/agents/agent1"
	flow := agent + "/flows/flow1"
	tool := agent + "/tools/tool1"
	transport := &cannedTransport{routes: []cannedRoute{
		{"/securitySettings", `{"securitySettings":[{"name":"` + loc + `/securitySettings/ss1"}]}`},
		{"/agents/agent1/generativeSettings", `{"name":"` + agent + `/generativeSettings"}`},
		{"/agents/agent1/entityTypes", `{"entityTypes":[{"name":"` + agent + `/entityTypes/et1"}]}`},
		{"/agents/agent1/environments", `{"environments":[{"name":"` + agent + `/environments/env1"}]}`},
		{"/flows/flow1/pages", `{"pages":[{"name":"` + flow + `/pages/page1"}]}`},
		{"/flows/flow1/versions", `{"versions":[{"name":"` + flow + `/versions/1"}]}`},
		{"/agents/agent1/flows", `{"flows":[{"name":"` + flow + `"}]}`},
		{"/agents/agent1/generators", `{"generators":[{"name":"` + agent + `/generators/gen1"}]}`},
		{"/agents/agent1/intents", `{"intents":[{"name":"` + agent + `/intents/intent1"}]}`},
		{"/agents/agent1/testCases", `{"testCases":[{"name":"` + agent + `/testCases/tc1"}]}`},
		{"/agents/agent1/webhooks", `{"webhooks":[{"name":"` + agent + `/webhooks/wh1"}]}`},
		{"/agents/agent1/playbooks", `{"playbooks":[{"name":"` + agent + `/playbooks/pb1"}]}`},
		{"/tools/tool1/versions", `{"toolVersions":[{"name":"` + tool + `/versions/1"}]}`},
		{"/agents/agent1/tools", `{"tools":[{"name":"` + tool + `"}]}`},
		{"/locations/us-central1/agents", `{"agents":[{"name":"` + agent + `"}]}`},
	}}
	ctx := context.Background()
	opts := []option.ClientOption{option.WithHTTPClient(transport.client()), option.WithoutAuthentication()}
	service, err := dialogflowv3.NewService(ctx, opts...)
	if err != nil {
		t.Fatalf("dialogflow v3 NewService: %v", err)
	}
	betaService, err := dialogflowv3beta1.NewService(ctx, opts...)
	if err != nil {
		t.Fatalf("dialogflow v3beta1 NewService: %v", err)
	}
	g := &DialogflowCxGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	g.initResources(ctx, service, betaService)
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_dialogflow_cx_agent":               {agent},
		"google_dialogflow_cx_entity_type":         {agent + "/entityTypes/et1"},
		"google_dialogflow_cx_environment":         {agent + "/environments/env1"},
		"google_dialogflow_cx_flow":                {flow},
		"google_dialogflow_cx_page":                {flow + "/pages/page1"},
		"google_dialogflow_cx_version":             {flow + "/versions/1"},
		"google_dialogflow_cx_generator":           {agent + "/generators/gen1"},
		"google_dialogflow_cx_intent":              {agent + "/intents/intent1"},
		"google_dialogflow_cx_test_case":           {agent + "/testCases/tc1"},
		"google_dialogflow_cx_webhook":             {agent + "/webhooks/wh1"},
		"google_dialogflow_cx_playbook":            {agent + "/playbooks/pb1"},
		"google_dialogflow_cx_tool":                {tool},
		"google_dialogflow_cx_tool_version":        {tool + "/versions/1"},
		"google_dialogflow_cx_generative_settings": {agent + "/generativeSettings"},
		"google_dialogflow_cx_security_settings":   {loc + "/securitySettings/ss1"},
	})
}
