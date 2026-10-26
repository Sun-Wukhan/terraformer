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
	"errors"
	"log"
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	quicksighttypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
)

var quicksightAllowEmptyValues = []string{"tags."}

type QuickSightGenerator struct {
	AWSService
}

// InitResources enumerates the Amazon QuickSight account: the account-level
// singletons (account settings, account subscription, IP restriction and
// customer managed key registration), namespaces, the content assets
// (analyses, dashboards, data sets with their ingestions and refresh
// schedules, data sources, folders and their memberships, templates and
// their aliases, themes, VPC connections), custom permission profiles, and
// the per-namespace identity resources (users, groups and memberships, IAM
// policy assignments, role custom permissions and role memberships).
func (g *QuickSightGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	accountIDPtr, e := g.getAccountNumber(config)
	if e != nil {
		return e
	}
	accountID := *accountIDPtr
	client := quicksight.NewFromConfig(config)

	namespaces, err := g.addNamespaces(client, accountID)
	if err != nil {
		return err
	}
	for _, add := range []func(*quicksight.Client, string) error{
		g.addAccountSettings,
		g.addAccountSubscription,
		g.addIPRestriction,
		g.addKeyRegistration,
		g.addAnalyses,
		g.addDashboards,
		g.addDataSets,
		g.addDataSources,
		g.addFolders,
		g.addTemplates,
		g.addThemes,
		g.addVPCConnections,
		g.addCustomPermissions,
	} {
		if err := add(client, accountID); err != nil {
			return err
		}
	}
	for _, namespace := range namespaces {
		if err := g.addNamespaceResources(client, accountID, namespace); err != nil {
			return err
		}
	}
	return nil
}

func (g *QuickSightGenerator) addNamespaces(client *quicksight.Client, accountID string) ([]string, error) {
	var namespaces []string
	input := &quicksight.ListNamespacesInput{
		AwsAccountId: aws.String(accountID),
	}
	for {
		page, err := client.ListNamespaces(context.TODO(), input)
		if err != nil {
			return nil, err
		}
		for _, namespace := range page.Namespaces {
			name := *namespace.Name
			namespaces = append(namespaces, name)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				accountID+","+name,
				name,
				"aws_quicksight_namespace",
				"aws",
				map[string]string{"aws_account_id": accountID},
				quicksightAllowEmptyValues,
				map[string]interface{}{},
			))
		}
		if page.NextToken == nil {
			return namespaces, nil
		}
		input.NextToken = page.NextToken
	}
}

// The account settings are a per-account singleton keyed by account ID.
func (g *QuickSightGenerator) addAccountSettings(client *quicksight.Client, accountID string) error {
	if _, err := client.DescribeAccountSettings(context.TODO(), &quicksight.DescribeAccountSettingsInput{
		AwsAccountId: aws.String(accountID),
	}); err != nil {
		return err
	}
	g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
		accountID,
		"quicksight_account_settings",
		"aws_quicksight_account_settings",
		"aws",
		quicksightAllowEmptyValues,
	))
	return nil
}

// The account subscription is a per-account singleton keyed by account ID.
func (g *QuickSightGenerator) addAccountSubscription(client *quicksight.Client, accountID string) error {
	out, err := client.DescribeAccountSubscription(context.TODO(), &quicksight.DescribeAccountSubscriptionInput{
		AwsAccountId: aws.String(accountID),
	})
	if err != nil {
		return err
	}
	name := "quicksight_account_subscription"
	if out.AccountInfo != nil && out.AccountInfo.AccountName != nil {
		name = *out.AccountInfo.AccountName
	}
	g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
		accountID,
		name,
		"aws_quicksight_account_subscription",
		"aws",
		quicksightAllowEmptyValues,
	))
	return nil
}

// The IP restriction configuration is a per-account singleton keyed by
// account ID.
func (g *QuickSightGenerator) addIPRestriction(client *quicksight.Client, accountID string) error {
	if _, err := client.DescribeIpRestriction(context.TODO(), &quicksight.DescribeIpRestrictionInput{
		AwsAccountId: aws.String(accountID),
	}); err != nil {
		return err
	}
	g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
		accountID,
		"quicksight_ip_restriction",
		"aws_quicksight_ip_restriction",
		"aws",
		quicksightAllowEmptyValues,
	))
	return nil
}

