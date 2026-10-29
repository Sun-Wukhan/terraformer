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

// RDS depth coverage (AWS sweep, database cluster).
//
// Skipped in this family, with reasons:
//   - aws_rds_instance_state: state-control action resource (start/stop) with no
//     independent configuration to import; it merely toggles DB instance state.

import (
	"context"
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"

	"github.com/aws/aws-sdk-go-v2/service/rds"
	rdstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
)

// loadRDSCertificates: aws_rds_certificate, ID = certificate identifier.
func (g *RDSGenerator) loadRDSCertificates(svc *rds.Client) error {
	p := rds.NewDescribeCertificatesPaginator(svc, &rds.DescribeCertificatesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, cert := range page.Certificates {
			name := StringValue(cert.CertificateIdentifier)
			if cert.CustomerOverride == nil || !*cert.CustomerOverride {
				continue // only account-level customer-override certs are managed resources;
				// the rest of DescribeCertificates lists AWS-published CA certs
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_rds_certificate", "aws", RDSAllowEmptyValues))
		}
	}
	return nil
}

// loadRDSClusterParameterGroups: aws_rds_cluster_parameter_group, ID = group name.
func (g *RDSGenerator) loadRDSClusterParameterGroups(svc *rds.Client) error {
	p := rds.NewDescribeDBClusterParameterGroupsPaginator(svc, &rds.DescribeDBClusterParameterGroupsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, pg := range page.DBClusterParameterGroups {
			name := StringValue(pg.DBClusterParameterGroupName)
			if strings.Contains(name, ".") {
				continue // default groups like default.aurora-mysql8.0
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_rds_cluster_parameter_group", "aws", RDSAllowEmptyValues))
		}
	}
	return nil
}

// loadRDSClusterEndpoints: aws_rds_cluster_endpoint, ID = endpoint identifier.
func (g *RDSGenerator) loadRDSClusterEndpoints(svc *rds.Client) error {
	p := rds.NewDescribeDBClusterEndpointsPaginator(svc, &rds.DescribeDBClusterEndpointsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, ep := range page.DBClusterEndpoints {
			name := StringValue(ep.DBClusterEndpointIdentifier)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_rds_cluster_endpoint", "aws", RDSAllowEmptyValues))
		}
	}
	return nil
}

// loadRDSClusterInstancesAndRoles walks clusters and instances once and emits:
//   - aws_rds_cluster_instance (instances that belong to a cluster), ID = instance identifier
//   - aws_rds_cluster_role_association, ID = "<cluster identifier>,<role ARN>"
//   - aws_db_instance_role_association, ID = "<instance identifier>,<role ARN>"
//   - aws_rds_cluster_activity_stream (clusters with an activity stream), ID = cluster identifier
func (g *RDSGenerator) loadRDSClusterInstancesAndRoles(svc *rds.Client) error {
	cp := rds.NewDescribeDBClustersPaginator(svc, &rds.DescribeDBClustersInput{})
	for cp.HasMorePages() {
		page, err := cp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, cluster := range page.DBClusters {
			clusterID := StringValue(cluster.DBClusterIdentifier)
			if cluster.ActivityStreamStatus != "" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					clusterID, clusterID, "aws_rds_cluster_activity_stream", "aws", RDSAllowEmptyValues))
			}
			for _, role := range cluster.AssociatedRoles {
				roleARN := StringValue(role.RoleArn)
				id := clusterID + "," + roleARN
				name := clusterID + "_" + roleNameFromARN(roleARN) + "_" + StringValue(role.FeatureName)
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, name, "aws_rds_cluster_role_association", "aws", RDSAllowEmptyValues))
			}
		}
	}

	ip := rds.NewDescribeDBInstancesPaginator(svc, &rds.DescribeDBInstancesInput{})
	for ip.HasMorePages() {
		page, err := ip.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, inst := range page.DBInstances {
			instID := StringValue(inst.DBInstanceIdentifier)
			if StringValue(inst.DBClusterIdentifier) != "" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					instID, instID, "aws_rds_cluster_instance", "aws", RDSAllowEmptyValues))
			}
			for _, role := range inst.AssociatedRoles {
				roleARN := StringValue(role.RoleArn)
				id := instID + "," + roleARN
				name := instID + "_" + roleNameFromARN(roleARN) + "_" + StringValue(role.FeatureName)
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, name, "aws_db_instance_role_association", "aws", RDSAllowEmptyValues))
			}
		}
	}
	return nil
}

