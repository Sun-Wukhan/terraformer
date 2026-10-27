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
	ncc "google.golang.org/api/networkconnectivity/v1"
)

var networkConnectivityAllowEmptyValues = []string{"tags."}

var networkConnectivityAdditionalFields = map[string]interface{}{}

type NetworkConnectivityGenerator struct {
	GCPService
}

// InitResources enumerates Network Connectivity Center resources: hubs
// under the global location (with the IAM policy, bindings and members
// expanded from each hub's live IAM policy, and each hub's groups),
// policy-based routes and internal ranges globally, regional endpoints
// and service connection policies in the configured region, and spokes
// in both the configured region and the global location.
//
// Deliberately not generated (verified against the generated v1 and
// v1alpha1 networkconnectivity clients — neither contains a gateways,
// transports or multicloud data transfer service):
//   - google_network_connectivity_multicloud_data_transfer_config and
//     its child google_network_connectivity_destination.
//   - google_network_connectivity_transport and its child
//     google_network_connectivity_gateway_advertised_route.
func (g *NetworkConnectivityGenerator) InitResources() error {
	ctx := context.Background()
	service, err := ncc.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, service)
}

func (g *NetworkConnectivityGenerator) initResources(ctx context.Context, service *ncc.Service) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region)
	parent := "projects/" + project
	globalParent := parent + "/locations/global"
	globalAttrs := map[string]string{"project": project, "location": "global"}

	err := service.Projects.Locations.Global.Hubs.List(globalParent).Pages(ctx, func(page *ncc.ListHubsResponse) error {
		for _, hub := range page.Hubs {
			g.Resources = append(g.Resources, g.newNCCResource(hub.Name, "google_network_connectivity_hub", globalAttrs))
			g.loadHubIam(ctx, service, hub.Name)
			groupErr := service.Projects.Locations.Global.Hubs.Groups.List(hub.Name).Pages(ctx, func(page *ncc.ListGroupsResponse) error {
				for _, group := range page.Groups {
					g.Resources = append(g.Resources, g.newNCCResource(group.Name, "google_network_connectivity_group", globalAttrs))
				}
				return nil
			})
			if groupErr != nil {
				return groupErr
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = service.Projects.Locations.Global.PolicyBasedRoutes.List(globalParent).Pages(ctx, func(page *ncc.ListPolicyBasedRoutesResponse) error {
		for _, route := range page.PolicyBasedRoutes {
			g.Resources = append(g.Resources, g.newNCCResource(route.Name, "google_network_connectivity_policy_based_route", globalAttrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	// Internal ranges exist at both global and regional locations.
	for _, location := range []string{"global", region.Name} {
		attrs := map[string]string{"project": project, "location": location}
		err = service.Projects.Locations.InternalRanges.List(parent+"/locations/"+location).Pages(ctx, func(page *ncc.ListInternalRangesResponse) error {
			for _, internalRange := range page.InternalRanges {
				g.Resources = append(g.Resources, g.newNCCResource(internalRange.Name, "google_network_connectivity_internal_range", attrs))
			}
			return nil
		})
		if err != nil {
			if location == "global" {
				return err
			}
			log.Println(err)
		}
	}

	regionalParent := parent + "/locations/" + region.Name
	regionalAttrs := map[string]string{"project": project, "location": region.Name}
	err = service.Projects.Locations.RegionalEndpoints.List(regionalParent).Pages(ctx, func(page *ncc.ListRegionalEndpointsResponse) error {
		for _, endpoint := range page.RegionalEndpoints {
			g.Resources = append(g.Resources, g.newNCCResource(endpoint.Name, "google_network_connectivity_regional_endpoint", regionalAttrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = service.Projects.Locations.ServiceConnectionPolicies.List(regionalParent).Pages(ctx, func(page *ncc.ListServiceConnectionPoliciesResponse) error {
		for _, policy := range page.ServiceConnectionPolicies {
			g.Resources = append(g.Resources, g.newNCCResource(policy.Name, "google_network_connectivity_service_connection_policy", regionalAttrs))
		}
		return nil
	})
	if err != nil {
		return err
	}
	// Spokes exist at both regional and global locations.
	for _, location := range []string{region.Name, "global"} {
		attrs := map[string]string{"project": project, "location": location}
		err = service.Projects.Locations.Spokes.List(parent+"/locations/"+location).Pages(ctx, func(page *ncc.ListSpokesResponse) error {
			for _, spoke := range page.Spokes {
				g.Resources = append(g.Resources, g.newNCCResource(spoke.Name, "google_network_connectivity_spoke", attrs))
			}
			return nil
		})
		if err != nil {
			if location == region.Name {
				return err
			}
			log.Println(err)
		}
	}
	return nil
}

func (g *NetworkConnectivityGenerator) newNCCResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
	parts := strings.Split(id, "/")
	return terraformutils.NewResource(
		id,
		parts[len(parts)-1],
		resourceType,
		g.ProviderName,
		attrs,
		networkConnectivityAllowEmptyValues,
		networkConnectivityAdditionalFields,
	)
}

// loadHubIam mirrors the Cloud Run IAM pattern: the policy resource ID
// is the hub's full name, and bindings and members are expanded from the
// hub's live IAM policy.
func (g *NetworkConnectivityGenerator) loadHubIam(ctx context.Context, service *ncc.Service, hubName string) {
	policy, err := service.Projects.Locations.Global.Hubs.GetIamPolicy(hubName).Context(ctx).Do()
	if err != nil {
		log.Println(err)
		return
	}
	parentAttrs := map[string]string{"hub": hubName}
	g.Resources = append(g.Resources, g.newNCCResource(
		hubName,
		"google_network_connectivity_hub_iam_policy",
		parentAttrs,
	))
	for _, binding := range policy.Bindings {
		bindingAttrs := map[string]string{"role": binding.Role}
		for k, v := range parentAttrs {
			bindingAttrs[k] = v
		}
		g.Resources = append(g.Resources, g.newNCCResource(
			hubName,
			"google_network_connectivity_hub_iam_binding",
			bindingAttrs,
		))
		for _, member := range binding.Members {
			memberAttrs := map[string]string{"role": binding.Role, "member": member}
			for k, v := range parentAttrs {
				memberAttrs[k] = v
			}
			g.Resources = append(g.Resources, g.newNCCResource(
				hubName,
				"google_network_connectivity_hub_iam_member",
				memberAttrs,
			))
		}
	}
}