// The key registration is a per-account singleton keyed by account ID; it
// only exists as a managed resource when at least one customer managed key
// is registered.
func (g *QuickSightGenerator) addKeyRegistration(client *quicksight.Client, accountID string) error {
	out, err := client.DescribeKeyRegistration(context.TODO(), &quicksight.DescribeKeyRegistrationInput{
		AwsAccountId: aws.String(accountID),
	})
	if err != nil {
		return err
	}
	if len(out.KeyRegistration) == 0 {
		return nil
	}
	g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
		accountID,
		"quicksight_key_registration",
		"aws_quicksight_key_registration",
		"aws",
		quicksightAllowEmptyValues,
	))
	return nil
}

// Provider ID format: AWS account ID, analysis ID.
func (g *QuickSightGenerator) addAnalyses(client *quicksight.Client, accountID string) error {
	input := &quicksight.ListAnalysesInput{
		AwsAccountId: aws.String(accountID),
	}
	for {
		page, err := client.ListAnalyses(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, analysis := range page.AnalysisSummaryList {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				accountID+","+*analysis.AnalysisId,
				*analysis.Name,
				"aws_quicksight_analysis",
				"aws",
				map[string]string{"aws_account_id": accountID},
				quicksightAllowEmptyValues,
				map[string]interface{}{},
			))
		}
		if page.NextToken == nil {
			return nil
		}
		input.NextToken = page.NextToken
	}
}

// Provider ID format: AWS account ID, dashboard ID.
func (g *QuickSightGenerator) addDashboards(client *quicksight.Client, accountID string) error {
	input := &quicksight.ListDashboardsInput{
		AwsAccountId: aws.String(accountID),
	}
	for {
		page, err := client.ListDashboards(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, dashboard := range page.DashboardSummaryList {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				accountID+","+*dashboard.DashboardId,
				*dashboard.Name,
				"aws_quicksight_dashboard",
				"aws",
				map[string]string{"aws_account_id": accountID},
				quicksightAllowEmptyValues,
				map[string]interface{}{},
			))
		}
		if page.NextToken == nil {
			return nil
		}
		input.NextToken = page.NextToken
	}
}

// Provider ID format: AWS account ID, data set ID.
func (g *QuickSightGenerator) addDataSets(client *quicksight.Client, accountID string) error {
	input := &quicksight.ListDataSetsInput{
		AwsAccountId: aws.String(accountID),
	}
	for {
		page, err := client.ListDataSets(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, dataSet := range page.DataSetSummaries {
			dataSetID := *dataSet.DataSetId
			g.Resources = append(g.Resources, terraformutils.NewResource(
				accountID+","+dataSetID,
				*dataSet.Name,
				"aws_quicksight_data_set",
				"aws",
				map[string]string{"aws_account_id": accountID},
				quicksightAllowEmptyValues,
				map[string]interface{}{},
			))
			if err := g.addIngestions(client, accountID, dataSetID); err != nil {
				return err
			}
			if err := g.addRefreshSchedules(client, accountID, dataSetID); err != nil {
				return err
			}
		}
		if page.NextToken == nil {
			return nil
		}
		input.NextToken = page.NextToken
	}
}

// Provider ID format: AWS account ID, data set ID, ingestion ID.
func (g *QuickSightGenerator) addIngestions(client *quicksight.Client, accountID, dataSetID string) error {
	input := &quicksight.ListIngestionsInput{
		AwsAccountId: aws.String(accountID),
		DataSetId:    aws.String(dataSetID),
	}
	for {
		page, err := client.ListIngestions(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, ingestion := range page.Ingestions {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				accountID+","+dataSetID+","+*ingestion.IngestionId,
				*ingestion.IngestionId,
				"aws_quicksight_ingestion",
				"aws",
				map[string]string{
					"aws_account_id": accountID,
					"data_set_id":    dataSetID,
				},
				quicksightAllowEmptyValues,
				map[string]interface{}{},
			))
		}
		if page.NextToken == nil {
			return nil
		}
		input.NextToken = page.NextToken
	}
}

