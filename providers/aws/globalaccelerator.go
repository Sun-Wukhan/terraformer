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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/globalaccelerator"
)

var globalAcceleratorAllowEmptyValues = []string{"tags."}

type GlobalAcceleratorGenerator struct {
	AWSService
}

// InitResources enumerates the AWS Global Accelerator objects:
// standard accelerators with their listeners and endpoint groups,
// custom routing accelerators with their listeners and endpoint
// groups, and cross-account attachments. Every resource in this
// family is keyed by its ARN. Global Accelerator is a global service
// homed in us-west-2, but the SDK endpoint follows the configured
// region like the other generators.
func (g *GlobalAcceleratorGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	client := globalaccelerator.NewFromConfig(config)
	if err := g.addAccelerators(client); err != nil {
		return err
	}
	if err := g.addCustomRoutingAccelerators(client); err != nil {
		return err
	}
	return g.addCrossAccountAttachments(client)
}

func (g *GlobalAcceleratorGenerator) addAccelerators(client *globalaccelerator.Client) error {
	input := &globalaccelerator.ListAcceleratorsInput{}
	for {
		out, err := client.ListAccelerators(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, accelerator := range out.Accelerators {
			arn := StringValue(accelerator.AcceleratorArn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn,
				StringValue(accelerator.Name),
				"aws_globalaccelerator_accelerator",
				"aws",
				globalAcceleratorAllowEmptyValues,
			))
			if err := g.addListeners(client, arn); err != nil {
				return err
			}
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *GlobalAcceleratorGenerator) addListeners(client *globalaccelerator.Client, acceleratorArn string) error {
	input := &globalaccelerator.ListListenersInput{AcceleratorArn: aws.String(acceleratorArn)}
	for {
		out, err := client.ListListeners(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, listener := range out.Listeners {
			listenerArn := StringValue(listener.ListenerArn)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				listenerArn,
				listenerArn,
				"aws_globalaccelerator_listener",
				"aws",
				map[string]string{"accelerator_arn": acceleratorArn},
				globalAcceleratorAllowEmptyValues,
				map[string]interface{}{},
			))
			if err := g.addEndpointGroups(client, listenerArn); err != nil {
				return err
			}
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *GlobalAcceleratorGenerator) addEndpointGroups(client *globalaccelerator.Client, listenerArn string) error {
	input := &globalaccelerator.ListEndpointGroupsInput{ListenerArn: aws.String(listenerArn)}
	for {
		out, err := client.ListEndpointGroups(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, group := range out.EndpointGroups {
			groupArn := StringValue(group.EndpointGroupArn)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				groupArn,
				groupArn,
				"aws_globalaccelerator_endpoint_group",
				"aws",
				map[string]string{"listener_arn": listenerArn},
				globalAcceleratorAllowEmptyValues,
				map[string]interface{}{},
			))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *GlobalAcceleratorGenerator) addCustomRoutingAccelerators(client *globalaccelerator.Client) error {
	input := &globalaccelerator.ListCustomRoutingAcceleratorsInput{}
	for {
		out, err := client.ListCustomRoutingAccelerators(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, accelerator := range out.Accelerators {
			arn := StringValue(accelerator.AcceleratorArn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn,
				StringValue(accelerator.Name),
				"aws_globalaccelerator_custom_routing_accelerator",
				"aws",
				globalAcceleratorAllowEmptyValues,
			))
			if err := g.addCustomRoutingListeners(client, arn); err != nil {
				return err
			}
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *GlobalAcceleratorGenerator) addCustomRoutingListeners(client *globalaccelerator.Client, acceleratorArn string) error {
	input := &globalaccelerator.ListCustomRoutingListenersInput{AcceleratorArn: aws.String(acceleratorArn)}
	for {
		out, err := client.ListCustomRoutingListeners(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, listener := range out.Listeners {
			listenerArn := StringValue(listener.ListenerArn)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				listenerArn,
				listenerArn,
				"aws_globalaccelerator_custom_routing_listener",
				"aws",
				map[string]string{"accelerator_arn": acceleratorArn},
				globalAcceleratorAllowEmptyValues,
				map[string]interface{}{},
			))
			if err := g.addCustomRoutingEndpointGroups(client, listenerArn); err != nil {
				return err
			}
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *GlobalAcceleratorGenerator) addCustomRoutingEndpointGroups(client *globalaccelerator.Client, listenerArn string) error {
	input := &globalaccelerator.ListCustomRoutingEndpointGroupsInput{ListenerArn: aws.String(listenerArn)}
	for {
		out, err := client.ListCustomRoutingEndpointGroups(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, group := range out.EndpointGroups {
			groupArn := StringValue(group.EndpointGroupArn)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				groupArn,
				groupArn,
				"aws_globalaccelerator_custom_routing_endpoint_group",
				"aws",
				map[string]string{"listener_arn": listenerArn},
				globalAcceleratorAllowEmptyValues,
				map[string]interface{}{},
			))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *GlobalAcceleratorGenerator) addCrossAccountAttachments(client *globalaccelerator.Client) error {
	input := &globalaccelerator.ListCrossAccountAttachmentsInput{}
	for {
		out, err := client.ListCrossAccountAttachments(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, attachment := range out.CrossAccountAttachments {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(attachment.AttachmentArn),
				StringValue(attachment.Name),
				"aws_globalaccelerator_cross_account_attachment",
				"aws",
				globalAcceleratorAllowEmptyValues,
			))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}
