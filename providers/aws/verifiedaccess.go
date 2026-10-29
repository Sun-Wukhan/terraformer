// Copyright 2024 The Terraformer Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aws

import (
	"context"
	"log"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// VerifiedAccessGenerator covers AWS Verified Access: instances,
// endpoints, trust providers, per-instance logging configurations and
// instance/trust-provider attachments. Verified Access is served by the
// EC2 API. Sub-resource failures are logged and skipped.
//
// Notable skip: aws_verifiedaccess_group has no provider importer.
type VerifiedAccessGenerator struct {
	AWSService
}

func (g *VerifiedAccessGenerator) InitResources() error {
	config, err := g.generateConfig()
	if err != nil {
		return err
	}
	svc := ec2.NewFromConfig(config)
	g.getVerifiedAccessInstances(svc)
	g.getVerifiedAccessEndpoints(svc)
	g.getVerifiedAccessTrustProviders(svc)
	return nil
}

func (g *VerifiedAccessGenerator) getVerifiedAccessInstances(svc *ec2.Client) {
	var instanceIDs []string
	p := ec2.NewDescribeVerifiedAccessInstancesPaginator(svc, &ec2.DescribeVerifiedAccessInstancesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, instance := range page.VerifiedAccessInstances {
			if instance.VerifiedAccessInstanceId == nil {
				continue
			}
			instanceID := *instance.VerifiedAccessInstanceId
			instanceIDs = append(instanceIDs, instanceID)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				instanceID,
				instanceID,
				"aws_verifiedaccess_instance",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
			// aws_verifiedaccess_instance_trust_provider_attachment is
			// imported by "<instance id>/<trust provider id>".
			for _, trustProvider := range instance.VerifiedAccessTrustProviders {
				if trustProvider.VerifiedAccessTrustProviderId == nil {
					continue
				}
				id := instanceID + "/" + *trustProvider.VerifiedAccessTrustProviderId
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id,
					id,
					"aws_verifiedaccess_instance_trust_provider_attachment",
					"aws",
					awsNetCoreAllowEmptyValues,
				))
			}
		}
	}
	if len(instanceIDs) == 0 {
		return
	}
	// aws_verifiedaccess_instance_logging_configuration is imported by
	// verified access instance id.
	out, err := svc.DescribeVerifiedAccessInstanceLoggingConfigurations(context.TODO(), &ec2.DescribeVerifiedAccessInstanceLoggingConfigurationsInput{
		VerifiedAccessInstanceIds: instanceIDs,
	})
	if err != nil {
		log.Println(err)
		return
	}
	for _, loggingConfiguration := range out.LoggingConfigurations {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			StringValue(loggingConfiguration.VerifiedAccessInstanceId),
			StringValue(loggingConfiguration.VerifiedAccessInstanceId),
			"aws_verifiedaccess_instance_logging_configuration",
			"aws",
			awsNetCoreAllowEmptyValues,
		))
	}
}

func (g *VerifiedAccessGenerator) getVerifiedAccessEndpoints(svc *ec2.Client) {
	p := ec2.NewDescribeVerifiedAccessEndpointsPaginator(svc, &ec2.DescribeVerifiedAccessEndpointsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, endpoint := range page.VerifiedAccessEndpoints {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(endpoint.VerifiedAccessEndpointId),
				StringValue(endpoint.VerifiedAccessEndpointId),
				"aws_verifiedaccess_endpoint",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
}

func (g *VerifiedAccessGenerator) getVerifiedAccessTrustProviders(svc *ec2.Client) {
	p := ec2.NewDescribeVerifiedAccessTrustProvidersPaginator(svc, &ec2.DescribeVerifiedAccessTrustProvidersInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, trustProvider := range page.VerifiedAccessTrustProviders {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(trustProvider.VerifiedAccessTrustProviderId),
				StringValue(trustProvider.VerifiedAccessTrustProviderId),
				"aws_verifiedaccess_trust_provider",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
}
