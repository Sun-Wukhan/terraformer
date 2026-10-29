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
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
)

// Depth enumeration for CloudFront, called from
// CloudFrontGenerator.InitResources. Most CloudFront list APIs share
// the Marker/NextMarker pagination shape; the helpers below walk them
// uniformly. Key value store keys themselves live in the separate
// cloudfrontkeyvaluestore family (see cloudfrontkeyvaluestore.go).
func (g *CloudFrontGenerator) addDepthResources(svc *cloudfront.Client) error {
	for _, add := range []func(*cloudfront.Client) error{
		g.addAnycastIPLists,
		g.addConnectionFunctions,
		g.addConnectionGroups,
		g.addContinuousDeploymentPolicies,
		g.addDistributionTenants,
		g.addMultitenantDistributions,
		g.addFieldLevelEncryptionConfigs,
		g.addFieldLevelEncryptionProfiles,
		g.addFunctions,
		g.addKeyGroups,
		g.addKeyValueStores,
		g.addMonitoringSubscriptions,
		g.addOriginAccessControls,
		g.addOriginAccessIdentities,
		g.addOriginRequestPolicies,
		g.addPublicKeys,
		g.addRealtimeLogConfigs,
		g.addResponseHeadersPolicies,
		g.addTrustStores,
		g.addVpcOrigins,
	} {
		if err := add(svc); err != nil {
			return err
		}
	}
	return nil
}

func (g *CloudFrontGenerator) addSimple(id, name, resourceType string) {
	g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
		id,
		name,
		resourceType,
		"aws",
		cloudFrontAllowEmptyValues,
	))
}

func (g *CloudFrontGenerator) addAnycastIPLists(svc *cloudfront.Client) error {
	var marker *string
	for {
		out, err := svc.ListAnycastIpLists(context.TODO(), &cloudfront.ListAnycastIpListsInput{Marker: marker})
		if err != nil {
			return err
		}
		if out.AnycastIpLists == nil {
			return nil
		}
		for _, item := range out.AnycastIpLists.Items {
			g.addSimple(StringValue(item.Id), StringValue(item.Name), "aws_cloudfront_anycast_ip_list")
		}
		if out.AnycastIpLists.NextMarker == nil {
			return nil
		}
		marker = out.AnycastIpLists.NextMarker
	}
}

func (g *CloudFrontGenerator) addConnectionFunctions(svc *cloudfront.Client) error {
	var marker *string
	for {
		out, err := svc.ListConnectionFunctions(context.TODO(), &cloudfront.ListConnectionFunctionsInput{Marker: marker})
		if err != nil {
			return err
		}
		for _, item := range out.ConnectionFunctions {
			g.addSimple(StringValue(item.Id), StringValue(item.Name), "aws_cloudfront_connection_function")
		}
		marker = out.NextMarker
		if marker == nil {
			return nil
		}
	}
}

func (g *CloudFrontGenerator) addConnectionGroups(svc *cloudfront.Client) error {
	var marker *string
	for {
		out, err := svc.ListConnectionGroups(context.TODO(), &cloudfront.ListConnectionGroupsInput{Marker: marker})
		if err != nil {
			return err
		}
		for _, item := range out.ConnectionGroups {
			g.addSimple(StringValue(item.Id), StringValue(item.Name), "aws_cloudfront_connection_group")
		}
		marker = out.NextMarker
		if marker == nil {
			return nil
		}
	}
}

func (g *CloudFrontGenerator) addContinuousDeploymentPolicies(svc *cloudfront.Client) error {
	var marker *string
	for {
		out, err := svc.ListContinuousDeploymentPolicies(context.TODO(), &cloudfront.ListContinuousDeploymentPoliciesInput{Marker: marker})
		if err != nil {
			return err
		}
		for _, item := range out.ContinuousDeploymentPolicyList.Items {
			g.addSimple(StringValue(item.ContinuousDeploymentPolicy.Id), StringValue(item.ContinuousDeploymentPolicy.Id), "aws_cloudfront_continuous_deployment_policy")
		}
		marker = out.ContinuousDeploymentPolicyList.NextMarker
		if marker == nil {
			return nil
		}
	}
}

func (g *CloudFrontGenerator) addDistributionTenants(svc *cloudfront.Client) error {
	var marker *string
	for {
		out, err := svc.ListDistributionTenants(context.TODO(), &cloudfront.ListDistributionTenantsInput{Marker: marker})
		if err != nil {
			return err
		}
		for _, item := range out.DistributionTenantList {
			name := StringValue(item.Name)
			if name == "" {
				name = StringValue(item.Id)
			}
			g.addSimple(StringValue(item.Id), name, "aws_cloudfront_distribution_tenant")
		}
		marker = out.NextMarker
		if marker == nil {
			return nil
		}
	}
}

