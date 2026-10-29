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
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	"github.com/aws/aws-sdk-go-v2/service/kafka/types"
	"github.com/aws/aws-sdk-go-v2/service/kafkaconnect"
)

// MSK depth for the AWS sweep, including MSK Connect. Provider importer
// IDs follow the terraform-provider-aws docs: configurations, replicators,
// VPC connections and all MSK Connect resources by ARN; the cluster policy
// and the SCRAM secret association by cluster ARN; serverless clusters by
// cluster ARN (ListClustersV2 filtered to the SERVERLESS cluster type, so
// provisioned clusters are not duplicated with aws_msk_cluster).
//
// Deliberately not generated:
//   - aws_msk_topic: topics live in the Kafka data plane; the MSK control
//     plane SDK has no ListTopics/DescribeTopic API at all.
//   - aws_msk_single_scram_secret_association: single-secret association
//     glue between a cluster and one secret; the whole-cluster association
//     is generated as aws_msk_scram_secret_association.
//   - aws_msk_channel IS generated (assembly_followup.go): the MSK Data
//     Delivery channel API (ListChannels) exists in kafka SDK v1.65.7 and
//     the provider imports by channel_arn,cluster_arn — both verified at
//     assembly time.
func (g *MskGenerator) loadExtended(config aws.Config) error {
	svc := kafka.NewFromConfig(config)
	clusterARNs, err := g.mskClusterARNs(svc)
	if err != nil {
		return err
	}
	for _, load := range []func(*kafka.Client, []string) error{
		g.loadClusterPolicies, g.loadScramSecretAssociations, g.loadChannels,
	} {
		if err := load(svc, clusterARNs); err != nil {
			return err
		}
	}
	if err := g.loadConfigurations(svc); err != nil {
		return err
	}
	if err := g.loadReplicators(svc); err != nil {
		return err
	}
	if err := g.loadVpcConnections(svc); err != nil {
		return err
	}
	if err := g.loadServerlessClusters(svc); err != nil {
		return err
	}
	return g.loadConnect(kafkaconnect.NewFromConfig(config))
}

func (g *MskGenerator) mskClusterARNs(svc *kafka.Client) ([]string, error) {
	var arns []string
	p := kafka.NewListClustersPaginator(svc, &kafka.ListClustersInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return nil, err
		}
		for _, cluster := range page.ClusterInfoList {
			arns = append(arns, StringValue(cluster.ClusterArn))
		}
	}
	return arns, nil
}

func (g *MskGenerator) loadClusterPolicies(svc *kafka.Client, clusterARNs []string) error {
	for _, arn := range clusterARNs {
		// Clusters without a policy return an error / empty policy; that
		// is not an enumeration failure.
		out, err := svc.GetClusterPolicy(context.TODO(), &kafka.GetClusterPolicyInput{ClusterArn: aws.String(arn)})
		if err == nil && out.Policy != nil && *out.Policy != "" {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_msk_cluster_policy", "aws", mskAllowEmptyValues))
		}
	}
	return nil
}

func (g *MskGenerator) loadScramSecretAssociations(svc *kafka.Client, clusterARNs []string) error {
	for _, arn := range clusterARNs {
		out, err := svc.ListScramSecrets(context.TODO(), &kafka.ListScramSecretsInput{ClusterArn: aws.String(arn)})
		if err != nil {
			continue // clusters without SCRAM auth have no association
		}
		if len(out.SecretArnList) > 0 {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_msk_scram_secret_association", "aws", mskAllowEmptyValues))
		}
	}
	return nil
}

func (g *MskGenerator) loadConfigurations(svc *kafka.Client) error {
	input := &kafka.ListConfigurationsInput{}
	for {
		out, err := svc.ListConfigurations(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, c := range out.Configurations {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(c.Arn), StringValue(c.Name), "aws_msk_configuration", "aws", mskAllowEmptyValues))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *MskGenerator) loadReplicators(svc *kafka.Client) error {
	input := &kafka.ListReplicatorsInput{}
	for {
		out, err := svc.ListReplicators(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, r := range out.Replicators {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(r.ReplicatorArn), StringValue(r.ReplicatorName), "aws_msk_replicator", "aws", mskAllowEmptyValues))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *MskGenerator) loadVpcConnections(svc *kafka.Client) error {
	input := &kafka.ListVpcConnectionsInput{}
	for {
		out, err := svc.ListVpcConnections(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, v := range out.VpcConnections {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(v.VpcConnectionArn), StringValue(v.VpcConnectionArn), "aws_msk_vpc_connection", "aws", mskAllowEmptyValues))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *MskGenerator) loadServerlessClusters(svc *kafka.Client) error {
	input := &kafka.ListClustersV2Input{ClusterTypeFilter: aws.String(string(types.ClusterTypeServerless))}
	for {
		out, err := svc.ListClustersV2(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, c := range out.ClusterInfoList {
			if c.ClusterType != types.ClusterTypeServerless {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(c.ClusterArn), StringValue(c.ClusterName), "aws_msk_serverless_cluster", "aws", mskAllowEmptyValues))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *MskGenerator) loadConnect(svc *kafkaconnect.Client) error {
	connInput := &kafkaconnect.ListConnectorsInput{}
	for {
		out, err := svc.ListConnectors(context.TODO(), connInput)
		if err != nil {
			return err
		}
		for _, c := range out.Connectors {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(c.ConnectorArn), StringValue(c.ConnectorName), "aws_mskconnect_connector", "aws", mskAllowEmptyValues))
		}
		if out.NextToken == nil {
			break
		}
		connInput.NextToken = out.NextToken
	}
	pluginInput := &kafkaconnect.ListCustomPluginsInput{}
	for {
		out, err := svc.ListCustomPlugins(context.TODO(), pluginInput)
		if err != nil {
			return err
		}
		for _, p := range out.CustomPlugins {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(p.CustomPluginArn), StringValue(p.Name), "aws_mskconnect_custom_plugin", "aws", mskAllowEmptyValues))
		}
		if out.NextToken == nil {
			break
		}
		pluginInput.NextToken = out.NextToken
	}
	workerInput := &kafkaconnect.ListWorkerConfigurationsInput{}
	for {
		out, err := svc.ListWorkerConfigurations(context.TODO(), workerInput)
		if err != nil {
			return err
		}
		for _, w := range out.WorkerConfigurations {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(w.WorkerConfigurationArn), StringValue(w.Name), "aws_mskconnect_worker_configuration", "aws", mskAllowEmptyValues))
		}
		if out.NextToken == nil {
			return nil
		}
		workerInput.NextToken = out.NextToken
	}
}
