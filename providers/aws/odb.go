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

// Oracle Database@AWS (ODB) coverage (AWS sweep, database cluster).
// Importer IDs are the resource IDs returned by the ODB List APIs.
//
// Skipped in this family, with reasons:
//   - aws_odb_autonomous_database_secrets_manager_integration: no list API;
//     the integration is a property managed on the autonomous database and
//     the SDK exposes no enumerable integration shape separate from it.
//   - aws_odb_iam_role_association: IAM role associations are attributes of
//     the parent cloud VM cluster / Exadata infrastructure resources
//     (AssociateIamRoleToResource); no independent enumeration API.

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/odb"
)

var odbAllowEmptyValues = []string{"tags."}

type ODBGenerator struct {
	AWSService
}

func (g *ODBGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := odb.NewFromConfig(config)

	ap := odb.NewListAutonomousDatabasesPaginator(svc, &odb.ListAutonomousDatabasesInput{})
	for ap.HasMorePages() {
		page, err := ap.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, db := range page.AutonomousDatabases {
			id := StringValue(db.AutonomousDatabaseId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(db.DisplayName), "aws_odb_autonomous_database", "aws", odbAllowEmptyValues))
		}
	}

	cp := odb.NewListCloudAutonomousVmClustersPaginator(svc, &odb.ListCloudAutonomousVmClustersInput{})
	for cp.HasMorePages() {
		page, err := cp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, cluster := range page.CloudAutonomousVmClusters {
			id := StringValue(cluster.CloudAutonomousVmClusterId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(cluster.DisplayName), "aws_odb_cloud_autonomous_vm_cluster", "aws", odbAllowEmptyValues))
		}
	}

	ip := odb.NewListCloudExadataInfrastructuresPaginator(svc, &odb.ListCloudExadataInfrastructuresInput{})
	for ip.HasMorePages() {
		page, err := ip.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, infra := range page.CloudExadataInfrastructures {
			id := StringValue(infra.CloudExadataInfrastructureId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(infra.DisplayName), "aws_odb_cloud_exadata_infrastructure", "aws", odbAllowEmptyValues))
		}
	}

	vp := odb.NewListCloudVmClustersPaginator(svc, &odb.ListCloudVmClustersInput{})
	for vp.HasMorePages() {
		page, err := vp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, cluster := range page.CloudVmClusters {
			id := StringValue(cluster.CloudVmClusterId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(cluster.DisplayName), "aws_odb_cloud_vm_cluster", "aws", odbAllowEmptyValues))
		}
	}

	np := odb.NewListOdbNetworksPaginator(svc, &odb.ListOdbNetworksInput{})
	for np.HasMorePages() {
		page, err := np.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, network := range page.OdbNetworks {
			id := StringValue(network.OdbNetworkId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(network.DisplayName), "aws_odb_network", "aws", odbAllowEmptyValues))
		}
	}

	pp := odb.NewListOdbPeeringConnectionsPaginator(svc, &odb.ListOdbPeeringConnectionsInput{})
	for pp.HasMorePages() {
		page, err := pp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, peering := range page.OdbPeeringConnections {
			id := StringValue(peering.OdbPeeringConnectionId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(peering.DisplayName), "aws_odb_network_peering_connection", "aws", odbAllowEmptyValues))
		}
	}

	return nil
}
