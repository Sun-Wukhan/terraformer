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
	"github.com/aws/aws-sdk-go-v2/service/arczonalshift"
)

var arcZonalShiftAllowEmptyValues = []string{"tags."}

type ArcZonalShiftGenerator struct {
	AWSService
}

// InitResources enumerates the Amazon Application Recovery
// Controller zonal shift managed resources in the configured region.
// A managed resource is a load balancer that has been enrolled for
// zonal shift / zonal autoshift; its zonal autoshift configuration
// is keyed by the resource identifier (the resource ARN).
//
// aws_arczonalshift_autoshift_observer_notification_status is
// intentionally not generated: it is the single account/region-level
// observer notification setting (the one result of
// GetAutoshiftObserverNotificationStatus), a singleton settings
// object rather than an enumerable resource.
func (g *ArcZonalShiftGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	client := arczonalshift.NewFromConfig(config)
	input := &arczonalshift.ListManagedResourcesInput{}
	for {
		out, err := client.ListManagedResources(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, resource := range out.Items {
			arn := StringValue(resource.Arn)
			// A managed resource only carries a zonal autoshift
			// configuration once one has been created for it; the
			// practice run configuration in GetManagedResource is
			// the existence signal (a merely zonal-shift-capable
			// load balancer has none).
			detail, err := client.GetManagedResource(context.TODO(), &arczonalshift.GetManagedResourceInput{
				ResourceIdentifier: resource.Arn,
			})
			if err != nil {
				return err
			}
			if detail.PracticeRunConfiguration == nil {
				continue
			}
			name := StringValue(resource.Name)
			if name == "" {
				name = arn
			}
			g.Resources = append(g.Resources, terraformutils.NewResource(
				arn,
				name,
				"aws_arczonalshift_zonal_autoshift_configuration",
				"aws",
				map[string]string{"resource_arn": arn},
				arcZonalShiftAllowEmptyValues,
				map[string]interface{}{},
			))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}
