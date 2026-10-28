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

func TestApiHubGeneratorEnumeration(t *testing.T) {
	parent := "projects/test-project/locations/us-central1"
	plugin := parent + "/plugins/plug1"
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: "/apiHubInstances", body: `{"apiHubInstances":[{"name":"` + parent + `/apiHubInstances/inst1"}]}`},
		{substr: "/curations", body: `{"curations":[{"name":"` + parent + `/curations/cur1"}]}`},
		{substr: "/hostProjectRegistrations", body: `{"hostProjectRegistrations":[{"name":"` + parent + `/hostProjectRegistrations/reg1"}]}`},
		{substr: "/runtimeProjectAttachments", body: `{"runtimeProjectAttachments":[{"name":"` + parent + `/runtimeProjectAttachments/att1"}]}`},
		{substr: "/instances", body: `{"pluginInstances":[{"name":"` + plugin + `/instances/pi1"}]}`},
		{substr: "/plugins", body: `{"plugins":[{"name":"` + plugin + `"}]}`},
	}}
	g := &ApiHubGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project", "region": compute.Region{Name: "us-central1"}})
	g.initResources(context.Background(), transport.client())
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_apihub_api_hub_instance":           {parent + "/apiHubInstances/inst1"},
		"google_apihub_curation":                   {parent + "/curations/cur1"},
		"google_apihub_host_project_registration":  {parent + "/hostProjectRegistrations/reg1"},
		"google_apihub_runtime_project_attachment": {parent + "/runtimeProjectAttachments/att1"},
		"google_apihub_plugin":                     {plugin},
		"google_apihub_plugin_instance":            {plugin + "/instances/pi1"},
	})
}
