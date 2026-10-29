package aws

// Sweep additions for ECS:
// Implemented: aws_ecs_capacity_provider (DescribeCapacityProviders),
// aws_ecs_cluster_capacity_providers (per-cluster association, ID = cluster name),
// aws_ecs_task_set (DescribeTaskSets per cluster/service; import ID
// "<task-set-id>,<service-arn>,<cluster-arn>").
// Also implemented: aws_ecs_daemon (ListDaemons per cluster, ID = daemon
// ARN), aws_ecs_daemon_task_definition (ListDaemonTaskDefinitions, ID =
// task definition ARN), aws_ecs_express_gateway_service
// (DescribeExpressGatewayService probe per service, ID = service ARN).
// Skipped: aws_ecs_account_setting_default (singleton account settings, no
// independent resource identity), aws_ecs_tag (tag glue on other resources).

import (
	"context"
	"fmt"
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

func (g *EcsGenerator) loadSweep(svc *ecs.Client) error {
	// Capacity providers
	var nextToken *string
	for {
		out, err := svc.DescribeCapacityProviders(context.TODO(), &ecs.DescribeCapacityProvidersInput{
			NextToken: nextToken,
		})
		if err != nil {
			return err
		}
		for _, cp := range out.CapacityProviders {
			arn := StringValue(cp.CapacityProviderArn)
			name := StringValue(cp.Name)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, name, "aws_ecs_capacity_provider", "aws", ecsAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}

	// Per-cluster: cluster capacity provider associations + task sets
	p := ecs.NewListClustersPaginator(svc, &ecs.ListClustersInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, clusterArn := range page.ClusterArns {
			parts := strings.Split(clusterArn, "/")
			clusterName := parts[len(parts)-1]
			desc, err := svc.DescribeClusters(context.TODO(), &ecs.DescribeClustersInput{
				Clusters: []string{clusterArn},
			})
			if err == nil && len(desc.Clusters) > 0 && len(desc.Clusters[0].CapacityProviders) > 0 {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					clusterName, clusterName, "aws_ecs_cluster_capacity_providers", "aws", ecsAllowEmptyValues))
			}
			// Daemons in this cluster
			var dToken *string
			for {
				dout, err := svc.ListDaemons(context.TODO(), &ecs.ListDaemonsInput{
					ClusterArn: &clusterArn, NextToken: dToken,
				})
				if err != nil {
					break
				}
				for _, daemon := range dout.DaemonSummariesList {
					arn := StringValue(daemon.DaemonArn)
					parts := strings.Split(arn, "/")
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						arn, parts[len(parts)-1], "aws_ecs_daemon", "aws", ecsAllowEmptyValues))
				}
				dToken = dout.NextToken
				if dToken == nil {
					break
				}
			}
			sp := ecs.NewListServicesPaginator(svc, &ecs.ListServicesInput{Cluster: &clusterArn})
			for sp.HasMorePages() {
				spage, err := sp.NextPage(context.TODO())
				if err != nil {
					break
				}
				for _, serviceArn := range spage.ServiceArns {
					if _, err := svc.DescribeExpressGatewayService(context.TODO(), &ecs.DescribeExpressGatewayServiceInput{
						ServiceArn: &serviceArn,
					}); err == nil {
						parts := strings.Split(serviceArn, "/")
						g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
							serviceArn, parts[len(parts)-1], "aws_ecs_express_gateway_service", "aws", ecsAllowEmptyValues))
					}
					ts, err := svc.DescribeTaskSets(context.TODO(), &ecs.DescribeTaskSetsInput{
						Cluster: &clusterArn, Service: &serviceArn,
					})
					if err != nil {
						continue
					}
					for _, taskSet := range ts.TaskSets {
						id := fmt.Sprintf("%s,%s,%s", StringValue(taskSet.Id), serviceArn, clusterArn)
						g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
							id, StringValue(taskSet.Id), "aws_ecs_task_set", "aws", ecsAllowEmptyValues))
					}
				}
			}
		}
	}
	// Daemon task definitions (account level; the pinned ecs SDK ships
	// no paginator for this operation, so paging is manual)
	var dtToken *string
	for {
		page, err := svc.ListDaemonTaskDefinitions(context.TODO(), &ecs.ListDaemonTaskDefinitionsInput{NextToken: dtToken})
		if err != nil {
			return err
		}
		for _, td := range page.DaemonTaskDefinitions {
			arn := StringValue(td.Arn)
			parts := strings.Split(arn, "/")
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, parts[len(parts)-1], "aws_ecs_daemon_task_definition", "aws", ecsAllowEmptyValues))
		}
		dtToken = page.NextToken
		if dtToken == nil {
			break
		}
	}
	return nil
}
