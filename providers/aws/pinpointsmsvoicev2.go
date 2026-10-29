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
	"log"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2"
)

var pinpointSmsVoiceV2AllowEmptyValues = []string{"tags."}

// PinpointSmsVoiceV2Generator covers End User Messaging SMS
// (aws_pinpointsmsvoicev2_*). Import IDs verified against the Terraform AWS
// provider docs: configuration sets, opt-out lists, phone numbers and pools
// import with name/ID; event destinations with
// configuration_set_name,event_destination_name; keywords with
// origination_identity_arn,keyword; sender IDs with sender_id,iso_country_code;
// resource policies with the parent resource ARN.
type PinpointSmsVoiceV2Generator struct {
	AWSService
}

func (g *PinpointSmsVoiceV2Generator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := pinpointsmsvoicev2.NewFromConfig(config)

	resourceArns, err := g.loadSmsVoiceConfigurationSets(svc)
	if err != nil {
		return err
	}
	if err := g.loadSmsVoiceOptOutLists(svc); err != nil {
		return err
	}
	phoneArns, err := g.loadSmsVoicePhoneNumbers(svc)
	if err != nil {
		return err
	}
	poolArns, err := g.loadSmsVoicePools(svc)
	if err != nil {
		return err
	}
	if err := g.loadSmsVoiceSenderIDs(svc); err != nil {
		return err
	}

	originationArns := append(append([]string{}, phoneArns...), poolArns...)
	for _, arn := range originationArns {
		if err := g.loadSmsVoiceKeywords(svc, arn); err != nil {
			return err
		}
	}
	resourceArns = append(resourceArns, originationArns...)
	g.loadSmsVoiceResourcePolicies(svc, resourceArns)
	return nil
}

func (g *PinpointSmsVoiceV2Generator) loadSmsVoiceConfigurationSets(svc *pinpointsmsvoicev2.Client) ([]string, error) {
	var arns []string
	p := pinpointsmsvoicev2.NewDescribeConfigurationSetsPaginator(svc, &pinpointsmsvoicev2.DescribeConfigurationSetsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return arns, err
		}
		for _, set := range page.ConfigurationSets {
			name := StringValue(set.ConfigurationSetName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name,
				name,
				"aws_pinpointsmsvoicev2_configuration_set",
				"aws",
				pinpointSmsVoiceV2AllowEmptyValues))
			arns = append(arns, StringValue(set.ConfigurationSetArn))
			for _, destination := range set.EventDestinations {
				id := name + "," + StringValue(destination.EventDestinationName)
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id,
					StringValue(destination.EventDestinationName),
					"aws_pinpointsmsvoicev2_event_destination",
					"aws",
					pinpointSmsVoiceV2AllowEmptyValues))
			}
		}
	}
	return arns, nil
}

func (g *PinpointSmsVoiceV2Generator) loadSmsVoiceOptOutLists(svc *pinpointsmsvoicev2.Client) error {
	p := pinpointsmsvoicev2.NewDescribeOptOutListsPaginator(svc, &pinpointsmsvoicev2.DescribeOptOutListsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, list := range page.OptOutLists {
			name := StringValue(list.OptOutListName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name,
				name,
				"aws_pinpointsmsvoicev2_opt_out_list",
				"aws",
				pinpointSmsVoiceV2AllowEmptyValues))
		}
	}
	return nil
}

func (g *PinpointSmsVoiceV2Generator) loadSmsVoicePhoneNumbers(svc *pinpointsmsvoicev2.Client) ([]string, error) {
	var arns []string
	p := pinpointsmsvoicev2.NewDescribePhoneNumbersPaginator(svc, &pinpointsmsvoicev2.DescribePhoneNumbersInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return arns, err
		}
		for _, number := range page.PhoneNumbers {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(number.PhoneNumberId),
				StringValue(number.PhoneNumber),
				"aws_pinpointsmsvoicev2_phone_number",
				"aws",
				pinpointSmsVoiceV2AllowEmptyValues))
			arns = append(arns, StringValue(number.PhoneNumberArn))
		}
	}
	return arns, nil
}

func (g *PinpointSmsVoiceV2Generator) loadSmsVoicePools(svc *pinpointsmsvoicev2.Client) ([]string, error) {
	var arns []string
	p := pinpointsmsvoicev2.NewDescribePoolsPaginator(svc, &pinpointsmsvoicev2.DescribePoolsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return arns, err
		}
		for _, pool := range page.Pools {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(pool.PoolId),
				StringValue(pool.PoolId),
				"aws_pinpointsmsvoicev2_pool",
				"aws",
				pinpointSmsVoiceV2AllowEmptyValues))
			arns = append(arns, StringValue(pool.PoolArn))
		}
	}
	return arns, nil
}

func (g *PinpointSmsVoiceV2Generator) loadSmsVoiceSenderIDs(svc *pinpointsmsvoicev2.Client) error {
	p := pinpointsmsvoicev2.NewDescribeSenderIdsPaginator(svc, &pinpointsmsvoicev2.DescribeSenderIdsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, sender := range page.SenderIds {
			id := StringValue(sender.SenderId) + "," + StringValue(sender.IsoCountryCode)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				StringValue(sender.SenderId),
				"aws_pinpointsmsvoicev2_sender_id",
				"aws",
				pinpointSmsVoiceV2AllowEmptyValues))
		}
	}
	return nil
}

func (g *PinpointSmsVoiceV2Generator) loadSmsVoiceKeywords(svc *pinpointsmsvoicev2.Client, originationArn string) error {
	p := pinpointsmsvoicev2.NewDescribeKeywordsPaginator(svc, &pinpointsmsvoicev2.DescribeKeywordsInput{
		OriginationIdentity: &originationArn,
	})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, keyword := range page.Keywords {
			id := originationArn + "," + StringValue(keyword.Keyword)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id,
				StringValue(keyword.Keyword),
				"aws_pinpointsmsvoicev2_keyword",
				"aws",
				pinpointSmsVoiceV2AllowEmptyValues))
		}
	}
	return nil
}

func (g *PinpointSmsVoiceV2Generator) loadSmsVoiceResourcePolicies(svc *pinpointsmsvoicev2.Client, arns []string) {
	for _, arn := range arns {
		if arn == "" {
			continue
		}
		out, err := svc.GetResourcePolicy(context.TODO(), &pinpointsmsvoicev2.GetResourcePolicyInput{ResourceArn: &arn})
		if err != nil {
			// Most resources have no resource policy; the API errors then.
			log.Println(err)
			continue
		}
		if out.Policy != nil && *out.Policy != "" {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn,
				arn,
				"aws_pinpointsmsvoicev2_resource_policy",
				"aws",
				pinpointSmsVoiceV2AllowEmptyValues))
		}
	}
}
