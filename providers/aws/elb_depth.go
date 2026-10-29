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
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
)

// getElbDepth extends ElbGenerator with app cookie stickiness policies.
// Failures are logged and skipped so they never block the base aws_elb
// import.
//
// Notable skips in the classic ELB family (see the cluster ledger):
// aws_elb_attachment, aws_lb_cookie_stickiness_policy,
// aws_lb_ssl_negotiation_policy, aws_load_balancer_policy,
// aws_load_balancer_listener_policy, aws_load_balancer_backend_server_policy
// and aws_proxy_protocol_policy all have no provider importer in their
// current docs; only the app cookie stickiness policy is importable.
func (g *ElbGenerator) getElbDepth(svc *elasticloadbalancing.Client) {
	p := elasticloadbalancing.NewDescribeLoadBalancersPaginator(svc, &elasticloadbalancing.DescribeLoadBalancersInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, loadBalancer := range page.LoadBalancerDescriptions {
			if loadBalancer.LoadBalancerName == nil {
				continue
			}
			lbName := *loadBalancer.LoadBalancerName
			appCookiePolicies := map[string]bool{}
			if loadBalancer.Policies != nil {
				for _, policy := range loadBalancer.Policies.AppCookieStickinessPolicies {
					appCookiePolicies[StringValue(policy.PolicyName)] = true
				}
			}
			// aws_app_cookie_stickiness_policy is imported by
			// "<elb name>:<port>:<policy name>".
			for _, listenerDescription := range loadBalancer.ListenerDescriptions {
				if listenerDescription.Listener == nil {
					continue
				}
				port := strconv.Itoa(int(listenerDescription.Listener.LoadBalancerPort))
				for _, policyName := range listenerDescription.PolicyNames {
					if !appCookiePolicies[policyName] {
						continue
					}
					id := lbName + ":" + port + ":" + policyName
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						id,
						id,
						"aws_app_cookie_stickiness_policy",
						"aws",
						awsNetCoreAllowEmptyValues,
					))
				}
			}
		}
	}
}