// Multi-tenant distributions are distributions whose connection mode
// is tenant-only; the dedicated list call filters on exactly that.
func (g *CloudFrontGenerator) addMultitenantDistributions(svc *cloudfront.Client) error {
	var marker *string
	for {
		out, err := svc.ListDistributionsByConnectionMode(context.TODO(), &cloudfront.ListDistributionsByConnectionModeInput{
			ConnectionMode: types.ConnectionModeTenantOnly,
			Marker:         marker,
		})
		if err != nil {
			return err
		}
		if out.DistributionList == nil {
			return nil
		}
		for _, item := range out.DistributionList.Items {
			g.addSimple(StringValue(item.Id), StringValue(item.Id), "aws_cloudfront_multitenant_distribution")
		}
		marker = out.DistributionList.NextMarker
		if marker == nil {
			return nil
		}
	}
}

func (g *CloudFrontGenerator) addFieldLevelEncryptionConfigs(svc *cloudfront.Client) error {
	var marker *string
	for {
		out, err := svc.ListFieldLevelEncryptionConfigs(context.TODO(), &cloudfront.ListFieldLevelEncryptionConfigsInput{Marker: marker})
		if err != nil {
			return err
		}
		for _, item := range out.FieldLevelEncryptionList.Items {
			g.addSimple(StringValue(item.Id), StringValue(item.Id), "aws_cloudfront_field_level_encryption_config")
		}
		marker = out.FieldLevelEncryptionList.NextMarker
		if marker == nil {
			return nil
		}
	}
}

func (g *CloudFrontGenerator) addFieldLevelEncryptionProfiles(svc *cloudfront.Client) error {
	var marker *string
	for {
		out, err := svc.ListFieldLevelEncryptionProfiles(context.TODO(), &cloudfront.ListFieldLevelEncryptionProfilesInput{Marker: marker})
		if err != nil {
			return err
		}
		for _, item := range out.FieldLevelEncryptionProfileList.Items {
			g.addSimple(StringValue(item.Id), StringValue(item.Name), "aws_cloudfront_field_level_encryption_profile")
		}
		marker = out.FieldLevelEncryptionProfileList.NextMarker
		if marker == nil {
			return nil
		}
	}
}

// Provider ID format: the function name.
func (g *CloudFrontGenerator) addFunctions(svc *cloudfront.Client) error {
	var marker *string
	for {
		out, err := svc.ListFunctions(context.TODO(), &cloudfront.ListFunctionsInput{Marker: marker})
		if err != nil {
			return err
		}
		for _, item := range out.FunctionList.Items {
			g.addSimple(StringValue(item.Name), StringValue(item.Name), "aws_cloudfront_function")
		}
		marker = out.FunctionList.NextMarker
		if marker == nil {
			return nil
		}
	}
}

func (g *CloudFrontGenerator) addKeyGroups(svc *cloudfront.Client) error {
	var marker *string
	for {
		out, err := svc.ListKeyGroups(context.TODO(), &cloudfront.ListKeyGroupsInput{Marker: marker})
		if err != nil {
			return err
		}
		for _, item := range out.KeyGroupList.Items {
			g.addSimple(StringValue(item.KeyGroup.Id), StringValue(item.KeyGroup.Id), "aws_cloudfront_key_group")
		}
		marker = out.KeyGroupList.NextMarker
		if marker == nil {
			return nil
		}
	}
}

// Provider ID format: the key value store name.
func (g *CloudFrontGenerator) addKeyValueStores(svc *cloudfront.Client) error {
	var marker *string
	for {
		out, err := svc.ListKeyValueStores(context.TODO(), &cloudfront.ListKeyValueStoresInput{Marker: marker})
		if err != nil {
			return err
		}
		for _, item := range out.KeyValueStoreList.Items {
			g.addSimple(StringValue(item.Name), StringValue(item.Name), "aws_cloudfront_key_value_store")
		}
		marker = out.KeyValueStoreList.NextMarker
		if marker == nil {
			return nil
		}
	}
}

// The monitoring subscription is a per-distribution singleton keyed
// by the distribution ID; it is generated only for distributions
// whose realtime metrics subscription is enabled.
func (g *CloudFrontGenerator) addMonitoringSubscriptions(svc *cloudfront.Client) error {
	p := cloudfront.NewListDistributionsPaginator(svc, &cloudfront.ListDistributionsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, distribution := range page.DistributionList.Items {
			distributionID := StringValue(distribution.Id)
			out, err := svc.GetMonitoringSubscription(context.TODO(), &cloudfront.GetMonitoringSubscriptionInput{
				DistributionId: aws.String(distributionID),
			})
			if err != nil {
				return err
			}
			if out.MonitoringSubscription == nil ||
				out.MonitoringSubscription.RealtimeMetricsSubscriptionConfig == nil ||
				out.MonitoringSubscription.RealtimeMetricsSubscriptionConfig.RealtimeMetricsSubscriptionStatus != types.RealtimeMetricsSubscriptionStatusEnabled {
				continue
			}
			g.addSimple(distributionID, distributionID, "aws_cloudfront_monitoring_subscription")
		}
	}
	return nil
}

