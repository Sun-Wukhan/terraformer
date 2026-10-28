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

func TestSaasRuntimeGeneratorEnumeration(t *testing.T) {
	parent := "projects/test-project/locations/us-central1"
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: "/saas", body: `{"saas":[{"name":"` + parent + `/saas/s1"}]}`},
		{substr: "/tenants", body: `{"tenants":[{"name":"` + parent + `/tenants/t1"}]}`},
		{substr: "/releases", body: `{"releases":[{"name":"` + parent + `/releases/r1"}]}`},
		{substr: "/rolloutKinds", body: `{"rolloutKinds":[{"name":"` + parent + `/rolloutKinds/rk1"}]}`},
		{substr: "/unitKinds", body: `{"unitKinds":[{"name":"` + parent + `/unitKinds/uk1"}]}`},
		{substr: "/units", body: `{"units":[{"name":"` + parent + `/units/u1"}]}`},
	}}
	g := &SaasRuntimeGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project", "region": compute.Region{Name: "us-central1"}})
	g.initResources(context.Background(), transport.client())
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_saas_runtime_saas":         {parent + "/saas/s1"},
		"google_saas_runtime_tenant":       {parent + "/tenants/t1"},
		"google_saas_runtime_release":      {parent + "/releases/r1"},
		"google_saas_runtime_rollout_kind": {parent + "/rolloutKinds/rk1"},
		"google_saas_runtime_unit_kind":    {parent + "/unitKinds/uk1"},
		"google_saas_runtime_unit":         {parent + "/units/u1"},
	})
}
