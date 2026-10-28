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

	binauthzv1 "google.golang.org/api/binaryauthorization/v1"
	"google.golang.org/api/option"
)

// TestBinaryAuthorizationGenerator feeds canned Binary Authorization
// responses through the generator and verifies the emitted types and
// IDs: the project policy singleton and the attestor with its IAM
// trio.
func TestBinaryAuthorizationGenerator(t *testing.T) {
	attestor := "projects/test-project/attestors/att-1"
	transport := &cannedTransport{routes: []cannedRoute{
		{":getIamPolicy", `{"bindings":[{"role":"roles/binaryauthorization.attestorsViewer","members":["user:a@example.com"]}]}`},
		{"/attestors", `{"attestors":[{"name":"` + attestor + `"}]}`},
		{"/v1/projects/test-project", `{"name":"projects/test-project/policy"}`},
	}}

	ctx := context.Background()
	service, err := binauthzv1.NewService(ctx,
		option.WithHTTPClient(transport.client()),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatalf("binaryauthorization.NewService: %v", err)
	}

	g := &BinaryAuthorizationGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project"})
	if err := g.initResources(ctx, service); err != nil {
		t.Fatalf("initResources: %v", err)
	}

	expected := map[string][]string{
		"google_binary_authorization_policy":               {"test-project"},
		"google_binary_authorization_attestor":             {attestor},
		"google_binary_authorization_attestor_iam_policy":  {attestor},
		"google_binary_authorization_attestor_iam_binding": {attestor},
		"google_binary_authorization_attestor_iam_member":  {attestor},
	}
	assertResourceIDs(t, g.Resources, expected)
}
