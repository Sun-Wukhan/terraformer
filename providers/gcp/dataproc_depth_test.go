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
	dataprocv1 "google.golang.org/api/dataproc/v1"
	"google.golang.org/api/option"
)

// TestDataprocDepthGenerator feeds canned Dataproc responses and
// verifies autoscaling policies and workflow templates expand their
// IAM trios, cluster IAM is expanded, and batches and session
// templates are emitted with full resource names as IDs.
func TestDataprocDepthGenerator(t *testing.T) {
	regionParent := "projects/test-project/regions/us-central1"
	locationParent := "projects/test-project/locations/us-central1"
	policy := regionParent + "/autoscalingPolicies/pol1"
	template := regionParent + "/workflowTemplates/wt1"
	cluster := regionParent + "/clusters/cl1"
	transport := &cannedTransport{routes: []cannedRoute{
		{":getIamPolicy", `{"bindings":[{"role":"roles/dataproc.editor","members":["user:a@example.com"]}]}`},
		{"/autoscalingPolicies", `{"policies":[{"name":"` + policy + `"}]}`},
		{"/workflowTemplates", `{"templates":[{"name":"` + template + `"}]}`},
		{"/regions/us-central1/clusters", `{"clusters":[{"clusterName":"cl1"}]}`},
		{"/locations/us-central1/batches", `{"batches":[{"name":"` + locationParent + `/batches/b1"}]}`},
		{"/locations/us-central1/sessionTemplates", `{"sessionTemplates":[{"name":"` + locationParent + `/sessionTemplates/st1"}]}`},
	}}
	ctx := context.Background()
	svc, err := dataprocv1.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("dataproc.NewService: %v", err)
	}
	g := &DataprocGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	g.initDepthResources(ctx, svc)
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_dataproc_autoscaling_policy":             {policy},
		"google_dataproc_autoscaling_policy_iam_policy":  {policy},
		"google_dataproc_autoscaling_policy_iam_binding": {policy},
		"google_dataproc_autoscaling_policy_iam_member":  {policy},
		"google_dataproc_workflow_template":              {template},
		"google_dataproc_cluster_iam_policy":             {cluster},
		"google_dataproc_cluster_iam_binding":            {cluster},
		"google_dataproc_cluster_iam_member":             {cluster},
		"google_dataproc_batch":                          {locationParent + "/batches/b1"},
		"google_dataproc_session_template":               {locationParent + "/sessionTemplates/st1"},
	})
}