// Provider ID format: AWS account ID, data set ID, schedule ID. The list
// API is not paginated.
func (g *QuickSightGenerator) addRefreshSchedules(client *quicksight.Client, accountID, dataSetID string) error {
	page, err := client.ListRefreshSchedules(context.TODO(), &quicksight.ListRefreshSchedulesInput{
		AwsAccountId: aws.String(accountID),
		DataSetId:    aws.String(dataSetID),
	})
	if err != nil {
		return err
	}
	for _, schedule := range page.RefreshSchedules {
		g.Resources = append(g.Resources, terraformutils.NewResource(
			accountID+","+dataSetID+","+*schedule.ScheduleId,
			*schedule.ScheduleId,
			"aws_quicksight_refresh_schedule",
			"aws",
			map[string]string{
				"aws_account_id": accountID,
				"data_set_id":    dataSetID,
			},
			quicksightAllowEmptyValues,
			map[string]interface{}{},
		))
	}
	return nil
}

// Provider ID format: AWS account ID, data source ID.
func (g *QuickSightGenerator) addDataSources(client *quicksight.Client, accountID string) error {
	input := &quicksight.ListDataSourcesInput{
		AwsAccountId: aws.String(accountID),
	}
	for {
		page, err := client.ListDataSources(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, dataSource := range page.DataSources {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				accountID+","+*dataSource.DataSourceId,
				*dataSource.Name,
				"aws_quicksight_data_source",
				"aws",
				map[string]string{"aws_account_id": accountID},
				quicksightAllowEmptyValues,
				map[string]interface{}{},
			))
		}
		if page.NextToken == nil {
			return nil
		}
		input.NextToken = page.NextToken
	}
}

