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
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
)

var bedrockagentAllowEmptyValues = []string{"tags."}

// Sub-resources of an agent (action groups, collaborators, knowledge base
// associations) exist per agent version; Terraformer enumerates the DRAFT
// version, which is the only one the provider resources manage directly.
const bedrockagentDraftVersion = "DRAFT"

type BedrockAgentGenerator struct {
	AWSService
}

func (g *BedrockAgentGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	client := bedrockagent.NewFromConfig(config)

	if err := g.addAgents(client); err != nil {
		return err
	}
	if err := g.addKnowledgeBases(client); err != nil {
		return err
	}
	if err := g.addFlows(client); err != nil {
		return err
	}
	return g.addPrompts(client)
}

func (g *BedrockAgentGenerator) addAgents(client *bedrockagent.Client) error {
	p := bedrockagent.NewListAgentsPaginator(client, &bedrockagent.ListAgentsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, agent := range page.AgentSummaries {
			agentID := *agent.AgentId
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				agentID,
				*agent.AgentName,
				"aws_bedrockagent_agent",
				"aws",
				bedrockagentAllowEmptyValues,
			))
			if err := g.addAgentAliases(client, agentID); err != nil {
				return err
			}
			if err := g.addAgentActionGroups(client, agentID); err != nil {
				return err
			}
			if err := g.addAgentCollaborators(client, agentID); err != nil {
				return err
			}
			if err := g.addAgentKnowledgeBaseAssociations(client, agentID); err != nil {
				return err
			}
		}
	}
	return nil
}

// Provider ID format: alias ID, agent ID.
func (g *BedrockAgentGenerator) addAgentAliases(client *bedrockagent.Client, agentID string) error {
	p := bedrockagent.NewListAgentAliasesPaginator(client, &bedrockagent.ListAgentAliasesInput{
		AgentId: aws.String(agentID),
	})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, alias := range page.AgentAliasSummaries {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				*alias.AgentAliasId+","+agentID,
				*alias.AgentAliasName,
				"aws_bedrockagent_agent_alias",
				"aws",
				map[string]string{"agent_id": agentID},
				bedrockagentAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

// Provider ID format: action group ID, agent ID, agent version.
func (g *BedrockAgentGenerator) addAgentActionGroups(client *bedrockagent.Client, agentID string) error {
	p := bedrockagent.NewListAgentActionGroupsPaginator(client, &bedrockagent.ListAgentActionGroupsInput{
		AgentId:      aws.String(agentID),
		AgentVersion: aws.String(bedrockagentDraftVersion),
	})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, group := range page.ActionGroupSummaries {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				*group.ActionGroupId+","+agentID+","+bedrockagentDraftVersion,
				*group.ActionGroupName,
				"aws_bedrockagent_agent_action_group",
				"aws",
				map[string]string{
					"agent_id":      agentID,
					"agent_version": bedrockagentDraftVersion,
				},
				bedrockagentAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

// Provider ID format: agent ID, agent version, collaborator ID.
func (g *BedrockAgentGenerator) addAgentCollaborators(client *bedrockagent.Client, agentID string) error {
	p := bedrockagent.NewListAgentCollaboratorsPaginator(client, &bedrockagent.ListAgentCollaboratorsInput{
		AgentId:      aws.String(agentID),
		AgentVersion: aws.String(bedrockagentDraftVersion),
	})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, collaborator := range page.AgentCollaboratorSummaries {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				agentID+","+bedrockagentDraftVersion+","+*collaborator.CollaboratorId,
				*collaborator.CollaboratorName,
				"aws_bedrockagent_agent_collaborator",
				"aws",
				map[string]string{
					"agent_id":      agentID,
					"agent_version": bedrockagentDraftVersion,
				},
				bedrockagentAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

// Provider ID format: agent ID, agent version, knowledge base ID.
func (g *BedrockAgentGenerator) addAgentKnowledgeBaseAssociations(client *bedrockagent.Client, agentID string) error {
	p := bedrockagent.NewListAgentKnowledgeBasesPaginator(client, &bedrockagent.ListAgentKnowledgeBasesInput{
		AgentId:      aws.String(agentID),
		AgentVersion: aws.String(bedrockagentDraftVersion),
	})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, kb := range page.AgentKnowledgeBaseSummaries {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				agentID+","+bedrockagentDraftVersion+","+*kb.KnowledgeBaseId,
				"kb_"+*kb.KnowledgeBaseId,
				"aws_bedrockagent_agent_knowledge_base_association",
				"aws",
				map[string]string{
					"agent_id":          agentID,
					"agent_version":     bedrockagentDraftVersion,
					"knowledge_base_id": *kb.KnowledgeBaseId,
				},
				bedrockagentAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

func (g *BedrockAgentGenerator) addKnowledgeBases(client *bedrockagent.Client) error {
	p := bedrockagent.NewListKnowledgeBasesPaginator(client, &bedrockagent.ListKnowledgeBasesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, kb := range page.KnowledgeBaseSummaries {
			kbID := *kb.KnowledgeBaseId
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				kbID,
				*kb.Name,
				"aws_bedrockagent_knowledge_base",
				"aws",
				bedrockagentAllowEmptyValues,
			))
			if err := g.addDataSources(client, kbID); err != nil {
				return err
			}
		}
	}
	return nil
}

// Provider ID format: data source ID, knowledge base ID.
func (g *BedrockAgentGenerator) addDataSources(client *bedrockagent.Client, kbID string) error {
	p := bedrockagent.NewListDataSourcesPaginator(client, &bedrockagent.ListDataSourcesInput{
		KnowledgeBaseId: aws.String(kbID),
	})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, ds := range page.DataSourceSummaries {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				*ds.DataSourceId+","+kbID,
				*ds.Name,
				"aws_bedrockagent_data_source",
				"aws",
				map[string]string{"knowledge_base_id": kbID},
				bedrockagentAllowEmptyValues,
				map[string]interface{}{},
			))
		}
	}
	return nil
}

func (g *BedrockAgentGenerator) addFlows(client *bedrockagent.Client) error {
	p := bedrockagent.NewListFlowsPaginator(client, &bedrockagent.ListFlowsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, flow := range page.FlowSummaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*flow.Id,
				*flow.Name,
				"aws_bedrockagent_flow",
				"aws",
				bedrockagentAllowEmptyValues,
			))
		}
	}
	return nil
}

func (g *BedrockAgentGenerator) addPrompts(client *bedrockagent.Client) error {
	p := bedrockagent.NewListPromptsPaginator(client, &bedrockagent.ListPromptsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, prompt := range page.PromptSummaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*prompt.Id,
				*prompt.Name,
				"aws_bedrockagent_prompt",
				"aws",
				bedrockagentAllowEmptyValues,
			))
		}
	}
	return nil
}
