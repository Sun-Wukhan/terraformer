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

	"google.golang.org/api/option"
	securityposturev1 "google.golang.org/api/securityposture/v1"
)

// TestSecurityPostureGenerator resolves the organization from canned
// Resource Manager ancestry and verifies the posture and posture
// deployment emitted from the canned organization lists.
func TestSecurityPostureGenerator(t *testing.T) {
	org := "organizations/6789/locations/global/"
	transport := &cannedTransport{routes: []cannedRoute{
		{":getAncestry", `{"ancestor":[{"resourceId":{"type":"organization","id":"6789"}}]}`},
		{"/postureDeployments", `{"postureDeployments":[{"name":"` + org + `postureDeployments/dep-1"}]}`},
		{"/postures", `{"postures":[{"name":"` + org + `postures/posture-1"}]}`},
	}}
	ctx := context.Background()
	service, err := securityposturev1.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("securityposture.NewService: %v", err)
	}
	g := &SecurityPostureGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project"})
	if err := g.initResources(ctx, service, crmTestService(t, transport)); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_securityposture_posture":            {org + "postures/posture-1"},
		"google_securityposture_posture_deployment": {org + "postureDeployments/dep-1"},
	})
}
