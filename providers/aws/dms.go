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

// DMS coverage (AWS sweep, database cluster). Classic DMS resources use
// their identifiers as provider IDs; the newer data-provider / migration
// resources use the ARN-shaped identifiers returned by the Describe APIs.
// Endpoints with engine "s3" belong to aws_dms_s3_endpoint, all others to
// aws_dms_endpoint.

import (
	"context"
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	dms "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice"
)

var dmsAllowEmptyValues = []string{"tags."}

type DMSGenerator struct {
	AWSService
}

func (g *DMSGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := dms.NewFromConfig(config)

	// aws_dms_certificate, ID = certificate identifier.
	cp := dms.NewDescribeCertificatesPaginator(svc, &dms.DescribeCertificatesInput{})
	for cp.HasMorePages() {
		page, err := cp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, cert := range page.Certificates {
			name := StringValue(cert.CertificateIdentifier)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_dms_certificate", "aws", dmsAllowEmptyValues))
		}
	}

	// aws_dms_data_provider, ID = data provider identifier (ARN).
	dp := dms.NewDescribeDataProvidersPaginator(svc, &dms.DescribeDataProvidersInput{})
	for dp.HasMorePages() {
		page, err := dp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, provider := range page.DataProviders {
			id := StringValue(provider.DataProviderArn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(provider.DataProviderName), "aws_dms_data_provider", "aws", dmsAllowEmptyValues))
		}
	}

	// aws_dms_endpoint / aws_dms_s3_endpoint, ID = endpoint identifier.
	ep := dms.NewDescribeEndpointsPaginator(svc, &dms.DescribeEndpointsInput{})
	for ep.HasMorePages() {
		page, err := ep.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, endpoint := range page.Endpoints {
			name := StringValue(endpoint.EndpointIdentifier)
			resourceType := "aws_dms_endpoint"
			if strings.EqualFold(StringValue(endpoint.EngineName), "s3") {
				resourceType = "aws_dms_s3_endpoint"
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, resourceType, "aws", dmsAllowEmptyValues))
		}
	}

	// aws_dms_event_subscription, ID = subscription ID.
	sp := dms.NewDescribeEventSubscriptionsPaginator(svc, &dms.DescribeEventSubscriptionsInput{})
	for sp.HasMorePages() {
		page, err := sp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, sub := range page.EventSubscriptionsList {
			name := StringValue(sub.CustSubscriptionId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_dms_event_subscription", "aws", dmsAllowEmptyValues))
		}
	}

	// aws_dms_instance_profile, ID = instance profile identifier (ARN).
	ip := dms.NewDescribeInstanceProfilesPaginator(svc, &dms.DescribeInstanceProfilesInput{})
	for ip.HasMorePages() {
		page, err := ip.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, profile := range page.InstanceProfiles {
			id := StringValue(profile.InstanceProfileArn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_dms_instance_profile", "aws", dmsAllowEmptyValues))
		}
	}

	// aws_dms_migration_project, ID = migration project identifier (ARN).
	mp := dms.NewDescribeMigrationProjectsPaginator(svc, &dms.DescribeMigrationProjectsInput{})
	for mp.HasMorePages() {
		page, err := mp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, project := range page.MigrationProjects {
			id := StringValue(project.MigrationProjectArn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, StringValue(project.MigrationProjectName), "aws_dms_migration_project", "aws", dmsAllowEmptyValues))
		}
	}

	// aws_dms_replication_config, ID = replication config identifier (ARN).
	rp := dms.NewDescribeReplicationConfigsPaginator(svc, &dms.DescribeReplicationConfigsInput{})
	for rp.HasMorePages() {
		page, err := rp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, cfg := range page.ReplicationConfigs {
			id := StringValue(cfg.ReplicationConfigIdentifier)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_dms_replication_config", "aws", dmsAllowEmptyValues))
		}
	}

	// aws_dms_replication_instance, ID = replication instance identifier.
	ri := dms.NewDescribeReplicationInstancesPaginator(svc, &dms.DescribeReplicationInstancesInput{})
	for ri.HasMorePages() {
		page, err := ri.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, inst := range page.ReplicationInstances {
			name := StringValue(inst.ReplicationInstanceIdentifier)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_dms_replication_instance", "aws", dmsAllowEmptyValues))
		}
	}

	// aws_dms_replication_subnet_group, ID = subnet group identifier.
	sg := dms.NewDescribeReplicationSubnetGroupsPaginator(svc, &dms.DescribeReplicationSubnetGroupsInput{})
	for sg.HasMorePages() {
		page, err := sg.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, group := range page.ReplicationSubnetGroups {
			name := StringValue(group.ReplicationSubnetGroupIdentifier)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_dms_replication_subnet_group", "aws", dmsAllowEmptyValues))
		}
	}

	// aws_dms_replication_task, ID = replication task identifier.
	rt := dms.NewDescribeReplicationTasksPaginator(svc, &dms.DescribeReplicationTasksInput{})
	for rt.HasMorePages() {
		page, err := rt.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, task := range page.ReplicationTasks {
			name := StringValue(task.ReplicationTaskIdentifier)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_dms_replication_task", "aws", dmsAllowEmptyValues))
		}
	}

	return nil
}
