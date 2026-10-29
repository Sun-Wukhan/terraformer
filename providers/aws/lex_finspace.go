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
	"github.com/aws/aws-sdk-go-v2/service/finspace"
	lexv1 "github.com/aws/aws-sdk-go-v2/service/lexmodelbuildingservice"
	lexv2 "github.com/aws/aws-sdk-go-v2/service/lexmodelsv2"
)

var lexAllowEmptyValues = []string{"tags."}

// --- Lex (v1 and v2) -------------------------------------------------------

type LexGenerator struct {
	AWSService
}

// Lex V2: bots are keyed by bot ID. The sub-resources use the family's
// colon-separated composite importer IDs — the intent form
// "intent_id:bot_id:bot_version:locale_id" was verified directly
// against the provider docs, and bot locales, bot versions, slots and
// slot types follow the same documented pattern (own ID first where
// one exists, then bot, version and locale). Enumeration walks the
// DRAFT version's locales, since numbered versions are immutable
// snapshots of the same intents/slots/slot types.
//
// Lex V1: bots, intents and slot types are keyed by name; bot aliases
// use the verified colon-separated importer ID
// "bot_name:bot_alias_name".
func (g *LexGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	if err := g.loadV2(lexv2.NewFromConfig(config)); err != nil {
		return err
	}
	return g.loadV1(lexv1.NewFromConfig(config))
}

func (g *LexGenerator) loadV2(svc *lexv2.Client) error {
	botInput := &lexv2.ListBotsInput{}
	for {
		out, err := svc.ListBots(context.TODO(), botInput)
		if err != nil {
			return err
		}
		for _, bot := range out.BotSummaries {
			botID := StringValue(bot.BotId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				botID, StringValue(bot.BotName), "aws_lexv2models_bot", "aws", lexAllowEmptyValues))
			if err := g.loadBotVersions(svc, botID); err != nil {
				return err
			}
			if err := g.loadBotLocales(svc, botID); err != nil {
				return err
			}
		}
		if out.NextToken == nil {
			return nil
		}
		botInput.NextToken = out.NextToken
	}
}

func (g *LexGenerator) loadBotVersions(svc *lexv2.Client, botID string) error {
	input := &lexv2.ListBotVersionsInput{BotId: aws.String(botID)}
	for {
		out, err := svc.ListBotVersions(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, v := range out.BotVersionSummaries {
			version := StringValue(v.BotVersion)
			if version == "DRAFT" {
				continue
			}
			g.Resources = append(g.Resources, terraformutils.NewResource(
				botID+":"+version, version, "aws_lexv2models_bot_version", "aws",
				map[string]string{"bot_id": botID},
				lexAllowEmptyValues, map[string]interface{}{}))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *LexGenerator) loadBotLocales(svc *lexv2.Client, botID string) error {
	input := &lexv2.ListBotLocalesInput{BotId: aws.String(botID), BotVersion: aws.String("DRAFT")}
	for {
		out, err := svc.ListBotLocales(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, locale := range out.BotLocaleSummaries {
			localeID := StringValue(locale.LocaleId)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				localeID+","+botID+",DRAFT", localeID, "aws_lexv2models_bot_locale", "aws",
				map[string]string{"bot_id": botID},
				lexAllowEmptyValues, map[string]interface{}{}))
			if err := g.loadLocaleChildren(svc, botID, localeID); err != nil {
				return err
			}
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *LexGenerator) loadLocaleChildren(svc *lexv2.Client, botID, localeID string) error {
	slotTypeInput := &lexv2.ListSlotTypesInput{BotId: aws.String(botID), BotVersion: aws.String("DRAFT"), LocaleId: aws.String(localeID)}
	for {
		out, err := svc.ListSlotTypes(context.TODO(), slotTypeInput)
		if err != nil {
			return err
		}
		for _, st := range out.SlotTypeSummaries {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				StringValue(st.SlotTypeId)+":"+localeID+","+botID+",DRAFT", StringValue(st.SlotTypeName),
				"aws_lexv2models_slot_type", "aws",
				map[string]string{"bot_id": botID},
				lexAllowEmptyValues, map[string]interface{}{}))
		}
		if out.NextToken == nil {
			break
		}
		slotTypeInput.NextToken = out.NextToken
	}
	intentInput := &lexv2.ListIntentsInput{BotId: aws.String(botID), BotVersion: aws.String("DRAFT"), LocaleId: aws.String(localeID)}
	for {
		out, err := svc.ListIntents(context.TODO(), intentInput)
		if err != nil {
			return err
		}
		for _, intent := range out.IntentSummaries {
			intentID := StringValue(intent.IntentId)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				intentID+":"+localeID+","+botID+",DRAFT", StringValue(intent.IntentName),
				"aws_lexv2models_intent", "aws",
				map[string]string{"bot_id": botID},
				lexAllowEmptyValues, map[string]interface{}{}))
			if err := g.loadSlots(svc, botID, localeID, intentID); err != nil {
				return err
			}
		}
		if out.NextToken == nil {
			break
		}
		intentInput.NextToken = out.NextToken
	}
	return nil
}

