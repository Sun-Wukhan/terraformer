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
	"github.com/aws/aws-sdk-go-v2/service/transfer"
)

// Transfer Family sweep. Importer IDs verified against provider docs:
// server = server ID; user = "server-id/user-name"; SSH key =
// "server-id/user-name/key-id"; access = "server-id/external-id";
// agreement = "server-id/agreement-id"; host key = "server-id,host-key-id"
// (comma); certificate / connector / profile / workflow / web app = their
// own IDs; web app customization = the web app ID.
//
// Skipped in this family (see results ledger):
//   - aws_transfer_tag: a tag-assignment glue resource (resource ARN +
//     key/value); tags are attributes of their parent resources, not an
//     enumerable standalone object, and the resource has no list API of
//     its own.

var transferAllowEmptyValues = []string{"tags."}

type TransferGenerator struct {
	AWSService
}

func (g *TransferGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := transfer.NewFromConfig(config)
	ctx := context.TODO()
	add := func(id, name, typ string) {
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(id, name, typ, "aws", transferAllowEmptyValues))
	}

	serverP := transfer.NewListServersPaginator(svc, &transfer.ListServersInput{})
	for serverP.HasMorePages() {
		page, err := serverP.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, server := range page.Servers {
			serverID := StringValue(server.ServerId)
			add(serverID, serverID, "aws_transfer_server")
			userP := transfer.NewListUsersPaginator(svc, &transfer.ListUsersInput{ServerId: server.ServerId})
			for userP.HasMorePages() {
				uPage, err := userP.NextPage(ctx)
				if err != nil {
					break
				}
				for _, user := range uPage.Users {
					userID := serverID + "/" + StringValue(user.UserName)
					add(userID, StringValue(user.UserName), "aws_transfer_user")
					if desc, err := svc.DescribeUser(ctx, &transfer.DescribeUserInput{ServerId: server.ServerId, UserName: user.UserName}); err == nil && desc.User != nil {
						for _, key := range desc.User.SshPublicKeys {
							keyID := userID + "/" + StringValue(key.SshPublicKeyId)
							add(keyID, StringValue(key.SshPublicKeyId), "aws_transfer_ssh_key")
						}
					}
				}
			}
			accessP := transfer.NewListAccessesPaginator(svc, &transfer.ListAccessesInput{ServerId: server.ServerId})
			for accessP.HasMorePages() {
				aPage, err := accessP.NextPage(ctx)
				if err != nil {
					break
				}
				for _, access := range aPage.Accesses {
					id := serverID + "/" + StringValue(access.ExternalId)
					add(id, StringValue(access.ExternalId), "aws_transfer_access")
				}
			}
			agrInput := &transfer.ListAgreementsInput{ServerId: server.ServerId}
			for {
				agrOut, err := svc.ListAgreements(ctx, agrInput)
				if err != nil {
					break
				}
				for _, agr := range agrOut.Agreements {
					id := serverID + "/" + StringValue(agr.AgreementId)
					add(id, StringValue(agr.AgreementId), "aws_transfer_agreement")
				}
				if agrOut.NextToken == nil {
					break
				}
				agrInput.NextToken = agrOut.NextToken
			}
			hkInput := &transfer.ListHostKeysInput{ServerId: server.ServerId}
			for {
				hkOut, err := svc.ListHostKeys(ctx, hkInput)
				if err != nil {
					break
				}
				for _, hk := range hkOut.HostKeys {
					id := serverID + "," + StringValue(hk.HostKeyId)
					add(id, StringValue(hk.HostKeyId), "aws_transfer_host_key")
				}
				if hkOut.NextToken == nil {
					break
				}
				hkInput.NextToken = hkOut.NextToken
			}
		}
	}

	certP := transfer.NewListCertificatesPaginator(svc, &transfer.ListCertificatesInput{})
	for certP.HasMorePages() {
		page, err := certP.NextPage(ctx)
		if err != nil {
			break
		}
		for _, cert := range page.Certificates {
			add(StringValue(cert.CertificateId), StringValue(cert.CertificateId), "aws_transfer_certificate")
		}
	}
	connP := transfer.NewListConnectorsPaginator(svc, &transfer.ListConnectorsInput{})
	for connP.HasMorePages() {
		page, err := connP.NextPage(ctx)
		if err != nil {
			break
		}
		for _, conn := range page.Connectors {
			add(StringValue(conn.ConnectorId), StringValue(conn.ConnectorId), "aws_transfer_connector")
		}
	}
	profP := transfer.NewListProfilesPaginator(svc, &transfer.ListProfilesInput{})
	for profP.HasMorePages() {
		page, err := profP.NextPage(ctx)
		if err != nil {
			break
		}
		for _, prof := range page.Profiles {
			add(StringValue(prof.ProfileId), StringValue(prof.ProfileId), "aws_transfer_profile")
		}
	}
	wfP := transfer.NewListWorkflowsPaginator(svc, &transfer.ListWorkflowsInput{})
	for wfP.HasMorePages() {
		page, err := wfP.NextPage(ctx)
		if err != nil {
			break
		}
		for _, wf := range page.Workflows {
			add(StringValue(wf.WorkflowId), StringValue(wf.WorkflowId), "aws_transfer_workflow")
		}
	}
	waInput := &transfer.ListWebAppsInput{}
	for {
		out, err := svc.ListWebApps(ctx, waInput)
		if err != nil {
			break
		}
		for _, app := range out.WebApps {
			appID := StringValue(app.WebAppId)
			add(appID, appID, "aws_transfer_web_app")
			if _, err := svc.DescribeWebAppCustomization(ctx, &transfer.DescribeWebAppCustomizationInput{WebAppId: app.WebAppId}); err == nil {
				add(appID, appID, "aws_transfer_web_app_customization")
			}
		}
		if out.NextToken == nil {
			break
		}
		waInput.NextToken = out.NextToken
	}
	return nil
}
