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
	"github.com/aws/aws-sdk-go-v2/service/servicediscovery"
	"github.com/aws/aws-sdk-go-v2/service/servicediscovery/types"
)

var serviceDiscoveryAllowEmptyValues = []string{"tags."}

// ServiceDiscoveryGenerator covers AWS Cloud Map. Namespaces are split by
// their API type into the three provider namespace resources; services
// import with their ID; instances with service_id/instance_id (verified
// against the Terraform AWS provider docs).
type ServiceDiscoveryGenerator struct {
	AWSService
}

func (g *ServiceDiscoveryGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := servicediscovery.NewFromConfig(config)

	if err := g.loadNamespaces(svc); err != nil {
		return err
	}
	if err := g.loadServices(svc); err != nil {
		return err
	}
	return nil
}

func (g *ServiceDiscoveryGenerator) loadNamespaces(svc *servicediscovery.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListNamespaces(context.TODO(), &servicediscovery.ListNamespacesInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, ns := range out.Namespaces {
			var resourceType string
			switch ns.Type {
			case types.NamespaceTypeHttp:
				resourceType = "aws_service_discovery_http_namespace"
			case types.NamespaceTypeDnsPrivate:
				resourceType = "aws_service_discovery_private_dns_namespace"
			case types.NamespaceTypeDnsPublic:
				resourceType = "aws_service_discovery_public_dns_namespace"
			default:
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(ns.Id),
				StringValue(ns.Name),
				resourceType,
				"aws",
				serviceDiscoveryAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

func (g *ServiceDiscoveryGenerator) loadServices(svc *servicediscovery.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListServices(context.TODO(), &servicediscovery.ListServicesInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, service := range out.Services {
			serviceID := StringValue(service.Id)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				serviceID,
				StringValue(service.Name),
				"aws_service_discovery_service",
				"aws",
				serviceDiscoveryAllowEmptyValues))
			if err := g.loadInstances(svc, serviceID); err != nil {
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

func (g *ServiceDiscoveryGenerator) loadInstances(svc *servicediscovery.Client, serviceID string) error {
	var nextToken *string
	for {
		out, err := svc.ListInstances(context.TODO(), &servicediscovery.ListInstancesInput{
			ServiceId: &serviceID,
			NextToken: nextToken,
		})
		if err != nil {
			return err
		}
		for _, instance := range out.Instances {
			id := serviceID + "/" + StringValue(instance.Id)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				StringValue(instance.Id),
				"aws_service_discovery_instance",
				"aws",
				serviceDiscoveryAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
