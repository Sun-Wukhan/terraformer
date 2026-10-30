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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/nginx/armnginx"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// NginxGenerator covers Nginx.NginxPlus deployments and their
// certificates and configurations.
//
// Skipped in this service (see ledger):
//   - azurerm_nginx_api_key: armnginx v1.0.1 has no API keys client /
//     List operation (only Deployments, Certificates, Configurations).
type NginxGenerator struct {
	AzureService
}

func (g *NginxGenerator) listDeployments() ([]*armnginx.Deployment, error) {
	client, err := armnginx.NewDeploymentsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	var out []*armnginx.Deployment
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return out, err
			}
			out = append(out, page.Value...)
		}
	} else {
		pager := client.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return out, err
			}
			out = append(out, page.Value...)
		}
	}
	return out, nil
}

func (g *NginxGenerator) listDeploymentChildren(rg, deployment string) []terraformutils.Resource {
	var resources []terraformutils.Resource
	ctx := context.Background()
	if client, err := armnginx.NewCertificatesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions()); err == nil {
		pager := client.NewListPager(rg, deployment, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					resources = append(resources, terraformutils.NewSimpleResource(*item.ID, *item.Name, "azurerm_nginx_certificate", g.ProviderName, []string{}))
				}
			}
		}
	}
	if client, err := armnginx.NewConfigurationsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions()); err == nil {
		pager := client.NewListPager(rg, deployment, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					resources = append(resources, terraformutils.NewSimpleResource(*item.ID, *item.Name, "azurerm_nginx_configuration", g.ProviderName, []string{}))
				}
			}
		}
	}
	return resources
}

func (g *NginxGenerator) InitResources() error {
	deployments, err := g.listDeployments()
	if err != nil {
		return err
	}
	for _, d := range deployments {
		if d.ID == nil || d.Name == nil {
			continue
		}
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(*d.ID, *d.Name, "azurerm_nginx_deployment", g.ProviderName, []string{}))
		parsed, err := ParseAzureResourceID(*d.ID)
		if err != nil {
			return err
		}
		g.Resources = append(g.Resources, g.listDeploymentChildren(parsed.ResourceGroup, *d.Name)...)
	}
	return nil
}
