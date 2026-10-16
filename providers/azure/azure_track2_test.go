package azure

import (
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
)

// Verifies every registered service can be initialized (Args wiring) without
// credentials, and that migrated services receive the track-2 credential key.
func TestInitServiceWiring(t *testing.T) {
	p := &AzureProvider{}
	for name, gen := range p.GetSupportedService() {
		if err := p.InitService(name, false); err != nil {
			t.Fatalf("InitService(%s): %v", name, err)
		}
		if _, ok := p.Service.GetArgs()["credential"]; !ok {
			t.Errorf("service %s missing credential arg", name)
		}
		_ = gen
	}
}

func TestAzureCloudConfig(t *testing.T) {
	cases := map[string]cloud.Configuration{
		"https://management.azure.com/":                 cloud.AzurePublic,
		"https://management.usgovcloudapi.net/":        cloud.AzureGovernment,
		"https://management.chinacloudapi.cn/":         cloud.AzureChina,
		"https://management.custom.example.com/":       cloud.AzurePublic,
		"":                                             cloud.AzurePublic,
	}
	for endpoint, want := range cases {
		if got := azureCloudConfig(endpoint); !reflect.DeepEqual(got, want) {
			t.Errorf("azureCloudConfig(%q) mismatch", endpoint)
		}
	}
}

func TestParseAzureResourceIDSanity(t *testing.T) {
	id, err := ParseAzureResourceID("/subscriptions/sub1/resourceGroups/rg1/providers/Microsoft.Network/virtualNetworks/vnet1")
	if err != nil {
		t.Fatal(err)
	}
	if id.SubscriptionID != "sub1" || id.ResourceGroup != "rg1" {
		t.Fatalf("unexpected parse result: %+v", id)
	}
}
