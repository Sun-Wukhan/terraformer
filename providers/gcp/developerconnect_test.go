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
	developerconnect "google.golang.org/api/developerconnect/v1"
	"google.golang.org/api/option"
)

func TestDeveloperConnectGeneratorEnumeration(t *testing.T) {
	ctx := context.Background()
	parent := "projects/test-project/locations/us-central1"
	conn := parent + "/connections/conn1"
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: "/accountConnectors", body: `{"accountConnectors":[{"name":"` + parent + `/accountConnectors/ac1"}]}`},
		{substr: "/gitRepositoryLinks", body: `{"gitRepositoryLinks":[{"name":"` + conn + `/gitRepositoryLinks/link1"}]}`},
		{substr: "/connections", body: `{"connections":[{"name":"` + conn + `"}]}`},
	}}
	svc, err := developerconnect.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	g := &DeveloperConnectGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project", "region": compute.Region{Name: "us-central1"}})
	if err := g.initResources(ctx, svc); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_developer_connect_account_connector":   {parent + "/accountConnectors/ac1"},
		"google_developer_connect_connection":          {conn},
		"google_developer_connect_git_repository_link": {conn + "/gitRepositoryLinks/link1"},
	})
}
