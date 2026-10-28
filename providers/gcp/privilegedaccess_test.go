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

func TestPrivilegedAccessGeneratorEnumeration(t *testing.T) {
	parent := "projects/test-project/locations/us-central1"
	ent := parent + "/entitlements/ent1"
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: "/entitlements", body: `{"entitlements":[{"name":"` + ent + `"}]}`},
	}}
	g := &PrivilegedAccessGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project", "region": compute.Region{Name: "us-central1"}})
	g.initResources(context.Background(), transport.client())
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_privileged_access_manager_entitlement": {ent},
	})
}
