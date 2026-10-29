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
	"github.com/aws/aws-sdk-go-v2/service/route53profiles"
)

var route53ProfilesAllowEmptyValues = []string{"tags."}

type Route53ProfilesGenerator struct {
	AWSService
}

// InitResources enumerates the Route 53 Profiles objects in the
// configured region: profiles, their VPC associations, and their
// resource associations. All three resources are keyed by their own
// ID.
func (g *Route53ProfilesGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	client := route53profiles.NewFromConfig(config)
	if err := g.addProfiles(client); err != nil {
		return err
	}
	if err := g.addAssociations(client); err != nil {
		return err
	}
	return g.addResourceAssociations(client)
}

func (g *Route53ProfilesGenerator) addProfiles(client *route53profiles.Client) error {
	input := &route53profiles.ListProfilesInput{}
	for {
		out, err := client.ListProfiles(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, profile := range out.ProfileSummaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(profile.Id),
				StringValue(profile.Name),
				"aws_route53profiles_profile",
				"aws",
				route53ProfilesAllowEmptyValues,
			))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *Route53ProfilesGenerator) addAssociations(client *route53profiles.Client) error {
	input := &route53profiles.ListProfileAssociationsInput{}
	for {
		out, err := client.ListProfileAssociations(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, association := range out.ProfileAssociations {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				StringValue(association.Id),
				StringValue(association.Name),
				"aws_route53profiles_association",
				"aws",
				map[string]string{"profile_id": StringValue(association.ProfileId)},
				route53ProfilesAllowEmptyValues,
				map[string]interface{}{},
			))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *Route53ProfilesGenerator) addResourceAssociations(client *route53profiles.Client) error {
	input := &route53profiles.ListProfileResourceAssociationsInput{}
	for {
		out, err := client.ListProfileResourceAssociations(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, association := range out.ProfileResourceAssociations {
			name := StringValue(association.Name)
			if name == "" {
				name = StringValue(association.Id)
			}
			g.Resources = append(g.Resources, terraformutils.NewResource(
				StringValue(association.Id),
				name,
				"aws_route53profiles_resource_association",
				"aws",
				map[string]string{"profile_id": StringValue(association.ProfileId)},
				route53ProfilesAllowEmptyValues,
				map[string]interface{}{},
			))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}
