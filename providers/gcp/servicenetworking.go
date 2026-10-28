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
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	svcnetv1 "google.golang.org/api/servicenetworking/v1"
)

const serviceNetworkingService = "servicenetworking.googleapis.com"

var serviceNetworkingAllowEmptyValues = []string{"tags."}

var serviceNetworkingAdditionalFields = map[string]interface{}{}

type ServiceNetworkingGenerator struct {
	GCPService
}

// InitResources enumerates Service Networking resources for the
// servicenetworking.googleapis.com peering service: the connections
// of the project's VPC networks (provider ID = the consumer network
// resource name, a colon, and the service), each connection's peered
// DNS domains, and — for connections where VPC Service Controls is
// enabled — the network's VPC Service Controls resource (provider ID
// = services/{service}/projects/{project number}/networks/{network}).
func (g *ServiceNetworkingGenerator) InitResources() error {
	ctx := context.Background()
	service, err := svcnetv1.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, service)
}

func (g *ServiceNetworkingGenerator) initResources(ctx context.Context, service *svcnetv1.APIService) error {
	// The connections list has no pagination surface in the API.
	resp, err := service.Services.Connections.List("services/" + serviceNetworkingService).Context(ctx).Do()
	if err != nil {
		return err
	}
	for _, connection := range resp.Connections {
		g.Resources = append(g.Resources, terraformutils.NewResource(
			connection.Network+":"+serviceNetworkingService,
			connectionNetworkName(connection.Network),
			"google_service_networking_connection",
			g.ProviderName,
			map[string]string{"network": connection.Network, "service": serviceNetworkingService},
			serviceNetworkingAllowEmptyValues,
			serviceNetworkingAdditionalFields,
		))
		g.loadServiceNetworkingNetwork(ctx, service, connection.Network)
	}
	return nil
}

// connectionNetworkName returns the short network name of a consumer
// network resource name (projects/{number}/global/networks/{name}).
func connectionNetworkName(network string) string {
	parts := strings.Split(network, "/")
	return parts[len(parts)-1]
}

func (g *ServiceNetworkingGenerator) loadServiceNetworkingNetwork(ctx context.Context, service *svcnetv1.APIService, network string) {
	// network is the consumer network resource name,
	// projects/{project number}/global/networks/{network}.
	networkParent := "services/" + serviceNetworkingService + "/" + network

	resp, err := service.Services.Projects.Global.Networks.PeeredDnsDomains.List(networkParent).Context(ctx).Do()
	if err != nil {
		log.Println(err)
	} else {
		for _, domain := range resp.PeeredDnsDomains {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				networkParent+"/peeredDnsDomains/"+domain.Name,
				domain.Name,
				"google_service_networking_peered_dns_domain",
				g.ProviderName,
				map[string]string{"network": network, "service": serviceNetworkingService},
				serviceNetworkingAllowEmptyValues,
				serviceNetworkingAdditionalFields,
			))
		}
	}

	controls, err := service.Services.Projects.Global.Networks.GetVpcServiceControls(networkParent).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		return
	}
	if controls.Enabled {
		// The provider ID drops the /global segment:
		// services/{service}/projects/{number}/networks/{network}.
		id := "services/" + serviceNetworkingService + "/" + strings.Replace(network, "/global/networks/", "/networks/", 1)
		g.Resources = append(g.Resources, terraformutils.NewResource(
			id,
			connectionNetworkName(network),
			"google_service_networking_vpc_service_controls",
			g.ProviderName,
			map[string]string{"network": network, "service": serviceNetworkingService},
			serviceNetworkingAllowEmptyValues,
			serviceNetworkingAdditionalFields,
		))
	}
}
