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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storagemover/armstoragemover"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// StorageMoverGenerator covers Azure Storage Mover via armstoragemover:
// movers, agents, projects and job definitions.
//
// Skipped in this family (see docs/azure.md "Known limits"):
//   - azurerm_storage_mover_source_endpoint and
//     azurerm_storage_mover_target_endpoint are the same ARM endpoints
//     collection; armstoragemover endpoint properties carry no source/target
//     discriminator (the two provider types differ only by how job
//     definitions reference the endpoint).
type StorageMoverGenerator struct {
	AzureService
}

func (g *StorageMoverGenerator) InitResources() error {
	ctx := context.Background()
	sub := g.getSubscriptionID()
	cred := g.getCredential()
	opts := g.armClientOptions()
	var resources []terraformutils.Resource
	add := func(id, name, typ string) {
		if id != "" && name != "" {
			resources = append(resources, terraformutils.NewSimpleResource(id, name, typ, "azurerm", []string{}))
		}
	}
	moversClient, err := armstoragemover.NewStorageMoversClient(sub, cred, opts)
	if err != nil {
		return err
	}
	agentsClient, err := armstoragemover.NewAgentsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	projectsClient, err := armstoragemover.NewProjectsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	jobsClient, err := armstoragemover.NewJobDefinitionsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	type moverRef struct{ rg, name string }
	var movers []moverRef
	if rg := g.getResourceGroup(); rg != "" {
		pager := moversClient.NewListPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_storage_mover")
					movers = append(movers, moverRef{rg, *item.Name})
				}
			}
		}
	} else {
		pager := moversClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, item := range page.Value {
				if item.ID == nil || item.Name == nil {
					continue
				}
				add(*item.ID, *item.Name, "azurerm_storage_mover")
				if parsed, err := ParseAzureResourceID(*item.ID); err == nil {
					movers = append(movers, moverRef{parsed.ResourceGroup, *item.Name})
				}
			}
		}
	}
	for _, mover := range movers {
		aPager := agentsClient.NewListPager(mover.rg, mover.name, nil)
		for aPager.More() {
			page, err := aPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					add(*item.ID, *item.Name, "azurerm_storage_mover_agent")
				}
			}
		}
		pPager := projectsClient.NewListPager(mover.rg, mover.name, nil)
		for pPager.More() {
			page, err := pPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, proj := range page.Value {
				if proj.ID == nil || proj.Name == nil {
					continue
				}
				add(*proj.ID, *proj.Name, "azurerm_storage_mover_project")
				jPager := jobsClient.NewListPager(mover.rg, mover.name, *proj.Name, nil)
				for jPager.More() {
					jPage, err := jPager.NextPage(ctx)
					if err != nil {
						log.Println(err)
						break
					}
					for _, job := range jPage.Value {
						if job.ID != nil && job.Name != nil {
							add(*job.ID, *job.Name, "azurerm_storage_mover_job_definition")
						}
					}
				}
			}
		}
	}
	g.Resources = resources
	return nil
}
