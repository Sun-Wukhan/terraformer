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
	"log"
	"net/http"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"google.golang.org/api/compute/v1"
	healthcare "google.golang.org/api/healthcare/v1"
)

// HealthcareGenerator enumerates the Cloud Healthcare API in the
// configured region: datasets with their consent, DICOM, FHIR and
// HL7v2 stores, via the generated healthcare v1 client. Every
// dataset/store also emits the combined google_healthcare_*_iam
// resource when it carries IAM bindings (the provider models
// Healthcare IAM through those combined _iam resources).
//
// Workspaces and pipeline jobs have no list surface in the pinned
// generated client (workspaces expose only IAM methods there), so
// they are listed over raw REST through gcp_rest.go against the
// healthcare v1 endpoints {{dataset}}/workspaces and
// {{dataset}}/pipelineJobs.
type HealthcareGenerator struct {
	GCPService
}

const healthcareBaseURL = "https://healthcare.googleapis.com/v1/"

func (g *HealthcareGenerator) InitResources() error {
	ctx := context.Background()
	service, err := healthcare.NewService(ctx)
	if err != nil {
		return err
	}
	restClient, err := newAuthenticatedHTTPClient(ctx)
	if err != nil {
		log.Println(err)
	}
	g.initResources(ctx, service, restClient)
	return nil
}

func (g *HealthcareGenerator) newResource(id, resourceType string) terraformutils.Resource {
	return terraformutils.NewResource(
		id, opsLastSegment(id), resourceType, g.ProviderName,
		map[string]string{}, []string{}, map[string]interface{}{},
	)
}

// emitIam records the combined _iam resource for parentID when the
// policy call succeeds and returns at least one binding.
func (g *HealthcareGenerator) emitIam(iamType string, policy *healthcare.Policy, err error, parentID string) {
	if err != nil {
		log.Println(err)
		return
	}
	if len(policy.Bindings) > 0 {
		g.Resources = append(g.Resources, g.newResource(parentID, iamType))
	}
}

func (g *HealthcareGenerator) initResources(ctx context.Context, service *healthcare.Service, restClient *http.Client) {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := "projects/" + project + "/locations/" + region
	if err := service.Projects.Locations.Datasets.List(parent).Pages(ctx, func(page *healthcare.ListDatasetsResponse) error {
		for _, dataset := range page.Datasets {
			g.Resources = append(g.Resources, g.newResource(dataset.Name, "google_healthcare_dataset"))
			policy, err := service.Projects.Locations.Datasets.GetIamPolicy(dataset.Name).Context(ctx).Do()
			g.emitIam("google_healthcare_dataset_iam", policy, err, dataset.Name)
			g.loadStores(ctx, service, dataset.Name)
			g.loadDatasetRest(ctx, restClient, dataset.Name)
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
}

func (g *HealthcareGenerator) loadStores(ctx context.Context, service *healthcare.Service, dataset string) {
	datasets := service.Projects.Locations.Datasets
	if err := datasets.ConsentStores.List(dataset).Pages(ctx, func(page *healthcare.ListConsentStoresResponse) error {
		for _, store := range page.ConsentStores {
			g.Resources = append(g.Resources, g.newResource(store.Name, "google_healthcare_consent_store"))
			policy, err := datasets.ConsentStores.GetIamPolicy(store.Name).Context(ctx).Do()
			g.emitIam("google_healthcare_consent_store_iam", policy, err, store.Name)
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := datasets.DicomStores.List(dataset).Pages(ctx, func(page *healthcare.ListDicomStoresResponse) error {
		for _, store := range page.DicomStores {
			g.Resources = append(g.Resources, g.newResource(store.Name, "google_healthcare_dicom_store"))
			policy, err := datasets.DicomStores.GetIamPolicy(store.Name).Context(ctx).Do()
			g.emitIam("google_healthcare_dicom_store_iam", policy, err, store.Name)
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := datasets.FhirStores.List(dataset).Pages(ctx, func(page *healthcare.ListFhirStoresResponse) error {
		for _, store := range page.FhirStores {
			g.Resources = append(g.Resources, g.newResource(store.Name, "google_healthcare_fhir_store"))
			policy, err := datasets.FhirStores.GetIamPolicy(store.Name).Context(ctx).Do()
			g.emitIam("google_healthcare_fhir_store_iam", policy, err, store.Name)
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	if err := datasets.Hl7V2Stores.List(dataset).Pages(ctx, func(page *healthcare.ListHl7V2StoresResponse) error {
		for _, store := range page.Hl7V2Stores {
			g.Resources = append(g.Resources, g.newResource(store.Name, "google_healthcare_hl7_v2_store"))
			policy, err := datasets.Hl7V2Stores.GetIamPolicy(store.Name).Context(ctx).Do()
			g.emitIam("google_healthcare_hl7_v2_store_iam", policy, err, store.Name)
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
}

func (g *HealthcareGenerator) loadDatasetRest(ctx context.Context, restClient *http.Client, dataset string) {
	if restClient == nil {
		return
	}
	workspaces, err := restListResources(ctx, restClient, healthcareBaseURL+dataset+"/workspaces")
	if err != nil {
		log.Println(err)
	} else {
		for _, workspace := range workspaces {
			g.Resources = append(g.Resources, g.newResource(workspace.Name, "google_healthcare_workspace"))
		}
	}
	jobs, err := restListResources(ctx, restClient, healthcareBaseURL+dataset+"/pipelineJobs")
	if err != nil {
		log.Println(err)
	} else {
		for _, job := range jobs {
			g.Resources = append(g.Resources, g.newResource(job.Name, "google_healthcare_pipeline_job"))
		}
	}
}
