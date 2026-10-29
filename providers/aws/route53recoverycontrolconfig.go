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
	"github.com/aws/aws-sdk-go-v2/service/route53recoverycontrolconfig"
)

var route53RecoveryControlAllowEmptyValues = []string{"tags."}

type Route53RecoveryControlConfigGenerator struct {
	AWSService
}

// InitResources enumerates the Amazon Route 53 Application Recovery
// Controller (recovery control configuration) objects: clusters with
// their control panels, routing controls and safety rules. Every
// resource in this family is keyed by its ARN.
func (g *Route53RecoveryControlConfigGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	client := route53recoverycontrolconfig.NewFromConfig(config)
	input := &route53recoverycontrolconfig.ListClustersInput{}
	for {
		out, err := client.ListClusters(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, cluster := range out.Clusters {
			clusterArn := StringValue(cluster.ClusterArn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				clusterArn,
				StringValue(cluster.Name),
				"aws_route53recoverycontrolconfig_cluster",
				"aws",
				route53RecoveryControlAllowEmptyValues,
			))
			if err := g.addControlPanels(client, clusterArn); err != nil {
				return err
			}
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *Route53RecoveryControlConfigGenerator) addControlPanels(client *route53recoverycontrolconfig.Client, clusterArn string) error {
	input := &route53recoverycontrolconfig.ListControlPanelsInput{ClusterArn: aws.String(clusterArn)}
	for {
		out, err := client.ListControlPanels(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, panel := range out.ControlPanels {
			panelArn := StringValue(panel.ControlPanelArn)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				panelArn,
				StringValue(panel.Name),
				"aws_route53recoverycontrolconfig_control_panel",
				"aws",
				map[string]string{"cluster_arn": clusterArn},
				route53RecoveryControlAllowEmptyValues,
				map[string]interface{}{},
			))
			if err := g.addRoutingControls(client, panelArn); err != nil {
				return err
			}
			if err := g.addSafetyRules(client, panelArn); err != nil {
				return err
			}
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *Route53RecoveryControlConfigGenerator) addRoutingControls(client *route53recoverycontrolconfig.Client, panelArn string) error {
	input := &route53recoverycontrolconfig.ListRoutingControlsInput{ControlPanelArn: aws.String(panelArn)}
	for {
		out, err := client.ListRoutingControls(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, control := range out.RoutingControls {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				StringValue(control.RoutingControlArn),
				StringValue(control.Name),
				"aws_route53recoverycontrolconfig_routing_control",
				"aws",
				map[string]string{"control_panel_arn": panelArn},
				route53RecoveryControlAllowEmptyValues,
				map[string]interface{}{},
			))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *Route53RecoveryControlConfigGenerator) addSafetyRules(client *route53recoverycontrolconfig.Client, panelArn string) error {
	input := &route53recoverycontrolconfig.ListSafetyRulesInput{ControlPanelArn: aws.String(panelArn)}
	for {
		out, err := client.ListSafetyRules(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, rule := range out.SafetyRules {
			var arn, name string
			switch {
			case rule.ASSERTION != nil:
				arn = StringValue(rule.ASSERTION.SafetyRuleArn)
				name = StringValue(rule.ASSERTION.Name)
			case rule.GATING != nil:
				arn = StringValue(rule.GATING.SafetyRuleArn)
				name = StringValue(rule.GATING.Name)
			default:
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewResource(
				arn,
				name,
				"aws_route53recoverycontrolconfig_safety_rule",
				"aws",
				map[string]string{"control_panel_arn": panelArn},
				route53RecoveryControlAllowEmptyValues,
				map[string]interface{}{},
			))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}
