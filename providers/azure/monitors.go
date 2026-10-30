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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/datadog/armdatadog"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dynatrace/armdynatrace"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/newrelic/armnewrelicobservability"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// DatadogGenerator covers Microsoft.Datadog monitors and their tag rules
// and single sign-on configurations (both have List operations in
// armdatadog v1.3.0, despite being singleton-style children).
type DatadogGenerator struct {
	AzureService
}

func (g *DatadogGenerator) listMonitors() ([]*armdatadog.MonitorResource, error) {
	client, err := armdatadog.NewMonitorsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	var out []*armdatadog.MonitorResource
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

func (g *DatadogGenerator) listMonitorChildren(rg, monitor string) []terraformutils.Resource {
	var resources []terraformutils.Resource
	ctx := context.Background()
	if client, err := armdatadog.NewTagRulesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions()); err == nil {
		pager := client.NewListPager(rg, monitor, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					resources = append(resources, terraformutils.NewSimpleResource(*item.ID, *item.Name, "azurerm_datadog_monitor_tag_rule", g.ProviderName, []string{}))
				}
			}
		}
	}
	if client, err := armdatadog.NewSingleSignOnConfigurationsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions()); err == nil {
		pager := client.NewListPager(rg, monitor, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					resources = append(resources, terraformutils.NewSimpleResource(*item.ID, *item.Name, "azurerm_datadog_monitor_sso_configuration", g.ProviderName, []string{}))
				}
			}
		}
	}
	return resources
}

func (g *DatadogGenerator) InitResources() error {
	monitors, err := g.listMonitors()
	if err != nil {
		return err
	}
	for _, m := range monitors {
		if m.ID == nil || m.Name == nil {
			continue
		}
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(*m.ID, *m.Name, "azurerm_datadog_monitor", g.ProviderName, []string{}))
		parsed, err := ParseAzureResourceID(*m.ID)
		if err != nil {
			return err
		}
		g.Resources = append(g.Resources, g.listMonitorChildren(parsed.ResourceGroup, *m.Name)...)
	}
	return nil
}

// DynatraceGenerator covers Dynatrace.Observability monitors and their
// tag rules (TagRulesClient List in armdynatrace v1.1.0).
type DynatraceGenerator struct {
	AzureService
}

func (g *DynatraceGenerator) listMonitors() ([]*armdynatrace.MonitorResource, error) {
	client, err := armdynatrace.NewMonitorsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	var out []*armdynatrace.MonitorResource
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
		pager := client.NewListBySubscriptionIDPager(nil)
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

func (g *DynatraceGenerator) InitResources() error {
	monitors, err := g.listMonitors()
	if err != nil {
		return err
	}
	tagClient, err := armdynatrace.NewTagRulesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, m := range monitors {
		if m.ID == nil || m.Name == nil {
			continue
		}
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(*m.ID, *m.Name, "azurerm_dynatrace_monitor", g.ProviderName, []string{}))
		parsed, err := ParseAzureResourceID(*m.ID)
		if err != nil {
			return err
		}
		pager := tagClient.NewListPager(parsed.ResourceGroup, *m.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(*item.ID, *item.Name, "azurerm_dynatrace_tag_rules", g.ProviderName, []string{}))
				}
			}
		}
	}
	return nil
}

// NewRelicGenerator covers NewRelic.Observability monitors and their tag
// rules (TagRulesClient ListByNewRelicMonitorResource in
// armnewrelicobservability v1.2.0).
type NewRelicGenerator struct {
	AzureService
}

func (g *NewRelicGenerator) InitResources() error {
	client, err := armnewrelicobservability.NewMonitorsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	tagClient, err := armnewrelicobservability.NewTagRulesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	ctx := context.Background()
	var monitors []*armnewrelicobservability.NewRelicMonitorResource
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			monitors = append(monitors, page.Value...)
		}
	} else {
		pager := client.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			monitors = append(monitors, page.Value...)
		}
	}
	for _, m := range monitors {
		if m.ID == nil || m.Name == nil {
			continue
		}
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(*m.ID, *m.Name, "azurerm_new_relic_monitor", g.ProviderName, []string{}))
		parsed, err := ParseAzureResourceID(*m.ID)
		if err != nil {
			return err
		}
		pager := tagClient.NewListByNewRelicMonitorResourcePager(parsed.ResourceGroup, *m.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				break
			}
			for _, item := range page.Value {
				if item.ID != nil && item.Name != nil {
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(*item.ID, *item.Name, "azurerm_new_relic_tag_rule", g.ProviderName, []string{}))
				}
			}
		}
	}
	return nil
}
