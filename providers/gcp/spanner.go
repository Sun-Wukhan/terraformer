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

package gcp

import (
	"context"
	"log"
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	spannerv1 "google.golang.org/api/spanner/v1"
)

var spannerAllowEmptyValues = []string{""}

var spannerAdditionalFields = map[string]interface{}{}

type SpannerGenerator struct {
	GCPService
}

// InitResources enumerates Cloud Spanner resources for the configured
// project: instance configs, instances with their IAM trio, databases
// per instance with their IAM trio, backup schedules per database and
// instance partitions per instance. Provider IDs are the full API
// resource names, the canonical import format for all of these.
func (g *SpannerGenerator) InitResources() error {
	ctx := context.Background()
	svc, err := spannerv1.NewService(ctx)
	if err != nil {
		return err
	}
	return g.initResources(ctx, svc)
}

func (g *SpannerGenerator) initResources(ctx context.Context, svc *spannerv1.Service) error {
	project := g.GetArgs()["project"].(string)
	projectParent := "projects/" + project
	baseAttrs := map[string]string{"project": project}
	err := svc.Projects.InstanceConfigs.List(projectParent).Pages(ctx, func(page *spannerv1.ListInstanceConfigsResponse) error {
		for _, config := range page.InstanceConfigs {
			g.Resources = append(g.Resources, g.newSpannerResource(config.Name, "google_spanner_instance_config", baseAttrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	err = svc.Projects.Instances.List(projectParent).Pages(ctx, func(page *spannerv1.ListInstancesResponse) error {
		for _, instance := range page.Instances {
			g.Resources = append(g.Resources, g.newSpannerResource(instance.Name, "google_spanner_instance", baseAttrs))
			g.loadSpannerIam(func() ([]iamBinding, error) {
				policy, err := svc.Projects.Instances.GetIamPolicy(instance.Name, &spannerv1.GetIamPolicyRequest{}).Context(ctx).Do()
				if err != nil {
					return nil, err
				}
				return spannerPolicyBindings(policy), nil
			}, instance.Name, "google_spanner_instance", "instance", baseAttrs)
			g.loadSpannerDatabases(ctx, svc, instance.Name, baseAttrs)
			g.loadInstancePartitions(ctx, svc, instance.Name, baseAttrs)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	return nil
}

func spannerPolicyBindings(policy *spannerv1.Policy) []iamBinding {
	bindings := []iamBinding{}
	for _, binding := range policy.Bindings {
		bindings = append(bindings, iamBinding{Role: binding.Role, Members: binding.Members})
	}
	return bindings
}

func (g *SpannerGenerator) newSpannerResource(id, resourceType string, attrs map[string]string) terraformutils.Resource {
	if attrs == nil {
		attrs = map[string]string{}
	}
	parts := strings.Split(id, "/")
	return terraformutils.NewResource(
		id,
		parts[len(parts)-1],
		resourceType,
		g.ProviderName,
		attrs,
		spannerAllowEmptyValues,
		spannerAdditionalFields,
	)
}

func (g *SpannerGenerator) loadSpannerIam(getBindings func() ([]iamBinding, error), name, resourcePrefix, parentKey string, baseAttrs map[string]string) {
	bindings, err := getBindings()
	if err != nil {
		log.Println(err)
		return
	}
	parentAttrs := map[string]string{parentKey: name}
	for k, v := range baseAttrs {
		parentAttrs[k] = v
	}
	g.Resources = append(g.Resources, expandIamPolicy(
		bindings,
		name,
		resourcePrefix+"_iam_policy",
		resourcePrefix+"_iam_binding",
		resourcePrefix+"_iam_member",
		parentAttrs,
		g.newSpannerResource,
	)...)
}

func (g *SpannerGenerator) loadSpannerDatabases(ctx context.Context, svc *spannerv1.Service, instanceName string, baseAttrs map[string]string) {
	err := svc.Projects.Instances.Databases.List(instanceName).Pages(ctx, func(page *spannerv1.ListDatabasesResponse) error {
		for _, database := range page.Databases {
			g.Resources = append(g.Resources, g.newSpannerResource(database.Name, "google_spanner_database", baseAttrs))
			g.loadSpannerIam(func() ([]iamBinding, error) {
				policy, err := svc.Projects.Instances.Databases.GetIamPolicy(database.Name, &spannerv1.GetIamPolicyRequest{}).Context(ctx).Do()
				if err != nil {
					return nil, err
				}
				return spannerPolicyBindings(policy), nil
			}, database.Name, "google_spanner_database", "database", baseAttrs)
			schedErr := svc.Projects.Instances.Databases.BackupSchedules.List(database.Name).Pages(ctx, func(page *spannerv1.ListBackupSchedulesResponse) error {
				for _, schedule := range page.BackupSchedules {
					g.Resources = append(g.Resources, g.newSpannerResource(schedule.Name, "google_spanner_backup_schedule", baseAttrs))
				}
				return nil
			})
			if schedErr != nil {
				log.Println(schedErr)
			}
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (g *SpannerGenerator) loadInstancePartitions(ctx context.Context, svc *spannerv1.Service, instanceName string, baseAttrs map[string]string) {
	err := svc.Projects.Instances.InstancePartitions.List(instanceName).Pages(ctx, func(page *spannerv1.ListInstancePartitionsResponse) error {
		for _, partition := range page.InstancePartitions {
			g.Resources = append(g.Resources, g.newSpannerResource(partition.Name, "google_spanner_instance_partition", baseAttrs))
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}
