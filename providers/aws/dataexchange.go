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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dataexchange"
	detypes "github.com/aws/aws-sdk-go-v2/service/dataexchange/types"
)

var dataexchangeAllowEmptyValues = []string{"tags."}

// DataExchangeGenerator covers Data Exchange data sets, revisions and event actions.
// Provider import IDs (verified): data set = ID; revision = "<data set ID>:<revision ID>";
// event action = ID.
// Skipped in this family:
//   - aws_dataexchange_revision_assets: the provider resource has no Import section
//     in its documentation (no provider importer); it is a write-path bundle that
//     creates assets inside a revision rather than an enumerable inventory object.
type DataExchangeGenerator struct {
	AWSService
}

func (g *DataExchangeGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := dataexchange.NewFromConfig(config)

	p := dataexchange.NewListDataSetsPaginator(svc, &dataexchange.ListDataSetsInput{Origin: aws.String(string(detypes.OriginOwned))})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, ds := range page.DataSets {
			dsID := StringValue(ds.Id)
			name := StringValue(ds.Name)
			if name == "" {
				name = dsID
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				dsID, name, "aws_dataexchange_data_set", "aws", dataexchangeAllowEmptyValues))
			rp := dataexchange.NewListDataSetRevisionsPaginator(svc, &dataexchange.ListDataSetRevisionsInput{DataSetId: ds.Id})
			for rp.HasMorePages() {
				rpage, err := rp.NextPage(context.TODO())
				if err != nil {
					break
				}
				for _, rev := range rpage.Revisions {
					id := dsID + ":" + StringValue(rev.Id)
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						id, id, "aws_dataexchange_revision", "aws", dataexchangeAllowEmptyValues))
				}
			}
		}
	}

	ep := dataexchange.NewListEventActionsPaginator(svc, &dataexchange.ListEventActionsInput{})
	for ep.HasMorePages() {
		page, err := ep.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, ea := range page.EventActions {
			id := StringValue(ea.Id)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_dataexchange_event_action", "aws", dataexchangeAllowEmptyValues))
		}
	}
	return nil
}
