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
	"google.golang.org/api/compute/v1"
	netsvcv1 "google.golang.org/api/networkservices/v1"
	netsvcbeta "google.golang.org/api/networkservices/v1beta1"
)

var networkServicesAllowEmptyValues = []string{"tags."}

var networkServicesAdditionalFields = map[string]interface{}{}

type NetworkServicesGenerator struct {
	GCPService
}

// InitResources enumerates Cloud Network Services resources — the
// Gateway API and service mesh estate: gateways, HTTP/gRPC/TCP/TLS
// routes, meshes, service bindings, service LB policies, endpoint
// policies, authz extensions, LB route/traffic/edge extensions and Wasm
// plugins. Every collection is listed at both the configured region and
// the global location; a resource lives in exactly one location, so
// listing both never duplicates.
//
// lb_edge_extension is served by the v1beta1 client: it has no GA list
// surface, matching the google provider, which also exposes it as a GA
// resource backed by the beta API.
//
// Deliberately not generated (verified against the generated v1 and
// v1beta1 networkservices clients):
//   - google_network_services_edge_cache_keyset,
//     google_network_services_edge_cache_origin and
//     google_network_services_edge_cache_service: the EdgeCache
//     services expose only IAM methods (get/set/test IAM policy); there
//     is no list or get surface to enumerate from.
//   - google_network_services_telemetry_policy: no telemetry policy
//     service exists in either client.
//   - google_network_services_agent_gateway and
//     google_network_services_agent_connectivity_template: no agent
//     connectivity services exist in either client.
//   - google_network_services_cdn_edge_extension: no CDN edge extension
//     service exists in either client (only lb_edge_extension, in
//     v1beta1).
//   - The nine multicast types google_network_services_multicast_domain,
//     google_network_services_multicast_domain_group,
//     google_network_services_multicast_domain_activation,
//     google_network_services_multicast_group_range,
//     google_network_services_multicast_group_range_activation,
//     google_network_services_multicast_group_consumer_activation,
//     google_network_services_multicast_group_producer_activation,
//     google_network_services_multicast_consumer_association and
//     google_network_services_multicast_producer_association: no
//     multicast services exist in either client.
func (g *NetworkServicesGenerator) InitResources() error {
	ctx := context.Background()
	v1Service, err := netsvcv1.NewService(ctx)
	if err != nil {
		return err
	}
	betaService, err := netsvcbeta.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, v1Service, betaService)
}

func (g *NetworkServicesGenerator) initResources(ctx context.Context, v1Service *netsvcv1.Service, betaService *netsvcbeta.Service) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region)
	parent := "projects/" + project
	if err := g.loadLocation(ctx, v1Service, betaService, parent, project, region.Name); err != nil {
		return err
	}
	if err := g.loadLocation(ctx, v1Service, betaService, parent, project, "global"); err != nil {
		log.Println(err)
	}
	return nil
}

func (g *NetworkServicesGenerator) newNetSvcResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
	parts := strings.Split(id, "/")
	return terraformutils.NewResource(
		id,
		parts[len(parts)-1],
		resourceType,
		g.ProviderName,
		attrs,
		networkServicesAllowEmptyValues,
		networkServicesAdditionalFields,
	)
}

