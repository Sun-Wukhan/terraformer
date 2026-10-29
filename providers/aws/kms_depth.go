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
	"log"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
)

// KMS depth pass (AWS full-coverage sweep, security cluster).
//
// Implemented here (provider importer ID = the key/store ID in every
// case, per the terraform-provider-aws docs):
//   - aws_kms_custom_key_store: from DescribeCustomKeyStores.
//   - aws_kms_external_key: customer keys whose DescribeKey origin is
//     EXTERNAL and which are not multi-region replicas.
//   - aws_kms_replica_key: multi-region replica keys with the default
//     AWS_KMS origin.
//   - aws_kms_replica_external_key: multi-region replica keys with
//     EXTERNAL origin.
//   - aws_kms_key_policy: for every customer-managed key, when
//     GetKeyPolicy returns the "default" policy.
//
// Intentionally not generated: aws_kms_ciphertext is the result of a
// one-off Encrypt operation, not a stored AWS object; it has no
// provider importer and nothing to enumerate.
func (g *KmsGenerator) addDepthResources(client *kms.Client) {
	g.addCustomKeyStores(client)
	g.addKeyVariantsAndPolicies(client)
}

func (g *KmsGenerator) addCustomKeyStores(client *kms.Client) {
	p := kms.NewDescribeCustomKeyStoresPaginator(client, &kms.DescribeCustomKeyStoresInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, store := range page.CustomKeyStores {
			id := StringValue(store.CustomKeyStoreId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				StringValue(store.CustomKeyStoreName),
				"aws_kms_custom_key_store",
				"aws",
				kmsAllowEmptyValues))
		}
	}
}

func (g *KmsGenerator) addKeyVariantsAndPolicies(client *kms.Client) {
	p := kms.NewListKeysPaginator(client, &kms.ListKeysInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, key := range page.Keys {
			described, err := client.DescribeKey(context.TODO(), &kms.DescribeKeyInput{KeyId: key.KeyId})
			if err != nil {
				log.Println(err)
				continue
			}
			meta := described.KeyMetadata
			if meta.KeyManager != types.KeyManagerTypeCustomer {
				continue
			}
			keyID := StringValue(key.KeyId)
			isReplica := meta.MultiRegionConfiguration != nil &&
				meta.MultiRegionConfiguration.MultiRegionKeyType == types.MultiRegionKeyTypeReplica
			switch {
			case isReplica && meta.Origin == types.OriginTypeExternal:
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					keyID, keyID, "aws_kms_replica_external_key", "aws", kmsAllowEmptyValues))
			case isReplica:
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					keyID, keyID, "aws_kms_replica_key", "aws", kmsAllowEmptyValues))
			case meta.Origin == types.OriginTypeExternal:
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					keyID, keyID, "aws_kms_external_key", "aws", kmsAllowEmptyValues))
			}
			g.addKeyPolicy(client, keyID)
		}
	}
}

func (g *KmsGenerator) addKeyPolicy(client *kms.Client, keyID string) {
	out, err := client.GetKeyPolicy(context.TODO(), &kms.GetKeyPolicyInput{
		KeyId:      &keyID,
		PolicyName: aws.String("default"),
	})
	if err != nil || out.Policy == nil || *out.Policy == "" {
		return
	}
	resource := terraformutils.NewSimpleResource(
		keyID,
		keyID,
		"aws_kms_key_policy",
		"aws",
		kmsAllowEmptyValues)
	resource.SlowQueryRequired = true
	g.Resources = append(g.Resources, resource)
}
