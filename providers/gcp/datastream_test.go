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
	datastreamv1 "google.golang.org/api/datastream/v1"
	"google.golang.org/api/option"
)

// TestDatastreamGenerator feeds canned Datastream responses and
// verifies connection profiles, private connections and streams with
// full resource names as IDs.
func TestDatastreamGenerator(t *testing.T) {
	loc := "projects/test-project/locations/us-central1"
	transport := &cannedTransport{routes: []cannedRoute{
		{"/locations/us-central1/connectionProfiles", `{"connectionProfiles":[{"name":"` + loc + `/connectionProfiles/cp1"}]}`},
		{"/locations/us-central1/privateConnections", `{"privateConnections":[{"name":"` + loc + `/privateConnections/pc1"}]}`},
		{"/locations/us-central1/streams", `{"streams":[{"name":"` + loc + `/streams/st1"}]}`},
	}}
	ctx := context.Background()
	svc, err := datastreamv1.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("datastream.NewService: %v", err)
	}
	g := &DatastreamGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	if err := g.initResources(ctx, svc); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_datastream_connection_profile": {loc + "/connectionProfiles/cp1"},
		"google_datastream_private_connection": {loc + "/privateConnections/pc1"},
		"google_datastream_stream":             {loc + "/streams/st1"},
	})
}
