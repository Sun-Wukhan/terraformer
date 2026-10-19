package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type SecurityCenterSubscriptionPricingGenerator struct {
	AzureService
}

func (g SecurityCenterSubscriptionPricingGenerator) listSubscriptionPricing() ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()

	if rg := g.getResourceGroup(); rg != "" {
		return resources, nil
	}

	securityCenterPricingClient, err := armsecurity.NewPricingsClient(g.getCredential(), g.armClientOptions())
	if err != nil {
		return resources, err
	}

	pricingList, err := securityCenterPricingClient.List(ctx, "subscriptions/"+g.getSubscriptionID(), nil)
	if err != nil {
		return resources, err
	}

	for _, pricing := range pricingList.Value {
		resources = append(resources, terraformutils.NewSimpleResource(
			*pricing.ID,
			*pricing.Name,
			"azurerm_security_center_subscription_pricing",
			g.ProviderName,
			[]string{}))
	}

	return resources, nil
}

func (g *SecurityCenterSubscriptionPricingGenerator) InitResources() error {
	resources, err := g.listSubscriptionPricing()
	if err != nil {
		return err
	}

	g.Resources = append(g.Resources, resources...)

	return nil
}