func (g *NetworkServicesGenerator) loadLocation(ctx context.Context, v1Service *netsvcv1.Service, betaService *netsvcbeta.Service, parent, project, location string) error {
	locationParent := parent + "/locations/" + location
	attrs := map[string]string{"project": project, "location": location}

	err := v1Service.Projects.Locations.AuthzExtensions.List(locationParent).Pages(ctx, func(page *netsvcv1.ListAuthzExtensionsResponse) error {
		for _, extension := range page.AuthzExtensions {
			g.Resources = append(g.Resources, g.newNetSvcResource(extension.Name, "google_network_services_authz_extension", attrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = v1Service.Projects.Locations.EndpointPolicies.List(locationParent).Pages(ctx, func(page *netsvcv1.ListEndpointPoliciesResponse) error {
		for _, policy := range page.EndpointPolicies {
			g.Resources = append(g.Resources, g.newNetSvcResource(policy.Name, "google_network_services_endpoint_policy", attrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = v1Service.Projects.Locations.Gateways.List(locationParent).Pages(ctx, func(page *netsvcv1.ListGatewaysResponse) error {
		for _, gateway := range page.Gateways {
			g.Resources = append(g.Resources, g.newNetSvcResource(gateway.Name, "google_network_services_gateway", attrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = v1Service.Projects.Locations.GrpcRoutes.List(locationParent).Pages(ctx, func(page *netsvcv1.ListGrpcRoutesResponse) error {
		for _, route := range page.GrpcRoutes {
			g.Resources = append(g.Resources, g.newNetSvcResource(route.Name, "google_network_services_grpc_route", attrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = v1Service.Projects.Locations.HttpRoutes.List(locationParent).Pages(ctx, func(page *netsvcv1.ListHttpRoutesResponse) error {
		for _, route := range page.HttpRoutes {
			g.Resources = append(g.Resources, g.newNetSvcResource(route.Name, "google_network_services_http_route", attrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = v1Service.Projects.Locations.LbRouteExtensions.List(locationParent).Pages(ctx, func(page *netsvcv1.ListLbRouteExtensionsResponse) error {
		for _, extension := range page.LbRouteExtensions {
			g.Resources = append(g.Resources, g.newNetSvcResource(extension.Name, "google_network_services_lb_route_extension", attrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = v1Service.Projects.Locations.LbTrafficExtensions.List(locationParent).Pages(ctx, func(page *netsvcv1.ListLbTrafficExtensionsResponse) error {
		for _, extension := range page.LbTrafficExtensions {
			g.Resources = append(g.Resources, g.newNetSvcResource(extension.Name, "google_network_services_lb_traffic_extension", attrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = v1Service.Projects.Locations.Meshes.List(locationParent).Pages(ctx, func(page *netsvcv1.ListMeshesResponse) error {
		for _, mesh := range page.Meshes {
			g.Resources = append(g.Resources, g.newNetSvcResource(mesh.Name, "google_network_services_mesh", attrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = v1Service.Projects.Locations.ServiceBindings.List(locationParent).Pages(ctx, func(page *netsvcv1.ListServiceBindingsResponse) error {
		for _, binding := range page.ServiceBindings {
			g.Resources = append(g.Resources, g.newNetSvcResource(binding.Name, "google_network_services_service_binding", attrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = v1Service.Projects.Locations.ServiceLbPolicies.List(locationParent).Pages(ctx, func(page *netsvcv1.ListServiceLbPoliciesResponse) error {
		for _, policy := range page.ServiceLbPolicies {
			g.Resources = append(g.Resources, g.newNetSvcResource(policy.Name, "google_network_services_service_lb_policies", attrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = v1Service.Projects.Locations.TcpRoutes.List(locationParent).Pages(ctx, func(page *netsvcv1.ListTcpRoutesResponse) error {
		for _, route := range page.TcpRoutes {
			g.Resources = append(g.Resources, g.newNetSvcResource(route.Name, "google_network_services_tcp_route", attrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = v1Service.Projects.Locations.TlsRoutes.List(locationParent).Pages(ctx, func(page *netsvcv1.ListTlsRoutesResponse) error {
		for _, route := range page.TlsRoutes {
			g.Resources = append(g.Resources, g.newNetSvcResource(route.Name, "google_network_services_tls_route", attrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = v1Service.Projects.Locations.WasmPlugins.List(locationParent).Pages(ctx, func(page *netsvcv1.ListWasmPluginsResponse) error {
		for _, plugin := range page.WasmPlugins {
			g.Resources = append(g.Resources, g.newNetSvcResource(plugin.Name, "google_network_services_wasm_plugin", attrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	return betaService.Projects.Locations.LbEdgeExtensions.List(locationParent).Pages(ctx, func(page *netsvcbeta.ListLbEdgeExtensionsResponse) error {
		for _, extension := range page.LbEdgeExtensions {
			g.Resources = append(g.Resources, g.newNetSvcResource(extension.Name, "google_network_services_lb_edge_extension", attrs))
		}
		return nil
	})
}
