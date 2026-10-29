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
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// getEipDepth extends ElasticIPGenerator with EIP associations and custom
// reverse-DNS domain names. Failures are logged and skipped so they never
// block the base aws_eip import.
func (g *ElasticIPGenerator) getEipDepth(svc *ec2.Client) {
	out, err := svc.DescribeAddresses(context.TODO(), &ec2.DescribeAddressesInput{})
	if err != nil {
		log.Println(err)
		return
	}
	for _, address := range out.Addresses {
		if address.AssociationId == nil {
			continue
		}
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			StringValue(address.AssociationId),
			StringValue(address.AssociationId),
			"aws_eip_association",
			"aws",
			awsNetCoreAllowEmptyValues,
		))
	}
	g.getEipDomainNames(svc)
}

// aws_eip_domain_name is imported by allocation id. The domain-name
// attribute query returns an entry for every address that has a custom
// reverse DNS domain configured; the SDK type does not model the domain
// value itself, but the allocation id is all the import needs.
func (g *ElasticIPGenerator) getEipDomainNames(svc *ec2.Client) {
	p := ec2.NewDescribeAddressesAttributePaginator(svc, &ec2.DescribeAddressesAttributeInput{
		Attribute: types.AddressAttributeNameDomainName,
	})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, address := range page.Addresses {
			if address.AllocationId == nil {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(address.AllocationId),
				StringValue(address.AllocationId),
				"aws_eip_domain_name",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
}