func (g *CloudFrontGenerator) addOriginAccessControls(svc *cloudfront.Client) error {
	var marker *string
	for {
		out, err := svc.ListOriginAccessControls(context.TODO(), &cloudfront.ListOriginAccessControlsInput{Marker: marker})
		if err != nil {
			return err
		}
		for _, item := range out.OriginAccessControlList.Items {
			g.addSimple(StringValue(item.Id), StringValue(item.Name), "aws_cloudfront_origin_access_control")
		}
		marker = out.OriginAccessControlList.NextMarker
		if marker == nil {
			return nil
		}
	}
}

func (g *CloudFrontGenerator) addOriginAccessIdentities(svc *cloudfront.Client) error {
	var marker *string
	for {
		out, err := svc.ListCloudFrontOriginAccessIdentities(context.TODO(), &cloudfront.ListCloudFrontOriginAccessIdentitiesInput{Marker: marker})
		if err != nil {
			return err
		}
		for _, item := range out.CloudFrontOriginAccessIdentityList.Items {
			g.addSimple(StringValue(item.Id), StringValue(item.Id), "aws_cloudfront_origin_access_identity")
		}
		marker = out.CloudFrontOriginAccessIdentityList.NextMarker
		if marker == nil {
			return nil
		}
	}
}

func (g *CloudFrontGenerator) addOriginRequestPolicies(svc *cloudfront.Client) error {
	var marker *string
	for {
		out, err := svc.ListOriginRequestPolicies(context.TODO(), &cloudfront.ListOriginRequestPoliciesInput{Marker: marker})
		if err != nil {
			return err
		}
		for _, item := range out.OriginRequestPolicyList.Items {
			g.addSimple(StringValue(item.OriginRequestPolicy.Id), StringValue(item.OriginRequestPolicy.Id), "aws_cloudfront_origin_request_policy")
		}
		marker = out.OriginRequestPolicyList.NextMarker
		if marker == nil {
			return nil
		}
	}
}

func (g *CloudFrontGenerator) addPublicKeys(svc *cloudfront.Client) error {
	var marker *string
	for {
		out, err := svc.ListPublicKeys(context.TODO(), &cloudfront.ListPublicKeysInput{Marker: marker})
		if err != nil {
			return err
		}
		for _, item := range out.PublicKeyList.Items {
			g.addSimple(StringValue(item.Id), StringValue(item.Name), "aws_cloudfront_public_key")
		}
		marker = out.PublicKeyList.NextMarker
		if marker == nil {
			return nil
		}
	}
}

// Provider ID format: the realtime log config ARN.
func (g *CloudFrontGenerator) addRealtimeLogConfigs(svc *cloudfront.Client) error {
	var marker *string
	for {
		out, err := svc.ListRealtimeLogConfigs(context.TODO(), &cloudfront.ListRealtimeLogConfigsInput{Marker: marker})
		if err != nil {
			return err
		}
		if out.RealtimeLogConfigs == nil {
			return nil
		}
		for _, item := range out.RealtimeLogConfigs.Items {
			g.addSimple(StringValue(item.ARN), StringValue(item.Name), "aws_cloudfront_realtime_log_config")
		}
		marker = out.RealtimeLogConfigs.NextMarker
		if marker == nil {
			return nil
		}
	}
}

func (g *CloudFrontGenerator) addResponseHeadersPolicies(svc *cloudfront.Client) error {
	var marker *string
	for {
		out, err := svc.ListResponseHeadersPolicies(context.TODO(), &cloudfront.ListResponseHeadersPoliciesInput{Marker: marker})
		if err != nil {
			return err
		}
		for _, item := range out.ResponseHeadersPolicyList.Items {
			g.addSimple(StringValue(item.ResponseHeadersPolicy.Id), StringValue(item.ResponseHeadersPolicy.Id), "aws_cloudfront_response_headers_policy")
		}
		marker = out.ResponseHeadersPolicyList.NextMarker
		if marker == nil {
			return nil
		}
	}
}

func (g *CloudFrontGenerator) addTrustStores(svc *cloudfront.Client) error {
	var marker *string
	for {
		out, err := svc.ListTrustStores(context.TODO(), &cloudfront.ListTrustStoresInput{Marker: marker})
		if err != nil {
			return err
		}
		for _, item := range out.TrustStoreList {
			g.addSimple(StringValue(item.Id), StringValue(item.Name), "aws_cloudfront_trust_store")
		}
		marker = out.NextMarker
		if marker == nil {
			return nil
		}
	}
}

func (g *CloudFrontGenerator) addVpcOrigins(svc *cloudfront.Client) error {
	var marker *string
	for {
		out, err := svc.ListVpcOrigins(context.TODO(), &cloudfront.ListVpcOriginsInput{Marker: marker})
		if err != nil {
			return err
		}
		for _, item := range out.VpcOriginList.Items {
			g.addSimple(StringValue(item.Id), StringValue(item.Name), "aws_cloudfront_vpc_origin")
		}
		marker = out.VpcOriginList.NextMarker
		if marker == nil {
			return nil
		}
	}
}
