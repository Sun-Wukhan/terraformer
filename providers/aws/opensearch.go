// Copyright 2026 The Terraformer Authors.
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

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/opensearch"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless"
	osstypes "github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
)

var opensearchAllowEmptyValues = []string{"tags."}

type OpenSearchGenerator struct {
	AWSService
}

// InitResources enumerates Amazon OpenSearch Service (domains and their
// policies, packages, cross-cluster connections, VPC endpoints and
// applications) and OpenSearch Serverless (collections, collection
// groups, access/lifecycle/security policies, security configs and VPC
// endpoints). Provider importer IDs follow the terraform-provider-aws
// docs: domains and domain policies by domain name, packages by package
// ID, connections and VPC endpoints by their IDs, applications by
// application ID; serverless collections, collection groups, security
// configs and VPC endpoints by ID, and serverless policies by
// "name/type".
//
// Deliberately not generated:
//   - aws_opensearch_domain_saml_options: per-domain SAML settings
//     subresource; the OpenSearch SDK exposes no get/list API for a
//     domain's SAML options, so existence cannot be enumerated.
//   - aws_opensearch_authorize_vpc_endpoint_access: an authorization
//     grant (domain, principal) — association glue, not a resource with
//     its own list API.
//   - aws_opensearch_package_association: package-to-domain association
//     glue.
func (g *OpenSearchGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := opensearch.NewFromConfig(config)
	domains, err := g.loadDomains(svc)
	if err != nil {
		return err
	}
	_ = domains
	for _, load := range []func(*opensearch.Client) error{
		g.loadPackages, g.loadInboundConnections, g.loadOutboundConnections,
		g.loadVpcEndpoints, g.loadApplications,
	} {
		if err := load(svc); err != nil {
			return err
		}
	}
	return g.loadServerless(opensearchserverless.NewFromConfig(config))
}

func (g *OpenSearchGenerator) loadDomains(svc *opensearch.Client) ([]string, error) {
	out, err := svc.ListDomainNames(context.TODO(), &opensearch.ListDomainNamesInput{})
	if err != nil {
		return nil, err
	}
	var names []string
	for _, d := range out.DomainNames {
		name := StringValue(d.DomainName)
		names = append(names, name)
		g.Resources = append(g.Resources, terraformutils.NewResource(
			name, name, "aws_opensearch_domain", "aws",
			map[string]string{"domain_name": name},
			opensearchAllowEmptyValues, map[string]interface{}{}))
		// The domain policy subresource is keyed by the same domain name;
		// every domain carries an access policy document.
		g.Resources = append(g.Resources, terraformutils.NewResource(
			name, name, "aws_opensearch_domain_policy", "aws",
			map[string]string{"domain_name": name},
			opensearchAllowEmptyValues, map[string]interface{}{}))
	}
	return names, nil
}

