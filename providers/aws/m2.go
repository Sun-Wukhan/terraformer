package aws

// AWS Mainframe Modernization sweep (new service):
// Implemented: aws_m2_application (ListApplications, ID = application ID),
// aws_m2_environment (ListEnvironments, ID = environment ID).
// Skipped: aws_m2_deployment (deployment operation record per application,
// not a standing configuration resource).

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/m2"
)

var m2AllowEmptyValues = []string{"tags."}

type M2Generator struct {
	AWSService
}

func (g *M2Generator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := m2.NewFromConfig(config)

	p := m2.NewListApplicationsPaginator(svc, &m2.ListApplicationsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, app := range page.Applications {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(app.ApplicationId), StringValue(app.Name), "aws_m2_application", "aws", m2AllowEmptyValues))
		}
	}
	p2 := m2.NewListEnvironmentsPaginator(svc, &m2.ListEnvironmentsInput{})
	for p2.HasMorePages() {
		page, err := p2.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, env := range page.Environments {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(env.EnvironmentId), StringValue(env.Name), "aws_m2_environment", "aws", m2AllowEmptyValues))
		}
	}
	return nil
}
