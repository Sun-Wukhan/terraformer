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

	cloudbuildv1 "google.golang.org/api/cloudbuild/v1"
	cloudbuildv2 "google.golang.org/api/cloudbuild/v2"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

func TestCloudBuildDepthEnumeration(t *testing.T) {
	ctx := context.Background()
	parent := "projects/test-project/locations/us-central1"
	conn := parent + "/connections/conn1"
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: ":getIamPolicy", body: `{"bindings":[{"role":"roles/cloudbuild.connectionAdmin","members":["user:a@example.com"]}]}`},
		{substr: "/workerPools", body: `{"workerPools":[{"name":"` + parent + `/workerPools/wp1"}]}`},
		{substr: "/bitbucketServerConfigs", body: `{"bitbucketServerConfigs":[{"name":"` + parent + `/bitbucketServerConfigs/bb1"}]}`},
		{substr: "/repositories", body: `{"repositories":[{"name":"` + conn + `/repositories/repo1"}]}`},
		{substr: "/connections", body: `{"connections":[{"name":"` + conn + `"}]}`},
	}}
	v1Service, err := cloudbuildv1.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("v1 NewService: %v", err)
	}
	v2Service, err := cloudbuildv2.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("v2 NewService: %v", err)
	}
	g := &CloudBuildDepthGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	g.initResources(ctx, v1Service, v2Service)
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_cloudbuild_worker_pool":             {parent + "/workerPools/wp1"},
		"google_cloudbuild_bitbucket_server_config": {parent + "/bitbucketServerConfigs/bb1"},
		"google_cloudbuildv2_connection":            {conn},
		"google_cloudbuildv2_connection_iam":        {conn},
		"google_cloudbuildv2_repository":            {conn + "/repositories/repo1"},
	})
}
