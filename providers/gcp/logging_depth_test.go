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

	"google.golang.org/api/logging/v2"
	"google.golang.org/api/option"
)

// TestLoggingDepthEnumeration feeds canned folder and organization
// Logging responses and asserts the folder/organization sink,
// exclusion, bucket config and settings resources with their
// provider IDs.
func TestLoggingDepthEnumeration(t *testing.T) {
	ctx := context.Background()
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: "/sinks", body: `{"sinks":[{"name":"sink1"}]}`},
		{substr: "/exclusions", body: `{"exclusions":[{"name":"excl1"}]}`},
		{substr: "folders/123/locations", body: `{"buckets":[{"name":"folders/123/locations/global/buckets/b1"}]}`},
		{substr: "organizations/456/locations", body: `{"buckets":[{"name":"organizations/456/locations/global/buckets/b2"}]}`},
		{substr: "/settings", body: `{}`},
	}}
	service, err := logging.NewService(ctx,
		option.WithHTTPClient(transport.client()),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	g := &LoggingGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project"})
	if err := g.loadLoggingScope(ctx, service, sccScope{kind: "folder", parent: "folders/123", id: "123"}); err != nil {
		t.Fatalf("folder scope: %v", err)
	}
	if err := g.loadLoggingScope(ctx, service, sccScope{kind: "organization", parent: "organizations/456", id: "456"}); err != nil {
		t.Fatalf("org scope: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_logging_folder_sink":                {"folders/123/sinks/sink1"},
		"google_logging_folder_exclusion":           {"folders/123/exclusions/excl1"},
		"google_logging_folder_bucket_config":       {"folders/123/locations/global/buckets/b1"},
		"google_logging_folder_settings":            {"folders/123"},
		"google_logging_organization_sink":          {"organizations/456/sinks/sink1"},
		"google_logging_organization_exclusion":     {"organizations/456/exclusions/excl1"},
		"google_logging_organization_bucket_config": {"organizations/456/locations/global/buckets/b2"},
		"google_logging_organization_settings":      {"organizations/456"},
	})
}
