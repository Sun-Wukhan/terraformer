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
	"strings"

	armappplatform "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appplatform/armappplatform/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/servicelinker/armservicelinker"
)

// SpringCloudGenerator covers Microsoft.AppPlatform/Spring services and
// their deep child tree (apps, deployments, gateways, APMs, ...). Several
// Terraform types share one ARM kind and are split by a discriminator:
// deployments by uploaded-source type, APMs by APM type, and app
// associations (bindings) by the bound resource's ARM type.
type SpringCloudGenerator struct {
	AzureService
}

// springDeploymentTerraformType maps a deployment's source type to its
// Terraform type.
func springDeploymentTerraformType(deployment *armappplatform.DeploymentResource) string {
	if deployment == nil || deployment.Properties == nil || deployment.Properties.Source == nil ||
		deployment.Properties.Source.GetUserSourceInfo() == nil ||
		deployment.Properties.Source.GetUserSourceInfo().Type == nil {
		return "azurerm_spring_cloud_java_deployment"
	}
	switch *deployment.Properties.Source.GetUserSourceInfo().Type {
	case "BuildResult":
		return "azurerm_spring_cloud_build_deployment"
	case "CustomContainer":
		return "azurerm_spring_cloud_container_deployment"
	default:
		// Jar, War, Source, NetCoreZip and plain uploads are all managed
		// through the Java deployment resource.
		return "azurerm_spring_cloud_java_deployment"
	}
}

// springAPMTerraformType maps an APM's type to its Terraform type.
func springAPMTerraformType(apm *armappplatform.ApmResource) string {
	if apm == nil || apm.Properties == nil || apm.Properties.Type == nil {
		return ""
	}
	switch strings.ToLower(*apm.Properties.Type) {
	case "applicationinsights":
		return "azurerm_spring_cloud_application_insights_application_performance_monitoring"
	case "appdynamics":
		return "azurerm_spring_cloud_app_dynamics_application_performance_monitoring"
	case "dynatrace":
		return "azurerm_spring_cloud_dynatrace_application_performance_monitoring"
	case "elasticapm":
		return "azurerm_spring_cloud_elastic_application_performance_monitoring"
	case "newrelic":
		return "azurerm_spring_cloud_new_relic_application_performance_monitoring"
	}
	return ""
}

// springBindingTerraformType maps a binding's bound-resource ARM type to
// its Terraform association type.
func springBindingTerraformType(binding *armappplatform.BindingResource) string {
	if binding == nil || binding.Properties == nil || binding.Properties.ResourceType == nil {
		return ""
	}
	resourceType := *binding.Properties.ResourceType
	switch {
	case strings.Contains(resourceType, "Microsoft.DocumentDB"):
		return "azurerm_spring_cloud_app_cosmosdb_association"
	case strings.Contains(resourceType, "Microsoft.DBforMySQL") || strings.Contains(resourceType, "Microsoft.DBforMariaDB"):
		return "azurerm_spring_cloud_app_mysql_association"
	case strings.Contains(resourceType, "Microsoft.Cache"):
		return "azurerm_spring_cloud_app_redis_association"
	}
	return ""
}

func (g *SpringCloudGenerator) InitResources() error {
	ctx := context.Background()
	servicesClient, err := armappplatform.NewServicesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	var services []*armappplatform.ServiceResource
	if rg := g.getResourceGroup(); rg != "" {
		pager := servicesClient.NewListPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			services = append(services, page.Value...)
		}
	} else {
		pager := servicesClient.NewListBySubscriptionPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			services = append(services, page.Value...)
		}
	}
	for _, service := range services {
		if err := g.appendServiceChildren(ctx, service); err != nil {
			return err
		}
	}
	return nil
}

