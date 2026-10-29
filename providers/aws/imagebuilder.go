package aws

// EC2 Image Builder sweep (new service): all types enumerated with the
// List* APIs; import ID for every type is its ARN (verified in provider
// docs).

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/imagebuilder"
)

var imagebuilderAllowEmptyValues = []string{"tags."}

type ImageBuilderGenerator struct {
	AWSService
}

func (g *ImageBuilderGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := imagebuilder.NewFromConfig(config)

	p := imagebuilder.NewListComponentsPaginator(svc, &imagebuilder.ListComponentsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, item := range page.ComponentVersionList {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(item.Arn), StringValue(item.Name), "aws_imagebuilder_component", "aws", imagebuilderAllowEmptyValues))
		}
	}
	p2 := imagebuilder.NewListContainerRecipesPaginator(svc, &imagebuilder.ListContainerRecipesInput{})
	for p2.HasMorePages() {
		page, err := p2.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, item := range page.ContainerRecipeSummaryList {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(item.Arn), StringValue(item.Name), "aws_imagebuilder_container_recipe", "aws", imagebuilderAllowEmptyValues))
		}
	}
	p3 := imagebuilder.NewListDistributionConfigurationsPaginator(svc, &imagebuilder.ListDistributionConfigurationsInput{})
	for p3.HasMorePages() {
		page, err := p3.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, item := range page.DistributionConfigurationSummaryList {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(item.Arn), StringValue(item.Name), "aws_imagebuilder_distribution_configuration", "aws", imagebuilderAllowEmptyValues))
		}
	}
	p4 := imagebuilder.NewListImagesPaginator(svc, &imagebuilder.ListImagesInput{})
	for p4.HasMorePages() {
		page, err := p4.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, item := range page.ImageVersionList {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(item.Arn), StringValue(item.Name), "aws_imagebuilder_image", "aws", imagebuilderAllowEmptyValues))
		}
	}
	p5 := imagebuilder.NewListImagePipelinesPaginator(svc, &imagebuilder.ListImagePipelinesInput{})
	for p5.HasMorePages() {
		page, err := p5.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, item := range page.ImagePipelineList {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(item.Arn), StringValue(item.Name), "aws_imagebuilder_image_pipeline", "aws", imagebuilderAllowEmptyValues))
		}
	}
	p6 := imagebuilder.NewListImageRecipesPaginator(svc, &imagebuilder.ListImageRecipesInput{})
	for p6.HasMorePages() {
		page, err := p6.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, item := range page.ImageRecipeSummaryList {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(item.Arn), StringValue(item.Name), "aws_imagebuilder_image_recipe", "aws", imagebuilderAllowEmptyValues))
		}
	}
	p7 := imagebuilder.NewListInfrastructureConfigurationsPaginator(svc, &imagebuilder.ListInfrastructureConfigurationsInput{})
	for p7.HasMorePages() {
		page, err := p7.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, item := range page.InfrastructureConfigurationSummaryList {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(item.Arn), StringValue(item.Name), "aws_imagebuilder_infrastructure_configuration", "aws", imagebuilderAllowEmptyValues))
		}
	}
	p8 := imagebuilder.NewListLifecyclePoliciesPaginator(svc, &imagebuilder.ListLifecyclePoliciesInput{})
	for p8.HasMorePages() {
		page, err := p8.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, item := range page.LifecyclePolicySummaryList {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(item.Arn), StringValue(item.Name), "aws_imagebuilder_lifecycle_policy", "aws", imagebuilderAllowEmptyValues))
		}
	}
	p9 := imagebuilder.NewListWorkflowsPaginator(svc, &imagebuilder.ListWorkflowsInput{})
	for p9.HasMorePages() {
		page, err := p9.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, item := range page.WorkflowVersionList {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(item.Arn), StringValue(item.Name), "aws_imagebuilder_workflow", "aws", imagebuilderAllowEmptyValues))
		}
	}
	return nil
}
