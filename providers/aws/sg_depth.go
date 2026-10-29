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

// getSgDepth extends SecurityGenerator with the default security group
// adoption resource, the standalone ingress/egress rule resources, the
// rules-exclusive overlay and cross-VPC security group associations.
// Failures are logged and skipped so they never block the base
// aws_security_group import.
func (g *SecurityGenerator) getSgDepth(svc *ec2.Client) {
	g.getDefaultSecurityGroups(svc)
	g.getSecurityGroupRules(svc)
	g.getSecurityGroupVpcAssociations(svc)
}

func (g *SecurityGenerator) getDefaultSecurityGroups(svc *ec2.Client) {
	p := ec2.NewDescribeSecurityGroupsPaginator(svc, &ec2.DescribeSecurityGroupsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, group := range page.SecurityGroups {
			if group.GroupId == nil {
				continue
			}
			groupID := *group.GroupId
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				groupID,
				groupID,
				"aws_vpc_security_group_rules_exclusive",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
			if StringValue(group.GroupName) == "default" {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					groupID,
					groupID,
					"aws_default_security_group",
					"aws",
					awsNetCoreAllowEmptyValues,
				))
			}
		}
	}
}

// Standalone rules are imported by their security group rule id
// (sgr-...) and split by direction.
func (g *SecurityGenerator) getSecurityGroupRules(svc *ec2.Client) {
	p := ec2.NewDescribeSecurityGroupRulesPaginator(svc, &ec2.DescribeSecurityGroupRulesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, rule := range page.SecurityGroupRules {
			if rule.SecurityGroupRuleId == nil {
				continue
			}
			ruleID := *rule.SecurityGroupRuleId
			resourceType := "aws_vpc_security_group_ingress_rule"
			if rule.IsEgress != nil && *rule.IsEgress {
				resourceType = "aws_vpc_security_group_egress_rule"
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				ruleID,
				ruleID,
				resourceType,
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
}

// aws_vpc_security_group_vpc_association is imported by
// "<security group id>,<vpc id>"; only established associations count.
func (g *SecurityGenerator) getSecurityGroupVpcAssociations(svc *ec2.Client) {
	p := ec2.NewDescribeSecurityGroupVpcAssociationsPaginator(svc, &ec2.DescribeSecurityGroupVpcAssociationsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, association := range page.SecurityGroupVpcAssociations {
			if association.GroupId == nil || association.VpcId == nil {
				continue
			}
			if string(association.State) != "associated" {
				continue
			}
			id := *association.GroupId + "," + *association.VpcId
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				id,
				"aws_vpc_security_group_vpc_association",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
}
