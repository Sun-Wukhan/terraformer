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
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
)

var bedrockagentcoreAllowEmptyValues = []string{"tags."}

type BedrockAgentCoreGenerator struct {
	AWSService
}

// InitResources enumerates the Amazon Bedrock AgentCore control plane:
// agent runtimes and their endpoints, gateways with their targets and rules,
// memories and their strategies, browser and code interpreter tools,
// credential providers, policy engines and policies, evaluators, online
// evaluation configs, registries, workload identities, harnesses, and the
// token vault CMK.
//
// Deliberately not generated: aws_bedrockagentcore_resource_policy. The
// control plane only exposes GetResourcePolicy for a resource ARN you
// already know; there is no list API to discover policies from.
func (g *BedrockAgentCoreGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	client := bedrockagentcorecontrol.NewFromConfig(config)

	if err := g.addAgentRuntimes(client); err != nil {
		return err
	}
	if err := g.addGateways(client); err != nil {
		return err
	}
	if err := g.addMemories(client); err != nil {
		return err
	}
	for _, add := range []func(*bedrockagentcorecontrol.Client) error{
		g.addBrowsers,
		g.addBrowserProfiles,
		g.addCodeInterpreters,
		g.addCredentialProviders,
		g.addEvaluators,
		g.addOnlineEvaluationConfigs,
		g.addPolicyEngines,
		g.addRegistries,
		g.addWorkloadIdentities,
		g.addHarnesses,
		g.addTokenVaultCMK,
	} {
		if err := add(client); err != nil {
			return err
		}
	}
	return nil
}

func (g *BedrockAgentCoreGenerator) addAgentRuntimes(client *bedrockagentcorecontrol.Client) error {
	input := &bedrockagentcorecontrol.ListAgentRuntimesInput{}
	for {
		page, err := client.ListAgentRuntimes(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, runtime := range page.AgentRuntimes {
			runtimeID := *runtime.AgentRuntimeId
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				runtimeID,
				*runtime.AgentRuntimeName,
				"aws_bedrockagentcore_agent_runtime",
				"aws",
				bedrockagentcoreAllowEmptyValues,
			))
			if err := g.addAgentRuntimeEndpoints(client, runtimeID); err != nil {
				return err
			}
		}
		if page.NextToken == nil {
			return nil
		}
		input.NextToken = page.NextToken
	}
}

// Provider ID format: agent runtime ID, endpoint name.
func (g *BedrockAgentCoreGenerator) addAgentRuntimeEndpoints(client *bedrockagentcorecontrol.Client, runtimeID string) error {
	input := &bedrockagentcorecontrol.ListAgentRuntimeEndpointsInput{
		AgentRuntimeId: aws.String(runtimeID),
	}
	for {
		page, err := client.ListAgentRuntimeEndpoints(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, endpoint := range page.RuntimeEndpoints {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				runtimeID+","+*endpoint.Name,
				*endpoint.Name,
				"aws_bedrockagentcore_agent_runtime_endpoint",
				"aws",
				map[string]string{"agent_runtime_id": runtimeID},
				bedrockagentcoreAllowEmptyValues,
				map[string]interface{}{},
			))
		}
		if page.NextToken == nil {
			return nil
		}
		input.NextToken = page.NextToken
	}
}

func (g *BedrockAgentCoreGenerator) addGateways(client *bedrockagentcorecontrol.Client) error {
	input := &bedrockagentcorecontrol.ListGatewaysInput{}
	for {
		page, err := client.ListGateways(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, gateway := range page.Items {
			gatewayID := *gateway.GatewayId
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				gatewayID,
				*gateway.Name,
				"aws_bedrockagentcore_gateway",
				"aws",
				bedrockagentcoreAllowEmptyValues,
			))
			if err := g.addGatewayTargets(client, gatewayID); err != nil {
				return err
			}
			if err := g.addGatewayRules(client, gatewayID); err != nil {
				return err
			}
		}
		if page.NextToken == nil {
			return nil
		}
		input.NextToken = page.NextToken
	}
}