func (g *SpringCloudGenerator) appendServiceChildren(ctx context.Context, service *armappplatform.ServiceResource) error {
	g.AppendSimpleResource(*service.ID, *service.Name, "azurerm_spring_cloud_service")
	parsed, err := ParseAzureResourceID(*service.ID)
	if err != nil {
		return err
	}
	rg, serviceName := parsed.ResourceGroup, *service.Name
	sub, cred, opts := g.getSubscriptionID(), g.getCredential(), g.armClientOptions()

	certificatesClient, err := armappplatform.NewCertificatesClient(sub, cred, opts)
	if err != nil {
		return err
	}
	pager := certificatesClient.NewListPager(rg, serviceName, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_spring_cloud_certificate")
		}
	}

	configServicesClient, err := armappplatform.NewConfigurationServicesClient(sub, cred, opts)
	if err != nil {
		return err
	}
	csPager := configServicesClient.NewListPager(rg, serviceName, nil)
	for csPager.More() {
		page, err := csPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_spring_cloud_configuration_service")
		}
	}

	storagesClient, err := armappplatform.NewStoragesClient(sub, cred, opts)
	if err != nil {
		return err
	}
	storagePager := storagesClient.NewListPager(rg, serviceName, nil)
	for storagePager.More() {
		page, err := storagePager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_spring_cloud_storage")
		}
	}

	apmsClient, err := armappplatform.NewApmsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	apmPager := apmsClient.NewListPager(rg, serviceName, nil)
	for apmPager.More() {
		page, err := apmPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			if resourceType := springAPMTerraformType(item); resourceType != "" {
				g.AppendSimpleResource(*item.ID, *item.Name, resourceType)
			}
		}
	}

	liveViewsClient, err := armappplatform.NewApplicationLiveViewsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	lvPager := liveViewsClient.NewListPager(rg, serviceName, nil)
	for lvPager.More() {
		page, err := lvPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_spring_cloud_application_live_view")
		}
	}

	devToolPortalsClient, err := armappplatform.NewDevToolPortalsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	dtpPager := devToolPortalsClient.NewListPager(rg, serviceName, nil)
	for dtpPager.More() {
		page, err := dtpPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_spring_cloud_dev_tool_portal")
		}
	}

	acceleratorsClient, err := armappplatform.NewApplicationAcceleratorsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	accPager := acceleratorsClient.NewListPager(rg, serviceName, nil)
	for accPager.More() {
		page, err := accPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, accelerator := range page.Value {
			g.AppendSimpleResource(*accelerator.ID, *accelerator.Name, "azurerm_spring_cloud_accelerator")

			customizedClient, err := armappplatform.NewCustomizedAcceleratorsClient(sub, cred, opts)
			if err != nil {
				return err
			}
			caPager := customizedClient.NewListPager(rg, serviceName, *accelerator.Name, nil)
			for caPager.More() {
				caPage, err := caPager.NextPage(ctx)
				if err != nil {
					log.Println(err)
					return err
				}
				for _, item := range caPage.Value {
					g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_spring_cloud_customized_accelerator")
				}
			}
		}
	}

	gatewaysClient, err := armappplatform.NewGatewaysClient(sub, cred, opts)
	if err != nil {
		return err
	}
	gatewayPager := gatewaysClient.NewListPager(rg, serviceName, nil)
	for gatewayPager.More() {
		page, err := gatewayPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, gateway := range page.Value {
			g.AppendSimpleResource(*gateway.ID, *gateway.Name, "azurerm_spring_cloud_gateway")
			if err := g.appendGatewayChildren(ctx, rg, serviceName, *gateway.Name); err != nil {
				return err
			}
		}
	}

	apiPortalsClient, err := armappplatform.NewAPIPortalsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	portalPager := apiPortalsClient.NewListPager(rg, serviceName, nil)
	for portalPager.More() {
		page, err := portalPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, portal := range page.Value {
			g.AppendSimpleResource(*portal.ID, *portal.Name, "azurerm_spring_cloud_api_portal")

			domainsClient, err := armappplatform.NewAPIPortalCustomDomainsClient(sub, cred, opts)
			if err != nil {
				return err
			}
			domainPager := domainsClient.NewListPager(rg, serviceName, *portal.Name, nil)
			for domainPager.More() {
				domainPage, err := domainPager.NextPage(ctx)
				if err != nil {
					log.Println(err)
					return err
				}
				for _, item := range domainPage.Value {
					g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_spring_cloud_api_portal_custom_domain")
				}
			}
		}
	}

	// Builders live under the (singleton, always named "default") build
	// service; build pack bindings live under each builder.
	buildersClient, err := armappplatform.NewBuildServiceBuilderClient(sub, cred, opts)
	if err != nil {
		return err
	}
	builderPager := buildersClient.NewListPager(rg, serviceName, "default", nil)
	for builderPager.More() {
		page, err := builderPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, builder := range page.Value {
			g.AppendSimpleResource(*builder.ID, *builder.Name, "azurerm_spring_cloud_builder")

			bindingsClient, err := armappplatform.NewBuildpackBindingClient(sub, cred, opts)
			if err != nil {
				return err
			}
			bpPager := bindingsClient.NewListPager(rg, serviceName, "default", *builder.Name, nil)
			for bpPager.More() {
				bpPage, err := bpPager.NextPage(ctx)
				if err != nil {
					log.Println(err)
					return err
				}
				for _, item := range bpPage.Value {
					g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_spring_cloud_build_pack_binding")
				}
			}
		}
	}

	appsClient, err := armappplatform.NewAppsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	appPager := appsClient.NewListPager(rg, serviceName, nil)
	for appPager.More() {
		page, err := appPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, app := range page.Value {
			if err := g.appendAppChildren(ctx, rg, serviceName, app); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *SpringCloudGenerator) appendGatewayChildren(ctx context.Context, rg, serviceName, gatewayName string) error {
	sub, cred, opts := g.getSubscriptionID(), g.getCredential(), g.armClientOptions()

	domainsClient, err := armappplatform.NewGatewayCustomDomainsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	domainPager := domainsClient.NewListPager(rg, serviceName, gatewayName, nil)
	for domainPager.More() {
		page, err := domainPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_spring_cloud_gateway_custom_domain")
		}
	}

	routeConfigsClient, err := armappplatform.NewGatewayRouteConfigsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	rcPager := routeConfigsClient.NewListPager(rg, serviceName, gatewayName, nil)
	for rcPager.More() {
		page, err := rcPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_spring_cloud_gateway_route_config")
		}
	}
	return nil
}

