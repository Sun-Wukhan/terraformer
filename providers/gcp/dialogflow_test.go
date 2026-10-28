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
	dialogflowv2 "google.golang.org/api/dialogflow/v2"
	dialogflowv2beta1 "google.golang.org/api/dialogflow/v2beta1"
	"google.golang.org/api/option"
)

// TestDialogflowGenerator feeds canned Dialogflow v2/v2beta1 and raw
// REST responses and verifies the whole ES/CCAI family: the agent
// and fulfillment singletons keyed by project, the agent's entity
// types, environments, intents and versions, and the global
// location's conversation profiles, generators, tools, SIP trunks
// and encryption spec.
func TestDialogflowGenerator(t *testing.T) {
	proj := "projects/test-project"
	agent := proj + "/agent"
	glob := proj + "/locations/global"
	transport := &cannedTransport{routes: []cannedRoute{
		{"/agent/entityTypes", `{"entityTypes":[{"name":"` + agent + `/entityTypes/et1"}]}`},
		{"/agent/environments", `{"environments":[{"name":"` + agent + `/environments/env1"}]}`},
		{"/agent/intents", `{"intents":[{"name":"` + agent + `/intents/intent1"}]}`},
		{"/agent/versions", `{"versions":[{"name":"` + agent + `/versions/1"}]}`},
		{"/agent/fulfillment", `{"name":"` + agent + `/fulfillment"}`},
		{"/v2/projects/test-project/agent", `{"parentProject":"test-project"}`},
		{"/locations/global/conversationProfiles", `{"conversationProfiles":[{"name":"` + glob + `/conversationProfiles/cp1"}]}`},
		{"/locations/global/generators", `{"generators":[{"name":"` + glob + `/generators/gen1"}]}`},
		{"/locations/global/tools", `{"tools":[{"name":"` + glob + `/tools/tool1"}]}`},
		{"/locations/global/sipTrunks", `{"sipTrunks":[{"name":"` + glob + `/sipTrunks/trunk1"}]}`},
		{"/locations/global/encryptionSpec", `{"name":"` + glob + `/encryptionSpec"}`},
	}}
	ctx := context.Background()
	opts := []option.ClientOption{option.WithHTTPClient(transport.client()), option.WithoutAuthentication()}
	service, err := dialogflowv2.NewService(ctx, opts...)
	if err != nil {
		t.Fatalf("dialogflow v2 NewService: %v", err)
	}
	betaService, err := dialogflowv2beta1.NewService(ctx, opts...)
	if err != nil {
		t.Fatalf("dialogflow v2beta1 NewService: %v", err)
	}
	g := &DialogflowGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: ""},
	})
	g.initResources(ctx, service, betaService, transport.client())
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_dialogflow_agent":                {"test-project"},
		"google_dialogflow_fulfillment":          {"test-project"},
		"google_dialogflow_entity_type":          {agent + "/entityTypes/et1"},
		"google_dialogflow_environment":          {agent + "/environments/env1"},
		"google_dialogflow_intent":               {agent + "/intents/intent1"},
		"google_dialogflow_version":              {agent + "/versions/1"},
		"google_dialogflow_conversation_profile": {glob + "/conversationProfiles/cp1"},
		"google_dialogflow_generator":            {glob + "/generators/gen1"},
		"google_dialogflow_tool":                 {glob + "/tools/tool1"},
		"google_dialogflow_sip_trunk":            {glob + "/sipTrunks/trunk1"},
		"google_dialogflow_encryption_spec":      {glob + "/encryptionSpec"},
	})
}
