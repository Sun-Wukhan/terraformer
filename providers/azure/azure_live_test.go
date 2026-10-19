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
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

// TestLiveTrack2ResourceGroups is a read-only smoke test for the track-2
// migration: it builds the azidentity credential the migrated services use
// and lists resource groups. It runs only when TERRAFORMER_AZURE_LIVE_TEST=1
// with credentials present (CI wires AZURE_* secrets; locally ARM_* also works).
func TestLiveTrack2ResourceGroups(t *testing.T) {
	if os.Getenv("TERRAFORMER_AZURE_LIVE_TEST") != "1" {
		t.Skip("set TERRAFORMER_AZURE_LIVE_TEST=1 to run live Azure tests")
	}
	// Accept the ARM_* vars terraformer has always used.
	for _, pair := range [][2]string{
		{"ARM_CLIENT_ID", "AZURE_CLIENT_ID"},
		{"ARM_TENANT_ID", "AZURE_TENANT_ID"},
		{"ARM_CLIENT_SECRET", "AZURE_CLIENT_SECRET"},
		{"ARM_SUBSCRIPTION_ID", "AZURE_SUBSCRIPTION_ID"},
	} {
		if v := os.Getenv(pair[0]); v != "" {
			if err := os.Setenv(pair[1], v); err != nil {
				t.Fatalf("setenv %s: %v", pair[1], err)
			}
		}
	}
	sub := os.Getenv("AZURE_SUBSCRIPTION_ID")
	if sub == "" {
		t.Fatal("AZURE_SUBSCRIPTION_ID (or ARM_SUBSCRIPTION_ID) must be set")
	}
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		t.Fatalf("credential: %v", err)
	}
	client, err := armresources.NewResourceGroupsClient(sub, cred, &arm.ClientOptions{})
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	pager := client.NewListPager(nil)
	ctx := context.Background()
	if !pager.More() {
		t.Fatal("expected at least one page of resource groups")
	}
	page, err := pager.NextPage(ctx)
	if err != nil {
		t.Fatalf("list resource groups: %v", err)
	}
	t.Logf("track-2 OK: listed %d resource group(s)", len(page.Value))
}
