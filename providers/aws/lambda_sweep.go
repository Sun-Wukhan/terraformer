package aws

// Sweep additions for Lambda:
// Implemented: aws_lambda_alias (ListAliases per function),
// aws_lambda_code_signing_config (ListCodeSigningConfigs, ID = config ARN),
// aws_lambda_function_url (ListFunctionUrlConfigs per function, ID = function name),
// aws_lambda_provisioned_concurrency_config (ListProvisionedConcurrencyConfigs,
// ID = "<function>,<alias>"), aws_lambda_runtime_management_config
// (GetRuntimeManagementConfig per function, ID = "<function>,$LATEST"),
// aws_lambda_layer_version_permission (GetLayerVersionPolicy per layer version,
// ID = "<layer-arn>,<version>").
// Also implemented: aws_lambda_capacity_provider (ListCapacityProviders,
// ID = provider name), aws_lambda_function_recursion_config
// (GetFunctionRecursionConfig per function, ID = function name),
// aws_lambda_function_scaling_config (GetFunctionScalingConfig per
// function with qualifier $LATEST.PUBLISHED).
// Skipped: aws_lambda_invocation (action/invocation result, not a managed
// resource), aws_lambda_resource_policy (policy glue already emitted as
// aws_lambda_permission from GetPolicy statements),
// aws_lambdamicrovms_image / aws_lambdamicrovms_microvm /
// aws_lambdacore_network_connector (no aws-sdk-go-v2 service module exists
// for the lambdamicrovms / lambdacore APIs).

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

type layerPolicyDoc struct {
	Statement []struct {
		Sid       string          `json:"Sid"`
		Principal json.RawMessage `json:"Principal"`
		Action    string          `json:"Action"`
	} `json:"Statement"`
}

func (g *LambdaGenerator) loadSweep(svc *lambda.Client) error {
	// Capacity providers (account level)
	var marker *string
	for {
		out, err := svc.ListCapacityProviders(context.TODO(), &lambda.ListCapacityProvidersInput{Marker: marker})
		if err != nil {
			return err
		}
		for _, cp := range out.CapacityProviders {
			// The SDK's CapacityProvider type carries no Name field; the
			// provider identity (capacity provider name) is the final
			// segment of the capacity provider ARN.
			arn := StringValue(cp.CapacityProviderArn)
			name := arn[strings.LastIndex(arn, ":")+1:]
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_lambda_capacity_provider", "aws", lambdaAllowEmptyValues))
		}
		marker = out.NextMarker
		if marker == nil {
			break
		}
	}
	// Code signing configs (account level)
	cp := lambda.NewListCodeSigningConfigsPaginator(svc, &lambda.ListCodeSigningConfigsInput{})
	for cp.HasMorePages() {
		page, err := cp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, csc := range page.CodeSigningConfigs {
			arn := StringValue(csc.CodeSigningConfigArn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, StringValue(csc.CodeSigningConfigId), "aws_lambda_code_signing_config", "aws", lambdaAllowEmptyValues))
		}
	}

	p := lambda.NewListFunctionsPaginator(svc, &lambda.ListFunctionsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, fn := range page.Functions {
			name := StringValue(fn.FunctionName)
			// Aliases
			ap := lambda.NewListAliasesPaginator(svc, &lambda.ListAliasesInput{FunctionName: fn.FunctionName})
			for ap.HasMorePages() {
				apage, err := ap.NextPage(context.TODO())
				if err != nil {
					break
				}
				for _, alias := range apage.Aliases {
					id := fmt.Sprintf("%s/%s", name, StringValue(alias.Name))
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						id, StringValue(alias.Name), "aws_lambda_alias", "aws", lambdaAllowEmptyValues))
				}
			}
			// Function URLs
			urls, err := svc.ListFunctionUrlConfigs(context.TODO(), &lambda.ListFunctionUrlConfigsInput{FunctionName: fn.FunctionName})
			if err == nil {
				for range urls.FunctionUrlConfigs {
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						name, name, "aws_lambda_function_url", "aws", lambdaAllowEmptyValues))
				}
			}
			// Provisioned concurrency
			pp := lambda.NewListProvisionedConcurrencyConfigsPaginator(svc, &lambda.ListProvisionedConcurrencyConfigsInput{FunctionName: fn.FunctionName})
			for pp.HasMorePages() {
				ppage, err := pp.NextPage(context.TODO())
				if err != nil {
					break
				}
				for _, cfg := range ppage.ProvisionedConcurrencyConfigs {
					// Import ID is "<function>,<alias>"; alias is last ARN segment.
					arnParts := strings.Split(StringValue(cfg.FunctionArn), ":")
					id := fmt.Sprintf("%s,%s", name, arnParts[len(arnParts)-1])
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						id, id, "aws_lambda_provisioned_concurrency_config", "aws", lambdaAllowEmptyValues))
				}
			}
			// Recursion config (present on every function)
			if _, err := svc.GetFunctionRecursionConfig(context.TODO(), &lambda.GetFunctionRecursionConfigInput{
				FunctionName: fn.FunctionName,
			}); err == nil {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					name, name, "aws_lambda_function_recursion_config", "aws", lambdaAllowEmptyValues))
			}
			// Scaling config ($LATEST.PUBLISHED)
			if _, err := svc.GetFunctionScalingConfig(context.TODO(), &lambda.GetFunctionScalingConfigInput{
				FunctionName: fn.FunctionName, Qualifier: aws.String("$LATEST.PUBLISHED"),
			}); err == nil {
				id := fmt.Sprintf("%s/%s", name, "$LATEST.PUBLISHED")
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, id, "aws_lambda_function_scaling_config", "aws", lambdaAllowEmptyValues))
			}
			// Runtime management config ($LATEST)
			if _, err := svc.GetRuntimeManagementConfig(context.TODO(), &lambda.GetRuntimeManagementConfigInput{
				FunctionName: fn.FunctionName, Qualifier: aws.String("$LATEST"),
			}); err == nil {
				id := fmt.Sprintf("%s,$LATEST", name)
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, id, "aws_lambda_runtime_management_config", "aws", lambdaAllowEmptyValues))
			}
		}
	}

	// Layer version permissions
	pl := lambda.NewListLayersPaginator(svc, &lambda.ListLayersInput{})
	for pl.HasMorePages() {
		plage, err := pl.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, layer := range plage.Layers {
			pv := lambda.NewListLayerVersionsPaginator(svc, &lambda.ListLayerVersionsInput{LayerName: layer.LayerName})
			for pv.HasMorePages() {
				pvage, err := pv.NextPage(context.TODO())
				if err != nil {
					break
				}
				for _, lv := range pvage.LayerVersions {
					pol, err := svc.GetLayerVersionPolicy(context.TODO(), &lambda.GetLayerVersionPolicyInput{
						LayerName: layer.LayerName, VersionNumber: &lv.Version,
					})
					if err != nil || pol.Policy == nil {
						continue
					}
					var doc layerPolicyDoc
					if err := json.Unmarshal([]byte(*pol.Policy), &doc); err != nil {
						continue
					}
					layerArn := StringValue(lv.LayerVersionArn)
					if idx := strings.LastIndex(layerArn, ":"); idx > 0 {
						layerArn = layerArn[:idx]
					}
					for _, st := range doc.Statement {
						id := fmt.Sprintf("%s,%d", layerArn, lv.Version)
						g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
							id, st.Sid, "aws_lambda_layer_version_permission", "aws", lambdaAllowEmptyValues))
					}
				}
			}
		}
	}
	return nil
}
