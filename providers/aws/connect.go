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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connect"
	"github.com/aws/aws-sdk-go-v2/service/connect/types"
)

var connectAllowEmptyValues = []string{"tags."}

type ConnectGenerator struct {
	AWSService
}

// InitResources enumerates the Amazon Connect objects in the configured
// region: instances with their contact flows, contact flow modules,
// hours of operation, queues, quick connects, routing profiles, security
// profiles, users, user hierarchy groups and hierarchy structure,
// vocabularies, instance storage configs for every resource type, Lex
// bot (V1) and Lambda function associations, and phone number contact
// flow associations; plus the account's phone numbers, which are listed
// regionally rather than per instance. Most per-instance resources are
// keyed by the instance ID and the resource ID separated by a colon,
// matching the Terraform AWS provider's ID format.
func (g *ConnectGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	client := connect.NewFromConfig(config)

	if err := g.addPhoneNumbers(client); err != nil {
		return err
	}

	paginator := connect.NewListInstancesPaginator(client, &connect.ListInstancesInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, instance := range page.InstanceSummaryList {
			instanceID := *instance.Id
			name := instanceID
			if instance.InstanceAlias != nil && *instance.InstanceAlias != "" {
				name = *instance.InstanceAlias
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				instanceID,
				name,
				"aws_connect_instance",
				"aws",
				connectAllowEmptyValues,
			))
			for _, add := range []func(*connect.Client, string) error{
				g.addContactFlows,
				g.addContactFlowModules,
				g.addHoursOfOperations,
				g.addQueues,
				g.addQuickConnects,
				g.addRoutingProfiles,
				g.addSecurityProfiles,
				g.addUsers,
				g.addUserHierarchyGroups,
				g.addVocabularies,
				g.addUserHierarchyStructure,
				g.addInstanceStorageConfigs,
				g.addBotAssociations,
				g.addLambdaFunctionAssociations,
				g.addPhoneNumberContactFlowAssociations,
			} {
				if err := add(client, instanceID); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Phone numbers are keyed by phone number ID alone and enumerated once
// for the region; the resource is named after the number itself.
func (g *ConnectGenerator) addPhoneNumbers(client *connect.Client) error {
	paginator := connect.NewListPhoneNumbersV2Paginator(client, &connect.ListPhoneNumbersV2Input{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, number := range page.ListPhoneNumbersSummaryList {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*number.PhoneNumberId,
				*number.PhoneNumber,
				"aws_connect_phone_number",
				"aws",
				connectAllowEmptyValues,
			))
		}
	}
	return nil
}

// Provider ID format: instance ID, contact flow ID (colon separated).
func (g *ConnectGenerator) addContactFlows(client *connect.Client, instanceID string) error {
	paginator := connect.NewListContactFlowsPaginator(client, &connect.ListContactFlowsInput{
		InstanceId: aws.String(instanceID),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, flow := range page.ContactFlowSummaryList {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				instanceID+":"+*flow.Id,
				*flow.Name,
				"aws_connect_contact_flow",
				"aws",
				map[string]string{"instance_id": instanceID},
				connectAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

// Provider ID format: instance ID, contact flow module ID (colon separated).
func (g *ConnectGenerator) addContactFlowModules(client *connect.Client, instanceID string) error {
	paginator := connect.NewListContactFlowModulesPaginator(client, &connect.ListContactFlowModulesInput{
		InstanceId: aws.String(instanceID),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, module := range page.ContactFlowModulesSummaryList {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				instanceID+":"+*module.Id,
				*module.Name,
				"aws_connect_contact_flow_module",
				"aws",
				map[string]string{"instance_id": instanceID},
				connectAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

// Provider ID format: instance ID, hours of operation ID (colon separated).
func (g *ConnectGenerator) addHoursOfOperations(client *connect.Client, instanceID string) error {
	paginator := connect.NewListHoursOfOperationsPaginator(client, &connect.ListHoursOfOperationsInput{
		InstanceId: aws.String(instanceID),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, hours := range page.HoursOfOperationSummaryList {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				instanceID+":"+*hours.Id,
				*hours.Name,
				"aws_connect_hours_of_operation",
				"aws",
				map[string]string{"instance_id": instanceID},
				connectAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

// Provider ID format: instance ID, queue ID (colon separated).
func (g *ConnectGenerator) addQueues(client *connect.Client, instanceID string) error {
	paginator := connect.NewListQueuesPaginator(client, &connect.ListQueuesInput{
		InstanceId: aws.String(instanceID),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, queue := range page.QueueSummaryList {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				instanceID+":"+*queue.Id,
				*queue.Name,
				"aws_connect_queue",
				"aws",
				map[string]string{"instance_id": instanceID},
				connectAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

// Provider ID format: instance ID, quick connect ID (colon separated).
func (g *ConnectGenerator) addQuickConnects(client *connect.Client, instanceID string) error {
	paginator := connect.NewListQuickConnectsPaginator(client, &connect.ListQuickConnectsInput{
		InstanceId: aws.String(instanceID),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, quickConnect := range page.QuickConnectSummaryList {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				instanceID+":"+*quickConnect.Id,
				*quickConnect.Name,
				"aws_connect_quick_connect",
				"aws",
				map[string]string{"instance_id": instanceID},
				connectAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

// Provider ID format: instance ID, routing profile ID (colon separated).
func (g *ConnectGenerator) addRoutingProfiles(client *connect.Client, instanceID string) error {
	paginator := connect.NewListRoutingProfilesPaginator(client, &connect.ListRoutingProfilesInput{
		InstanceId: aws.String(instanceID),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, profile := range page.RoutingProfileSummaryList {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				instanceID+":"+*profile.Id,
				*profile.Name,
				"aws_connect_routing_profile",
				"aws",
				map[string]string{"instance_id": instanceID},
				connectAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

// Provider ID format: instance ID, security profile ID (colon separated).
func (g *ConnectGenerator) addSecurityProfiles(client *connect.Client, instanceID string) error {
	paginator := connect.NewListSecurityProfilesPaginator(client, &connect.ListSecurityProfilesInput{
		InstanceId: aws.String(instanceID),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, profile := range page.SecurityProfileSummaryList {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				instanceID+":"+*profile.Id,
				*profile.Name,
				"aws_connect_security_profile",
				"aws",
				map[string]string{"instance_id": instanceID},
				connectAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

// Provider ID format: instance ID, user ID (colon separated).
func (g *ConnectGenerator) addUsers(client *connect.Client, instanceID string) error {
	paginator := connect.NewListUsersPaginator(client, &connect.ListUsersInput{
		InstanceId: aws.String(instanceID),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, user := range page.UserSummaryList {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				instanceID+":"+*user.Id,
				*user.Username,
				"aws_connect_user",
				"aws",
				map[string]string{"instance_id": instanceID},
				connectAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

// Provider ID format: instance ID, hierarchy group ID (colon separated).
func (g *ConnectGenerator) addUserHierarchyGroups(client *connect.Client, instanceID string) error {
	paginator := connect.NewListUserHierarchyGroupsPaginator(client, &connect.ListUserHierarchyGroupsInput{
		InstanceId: aws.String(instanceID),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, group := range page.UserHierarchyGroupSummaryList {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				instanceID+":"+*group.Id,
				*group.Name,
				"aws_connect_user_hierarchy_group",
				"aws",
				map[string]string{"instance_id": instanceID},
				connectAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

// Provider ID format: instance ID, vocabulary ID (colon separated).
func (g *ConnectGenerator) addVocabularies(client *connect.Client, instanceID string) error {
	paginator := connect.NewSearchVocabulariesPaginator(client, &connect.SearchVocabulariesInput{
		InstanceId: aws.String(instanceID),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, vocabulary := range page.VocabularySummaryList {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				instanceID+":"+*vocabulary.Id,
				*vocabulary.Name,
				"aws_connect_vocabulary",
				"aws",
				map[string]string{"instance_id": instanceID},
				connectAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

// The hierarchy structure is a per-instance singleton keyed by the
// instance ID; it only exists once at least one level has been named
// (the provider requires a non-empty hierarchy_structure block).
func (g *ConnectGenerator) addUserHierarchyStructure(client *connect.Client, instanceID string) error {
	out, err := client.DescribeUserHierarchyStructure(context.TODO(), &connect.DescribeUserHierarchyStructureInput{
		InstanceId: aws.String(instanceID),
	})
	if err != nil {
		return err
	}
	if out.HierarchyStructure == nil {
		return nil
	}
	named := false
	for _, level := range []*types.HierarchyLevel{
		out.HierarchyStructure.LevelOne,
		out.HierarchyStructure.LevelTwo,
		out.HierarchyStructure.LevelThree,
		out.HierarchyStructure.LevelFour,
		out.HierarchyStructure.LevelFive,
	} {
		if level != nil && level.Name != nil && *level.Name != "" {
			named = true
		}
	}
	if !named {
		return nil
	}
	g.Resources = append(g.Resources, terraformutils.NewResource(
		instanceID,
		instanceID,
		"aws_connect_user_hierarchy_structure",
		"aws",
		map[string]string{"instance_id": instanceID},
		connectAllowEmptyValues,
		map[string]interface{}{},
	))
	return nil
}

// Provider ID format: instance ID, association ID, resource type (colon
// separated). Every storage resource type the provider supports is
// probed in turn.
func (g *ConnectGenerator) addInstanceStorageConfigs(client *connect.Client, instanceID string) error {
	for _, resourceType := range []types.InstanceStorageResourceType{
		types.InstanceStorageResourceTypeAgentEvents,
		types.InstanceStorageResourceTypeAttachments,
		types.InstanceStorageResourceTypeCallRecordings,
		types.InstanceStorageResourceTypeChatTranscripts,
		types.InstanceStorageResourceTypeContactEvaluations,
		types.InstanceStorageResourceTypeContactTraceRecords,
		types.InstanceStorageResourceTypeEmailMessages,
		types.InstanceStorageResourceTypeMediaStreams,
		types.InstanceStorageResourceTypeRealTimeContactAnalysisSegments,
		types.InstanceStorageResourceTypeRealTimeContactAnalysisChatSegments,
		types.InstanceStorageResourceTypeRealTimeContactAnalysisVoiceSegments,
		types.InstanceStorageResourceTypeScheduledReports,
		types.InstanceStorageResourceTypeScreenRecordings,
	} {
		paginator := connect.NewListInstanceStorageConfigsPaginator(client, &connect.ListInstanceStorageConfigsInput{
			InstanceId:   aws.String(instanceID),
			ResourceType: resourceType,
		})
		for paginator.HasMorePages() {
			page, err := paginator.NextPage(context.TODO())
			if err != nil {
				return err
			}
			for _, storageConfig := range page.StorageConfigs {
				g.Resources = append(g.Resources, terraformutils.NewResource(
					instanceID+":"+*storageConfig.AssociationId+":"+string(resourceType),
					*storageConfig.AssociationId,
					"aws_connect_instance_storage_config",
					"aws",
					map[string]string{"instance_id": instanceID},
					connectAllowEmptyValues,
					map[string]interface{}{},
				))
			}
		}
	}
	return nil
}

// Provider ID format: instance ID, bot name, Lex region (colon
// separated). The provider resource manages Lex V1 bots only.
func (g *ConnectGenerator) addBotAssociations(client *connect.Client, instanceID string) error {
	paginator := connect.NewListBotsPaginator(client, &connect.ListBotsInput{
		InstanceId: aws.String(instanceID),
		LexVersion: types.LexVersionV1,
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, botConfig := range page.LexBots {
			if botConfig.LexBot == nil {
				continue
			}
			bot := botConfig.LexBot
			g.Resources = append(g.Resources, terraformutils.NewResource(
				instanceID+":"+*bot.Name+":"+*bot.LexRegion,
				*bot.Name,
				"aws_connect_bot_association",
				"aws",
				map[string]string{"instance_id": instanceID},
				connectAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

// Provider ID format: instance ID, Lambda function ARN (comma separated).
func (g *ConnectGenerator) addLambdaFunctionAssociations(client *connect.Client, instanceID string) error {
	paginator := connect.NewListLambdaFunctionsPaginator(client, &connect.ListLambdaFunctionsInput{
		InstanceId: aws.String(instanceID),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, functionArn := range page.LambdaFunctions {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				instanceID+","+functionArn,
				arnResourceName(functionArn),
				"aws_connect_lambda_function_association",
				"aws",
				map[string]string{"instance_id": instanceID},
				connectAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

// Provider ID format: phone number ID, instance ID, contact flow ID
// (comma separated). Flow associations are reported by ARN, so the
// instance's contact flows and phone numbers are indexed by ARN first.
func (g *ConnectGenerator) addPhoneNumberContactFlowAssociations(client *connect.Client, instanceID string) error {
	flowIDsByArn := map[string]string{}
	flowPaginator := connect.NewListContactFlowsPaginator(client, &connect.ListContactFlowsInput{
		InstanceId: aws.String(instanceID),
	})
	for flowPaginator.HasMorePages() {
		page, err := flowPaginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, flow := range page.ContactFlowSummaryList {
			flowIDsByArn[*flow.Arn] = *flow.Id
		}
	}

	phoneIDsByArn := map[string]string{}
	phonePaginator := connect.NewListPhoneNumbersPaginator(client, &connect.ListPhoneNumbersInput{
		InstanceId: aws.String(instanceID),
	})
	for phonePaginator.HasMorePages() {
		page, err := phonePaginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, number := range page.PhoneNumberSummaryList {
			phoneIDsByArn[*number.Arn] = *number.Id
		}
	}

	paginator := connect.NewListFlowAssociationsPaginator(client, &connect.ListFlowAssociationsInput{
		InstanceId:   aws.String(instanceID),
		ResourceType: types.ListFlowAssociationResourceTypeVoicePhoneNumber,
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, association := range page.FlowAssociationSummaryList {
			phoneNumberID, okPhone := phoneIDsByArn[*association.ResourceId]
			flowID, okFlow := flowIDsByArn[*association.FlowId]
			if !okPhone || !okFlow {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewResource(
				phoneNumberID+","+instanceID+","+flowID,
				phoneNumberID,
				"aws_connect_phone_number_contact_flow_association",
				"aws",
				map[string]string{"instance_id": instanceID},
				connectAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}
