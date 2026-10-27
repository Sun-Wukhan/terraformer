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
	"errors"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
)

var vpcLatticeAllowEmptyValues = []string{"tags."}

type VPCLatticeGenerator struct {
	AWSService
}

// InitResources enumerates the Amazon VPC Lattice objects in the
// configured region: services with their listeners, listener rules and
// auth policies; service networks with their VPC, service and resource
// associations and auth policies; resource configurations, resource
// gateways, domain verifications, target groups and access log
// subscriptions; and resource policies, which are probed on the service
// networks, resource configurations and resource gateways that can
// carry one (a ResourceNotFoundException from the probe simply means
// none is attached).
//
// aws_vpclattice_target_group_attachment is intentionally not
// generated: it has no importer in the Terraform AWS provider and its
// state ID is a create-time composite, while its Read derives the
// target from configuration rather than from any listable object, so
// attachments cannot be reconstructed from ListTargets output.
func (g *VPCLatticeGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	client := vpclattice.NewFromConfig(config)

	for _, add := range []func(*vpclattice.Client) error{
		g.addServices,
		g.addServiceNetworks,
		g.addResourceConfigurations,
		g.addResourceGateways,
		g.addDomainVerifications,
		g.addTargetGroups,
		g.addAccessLogSubscriptions,
		g.addServiceNetworkVpcAssociations,
		g.addServiceNetworkServiceAssociations,
		g.addServiceNetworkResourceAssociations,
	} {
		if err := add(client); err != nil {
			return err
		}
	}
	return nil
}

func (g *VPCLatticeGenerator) addServices(client *vpclattice.Client) error {
	paginator := vpclattice.NewListServicesPaginator(client, &vpclattice.ListServicesInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, service := range page.Items {
			serviceID := *service.Id
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				serviceID,
				*service.Name,
				"aws_vpclattice_service",
				"aws",
				vpcLatticeAllowEmptyValues,
			))
			if err := g.addListeners(client, serviceID); err != nil {
				return err
			}
			if err := g.addAuthPolicy(client, *service.Arn, *service.Name); err != nil {
				return err
			}
			if err := g.addResourcePolicy(client, *service.Arn, *service.Name); err != nil {
				return err
			}
		}
	}
	return nil
}