// Provider ID format: AWS account ID, folder ID.
func (g *QuickSightGenerator) addFolders(client *quicksight.Client, accountID string) error {
	input := &quicksight.ListFoldersInput{
		AwsAccountId: aws.String(accountID),
	}
	for {
		page, err := client.ListFolders(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, folder := range page.FolderSummaryList {
			folderID := *folder.FolderId
			g.Resources = append(g.Resources, terraformutils.NewResource(
				accountID+","+folderID,
				*folder.Name,
				"aws_quicksight_folder",
				"aws",
				map[string]string{"aws_account_id": accountID},
				quicksightAllowEmptyValues,
				map[string]interface{}{},
			))
			if err := g.addFolderMemberships(client, accountID, folderID); err != nil {
				return err
			}
		}
		if page.NextToken == nil {
			return nil
		}
		input.NextToken = page.NextToken
	}
}

// Provider ID format: AWS account ID, folder ID, member type, member ID.
// The list API returns member ID/ARN pairs only; the member type the
// provider requires is the resource type segment of the member ARN.
func (g *QuickSightGenerator) addFolderMemberships(client *quicksight.Client, accountID, folderID string) error {
	input := &quicksight.ListFolderMembersInput{
		AwsAccountId: aws.String(accountID),
		FolderId:     aws.String(folderID),
	}
	for {
		page, err := client.ListFolderMembers(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, member := range page.FolderMemberList {
			memberType := folderMemberType(*member.MemberArn)
			if memberType == "" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewResource(
				accountID+","+folderID+","+memberType+","+*member.MemberId,
				*member.MemberId,
				"aws_quicksight_folder_membership",
				"aws",
				map[string]string{
					"aws_account_id": accountID,
					"folder_id":      folderID,
				},
				quicksightAllowEmptyValues,
				map[string]interface{}{},
			))
		}
		if page.NextToken == nil {
			return nil
		}
		input.NextToken = page.NextToken
	}
}

func folderMemberType(memberArn string) string {
	parts := strings.Split(memberArn, ":")
	if len(parts) != 6 {
		return ""
	}
	resource := strings.SplitN(parts[5], "/", 2)
	if len(resource) != 2 {
		return ""
	}
	switch resource[0] {
	case "analysis", "dashboard", "dataset":
		return strings.ToUpper(resource[0])
	}
	return ""
}

// Provider ID format: AWS account ID, template ID.
func (g *QuickSightGenerator) addTemplates(client *quicksight.Client, accountID string) error {
	input := &quicksight.ListTemplatesInput{
		AwsAccountId: aws.String(accountID),
	}
	for {
		page, err := client.ListTemplates(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, template := range page.TemplateSummaryList {
			templateID := *template.TemplateId
			g.Resources = append(g.Resources, terraformutils.NewResource(
				accountID+","+templateID,
				*template.Name,
				"aws_quicksight_template",
				"aws",
				map[string]string{"aws_account_id": accountID},
				quicksightAllowEmptyValues,
				map[string]interface{}{},
			))
			if err := g.addTemplateAliases(client, accountID, templateID); err != nil {
				return err
			}
		}
		if page.NextToken == nil {
			return nil
		}
		input.NextToken = page.NextToken
	}
}

// Provider ID format: AWS account ID, template ID, alias name.
func (g *QuickSightGenerator) addTemplateAliases(client *quicksight.Client, accountID, templateID string) error {
	input := &quicksight.ListTemplateAliasesInput{
		AwsAccountId: aws.String(accountID),
		TemplateId:   aws.String(templateID),
	}
	for {
		page, err := client.ListTemplateAliases(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, alias := range page.TemplateAliasList {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				accountID+","+templateID+","+*alias.AliasName,
				*alias.AliasName,
				"aws_quicksight_template_alias",
				"aws",
				map[string]string{
					"aws_account_id": accountID,
					"template_id":    templateID,
				},
				quicksightAllowEmptyValues,
				map[string]interface{}{},
			))
		}
		if page.NextToken == nil {
			return nil
		}
		input.NextToken = page.NextToken
	}
}

// Provider ID format: AWS account ID, theme ID.
func (g *QuickSightGenerator) addThemes(client *quicksight.Client, accountID string) error {
	input := &quicksight.ListThemesInput{
		AwsAccountId: aws.String(accountID),
	}
	for {
		page, err := client.ListThemes(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, theme := range page.ThemeSummaryList {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				accountID+","+*theme.ThemeId,
				*theme.Name,
				"aws_quicksight_theme",
				"aws",
				map[string]string{"aws_account_id": accountID},
				quicksightAllowEmptyValues,
				map[string]interface{}{},
			))
		}
		if page.NextToken == nil {
			return nil
		}
		input.NextToken = page.NextToken
	}
}

// Provider ID format: AWS account ID, VPC connection ID.
func (g *QuickSightGenerator) addVPCConnections(client *quicksight.Client, accountID string) error {
	input := &quicksight.ListVPCConnectionsInput{
		AwsAccountId: aws.String(accountID),
	}
	for {
		page, err := client.ListVPCConnections(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, connection := range page.VPCConnectionSummaries {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				accountID+","+*connection.VPCConnectionId,
				*connection.Name,
				"aws_quicksight_vpc_connection",
				"aws",
				map[string]string{"aws_account_id": accountID},
				quicksightAllowEmptyValues,
				map[string]interface{}{},
			))
		}
		if page.NextToken == nil {
			return nil
		}
		input.NextToken = page.NextToken
	}
}

// Provider ID format: AWS account ID, custom permissions profile name.
func (g *QuickSightGenerator) addCustomPermissions(client *quicksight.Client, accountID string) error {
	input := &quicksight.ListCustomPermissionsInput{
		AwsAccountId: aws.String(accountID),
	}
	for {
		page, err := client.ListCustomPermissions(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, profile := range page.CustomPermissionsList {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				accountID+","+*profile.CustomPermissionsName,
				*profile.CustomPermissionsName,
				"aws_quicksight_custom_permissions",
				"aws",
				map[string]string{"aws_account_id": accountID},
				quicksightAllowEmptyValues,
				map[string]interface{}{},
			))
		}
		if page.NextToken == nil {
			return nil
		}
		input.NextToken = page.NextToken
	}
}

