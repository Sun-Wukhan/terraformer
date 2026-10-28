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

	"google.golang.org/api/cloudfunctions/v1"
	cfv2 "google.golang.org/api/cloudfunctions/v2"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

func TestCloudFunctionsIamGeneratorEnumeration(t *testing.T) {
	ctx := context.Background()
	parent := "projects/test-project/locations/us-central1"
	gen1 := parent + "/functions/fn1"
	gen2 := parent + "/functions/fn2"
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: ":getIamPolicy", body: `{"bindings":[{"role":"roles/cloudfunctions.invoker","members":["allUsers"]}]}`},
		{substr: "/functions", body: `{"functions":[{"name":"` + gen1 + `","environment":"GEN_1"},{"name":"` + gen2 + `","environment":"GEN_2"}]}`},
	}}
	v1Svc, err := cloudfunctions.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("v1 NewService: %v", err)
	}
	v2Svc, err := cfv2.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("v2 NewService: %v", err)
	}
	g := &CloudFunctionsIamGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project", "region": compute.Region{Name: "us-central1"}})
	if err := g.initResources(ctx, v1Svc, v2Svc); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_cloudfunctions_function_iam_policy":   {"test-project/us-central1/fn1"},
		"google_cloudfunctions_function_iam_binding":  {"test-project/us-central1/fn1"},
		"google_cloudfunctions_function_iam_member":   {"test-project/us-central1/fn1"},
		"google_cloudfunctions2_function_iam_policy":  {gen2},
		"google_cloudfunctions2_function_iam_binding": {gen2},
		"google_cloudfunctions2_function_iam_member":  {gen2},
	})
}
