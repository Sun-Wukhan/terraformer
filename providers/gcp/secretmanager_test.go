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
	"google.golang.org/api/option"
	secretmanager "google.golang.org/api/secretmanager/v1"
)

func TestSecretManagerGeneratorEnumeration(t *testing.T) {
	ctx := context.Background()
	secretName := "projects/test-project/secrets/db-pass"
	regionalName := "projects/test-project/locations/us-central1/secrets/api-key"
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: ":getIamPolicy", body: `{"bindings":[{"role":"roles/secretmanager.admin","members":["user:a@example.com"]}]}`},
		{substr: "/locations/us-central1/secrets/api-key/versions", body: `{"versions":[{"name":"` + regionalName + `/versions/1"}]}`},
		{substr: "/locations/us-central1/secrets", body: `{"secrets":[{"name":"` + regionalName + `"}]}`},
		{substr: "/versions", body: `{"versions":[{"name":"` + secretName + `/versions/1"}]}`},
		{substr: "/secrets", body: `{"secrets":[{"name":"` + secretName + `"}]}`},
	}}
	svc, err := secretmanager.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	g := &SecretManagerGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project", "region": compute.Region{Name: "us-central1"}})
	if err := g.initResources(ctx, svc); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_secret_manager_secret":                      {secretName},
		"google_secret_manager_secret_iam_policy":           {secretName},
		"google_secret_manager_secret_iam_binding":          {secretName},
		"google_secret_manager_secret_iam_member":           {secretName},
		"google_secret_manager_secret_version":              {secretName + "/versions/1"},
		"google_secret_manager_regional_secret":             {regionalName},
		"google_secret_manager_regional_secret_iam_policy":  {regionalName},
		"google_secret_manager_regional_secret_iam_binding": {regionalName},
		"google_secret_manager_regional_secret_iam_member":  {regionalName},
		"google_secret_manager_regional_secret_version":     {regionalName + "/versions/1"},
	})
}
