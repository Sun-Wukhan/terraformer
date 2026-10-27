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
	runv1 "google.golang.org/api/run/v1"
	runv2 "google.golang.org/api/run/v2"
)

func TestCloudRunGeneratorEnumeration(t *testing.T) {
	ctx := context.Background()
	parent := "projects/test-project/locations/us-central1"
	v1SvcName := "locations/us-central1/namespaces/test-project/services/api-v1"
	v2SvcName := parent + "/services/api-v2"
	jobName := parent + "/jobs/nightly"
	iamBody := `{"bindings":[{"role":"roles/run.invoker","members":["allUsers"]}]}`

	v1Transport := &cannedTransport{routes: []cannedRoute{
		{substr: ":getIamPolicy", body: iamBody},
		{substr: "/domainmappings", body: `{"items":[{"metadata":{"name":"example.com","namespace":"test-project"}}]}`},
		{substr: "/services", body: `{"items":[{"metadata":{"name":"api-v1","namespace":"test-project"}}]}`},
	}}
	v2Transport := &cannedTransport{routes: []cannedRoute{
		{substr: ":getIamPolicy", body: iamBody},
		{substr: "/jobs", body: `{"jobs":[{"name":"` + jobName + `"}]}`},
		{substr: "/workerPools", body: `{"workerPools":[{"name":"` + parent + `/workerPools/pool1"}]}`},
		{substr: "/services", body: `{"services":[{"name":"` + v2SvcName + `"}]}`},
	}}

	v1Service, err := runv1.NewService(ctx,
		option.WithHTTPClient(v1Transport.client()),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatalf("run v1 NewService: %v", err)
	}
	v2Service, err := runv2.NewService(ctx,
		option.WithHTTPClient(v2Transport.client()),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatalf("run v2 NewService: %v", err)
	}

	g := &CloudRunGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	if err := g.initResources(ctx, v1Service, v2Service); err != nil {
		t.Fatalf("initResources: %v", err)
	}

	got := map[string][]string{}
	for _, r := range g.Resources {
		got[r.InstanceInfo.Type] = append(got[r.InstanceInfo.Type], r.InstanceState.ID)
	}

	wantIDs := map[string]string{
		"google_cloud_run_service":            v1SvcName,
		"google_cloud_run_service_iam_policy": v1SvcName,
		"google_cloud_run_domain_mapping":     "locations/us-central1/namespaces/test-project/domainmappings/example.com",
		"google_cloud_run_v2_service":         v2SvcName,
		"google_cloud_run_v2_job":             jobName,
		"google_cloud_run_v2_job_iam_policy":  jobName,
		"google_cloud_run_v2_worker_pool":     parent + "/workerPools/pool1",
	}
	for resourceType, wantID := range wantIDs {
		ids := got[resourceType]
		if len(ids) != 1 || ids[0] != wantID {
			t.Errorf("%s: got IDs %v, want [%s]", resourceType, ids, wantID)
		}
	}
	// One binding with one member in each canned policy.
	for _, iamType := range []string{
		"google_cloud_run_service_iam_binding",
		"google_cloud_run_service_iam_member",
		"google_cloud_run_v2_service_iam_policy",
		"google_cloud_run_v2_service_iam_binding",
		"google_cloud_run_v2_service_iam_member",
		"google_cloud_run_v2_job_iam_binding",
		"google_cloud_run_v2_job_iam_member",
	} {
		if len(got[iamType]) != 1 {
			t.Errorf("%s: got %d resources, want 1", iamType, len(got[iamType]))
		}
	}
}
