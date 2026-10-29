package aws

// App Runner sweep (new service):
// Implemented: aws_apprunner_service (ListServices, ID = service ARN),
// aws_apprunner_connection (ListConnections, ID = connection name),
// aws_apprunner_auto_scaling_configuration_version
// (ListAutoScalingConfigurations, ID = configuration ARN),
// aws_apprunner_observability_configuration
// (ListObservabilityConfigurations, ID = configuration ARN),
// aws_apprunner_vpc_connector (ListVpcConnectors, ID = connector ARN),
// aws_apprunner_vpc_ingress_connection (ListVpcIngressConnections,
// ID = connection ARN), aws_apprunner_custom_domain_association
// (DescribeCustomDomains per service, ID = "<domain>,<service-arn>").
// Skipped: aws_apprunner_deployment (deployment operation record, provider
// docs have no Import section), aws_apprunner_default_auto_scaling_configuration_version
// (singleton per-region default pointer).

import (
	"context"
	"fmt"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/apprunner"
)

var apprunnerAllowEmptyValues = []string{"tags."}

type AppRunnerGenerator struct {
	AWSService
}

func (g *AppRunnerGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := apprunner.NewFromConfig(config)

	var serviceArns []string
	var nextToken *string
	for {
		out, err := svc.ListServices(context.TODO(), &apprunner.ListServicesInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, s := range out.ServiceSummaryList {
			arn := StringValue(s.ServiceArn)
			serviceArns = append(serviceArns, arn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(s.ServiceName), "aws_apprunner_service", "aws", apprunnerAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	for _, arn := range serviceArns {
		var nt *string
		for {
			out, err := svc.DescribeCustomDomains(context.TODO(), &apprunner.DescribeCustomDomainsInput{
				ServiceArn: &arn, NextToken: nt,
			})
			if err != nil {
				break
			}
			for _, d := range out.CustomDomains {
				id := fmt.Sprintf("%s,%s", StringValue(d.DomainName), arn)
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, StringValue(d.DomainName), "aws_apprunner_custom_domain_association", "aws", apprunnerAllowEmptyValues))
			}
			nt = out.NextToken
			if nt == nil {
				break
			}
		}
	}

	nextToken = nil
	for {
		out, err := svc.ListConnections(context.TODO(), &apprunner.ListConnectionsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, c := range out.ConnectionSummaryList {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(c.ConnectionName), StringValue(c.ConnectionName), "aws_apprunner_connection", "aws", apprunnerAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	nextToken = nil
	for {
		out, err := svc.ListAutoScalingConfigurations(context.TODO(), &apprunner.ListAutoScalingConfigurationsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, c := range out.AutoScalingConfigurationSummaryList {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(c.AutoScalingConfigurationArn), StringValue(c.AutoScalingConfigurationName), "aws_apprunner_auto_scaling_configuration_version", "aws", apprunnerAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	nextToken = nil
	for {
		out, err := svc.ListObservabilityConfigurations(context.TODO(), &apprunner.ListObservabilityConfigurationsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, c := range out.ObservabilityConfigurationSummaryList {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(c.ObservabilityConfigurationArn), StringValue(c.ObservabilityConfigurationName), "aws_apprunner_observability_configuration", "aws", apprunnerAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	nextToken = nil
	for {
		out, err := svc.ListVpcConnectors(context.TODO(), &apprunner.ListVpcConnectorsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, c := range out.VpcConnectors {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(c.VpcConnectorArn), StringValue(c.VpcConnectorName), "aws_apprunner_vpc_connector", "aws", apprunnerAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	nextToken = nil
	for {
		out, err := svc.ListVpcIngressConnections(context.TODO(), &apprunner.ListVpcIngressConnectionsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, c := range out.VpcIngressConnectionSummaryList {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(c.VpcIngressConnectionArn), StringValue(c.VpcIngressConnectionArn), "aws_apprunner_vpc_ingress_connection", "aws", apprunnerAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
