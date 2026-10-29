package aws

// Sweep additions for EKS, enumerated per cluster:
// Implemented: aws_eks_addon (ListAddons), aws_eks_fargate_profile
// (ListFargateProfiles), aws_eks_identity_provider_config
// (ListIdentityProviderConfigs), aws_eks_pod_identity_association
// (ListPodIdentityAssociations), aws_eks_access_entry (ListAccessEntries),
// aws_eks_access_policy_association (ListAssociatedAccessPolicies per
// access entry). Import identities are (cluster_name, name/principal)
// pairs per the provider docs, joined with ":" in the legacy ID.
// Also implemented: aws_eks_capability (ListCapabilities per cluster,
// identity cluster_name + capability_name).

import (
	"context"
	"fmt"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/eks"
)

func (g *EksGenerator) loadSweep(svc *eks.Client) error {
	p := eks.NewListClustersPaginator(svc, &eks.ListClustersInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, clusterName := range page.Clusters {
			var capToken *string
			for {
				caps, err := svc.ListCapabilities(context.TODO(), &eks.ListCapabilitiesInput{
					ClusterName: &clusterName, NextToken: capToken,
				})
				if err != nil {
					break
				}
				for _, cap := range caps.Capabilities {
					id := fmt.Sprintf("%s:%s", clusterName, StringValue(cap.CapabilityName))
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						id, StringValue(cap.CapabilityName), "aws_eks_capability", "aws", eksAllowEmptyValues))
				}
				capToken = caps.NextToken
				if capToken == nil {
					break
				}
			}
			ap := eks.NewListAddonsPaginator(svc, &eks.ListAddonsInput{ClusterName: &clusterName})
			for ap.HasMorePages() {
				apage, err := ap.NextPage(context.TODO())
				if err != nil {
					break
				}
				for _, addon := range apage.Addons {
					id := fmt.Sprintf("%s:%s", clusterName, addon)
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						id, addon, "aws_eks_addon", "aws", eksAllowEmptyValues))
				}
			}
			fp := eks.NewListFargateProfilesPaginator(svc, &eks.ListFargateProfilesInput{ClusterName: &clusterName})
			for fp.HasMorePages() {
				fpage, err := fp.NextPage(context.TODO())
				if err != nil {
					break
				}
				for _, profile := range fpage.FargateProfileNames {
					id := fmt.Sprintf("%s:%s", clusterName, profile)
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						id, profile, "aws_eks_fargate_profile", "aws", eksAllowEmptyValues))
				}
			}
			ipc, err := svc.ListIdentityProviderConfigs(context.TODO(), &eks.ListIdentityProviderConfigsInput{ClusterName: &clusterName})
			if err == nil {
				for _, cfg := range ipc.IdentityProviderConfigs {
					id := fmt.Sprintf("%s:%s", clusterName, StringValue(cfg.Name))
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						id, StringValue(cfg.Name), "aws_eks_identity_provider_config", "aws", eksAllowEmptyValues))
				}
			}
			pp := eks.NewListPodIdentityAssociationsPaginator(svc, &eks.ListPodIdentityAssociationsInput{ClusterName: &clusterName})
			for pp.HasMorePages() {
				ppage, err := pp.NextPage(context.TODO())
				if err != nil {
					break
				}
				for _, assoc := range ppage.Associations {
					id := fmt.Sprintf("%s:%s", clusterName, StringValue(assoc.AssociationId))
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						id, StringValue(assoc.AssociationId), "aws_eks_pod_identity_association", "aws", eksAllowEmptyValues))
				}
			}
			ep := eks.NewListAccessEntriesPaginator(svc, &eks.ListAccessEntriesInput{ClusterName: &clusterName})
			for ep.HasMorePages() {
				epage, err := ep.NextPage(context.TODO())
				if err != nil {
					break
				}
				for _, principalArn := range epage.AccessEntries {
					id := fmt.Sprintf("%s:%s", clusterName, principalArn)
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						id, principalArn, "aws_eks_access_entry", "aws", eksAllowEmptyValues))
					policies, err := svc.ListAssociatedAccessPolicies(context.TODO(), &eks.ListAssociatedAccessPoliciesInput{
						ClusterName: &clusterName, PrincipalArn: &principalArn,
					})
					if err == nil {
						for _, pol := range policies.AssociatedAccessPolicies {
							pid := fmt.Sprintf("%s:%s:%s", clusterName, principalArn, StringValue(pol.PolicyArn))
							g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
								pid, StringValue(pol.PolicyArn), "aws_eks_access_policy_association", "aws", eksAllowEmptyValues))
						}
					}
				}
			}
		}
	}
	return nil
}
