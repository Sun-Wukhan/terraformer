package aws

// Sweep additions for WorkSpaces:
// Implemented: aws_workspaces_connection_alias (DescribeConnectionAliases,
// ID = alias ID), aws_workspaces_pool (DescribeWorkspacesPools, ID = pool
// ID). The pinned workspaces SDK ships no paginators for these two
// operations, so paging is manual.

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/workspaces"
)

func (g *WorkspacesGenerator) loadSweep(svc *workspaces.Client) error {
	var nextToken *string
	for {
		page, err := svc.DescribeConnectionAliases(context.TODO(), &workspaces.DescribeConnectionAliasesInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, alias := range page.ConnectionAliases {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(alias.AliasId), StringValue(alias.AliasId), "aws_workspaces_connection_alias", "aws", workspacesAllowEmptyValues))
		}
		nextToken = page.NextToken
		if nextToken == nil {
			break
		}
	}
	nextToken = nil
	for {
		page, err := svc.DescribeWorkspacesPools(context.TODO(), &workspaces.DescribeWorkspacesPoolsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, pool := range page.WorkspacesPools {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(pool.PoolId), StringValue(pool.PoolName), "aws_workspaces_pool", "aws", workspacesAllowEmptyValues))
		}
		nextToken = page.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
