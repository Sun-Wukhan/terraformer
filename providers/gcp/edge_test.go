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

// TestEdgeContainerGenerator feeds canned Distributed Cloud Edge
// raw REST responses and verifies clusters, node pools and VPN
// connections.
func TestEdgeContainerGenerator(t *testing.T) {
	loc := "projects/test-project/locations/us-central1"
	cluster := loc + "/clusters/c1"
	transport := &cannedTransport{routes: []cannedRoute{
		{"/clusters/c1/nodePools", `{"nodePools":[{"name":"` + cluster + `/nodePools/np1"}]}`},
		{"/clusters", `{"clusters":[{"name":"` + cluster + `"}]}`},
		{"/vpnConnections", `{"vpnConnections":[{"name":"` + loc + `/vpnConnections/v1"}]}`},
	}}
	g := &EdgeContainerGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project", "region": compute.Region{Name: "us-central1"}})
	g.initResources(context.Background(), transport.client())
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_edgecontainer_cluster":        {cluster},
		"google_edgecontainer_node_pool":      {cluster + "/nodePools/np1"},
		"google_edgecontainer_vpn_connection": {loc + "/vpnConnections/v1"},
	})
}

// TestEdgeNetworkGenerator feeds canned edge network raw REST
// responses and verifies networks, subnets and interconnect
// attachments.
func TestEdgeNetworkGenerator(t *testing.T) {
	loc := "projects/test-project/locations/us-central1"
	network := loc + "/networks/n1"
	transport := &cannedTransport{routes: []cannedRoute{
		{"/networks/n1/subnets", `{"subnets":[{"name":"` + network + `/subnets/s1"}]}`},
		{"/networks", `{"networks":[{"name":"` + network + `"}]}`},
		{"/interconnectAttachments", `{"interconnectAttachments":[{"name":"` + loc + `/interconnectAttachments/ia1"}]}`},
	}}
	g := &EdgeNetworkGenerator{}
	g.ProviderName = "google"
	g.SetArgs(map[string]interface{}{"project": "test-project", "region": compute.Region{Name: "us-central1"}})
	g.initResources(context.Background(), transport.client())
	assertResourceIDs(t, g.Resources, map[string][]string{
		"google_edgenetwork_network":                 {network},
		"google_edgenetwork_subnet":                  {network + "/subnets/s1"},
		"google_edgenetwork_interconnect_attachment": {loc + "/interconnectAttachments/ia1"},
	})
}
