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
	"github.com/aws/aws-sdk-go-v2/service/glacier"
)

// Glacier sweep. Both provider types import with the vault name
// (verified against provider docs); vault locks are emitted per vault
// only when GetVaultLock reports a lock policy/state.

var glacierAllowEmptyValues = []string{"tags."}

type GlacierGenerator struct {
	AWSService
}

func (g *GlacierGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := glacier.NewFromConfig(config)
	ctx := context.TODO()
	dash := "-"
	input := &glacier.ListVaultsInput{AccountId: &dash}
	for {
		out, err := svc.ListVaults(ctx, input)
		if err != nil {
			return err
		}
		for _, vault := range out.VaultList {
			name := StringValue(vault.VaultName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(name, name, "aws_glacier_vault", "aws", glacierAllowEmptyValues))
			if lock, err := svc.GetVaultLock(ctx, &glacier.GetVaultLockInput{AccountId: &dash, VaultName: vault.VaultName}); err == nil && lock.Policy != nil {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(name, name, "aws_glacier_vault_lock", "aws", glacierAllowEmptyValues))
			}
		}
		if out.Marker == nil {
			break
		}
		input.Marker = out.Marker
	}
	return nil
}
