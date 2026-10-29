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

// Redshift depth coverage (AWS sweep, database cluster).
//
// Skipped in this family, with reasons:
//   - aws_redshift_logging: singleton per-cluster settings object (bucket /
//     log exports of one cluster), not an independent resource.
//   - aws_redshift_snapshot_copy: singleton per-cluster settings object
//     (cross-region copy configuration of one cluster).
//   - aws_redshift_resource_policy: policies attach to individual resources
//     (snapshots, datashares, ...); the SDK exposes only per-resource
//     GetResourcePolicy with no list/enumeration API.
//   - aws_redshift_namespace_registration: no list/describe API in the
//     Redshift SDK (RegisterNamespace/DeregisterNamespace only), so existing
//     registrations cannot be enumerated.

import (
	"context"
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
)

func (g *RedshiftGenerator) loadRedshiftDepth(svc *redshift.Client, config aws.Config) error {
	// aws_redshift_authentication_profile, ID = profile name.
	{
		out, err := svc.DescribeAuthenticationProfiles(context.TODO(), &redshift.DescribeAuthenticationProfilesInput{})
		if err != nil {
			return err
		}
		for _, profile := range out.AuthenticationProfiles {
			name := StringValue(profile.AuthenticationProfileName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_redshift_authentication_profile", "aws", RedshiftAllowEmptyValues))
		}
	}

	// aws_redshift_cluster_snapshot, ID = snapshot identifier.
	sp := redshift.NewDescribeClusterSnapshotsPaginator(svc, &redshift.DescribeClusterSnapshotsInput{})
	for sp.HasMorePages() {
		page, err := sp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, snap := range page.Snapshots {
			name := StringValue(snap.SnapshotIdentifier)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_redshift_cluster_snapshot", "aws", RedshiftAllowEmptyValues))
		}
	}

	// Per-cluster depth: IAM role associations, partners.
	account, err := g.getAccountNumber(config)
	if err != nil {
		return err
	}
	cp := redshift.NewDescribeClustersPaginator(svc, &redshift.DescribeClustersInput{})
	for cp.HasMorePages() {
		page, err := cp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, cluster := range page.Clusters {
			clusterID := StringValue(cluster.ClusterIdentifier)

			// aws_redshift_cluster_iam_roles, ID = cluster identifier;
			// one resource per cluster that has IAM roles associated.
			if len(cluster.IamRoles) > 0 {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					clusterID, clusterID, "aws_redshift_cluster_iam_roles", "aws", RedshiftAllowEmptyValues))
			}

			// aws_redshift_partner, ID = "<account>:<cluster>:<database>:<partner>".
			dbName := StringValue(cluster.DBName)
			if dbName != "" {
				partners, err := svc.DescribePartners(context.TODO(), &redshift.DescribePartnersInput{
					AccountId:         account,
					ClusterIdentifier: aws.String(clusterID),
					DatabaseName:      aws.String(dbName),
				})
				if err != nil {
					return err
				}
				for _, partner := range partners.PartnerIntegrationInfoList {
					partnerName := StringValue(partner.PartnerName)
					id := StringValue(account) + ":" + clusterID + ":" + dbName + ":" + partnerName
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						id, clusterID+"_"+partnerName, "aws_redshift_partner", "aws", RedshiftAllowEmptyValues))
				}
			}
		}
	}

	// aws_redshift_data_share_authorization (producer side),
	// ID = "<datashare ARN>,<consumer identifier>".
	dsp := redshift.NewDescribeDataSharesPaginator(svc, &redshift.DescribeDataSharesInput{})
	for dsp.HasMorePages() {
		page, err := dsp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, share := range page.DataShares {
			shareARN := StringValue(share.DataShareArn)
			for _, assoc := range share.DataShareAssociations {
				consumer := StringValue(assoc.ConsumerIdentifier)
				if consumer == "" {
					continue
				}
				id := shareARN + "," + consumer
				name := shareARN + "_" + consumer
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, name, "aws_redshift_data_share_authorization", "aws", RedshiftAllowEmptyValues))
			}
		}
	}

	// aws_redshift_data_share_consumer_association (consumer side),
	// ID = "<datashare ARN>,<associate_entire_account>,<consumer ARN>,<consumer region>".
	csp := redshift.NewDescribeDataSharesForConsumerPaginator(svc, &redshift.DescribeDataSharesForConsumerInput{})
	for csp.HasMorePages() {
		page, err := csp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, share := range page.DataShares {
			shareARN := StringValue(share.DataShareArn)
			for _, assoc := range share.DataShareAssociations {
				consumer := StringValue(assoc.ConsumerIdentifier)
				consumerRegion := StringValue(assoc.ConsumerRegion)
				var id string
				switch {
				case strings.HasPrefix(consumer, "arn:"):
					id = shareARN + ",," + consumer + ","
				case consumerRegion != "":
					id = shareARN + ",,," + consumerRegion
				case consumer != "":
					id = shareARN + ",true,,"
				default:
					continue
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, shareARN+"_"+consumer, "aws_redshift_data_share_consumer_association", "aws", RedshiftAllowEmptyValues))
			}
		}
	}

	// aws_redshift_endpoint_access, ID = endpoint name.
	eap := redshift.NewDescribeEndpointAccessPaginator(svc, &redshift.DescribeEndpointAccessInput{})
	for eap.HasMorePages() {
		page, err := eap.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, endpoint := range page.EndpointAccessList {
			name := StringValue(endpoint.EndpointName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_redshift_endpoint_access", "aws", RedshiftAllowEmptyValues))
		}
	}

	// aws_redshift_endpoint_authorization, ID = "<cluster identifier>,<grantee account>".
	eauth, err := svc.DescribeEndpointAuthorization(context.TODO(), &redshift.DescribeEndpointAuthorizationInput{})
	if err != nil {
		return err
	}
	for _, auth := range eauth.EndpointAuthorizationList {
		clusterID := StringValue(auth.ClusterIdentifier)
		grantee := StringValue(auth.Grantee)
		id := clusterID + "," + grantee
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			id, clusterID+"_"+grantee, "aws_redshift_endpoint_authorization", "aws", RedshiftAllowEmptyValues))
	}

	// aws_redshift_hsm_client_certificate, ID = certificate identifier.
	hcp := redshift.NewDescribeHsmClientCertificatesPaginator(svc, &redshift.DescribeHsmClientCertificatesInput{})
	for hcp.HasMorePages() {
		page, err := hcp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, cert := range page.HsmClientCertificates {
			name := StringValue(cert.HsmClientCertificateIdentifier)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_redshift_hsm_client_certificate", "aws", RedshiftAllowEmptyValues))
		}
	}

	// aws_redshift_hsm_configuration, ID = configuration identifier.
	hfp := redshift.NewDescribeHsmConfigurationsPaginator(svc, &redshift.DescribeHsmConfigurationsInput{})
	for hfp.HasMorePages() {
		page, err := hfp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, conf := range page.HsmConfigurations {
			name := StringValue(conf.HsmConfigurationIdentifier)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_redshift_hsm_configuration", "aws", RedshiftAllowEmptyValues))
		}
	}

	// aws_redshift_idc_application, ID = application ARN.
	{
		out, err := svc.DescribeRedshiftIdcApplications(context.TODO(), &redshift.DescribeRedshiftIdcApplicationsInput{})
		if err != nil {
			return err
		}
		for _, app := range out.RedshiftIdcApplications {
			arn := StringValue(app.RedshiftIdcApplicationArn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(app.RedshiftIdcApplicationName), "aws_redshift_idc_application", "aws", RedshiftAllowEmptyValues))
		}
	}

	// aws_redshift_integration (inbound zero-ETL), ID = integration ARN.
	iip := redshift.NewDescribeInboundIntegrationsPaginator(svc, &redshift.DescribeInboundIntegrationsInput{})
	for iip.HasMorePages() {
		page, err := iip.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, integration := range page.InboundIntegrations {
			arn := StringValue(integration.IntegrationArn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_redshift_integration", "aws", RedshiftAllowEmptyValues))
		}
	}

	// aws_redshift_scheduled_action, ID = scheduled action name.
	sap := redshift.NewDescribeScheduledActionsPaginator(svc, &redshift.DescribeScheduledActionsInput{})
	for sap.HasMorePages() {
		page, err := sap.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, action := range page.ScheduledActions {
			name := StringValue(action.ScheduledActionName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_redshift_scheduled_action", "aws", RedshiftAllowEmptyValues))
		}
	}

	// aws_redshift_snapshot_copy_grant, ID = grant name.
	gp := redshift.NewDescribeSnapshotCopyGrantsPaginator(svc, &redshift.DescribeSnapshotCopyGrantsInput{})
	for gp.HasMorePages() {
		page, err := gp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, grant := range page.SnapshotCopyGrants {
			name := StringValue(grant.SnapshotCopyGrantName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_redshift_snapshot_copy_grant", "aws", RedshiftAllowEmptyValues))
		}
	}

	// aws_redshift_usage_limit, ID = usage limit ID.
	ulp := redshift.NewDescribeUsageLimitsPaginator(svc, &redshift.DescribeUsageLimitsInput{})
	for ulp.HasMorePages() {
		page, err := ulp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, limit := range page.UsageLimits {
			id := StringValue(limit.UsageLimitId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_redshift_usage_limit", "aws", RedshiftAllowEmptyValues))
		}
	}

	return nil
}