func roleNameFromARN(arn string) string {
	if idx := strings.LastIndex(arn, "/"); idx >= 0 {
		return arn[idx+1:]
	}
	return arn
}

// loadRDSSnapshotCopies emits the snapshot-copy resources for snapshots that
// were themselves produced by a copy (SourceDBSnapshotIdentifier /
// SourceDBClusterSnapshotArn set). ID = target snapshot identifier.
func (g *RDSGenerator) loadRDSSnapshotCopies(svc *rds.Client) error {
	p := rds.NewDescribeDBSnapshotsPaginator(svc, &rds.DescribeDBSnapshotsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, snap := range page.DBSnapshots {
			if StringValue(snap.SourceDBSnapshotIdentifier) == "" {
				continue
			}
			name := StringValue(snap.DBSnapshotIdentifier)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_db_snapshot_copy", "aws", RDSAllowEmptyValues))
		}
	}
	cp := rds.NewDescribeDBClusterSnapshotsPaginator(svc, &rds.DescribeDBClusterSnapshotsInput{})
	for cp.HasMorePages() {
		page, err := cp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, snap := range page.DBClusterSnapshots {
			if StringValue(snap.SourceDBClusterSnapshotArn) == "" {
				continue
			}
			name := StringValue(snap.DBClusterSnapshotIdentifier)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_rds_cluster_snapshot_copy", "aws", RDSAllowEmptyValues))
		}
	}
	return nil
}

// loadRDSCustomEngineVersions: aws_rds_custom_db_engine_version,
// ID = "<engine>/<version>".
func (g *RDSGenerator) loadRDSCustomEngineVersions(svc *rds.Client) error {
	p := rds.NewDescribeDBEngineVersionsPaginator(svc, &rds.DescribeDBEngineVersionsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, v := range page.DBEngineVersions {
			engine := StringValue(v.Engine)
			if !strings.HasPrefix(engine, "custom-") {
				continue
			}
			id := engine + "/" + StringValue(v.EngineVersion)
			name := engine + "_" + StringValue(v.EngineVersion)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, name, "aws_rds_custom_db_engine_version", "aws", RDSAllowEmptyValues))
		}
	}
	return nil
}

// loadRDSExportTasks: aws_rds_export_task, ID = export task identifier.
func (g *RDSGenerator) loadRDSExportTasks(svc *rds.Client) error {
	p := rds.NewDescribeExportTasksPaginator(svc, &rds.DescribeExportTasksInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, t := range page.ExportTasks {
			name := StringValue(t.ExportTaskIdentifier)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_rds_export_task", "aws", RDSAllowEmptyValues))
		}
	}
	return nil
}

// loadRDSIntegrations: aws_rds_integration (zero-ETL), ID = integration ARN.
func (g *RDSGenerator) loadRDSIntegrations(svc *rds.Client) error {
	p := rds.NewDescribeIntegrationsPaginator(svc, &rds.DescribeIntegrationsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, in := range page.Integrations {
			arn := StringValue(in.IntegrationArn)
			name := StringValue(in.IntegrationName)
			if name == "" {
				name = arn
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, name, "aws_rds_integration", "aws", RDSAllowEmptyValues))
		}
	}
	return nil
}

// loadRDSReservedInstances: aws_rds_reserved_instance, ID = reservation ID.
func (g *RDSGenerator) loadRDSReservedInstances(svc *rds.Client) error {
	p := rds.NewDescribeReservedDBInstancesPaginator(svc, &rds.DescribeReservedDBInstancesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, ri := range page.ReservedDBInstances {
			name := StringValue(ri.ReservedDBInstanceId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_rds_reserved_instance", "aws", RDSAllowEmptyValues))
		}
	}
	return nil
}

// loadRDSShardGroups: aws_rds_shard_group (Aurora Limitless),
// ID = shard group identifier.
func (g *RDSGenerator) loadRDSShardGroups(svc *rds.Client) error {
	out, err := svc.DescribeDBShardGroups(context.TODO(), &rds.DescribeDBShardGroupsInput{})
	if err != nil {
		return err
	}
	for _, sg := range out.DBShardGroups {
		name := StringValue(sg.DBShardGroupIdentifier)
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			name, name, "aws_rds_shard_group", "aws", RDSAllowEmptyValues))
	}
	return nil
}

