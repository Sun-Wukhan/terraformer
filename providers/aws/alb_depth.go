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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
)

// getAlbDepth extends AlbGenerator with ELBv2 trust stores and their
// revocations. Failures are logged and skipped so they never block the
// base load balancer import.
func (g *AlbGenerator) getAlbDepth(svc *elasticloadbalancingv2.Client) {
	input := &elasticloadbalancingv2.DescribeTrustStoresInput{}
	for {
		out, err := svc.DescribeTrustStores(context.TODO(), input)
		if err != nil {
			log.Println(err)
			return
		}
		for _, trustStore := range out.TrustStores {
			if trustStore.TrustStoreArn == nil {
				continue
			}
			arn := *trustStore.TrustStoreArn
			name := StringValue(trustStore.Name)
			if name == "" {
				name = arn
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn,
				name,
				"aws_lb_trust_store",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
			g.getTrustStoreRevocations(svc, arn)
		}
		if out.NextMarker == nil {
			break
		}
		input.Marker = out.NextMarker
	}
}

// aws_lb_trust_store_revocation is imported by
// "<trust store arn>,<revocation id>".
func (g *AlbGenerator) getTrustStoreRevocations(svc *elasticloadbalancingv2.Client, trustStoreArn string) {
	out, err := svc.DescribeTrustStoreRevocations(context.TODO(), &elasticloadbalancingv2.DescribeTrustStoreRevocationsInput{
		TrustStoreArn: aws.String(trustStoreArn),
	})
	if err != nil {
		log.Println(err)
		return
	}
	for _, revocation := range out.TrustStoreRevocations {
		if revocation.RevocationId == nil {
			continue
		}
		id := trustStoreArn + "," + strconv.FormatInt(*revocation.RevocationId, 10)
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			id,
			id,
			"aws_lb_trust_store_revocation",
			"aws",
			awsNetCoreAllowEmptyValues,
		))
	}
}
