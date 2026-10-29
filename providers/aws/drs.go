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
	"github.com/aws/aws-sdk-go-v2/service/drs"
)

// DRS (Elastic Disaster Recovery) sweep.
// aws_drs_replication_configuration_template imports with the template
// ID (verified against provider docs), enumerated via
// DescribeReplicationConfigurationTemplates.

var drsAllowEmptyValues = []string{"tags."}

type DrsGenerator struct {
	AWSService
}

func (g *DrsGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := drs.NewFromConfig(config)
	ctx := context.TODO()
	input := &drs.DescribeReplicationConfigurationTemplatesInput{}
	for {
		out, err := svc.DescribeReplicationConfigurationTemplates(ctx, input)
		if err != nil {
			return err
		}
		for _, tmpl := range out.Items {
			id := StringValue(tmpl.ReplicationConfigurationTemplateID)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(id, id, "aws_drs_replication_configuration_template", "aws", drsAllowEmptyValues))
		}
		if out.NextToken == nil {
			break
		}
		input.NextToken = out.NextToken
	}
	return nil
}
