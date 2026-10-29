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

// CloudWatch Network Monitor generator (service key "networkmonitor").

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/networkmonitor"
)

var networkmonitorAllowEmptyValues = []string{"tags."}

type NetworkMonitorGenerator struct {
	AWSService
}

func (g *NetworkMonitorGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := networkmonitor.NewFromConfig(config)
	var nextToken *string
	for {
		out, err := svc.ListMonitors(context.TODO(), &networkmonitor.ListMonitorsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, monitor := range out.Monitors {
			monitorName := StringValue(monitor.MonitorName)
			// aws_networkmonitor_monitor — importer ID is the monitor name.
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				monitorName, monitorName, "aws_networkmonitor_monitor", "aws", networkmonitorAllowEmptyValues))
			if err := g.createProbes(svc, monitorName); err != nil {
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

// aws_networkmonitor_probe — importer ID "monitor_name,probe_id". Probes are
// enumerated from the parent monitor's GetMonitor response (the SDK has no
// ListProbes operation).
func (g *NetworkMonitorGenerator) createProbes(svc *networkmonitor.Client, monitorName string) error {
	out, err := svc.GetMonitor(context.TODO(), &networkmonitor.GetMonitorInput{MonitorName: &monitorName})
	if err != nil {
		return err
	}
	for _, probe := range out.Probes {
		id := monitorName + "," + StringValue(probe.ProbeId)
		g.Resources = append(g.Resources, terraformutils.NewResource(
			id, StringValue(probe.ProbeId), "aws_networkmonitor_probe", "aws",
			map[string]string{"monitor_name": monitorName}, networkmonitorAllowEmptyValues, map[string]interface{}{}))
	}
	return nil
}
