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

package azure

import (
	"context"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// MSSQLGenerator covers the azurerm_mssql_* depth types via armsql. The
// older azurerm_sql_* equivalents of some of these live in database.go;
// the mssql_* provider resources share the same ARM IDs.
//
// Skipped in this family:
//   - azurerm_mssql_job_schedule: a job schedule is embedded in the parent
//     job (its provider importer ID is the job ID itself) and armsql has no
//     JobSchedulesClient or List; there is no separate ARM resource.
//   - azurerm_mssql_database_vulnerability_assessment_rule_baseline:
//     armsql DatabaseVulnerabilityAssessmentRuleBaselinesClient is Get-only
//     (per rule + baseline name), no List pager.
//   - azurerm_mssql_managed_instance_start_stop_schedule: armsql v1.2.0 has
//     no StartStopSchedulesClient at all; the schedule is Get-only by the
//     fixed name "default" where the API exists.
type MSSQLGenerator struct {
	AzureService
}

func (g *MSSQLGenerator) appendSimple(id, name, resourceType string) {
	g.Resources = append(g.Resources, terraformutils.NewSimpleResource(id, name, resourceType, g.ProviderName, []string{}))
}

// listServerChildren enumerates every per-server child type in the sweep.
func (g *MSSQLGenerator) listServerChildren(ctx context.Context, resourceGroup, serverName string) error {
	elasticPoolsClient, err := armsql.NewElasticPoolsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	pager := elasticPoolsClient.NewListByServerPager(resourceGroup, serverName, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			g.appendSimple(*item.ID, *item.Name, "azurerm_mssql_elasticpool")
		}
	}

	failoverGroupsClient, err := armsql.NewFailoverGroupsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	failoverPager := failoverGroupsClient.NewListByServerPager(resourceGroup, serverName, nil)
	for failoverPager.More() {
		page, err := failoverPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			g.appendSimple(*item.ID, *item.Name, "azurerm_mssql_failover_group")
		}
	}

	outboundClient, err := armsql.NewOutboundFirewallRulesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	outboundPager := outboundClient.NewListByServerPager(resourceGroup, serverName, nil)
	for outboundPager.More() {
		page, err := outboundPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			g.appendSimple(*item.ID, *item.Name, "azurerm_mssql_outbound_firewall_rule")
		}
	}

	dnsAliasesClient, err := armsql.NewServerDNSAliasesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	dnsPager := dnsAliasesClient.NewListByServerPager(resourceGroup, serverName, nil)
	for dnsPager.More() {
		page, err := dnsPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			g.appendSimple(*item.ID, *item.Name, "azurerm_mssql_server_dns_alias")
		}
	}

	virtualNetworkRulesClient, err := armsql.NewVirtualNetworkRulesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	vnetPager := virtualNetworkRulesClient.NewListByServerPager(resourceGroup, serverName, nil)
	for vnetPager.More() {
		page, err := vnetPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			g.appendSimple(*item.ID, *item.Name, "azurerm_mssql_virtual_network_rule")
		}
	}

	// Singleton-per-server policies: each of these List endpoints returns
	// the single "default"/"current" policy, whose ID is exactly the
	// provider importer ID.
	extendedAuditingClient, err := armsql.NewExtendedServerBlobAuditingPoliciesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	extendedAuditingPager := extendedAuditingClient.NewListByServerPager(resourceGroup, serverName, nil)
	for extendedAuditingPager.More() {
		page, err := extendedAuditingPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			g.appendSimple(*item.ID, *item.Name+"-"+serverName, "azurerm_mssql_server_extended_auditing_policy")
		}
	}

	devOpsAuditingClient, err := armsql.NewServerDevOpsAuditSettingsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	devOpsPager := devOpsAuditingClient.NewListByServerPager(resourceGroup, serverName, nil)
	for devOpsPager.More() {
		page, err := devOpsPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			g.appendSimple(*item.ID, *item.Name+"-"+serverName, "azurerm_mssql_server_microsoft_support_auditing_policy")
		}
	}

	securityAlertClient, err := armsql.NewServerSecurityAlertPoliciesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	securityAlertPager := securityAlertClient.NewListByServerPager(resourceGroup, serverName, nil)
	for securityAlertPager.More() {
		page, err := securityAlertPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			g.appendSimple(*item.ID, *item.Name+"-"+serverName, "azurerm_mssql_server_security_alert_policy")
		}
	}

	encryptionProtectorsClient, err := armsql.NewEncryptionProtectorsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	encryptionPager := encryptionProtectorsClient.NewListByServerPager(resourceGroup, serverName, nil)
	for encryptionPager.More() {
		page, err := encryptionPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			g.appendSimple(*item.ID, *item.Name+"-"+serverName, "azurerm_mssql_server_transparent_data_encryption")
		}
	}

	vulnerabilityClient, err := armsql.NewServerVulnerabilityAssessmentsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	vulnerabilityPager := vulnerabilityClient.NewListByServerPager(resourceGroup, serverName, nil)
	for vulnerabilityPager.More() {
		page, err := vulnerabilityPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			g.appendSimple(*item.ID, *item.Name+"-"+serverName, "azurerm_mssql_server_vulnerability_assessment")
		}
	}

	return nil
}

