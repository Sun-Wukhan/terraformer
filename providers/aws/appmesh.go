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

package aws

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/appmesh"
)

var appMeshAllowEmptyValues = []string{"tags."}

// AppMeshGenerator covers AWS App Mesh. Import IDs follow the provider's
// documented mesh_name/name composite format (verified for virtual
// services; the same composite pattern applies to the other mesh-scoped
// resources, with mesh/router/route and mesh/gateway/gateway_route for the
// nested types).
type AppMeshGenerator struct {
	AWSService
}

func (g *AppMeshGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := appmesh.NewFromConfig(config)

	var nextToken *string
	for {
		out, err := svc.ListMeshes(context.TODO(), &appmesh.ListMeshesInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, mesh := range out.Meshes {
			meshName := StringValue(mesh.MeshName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				meshName,
				meshName,
				"aws_appmesh_mesh",
				"aws",
				appMeshAllowEmptyValues))
			if err := g.loadAppMeshChildren(svc, meshName); err != nil {
				return err
			}
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

func (g *AppMeshGenerator) loadAppMeshChildren(svc *appmesh.Client, meshName string) error {
	var nextToken *string
	for {
		out, err := svc.ListVirtualNodes(context.TODO(), &appmesh.ListVirtualNodesInput{
			MeshName:  &meshName,
			NextToken: nextToken,
		})
		if err != nil {
			return err
		}
		for _, node := range out.VirtualNodes {
			id := meshName + "/" + StringValue(node.VirtualNodeName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				StringValue(node.VirtualNodeName),
				"aws_appmesh_virtual_node",
				"aws",
				appMeshAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}

	nextToken = nil
	for {
		out, err := svc.ListVirtualServices(context.TODO(), &appmesh.ListVirtualServicesInput{
			MeshName:  &meshName,
			NextToken: nextToken,
		})
		if err != nil {
			return err
		}
		for _, service := range out.VirtualServices {
			id := meshName + "/" + StringValue(service.VirtualServiceName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				StringValue(service.VirtualServiceName),
				"aws_appmesh_virtual_service",
				"aws",
				appMeshAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}

	nextToken = nil
	for {
		out, err := svc.ListVirtualRouters(context.TODO(), &appmesh.ListVirtualRoutersInput{
			MeshName:  &meshName,
			NextToken: nextToken,
		})
		if err != nil {
			return err
		}
		for _, router := range out.VirtualRouters {
			routerName := StringValue(router.VirtualRouterName)
			id := meshName + "/" + routerName
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				routerName,
				"aws_appmesh_virtual_router",
				"aws",
				appMeshAllowEmptyValues))
			if err := g.loadAppMeshRoutes(svc, meshName, routerName); err != nil {
				return err
			}
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}

	nextToken = nil
	for {
		out, err := svc.ListVirtualGateways(context.TODO(), &appmesh.ListVirtualGatewaysInput{
			MeshName:  &meshName,
			NextToken: nextToken,
		})
		if err != nil {
			return err
		}
		for _, gateway := range out.VirtualGateways {
			gatewayName := StringValue(gateway.VirtualGatewayName)
			id := meshName + "/" + gatewayName
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				gatewayName,
				"aws_appmesh_virtual_gateway",
				"aws",
				appMeshAllowEmptyValues))
			if err := g.loadAppMeshGatewayRoutes(svc, meshName, gatewayName); err != nil {
				return err
			}
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

func (g *AppMeshGenerator) loadAppMeshRoutes(svc *appmesh.Client, meshName, routerName string) error {
	var nextToken *string
	for {
		out, err := svc.ListRoutes(context.TODO(), &appmesh.ListRoutesInput{
			MeshName:          &meshName,
			VirtualRouterName: &routerName,
			NextToken:         nextToken,
		})
		if err != nil {
			return err
		}
		for _, route := range out.Routes {
			id := meshName + "/" + routerName + "/" + StringValue(route.RouteName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				StringValue(route.RouteName),
				"aws_appmesh_route",
				"aws",
				appMeshAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

func (g *AppMeshGenerator) loadAppMeshGatewayRoutes(svc *appmesh.Client, meshName, gatewayName string) error {
	var nextToken *string
	for {
		out, err := svc.ListGatewayRoutes(context.TODO(), &appmesh.ListGatewayRoutesInput{
			MeshName:           &meshName,
			VirtualGatewayName: &gatewayName,
			NextToken:          nextToken,
		})
		if err != nil {
			return err
		}
		for _, route := range out.GatewayRoutes {
			id := meshName + "/" + gatewayName + "/" + StringValue(route.GatewayRouteName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				StringValue(route.GatewayRouteName),
				"aws_appmesh_gateway_route",
				"aws",
				appMeshAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
