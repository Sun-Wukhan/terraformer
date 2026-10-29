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
	"fmt"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53domains"
)

var route53DomainsAllowEmptyValues = []string{"tags."}

type Route53DomainsGenerator struct {
	AWSService
}

// InitResources enumerates the domains registered with Route 53
// Domains. The same registered domain backs both the (legacy)
// aws_route53domains_domain resource and the
// aws_route53domains_registered_domain resource; both are keyed by
// the domain name. A domain's delegation signer (DS) records are
// managed per signing key by the
// aws_route53domains_delegation_signer_record resource, keyed by
// domain_name,flags,algorithm,public_key, and are generated from the
// DNSSEC keys in each domain's detail.
func (g *Route53DomainsGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	client := route53domains.NewFromConfig(config)
	input := &route53domains.ListDomainsInput{}
	for {
		out, err := client.ListDomains(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, domain := range out.Domains {
			name := StringValue(domain.DomainName)
			for _, resourceType := range []string{"aws_route53domains_domain", "aws_route53domains_registered_domain"} {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					name,
					name,
					resourceType,
					"aws",
					route53DomainsAllowEmptyValues,
				))
			}
			if err := g.addDelegationSignerRecord(client, name); err != nil {
				return err
			}
		}
		if out.NextPageMarker == nil {
			return nil
		}
		input.Marker = out.NextPageMarker
	}
}

func (g *Route53DomainsGenerator) addDelegationSignerRecord(client *route53domains.Client, domainName string) error {
	detail, err := client.GetDomainDetail(context.TODO(), &route53domains.GetDomainDetailInput{
		DomainName: aws.String(domainName),
	})
	if err != nil {
		return err
	}
	for _, key := range detail.DnssecKeys {
		var flags, algorithm int32
		if key.Flags != nil {
			flags = *key.Flags
		}
		if key.Algorithm != nil {
			algorithm = *key.Algorithm
		}
		id := fmt.Sprintf("%s,%d,%d,%s", domainName, flags, algorithm, StringValue(key.PublicKey))
		g.Resources = append(g.Resources, terraformutils.NewResource(
			id,
			domainName,
			"aws_route53domains_delegation_signer_record",
			"aws",
			map[string]string{"domain_name": domainName},
			route53DomainsAllowEmptyValues,
			map[string]interface{}{},
		))
	}
	return nil
}
