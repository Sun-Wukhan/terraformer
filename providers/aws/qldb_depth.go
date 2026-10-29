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

// QLDB depth coverage (AWS sweep, database cluster).

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/qldb"
)

// loadQLDBStreams emits aws_qldb_stream per ledger,
// ID = "<ledger name>:<stream ID>".
func (g *QLDBGenerator) loadQLDBStreams(svc *qldb.Client, ledgerNames []string) error {
	for _, ledgerName := range ledgerNames {
		p := qldb.NewListJournalKinesisStreamsForLedgerPaginator(svc, &qldb.ListJournalKinesisStreamsForLedgerInput{LedgerName: aws.String(ledgerName)})
		for p.HasMorePages() {
			page, err := p.NextPage(context.TODO())
			if err != nil {
				return err
			}
			for _, stream := range page.Streams {
				streamID := StringValue(stream.StreamId)
				id := ledgerName + ":" + streamID
				name := ledgerName + "_" + StringValue(stream.StreamName)
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, name, "aws_qldb_stream", "aws", qldbAllowEmptyValues))
			}
		}
	}
	return nil
}
