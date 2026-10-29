package aws

// Sweep additions for CodeBuild:
// Implemented: aws_codebuild_fleet (ListFleets returns fleet names, then
// BatchGetFleets for ARNs, ID = fleet ARN), aws_codebuild_report_group
// (ListReportGroups returns ARNs, then BatchGetReportGroups for names,
// ID = report group ARN), aws_codebuild_webhook (projects whose
// BatchGetProjects entry has a webhook, ID = project name),
// aws_codebuild_source_credential (ListSourceCredentials, ID = credential
// ARN). The codebuild SDK ships no paginators, so paging is manual.
// Skipped: aws_codebuild_resource_policy (resource-policy glue; only a
// per-ARN GetResourcePolicy exists, no enumeration).

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/codebuild"
)

func (g *CodeBuildGenerator) loadSweep(svc *codebuild.Client) error {
	// Fleets: ListFleets yields names only; resolve ARNs via BatchGetFleets.
	var fleetNames []string
	var nextToken *string
	for {
		out, err := svc.ListFleets(context.TODO(), &codebuild.ListFleetsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		fleetNames = append(fleetNames, out.Fleets...)
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	if len(fleetNames) > 0 {
		out, err := svc.BatchGetFleets(context.TODO(), &codebuild.BatchGetFleetsInput{Names: fleetNames})
		if err != nil {
			return err
		}
		for _, fleet := range out.Fleets {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(fleet.Arn), StringValue(fleet.Name), "aws_codebuild_fleet", "aws", codebuildAllowEmptyValues))
		}
	}
	// Report groups: ListReportGroups yields ARNs; resolve names.
	var rgArns []string
	nextToken = nil
	for {
		out, err := svc.ListReportGroups(context.TODO(), &codebuild.ListReportGroupsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		rgArns = append(rgArns, out.ReportGroups...)
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	if len(rgArns) > 0 {
		out, err := svc.BatchGetReportGroups(context.TODO(), &codebuild.BatchGetReportGroupsInput{ReportGroupArns: rgArns})
		if err != nil {
			return err
		}
		for _, rg := range out.ReportGroups {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(rg.Arn), StringValue(rg.Name), "aws_codebuild_report_group", "aws", codebuildAllowEmptyValues))
		}
	}
	// Source credentials: one unfiltered List call covers all server types.
	if out, err := svc.ListSourceCredentials(context.TODO(), &codebuild.ListSourceCredentialsInput{}); err == nil {
		for _, cred := range out.SourceCredentialsInfos {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(cred.Arn), StringValue(cred.Arn), "aws_codebuild_source_credential", "aws", codebuildAllowEmptyValues))
		}
	}
	// Webhooks: one per project that has one configured.
	nextToken = nil
	for {
		page, err := svc.ListProjects(context.TODO(), &codebuild.ListProjectsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		if len(page.Projects) > 0 {
			projects, err := svc.BatchGetProjects(context.TODO(), &codebuild.BatchGetProjectsInput{Names: page.Projects})
			if err == nil {
				for _, project := range projects.Projects {
					if project.Webhook != nil {
						name := StringValue(project.Name)
						g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
							name, name, "aws_codebuild_webhook", "aws", codebuildAllowEmptyValues))
					}
				}
			}
		}
		nextToken = page.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