// loadRDSAutomatedBackupsReplication: an entry returned by
// DescribeDBInstanceAutomatedBackups in this region is a backup replicated
// into this region; the provider resource ID is the source DB instance ARN.
func (g *RDSGenerator) loadRDSAutomatedBackupsReplication(svc *rds.Client) error {
	p := rds.NewDescribeDBInstanceAutomatedBackupsPaginator(svc, &rds.DescribeDBInstanceAutomatedBackupsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, b := range page.DBInstanceAutomatedBackups {
			id := StringValue(b.DBInstanceArn)
			name := StringValue(b.DBInstanceIdentifier)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, name, "aws_db_instance_automated_backups_replication", "aws", RDSAllowEmptyValues))
		}
	}
	return nil
}

// loadDBProxyDepth emits, per proxy:
//   - aws_db_proxy_default_target_group, ID = proxy name (one per proxy)
//   - aws_db_proxy_endpoint, ID = "<proxy name>/<endpoint name>"
//   - aws_db_proxy_target, ID = "<proxy name>/<target group name>/<type>/<resource identifier>"
func (g *RDSGenerator) loadDBProxyDepth(svc *rds.Client) error {
	// Map RDS resource IDs (db-XXXX / cluster-XXXX) to identifiers for targets.
	resourceIDToIdentifier := map[string]string{}
	ip := rds.NewDescribeDBInstancesPaginator(svc, &rds.DescribeDBInstancesInput{})
	for ip.HasMorePages() {
		page, err := ip.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, inst := range page.DBInstances {
			resourceIDToIdentifier[StringValue(inst.DbiResourceId)] = StringValue(inst.DBInstanceIdentifier)
		}
	}

	pp := rds.NewDescribeDBProxiesPaginator(svc, &rds.DescribeDBProxiesInput{})
	for pp.HasMorePages() {
		page, err := pp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, proxy := range page.DBProxies {
			proxyName := StringValue(proxy.DBProxyName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				proxyName, proxyName, "aws_db_proxy_default_target_group", "aws", RDSAllowEmptyValues))

			ep, err := svc.DescribeDBProxyEndpoints(context.TODO(), &rds.DescribeDBProxyEndpointsInput{DBProxyName: proxy.DBProxyName})
			if err != nil {
				return err
			}
			for _, endpoint := range ep.DBProxyEndpoints {
				epName := StringValue(endpoint.DBProxyEndpointName)
				id := proxyName + "/" + epName
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, proxyName+"_"+epName, "aws_db_proxy_endpoint", "aws", RDSAllowEmptyValues))
			}

			tg, err := svc.DescribeDBProxyTargetGroups(context.TODO(), &rds.DescribeDBProxyTargetGroupsInput{DBProxyName: proxy.DBProxyName})
			if err != nil {
				return err
			}
			for _, group := range tg.TargetGroups {
				groupName := StringValue(group.TargetGroupName)
				tp := rds.NewDescribeDBProxyTargetsPaginator(svc, &rds.DescribeDBProxyTargetsInput{DBProxyName: proxy.DBProxyName, TargetGroupName: group.TargetGroupName})
				for tp.HasMorePages() {
					tpage, err := tp.NextPage(context.TODO())
					if err != nil {
						return err
					}
					for _, target := range tpage.Targets {
						resourceID := StringValue(target.RdsResourceId)
						if target.Type == rdstypes.TargetTypeTrackedCluster {
							resourceID = StringValue(target.TrackedClusterId)
						} else if ident, ok := resourceIDToIdentifier[resourceID]; ok {
							resourceID = ident
						}
						id := proxyName + "/" + groupName + "/" + string(target.Type) + "/" + resourceID
						name := proxyName + "_" + groupName + "_" + resourceID
						g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
							id, name, "aws_db_proxy_target", "aws", RDSAllowEmptyValues))
					}
				}
			}
		}
	}
	return nil
}

func (g *RDSGenerator) loadRDSDepth(svc *rds.Client) error {
	for _, load := range []func(*rds.Client) error{
		g.loadRDSCertificates,
		g.loadRDSClusterParameterGroups,
		g.loadRDSClusterEndpoints,
		g.loadRDSClusterInstancesAndRoles,
		g.loadRDSSnapshotCopies,
		g.loadRDSCustomEngineVersions,
		g.loadRDSExportTasks,
		g.loadRDSIntegrations,
		g.loadRDSReservedInstances,
		g.loadRDSShardGroups,
		g.loadRDSAutomatedBackupsReplication,
		g.loadDBProxyDepth,
	} {
		if err := load(svc); err != nil {
			return err
		}
	}
	return nil
}
