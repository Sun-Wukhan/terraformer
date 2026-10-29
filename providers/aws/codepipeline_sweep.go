package aws

// Sweep additions for CodePipeline:
// Implemented: aws_codepipeline_custom_action_type (ListActionTypes with
// the Custom owner filter, ID = "<category>:<provider>:<version>").

import (
	"context"
	"fmt"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/codepipeline"
	"github.com/aws/aws-sdk-go-v2/service/codepipeline/types"
)

func (g *CodePipelineGenerator) loadSweep(svc *codepipeline.Client) error {
	out, err := svc.ListActionTypes(context.TODO(), &codepipeline.ListActionTypesInput{
		ActionOwnerFilter: types.ActionOwnerCustom,
	})
	if err != nil {
		return err
	}
	for _, at := range out.ActionTypes {
		if at.Id == nil {
			continue
		}
		id := fmt.Sprintf("%s:%s:%s", string(at.Id.Category), StringValue(at.Id.Provider), StringValue(at.Id.Version))
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			id, StringValue(at.Id.Provider), "aws_codepipeline_custom_action_type", "aws", codepipelineAllowEmptyValues))
	}
	return nil
}
