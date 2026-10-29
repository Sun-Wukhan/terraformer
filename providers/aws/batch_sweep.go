package aws

// Sweep additions for Batch:
// Implemented: aws_batch_scheduling_policy (ListSchedulingPolicies for
// ARNs, then DescribeSchedulingPolicies for names, ID = ARN). The pinned
// batch SDK has no paginator for either operation, so paging is manual.

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/batch"
)

func (g *BatchGenerator) loadSweep(batchClient *batch.Client) error {
	var arns []string
	var nextToken *string
	for {
		out, err := batchClient.ListSchedulingPolicies(context.TODO(), &batch.ListSchedulingPoliciesInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, pol := range out.SchedulingPolicies {
			arns = append(arns, StringValue(pol.Arn))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	if len(arns) == 0 {
		return nil
	}
	out, err := batchClient.DescribeSchedulingPolicies(context.TODO(), &batch.DescribeSchedulingPoliciesInput{Arns: arns})
	if err != nil {
		// Fall back to ARN-only resources if the describe call fails.
		for _, arn := range arns {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_batch_scheduling_policy", "aws", BatchAllowEmptyValues))
		}
		return nil
	}
	for _, pol := range out.SchedulingPolicies {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			StringValue(pol.Arn), StringValue(pol.Name), "aws_batch_scheduling_policy", "aws", BatchAllowEmptyValues))
	}
	return nil
}
