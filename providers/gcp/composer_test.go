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

	composer "google.golang.org/api/composer/v1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

func TestComposerEnumeration(t *testing.T) {
	ctx := context.Background()
	parent := "projects/test-project/locations/us-central1"
	env := parent + "/environments/env1"
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: "/userWorkloadsConfigMaps", body: `{"userWorkloadsConfigMaps":[{"name":"` + env + `/userWorkloadsConfigMaps/cm1"}]}`},
		{substr: "/userWorkloadsSecrets", body: `{"userWorkloadsSecrets":[{"name":"` + env + `/userWorkloadsSecrets/s1"}]}`},
		{substr: "/environments", body: `{"environments":[{"name":"` + env + `"}]}`},
	}}
	service, err := composer.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	g := &ComposerGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	g.initResources(ctx, service)
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_composer_environment":               {env},
		"google_composer_user_workloads_config_map": {env + "/userWorkloadsConfigMaps/cm1"},
		"google_composer_user_workloads_secret":     {env + "/userWorkloadsSecrets/s1"},
	})
}
