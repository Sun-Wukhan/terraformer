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
	"github.com/aws/aws-sdk-go-v2/service/chimesdkvoice"
)

var chimeAllowEmptyValues = []string{"tags."}

// ChimeGenerator covers Amazon Chime voice resources (aws_chime_* and
// aws_chimesdkvoice_*), enumerated with the Chime SDK Voice client. The
// per-connector configuration resources (logging, origination, streaming,
// termination) import with the voice connector ID, as do the connectors
// themselves in the provider.
//
// Skipped in this family:
//   - aws_chime_voice_connector_termination_credentials: the API returns
//     only credential usernames; passwords are write-only and cannot be
//     enumerated, so the resource cannot be faithfully imported.
//   - aws_chimesdkvoice_global_settings: singleton account-level settings.
type ChimeGenerator struct {
	AWSService
}

func (g *ChimeGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := chimesdkvoice.NewFromConfig(config)

	if err := g.loadVoiceConnectors(svc); err != nil {
		return err
	}
	if err := g.loadVoiceConnectorGroups(svc); err != nil {
		return err
	}
	if err := g.loadSipMediaApplications(svc); err != nil {
		return err
	}
	if err := g.loadSipRules(svc); err != nil {
		return err
	}
	if err := g.loadVoiceProfileDomains(svc); err != nil {
		return err
	}
	return nil
}

func (g *ChimeGenerator) loadVoiceConnectors(svc *chimesdkvoice.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListVoiceConnectors(context.TODO(), &chimesdkvoice.ListVoiceConnectorsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, connector := range out.VoiceConnectors {
			connectorID := StringValue(connector.VoiceConnectorId)
			name := StringValue(connector.Name)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				connectorID,
				name,
				"aws_chime_voice_connector",
				"aws",
				chimeAllowEmptyValues))
			for _, configType := range []string{
				"aws_chime_voice_connector_logging",
				"aws_chime_voice_connector_origination",
				"aws_chime_voice_connector_streaming",
				"aws_chime_voice_connector_termination",
			} {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					connectorID,
					name,
					configType,
					"aws",
					chimeAllowEmptyValues))
			}
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

func (g *ChimeGenerator) loadVoiceConnectorGroups(svc *chimesdkvoice.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListVoiceConnectorGroups(context.TODO(), &chimesdkvoice.ListVoiceConnectorGroupsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, group := range out.VoiceConnectorGroups {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(group.VoiceConnectorGroupId),
				StringValue(group.Name),
				"aws_chime_voice_connector_group",
				"aws",
				chimeAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

func (g *ChimeGenerator) loadSipMediaApplications(svc *chimesdkvoice.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListSipMediaApplications(context.TODO(), &chimesdkvoice.ListSipMediaApplicationsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, app := range out.SipMediaApplications {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(app.SipMediaApplicationId),
				StringValue(app.Name),
				"aws_chimesdkvoice_sip_media_application",
				"aws",
				chimeAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

func (g *ChimeGenerator) loadSipRules(svc *chimesdkvoice.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListSipRules(context.TODO(), &chimesdkvoice.ListSipRulesInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, rule := range out.SipRules {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(rule.SipRuleId),
				StringValue(rule.Name),
				"aws_chimesdkvoice_sip_rule",
				"aws",
				chimeAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

func (g *ChimeGenerator) loadVoiceProfileDomains(svc *chimesdkvoice.Client) error {
	var nextToken *string
	for {
		out, err := svc.ListVoiceProfileDomains(context.TODO(), &chimesdkvoice.ListVoiceProfileDomainsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, domain := range out.VoiceProfileDomains {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(domain.VoiceProfileDomainId),
				StringValue(domain.Name),
				"aws_chimesdkvoice_voice_profile_domain",
				"aws",
				chimeAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
