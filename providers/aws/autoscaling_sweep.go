package aws

// Sweep additions for Auto Scaling:
// Implemented per Auto Scaling group: aws_autoscaling_policy
// (DescribePolicies), aws_autoscaling_schedule (DescribeScheduledActions),
// aws_autoscaling_lifecycle_hook (DescribeLifecycleHooks). Import identity
// for all three is (autoscaling_group_name, name).
// Skipped: aws_autoscaling_attachment and aws_autoscaling_notification
// (provider docs have no Import section / no importer),
// aws_autoscaling_group_tag (tag glue on the ASG),
// aws_autoscaling_traffic_source_attachment (association glue between an
// ASG and a traffic source, no independent enumeration).

import (
	"context"
	"fmt"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
)

func (g *AutoScalingGenerator) loadSweep(svc *autoscaling.Client) error {
	p := autoscaling.NewDescribeAutoScalingGroupsPaginator(svc, &autoscaling.DescribeAutoScalingGroupsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, asg := range page.AutoScalingGroups {
			asgName := StringValue(asg.AutoScalingGroupName)
			pp := autoscaling.NewDescribePoliciesPaginator(svc, &autoscaling.DescribePoliciesInput{AutoScalingGroupName: &asgName})
			for pp.HasMorePages() {
				pg, err := pp.NextPage(context.TODO())
				if err != nil {
					break
				}
				for _, pol := range pg.ScalingPolicies {
					id := fmt.Sprintf("%s/%s", asgName, StringValue(pol.PolicyName))
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						id, StringValue(pol.PolicyName), "aws_autoscaling_policy", "aws", AsgAllowEmptyValues))
				}
			}
			sp := autoscaling.NewDescribeScheduledActionsPaginator(svc, &autoscaling.DescribeScheduledActionsInput{AutoScalingGroupName: &asgName})
			for sp.HasMorePages() {
				pg, err := sp.NextPage(context.TODO())
				if err != nil {
					break
				}
				for _, act := range pg.ScheduledUpdateGroupActions {
					id := fmt.Sprintf("%s/%s", asgName, StringValue(act.ScheduledActionName))
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						id, StringValue(act.ScheduledActionName), "aws_autoscaling_schedule", "aws", AsgAllowEmptyValues))
				}
			}
			hooks, err := svc.DescribeLifecycleHooks(context.TODO(), &autoscaling.DescribeLifecycleHooksInput{AutoScalingGroupName: &asgName})
			if err == nil {
				for _, hook := range hooks.LifecycleHooks {
					id := fmt.Sprintf("%s/%s", asgName, StringValue(hook.LifecycleHookName))
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						id, StringValue(hook.LifecycleHookName), "aws_autoscaling_lifecycle_hook", "aws", AsgAllowEmptyValues))
				}
			}
		}
	}
	return nil
}
