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
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/controltower"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
)

var controltowerAllowEmptyValues = []string{"tags."}

// ControlTowerGenerator covers Control Tower landing zones, enabled baselines
// and enabled controls.
// Provider import IDs (verified): landing zone = landing zone ID; baseline =
// enabled baseline ARN; control = "<OU ARN>,<control ARN>".
type ControlTowerGenerator struct {
	AWSService
}

func (g *ControlTowerGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := controltower.NewFromConfig(config)

	lzp := controltower.NewListLandingZonesPaginator(svc, &controltower.ListLandingZonesInput{})
	for lzp.HasMorePages() {
		page, err := lzp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, lz := range page.LandingZones {
			// ListLandingZones summaries carry only the ARN; the provider ID is
			// the landing zone identifier in the ARN's "landingzone/<id>" suffix.
			arn := StringValue(lz.Arn)
			id := arn
			if idx := strings.LastIndex(arn, "/"); idx >= 0 {
				id = arn[idx+1:]
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_controltower_landing_zone", "aws", controltowerAllowEmptyValues))
		}
	}

	bp := controltower.NewListEnabledBaselinesPaginator(svc, &controltower.ListEnabledBaselinesInput{})
	for bp.HasMorePages() {
		page, err := bp.NextPage(context.TODO())
		if err != nil {
			break
		}
		for _, b := range page.EnabledBaselines {
			arn := StringValue(b.Arn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_controltower_baseline", "aws", controltowerAllowEmptyValues))
		}
	}

	// Enabled controls are listed per target (OU). Walk the organization OUs.
	orgSvc := organizations.NewFromConfig(config)
	roots, err := orgSvc.ListRoots(context.TODO(), &organizations.ListRootsInput{})
	if err == nil {
		var ouARNs []string
		var walk func(parentID string)
		walk = func(parentID string) {
			out, err := orgSvc.ListOrganizationalUnitsForParent(context.TODO(),
				&organizations.ListOrganizationalUnitsForParentInput{ParentId: aws.String(parentID)})
			if err != nil {
				return
			}
			for _, ou := range out.OrganizationalUnits {
				ouARNs = append(ouARNs, StringValue(ou.Arn))
				walk(StringValue(ou.Id))
			}
		}
		for _, root := range roots.Roots {
			ouARNs = append(ouARNs, StringValue(root.Arn))
			walk(StringValue(root.Id))
		}
		for _, ouARN := range ouARNs {
			ep := controltower.NewListEnabledControlsPaginator(svc, &controltower.ListEnabledControlsInput{TargetIdentifier: aws.String(ouARN)})
			for ep.HasMorePages() {
				page, err := ep.NextPage(context.TODO())
				if err != nil {
					break
				}
				for _, c := range page.EnabledControls {
					controlARN := StringValue(c.ControlIdentifier)
					id := ouARN + "," + controlARN
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						id, id, "aws_controltower_control", "aws", controltowerAllowEmptyValues))
				}
			}
		}
	}
	return nil
}
