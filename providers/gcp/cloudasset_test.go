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

	cloudasset "google.golang.org/api/cloudasset/v1"
	"google.golang.org/api/option"
)

// TestCloudAssetEnumeration asserts project feeds are emitted (no
// CRM service is supplied, so folder/org scopes are not walked).
func TestCloudAssetEnumeration(t *testing.T) {
	ctx := context.Background()
	feed := "projects/test-project/feeds/feed1"
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: "/feeds", body: `{"feeds":[{"name":"` + feed + `"}]}`},
	}}
	service, err := cloudasset.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	g := &CloudAssetGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project"})
	g.initResources(ctx, service, nil)
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_cloud_asset_project_feed": {feed},
	})
}