// Provider ID format: service ID, listener ID (slash separated).
func (g *VPCLatticeGenerator) addListeners(client *vpclattice.Client, serviceID string) error {
	paginator := vpclattice.NewListListenersPaginator(client, &vpclattice.ListListenersInput{
		ServiceIdentifier: aws.String(serviceID),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, listener := range page.Items {
			listenerID := *listener.Id
			g.Resources = append(g.Resources, terraformutils.NewResource(
				serviceID+"/"+listenerID,
				*listener.Name,
				"aws_vpclattice_listener",
				"aws",
				map[string]string{"service_identifier": serviceID},
				vpcLatticeAllowEmptyValues,
				map[string]interface{}{},
			))
			if err := g.addListenerRules(client, serviceID, listenerID); err != nil {
				return err
			}
		}
	}
	return nil
}

// Provider ID format: service ID, listener ID, rule ID (slash separated).
func (g *VPCLatticeGenerator) addListenerRules(client *vpclattice.Client, serviceID, listenerID string) error {
	paginator := vpclattice.NewListRulesPaginator(client, &vpclattice.ListRulesInput{
		ServiceIdentifier:  aws.String(serviceID),
		ListenerIdentifier: aws.String(listenerID),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, rule := range page.Items {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				serviceID+"/"+listenerID+"/"+*rule.Id,
				*rule.Name,
				"aws_vpclattice_listener_rule",
				"aws",
				map[string]string{
					"service_identifier":  serviceID,
					"listener_identifier": listenerID,
				},
				vpcLatticeAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

func (g *VPCLatticeGenerator) addServiceNetworks(client *vpclattice.Client) error {
	paginator := vpclattice.NewListServiceNetworksPaginator(client, &vpclattice.ListServiceNetworksInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, network := range page.Items {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*network.Id,
				*network.Name,
				"aws_vpclattice_service_network",
				"aws",
				vpcLatticeAllowEmptyValues,
			))
			if err := g.addAuthPolicy(client, *network.Arn, *network.Name); err != nil {
				return err
			}
			if err := g.addResourcePolicy(client, *network.Arn, *network.Name); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *VPCLatticeGenerator) addResourceConfigurations(client *vpclattice.Client) error {
	paginator := vpclattice.NewListResourceConfigurationsPaginator(client, &vpclattice.ListResourceConfigurationsInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, configuration := range page.Items {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*configuration.Id,
				*configuration.Name,
				"aws_vpclattice_resource_configuration",
				"aws",
				vpcLatticeAllowEmptyValues,
			))
			if err := g.addResourcePolicy(client, *configuration.Arn, *configuration.Name); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *VPCLatticeGenerator) addResourceGateways(client *vpclattice.Client) error {
	paginator := vpclattice.NewListResourceGatewaysPaginator(client, &vpclattice.ListResourceGatewaysInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, gateway := range page.Items {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*gateway.Id,
				*gateway.Name,
				"aws_vpclattice_resource_gateway",
				"aws",
				vpcLatticeAllowEmptyValues,
			))
			if err := g.addResourcePolicy(client, *gateway.Arn, *gateway.Name); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *VPCLatticeGenerator) addDomainVerifications(client *vpclattice.Client) error {
	paginator := vpclattice.NewListDomainVerificationsPaginator(client, &vpclattice.ListDomainVerificationsInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, verification := range page.Items {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*verification.Id,
				*verification.DomainName,
				"aws_vpclattice_domain_verification",
				"aws",
				vpcLatticeAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *VPCLatticeGenerator) addTargetGroups(client *vpclattice.Client) error {
	paginator := vpclattice.NewListTargetGroupsPaginator(client, &vpclattice.ListTargetGroupsInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, group := range page.Items {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*group.Id,
				*group.Name,
				"aws_vpclattice_target_group",
				"aws",
				vpcLatticeAllowEmptyValues,
			))
		}
	}
	return nil
}

// Access log subscriptions have no name; they are keyed and named by
// their subscription ID.
func (g *VPCLatticeGenerator) addAccessLogSubscriptions(client *vpclattice.Client) error {
	paginator := vpclattice.NewListAccessLogSubscriptionsPaginator(client, &vpclattice.ListAccessLogSubscriptionsInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, subscription := range page.Items {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*subscription.Id,
				*subscription.Id,
				"aws_vpclattice_access_log_subscription",
				"aws",
				vpcLatticeAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *VPCLatticeGenerator) addServiceNetworkVpcAssociations(client *vpclattice.Client) error {
	paginator := vpclattice.NewListServiceNetworkVpcAssociationsPaginator(client, &vpclattice.ListServiceNetworkVpcAssociationsInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, association := range page.Items {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				*association.Id,
				*association.Id,
				"aws_vpclattice_service_network_vpc_association",
				"aws",
				map[string]string{"service_network_identifier": *association.ServiceNetworkId},
				vpcLatticeAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

func (g *VPCLatticeGenerator) addServiceNetworkServiceAssociations(client *vpclattice.Client) error {
	paginator := vpclattice.NewListServiceNetworkServiceAssociationsPaginator(client, &vpclattice.ListServiceNetworkServiceAssociationsInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, association := range page.Items {
			name := *association.Id
			if association.ServiceName != nil && *association.ServiceName != "" {
				name = *association.ServiceName
			}
			g.Resources = append(g.Resources, terraformutils.NewResource(
				*association.Id,
				name,
				"aws_vpclattice_service_network_service_association",
				"aws",
				map[string]string{"service_network_identifier": *association.ServiceNetworkId},
				vpcLatticeAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

func (g *VPCLatticeGenerator) addServiceNetworkResourceAssociations(client *vpclattice.Client) error {
	paginator := vpclattice.NewListServiceNetworkResourceAssociationsPaginator(client, &vpclattice.ListServiceNetworkResourceAssociationsInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, association := range page.Items {
			name := *association.Id
			if association.ResourceConfigurationName != nil && *association.ResourceConfigurationName != "" {
				name = *association.ResourceConfigurationName
			}
			g.Resources = append(g.Resources, terraformutils.NewResource(
				*association.Id,
				name,
				"aws_vpclattice_service_network_resource_association",
				"aws",
				map[string]string{"service_network_identifier": *association.ServiceNetworkId},
				vpcLatticeAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

// The auth policy resource is keyed by the resource identifier (the ARN
// of a service or service network) and exists only when a policy is
// actually attached; the probe reports ResourceNotFoundException
// otherwise.
func (g *VPCLatticeGenerator) addAuthPolicy(client *vpclattice.Client, resourceArn, resourceName string) error {
	out, err := client.GetAuthPolicy(context.TODO(), &vpclattice.GetAuthPolicyInput{
		ResourceIdentifier: aws.String(resourceArn),
	})
	if err != nil {
		var notFound *types.ResourceNotFoundException
		if errors.As(err, &notFound) {
			return nil
		}
		return err
	}
	if out.Policy == nil || *out.Policy == "" {
		return nil
	}
	g.Resources = append(g.Resources, terraformutils.NewResource(
		resourceArn,
		resourceName,
		"aws_vpclattice_auth_policy",
		"aws",
		map[string]string{"resource_identifier": resourceArn},
		vpcLatticeAllowEmptyValues,
		map[string]interface{}{},
	))
	return nil
}

// The resource policy resource is keyed by the resource ARN and exists
// only when a policy document is attached.
func (g *VPCLatticeGenerator) addResourcePolicy(client *vpclattice.Client, resourceArn, resourceName string) error {
	out, err := client.GetResourcePolicy(context.TODO(), &vpclattice.GetResourcePolicyInput{
		ResourceArn: aws.String(resourceArn),
	})
	if err != nil {
		var notFound *types.ResourceNotFoundException
		if errors.As(err, &notFound) {
			return nil
		}
		return err
	}
	if out.Policy == nil || *out.Policy == "" {
		return nil
	}
	g.Resources = append(g.Resources, terraformutils.NewResource(
		resourceArn,
		resourceName,
		"aws_vpclattice_resource_policy",
		"aws",
		map[string]string{"resource_identifier": resourceArn},
		vpcLatticeAllowEmptyValues,
		map[string]interface{}{},
	))
	return nil
}
