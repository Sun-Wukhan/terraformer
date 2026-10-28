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

func TestParameterManagerGeneratorEnumeration(t *testing.T) {
	globalParent := "projects/test-project/locations/global"
	regionalParent := "projects/test-project/locations/us-central1"
	gp := globalParent + "/parameters/p1"
	rp := regionalParent + "/parameters/p2"
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: "/locations/global/parameters/p1/versions", body: `{"parameterVersions":[{"name":"` + gp + `/versions/v1"}]}`},
		{substr: "/locations/us-central1/parameters/p2/versions", body: `{"parameterVersions":[{"name":"` + rp + `/versions/v1"}]}`},
		{substr: "/locations/global/parameters", body: `{"parameters":[{"name":"` + gp + `"}]}`},
		{substr: "/locations/global/templates", body: `{"templates":[{"name":"` + globalParent + `/templates/t1"}]}`},
		{substr: "/locations/us-central1/parameters", body: `{"parameters":[{"name":"` + rp + `"}]}`},
	}}
	g := &ParameterManagerGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project", "region": compute.Region{Name: "us-central1"}})
	g.initResources(context.Background(), transport.client())
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_parameter_manager_parameter":                  {gp},
		"google_parameter_manager_parameter_version":          {gp + "/versions/v1"},
		"google_parameter_manager_template":                   {globalParent + "/templates/t1"},
		"google_parameter_manager_regional_parameter":         {rp},
		"google_parameter_manager_regional_parameter_version": {rp + "/versions/v1"},
	})
}
