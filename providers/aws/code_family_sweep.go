package aws

// Sweep additions for the remaining Code* services (new generators):
// aws_codeartifact_domain / aws_codeartifact_repository (import ID = ARN),
// aws_codeconnections_connection / aws_codeconnections_host (ID = ARN),
// aws_codestarnotifications_notification_rule (ListNotificationRules,
// ID = rule ARN), aws_codeguruprofiler_profiling_group
// (ListProfilingGroups, ID = group name).
// Skipped: aws_codeartifact_domain_permissions_policy and
// aws_codeartifact_repository_permissions_policy (policy glue; only a
// per-parent GetDomainPermissionsPolicy / GetRepositoryPermissionsPolicy
// exists), aws_codestarconnections_connection / aws_codestarconnections_host
// (deprecated legacy aliases of the codeconnections resources over the
// same underlying connections/hosts), aws_codegurureviewer_repository_association
// (provider docs have no Import section / no importer),
// aws_codecatalyst_dev_environment / aws_codecatalyst_project /
// aws_codecatalyst_source_repository (CodeCatalyst is closed to new
// customers and has no aws-sdk-go-v2 enumeration path used by this sweep;
// dev environment additionally has no provider importer).

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/codeartifact"
	"github.com/aws/aws-sdk-go-v2/service/codeconnections"
	"github.com/aws/aws-sdk-go-v2/service/codeguruprofiler"
	"github.com/aws/aws-sdk-go-v2/service/codestarnotifications"
)

var codeFamilyAllowEmptyValues = []string{"tags."}

// --- CodeArtifact ---

type CodeArtifactGenerator struct {
	AWSService
}

func (g *CodeArtifactGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := codeartifact.NewFromConfig(config)
	p := codeartifact.NewListDomainsPaginator(svc, &codeartifact.ListDomainsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, domain := range page.Domains {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(domain.Arn), StringValue(domain.Name), "aws_codeartifact_domain", "aws", codeFamilyAllowEmptyValues))
			rp := codeartifact.NewListRepositoriesInDomainPaginator(svc, &codeartifact.ListRepositoriesInDomainInput{
				Domain: domain.Name,
			})
			for rp.HasMorePages() {
				rpage, err := rp.NextPage(context.TODO())
				if err != nil {
					break
				}
				for _, repo := range rpage.Repositories {
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						StringValue(repo.Arn), StringValue(repo.Name), "aws_codeartifact_repository", "aws", codeFamilyAllowEmptyValues))
				}
			}
		}
	}
	return nil
}

// --- CodeConnections ---

type CodeConnectionsGenerator struct {
	AWSService
}

func (g *CodeConnectionsGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := codeconnections.NewFromConfig(config)
	p := codeconnections.NewListConnectionsPaginator(svc, &codeconnections.ListConnectionsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, conn := range page.Connections {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(conn.ConnectionArn), StringValue(conn.ConnectionName), "aws_codeconnections_connection", "aws", codeFamilyAllowEmptyValues))
		}
	}
	p2 := codeconnections.NewListHostsPaginator(svc, &codeconnections.ListHostsInput{})
	for p2.HasMorePages() {
		page, err := p2.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, host := range page.Hosts {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(host.HostArn), StringValue(host.Name), "aws_codeconnections_host", "aws", codeFamilyAllowEmptyValues))
		}
	}
	return nil
}

// --- CodeStar Notifications ---

type CodeStarNotificationsGenerator struct {
	AWSService
}

func (g *CodeStarNotificationsGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := codestarnotifications.NewFromConfig(config)
	p := codestarnotifications.NewListNotificationRulesPaginator(svc, &codestarnotifications.ListNotificationRulesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, rule := range page.NotificationRules {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(rule.Arn), StringValue(rule.Id), "aws_codestarnotifications_notification_rule", "aws", codeFamilyAllowEmptyValues))
		}
	}
	return nil
}

// --- CodeGuru Profiler ---

type CodeGuruProfilerGenerator struct {
	AWSService
}

func (g *CodeGuruProfilerGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := codeguruprofiler.NewFromConfig(config)
	p := codeguruprofiler.NewListProfilingGroupsPaginator(svc, &codeguruprofiler.ListProfilingGroupsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, name := range page.ProfilingGroupNames {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_codeguruprofiler_profiling_group", "aws", codeFamilyAllowEmptyValues))
		}
	}
	return nil
}
