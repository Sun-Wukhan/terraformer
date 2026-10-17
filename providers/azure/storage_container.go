package azure

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/hashicorp/go-azure-helpers/authentication"
)

const (
	containerIDFormat = "https://%s.blob.core.windows.net/%s"
)

type StorageContainerGenerator struct {
	AzureService
}

func NewStorageContainerGenerator(subscriptionID, resourceManagerEndpoint string, credential azcore.TokenCredential, rg string) *StorageContainerGenerator {
	storageContainerGenerator := new(StorageContainerGenerator)
	storageContainerGenerator.Args = map[string]interface{}{}
	storageContainerGenerator.Args["config"] = authentication.Config{CustomResourceManagerEndpoint: resourceManagerEndpoint, SubscriptionID: subscriptionID}
	storageContainerGenerator.Args["credential"] = credential
	storageContainerGenerator.Args["resource_group"] = rg

	return storageContainerGenerator
}

func (g StorageContainerGenerator) ListBlobContainers() ([]terraformutils.Resource, error) {
	var containerResources []terraformutils.Resource
	ctx := context.Background()

	blobContainersClient, err := armstorage.NewBlobContainersClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return containerResources, err
	}

	accounts, err := g.getStorageAccounts()
	if err != nil {
		return containerResources, err
	}

	for _, storageAccount := range accounts {
		parsedStorageAccountResourceID, err := ParseAzureResourceID(*storageAccount.ID)
		if err != nil {
			break
		}
		pager := blobContainersClient.NewListPager(parsedStorageAccountResourceID.ResourceGroup, *storageAccount.Name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return containerResources, err
			}
			for _, containerItem := range page.Value {
				containerResources = append(containerResources,
					terraformutils.NewResource(
						fmt.Sprintf(containerIDFormat, *storageAccount.Name, *containerItem.Name),
						*containerItem.Name,
						"azurerm_storage_container",
						"azurerm",
						map[string]string{
							"storage_account_name": *storageAccount.Name,
							"name":                 *containerItem.Name,
						},
						[]string{},
						map[string]interface{}{}))
			}
		}
	}

	return containerResources, nil
}

func (g *StorageContainerGenerator) getStorageAccounts() ([]*armstorage.Account, error) {
	ctx := context.Background()
	client, err := armstorage.NewAccountsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	var accounts []*armstorage.Account
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return accounts, err
			}
			accounts = append(accounts, page.Value...)
		}
	} else {
		pager := client.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return accounts, err
			}
			accounts = append(accounts, page.Value...)
		}
	}

	return accounts, nil
}

func (g *StorageContainerGenerator) InitResources() error {
	storageAccounts, err := g.ListBlobContainers()
	if err != nil {
		return err
	}

	g.Resources = storageAccounts

	return nil
}