// listDatabaseChildren enumerates the per-database extended auditing
// policies (the List endpoint returns the single "default" policy).
func (g *MSSQLGenerator) listDatabaseChildren(ctx context.Context, resourceGroup, serverName string) error {
	databasesClient, err := armsql.NewDatabasesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	extendedAuditingClient, err := armsql.NewExtendedDatabaseBlobAuditingPoliciesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	pager := databasesClient.NewListByServerPager(resourceGroup, serverName, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, database := range page.Value {
			auditingPager := extendedAuditingClient.NewListByDatabasePager(resourceGroup, serverName, *database.Name, nil)
			for auditingPager.More() {
				auditingPage, err := auditingPager.NextPage(ctx)
				if err != nil {
					log.Println(err)
					return err
				}
				for _, item := range auditingPage.Value {
					g.appendSimple(*item.ID, *item.Name+"-"+*database.Name, "azurerm_mssql_database_extended_auditing_policy")
				}
			}
		}
	}
	return nil
}

// listJobChildren enumerates job agents and, per agent, jobs (with their
// steps), credentials and target groups.
func (g *MSSQLGenerator) listJobChildren(ctx context.Context, resourceGroup, serverName string) error {
	jobAgentsClient, err := armsql.NewJobAgentsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	jobsClient, err := armsql.NewJobsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	jobStepsClient, err := armsql.NewJobStepsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	jobCredentialsClient, err := armsql.NewJobCredentialsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	jobTargetGroupsClient, err := armsql.NewJobTargetGroupsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}

	pager := jobAgentsClient.NewListByServerPager(resourceGroup, serverName, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, agent := range page.Value {
			g.appendSimple(*agent.ID, *agent.Name, "azurerm_mssql_job_agent")

			jobsPager := jobsClient.NewListByAgentPager(resourceGroup, serverName, *agent.Name, nil)
			for jobsPager.More() {
				jobsPage, err := jobsPager.NextPage(ctx)
				if err != nil {
					log.Println(err)
					return err
				}
				for _, job := range jobsPage.Value {
					g.appendSimple(*job.ID, *job.Name, "azurerm_mssql_job")
					stepsPager := jobStepsClient.NewListByJobPager(resourceGroup, serverName, *agent.Name, *job.Name, nil)
					for stepsPager.More() {
						stepsPage, err := stepsPager.NextPage(ctx)
						if err != nil {
							log.Println(err)
							return err
						}
						for _, step := range stepsPage.Value {
							g.appendSimple(*step.ID, *step.Name, "azurerm_mssql_job_step")
						}
					}
				}
			}

			credentialsPager := jobCredentialsClient.NewListByAgentPager(resourceGroup, serverName, *agent.Name, nil)
			for credentialsPager.More() {
				credentialsPage, err := credentialsPager.NextPage(ctx)
				if err != nil {
					log.Println(err)
					return err
				}
				for _, credential := range credentialsPage.Value {
					g.appendSimple(*credential.ID, *credential.Name, "azurerm_mssql_job_credential")
				}
			}

			targetGroupsPager := jobTargetGroupsClient.NewListByAgentPager(resourceGroup, serverName, *agent.Name, nil)
			for targetGroupsPager.More() {
				targetGroupsPage, err := targetGroupsPager.NextPage(ctx)
				if err != nil {
					log.Println(err)
					return err
				}
				for _, targetGroup := range targetGroupsPage.Value {
					g.appendSimple(*targetGroup.ID, *targetGroup.Name, "azurerm_mssql_job_target_group")
				}
			}
		}
	}
	return nil
}

func (g *MSSQLGenerator) listServers() error {
	ctx := context.Background()
	client, err := armsql.NewServersClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	process := func(servers []*armsql.Server) error {
		for _, server := range servers {
			id, err := ParseAzureResourceID(*server.ID)
			if err != nil {
				return err
			}
			if err := g.listServerChildren(ctx, id.ResourceGroup, *server.Name); err != nil {
				return err
			}
			if err := g.listDatabaseChildren(ctx, id.ResourceGroup, *server.Name); err != nil {
				return err
			}
			if err := g.listJobChildren(ctx, id.ResourceGroup, *server.Name); err != nil {
				return err
			}
		}
		return nil
	}
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			if err := process(page.Value); err != nil {
				return err
			}
		}
	} else {
		pager := client.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			if err := process(page.Value); err != nil {
				return err
			}
		}
	}
	return nil
}

