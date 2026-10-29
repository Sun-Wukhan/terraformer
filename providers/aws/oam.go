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

// CloudWatch Observability Access Manager (OAM) generator (service key
// "oam"). Importer IDs are the resource ARNs for links and sinks.

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/oam"
)

var oamAllowEmptyValues = []string{"tags."}

type OamGenerator struct {
	AWSService
}

func (g *OamGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := oam.NewFromConfig(config)
	if err := g.createLinks(svc); err != nil {
		return err
	}
	return g.createSinks(svc)
}

// aws_oam_link — importer ID is the link ARN.
func (g *OamGenerator) createLinks(svc *oam.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListLinks(context.TODO(), &oam.ListLinksInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, link := range out.Items {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(link.Arn), StringValue(link.Label),
				"aws_oam_link", "aws", oamAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

// aws_oam_sink — importer ID is the sink ARN.
// aws_oam_sink_policy — importer ID is the sink ARN; emitted for sinks that
// carry a non-empty policy.
func (g *OamGenerator) createSinks(svc *oam.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListSinks(context.TODO(), &oam.ListSinksInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, sink := range out.Items {
			arn := StringValue(sink.Arn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(sink.Name), "aws_oam_sink", "aws", oamAllowEmptyValues))
			policy, err := svc.GetSinkPolicy(context.TODO(), &oam.GetSinkPolicyInput{SinkIdentifier: sink.Arn})
			if err != nil {
				return err
			}
			if policy.Policy != nil && *policy.Policy != "" {
				g.Resources = append(g.Resources, terraformutils.NewResource(
					arn, StringValue(sink.Name), "aws_oam_sink_policy", "aws",
					map[string]string{"sink_identifier": arn}, oamAllowEmptyValues, map[string]interface{}{}))
			}
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
