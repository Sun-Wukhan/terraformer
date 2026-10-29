package aws

// Application Auto Scaling sweep (new service):
// Implemented: aws_appautoscaling_target (DescribeScalableTargets across
// service namespaces; import identity service_namespace + resource_id +
// scalable_dimension) and aws_appautoscaling_policy
// (DescribeScalingPolicies per target, same identity plus policy name).
// Skipped: aws_appautoscaling_scheduled_action (provider docs have no
// Import section / no importer), aws_autoscalingplans_scaling_plan is
// handled in autoscalingplans_sweep.go under the auto_scaling service.

import (
	"context"
	"fmt"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling/types"
)

var appautoscalingAllowEmptyValues = []string{"tags."}

type AppAutoScalingGenerator struct {
	AWSService
}

var appAutoScalingNamespaces = []types.ServiceNamespace{
	types.ServiceNamespaceEcs, types.ServiceNamespaceEmr,
	types.ServiceNamespaceEc2, types.ServiceNamespaceAppstream,
	types.ServiceNamespaceDynamodb, types.ServiceNamespaceRds,
	types.ServiceNamespaceSagemaker, types.ServiceNamespaceCustomResource,
	types.ServiceNamespaceComprehend, types.ServiceNamespaceLambda,
	types.ServiceNamespaceCassandra, types.ServiceNamespaceKafka,
	types.ServiceNamespaceElasticache, types.ServiceNamespaceNeptune,
	types.ServiceNamespaceWorkspaces,
}

func (g *AppAutoScalingGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := applicationautoscaling.NewFromConfig(config)
	for _, ns := range appAutoScalingNamespaces {
		p := applicationautoscaling.NewDescribeScalableTargetsPaginator(svc, &applicationautoscaling.DescribeScalableTargetsInput{
			ServiceNamespace: ns,
		})
		for p.HasMorePages() {
			page, err := p.NextPage(context.TODO())
			if err != nil {
				break
			}
			for _, target := range page.ScalableTargets {
				targetID := fmt.Sprintf("%s/%s/%s", string(ns), StringValue(target.ResourceId), string(target.ScalableDimension))
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					targetID, StringValue(target.ResourceId), "aws_appautoscaling_target", "aws", appautoscalingAllowEmptyValues))
				pp := applicationautoscaling.NewDescribeScalingPoliciesPaginator(svc, &applicationautoscaling.DescribeScalingPoliciesInput{
					ServiceNamespace:  ns,
					ResourceId:        target.ResourceId,
					ScalableDimension: target.ScalableDimension,
				})
				for pp.HasMorePages() {
					ppage, err := pp.NextPage(context.TODO())
					if err != nil {
						break
					}
					for _, pol := range ppage.ScalingPolicies {
						polID := fmt.Sprintf("%s/%s", targetID, StringValue(pol.PolicyName))
						g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
							polID, StringValue(pol.PolicyName), "aws_appautoscaling_policy", "aws", appautoscalingAllowEmptyValues))
					}
				}
			}
		}
	}
	return nil
}
