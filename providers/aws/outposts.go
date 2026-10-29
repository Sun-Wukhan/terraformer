package aws

// Outposts sweep (new service):
// Implemented: aws_outposts_capacity_task (ListOutposts then
// ListCapacityTasks per outpost; import identity is outpost + task ID).

import (
	"context"
	"fmt"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/outposts"
)

var outpostsAllowEmptyValues = []string{"tags."}

type OutpostsGenerator struct {
	AWSService
}

func (g *OutpostsGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := outposts.NewFromConfig(config)
	p := outposts.NewListOutpostsPaginator(svc, &outposts.ListOutpostsInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, outpost := range page.Outposts {
			outpostID := StringValue(outpost.OutpostId)
			tp := outposts.NewListCapacityTasksPaginator(svc, &outposts.ListCapacityTasksInput{
				OutpostIdentifierFilter: &outpostID,
			})
			for tp.HasMorePages() {
				tpage, err := tp.NextPage(context.TODO())
				if err != nil {
					break
				}
				for _, task := range tpage.CapacityTasks {
					id := fmt.Sprintf("%s/%s", outpostID, StringValue(task.CapacityTaskId))
					g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
						id, StringValue(task.CapacityTaskId), "aws_outposts_capacity_task", "aws", outpostsAllowEmptyValues))
				}
			}
		}
	}
	return nil
}