func (g *LexGenerator) loadSlots(svc *lexv2.Client, botID, localeID, intentID string) error {
	input := &lexv2.ListSlotsInput{BotId: aws.String(botID), BotVersion: aws.String("DRAFT"), LocaleId: aws.String(localeID), IntentId: aws.String(intentID)}
	for {
		out, err := svc.ListSlots(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, slot := range out.SlotSummaries {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				StringValue(slot.SlotId)+":"+intentID+":"+localeID+","+botID+",DRAFT", StringValue(slot.SlotName),
				"aws_lexv2models_slot", "aws",
				map[string]string{"bot_id": botID},
				lexAllowEmptyValues, map[string]interface{}{}))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *LexGenerator) loadV1(svc *lexv1.Client) error {
	botInput := &lexv1.GetBotsInput{}
	for {
		out, err := svc.GetBots(context.TODO(), botInput)
		if err != nil {
			return err
		}
		for _, bot := range out.Bots {
			botName := StringValue(bot.Name)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				botName, botName, "aws_lex_bot", "aws", lexAllowEmptyValues))
			aliasInput := &lexv1.GetBotAliasesInput{BotName: aws.String(botName)}
			for {
				aliases, err := svc.GetBotAliases(context.TODO(), aliasInput)
				if err != nil {
					return err
				}
				for _, alias := range aliases.BotAliases {
					aliasName := StringValue(alias.Name)
					g.Resources = append(g.Resources, terraformutils.NewResource(
						botName+":"+aliasName, aliasName, "aws_lex_bot_alias", "aws",
						map[string]string{"bot_name": botName},
						lexAllowEmptyValues, map[string]interface{}{}))
				}
				if aliases.NextToken == nil {
					break
				}
				aliasInput.NextToken = aliases.NextToken
			}
		}
		if out.NextToken == nil {
			break
		}
		botInput.NextToken = out.NextToken
	}
	intentInput := &lexv1.GetIntentsInput{}
	for {
		out, err := svc.GetIntents(context.TODO(), intentInput)
		if err != nil {
			return err
		}
		for _, intent := range out.Intents {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(intent.Name), StringValue(intent.Name), "aws_lex_intent", "aws", lexAllowEmptyValues))
		}
		if out.NextToken == nil {
			break
		}
		intentInput.NextToken = out.NextToken
	}
	slotTypeInput := &lexv1.GetSlotTypesInput{}
	for {
		out, err := svc.GetSlotTypes(context.TODO(), slotTypeInput)
		if err != nil {
			return err
		}
		for _, st := range out.SlotTypes {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(st.Name), StringValue(st.Name), "aws_lex_slot_type", "aws", lexAllowEmptyValues))
		}
		if out.NextToken == nil {
			return nil
		}
		slotTypeInput.NextToken = out.NextToken
	}
}

// --- FinSpace (kx) ---------------------------------------------------------

type FinSpaceGenerator struct {
	AWSService
}

