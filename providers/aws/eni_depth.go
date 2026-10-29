// Copyright 2024 The Terraformer Authors.
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
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// getEniDepth extends EniGenerator with network interface attachments,
// permissions and security group attachments. Failures are logged and
// skipped so they never block the base aws_network_interface import.
//
// Note: aws_network_interface_attachment is also emitted for primary
// (device index 0) attachments; the provider manages those the same way,
// but configurations that also set the instance's primary interface inline
// should drop the duplicate.
func (g *EniGenerator) getEniDepth(svc *ec2.Client) {
	p := ec2.NewDescribeNetworkInterfacesPaginator(svc, &ec2.DescribeNetworkInterfacesInput{})
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			break
		}
		for _, eni := range page.NetworkInterfaces {
			if eni.NetworkInterfaceId == nil {
				continue
			}
			eniID := *eni.NetworkInterfaceId
			if eni.Attachment != nil && eni.Attachment.AttachmentId != nil {
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					StringValue(eni.Attachment.AttachmentId),
					StringValue(eni.Attachment.AttachmentId),
					"aws_network_interface_attachment",
					"aws",
					awsNetCoreAllowEmptyValues,
				))
			}
			// aws_network_interface_sg_attachment is imported by
			// "<eni id>_<security group id>".
			for _, group := range eni.Groups {
				if group.GroupId == nil {
					continue
				}
				id := eniID + "_" + *group.GroupId
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id,
					id,
					"aws_network_interface_sg_attachment",
					"aws",
					awsNetCoreAllowEmptyValues,
				))
			}
		}
	}
	permissionPaginator := ec2.NewDescribeNetworkInterfacePermissionsPaginator(svc, &ec2.DescribeNetworkInterfacePermissionsInput{})
	for permissionPaginator.HasMorePages() {
		page, err := permissionPaginator.NextPage(context.TODO())
		if err != nil {
			log.Println(err)
			break
		}
		for _, permission := range page.NetworkInterfacePermissions {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(permission.NetworkInterfacePermissionId),
				StringValue(permission.NetworkInterfacePermissionId),
				"aws_network_interface_permission",
				"aws",
				awsNetCoreAllowEmptyValues,
			))
		}
	}
}