func (g *QuickSightGenerator) addNamespaceResources(client *quicksight.Client, accountID, namespace string) error {
	if err := g.addGroups(client, accountID, namespace); err != nil {
		return err
	}
	if err := g.addUsers(client, accountID, namespace); err != nil {
		return err
	}
	if err := g.addIAMPolicyAssignments(client, accountID, namespace); err != nil {
		return err
	}
	if err := g.addRoleCustomPermissions(client, accountID, namespace); err != nil {
		return err
	}
	return g.addRoleMemberships(client, accountID, namespace)
}

// Provider ID format: AWS account ID, namespace, group name (slash
// separated).
func (g *QuickSightGenerator) addGroups(client *quicksight.Client, accountID, namespace string) error {
	input := &quicksight.ListGroupsInput{
		AwsAccountId: aws.String(accountID),
		Namespace:    aws.String(namespace),
	}
	for {
		page, err := client.ListGroups(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, group := range page.GroupList {
			groupName := *group.GroupName
			g.Resources = append(g.Resources, terraformutils.NewResource(
				accountID+"/"+namespace+"/"+groupName,
				groupName,
				"aws_quicksight_group",
				"aws",
				map[string]string{
					"aws_account_id": accountID,
					"namespace":      namespace,
				},
				quicksightAllowEmptyValues,
				map[string]interface{}{},
			))
			if err := g.addGroupMemberships(client, accountID, namespace, groupName); err != nil {
				return err
			}
		}
		if page.NextToken == nil {
			return nil
		}
		input.NextToken = page.NextToken
	}
}

// Provider ID format: AWS account ID, namespace, group name, member name
// (slash separated).
func (g *QuickSightGenerator) addGroupMemberships(client *quicksight.Client, accountID, namespace, groupName string) error {
	input := &quicksight.ListGroupMembershipsInput{
		AwsAccountId: aws.String(accountID),
		Namespace:    aws.String(namespace),
		GroupName:    aws.String(groupName),
	}
	for {
		page, err := client.ListGroupMemberships(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, member := range page.GroupMemberList {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				accountID+"/"+namespace+"/"+groupName+"/"+*member.MemberName,
				*member.MemberName,
				"aws_quicksight_group_membership",
				"aws",
				map[string]string{
					"aws_account_id": accountID,
					"namespace":      namespace,
					"group_name":     groupName,
				},
				quicksightAllowEmptyValues,
				map[string]interface{}{},
			))
		}
		if page.NextToken == nil {
			return nil
		}
		input.NextToken = page.NextToken
	}
}

// Provider ID format: AWS account ID, namespace, user name (slash
// separated).
func (g *QuickSightGenerator) addUsers(client *quicksight.Client, accountID, namespace string) error {
	input := &quicksight.ListUsersInput{
		AwsAccountId: aws.String(accountID),
		Namespace:    aws.String(namespace),
	}
	for {
		page, err := client.ListUsers(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, user := range page.UserList {
			userName := *user.UserName
			g.Resources = append(g.Resources, terraformutils.NewResource(
				accountID+"/"+namespace+"/"+userName,
				userName,
				"aws_quicksight_user",
				"aws",
				map[string]string{
					"aws_account_id": accountID,
					"namespace":      namespace,
				},
				quicksightAllowEmptyValues,
				map[string]interface{}{},
			))
			if user.CustomPermissionsName == nil || *user.CustomPermissionsName == "" {
				continue
			}
			// Provider ID format: AWS account ID, namespace, user name.
			g.Resources = append(g.Resources, terraformutils.NewResource(
				accountID+","+namespace+","+userName,
				userName,
				"aws_quicksight_user_custom_permission",
				"aws",
				map[string]string{
					"aws_account_id": accountID,
					"namespace":      namespace,
					"user_name":      userName,
				},
				quicksightAllowEmptyValues,
				map[string]interface{}{},
			))
		}
		if page.NextToken == nil {
			return nil
		}
		input.NextToken = page.NextToken
	}
}

// Provider ID format: AWS account ID, namespace, assignment name.
func (g *QuickSightGenerator) addIAMPolicyAssignments(client *quicksight.Client, accountID, namespace string) error {
	input := &quicksight.ListIAMPolicyAssignmentsInput{
		AwsAccountId: aws.String(accountID),
		Namespace:    aws.String(namespace),
	}
	for {
		page, err := client.ListIAMPolicyAssignments(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, assignment := range page.IAMPolicyAssignments {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				accountID+","+namespace+","+*assignment.AssignmentName,
				*assignment.AssignmentName,
				"aws_quicksight_iam_policy_assignment",
				"aws",
				map[string]string{
					"aws_account_id": accountID,
					"namespace":      namespace,
				},
				quicksightAllowEmptyValues,
				map[string]interface{}{},
			))
		}
		if page.NextToken == nil {
			return nil
		}
		input.NextToken = page.NextToken
	}
}

