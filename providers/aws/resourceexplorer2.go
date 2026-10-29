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
	"github.com/aws/aws-sdk-go-v2/service/resourceexplorer2"
)

var resourceExplorer2AllowEmptyValues = []string{"tags."}

type ResourceExplorer2Generator struct {
	AWSService
}

// InitResources enumerates Resource Explorer indexes (ListIndexes;
// the provider importer ID of aws_resourceexplorer2_index is the
// region the index lives in) and views (ListViews returns view ARNs,
// which are the provider importer IDs of aws_resourceexplorer2_view).
func (g *ResourceExplorer2Generator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := resourceexplorer2.NewFromConfig(config)

	indexes, err := svc.ListIndexes(context.TODO(), &resourceexplorer2.ListIndexesInput{})
	if err != nil {
		return err
	}
	for _, index := range indexes.Indexes {
		region := StringValue(index.Region)
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			region,
			region,
			"aws_resourceexplorer2_index",
			"aws",
			resourceExplorer2AllowEmptyValues))
	}

	views, err := svc.ListViews(context.TODO(), &resourceexplorer2.ListViewsInput{})
	if err != nil {
		return err
	}
	for _, viewArn := range views.Views {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			viewArn,
			arnResourceName(viewArn),
			"aws_resourceexplorer2_view",
			"aws",
			resourceExplorer2AllowEmptyValues))
	}
	return nil
}
