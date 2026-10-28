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

	clouddeployv1 "google.golang.org/api/clouddeploy/v1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

// TestCloudDeployGenerator feeds canned Cloud Deploy responses and
// verifies pipelines, automations, targets, custom target types,
// deploy policies and the IAM trios.
func TestCloudDeployGenerator(t *testing.T) {
	loc := "projects/test-project/locations/us-central1"
	pipeline := loc + "/deliveryPipelines/dp1"
	target := loc + "/targets/t1"
	ctt := loc + "/customTargetTypes/ctt1"
	policy := `{"bindings":[{"role":"roles/clouddeploy.viewer","members":["user:a@example.com"]}]}`
	transport := &cannedTransport{routes: []cannedRoute{
		{":getIamPolicy", policy},
		{"/deliveryPipelines/dp1/automations", `{"automations":[{"name":"` + pipeline + `/automations/a1"}]}`},
		{"/customTargetTypes", `{"customTargetTypes":[{"name":"` + ctt + `"}]}`},
		{"/targets", `{"targets":[{"name":"` + target + `"}]}`},
		{"/deployPolicies", `{"deployPolicies":[{"name":"` + loc + `/deployPolicies/pol1"}]}`},
		{"/deliveryPipelines", `{"deliveryPipelines":[{"name":"` + pipeline + `"}]}`},
	}}
	ctx := context.Background()
	clientOpts := []option.ClientOption{option.WithHTTPClient(transport.client()), option.WithoutAuthentication()}
	deployService, err := clouddeployv1.NewService(ctx, clientOpts...)
	if err != nil {
		t.Fatalf("clouddeploy.NewService: %v", err)
	}
	g := &CloudDeployGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project", "region": compute.Region{Name: "us-central1"}})
	if err := g.initResources(ctx, deployService); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_clouddeploy_custom_target_type":             {ctt},
		"google_clouddeploy_custom_target_type_iam_policy":  {ctt},
		"google_clouddeploy_custom_target_type_iam_binding": {ctt},
		"google_clouddeploy_custom_target_type_iam_member":  {ctt},
		"google_clouddeploy_target":                         {target},
		"google_clouddeploy_target_iam_policy":              {target},
		"google_clouddeploy_target_iam_binding":             {target},
		"google_clouddeploy_target_iam_member":              {target},
		"google_clouddeploy_deploy_policy":                  {loc + "/deployPolicies/pol1"},
		"google_clouddeploy_delivery_pipeline":              {pipeline},
		"google_clouddeploy_delivery_pipeline_iam_policy":   {pipeline},
		"google_clouddeploy_delivery_pipeline_iam_binding":  {pipeline},
		"google_clouddeploy_delivery_pipeline_iam_member":   {pipeline},
		"google_clouddeploy_automation":                     {pipeline + "/automations/a1"},
	})
}
