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
	"log"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/inspector"
)

var inspectorAllowEmptyValues = []string{"tags."}

type InspectorGenerator struct {
	AWSService
}

// InitResources enumerates Inspector Classic assessment targets and
// assessment templates in the configured region; both list APIs
// return ARNs, which are also the provider importer IDs.
//
// Intentionally not generated: aws_inspector_resource_group has no
// list API in the SDK (resource groups are only reachable by ARN
// through DescribeResourceGroups, referenced from assessment
// targets/templates), so there is no enumeration entry point.
func (g *InspectorGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := inspector.NewFromConfig(config)

	targets, err := svc.ListAssessmentTargets(context.TODO(), &inspector.ListAssessmentTargetsInput{})
	if err != nil {
		log.Println(err)
	} else {
		described, err := svc.DescribeAssessmentTargets(context.TODO(), &inspector.DescribeAssessmentTargetsInput{
			AssessmentTargetArns: targets.AssessmentTargetArns,
		})
		if err != nil {
			log.Println(err)
		} else {
			for _, target := range described.AssessmentTargets {
				arn := StringValue(target.Arn)
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					arn,
					StringValue(target.Name),
					"aws_inspector_assessment_target",
					"aws",
					inspectorAllowEmptyValues))
			}
		}
	}

	templates, err := svc.ListAssessmentTemplates(context.TODO(), &inspector.ListAssessmentTemplatesInput{})
	if err != nil {
		log.Println(err)
		return nil
	}
	described, err := svc.DescribeAssessmentTemplates(context.TODO(), &inspector.DescribeAssessmentTemplatesInput{
		AssessmentTemplateArns: templates.AssessmentTemplateArns,
	})
	if err != nil {
		log.Println(err)
		return nil
	}
	for _, template := range described.AssessmentTemplates {
		arn := StringValue(template.Arn)
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			arn,
			StringValue(template.Name),
			"aws_inspector_assessment_template",
			"aws",
			inspectorAllowEmptyValues))
	}
	return nil
}
