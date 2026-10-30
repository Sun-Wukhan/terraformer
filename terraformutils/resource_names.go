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

package terraformutils

import (
	"fmt"
	"log"
	"sort"
)

// DedupeResourceNames makes every (resource type, resource name) pair unique
// across all imported services. Child resources frequently share a name under
// different parents (e.g. a "MDE.Windows" extension on every VM, or a
// "$Default" consumer group in every event hub); without this the HCL writer
// silently keeps one of them while the state file lists all of them, which
// Terraform rejects as "Duplicate resource instance in state".
//
// The first occurrence keeps its name; later ones get a "-2", "-3", ... suffix.
// Services are visited in sorted order so the result is deterministic.
func DedupeResourceNames(importResources map[string][]Resource) {
	services := make([]string, 0, len(importResources))
	for service := range importResources {
		services = append(services, service)
	}
	sort.Strings(services)

	used := map[string]map[string]bool{} // type -> name -> taken
	for _, service := range services {
		resources := importResources[service]
		for i := range resources {
			r := &resources[i]
			if r.InstanceInfo == nil {
				continue
			}
			typ := r.InstanceInfo.Type
			if used[typ] == nil {
				used[typ] = map[string]bool{}
			}
			name := r.ResourceName
			for n := 2; used[typ][name]; n++ {
				name = fmt.Sprintf("%s-%d", r.ResourceName, n)
			}
			if name != r.ResourceName {
				log.Printf("WARN: renaming duplicate %s.%s to %s.%s", typ, r.ResourceName, typ, name)
				r.ResourceName = name
			}
			used[typ][name] = true
		}
	}
}