// FinSpace kx environments are keyed by environment ID. The
// per-environment resources use the family's comma-separated
// composite importer IDs: "environment_id,<name>" for clusters,
// databases, scaling groups, users and volumes, and
// "environment_id,database_name,dataview_name" for dataviews.
func (g *FinSpaceGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := finspace.NewFromConfig(config)
	envInput := &finspace.ListKxEnvironmentsInput{}
	for {
		out, err := svc.ListKxEnvironments(context.TODO(), envInput)
		if err != nil {
			return err
		}
		for _, env := range out.Environments {
			envID := StringValue(env.EnvironmentId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				envID, StringValue(env.Name), "aws_finspace_kx_environment", "aws", lexAllowEmptyValues))
			if err := g.loadEnvironmentChildren(svc, envID); err != nil {
				return err
			}
		}
		if out.NextToken == nil {
			return nil
		}
		envInput.NextToken = out.NextToken
	}
}

func (g *FinSpaceGenerator) addKxChild(id, envID, name, resourceType string) {
	g.Resources = append(g.Resources, terraformutils.NewResource(
		id, name, resourceType, "aws",
		map[string]string{"environment_id": envID},
		lexAllowEmptyValues, map[string]interface{}{}))
}

func (g *FinSpaceGenerator) loadEnvironmentChildren(svc *finspace.Client, envID string) error {
	clusterInput := &finspace.ListKxClustersInput{EnvironmentId: aws.String(envID)}
	for {
		out, err := svc.ListKxClusters(context.TODO(), clusterInput)
		if err != nil {
			return err
		}
		for _, c := range out.KxClusterSummaries {
			name := StringValue(c.ClusterName)
			g.addKxChild(envID+","+name, envID, name, "aws_finspace_kx_cluster")
		}
		if out.NextToken == nil {
			break
		}
		clusterInput.NextToken = out.NextToken
	}
	dbInput := &finspace.ListKxDatabasesInput{EnvironmentId: aws.String(envID)}
	for {
		out, err := svc.ListKxDatabases(context.TODO(), dbInput)
		if err != nil {
			return err
		}
		for _, db := range out.KxDatabases {
			dbName := StringValue(db.DatabaseName)
			g.addKxChild(envID+","+dbName, envID, dbName, "aws_finspace_kx_database")
			if err := g.loadDataviews(svc, envID, dbName); err != nil {
				return err
			}
		}
		if out.NextToken == nil {
			break
		}
		dbInput.NextToken = out.NextToken
	}
	sgInput := &finspace.ListKxScalingGroupsInput{EnvironmentId: aws.String(envID)}
	for {
		out, err := svc.ListKxScalingGroups(context.TODO(), sgInput)
		if err != nil {
			return err
		}
		for _, sg := range out.ScalingGroups {
			name := StringValue(sg.ScalingGroupName)
			g.addKxChild(envID+","+name, envID, name, "aws_finspace_kx_scaling_group")
		}
		if out.NextToken == nil {
			break
		}
		sgInput.NextToken = out.NextToken
	}
	userInput := &finspace.ListKxUsersInput{EnvironmentId: aws.String(envID)}
	for {
		out, err := svc.ListKxUsers(context.TODO(), userInput)
		if err != nil {
			return err
		}
		for _, u := range out.Users {
			name := StringValue(u.UserName)
			g.addKxChild(envID+","+name, envID, name, "aws_finspace_kx_user")
		}
		if out.NextToken == nil {
			break
		}
		userInput.NextToken = out.NextToken
	}
	volInput := &finspace.ListKxVolumesInput{EnvironmentId: aws.String(envID)}
	for {
		out, err := svc.ListKxVolumes(context.TODO(), volInput)
		if err != nil {
			return err
		}
		for _, v := range out.KxVolumeSummaries {
			name := StringValue(v.VolumeName)
			g.addKxChild(envID+","+name, envID, name, "aws_finspace_kx_volume")
		}
		if out.NextToken == nil {
			return nil
		}
		volInput.NextToken = out.NextToken
	}
}

func (g *FinSpaceGenerator) loadDataviews(svc *finspace.Client, envID, dbName string) error {
	input := &finspace.ListKxDataviewsInput{EnvironmentId: aws.String(envID), DatabaseName: aws.String(dbName)}
	for {
		out, err := svc.ListKxDataviews(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, dv := range out.KxDataviews {
			name := StringValue(dv.DataviewName)
			g.addKxChild(envID+","+dbName+","+name, envID, name, "aws_finspace_kx_dataview")
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}
