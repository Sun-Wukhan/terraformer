package aws

// AppStream 2.0 sweep (new service):
// Implemented: aws_appstream_fleet (DescribeFleets, ID = name),
// aws_appstream_stack (DescribeStacks, ID = name),
// aws_appstream_image_builder (DescribeImageBuilders, ID = name),
// aws_appstream_directory_config (DescribeDirectoryConfigs, ID = directory
// name), aws_appstream_user (DescribeUsers for SAML and USERPOOL auth,
// ID = "<user-name>/<auth-type>"), aws_appstream_user_stack_association
// (DescribeUserStackAssociations per user, ID =
// "<user-name>/<auth-type>/<stack-name>").
// Skipped: aws_appstream_fleet_stack_association (association glue; the
// SDK exposes no list/describe operation enumerating fleet-stack pairs).

import (
	"context"
	"fmt"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/appstream"
	"github.com/aws/aws-sdk-go-v2/service/appstream/types"
)

var appstreamAllowEmptyValues = []string{"tags."}

type AppStreamGenerator struct {
	AWSService
}

func (g *AppStreamGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := appstream.NewFromConfig(config)

	var nextToken *string
	for {
		out, err := svc.DescribeFleets(context.TODO(), &appstream.DescribeFleetsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, fleet := range out.Fleets {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(fleet.Name), StringValue(fleet.Name), "aws_appstream_fleet", "aws", appstreamAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	nextToken = nil
	for {
		out, err := svc.DescribeStacks(context.TODO(), &appstream.DescribeStacksInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, stack := range out.Stacks {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(stack.Name), StringValue(stack.Name), "aws_appstream_stack", "aws", appstreamAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	nextToken = nil
	for {
		out, err := svc.DescribeImageBuilders(context.TODO(), &appstream.DescribeImageBuildersInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, ib := range out.ImageBuilders {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(ib.Name), StringValue(ib.Name), "aws_appstream_image_builder", "aws", appstreamAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	nextToken = nil
	for {
		out, err := svc.DescribeDirectoryConfigs(context.TODO(), &appstream.DescribeDirectoryConfigsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, dc := range out.DirectoryConfigs {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(dc.DirectoryName), StringValue(dc.DirectoryName), "aws_appstream_directory_config", "aws", appstreamAllowEmptyValues))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	for _, authType := range []types.AuthenticationType{types.AuthenticationTypeSaml, types.AuthenticationTypeUserpool} {
		nextToken = nil
		for {
			out, err := svc.DescribeUsers(context.TODO(), &appstream.DescribeUsersInput{
				AuthenticationType: authType, NextToken: nextToken,
			})
			if err != nil {
				return err
			}
			for _, user := range out.Users {
				userName := StringValue(user.UserName)
				id := fmt.Sprintf("%s/%s", userName, string(authType))
				g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
					id, userName, "aws_appstream_user", "aws", appstreamAllowEmptyValues))
				assoc, err := svc.DescribeUserStackAssociations(context.TODO(), &appstream.DescribeUserStackAssociationsInput{
					UserName: &userName, AuthenticationType: authType,
				})
				if err == nil {
					for _, a := range assoc.UserStackAssociations {
						aid := fmt.Sprintf("%s/%s/%s", userName, string(authType), StringValue(a.StackName))
						g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
							aid, aid, "aws_appstream_user_stack_association", "aws", appstreamAllowEmptyValues))
					}
				}
			}
			nextToken = out.NextToken
			if nextToken == nil {
				break
			}
		}
	}
	return nil
}