// Provider ID format: gateway identifier, target ID.
func (g *BedrockAgentCoreGenerator) addGatewayTargets(client *bedrockagentcorecontrol.Client, gatewayID string) error {
	input := &bedrockagentcorecontrol.ListGatewayTargetsInput{
		GatewayIdentifier: aws.String(gatewayID),
	}
	for {
		page, err := client.ListGatewayTargets(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, target := range page.Items {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				gatewayID+","+*target.TargetId,
				*target.Name,
				"aws_bedrockagentcore_gateway_target",
				"aws",
				map[string]string{"gateway_identifier": gatewayID},
				bedrockagentcoreAllowEmptyValues,
				map[string]interface{}{},
			))
		}
		if page.NextToken == nil {
			return nil
		}
		input.NextToken = page.NextToken
	}
}

// Provider ID format: gateway identifier, rule ID. Rules have no name in
// the API, so the rule ID doubles as the Terraform resource name.
func (g *BedrockAgentCoreGenerator) addGatewayRules(client *bedrockagentcorecontrol.Client, gatewayID string) error {
	input := &bedrockagentcorecontrol.ListGatewayRulesInput{
		GatewayIdentifier: aws.String(gatewayID),
	}
	for {
		page, err := client.ListGatewayRules(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, rule := range page.GatewayRules {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				gatewayID+","+*rule.RuleId,
				*rule.RuleId,
				"aws_bedrockagentcore_gateway_rule",
				"aws",
				map[string]string{"gateway_identifier": gatewayID},
				bedrockagentcoreAllowEmptyValues,
				map[string]interface{}{},
			))
		}
		if page.NextToken == nil {
			return nil
		}
		input.NextToken = page.NextToken
	}
}

func (g *BedrockAgentCoreGenerator) addMemories(client *bedrockagentcorecontrol.Client) error {
	input := &bedrockagentcorecontrol.ListMemoriesInput{}
	for {
		page, err := client.ListMemories(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, summary := range page.Memories {
			memoryID := *summary.Id
			detail, err := client.GetMemory(context.TODO(), &bedrockagentcorecontrol.GetMemoryInput{
				MemoryId: aws.String(memoryID),
			})
			if err != nil {
				return err
			}
			name := memoryID
			if detail.Memory != nil && detail.Memory.Name != nil {
				name = *detail.Memory.Name
			}
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				memoryID,
				name,
				"aws_bedrockagentcore_memory",
				"aws",
				bedrockagentcoreAllowEmptyValues,
			))
			if detail.Memory == nil {
				continue
			}
			for _, strategy := range detail.Memory.Strategies {
				g.Resources = append(g.Resources, terraformutils.NewResource(
					memoryID+","+*strategy.StrategyId,
					*strategy.Name,
					"aws_bedrockagentcore_memory_strategy",
					"aws",
					map[string]string{"memory_id": memoryID},
					bedrockagentcoreAllowEmptyValues,
					map[string]interface{}{},
				))
			}
		}
		if page.NextToken == nil {
			return nil
		}
		input.NextToken = page.NextToken
	}
}

func (g *BedrockAgentCoreGenerator) addBrowsers(client *bedrockagentcorecontrol.Client) error {
	input := &bedrockagentcorecontrol.ListBrowsersInput{}
	for {
		page, err := client.ListBrowsers(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, browser := range page.BrowserSummaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*browser.BrowserId,
				*browser.Name,
				"aws_bedrockagentcore_browser",
				"aws",
				bedrockagentcoreAllowEmptyValues,
			))
		}
		if page.NextToken == nil {
			return nil
		}
		input.NextToken = page.NextToken
	}
}

