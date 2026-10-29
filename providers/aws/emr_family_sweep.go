package aws

// Sweep additions for the EMR family, folded into the existing emr
// generator: aws_emrcontainers_virtual_cluster (ListVirtualClusters,
// ID = cluster ID), aws_emrcontainers_job_template (ListJobTemplates,
// ID = template ID), aws_emrserverless_application (ListApplications,
// ID = application ID).
// Skipped: aws_serverlessapplicationrepository_cloudformation_stack (a
// CloudFormation stack created by the Serverless Application Repository;
// the SAR API exposes no list operation for deployed stacks, so there is
// no independent enumeration distinct from CloudFormation itself).

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/emrcontainers"
	"github.com/aws/aws-sdk-go-v2/service/emrserverless"
)

func (g *EmrGenerator) loadFamilySweep() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	containersSvc := emrcontainers.NewFromConfig(config)
	p := emrcontainers.NewListVirtualClustersPaginator(containersSvc, &emrcontainers.ListVirtualClustersInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, vc := range page.VirtualClusters {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(vc.Id), StringValue(vc.Name), "aws_emrcontainers_virtual_cluster", "aws", emrAllowEmptyValues))
		}
	}
	p2 := emrcontainers.NewListJobTemplatesPaginator(containersSvc, &emrcontainers.ListJobTemplatesInput{})
	for p2.HasMorePages() {
		page, err := p2.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, tmpl := range page.Templates {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(tmpl.Id), StringValue(tmpl.Name), "aws_emrcontainers_job_template", "aws", emrAllowEmptyValues))
		}
	}
	serverlessSvc := emrserverless.NewFromConfig(config)
	p3 := emrserverless.NewListApplicationsPaginator(serverlessSvc, &emrserverless.ListApplicationsInput{})
	for p3.HasMorePages() {
		page, err := p3.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, app := range page.Applications {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(app.Id), StringValue(app.Name), "aws_emrserverless_application", "aws", emrAllowEmptyValues))
		}
	}
	return nil
}