var quicksightRoles = []quicksighttypes.Role{
	quicksighttypes.RoleAdmin,
	quicksighttypes.RoleAuthor,
	quicksighttypes.RoleReader,
	quicksighttypes.RoleAdminPro,
	quicksighttypes.RoleAuthorPro,
	quicksighttypes.RoleReaderPro,
}

// Role custom permissions are not listable; the role set is fixed, so each
// role is described in turn. A role with no profile attached reports
// ResourceNotFoundException and is skipped.
//
// Provider ID format: AWS account ID, namespace, role.
func (g *QuickSightGenerator) addRoleCustomPermissions(client *quicksight.Client, accountID, namespace string) error {
	for _, role := range quicksightRoles {
		out, err := client.DescribeRoleCustomPermission(context.TODO(), &quicksight.DescribeRoleCustomPermissionInput{
			AwsAccountId: aws.String(accountID),
			Namespace:    aws.String(namespace),
			Role:         role,
		})
		if err != nil {
			var notFound *quicksighttypes.ResourceNotFoundException
			if errors.As(err, &notFound) {
				continue
			}
			return err
		}
		if out.CustomPermissionsName == nil || *out.CustomPermissionsName == "" {
			continue
		}
		g.Resources = append(g.Resources, terraformutils.NewResource(
			accountID+","+namespace+","+string(role),
			string(role),
			"aws_quicksight_role_custom_permission",
			"aws",
			map[string]string{
				"aws_account_id": accountID,
				"namespace":      namespace,
			},
			quicksightAllowEmptyValues,
			map[string]interface{}{},
		))
	}
	return nil
}

// The role membership APIs are disabled for identities managed by
// QuickSight itself (the common account type), where ListRoleMemberships
// fails outright; following the S3 Express precedent in the S3 generator,
// a failure here is logged and skipped rather than failing the whole
// QuickSight import.
//
// Provider ID format: AWS account ID, namespace, role, member name.
func (g *QuickSightGenerator) addRoleMemberships(client *quicksight.Client, accountID, namespace string) error {
	for _, role := range quicksightRoles {
		input := &quicksight.ListRoleMembershipsInput{
			AwsAccountId: aws.String(accountID),
			Namespace:    aws.String(namespace),
			Role:         role,
		}
		for {
			page, err := client.ListRoleMemberships(context.TODO(), input)
			if err != nil {
				log.Printf("quicksight: skipping role memberships for role %s in namespace %s: %v", role, namespace, err)
				break
			}
			for _, member := range page.MembersList {
				g.Resources = append(g.Resources, terraformutils.NewResource(
					accountID+","+namespace+","+string(role)+","+member,
					member,
					"aws_quicksight_role_membership",
					"aws",
					map[string]string{
						"aws_account_id": accountID,
						"namespace":      namespace,
						"role":           string(role),
					},
					quicksightAllowEmptyValues,
					map[string]interface{}{},
				))
			}
			if page.NextToken == nil {
				break
			}
			input.NextToken = page.NextToken
		}
	}
	return nil
}
