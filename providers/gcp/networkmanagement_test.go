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
	netmgmtv1 "google.golang.org/api/networkmanagement/v1"
	"google.golang.org/api/option"
)

// TestNetworkManagementGenerator feeds canned Network Management
// responses through the generator and verifies the emitted types and
// IDs: the connectivity test and project VPC flow logs config from
// the v1 client, the network monitoring provider from the raw REST
// v1alpha1 list, and the organization VPC flow logs config from the
// raw REST v1 list under the ancestry-resolved organization.
func TestNetworkManagementGenerator(t *testing.T) {
	transport := &cannedTransport{routes: []cannedRoute{
		{":getAncestry", `{"ancestor":[{"resourceId":{"type":"organization","id":"6789"}}]}`},
		{"/organizations/6789/locations/global/vpcFlowLogsConfigs", `{"vpcFlowLogsConfigs":[{"name":"organizations/6789/locations/global/vpcFlowLogsConfigs/org-config"}]}`},
		{"/locations/us-central1/networkMonitoringProviders", `{"networkMonitoringProviders":[{"name":"projects/test-project/locations/us-central1/networkMonitoringProviders/external"}]}`},
		{"/locations/global/networkMonitoringProviders", `{"networkMonitoringProviders":[]}`},
		{"/locations/global/connectivityTests", `{"resources":[{"name":"projects/test-project/locations/global/connectivityTests/test-1"}]}`},
		{"/locations/global/vpcFlowLogsConfigs", `{"vpcFlowLogsConfigs":[{"name":"projects/test-project/locations/global/vpcFlowLogsConfigs/proj-config"}]}`},
	}}
	ctx := context.Background()
	service, err := netmgmtv1.NewService(ctx, option.WithHTTPClient(transport.client()), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("networkmanagement.NewService: %v", err)
	}
	g := &NetworkManagementGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{
		"project": "test-project",
		"region":  compute.Region{Name: "us-central1"},
	})
	if err := g.initResources(ctx, service, crmTestService(t, transport), transport.client()); err != nil {
		t.Fatalf("initResources: %v", err)
	}
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_network_management_connectivity_test":                 {"projects/test-project/locations/global/connectivityTests/test-1"},
		"google_network_management_vpc_flow_logs_config":              {"projects/test-project/locations/global/vpcFlowLogsConfigs/proj-config"},
		"google_network_management_network_monitoring_provider":       {"projects/test-project/locations/us-central1/networkMonitoringProviders/external"},
		"google_network_management_organization_vpc_flow_logs_config": {"organizations/6789/locations/global/vpcFlowLogsConfigs/org-config"},
	})
}
