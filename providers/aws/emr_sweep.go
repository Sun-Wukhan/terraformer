package aws

// Sweep additions for EMR:
// Implemented: aws_emr_instance_fleet and aws_emr_instance_group
// (per-cluster ListInstanceFleets / ListInstanceGroups, ID =
// "<cluster-id>/<fleet-or-group-id>"), aws_emr_managed_scaling_policy
// (GetManagedScalingPolicy per active cluster, ID = cluster ID),
// aws_emr_studio (ListStudios), aws_emr_studio_session_mapping
// (ListStudioSessionMappings per studio, ID =
// "<studio-id>:<identity-type>:<identity-id>").
// Skipped: aws_emr_block_public_access_configuration (singleton account
// configuration, ID is the literal "current").

import (
	"context"
	"fmt"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/emr"
	"github.com/aws/aws-sdk-go-v2/service/emr/types"
)

func (g *EmrGenerator) loadSweep(client *emr.Client) error {
	// Studios + session mappings
	sp := emr.NewListStudiosPaginator(client, &emr.ListStudiosInput{})
	for sp.HasMorePages() {
		page, err := sp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, studio := range page.Studios {
			studioID := StringValue(studio.StudioId)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				studioID, StringValue(studio.Name), "aws_emr_studio", "aws", emrAllowEmptyValues))
			mp := emr.NewListStudioSessionMappingsPaginator(client, &emr.ListStudioSessionMappingsInput{StudioId: &studioID})
			for mp.HasMorePages() {
				mpage, err := mp.NextPage(context.TODO())
				if err != nil {
					break
				}
				for _, m := range mpage.SessionMappings {
					id := fmt.Sprintf("%s:%s:%s", studioID, string(m.IdentityType), StringValue(m.IdentityId))
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						id, id, "aws_emr_studio_session_mapping", "aws", emrAllowEmptyValues))
				}
			}
		}
	}

	// Per-cluster fleets, groups, managed scaling policies
	cp := emr.NewListClustersPaginator(client, &emr.ListClustersInput{
		ClusterStates: []types.ClusterState{
			types.ClusterStateStarting, types.ClusterStateBootstrapping,
			types.ClusterStateRunning, types.ClusterStateWaiting,
		},
	})
	for cp.HasMorePages() {
		page, err := cp.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, cluster := range page.Clusters {
			clusterID := StringValue(cluster.Id)
			fp := emr.NewListInstanceFleetsPaginator(client, &emr.ListInstanceFleetsInput{ClusterId: &clusterID})
			for fp.HasMorePages() {
				fpage, err := fp.NextPage(context.TODO())
				if err != nil {
					break
				}
				for _, fleet := range fpage.InstanceFleets {
					id := fmt.Sprintf("%s/%s", clusterID, StringValue(fleet.Id))
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						id, StringValue(fleet.Name), "aws_emr_instance_fleet", "aws", emrAllowEmptyValues))
				}
			}
			gp := emr.NewListInstanceGroupsPaginator(client, &emr.ListInstanceGroupsInput{ClusterId: &clusterID})
			for gp.HasMorePages() {
				gpage, err := gp.NextPage(context.TODO())
				if err != nil {
					break
				}
				for _, grp := range gpage.InstanceGroups {
					id := fmt.Sprintf("%s/%s", clusterID, StringValue(grp.Id))
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						id, StringValue(grp.Name), "aws_emr_instance_group", "aws", emrAllowEmptyValues))
				}
			}
			if pol, err := client.GetManagedScalingPolicy(context.TODO(), &emr.GetManagedScalingPolicyInput{ClusterId: &clusterID}); err == nil && pol.ManagedScalingPolicy != nil {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					clusterID, clusterID, "aws_emr_managed_scaling_policy", "aws", emrAllowEmptyValues))
			}
		}
	}
	return nil
}
