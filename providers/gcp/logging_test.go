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

func TestLoggingGeneratorEnumeration(t *testing.T) {
	ctx := context.Background()
	bucketName := "projects/test-project/locations/global/buckets/my-bucket"
	viewName := bucketName + "/views/my-view"
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: ":getIamPolicy", body: `{"bindings":[{"role":"roles/logging.viewAccessor","members":["user:jane@example.com"]}]}`},
		{substr: "/links", body: `{"links":[{"name":"` + bucketName + `/links/link1","bigqueryDataset":{"datasetId":"projects/test-project/datasets/ds1"}}]}`},
		{substr: "/views", body: `{"views":[{"name":"` + viewName + `"}]}`},
		{substr: "/buckets", body: `{"buckets":[{"name":"` + bucketName + `"}]}`},
		{substr: "/sinks", body: `{"sinks":[{"name":"audit-sink"}]}`},
		{substr: "/exclusions", body: `{"exclusions":[{"name":"excl1"}]}`},
		{substr: "/logScopes", body: `{"logScopes":[{"name":"projects/test-project/locations/global/logScopes/my-scope"}]}`},
		{substr: "/savedQueries", body: `{"savedQueries":[{"name":"projects/test-project/locations/global/savedQueries/sq1"}]}`},
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
	if err := g.initResources(ctx, service); err != nil {
		t.Fatalf("initResources: %v", err)
	}

	got := map[string][]string{}
	for _, r := range g.Resources {
		got[r.InstanceInfo.Type] = append(got[r.InstanceInfo.Type], r.InstanceState.ID)
	}

	wantIDs := map[string]string{
		"google_logging_project_sink":          "projects/test-project/sinks/audit-sink",
		"google_logging_project_exclusion":     "projects/test-project/exclusions/excl1",
		"google_logging_project_bucket_config": bucketName,
		"google_logging_linked_dataset":        bucketName + "/links/link1",
		"google_logging_log_view":              viewName,
		"google_logging_log_view_iam_policy":   viewName,
		"google_logging_log_view_iam_binding":  viewName,
		"google_logging_log_view_iam_member":   viewName,
		"google_logging_log_scope":             "projects/test-project/locations/global/logScopes/my-scope",
		"google_logging_saved_query":           "projects/test-project/locations/global/savedQueries/sq1",
	}
	for resourceType, wantID := range wantIDs {
		ids := got[resourceType]
		if len(ids) != 1 || ids[0] != wantID {
			t.Errorf("%s: got IDs %v, want [%s]", resourceType, ids, wantID)
		}
	}
}
