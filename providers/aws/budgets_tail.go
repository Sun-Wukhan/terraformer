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
	"fmt"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/budgets"
)

// loadBudgetActions adds aws_budgets_budget_action resources.
// Provider import ID (verified): "<account ID>:<action ID>:<budget name>".
func (g *BudgetsGenerator) loadBudgetActions(svc *budgets.Client, account *string) {
	out, err := svc.DescribeBudgets(context.TODO(), &budgets.DescribeBudgetsInput{AccountId: account})
	if err != nil {
		return
	}
	for _, budget := range out.Budgets {
		actions, err := svc.DescribeBudgetActionsForBudget(context.TODO(),
			&budgets.DescribeBudgetActionsForBudgetInput{AccountId: account, BudgetName: budget.BudgetName})
		if err != nil {
			continue
		}
		for _, action := range actions.Actions {
			id := fmt.Sprintf("%s:%s:%s", StringValue(account), StringValue(action.ActionId), StringValue(budget.BudgetName))
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				id, id, "aws_budgets_budget_action", "aws", []string{}))
		}
	}
}
