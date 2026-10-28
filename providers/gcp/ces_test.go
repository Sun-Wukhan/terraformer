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

// TestCesGenerator feeds canned CES v1beta responses and verifies
// the whole family: the security settings singleton, an app with a
// root agent (hence its root-agent association), and one item in
// each per-app collection.
func TestCesGenerator(t *testing.T) {
	parent := "projects/test-project/locations/us"
	app := parent + "/apps/app1"
	agent := app + "/agents/agent1"
	transport := &cannedTransport{routes: []cannedRoute{
		{"/locations/us/securitySettings", `{"name":"` + parent + `/securitySettings"}`},
		{"/apps/app1/agents", `{"agents":[{"name":"` + agent + `"}]}`},
		{"/apps/app1/deployments", `{"deployments":[{"name":"` + app + `/deployments/dep1"}]}`},
		{"/apps/app1/evaluations", `{"evaluations":[{"name":"` + app + `/evaluations/ev1"}]}`},
		{"/apps/app1/examples", `{"examples":[{"name":"` + app + `/examples/ex1"}]}`},
		{"/apps/app1/guardrails", `{"guardrails":[{"name":"` + app + `/guardrails/gr1"}]}`},
		{"/apps/app1/toolsets", `{"toolsets":[{"name":"` + app + `/toolsets/ts1"}]}`},
		{"/apps/app1/tools", `{"tools":[{"name":"` + app + `/tools/tool1"}]}`},
		{"/apps/app1/versions", `{"versions":[{"name":"` + app + `/versions/v1"}]}`},
		{"/locations/us/apps", `{"apps":[{"name":"` + app + `","rootAgent":"` + agent + `"}]}`},
	}}
	g := &CesGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us"},
	})
	g.initResources(context.Background(), transport.client())
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_ces_security_settings":          {parent + "/securitySettings"},
		"google_ces_app":                        {app},
		"google_ces_app_root_agent_association": {app},
		"google_ces_agent":                      {agent},
		"google_ces_deployment":                 {app + "/deployments/dep1"},
		"google_ces_evaluation":                 {app + "/evaluations/ev1"},
		"google_ces_example":                    {app + "/examples/ex1"},
		"google_ces_guardrail":                  {app + "/guardrails/gr1"},
		"google_ces_tool":                       {app + "/tools/tool1"},
		"google_ces_toolset":                    {app + "/toolsets/ts1"},
		"google_ces_app_version":                {app + "/versions/v1"},
	})
}
