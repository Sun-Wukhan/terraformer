package azure

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

const (
	blobFormatString = `https://%s.blob.core.windows.net`
	blobIDFormat     = `https://%s.blob.core.windows.net/%s/%s`
)

type StorageBlobGenerator struct {
	AzureService
}

func (g StorageBlobGenerator) getAccountPrimaryKey(ctx context.Context, accountName, accountGroupName string) string {
	client, err := armstorage.NewAccountsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		log.Fatalf("failed to create storage accounts client: %v", err)
	}

	response, err := client.ListKeys(ctx, accountGroupName, accountName, nil)
	if err != nil {
		log.Fatalf("failed to list keys: %v", err)
	}
	return *response.Keys[0].Value
}

func (g StorageBlobGenerator) getContainerClient(ctx context.Context, accountName, accountGroupName, containerName string) (*container.Client, error) {
	accountPrimaryKey := g.getAccountPrimaryKey(ctx, accountName, accountGroupName)
	sharedKeyCredential, err := azblob.NewSharedKeyCredential(accountName, accountPrimaryKey)
	if err != nil {
		return nil, err
	}

	return container.NewClientWithSharedKeyCredential(
		fmt.Sprintf(blobFormatString+"/%s", accountName, containerName),
		sharedKeyCredential,
		nil,
	)
}

func (g StorageBlobGenerator) getBlobsFromContainer(ctx context.Context, accountName, accountGroupName, containerName string) ([]*container.BlobItem, error) {
	containerClient, err := g.getContainerClient(ctx, accountName, accountGroupName, containerName)
	if err != nil {
		return nil, err
	}

	var blobs []*container.BlobItem
	pager := containerClient.NewListBlobsFlatPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		blobs = append(blobs, page.Segment.BlobItems...)
	}

	return blobs, nil
}

func (g StorageBlobGenerator) listStorageBlobs() ([]terraformutils.Resource, error) {
	var storageBlobsResources []terraformutils.Resource
	ctx := context.Background()

	blobContainerGenerator := NewStorageContainerGenerator(
		g.getSubscriptionID(),
		g.getResourceManagerEndpoint(),
		g.getCredential(),
		g.getResourceGroup(),
	)
	blobContainersResources, err := blobContainerGenerator.ListBlobContainers()
	if err != nil {
		return storageBlobsResources, err
	}

	for _, blobContainerResource := range blobContainersResources {
		containerID := blobContainerResource.InstanceState.ID
		parsedContainerID, err := ParseAzureResourceID(containerID)
		if err != nil {
			return storageBlobsResources, err
		}

		storageAccountName := blobContainerResource.InstanceState.Attributes["storage_account_name"]
		containerName := blobContainerResource.InstanceState.Attributes["name"]
		blobsList, err := g.getBlobsFromContainer(ctx, storageAccountName, parsedContainerID.ResourceGroup, containerName)
		if err != nil {
			return storageBlobsResources, err
		}

		for _, blobItem := range blobsList {
			storageBlobsResources = append(storageBlobsResources, terraformutils.NewSimpleResource(
				fmt.Sprintf(blobIDFormat, storageAccountName, containerName, *blobItem.Name),
				*blobItem.Name,
				"azurerm_storage_blob",
				"azurerm",
				[]string{}))
		}
	}

	return storageBlobsResources, err
}

func (g *StorageBlobGenerator) InitResources() error {
	resources, err := g.listStorageBlobs()
	if err != nil {
		return err
	}

	g.Resources = append(g.Resources, resources...)

	return nil
}
