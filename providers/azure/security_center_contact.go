package azure

import (
	"context"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type SecurityCenterContactGenerator struct {
	AzureService
}

func (g SecurityCenterContactGenerator) listContacts() ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	ctx := context.Background()

	if rg := g.getResourceGroup(); rg != "" {
		return resources, nil
	}

	securityCenterContactClient, err := armsecurity.NewContactsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return resources, err
	}

	pager := securityCenterContactClient.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return resources, err
		}
		for _, contact := range page.Value {
			resources = append(resources, terraformutils.NewSimpleResource(
				*contact.ID,
				*contact.Name,
				"azurerm_security_center_contact",
				g.ProviderName,
				[]string{}))
		}
	}

	return resources, nil
}

func (g *SecurityCenterContactGenerator) InitResources() error {
	resources, err := g.listContacts()
	if err != nil {
		return err
	}

	g.Resources = append(g.Resources, resources...)

	return nil
}
