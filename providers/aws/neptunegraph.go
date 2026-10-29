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

// Neptune Analytics (neptunegraph) coverage (AWS sweep, database cluster).

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/neptunegraph"
)

var neptuneGraphAllowEmptyValues = []string{"tags."}

type NeptuneGraphGenerator struct {
	AWSService
}

func (g *NeptuneGraphGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := neptunegraph.NewFromConfig(config)

	p := neptunegraph.NewListGraphsPaginator(svc, &neptunegraph.ListGraphsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, graph := range page.Graphs {
			graphID := StringValue(graph.Id)
			graphName := StringValue(graph.Name)
			// aws_neptunegraph_graph, ID = graph ID.
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				graphID, graphName, "aws_neptunegraph_graph", "aws", neptuneGraphAllowEmptyValues))

			// aws_neptunegraph_private_graph_endpoint,
			// ID = "<graph ID>,<VPC endpoint ID>".
			ep, err := svc.ListPrivateGraphEndpoints(context.TODO(), &neptunegraph.ListPrivateGraphEndpointsInput{GraphIdentifier: aws.String(graphID)})
			if err != nil {
				return err
			}
			for _, endpoint := range ep.PrivateGraphEndpoints {
				vpceID := StringValue(endpoint.VpcEndpointId)
				id := graphID + "," + vpceID
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, graphName+"_"+vpceID, "aws_neptunegraph_private_graph_endpoint", "aws", neptuneGraphAllowEmptyValues))
			}
		}
	}
	return nil
}
