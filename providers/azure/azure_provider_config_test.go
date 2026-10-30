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
	"testing"

	"github.com/hashicorp/go-azure-helpers/authentication"
	"github.com/zclconf/go-cty/cty"
)

func TestAzureProviderGetConfig(t *testing.T) {
	p := &AzureProvider{config: authentication.Config{SubscriptionID: "sub-1"}}
	got := p.GetConfig().AsValueMap()

	features, ok := got["features"]
	if !ok || !features.Type().IsObjectType() || len(features.Type().AttributeTypes()) != 0 {
		t.Fatalf("features must be an empty object (rendered as `features {}`), got %#v", features)
	}
	if sub := got["subscription_id"]; !sub.RawEquals(cty.StringVal("sub-1")) {
		t.Fatalf("subscription_id = %#v, want sub-1", sub)
	}

	if _, ok := (&AzureProvider{}).GetConfig().AsValueMap()["subscription_id"]; ok {
		t.Fatalf("subscription_id must be omitted when unknown")
	}
}
