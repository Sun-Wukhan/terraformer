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

	"google.golang.org/api/aiplatform/v1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

// TestColabGenerator feeds canned aiplatform v1 responses and
// verifies the Colab Enterprise family: a runtime template with its
// IAM trio, a runtime, a notebook execution job, and a schedule —
// plus the negative case that a pipeline-job schedule is NOT
// emitted as a Colab schedule.
func TestColabGenerator(t *testing.T) {
	parent := "projects/test-project/locations/us-central1"
	template := parent + "/notebookRuntimeTemplates/tpl1"
	transport := &cannedTransport{routes: []cannedRoute{
		{":getIamPolicy", `{"bindings":[{"role":"roles/aiplatform.colabEnterpriseAdmin","members":["user:jane@example.com"]}]}`},
		{"/notebookRuntimeTemplates", `{"notebookRuntimeTemplates":[{"name":"` + template + `"}]}`},
		{"/notebookRuntimes", `{"notebookRuntimes":[{"name":"` + parent + `/notebookRuntimes/rt1"}]}`},
		{"/notebookExecutionJobs", `{"notebookExecutionJobs":[{"name":"` + parent + `/notebookExecutionJobs/job1"}]}`},
		{"/schedules", `{"schedules":[{"name":"` + parent + `/schedules/sched1","createNotebookExecutionJobRequest":{}},{"name":"` + parent + `/schedules/sched2","createPipelineJobRequest":{}}]}`},
	}}
	ctx := context.Background()
	service, err := aiplatform.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("aiplatform NewService: %v", err)
	}
	g := &ColabGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	g.initResources(ctx, service)
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_colab_runtime_template":             {template},
		"google_colab_runtime_template_iam_policy":  {template},
		"google_colab_runtime_template_iam_binding": {template},
		"google_colab_runtime_template_iam_member":  {template},
		"google_colab_runtime":                      {parent + "/notebookRuntimes/rt1"},
		"google_colab_notebook_execution":           {parent + "/notebookExecutionJobs/job1"},
		"google_colab_schedule":                     {parent + "/schedules/sched1"},
	})
}
