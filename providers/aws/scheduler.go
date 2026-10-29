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

// Amazon EventBridge Scheduler generator (service key "scheduler").

import (
	"context"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/scheduler"
)

var schedulerAllowEmptyValues = []string{"tags."}

type SchedulerGenerator struct {
	AWSService
}

func (g *SchedulerGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	svc := scheduler.NewFromConfig(config)
	var nextToken *string
	for {
		out, err := svc.ListScheduleGroups(context.TODO(), &scheduler.ListScheduleGroupsInput{NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, group := range out.ScheduleGroups {
			groupName := StringValue(group.Name)
			// aws_scheduler_schedule_group — importer ID is the group name.
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				groupName, groupName, "aws_scheduler_schedule_group", "aws", schedulerAllowEmptyValues))
			if err := g.createSchedules(svc, groupName); err != nil {
				return err
			}
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}

// aws_scheduler_schedule — importer ID is "group_name/name".
func (g *SchedulerGenerator) createSchedules(svc *scheduler.Client, groupName string) error {
	var nextToken *string
	for {
		out, err := svc.ListSchedules(context.TODO(), &scheduler.ListSchedulesInput{GroupName: &groupName, NextToken: nextToken})
		if err != nil {
			return err
		}
		for _, schedule := range out.Schedules {
			scheduleName := StringValue(schedule.Name)
			id := groupName + "/" + scheduleName
			g.Resources = append(g.Resources, terraformutils.NewResource(
				id, scheduleName, "aws_scheduler_schedule", "aws",
				map[string]string{"group_name": groupName, "name": scheduleName},
				schedulerAllowEmptyValues, map[string]interface{}{}))
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}
	return nil
}
