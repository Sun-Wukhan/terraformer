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

	"google.golang.org/api/option"
	runtimeconfig "google.golang.org/api/runtimeconfig/v1beta1"
)

func TestRuntimeConfigEnumeration(t *testing.T) {
	ctx := context.Background()
	cfg := "projects/test-project/configs/cfg1"
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: ":getIamPolicy", body: `{"bindings":[{"role":"roles/runtimeconfig.configViewer","members":["user:a@example.com"]}]}`},
		{substr: "/variables", body: `{"variables":[{"name":"` + cfg + `/variables/v1"}]}`},
		{substr: "/configs", body: `{"configs":[{"name":"` + cfg + `"}]}`},
	}}
	service, err := runtimeconfig.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	g := &RuntimeConfigGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project"})
	g.initResources(ctx, service)
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_runtimeconfig_config":     {cfg},
		"google_runtimeconfig_config_iam": {cfg},
		"google_runtimeconfig_variable":   {cfg + "/variables/v1"},
	})
}
