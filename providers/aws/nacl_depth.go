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
	"strconv"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// getNaclDepth extends NaclGenerator with network ACL associations and
// standalone rules, both derived from the network ACL object. Failures are
// logged and skipped so they never block the base aws_network_acl import.
func (g *NaclGenerator) getNaclDepth(svc *ec2.Client) {
	p := ec2.NewDescribeNetworkAclsPaginator(svc, &ec2.DescribeNetworkAclsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, acl := range page.NetworkAcls {
			if acl.NetworkAclId == nil {
				continue
			}
			aclID := *acl.NetworkAclId
			for _, association := range acl.Associations {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					StringValue(association.NetworkAclAssociationId),
					StringValue(association.NetworkAclAssociationId),
					"aws_network_acl_association",
					"aws",
					awsNetCoreAllowEmptyValues,
				))
			}
			// aws_network_acl_rule is imported by
			// "<acl id>:<rule number>:<protocol>:<egress>".
			for _, entry := range acl.Entries {
				if entry.RuleNumber == nil || entry.Protocol == nil || entry.Egress == nil {
					continue
				}
				id := aclID + ":" + strconv.Itoa(int(*entry.RuleNumber)) + ":" +
					*entry.Protocol + ":" + strconv.FormatBool(*entry.Egress)
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id,
					id,
					"aws_network_acl_rule",
					"aws",
					awsNetCoreAllowEmptyValues,
				))
			}
		}
	}
}
