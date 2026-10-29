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

package aws

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/rolesanywhere"
)

var rolesAnywhereAllowEmptyValues = []string{"tags."}

type RolesAnywhereGenerator struct {
	AWSService
}

// InitResources enumerates IAM Roles Anywhere profiles and trust
// anchors in the configured region; the provider importer IDs are
// the profile ID and the trust anchor ID respectively.
func (g *RolesAnywhereGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := rolesanywhere.NewFromConfig(config)

	profiles, err := svc.ListProfiles(context.TODO(), &rolesanywhere.ListProfilesInput{})
	if err != nil {
		return err
	}
	for _, profile := range profiles.Profiles {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			StringValue(profile.ProfileId),
			StringValue(profile.Name),
			"aws_rolesanywhere_profile",
			"aws",
			rolesAnywhereAllowEmptyValues))
	}

	anchors, err := svc.ListTrustAnchors(context.TODO(), &rolesanywhere.ListTrustAnchorsInput{})
	if err != nil {
		return err
	}
	for _, anchor := range anchors.TrustAnchors {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			StringValue(anchor.TrustAnchorId),
			StringValue(anchor.Name),
			"aws_rolesanywhere_trust_anchor",
			"aws",
			rolesAnywhereAllowEmptyValues))
	}
	return nil
}
