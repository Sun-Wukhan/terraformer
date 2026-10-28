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
)

// GeminiGenerator enumerates the Gemini for Google Cloud surface of
// one project. Neither API has a generated client in the pinned
// google.golang.org/api module, so everything runs over raw REST
// through gcp_rest.go, with endpoints and ID formats taken from the
// google provider's Magic Modules definitions:
//
//   - Cloud AI Companion (cloudaicompanion.googleapis.com/v1):
//     code repository indexes (with their repository groups and IAM)
//     and the seven per-location setting collections, each with its
//     setting bindings. Settings exist in the "global" location and
//     in regional locations; both the global location and the
//     configured region are enumerated.
//   - Gemini Data Analytics (geminidataanalytics.googleapis.com/v1):
//     data agents and their IAM, in the same two locations.
//
// The setting collections follow one shape, captured in
// geminiSettingCollections: list the settings under a location, then
// list each setting's settingBindings.
type GeminiGenerator struct {
	GCPService
}

const (
	cloudAICompanionBaseURL    = "https://cloudaicompanion.googleapis.com/v1/"
	geminiDataAnalyticsBaseURL = "https://geminidataanalytics.googleapis.com/v1/"
)

// geminiSettingCollection maps one Cloud AI Companion settings
// collection segment onto the provider's setting and setting-binding
// resource types.
type geminiSettingCollection struct {
	segment     string
	settingType string
	bindingType string
}

var geminiSettingCollections = []geminiSettingCollection{
	{"codeToolsSettings", "google_gemini_code_tools_setting", "google_gemini_code_tools_setting_binding"},
	{"dataSharingWithGoogleSettings", "google_gemini_data_sharing_with_google_setting", "google_gemini_data_sharing_with_google_setting_binding"},
	{"geminiGcpEnablementSettings", "google_gemini_gemini_gcp_enablement_setting", "google_gemini_gemini_gcp_enablement_setting_binding"},
	{"gdaObservabilitySettings", "google_gemini_gda_observability_setting", "google_gemini_gda_observability_setting_binding"},
	{"gibqObservabilitySettings", "google_gemini_gibq_observability_setting", "google_gemini_gibq_observability_setting_binding"},
	{"loggingSettings", "google_gemini_logging_setting", "google_gemini_logging_setting_binding"},
	{"releaseChannelSettings", "google_gemini_release_channel_setting", "google_gemini_release_channel_setting_binding"},
}

func (g *GeminiGenerator) InitResources() error {
	ctx := context.Background()
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		log.Println(err)
		g.RecordFailure("", err)
		return nil
	}
	g.initResources(ctx, restClient)
	return nil
}

func (g *GeminiGenerator) initResources(ctx context.Context, restClient *http.Client) {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	locations := []string{"global", region}
	for _, location := range locations {
		parent := "projects/" + project + "/locations/" + location
		g.geminiCodeRepositoryIndexes(ctx, restClient, parent)
		for _, collection := range geminiSettingCollections {
			g.geminiSettings(ctx, restClient, parent, collection)
		}
		g.geminiDataAgents(ctx, restClient, parent)
	}
}

func (g *GeminiGenerator) newGeminiResource(id, resourceType string) terraformutils.Resource {
	parts := strings.Split(id, "/")
	return terraformutils.NewResource(
		id,
		parts[len(parts)-1],
		resourceType,
		g.ProviderName,
		map[string]string{},
		[]string{},
		map[string]interface{}{},
	)
}

// geminiCodeRepositoryIndexes lists code repository indexes under
// one location, their repository groups, and each repository
// group's combined IAM resource.
func (g *GeminiGenerator) geminiCodeRepositoryIndexes(ctx context.Context, restClient *http.Client, parent string) {
	indexes, err := restListResources(ctx, restClient, cloudAICompanionBaseURL+parent+"/codeRepositoryIndexes")
	if err != nil {
		log.Println(err)
		return
	}
	for _, index := range indexes {
		g.Resources = append(g.Resources, g.newGeminiResource(index.Name, "google_gemini_code_repository_index"))
		groups, err := restListResources(ctx, restClient, cloudAICompanionBaseURL+index.Name+"/repositoryGroups")
		if err != nil {
			log.Println(err)
		}
		for _, group := range groups {
			g.Resources = append(g.Resources, g.newGeminiResource(group.Name, "google_gemini_repository_group"))
			// The provider models repository group IAM as the
			// single combined _iam resource, emitted when the
			// group carries bindings.
			bindings, err := restGetIamPolicy(ctx, restClient, cloudAICompanionBaseURL+group.Name+":getIamPolicy")
			if err != nil {
				log.Println(err)
				continue
			}
			if len(bindings) > 0 {
				g.Resources = append(g.Resources, g.newGeminiResource(group.Name, "google_gemini_repository_group_iam"))
			}
		}
	}
}

// geminiSettings lists one settings collection under a location and
// each setting's bindings.
func (g *GeminiGenerator) geminiSettings(ctx context.Context, restClient *http.Client, parent string, collection geminiSettingCollection) {
	settings, err := restListResources(ctx, restClient, cloudAICompanionBaseURL+parent+"/"+collection.segment)
	if err != nil {
		log.Println(err)
		return
	}
	for _, setting := range settings {
		g.Resources = append(g.Resources, g.newGeminiResource(setting.Name, collection.settingType))
		bindings, err := restListResources(ctx, restClient, cloudAICompanionBaseURL+setting.Name+"/settingBindings")
		if err != nil {
			log.Println(err)
			continue
		}
		for _, binding := range bindings {
			g.Resources = append(g.Resources, g.newGeminiResource(binding.Name, collection.bindingType))
		}
	}
}

// geminiDataAgents lists Gemini Data Analytics data agents under
// one location and each agent's combined IAM resource.
func (g *GeminiGenerator) geminiDataAgents(ctx context.Context, restClient *http.Client, parent string) {
	agents, err := restListResources(ctx, restClient, geminiDataAnalyticsBaseURL+parent+"/dataAgents")
	if err != nil {
		log.Println(err)
		return
	}
	for _, agent := range agents {
		g.Resources = append(g.Resources, g.newGeminiResource(agent.Name, "google_gemini_data_analytics_data_agent"))
		// The provider models data agent IAM as the single
		// combined _iam resource, emitted when the agent carries
		// bindings.
		bindings, err := restGetIamPolicy(ctx, restClient, geminiDataAnalyticsBaseURL+agent.Name+":getIamPolicy")
		if err != nil {
			log.Println(err)
			continue
		}
		if len(bindings) > 0 {
			g.Resources = append(g.Resources, g.newGeminiResource(agent.Name, "google_gemini_data_analytics_data_agent_iam"))
		}
	}
}
