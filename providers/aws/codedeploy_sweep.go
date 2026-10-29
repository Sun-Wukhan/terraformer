package aws

// Sweep additions for CodeDeploy:
// Implemented: aws_codedeploy_deployment_config (ListDeploymentConfigs,
// ID = config name), aws_codedeploy_deployment_group (per-application
// ListDeploymentGroups, ID = "<application>:<group>").

import (
	"context"
	"fmt"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/codedeploy"
)

func (g *CodeDeployGenerator) loadSweep(svc *codedeploy.Client) error {
	p := codedeploy.NewListDeploymentConfigsPaginator(svc, &codedeploy.ListDeploymentConfigsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, cfg := range page.DeploymentConfigsList {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				cfg, cfg, "aws_codedeploy_deployment_config", "aws", codedeployAllowEmptyValues))
		}
	}
	ap := codedeploy.NewListApplicationsPaginator(svc, &codedeploy.ListApplicationsInput{})
	for ap.HasMorePages() {
		page, err := ap.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, app := range page.Applications {
			gp := codedeploy.NewListDeploymentGroupsPaginator(svc, &codedeploy.ListDeploymentGroupsInput{
				ApplicationName: &app,
			})
			for gp.HasMorePages() {
				gpage, err := gp.NextPage(context.TODO())
				if err != nil {
					break
				}
				for _, grp := range gpage.DeploymentGroups {
					id := fmt.Sprintf("%s:%s", app, grp)
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						id, grp, "aws_codedeploy_deployment_group", "aws", codedeployAllowEmptyValues))
				}
			}
		}
	}
	return nil
}
