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

// CloudWatch Internet Monitor generator (service key "internetmonitor").

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/internetmonitor"
)

var internetmonitorAllowEmptyValues = []string{"tags."}

type InternetMonitorGenerator struct {
	AWSService
}

// aws_internetmonitor_monitor — importer ID is the monitor name.
func (g *InternetMonitorGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := internetmonitor.NewFromConfig(config)
	var nextToken *string
	for {
		out, err := svc.ListMonitors(context.TODO(), &internetmonitor.ListMonitorsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, monitor := range out.Monitors {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(monitor.MonitorName), StringValue(monitor.MonitorName),
				"aws_internetmonitor_monitor", "aws", internetmonitorAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
