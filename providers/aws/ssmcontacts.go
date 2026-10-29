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
	"github.com/aws/aws-sdk-go-v2/service/ssmcontacts"
)

var ssmcontactsAllowEmptyValues = []string{"tags."}

// SsmContactsGenerator covers Incident Manager Contacts: contacts, contact
// channels, contact plans and rotations.
// Provider import IDs (verified): contact = ARN; contact channel = ARN;
// plan = contact ARN; rotation = ARN.
type SsmContactsGenerator struct {
	AWSService
}

func (g *SsmContactsGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := ssmcontacts.NewFromConfig(config)
	p := ssmcontacts.NewListContactsPaginator(svc, &ssmcontacts.ListContactsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, contact := range page.Contacts {
			arn := StringValue(contact.ContactArn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_ssmcontacts_contact", "aws", ssmcontactsAllowEmptyValues))
			// A contact plan exists per contact; its provider ID is the contact ARN.
			if _, err := svc.GetContactPolicy(context.TODO(), &ssmcontacts.GetContactPolicyInput{ContactArn: contact.ContactArn}); err == nil {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					arn, arn, "aws_ssmcontacts_plan", "aws", ssmcontactsAllowEmptyValues))
			}
			cp := ssmcontacts.NewListContactChannelsPaginator(svc, &ssmcontacts.ListContactChannelsInput{ContactId: contact.ContactArn})
			for cp.HasMorePages() {
				cpage, err := cp.NextPage(context.TODO())
				if err != nil {
					break
				}
				for _, ch := range cpage.ContactChannels {
					chARN := StringValue(ch.ContactChannelArn)
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						chARN, chARN, "aws_ssmcontacts_contact_channel", "aws", ssmcontactsAllowEmptyValues))
				}
			}
		}
	}
	rp := ssmcontacts.NewListRotationsPaginator(svc, &ssmcontacts.ListRotationsInput{})
	for rp.HasMorePages() {
		page, err := rp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, rot := range page.Rotations {
			arn := StringValue(rot.RotationArn)
			name := StringValue(rot.Name)
			if name == "" {
				name = arn
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, name, "aws_ssmcontacts_rotation", "aws", ssmcontactsAllowEmptyValues))
		}
	}
	return nil
}
