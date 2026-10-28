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
	"fmt"
	"log"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"google.golang.org/api/cloudtasks/v2"
	"google.golang.org/api/compute/v1"
)

// Cloud Tasks queue IAM (ops sweep), extending CloudTaskGenerator:
// for every queue in the configured region, the combined
// google_cloud_tasks_queue_iam resource is emitted when the queue
// carries IAM bindings (the provider models queue IAM through that
// single combined resource). Enumeration uses the generated
// cloudtasks v2 REST client; the queue resources themselves come
// from the GAPIC pass in cloudtasks.go.
func (g *CloudTaskGenerator) loadCloudTaskQueueIam(ctx context.Context, service *cloudtasks.Service) error {
	project := g.GetArgs()["project"].(string)
	region := g.GetArgs()["region"].(compute.Region).Name
	parent := fmt.Sprintf("projects/%s/locations/%s", project, region)
	return service.Projects.Locations.Queues.List(parent).Pages(ctx, func(page *cloudtasks.ListQueuesResponse) error {
		for _, queue := range page.Queues {
			policy, err := service.Projects.Locations.Queues.GetIamPolicy(queue.Name, &cloudtasks.GetIamPolicyRequest{}).Context(ctx).Do()
			if err != nil {
				log.Println(err)
				continue
			}
			if len(policy.Bindings) == 0 {
				continue
			}
			queueName := opsLastSegment(queue.Name)
			g.Resources = append(g.Resources, terraformutils.NewResource(
				queue.Name,
				queueName,
				"google_cloud_tasks_queue_iam",
				g.ProviderName,
				map[string]string{
					"name":     queueName,
					"project":  project,
					"location": region,
				},
				cloudTasksAllowEmptyValues,
				cloudTasksAdditionalFields,
			))
		}
		return nil
	})
}
