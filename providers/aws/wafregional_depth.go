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
	"github.com/aws/aws-sdk-go-v2/service/wafregional"
	"github.com/aws/aws-sdk-go-v2/service/wafregional/types"
)

// Depth enumeration for WAF Regional, called from
// WafRegionalGenerator.InitResources: web ACL associations, listed
// per web ACL via ListResourcesForWebACL for both associable resource
// types (Application Load Balancers and API Gateway stages).
// Provider ID format: web_acl_id:resource_arn.
func (g *WafRegionalGenerator) addDepthResources(svc *wafregional.Client) error {
	output, err := svc.ListWebACLs(context.TODO(), &wafregional.ListWebACLsInput{})
	if err != nil {
		return err
	}
	for _, acl := range output.WebACLs {
		webACLID := StringValue(acl.WebACLId)
		for _, resourceType := range []types.ResourceType{types.ResourceTypeApplicationLoadBalancer, types.ResourceTypeApiGateway} {
			out, err := svc.ListResourcesForWebACL(context.TODO(), &wafregional.ListResourcesForWebACLInput{
				WebACLId:     acl.WebACLId,
				ResourceType: resourceType,
			})
			if err != nil {
				return err
			}
			for _, resourceArn := range out.ResourceArns {
				g.Resources = append(g.Resources, terraformutils.NewResource(
					webACLID+":"+resourceArn,
					resourceArn,
					"aws_wafregional_web_acl_association",
					"aws",
					map[string]string{
						"resource_arn": resourceArn,
						"web_acl_id":   webACLID,
					},
					wafAllowEmptyValues,
					map[string]interface{}{},
				))
			}
		}
	}
	return nil
}
