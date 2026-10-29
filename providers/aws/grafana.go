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

// Amazon Managed Grafana generator (service key "grafana").
//
// Skipped types in this family (see results ledger):
//   - aws_grafana_role_association: the provider resource has no Import
//     section (no importer); associations are also only readable as the
//     aggregated ListPermissions view per workspace/role.
//   - aws_grafana_workspace_api_key: no provider importer (the provider
//     documents no Import section; API keys are write-only secrets and the
//     Grafana API exposes no list operation for them).
//   - aws_grafana_workspace_service_account_token: no provider importer
//     (no Import section documented; tokens are write-only secrets).

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/grafana"
	"github.com/aws/aws-sdk-go-v2/service/grafana/types"
)

var grafanaAllowEmptyValues = []string{"tags."}

type GrafanaGenerator struct {
	AWSService
}

func (g *GrafanaGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := grafana.NewFromConfig(config)
	var nextToken *string
	for {
		out, err := svc.ListWorkspaces(context.TODO(), &grafana.ListWorkspacesInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, workspace := range out.Workspaces {
			id := StringValue(workspace.Id)
			name := StringValue(workspace.Name)
			if name == "" {
				name = id
			}
			// aws_grafana_workspace — importer ID is the workspace ID.
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, name, "aws_grafana_workspace", "aws", grafanaAllowEmptyValues))
			if err := g.createWorkspaceChildren(svc, workspace); err != nil {
				return err
			}
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

func (g *GrafanaGenerator) createWorkspaceChildren(svc *grafana.Client, workspace types.WorkspaceSummary) error {
	workspaceID := StringValue(workspace.Id)
	// aws_grafana_license_association — importer ID is the workspace ID;
	// emitted only when the workspace carries a license type.
	if workspace.LicenseType != "" {
		g.Resources = append(g.Resources, terraformutils.NewResource(
			workspaceID, workspaceID, "aws_grafana_license_association", "aws",
			map[string]string{"workspace_id": workspaceID}, grafanaAllowEmptyValues, map[string]interface{}{}))
	}
	// aws_grafana_workspace_saml_configuration — importer ID is the workspace
	// ID; emitted only when SAML is configured for the workspace.
	auth, err := svc.DescribeWorkspaceAuthentication(context.TODO(), &grafana.DescribeWorkspaceAuthenticationInput{WorkspaceId: workspace.Id})
	if err != nil {
		return err
	}
	if auth.Authentication != nil && auth.Authentication.Saml != nil &&
		auth.Authentication.Saml.Status != types.SamlConfigurationStatusNotConfigured {
		g.Resources = append(g.Resources, terraformutils.NewResource(
			workspaceID, workspaceID, "aws_grafana_workspace_saml_configuration", "aws",
			map[string]string{"workspace_id": workspaceID}, grafanaAllowEmptyValues, map[string]interface{}{}))
	}
	// aws_grafana_workspace_service_account — importer ID
	// "workspace_id,service_account_id".
	var nextToken *string
	for {
		out, err := svc.ListWorkspaceServiceAccounts(context.TODO(), &grafana.ListWorkspaceServiceAccountsInput{WorkspaceId: workspace.Id, NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, account := range out.ServiceAccounts {
			id := workspaceID + "," + StringValue(account.Id)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				id, StringValue(account.Name), "aws_grafana_workspace_service_account", "aws",
				map[string]string{"workspace_id": workspaceID}, grafanaAllowEmptyValues, map[string]interface{}{}))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
