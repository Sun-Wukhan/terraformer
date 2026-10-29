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
	"log"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/ram"
	ramtypes "github.com/aws/aws-sdk-go-v2/service/ram/types"
)

var ramAllowEmptyValues = []string{"tags."}

type RamGenerator struct {
	AWSService
}

// InitResources enumerates RAM resource shares owned by this
// account (GetResourceShares with ResourceOwner=SELF; provider ID =
// share ARN), each share's principal and resource associations
// (GetResourceShareAssociations; provider ID = share ARN and the
// associated entity, comma-separated — verified against the
// provider docs for both aws_ram_principal_association and
// aws_ram_resource_association), and the RAM permissions known to
// the account (ListPermissions; provider ID = permission ARN).
//
// Intentionally not generated:
//   - aws_ram_resource_share_accepter is the accepter side of a
//     share invitation in the receiving account.
//   - aws_ram_resource_share_associations_exclusive conflicts with
//     the individual association resources for the same share.
//   - aws_ram_sharing_with_organization is the singleton
//     organization-wide sharing enablement; the SDK exposes no
//     get/list for its state.
func (g *RamGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := ram.NewFromConfig(config)

	p := ram.NewGetResourceSharesPaginator(svc, &ram.GetResourceSharesInput{
		ResourceOwner: ramtypes.ResourceOwnerSelf,
	})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, share := range page.ResourceShares {
			arn := StringValue(share.ResourceShareArn)
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn,
				StringValue(share.Name),
				"aws_ram_resource_share",
				"aws",
				ramAllowEmptyValues))
			g.addAssociations(svc, arn, ramtypes.ResourceShareAssociationTypePrincipal, "aws_ram_principal_association")
			g.addAssociations(svc, arn, ramtypes.ResourceShareAssociationTypeResource, "aws_ram_resource_association")
		}
	}

	permissions, err := svc.ListPermissions(context.TODO(), &ram.ListPermissionsInput{})
	if err != nil {
		return err
	}
	for _, permission := range permissions.Permissions {
		arn := StringValue(permission.Arn)
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			arn,
			StringValue(permission.Name),
			"aws_ram_permission",
			"aws",
			ramAllowEmptyValues))
	}
	return nil
}

func (g *RamGenerator) addAssociations(svc *ram.Client, shareArn string, associationType ramtypes.ResourceShareAssociationType, resourceType string) {
	p := ram.NewGetResourceShareAssociationsPaginator(svc, &ram.GetResourceShareAssociationsInput{
		AssociationType:   associationType,
		ResourceShareArns: []string{shareArn},
	})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			return
		}
		for _, association := range page.ResourceShareAssociations {
			entity := StringValue(association.AssociatedEntity)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				shareArn+","+entity,
				entity,
				resourceType,
				"aws",
				map[string]string{"resource_share_arn": shareArn},
				ramAllowEmptyValues,
				map[string]interface{}{}))
		}
	}
}
