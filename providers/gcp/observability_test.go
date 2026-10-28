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

// TestObservabilityEnumeration asserts buckets, their links, trace
// scopes and the project settings singleton are emitted (project
// scope only: no CRM service is supplied, so no folder/org scopes).
func TestObservabilityEnumeration(t *testing.T) {
	proj := "projects/test-project"
	loc := proj + "/locations/us-central1"
	bucket := loc + "/buckets/b1"
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: "/links", body: `{"links":[{"name":"` + bucket + `/links/l1"}]}`},
		{substr: "/buckets", body: `{"buckets":[{"name":"` + bucket + `"}]}`},
		{substr: "/traceScopes", body: `{"traceScopes":[{"name":"` + loc + `/traceScopes/ts1"}]}`},
		{substr: "/settings", body: `{}`},
	}}
	g := &ObservabilityGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	g.initResources(context.Background(), transport.client(), nil)
	// The canned transport answers both the regional and the global
	// location walks with the same body; assertResourceIDs compares
	// ID sets, so each ID is listed once.
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_observability_bucket":           {bucket},
		"google_observability_link":             {bucket + "/links/l1"},
		"google_observability_trace_scope":      {loc + "/traceScopes/ts1"},
		"google_observability_project_settings": {"projects/test-project"},
	})
}
