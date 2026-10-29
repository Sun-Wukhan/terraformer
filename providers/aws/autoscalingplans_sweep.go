package aws

// Sweep addition for Auto Scaling Plans, folded into the auto_scaling
// generator: aws_autoscalingplans_scaling_plan (DescribeScalingPlans,
// ID = plan name). The pinned autoscalingplans SDK ships no paginator
// for DescribeScalingPlans, so paging is manual.

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/autoscalingplans"
)

func (g *AutoScalingGenerator) loadPlansSweep() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := autoscalingplans.NewFromConfig(config)
	var nextToken *string
	for {
		out, err := svc.DescribeScalingPlans(context.TODO(), &autoscalingplans.DescribeScalingPlansInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, plan := range out.ScalingPlans {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(plan.ScalingPlanName), StringValue(plan.ScalingPlanName), "aws_autoscalingplans_scaling_plan", "aws", AsgAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