// listManagedInstanceChildren enumerates every per-instance child type.
func (g *MSSQLGenerator) listManagedInstanceChildren(ctx context.Context, resourceGroup, instanceName, location string, includeFailoverGroups bool) error {
	databasesClient, err := armsql.NewManagedDatabasesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	pager := databasesClient.NewListByInstancePager(resourceGroup, instanceName, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			g.appendSimple(*item.ID, *item.Name, "azurerm_mssql_managed_database")
		}
	}

	administratorsClient, err := armsql.NewManagedInstanceAdministratorsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	administratorsPager := administratorsClient.NewListByInstancePager(resourceGroup, instanceName, nil)
	for administratorsPager.More() {
		page, err := administratorsPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			g.appendSimple(*item.ID, *item.Name+"-"+instanceName, "azurerm_mssql_managed_instance_active_directory_administrator")
		}
	}

	securityAlertClient, err := armsql.NewManagedServerSecurityAlertPoliciesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	securityAlertPager := securityAlertClient.NewListByInstancePager(resourceGroup, instanceName, nil)
	for securityAlertPager.More() {
		page, err := securityAlertPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			g.appendSimple(*item.ID, *item.Name+"-"+instanceName, "azurerm_mssql_managed_instance_security_alert_policy")
		}
	}

	encryptionClient, err := armsql.NewManagedInstanceEncryptionProtectorsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	encryptionPager := encryptionClient.NewListByInstancePager(resourceGroup, instanceName, nil)
	for encryptionPager.More() {
		page, err := encryptionPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			g.appendSimple(*item.ID, *item.Name+"-"+instanceName, "azurerm_mssql_managed_instance_transparent_data_encryption")
		}
	}

	vulnerabilityClient, err := armsql.NewManagedInstanceVulnerabilityAssessmentsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	vulnerabilityPager := vulnerabilityClient.NewListByInstancePager(resourceGroup, instanceName, nil)
	for vulnerabilityPager.More() {
		page, err := vulnerabilityPager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, item := range page.Value {
			g.appendSimple(*item.ID, *item.Name+"-"+instanceName, "azurerm_mssql_managed_instance_vulnerability_assessment")
		}
	}

	// Instance failover groups are scoped by (resource group, location),
	// not by instance; the caller enumerates them once per location.
	if includeFailoverGroups && location != "" {
		failoverClient, err := armsql.NewInstanceFailoverGroupsClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
		if err != nil {
			return err
		}
		failoverPager := failoverClient.NewListByLocationPager(resourceGroup, location, nil)
		for failoverPager.More() {
			page, err := failoverPager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			for _, item := range page.Value {
				g.appendSimple(*item.ID, *item.Name, "azurerm_mssql_managed_instance_failover_group")
			}
		}
	}

	return nil
}

func (g *MSSQLGenerator) listManagedInstances() error {
	ctx := context.Background()
	client, err := armsql.NewManagedInstancesClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return err
	}
	seenFailoverLocations := map[string]bool{}
	process := func(instances []*armsql.ManagedInstance) error {
		for _, instance := range instances {
			g.appendSimple(*instance.ID, *instance.Name, "azurerm_mssql_managed_instance")
			id, err := ParseAzureResourceID(*instance.ID)
			if err != nil {
				return err
			}
			location := ""
			if instance.Location != nil {
				location = *instance.Location
			}
			// Failover groups are location-scoped, not instance-scoped:
			// enumerate them once per (resource group, location).
			failoverKey := id.ResourceGroup + "/" + location
			includeFailover := !seenFailoverLocations[failoverKey]
			seenFailoverLocations[failoverKey] = true
			if err := g.listManagedInstanceChildren(ctx, id.ResourceGroup, *instance.Name, location, includeFailover); err != nil {
				return err
			}
		}
		return nil
	}
	if rg := g.getResourceGroup(); rg != "" {
		pager := client.NewListByResourceGroupPager(rg, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			if err := process(page.Value); err != nil {
				return err
			}
		}
	} else {
		pager := client.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				log.Println(err)
				return err
			}
			if err := process(page.Value); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *MSSQLGenerator) InitResources() error {
	if err := g.listServers(); err != nil {
		return err
	}
	return g.listManagedInstances()
}
