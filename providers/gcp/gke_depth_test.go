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
	containeranalysisv1 "google.golang.org/api/containeranalysis/v1"
	"google.golang.org/api/option"
)

// TestGkeContainerDepth feeds canned Container Analysis responses
// and multicloud raw REST attached clusters through the GKE
// generator's depth extension and verifies notes (with IAM),
// occurrences and attached clusters.
func TestGkeContainerDepth(t *testing.T) {
	proj := "projects/test-project"
	note := proj + "/notes/n1"
	loc := proj + "/locations/us-central1"
	policy := `{"bindings":[{"role":"roles/containeranalysis.notes.viewer","members":["user:a@example.com"]}]}`
	transport := &cannedTransport{routes: []cannedRoute{
		{":getIamPolicy", policy},
		{"/notes", `{"notes":[{"name":"` + note + `"}]}`},
		{"/occurrences", `{"occurrences":[{"name":"` + proj + `/occurrences/o1"}]}`},
		{"/attachedClusters", `{"attachedClusters":[{"name":"` + loc + `/attachedClusters/ac1"}]}`},
	}}
	ctx := context.Background()
	clientOpts := []option.ClientOption{option.WithHTTPClient(transport.client()), option.WithoutAuthentication()}
	caService, err := containeranalysisv1.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("containeranalysis.NewService: %v", err)
	}
	g := &GkeGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project", "region": compute.Region{Name: "us-central1"}})
	g.initContainerDepthWith(ctx, caService, transport.client())
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_container_analysis_note":             {note},
		"google_container_analysis_note_iam_policy":  {note},
		"google_container_analysis_note_iam_binding": {note},
		"google_container_analysis_note_iam_member":  {note},
		"google_container_analysis_occurrence":       {proj + "/occurrences/o1"},
		"google_container_attached_cluster":          {loc + "/attachedClusters/ac1"},
	})
}
