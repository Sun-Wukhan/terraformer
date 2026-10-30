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

package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/attestation/armattestation"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// AttestationGenerator imports attestation providers via the track-2
// armattestation SDK (ProvidersClient exposes non-paged List operations).
type AttestationGenerator struct {
	AzureService
}

func (g *AttestationGenerator) InitResources() error {
	ctx := context.Background()
	client, err := armattestation.NewProvidersClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	var items []*armattestation.Provider
	if rg := g.getResourceGroup(); rg != "" {
		resp, err := client.ListByResourceGroup(ctx, rg, nil)
		if err != nil {
			return err
		}
		items = resp.Value
	} else {
		resp, err := client.List(ctx, nil)
		if err != nil {
			return err
		}
		items = resp.Value
	}
	for _, item := range items {
		if item.ID == nil || item.Name == nil {
			continue
		}
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			*item.ID, *item.Name, "azurerm_attestation_provider", g.ProviderName, []string{}))
	}
	return nil
}
