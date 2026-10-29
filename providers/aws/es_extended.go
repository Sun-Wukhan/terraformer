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
	es "github.com/aws/aws-sdk-go-v2/service/elasticsearchservice"
)

// Elasticsearch remainder for the AWS sweep: the per-domain access
// policy (keyed by domain name, the provider importer ID, mirroring
// aws_opensearch_domain_policy) and VPC endpoints (by VPC endpoint ID).
//
// Deliberately not generated: aws_elasticsearch_domain_saml_options —
// per-domain SAML settings subresource; the Elasticsearch SDK exposes
// no get/list API for a domain's SAML options, so existence cannot be
// enumerated.
func (g *EsGenerator) loadExtended(svc *es.Client) error {
	out, err := svc.ListDomainNames(context.TODO(), &es.ListDomainNamesInput{})
	if err != nil {
		return err
	}
	for _, d := range out.DomainNames {
		name := StringValue(d.DomainName)
		g.Resources = append(g.Resources, terraformutils.NewResource(
			name, name, "aws_elasticsearch_domain_policy", "aws",
			map[string]string{"domain_name": name},
			esAllowEmptyValues, map[string]interface{}{}))
	}
	input := &es.ListVpcEndpointsInput{}
	for {
		page, err := svc.ListVpcEndpoints(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, v := range page.VpcEndpointSummaryList {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(v.VpcEndpointId), StringValue(v.VpcEndpointId),
				"aws_elasticsearch_vpc_endpoint", "aws", esAllowEmptyValues))
		}
		if page.NextToken == nil {
			return nil
		}
		input.NextToken = page.NextToken
	}
}
