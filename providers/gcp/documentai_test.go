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
	documentaiv1 "google.golang.org/api/documentai/v1"
	documentaiv1beta3 "google.golang.org/api/documentai/v1beta3"
	"google.golang.org/api/option"
)

// TestDocumentAiGenerator feeds canned Document AI responses for
// the "us" location and verifies the processor, its default-version
// assignment (the processor reports a default version) and its
// dataset schema singleton.
func TestDocumentAiGenerator(t *testing.T) {
	parent := "projects/test-project/locations/us"
	processor := parent + "/processors/p1"
	transport := &cannedTransport{routes: []cannedRoute{
		{"/processors/p1/dataset/datasetSchema", `{"name":"` + processor + `/dataset/datasetSchema"}`},
		{"/locations/us/processors", `{"processors":[{"name":"` + processor + `","displayName":"invoice","defaultProcessorVersion":"` + processor + `/processorVersions/v1"}]}`},
	}}
	ctx := context.Background()
	opts := []option.ClientOption{option.WithHTTPClient(transport.client()), option.WithoutAuthentication()}
	service, err := documentaiv1.NewService(ctx, opts...)
	if err != nil {
		t.Fatalf("documentai v1 NewService: %v", err)
	}
	betaService, err := documentaiv1beta3.NewService(ctx, opts...)
	if err != nil {
		t.Fatalf("documentai v1beta3 NewService: %v", err)
	}
	g := &DocumentAiGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	g.initResources(ctx, service, betaService)
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_document_ai_processor":                 {processor},
		"google_document_ai_processor_default_version": {processor},
		"google_document_ai_schema":                    {processor + "/dataset/datasetSchema"},
	})
}

// TestDocumentAiWarehouseGenerator feeds canned Content Warehouse
// responses and verifies the location singleton and a document
// schema.
func TestDocumentAiWarehouseGenerator(t *testing.T) {
	parent := "projects/test-project/locations/us-central1"
	transport := &cannedTransport{routes: []cannedRoute{
		{"/locations/us-central1/documentSchemas", `{"documentSchemas":[{"name":"` + parent + `/documentSchemas/ds1"}]}`},
		{"/v1/projects/test-project/locations/us-central1", `{"name":"` + parent + `"}`},
	}}
	g := &DocumentAiWarehouseGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	g.initResources(context.Background(), transport.client())
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_document_ai_warehouse_location":        {parent},
		"google_document_ai_warehouse_document_schema": {parent + "/documentSchemas/ds1"},
	})
}

// TestModelArmorGenerator feeds canned Model Armor responses in
// the "us" multi-region and verifies a template and the floor
// setting singleton.
func TestModelArmorGenerator(t *testing.T) {
	parent := "projects/test-project/locations/us"
	transport := &cannedTransport{routes: []cannedRoute{
		{"/locations/us/templates", `{"templates":[{"name":"` + parent + `/templates/t1"}]}`},
		{"/locations/us/floorSetting", `{"name":"` + parent + `/floorSetting"}`},
	}}
	g := &ModelArmorGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us"},
	})
	g.initResources(context.Background(), transport.client())
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_model_armor_template":     {parent + "/templates/t1"},
		"google_model_armor_floorsetting": {parent + "/floorSetting"},
	})
}
