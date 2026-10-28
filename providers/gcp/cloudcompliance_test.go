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
)

func TestCloudComplianceGeneratorEnumeration(t *testing.T) {
	parent := "organizations/123456789/locations/global"
	transport := &cannedTransport{routes: []cannedRoute{
		{substr: ":getAncestry", body: `{"ancestor":[{"resourceId":{"type":"organization","id":"123456789"}}]}`},
		{substr: "/frameworks", body: `{"frameworks":[{"name":"` + parent + `/frameworks/fw1"}]}`},
		{substr: "/cloudControls", body: `{"cloudControls":[{"name":"` + parent + `/cloudControls/cc1"}]}`},
		{substr: "/frameworkDeployments", body: `{"frameworkDeployments":[{"name":"` + parent + `/frameworkDeployments/fd1"}]}`},
	}}
	crmService := crmTestService(t, transport)
	g := &CloudComplianceGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project", "region": compute.Region{Name: "us-central1"}})
	g.initResources(context.Background(), crmService, transport.client())
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_cloud_security_compliance_framework":            {parent + "/frameworks/fw1"},
		"google_cloud_security_compliance_cloud_control":        {parent + "/cloudControls/cc1"},
		"google_cloud_security_compliance_framework_deployment": {parent + "/frameworkDeployments/fd1"},
	})
}
