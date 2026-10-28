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
	eventarc "google.golang.org/api/eventarc/v1"
	"google.golang.org/api/option"
)

// TestEventarcEnumeration asserts every Eventarc collection plus the
// pipeline combined IAM resource and the channel config singleton.
func TestEventarcEnumeration(t *testing.T) {
	ctx := context.Background()
	parent := "projects/test-project/locations/us-central1"
	pipeline := parent + "/pipelines/p1"
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: ":getIamPolicy", body: `{"bindings":[{"role":"roles/eventarc.publisher","members":["user:a@example.com"]}]}`},
		{substr: "/channels", body: `{"channels":[{"name":"` + parent + `/channels/c1"}]}`},
		{substr: "/enrollments", body: `{"enrollments":[{"name":"` + parent + `/enrollments/e1"}]}`},
		{substr: "/googleApiSources", body: `{"googleApiSources":[{"name":"` + parent + `/googleApiSources/g1"}]}`},
		{substr: "/messageBuses", body: `{"messageBuses":[{"name":"` + parent + `/messageBuses/m1"}]}`},
		{substr: "/pipelines", body: `{"pipelines":[{"name":"` + pipeline + `"}]}`},
		{substr: "/triggers", body: `{"triggers":[{"name":"` + parent + `/triggers/t1"}]}`},
		{substr: "/googleChannelConfig", body: `{}`},
	}}
	service, err := eventarc.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	g := &EventarcGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	g.initResources(ctx, service)
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_eventarc_channel":               {parent + "/channels/c1"},
		"google_eventarc_enrollment":            {parent + "/enrollments/e1"},
		"google_eventarc_google_api_source":     {parent + "/googleApiSources/g1"},
		"google_eventarc_message_bus":           {parent + "/messageBuses/m1"},
		"google_eventarc_pipeline":              {pipeline},
		"google_eventarc_pipeline_iam":          {pipeline},
		"google_eventarc_trigger":               {parent + "/triggers/t1"},
		"google_eventarc_google_channel_config": {parent},
	})
}
