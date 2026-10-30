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
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/domainservices/armdomainservices"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// DomainServicesGenerator covers Microsoft Entra Domain Services
// (Microsoft.AAD/domainServices).
//
// The azurerm_active_directory_domain_service importer ID is
// "<domain service ARM ID>/initialReplicaSetId/<replica set ID>", so the ID
// is constructed from the first replica set embedded in the domain service.
// Replica sets themselves are embedded in the same properties (they have no
// standalone list call); their importer ID is
// "<domain service ARM ID>/replicaSets/<replica set ID>".
//
// Skipped in this family:
//   - azurerm_active_directory_domain_service_trust: forest trusts are
//     embedded in the domain service's forestTrusts property and
//     armdomainservices has no Trusts client or list call for them.
type DomainServicesGenerator struct {
	AzureService
}

func (g *DomainServicesGenerator) InitResources() error {
	ctx := context.Background()
	client, err := armdomainservices.NewClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	pager := client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			break
		}
		for _, item := range page.Value {
			if item.ID == nil || item.Name == nil || item.Properties == nil {
				continue
			}
			var replicaSets []*armdomainservices.ReplicaSet
			for _, rs := range item.Properties.ReplicaSets {
				if rs != nil && rs.ReplicaSetID != nil {
					replicaSets = append(replicaSets, rs)
				}
			}
			if len(replicaSets) == 0 {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*item.ID+"/initialReplicaSetId/"+*replicaSets[0].ReplicaSetID,
				*item.Name,
				"azurerm_active_directory_domain_service",
				g.ProviderName,
				[]string{}))
			for _, rs := range replicaSets {
				name := *rs.ReplicaSetID
				if rs.Location != nil && *rs.Location != "" {
					name = *item.Name + "-" + *rs.Location
				}
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					*item.ID+"/replicaSets/"+*rs.ReplicaSetID,
					name,
					"azurerm_active_directory_domain_service_replica_set",
					g.ProviderName,
					[]string{}))
			}
		}
	}
	return nil
}
