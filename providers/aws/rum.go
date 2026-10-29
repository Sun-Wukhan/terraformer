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

// CloudWatch RUM generator (service key "rum").

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/rum"
)

var rumAllowEmptyValues = []string{"tags."}

type RumGenerator struct {
	AWSService
}

func (g *RumGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := rum.NewFromConfig(config)
	var nextToken *string
	for {
		out, err := svc.ListAppMonitors(context.TODO(), &rum.ListAppMonitorsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, monitor := range out.AppMonitorSummaries {
			name := StringValue(monitor.Name)
			// aws_rum_app_monitor — importer ID is the app monitor name.
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_rum_app_monitor", "aws", rumAllowEmptyValues))
			if err := g.createMetricsDestinations(svc, name); err != nil {
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

// aws_rum_metrics_destination — the provider ID is the app monitor name
// (one resource per app monitor); emitted when the monitor has at least one
// metrics destination.
func (g *RumGenerator) createMetricsDestinations(svc *rum.Client, appMonitorName string) error {
	var nextToken *string
	found := false
	for {
		out, err := svc.ListRumMetricsDestinations(context.TODO(), &rum.ListRumMetricsDestinationsInput{AppMonitorName: &appMonitorName, NextToken: nextToken})
		if err != nil {
			return err
		}
		if len(out.Destinations) > 0 {
			found = true
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	if found {
		g.Resources = append(g.Resources, terraformutils.NewResource(
			appMonitorName, appMonitorName, "aws_rum_metrics_destination", "aws",
			map[string]string{"app_monitor_name": appMonitorName}, rumAllowEmptyValues, map[string]interface{}{}))
	}
	return nil
}
