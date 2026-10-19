// Copyright 2019 The Terraformer Authors.
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
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/go-autorest/autorest"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/hashicorp/go-azure-helpers/authentication"
)

type AzureService struct { //nolint
	terraformutils.Service
}

func (az *AzureService) getClientArgs() (subscriptionID string, resourceGroup string, authorizer autorest.Authorizer, resourceManagerEndpoint string) {
	subs := az.Args["config"].(authentication.Config).SubscriptionID
	auth := az.Args["authorizer"].(autorest.Authorizer)
	resg := az.Args["resource_group"].(string)
	rEndpoint := az.Args["config"].(authentication.Config).CustomResourceManagerEndpoint
	return subs, resg, auth, rEndpoint
}

// getCredential returns the track-2 (azidentity) token credential.
// Migrated services use this instead of the track-1 autorest Authorizer.
func (az *AzureService) getCredential() azcore.TokenCredential {
	return az.Args["credential"].(azcore.TokenCredential)
}

func (az *AzureService) getSubscriptionID() string {
	return az.Args["config"].(authentication.Config).SubscriptionID
}

func (az *AzureService) getResourceGroup() string {
	return az.Args["resource_group"].(string)
}

func (az *AzureService) getResourceManagerEndpoint() string {
	return az.Args["config"].(authentication.Config).CustomResourceManagerEndpoint
}

// armClientOptions maps the configured ARM environment to the track-2
// cloud configuration used by all arm* clients.
func (az *AzureService) armClientOptions() *arm.ClientOptions {
	return &arm.ClientOptions{
		ClientOptions: azcore.ClientOptions{
			Cloud: azureCloudConfig(az.Args["config"].(authentication.Config).CustomResourceManagerEndpoint),
		},
	}
}

// azureCloudConfig maps a Resource Manager endpoint to a track-2 cloud
// configuration, mirroring the environment selection in setEnvConfig.
func azureCloudConfig(resourceManagerEndpoint string) cloud.Configuration {
	switch {
	case strings.Contains(resourceManagerEndpoint, "usgovcloudapi"):
		return cloud.AzureGovernment
	case strings.Contains(resourceManagerEndpoint, "chinacloudapi"):
		return cloud.AzureChina
	default:
		return cloud.AzurePublic
	}
}

func (az *AzureService) AppendSimpleResource(id string, resourceName string, resourceType string) {
	newResource := terraformutils.NewSimpleResource(id, resourceName, resourceType, az.ProviderName, []string{})
	az.Resources = append(az.Resources, newResource)
}

func (az *AzureService) AppendSimpleResourceWithDuplicateCheck(id string, resourceName string, resourceType string) {
	tferexist, _ := az.DuplicateCheck(id, resourceName, resourceType)
	if !tferexist {
		resourceName = resourceName + "_" + id
	}
	newResource := terraformutils.NewSimpleResource(id, resourceName, resourceType, az.ProviderName, []string{})
	az.Resources = append(az.Resources, newResource)
}

// This method checks if same resource name(tfer) exists with
// same id
func (az *AzureService) DuplicateCheck(id string, resourceName string, resourceType string) (bool, bool) {
	var tferexist, idexist bool
	tferName := terraformutils.TfSanitize(resourceName)
	for _, resource := range az.Resources {
		if tferName == resource.ResourceName {
			if id == resource.InstanceState.ID {
				tferexist = true
				idexist = true
			} else {
				tferexist = true
				idexist = false
			}
		}
	}
	return tferexist, idexist
}

func (az *AzureService) appendSimpleAssociation(id string, linkedResourceName string, resourceName *string, resourceType string, attributes map[string]string) {
	var resourceName2 string
	if resourceName != nil {
		resourceName2 = *resourceName
	} else {
		resourceName0 := strings.ReplaceAll(resourceType, "azurerm_", "")
		resourceName1 := resourceName0[strings.IndexByte(resourceName0, '_'):]
		resourceName2 = linkedResourceName + resourceName1
	}
	newResource := terraformutils.NewResource(
		id, resourceName2, resourceType, az.ProviderName, attributes,
		[]string{"name"},
		map[string]interface{}{},
	)
	az.Resources = append(az.Resources, newResource)
}
