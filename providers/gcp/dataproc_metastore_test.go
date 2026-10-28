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
	metastorev1beta "google.golang.org/api/metastore/v1beta"
	"google.golang.org/api/option"
)

// TestDataprocMetastoreGenerator feeds canned Metastore responses
// and verifies services and federations expand their IAM trios with
// full resource names as IDs.
func TestDataprocMetastoreGenerator(t *testing.T) {
	loc := "projects/test-project/locations/us-central1"
	svcName := loc + "/services/svc1"
	fedName := loc + "/federations/fed1"
	transport := &cannedTransport{routes: []cannedRoute{
		{":getIamPolicy", `{"bindings":[{"role":"roles/metastore.editor","members":["user:a@example.com"]}]}`},
		{"/locations/us-central1/services", `{"services":[{"name":"` + svcName + `"}]}`},
		{"/locations/us-central1/federations", `{"federations":[{"name":"` + fedName + `"}]}`},
	}}
	ctx := context.Background()
	svc, err := metastorev1beta.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("metastore.NewService: %v", err)
	}
	g := &DataprocMetastoreGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	if err := g.initResources(ctx, svc); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_dataproc_metastore_service":                {svcName},
		"google_dataproc_metastore_service_iam_policy":     {svcName},
		"google_dataproc_metastore_service_iam_binding":    {svcName},
		"google_dataproc_metastore_service_iam_member":     {svcName},
		"google_dataproc_metastore_federation":             {fedName},
		"google_dataproc_metastore_federation_iam_policy":  {fedName},
		"google_dataproc_metastore_federation_iam_binding": {fedName},
		"google_dataproc_metastore_federation_iam_member":  {fedName},
	})
}