func (g *BedrockAgentCoreGenerator) addBrowserProfiles(client *bedrockagentcorecontrol.Client) error {
	input := &bedrockagentcorecontrol.ListBrowserProfilesInput{}
	for {
		page, err := client.ListBrowserProfiles(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, profile := range page.ProfileSummaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*profile.ProfileId,
				*profile.Name,
				"aws_bedrockagentcore_browser_profile",
				"aws",
				bedrockagentcoreAllowEmptyValues,
			))
		}
		if page.NextToken == nil {
			return nil
		}
		input.NextToken = page.NextToken
	}
}

func (g *BedrockAgentCoreGenerator) addCodeInterpreters(client *bedrockagentcorecontrol.Client) error {
	input := &bedrockagentcorecontrol.ListCodeInterpretersInput{}
	for {
		page, err := client.ListCodeInterpreters(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, interpreter := range page.CodeInterpreterSummaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*interpreter.CodeInterpreterId,
				*interpreter.Name,
				"aws_bedrockagentcore_code_interpreter",
				"aws",
				bedrockagentcoreAllowEmptyValues,
			))
		}
		if page.NextToken == nil {
			return nil
		}
		input.NextToken = page.NextToken
	}
}

// Credential providers are keyed by name in both the API and the provider.
func (g *BedrockAgentCoreGenerator) addCredentialProviders(client *bedrockagentcorecontrol.Client) error {
	apiKeyInput := &bedrockagentcorecontrol.ListApiKeyCredentialProvidersInput{}
	for {
		page, err := client.ListApiKeyCredentialProviders(context.TODO(), apiKeyInput)
		if err != nil {
			return err
		}
		for _, provider := range page.CredentialProviders {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*provider.Name,
				*provider.Name,
				"aws_bedrockagentcore_api_key_credential_provider",
				"aws",
				bedrockagentcoreAllowEmptyValues,
			))
		}
		if page.NextToken == nil {
			break
		}
		apiKeyInput.NextToken = page.NextToken
	}

	oauthInput := &bedrockagentcorecontrol.ListOauth2CredentialProvidersInput{}
	for {
		page, err := client.ListOauth2CredentialProviders(context.TODO(), oauthInput)
		if err != nil {
			return err
		}
		for _, provider := range page.CredentialProviders {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*provider.Name,
				*provider.Name,
				"aws_bedrockagentcore_oauth2_credential_provider",
				"aws",
				bedrockagentcoreAllowEmptyValues,
			))
		}
		if page.NextToken == nil {
			return nil
		}
		oauthInput.NextToken = page.NextToken
	}
}

func (g *BedrockAgentCoreGenerator) addEvaluators(client *bedrockagentcorecontrol.Client) error {
	input := &bedrockagentcorecontrol.ListEvaluatorsInput{}
	for {
		page, err := client.ListEvaluators(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, evaluator := range page.Evaluators {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*evaluator.EvaluatorId,
				*evaluator.EvaluatorName,
				"aws_bedrockagentcore_evaluator",
				"aws",
				bedrockagentcoreAllowEmptyValues,
			))
		}
		if page.NextToken == nil {
			return nil
		}
		input.NextToken = page.NextToken
	}
}

func (g *BedrockAgentCoreGenerator) addOnlineEvaluationConfigs(client *bedrockagentcorecontrol.Client) error {
	input := &bedrockagentcorecontrol.ListOnlineEvaluationConfigsInput{}
	for {
		page, err := client.ListOnlineEvaluationConfigs(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, evalConfig := range page.OnlineEvaluationConfigs {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*evalConfig.OnlineEvaluationConfigId,
				*evalConfig.OnlineEvaluationConfigName,
				"aws_bedrockagentcore_online_evaluation_config",
				"aws",
				bedrockagentcoreAllowEmptyValues,
			))
		}
		if page.NextToken == nil {
			return nil
		}
		input.NextToken = page.NextToken
	}
}

