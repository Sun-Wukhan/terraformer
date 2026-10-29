package aws

// GameLift sweep (new service): ListAliases / ListBuilds / ListFleets /
// ListScripts return IDs (import ID = the ID); game server groups import
// by name (ListGameServerGroups).
// Skipped: aws_gamelift_game_session_queue (the GameLift API has no list
// operation for game session queues — DescribeGameSessionQueues requires
// queue names/ARNs up front, so there is no enumeration path).

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/gamelift"
)

var gameliftAllowEmptyValues = []string{"tags."}

type GameLiftGenerator struct {
	AWSService
}

func (g *GameLiftGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := gamelift.NewFromConfig(config)

	var nextToken *string
	for {
		out, err := svc.ListAliases(context.TODO(), &gamelift.ListAliasesInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, a := range out.Aliases {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(a.AliasId), StringValue(a.Name), "aws_gamelift_alias", "aws", gameliftAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	nextToken = nil
	for {
		out, err := svc.ListBuilds(context.TODO(), &gamelift.ListBuildsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, b := range out.Builds {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(b.BuildId), StringValue(b.Name), "aws_gamelift_build", "aws", gameliftAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	nextToken = nil
	for {
		out, err := svc.ListFleets(context.TODO(), &gamelift.ListFleetsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, fleetID := range out.FleetIds {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				fleetID, fleetID, "aws_gamelift_fleet", "aws", gameliftAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	nextToken = nil
	for {
		out, err := svc.ListScripts(context.TODO(), &gamelift.ListScriptsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, s := range out.Scripts {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(s.ScriptId), StringValue(s.Name), "aws_gamelift_script", "aws", gameliftAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	nextToken = nil
	for {
		out, err := svc.ListGameServerGroups(context.TODO(), &gamelift.ListGameServerGroupsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, grp := range out.GameServerGroups {
			name := StringValue(grp.GameServerGroupName)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				name, name, "aws_gamelift_game_server_group", "aws", gameliftAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
