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
)

// TestGeminiGenerator feeds canned Cloud AI Companion and Gemini
// Data Analytics responses through the raw-REST generator and
// verifies the emitted types and IDs: a code repository index with
// its repository group and the group's combined IAM resource, one setting + setting binding
// for each of the seven settings collections in the global location,
// a second logging setting in the configured region (proving both
// locations are walked), and a data agent with its combined IAM resource.
func TestGeminiGenerator(t *testing.T) {
	proj := "projects/test-project"
	globalLoc := proj + "/locations/global"
	regionLoc := proj + "/locations/us-central1"
	index := globalLoc + "/codeRepositoryIndexes/idx1"
	agent := globalLoc + "/dataAgents/da1"
	transport := &cannedTransport{routes: []cannedRoute{
		{":getIamPolicy", `{"bindings":[{"role":"roles/cloudaicompanion.user","members":["user:a@example.com"]}]}`},
		{"/codeRepositoryIndexes/idx1/repositoryGroups", `{"repositoryGroups":[{"name":"` + index + `/repositoryGroups/rg1"}]}`},
		{"/locations/global/codeRepositoryIndexes", `{"codeRepositoryIndexes":[{"name":"` + index + `"}]}`},
		{"/locations/global/dataAgents", `{"dataAgents":[{"name":"` + agent + `"}]}`},
		// One regional setting: the region location is walked too.
		// (Bindings route first — its path contains the list path.)
		{"/locations/us-central1/loggingSettings/s1/settingBindings", `{"settingBindings":[{"name":"` + regionLoc + `/loggingSettings/s1/settingBindings/b1"}]}`},
		{"/locations/us-central1/loggingSettings", `{"loggingSettings":[{"name":"` + regionLoc + `/loggingSettings/s1"}]}`},
	}}
	// The seven settings collections share one shape in the global
	// location. Binding routes are registered first: a setting's
	// settingBindings path also contains the collection segment the
	// settings-list route matches on.
	for _, segment := range []string{
		"codeToolsSettings",
		"dataSharingWithGoogleSettings",
		"geminiGcpEnablementSettings",
		"gdaObservabilitySettings",
		"gibqObservabilitySettings",
		"loggingSettings",
		"releaseChannelSettings",
	} {
		setting := globalLoc + "/" + segment + "/s1"
		transport.routes = append(transport.routes,
			cannedRoute{"/" + segment + "/s1/settingBindings", `{"settingBindings":[{"name":"` + setting + `/settingBindings/b1"}]}`},
		)
	}
	for _, segment := range []string{
		"codeToolsSettings",
		"dataSharingWithGoogleSettings",
		"geminiGcpEnablementSettings",
		"gdaObservabilitySettings",
		"gibqObservabilitySettings",
		"loggingSettings",
		"releaseChannelSettings",
	} {
		setting := globalLoc + "/" + segment + "/s1"
		transport.routes = append(transport.routes,
			cannedRoute{"/locations/global/" + segment, `{"settings":[{"name":"` + setting + `"}]}`},
		)
	}

	g := &GeminiGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	g.initResources(context.Background(), transport.client())

	expected := map[string][]string{
		"google_gemini_code_repository_index":         {index},
		"google_gemini_repository_group":              {index + "/repositoryGroups/rg1"},
		"google_gemini_repository_group_iam":          {index + "/repositoryGroups/rg1"},
		"google_gemini_data_analytics_data_agent":     {agent},
		"google_gemini_data_analytics_data_agent_iam": {agent},
	}
	settingTypes := map[string][2]string{
		"codeToolsSettings":             {"google_gemini_code_tools_setting", "google_gemini_code_tools_setting_binding"},
		"dataSharingWithGoogleSettings": {"google_gemini_data_sharing_with_google_setting", "google_gemini_data_sharing_with_google_setting_binding"},
		"geminiGcpEnablementSettings":   {"google_gemini_gemini_gcp_enablement_setting", "google_gemini_gemini_gcp_enablement_setting_binding"},
		"gdaObservabilitySettings":      {"google_gemini_gda_observability_setting", "google_gemini_gda_observability_setting_binding"},
		"gibqObservabilitySettings":     {"google_gemini_gibq_observability_setting", "google_gemini_gibq_observability_setting_binding"},
		"loggingSettings":               {"google_gemini_logging_setting", "google_gemini_logging_setting_binding"},
		"releaseChannelSettings":        {"google_gemini_release_channel_setting", "google_gemini_release_channel_setting_binding"},
	}
	for segment, types := range settingTypes {
		setting := globalLoc + "/" + segment + "/s1"
		expected[types[0]] = []string{setting}
		expected[types[1]] = []string{setting + "/settingBindings/b1"}
	}
	// Logging also exists in the configured region.
	expected["google_gemini_logging_setting"] = append(expected["google_gemini_logging_setting"], regionLoc+"/loggingSettings/s1")
	expected["google_gemini_logging_setting_binding"] = append(expected["google_gemini_logging_setting_binding"], regionLoc+"/loggingSettings/s1/settingBindings/b1")
	assertResourceIDs(t, g.Resources, expected)
}