func (g *OpenSearchGenerator) loadPackages(svc *opensearch.Client) error {
	input := &opensearch.DescribePackagesInput{}
	for {
		out, err := svc.DescribePackages(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, p := range out.PackageDetailsList {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(p.PackageID), StringValue(p.PackageName),
				"aws_opensearch_package", "aws", opensearchAllowEmptyValues))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *OpenSearchGenerator) loadInboundConnections(svc *opensearch.Client) error {
	input := &opensearch.DescribeInboundConnectionsInput{}
	for {
		out, err := svc.DescribeInboundConnections(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, c := range out.Connections {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(c.ConnectionId), StringValue(c.ConnectionId),
				"aws_opensearch_inbound_connection_accepter", "aws", opensearchAllowEmptyValues))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *OpenSearchGenerator) loadOutboundConnections(svc *opensearch.Client) error {
	input := &opensearch.DescribeOutboundConnectionsInput{}
	for {
		out, err := svc.DescribeOutboundConnections(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, c := range out.Connections {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(c.ConnectionId), StringValue(c.ConnectionAlias),
				"aws_opensearch_outbound_connection", "aws", opensearchAllowEmptyValues))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *OpenSearchGenerator) loadVpcEndpoints(svc *opensearch.Client) error {
	input := &opensearch.ListVpcEndpointsInput{}
	for {
		out, err := svc.ListVpcEndpoints(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, v := range out.VpcEndpointSummaryList {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(v.VpcEndpointId), StringValue(v.VpcEndpointId),
				"aws_opensearch_vpc_endpoint", "aws", opensearchAllowEmptyValues))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *OpenSearchGenerator) loadApplications(svc *opensearch.Client) error {
	input := &opensearch.ListApplicationsInput{}
	for {
		out, err := svc.ListApplications(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, a := range out.ApplicationSummaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(a.Id), StringValue(a.Name),
				"aws_opensearch_application", "aws", opensearchAllowEmptyValues))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *OpenSearchGenerator) loadServerless(svc *opensearchserverless.Client) error {
	collInput := &opensearchserverless.ListCollectionsInput{}
	for {
		out, err := svc.ListCollections(context.TODO(), collInput)
		if err != nil {
			return err
		}
		for _, c := range out.CollectionSummaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(c.Id), StringValue(c.Name),
				"aws_opensearchserverless_collection", "aws", opensearchAllowEmptyValues))
		}
		if out.NextToken == nil {
			break
		}
		collInput.NextToken = out.NextToken
	}
	groupInput := &opensearchserverless.ListCollectionGroupsInput{}
	for {
		out, err := svc.ListCollectionGroups(context.TODO(), groupInput)
		if err != nil {
			return err
		}
		for _, c := range out.CollectionGroupSummaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(c.Id), StringValue(c.Name),
				"aws_opensearchserverless_collection_group", "aws", opensearchAllowEmptyValues))
		}
		if out.NextToken == nil {
			break
		}
		groupInput.NextToken = out.NextToken
	}
	vpcInput := &opensearchserverless.ListVpcEndpointsInput{}
	for {
		out, err := svc.ListVpcEndpoints(context.TODO(), vpcInput)
		if err != nil {
			return err
		}
		for _, v := range out.VpcEndpointSummaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(v.Id), StringValue(v.Name),
				"aws_opensearchserverless_vpc_endpoint", "aws", opensearchAllowEmptyValues))
		}
		if out.NextToken == nil {
			break
		}
		vpcInput.NextToken = out.NextToken
	}
	cfgInput := &opensearchserverless.ListSecurityConfigsInput{}
	for {
		out, err := svc.ListSecurityConfigs(context.TODO(), cfgInput)
		if err != nil {
			return err
		}
		for _, c := range out.SecurityConfigSummaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(c.Id), StringValue(c.Id),
				"aws_opensearchserverless_security_config", "aws", opensearchAllowEmptyValues))
		}
		if out.NextToken == nil {
			break
		}
		cfgInput.NextToken = out.NextToken
	}
	for _, policyType := range []osstypes.AccessPolicyType{osstypes.AccessPolicyTypeData} {
		input := &opensearchserverless.ListAccessPoliciesInput{Type: policyType}
		for {
			out, err := svc.ListAccessPolicies(context.TODO(), input)
			if err != nil {
				return err
			}
			for _, p := range out.AccessPolicySummaries {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					StringValue(p.Name)+"/"+string(p.Type), StringValue(p.Name),
					"aws_opensearchserverless_access_policy", "aws", opensearchAllowEmptyValues))
			}
			if out.NextToken == nil {
				break
			}
			input.NextToken = out.NextToken
		}
	}
	for _, policyType := range []osstypes.SecurityPolicyType{osstypes.SecurityPolicyTypeEncryption, osstypes.SecurityPolicyTypeNetwork} {
		input := &opensearchserverless.ListSecurityPoliciesInput{Type: policyType}
		for {
			out, err := svc.ListSecurityPolicies(context.TODO(), input)
			if err != nil {
				return err
			}
			for _, p := range out.SecurityPolicySummaries {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					StringValue(p.Name)+"/"+string(p.Type), StringValue(p.Name),
					"aws_opensearchserverless_security_policy", "aws", opensearchAllowEmptyValues))
			}
			if out.NextToken == nil {
				break
			}
			input.NextToken = out.NextToken
		}
	}
	input := &opensearchserverless.ListLifecyclePoliciesInput{Type: osstypes.LifecyclePolicyTypeRetention}
	for {
		out, err := svc.ListLifecyclePolicies(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, p := range out.LifecyclePolicySummaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(p.Name)+"/"+string(p.Type), StringValue(p.Name),
				"aws_opensearchserverless_lifecycle_policy", "aws", opensearchAllowEmptyValues))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}
