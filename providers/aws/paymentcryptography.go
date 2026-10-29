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
	"github.com/aws/aws-sdk-go-v2/service/paymentcryptography"
)

var paymentcryptographyAllowEmptyValues = []string{"tags."}

// PaymentCryptographyGenerator covers Payment Cryptography keys and key aliases.
// Provider import IDs (verified): key = key ARN; alias = alias name (e.g. "alias/<id>").
type PaymentCryptographyGenerator struct {
	AWSService
}

func (g *PaymentCryptographyGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := paymentcryptography.NewFromConfig(config)
	kp := paymentcryptography.NewListKeysPaginator(svc, &paymentcryptography.ListKeysInput{})
	for kp.HasMorePages() {
		page, err := kp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, key := range page.Keys {
			arn := StringValue(key.KeyArn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_paymentcryptography_key", "aws", paymentcryptographyAllowEmptyValues))
		}
	}
	ap := paymentcryptography.NewListAliasesPaginator(svc, &paymentcryptography.ListAliasesInput{})
	for ap.HasMorePages() {
		page, err := ap.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, alias := range page.Aliases {
			name := StringValue(alias.AliasName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_paymentcryptography_key_alias", "aws", paymentcryptographyAllowEmptyValues))
		}
	}
	return nil
}