func (g *SpringCloudGenerator) appendAppChildren(ctx context.Context, rg, serviceName string, app *armappplatform.AppResource) error {
	g.AppendSimpleResource(*app.ID, *app.Name, "azurerm_spring_cloud_app")
	appName := *app.Name
	sub, cred, opts := g.getSubscriptionID(), g.getCredential(), g.armClientOptions()

	domainsClient, err := armappplatform.NewCustomDomainsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	domainPager := domainsClient.NewListPager(rg, serviceName, appName, nil)
	for domainPager.More() {
		page, err := domainPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_spring_cloud_custom_domain")
		}
	}

	bindingsClient, err := armappplatform.NewBindingsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	bindingPager := bindingsClient.NewListPager(rg, serviceName, appName, nil)
	for bindingPager.More() {
		page, err := bindingPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			if resourceType := springBindingTerraformType(item); resourceType != "" {
				g.AppendSimpleResource(*item.ID, *item.Name, resourceType)
			}
		}
	}

	deploymentsClient, err := armappplatform.NewDeploymentsClient(sub, cred, opts)
	if err != nil {
		return err
	}
	linkerClient, err := armservicelinker.NewLinkerClient(cred, opts)
	if err != nil {
		return err
	}
	deploymentPager := deploymentsClient.NewListPager(rg, serviceName, appName, nil)
	for deploymentPager.More() {
		page, err := deploymentPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, deployment := range page.Value {
			g.AppendSimpleResource(*deployment.ID, *deployment.Name, springDeploymentTerraformType(deployment))

			// The active deployment is a flag on one deployment of the app;
			// its Terraform ID is the app ID.
			if deployment.Properties != nil && deployment.Properties.Active != nil && *deployment.Properties.Active {
				g.AppendSimpleResource(*app.ID, appName, "azurerm_spring_cloud_active_deployment")
			}

			// Service-connector connections are ServiceLinker linkers
			// scoped to the deployment resource.
			linkerPager := linkerClient.NewListPager(*deployment.ID, nil)
			for linkerPager.More() {
				linkerPage, err := linkerPager.NextPage(ctx)
				if err != nil {
					log.Println(err)
					return err
				}
				for _, item := range linkerPage.Value {
					g.AppendSimpleResource(*item.ID, *item.Name, "azurerm_spring_cloud_connection")
				}
			}
		}
	}
	return nil
}