func (g *BedrockAgentCoreGenerator) addPolicyEngines(client *bedrockagentcorecontrol.Client) error {
	input := &bedrockagentcorecontrol.ListPolicyEnginesInput{}
	for {
		page, err := client.ListPolicyEngines(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, engine := range page.PolicyEngines {
			engineID := *engine.PolicyEngineId
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				engineID,
				*engine.Name,
				"aws_bedrockagentcore_policy_engine",
				"aws",
				bedrockagentcoreAllowEmptyValues,
			))
			if err := g.addPolicies(client, engineID); err != nil {
				return err
			}
		}
		if page.NextToken == nil {
			return nil
		}
		input.NextToken = page.NextToken
	}
}

// Provider ID format: policy engine ID, policy ID.
func (g *BedrockAgentCoreGenerator) addPolicies(client *bedrockagentcorecontrol.Client, engineID string) error {
	input := &bedrockagentcorecontrol.ListPoliciesInput{
		PolicyEngineId: aws.String(engineID),
	}
	for {
		page, err := client.ListPolicies(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, policy := range page.Policies {
			g.Resources = append(g.Resources, terraformutils.NewResource(
				engineID+","+*policy.PolicyId,
				*policy.Name,
				"aws_bedrockagentcore_policy",
				"aws",
				map[string]string{"policy_engine_id": engineID},
				bedrockagentcoreAllowEmptyValues,
				map[string]interface{}{},
			))
		}
		if page.NextToken == nil {
			return nil
		}
		input.NextToken = page.NextToken
	}
}

func (g *BedrockAgentCoreGenerator) addRegistries(client *bedrockagentcorecontrol.Client) error {
	input := &bedrockagentcorecontrol.ListRegistriesInput{}
	for {
		page, err := client.ListRegistries(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, registry := range page.Registries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*registry.RegistryId,
				*registry.Name,
				"aws_bedrockagentcore_registry",
				"aws",
				bedrockagentcoreAllowEmptyValues,
			))
		}
		if page.NextToken == nil {
			return nil
		}
		input.NextToken = page.NextToken
	}
}

// Workload identities are keyed by name in both the API and the provider.
func (g *BedrockAgentCoreGenerator) addWorkloadIdentities(client *bedrockagentcorecontrol.Client) error {
	input := &bedrockagentcorecontrol.ListWorkloadIdentitiesInput{}
	for {
		page, err := client.ListWorkloadIdentities(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, identity := range page.WorkloadIdentities {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*identity.Name,
				*identity.Name,
				"aws_bedrockagentcore_workload_identity",
				"aws",
				bedrockagentcoreAllowEmptyValues,
			))
		}
		if page.NextToken == nil {
			return nil
		}
		input.NextToken = page.NextToken
	}
}

func (g *BedrockAgentCoreGenerator) addHarnesses(client *bedrockagentcorecontrol.Client) error {
	input := &bedrockagentcorecontrol.ListHarnessesInput{}
	for {
		page, err := client.ListHarnesses(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, harness := range page.Harnesses {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*harness.HarnessId,
				*harness.HarnessName,
				"aws_bedrockagentcore_harness",
				"aws",
				bedrockagentcoreAllowEmptyValues,
			))
		}
		if page.NextToken == nil {
			return nil
		}
		input.NextToken = page.NextToken
	}
}

// The token vault CMK is a singleton keyed by token vault ID ("default"
// unless a custom vault is in use); it only exists as a managed resource
// when a customer-managed key is configured.
func (g *BedrockAgentCoreGenerator) addTokenVaultCMK(client *bedrockagentcorecontrol.Client) error {
	out, err := client.GetTokenVault(context.TODO(), &bedrockagentcorecontrol.GetTokenVaultInput{})
	if err != nil {
		return err
	}
	if out.KmsConfiguration == nil || out.KmsConfiguration.KmsKeyArn == nil {
		return nil
	}
	vaultID := "default"
	if out.TokenVaultId != nil && *out.TokenVaultId != "" {
		vaultID = *out.TokenVaultId
	}
	g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
		vaultID,
		"token_vault_cmk",
		"aws_bedrockagentcore_token_vault_cmk",
		"aws",
		bedrockagentcoreAllowEmptyValues,
	))
	return nil
}
