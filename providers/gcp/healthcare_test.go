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
	healthcare "google.golang.org/api/healthcare/v1"
	"google.golang.org/api/option"
)

// TestHealthcareEnumeration asserts datasets, all four store kinds,
// their combined IAM resources, and the raw-REST workspaces and
// pipeline jobs.
func TestHealthcareEnumeration(t *testing.T) {
	ctx := context.Background()
	dataset := "projects/test-project/locations/us-central1/datasets/ds1"
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: ":getIamPolicy", body: `{"bindings":[{"role":"roles/healthcare.datasetAdmin","members":["user:a@example.com"]}]}`},
		{substr: "/consentStores", body: `{"consentStores":[{"name":"` + dataset + `/consentStores/cs1"}]}`},
		{substr: "/dicomStores", body: `{"dicomStores":[{"name":"` + dataset + `/dicomStores/dc1"}]}`},
		{substr: "/fhirStores", body: `{"fhirStores":[{"name":"` + dataset + `/fhirStores/f1"}]}`},
		{substr: "/hl7V2Stores", body: `{"hl7V2Stores":[{"name":"` + dataset + `/hl7V2Stores/h1"}]}`},
		{substr: "/workspaces", body: `{"workspaces":[{"name":"` + dataset + `/workspaces/w1"}]}`},
		{substr: "/pipelineJobs", body: `{"pipelineJobs":[{"name":"` + dataset + `/pipelineJobs/pj1"}]}`},
		{substr: "/datasets", body: `{"datasets":[{"name":"` + dataset + `"}]}`},
	}}
	service, err := healthcare.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	g := &HealthcareGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	g.initResources(ctx, service, transport.client())
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_healthcare_dataset":           {dataset},
		"google_healthcare_dataset_iam":       {dataset},
		"google_healthcare_consent_store":     {dataset + "/consentStores/cs1"},
		"google_healthcare_consent_store_iam": {dataset + "/consentStores/cs1"},
		"google_healthcare_dicom_store":       {dataset + "/dicomStores/dc1"},
		"google_healthcare_dicom_store_iam":   {dataset + "/dicomStores/dc1"},
		"google_healthcare_fhir_store":        {dataset + "/fhirStores/f1"},
		"google_healthcare_fhir_store_iam":    {dataset + "/fhirStores/f1"},
		"google_healthcare_hl7_v2_store":      {dataset + "/hl7V2Stores/h1"},
		"google_healthcare_hl7_v2_store_iam":  {dataset + "/hl7V2Stores/h1"},
		"google_healthcare_workspace":         {dataset + "/workspaces/w1"},
		"google_healthcare_pipeline_job":      {dataset + "/pipelineJobs/pj1"},
	})
}
