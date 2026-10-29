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
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/kinesisanalytics"
	"github.com/aws/aws-sdk-go-v2/service/kinesisanalyticsv2"
	"github.com/aws/aws-sdk-go-v2/service/kinesisvideo"
)

// Kinesis depth for the AWS sweep. Provider importer IDs verified against
// the terraform-provider-aws docs: stream consumers by consumer ARN
// (verified directly), the stream resource policy by stream ARN, Kinesis
// Video streams and both Kinesis Analytics application generations by
// application/stream ARN.
//
// Deliberately not generated:
//   - aws_kinesis_account_settings: account-level singleton settings, no
//     list API and nothing to enumerate.
//   - aws_kinesisanalyticsv2_application_snapshot: a snapshot is a
//     point-in-time version artifact of an application (derived/version
//     glue), not a standalone managed resource.
func (g *KinesisGenerator) loadExtended(config aws.Config) error {
	svc := kinesis.NewFromConfig(config)
	if err := g.loadStreamConsumersAndPolicies(svc); err != nil {
		return err
	}
	if err := g.loadVideoStreams(kinesisvideo.NewFromConfig(config)); err != nil {
		return err
	}
	if err := g.loadAnalyticsApplications(kinesisanalytics.NewFromConfig(config)); err != nil {
		return err
	}
	return g.loadAnalyticsV2Applications(kinesisanalyticsv2.NewFromConfig(config))
}

func (g *KinesisGenerator) streamARNs(svc *kinesis.Client) ([]string, error) {
	var arns []string
	var results *kinesis.ListStreamsOutput
	request := kinesis.ListStreamsInput{}
	for results == nil || aws.ToBool(results.HasMoreStreams) {
		var err error
		results, err = svc.ListStreams(context.TODO(), &request)
		if err != nil {
			return nil, err
		}
		for _, name := range results.StreamNames {
			summary, err := svc.DescribeStreamSummary(context.TODO(), &kinesis.DescribeStreamSummaryInput{StreamName: aws.String(name)})
			if err != nil {
				return nil, err
			}
			arns = append(arns, StringValue(summary.StreamDescriptionSummary.StreamARN))
			request = kinesis.ListStreamsInput{ExclusiveStartStreamName: aws.String(name)}
		}
		if len(results.StreamNames) == 0 {
			break
		}
	}
	return arns, nil
}

func (g *KinesisGenerator) loadStreamConsumersAndPolicies(svc *kinesis.Client) error {
	arns, err := g.streamARNs(svc)
	if err != nil {
		return err
	}
	for _, arn := range arns {
		input := &kinesis.ListStreamConsumersInput{StreamARN: aws.String(arn)}
		for {
			out, err := svc.ListStreamConsumers(context.TODO(), input)
			if err != nil {
				return err
			}
			for _, consumer := range out.Consumers {
				g.Resources = append(g.Resources, terraformutils.NewResource(
					StringValue(consumer.ConsumerARN),
					StringValue(consumer.ConsumerName),
					"aws_kinesis_stream_consumer", "aws",
					map[string]string{"stream_arn": arn},
					kinesisAllowEmptyValues, map[string]interface{}{}))
			}
			if out.NextToken == nil {
				break
			}
			input.NextToken = out.NextToken
		}
		// A stream only has a resource policy when one was attached; the
		// Get call fails for streams without one, which is not an error
		// for enumeration purposes.
		policy, err := svc.GetResourcePolicy(context.TODO(), &kinesis.GetResourcePolicyInput{ResourceARN: aws.String(arn)})
		if err == nil && policy.Policy != nil && *policy.Policy != "" {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				arn, arn, "aws_kinesis_resource_policy", "aws", kinesisAllowEmptyValues))
		}
	}
	return nil
}

func (g *KinesisGenerator) loadVideoStreams(svc *kinesisvideo.Client) error {
	input := &kinesisvideo.ListStreamsInput{}
	for {
		out, err := svc.ListStreams(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, stream := range out.StreamInfoList {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(stream.StreamARN), StringValue(stream.StreamName),
				"aws_kinesis_video_stream", "aws", kinesisAllowEmptyValues))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}

func (g *KinesisGenerator) loadAnalyticsApplications(svc *kinesisanalytics.Client) error {
	input := &kinesisanalytics.ListApplicationsInput{}
	for {
		out, err := svc.ListApplications(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, app := range out.ApplicationSummaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(app.ApplicationARN), StringValue(app.ApplicationName),
				"aws_kinesis_analytics_application", "aws", kinesisAllowEmptyValues))
		}
		if !aws.ToBool(out.HasMoreApplications) {
			return nil
		}
	}
}

func (g *KinesisGenerator) loadAnalyticsV2Applications(svc *kinesisanalyticsv2.Client) error {
	input := &kinesisanalyticsv2.ListApplicationsInput{}
	for {
		out, err := svc.ListApplications(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, app := range out.ApplicationSummaries {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				StringValue(app.ApplicationARN), StringValue(app.ApplicationName),
				"aws_kinesisanalyticsv2_application", "aws", kinesisAllowEmptyValues))
		}
		if out.NextToken == nil {
			return nil
		}
		input.NextToken = out.NextToken
	}
}
