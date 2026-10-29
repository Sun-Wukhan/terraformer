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

// CloudWatch Network Flow Monitor generator (service key
// "networkflowmonitor").

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/networkflowmonitor"
)

var networkflowmonitorAllowEmptyValues = []string{"tags."}

type NetworkFlowMonitorGenerator struct {
	AWSService
}

func (g *NetworkFlowMonitorGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := networkflowmonitor.NewFromConfig(config)
	if err := g.createMonitors(svc); err != nil {
		return err
	}
	return g.createScopes(svc)
}

// aws_networkflowmonitor_monitor — importer ID is the monitor name (the API
// exposes no separate monitor ID on list/get responses).
func (g *NetworkFlowMonitorGenerator) createMonitors(svc *networkflowmonitor.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListMonitors(context.TODO(), &networkflowmonitor.ListMonitorsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, monitor := range out.Monitors {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(monitor.MonitorName), StringValue(monitor.MonitorName),
				"aws_networkflowmonitor_monitor", "aws", networkflowmonitorAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

// aws_networkflowmonitor_scope — importer ID is the scope ID.
func (g *NetworkFlowMonitorGenerator) createScopes(svc *networkflowmonitor.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListScopes(context.TODO(), &networkflowmonitor.ListScopesInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, scope := range out.Scopes {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(scope.ScopeId), StringValue(scope.ScopeId),
				"aws_networkflowmonitor_scope", "aws", networkflowmonitorAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
