// Copyright 2019 The Terraformer Authors.
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
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/go-autorest/autorest"
	"github.com/hashicorp/go-azure-helpers/authentication"
	"github.com/hashicorp/go-azure-helpers/sender"
	"github.com/manicminer/hamilton/environments"
	"github.com/zclconf/go-cty/cty"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

type AzureProvider struct { //nolint
	terraformutils.Provider
	config        authentication.Config
	authorizer    autorest.Authorizer
	credential    azcore.TokenCredential
	resourceGroup string
}

func (p *AzureProvider) setEnvConfig() error {
	subscriptionID := os.Getenv("ARM_SUBSCRIPTION_ID")
	if subscriptionID == "" {
		return errors.New("set ARM_SUBSCRIPTION_ID env var")
	}
	var auxTenants []string
	if v := os.Getenv("ARM_AUXILIARY_TENANT_IDS"); v != "" {
		auxTenants = strings.Split(v, ";")
		if len(auxTenants) > 3 {
			return fmt.Errorf("the provider only supports 3 auxiliary tenant IDs for ARM_AUXILIARY_TENANT_IDS")
		}
	}
	builder := &authentication.Builder{
		ClientID:            os.Getenv("ARM_CLIENT_ID"),
		SubscriptionID:      subscriptionID,
		TenantID:            os.Getenv("ARM_TENANT_ID"),
		AuxiliaryTenantIDs:  auxTenants,
		Environment:         os.Getenv("ARM_ENVIRONMENT"),
		MetadataHost:        os.Getenv("ARM_METADATA_HOSTNAME"),
		MsiEndpoint:         os.Getenv("ARM_MSI_ENDPOINT"),
		ClientSecret:        os.Getenv("ARM_CLIENT_SECRET"),
		ClientCertPath:      os.Getenv("ARM_CLIENT_CERTIFICATE_PATH"),
		ClientCertPassword:  os.Getenv("ARM_CLIENT_CERTIFICATE_PASSWORD"),
		IDTokenRequestToken: os.Getenv("ARM_OIDC_REQUEST_TOKEN"),
		IDTokenRequestURL:   os.Getenv("ARM_OIDC_REQUEST_URL"),

		// Feature Toggles
		SupportsAzureCliToken:          true,
		SupportsClientSecretAuth:       true,
		SupportsClientCertAuth:         true,
		SupportsManagedServiceIdentity: os.Getenv("ARM_USE_MSI") != "",
		SupportsOIDCAuth:               os.Getenv("ARM_USE_OIDC") != "",
		UseMicrosoftGraph:              os.Getenv("ARM_USE_ADAL") == "",
	}

	if builder.Environment == "" {
		builder.Environment = "public"
	}
	config, err := builder.Build()
	if err != nil {
		return nil
	}
	p.config = *config

	return nil
}

func (p *AzureProvider) getAuthorizer() (autorest.Authorizer, error) {
	env, err := authentication.DetermineEnvironment(p.config.Environment)
	if err != nil {
		return nil, err
	}
	p.config.CustomResourceManagerEndpoint = env.ResourceManagerEndpoint
	oauthConfig, err := p.config.BuildOAuthConfig(env.ActiveDirectoryEndpoint)
	if err != nil {
		return nil, err
	}
	if oauthConfig == nil {
		return nil, fmt.Errorf("unable to configure OAuthConfig for tenant %s", p.config.TenantID)
	}
	sender := sender.BuildSender("terraformer")
	ctx := context.Background()
	var auth autorest.Authorizer

	if p.config.UseMicrosoftGraph {
		hamiltonEnv, ero := environments.EnvironmentFromString(p.config.Environment)
		if ero != nil {
			return nil, ero
		}
		auth, err = p.config.GetMSALToken(ctx, hamiltonEnv.ResourceManager, sender, oauthConfig, env.TokenAudience)
	} else {
		// Deprecated
		auth, err = p.config.GetADALToken(ctx, sender, oauthConfig, env.ResourceManagerEndpoint)
	}
	if err != nil {
		return nil, err
	}
	return auth, nil
}

func (p *AzureProvider) Init(args []string) error {
	err := p.setEnvConfig()
	if err != nil {
		return err
	}

	authorizer, err := p.getAuthorizer()
	if err != nil {
		return err
	}
	p.authorizer = authorizer

	credential, err := p.getCredential()
	if err != nil {
		return err
	}
	p.credential = credential

	p.resourceGroup = args[0]

	return nil
}

// getCredential builds the track-2 token credential used by migrated services.
// azidentity's DefaultAzureCredential understands AZURE_* env vars, so the
// ARM_* vars terraformer has always used are mapped across. This preserves
// the service-principal, client-certificate, managed-identity and Azure CLI
// flows; the GitHub-OIDC (ARM_OIDC_*) flow has no azidentity equivalent and
// is not carried over.
func (p *AzureProvider) getCredential() (azcore.TokenCredential, error) {
	if v := os.Getenv("ARM_CLIENT_ID"); v != "" {
		_ = os.Setenv("AZURE_CLIENT_ID", v)
	}
	if v := os.Getenv("ARM_TENANT_ID"); v != "" {
		_ = os.Setenv("AZURE_TENANT_ID", v)
	}
	if v := os.Getenv("ARM_CLIENT_SECRET"); v != "" {
		_ = os.Setenv("AZURE_CLIENT_SECRET", v)
	}
	if v := os.Getenv("ARM_CLIENT_CERTIFICATE_PATH"); v != "" {
		_ = os.Setenv("AZURE_CLIENT_CERTIFICATE_PATH", v)
	}
	if v := os.Getenv("ARM_CLIENT_CERTIFICATE_PASSWORD"); v != "" {
		_ = os.Setenv("AZURE_CLIENT_CERTIFICATE_PASSWORD", v)
	}
	cred, err := azidentity.NewDefaultAzureCredential(&azidentity.DefaultAzureCredentialOptions{
		ClientOptions: azcore.ClientOptions{
			Cloud: azureCloudConfig(p.config.CustomResourceManagerEndpoint),
		},
	})
	if err != nil {
		return nil, err
	}
	return cred, nil
}

// GetConfig returns the provider block used for the CLI refresh. azurerm
// rejects any configuration without a `features {}` block, and v4+ also
// requires an explicit subscription_id.
func (p *AzureProvider) GetConfig() cty.Value {
	config := map[string]cty.Value{
		"features": cty.EmptyObjectVal,
	}
	if p.config.SubscriptionID != "" {
		config["subscription_id"] = cty.StringVal(p.config.SubscriptionID)
	}
	return cty.ObjectVal(config)
}

func (p *AzureProvider) GetName() string {
	return "azurerm"
}

// GetProviderData returns the generated provider block. azurerm >= 2.0
// requires a `features {}` block; the version constraint lives in
// required_providers.
func (p *AzureProvider) GetProviderData(arg ...string) map[string]interface{} {
	return map[string]interface{}{
		"provider": map[string]interface{}{
			"azurerm": map[string]interface{}{
				"features": map[string]interface{}{},
			},
		},
	}
}

func (AzureProvider) GetResourceConnections() map[string]map[string][]string {
	return map[string]map[string][]string{
		"analysis": {
			"resource_group": []string{"resource_group_name", "name"},
		},
		"automation": {
			"resource_group": []string{"resource_group_name", "name"},
		},
		"datadog": {
			"resource_group": []string{"resource_group_name", "name"},
		},
		"dynatrace": {
			"resource_group": []string{"resource_group_name", "name"},
		},
		"machine_learning": {
			"resource_group": []string{"resource_group_name", "name"},
		},
		"new_relic": {
			"resource_group": []string{"resource_group_name", "name"},
		},
		"nginx": {
			"resource_group": []string{"resource_group_name", "name"},
		},
		"oracle": {
			"resource_group": []string{"resource_group_name", "name"},
		},
		"palo_alto": {
			"resource_group": []string{"resource_group_name", "name"},
		},
		"app_service": {
			"resource_group": []string{"resource_group_name", "name"},
		},
		"application_gateway": {
			"resource_group": []string{"resource_group_name", "name"},
		},
		"cosmosdb": {
			"resource_group": []string{
				"resource_group_name", "name",
				"location", "location",
			},
		},
		"container": {
			"resource_group": []string{
				"resource_group_name", "name",
				"location", "location",
			},
		},
		"database": {
			"resource_group": []string{
				"resource_group_name", "name",
				"location", "location",
			},
		},
		"databricks": {
			"resource_group": []string{
				"resource_group_name", "name",
				"managed_resource_group_name", "name",
				"location", "location",
			},
			"storage_account": []string{"storage_account_name", "name"},
			"subnet": []string{
				"public_subnet_name", "name",
				"private_subnet_name", "name",
			},
			"virtual_network": []string{"virtual_network_id", "id"},
		},
		"data_factory": {
			"resource_group": []string{
				"resource_group_name", "name",
				"location", "location",
			},
			"data_factory": []string{
				"data_factory_name", "name",
				"data_factory_id", "id",
				"linked_service_name", "name",
				"integration_runtime_name", "name",
			},
			"databricks":      []string{"existing_cluster_id", "id"},
			"keyvault":        []string{"keyvault_id", "id"},
			"storage_account": []string{"storage_account_id", "id"},
		},
		"disk": {
			"resource_group": []string{"resource_group_name", "name"},
		},
		"dns": {
			"resource_group": []string{"resource_group_name", "name"},
		},
		"eventhub": {
			"resource_group": []string{"resource_group_name", "name"},
			"eventhub": []string{
				"eventhub_name", "name",
				"namespace_name", "name",
			},
		},
		"keyvault": {
			"resource_group": []string{
				"resource_group_name", "name",
				"location", "location",
			},
		},
		"load_balancer": {
			"resource_group": []string{"resource_group_name", "name"},
		},
		"network_interface": {
			"resource_group": []string{
				"resource_group_name", "name",
				"location", "location",
			},
			"subnet": []string{"subnet_id", "id"},
		},
		"network_security_group": {
			"resource_group": []string{
				"resource_group_name", "name",
				"location", "location",
			},
			"network_security_group": []string{"network_security_group_name", "name"},
		},
		"network_watcher": {
			"resource_group": []string{
				"resource_group_name", "name",
				"location", "location",
			},
			"network_watcher": []string{"network_watcher_name", "name"},
			"storage_account": []string{"storage_account_id", "id"},
		},
		"private_dns": {
			"resource_group":  []string{"resource_group_name", "name"},
			"virtual_network": []string{"virtual_network_id", "id"},
			"private_dns": []string{
				"zone_name", "name",
				"private_dns_zone_name", "name",
			},
		},
		"private_endpoint": {
			"resource_group": []string{
				"resource_group_name", "name",
				"location", "location",
			},
			"subnet": []string{"subnet_id", "id"},
		},
		"public_ip": {
			"resource_group": []string{
				"resource_group_name", "name",
				"location", "location",
			},
		},
		"purview": {
			"resource_group": []string{"resource_group_name", "name"},
		},
		"redis": {
			"resource_group": []string{"resource_group_name", "name"},
		},
		"route_table": {
			"resource_group": []string{
				"resource_group_name", "name",
				"location", "location",
			},
			"route_table": []string{"route_table_name", "name"},
		},
		"scaleset": {
			"resource_group": []string{"resource_group_name", "name"},
		},
		"ssh_public_key": {
			"resource_group": []string{
				"resource_group_name", "name",
				"location", "location",
			},
		},
		"storage_account": {
			"resource_group": []string{
				"resource_group_name", "name",
				"location", "location",
			},
			"virtual_network": []string{"virtual_network_subnet_ids", "id"},
		},
		"storage_blob": {
			"storage_account":   []string{"storage_account_name", "name"},
			"storage_container": []string{"storage_container_name", "name"},
		},
		"storage_container": {
			"storage_account": []string{"storage_account_name", "name"},
		},
		"synapse": {
			"resource_group": []string{
				"resource_group_name", "name",
				"managed_resource_group_name", "name",
			},
			"synapse": []string{"synapse_workspace_id", "id"},
		},
		"subnet": {
			"resource_group":         []string{"resource_group_name", "name"},
			"virtual_network":        []string{"virtual_network_name", "name"},
			"network_security_group": []string{"network_security_group_id", "id"},
			"route_table":            []string{"route_table_id", "id"},
			"subnet":                 []string{"subnet_id", "id"},
		},
		"virtual_machine": {
			"resource_group": []string{
				"resource_group_name", "name",
				"location", "location",
			},
			"network_interface": []string{"network_interface_ids", "id"},
		},
		"virtual_network": {
			"resource_group": []string{"resource_group_name", "name"},
		},
	}
}

func (p *AzureProvider) GetSupportedService() map[string]terraformutils.ServiceGenerator {
	return map[string]terraformutils.ServiceGenerator{
		"ai_foundry":                           &AIFoundryGenerator{},
		"analysis":                             &AnalysisGenerator{},
		"api_management":                       &APIManagementGenerator{},
		"app_configuration":                    &AppConfigurationGenerator{},
		"app_service":                          &AppServiceGenerator{},
		"automation":                           &AutomationGenerator{},
		"datadog":                              &DatadogGenerator{},
		"dynatrace":                            &DynatraceGenerator{},
		"machine_learning":                     &MachineLearningGenerator{},
		"new_relic":                            &NewRelicGenerator{},
		"nginx":                                &NginxGenerator{},
		"oracle":                               &OracleGenerator{},
		"palo_alto":                            &PaloAltoGenerator{},
		"arc":                                  &ArcGenerator{},
		"automanage":                           &AutomanageGenerator{},
		"batch":                                &BatchGenerator{},
		"compute":                              &ComputeGenerator{},
		"dev_center":                           &DevCenterGenerator{},
		"dev_test":                             &DevTestGenerator{},
		"maintenance":                          &MaintenanceGenerator{},
		"managed_compute":                      &ManagedComputeGenerator{},
		"scvmm":                                &SCVMMGenerator{},
		"service_fabric":                       &ServiceFabricGenerator{},
		"stack_hci":                            &StackHCIGenerator{},
		"virtual_desktop":                      &VirtualDesktopGenerator{},
		"virtual_machine_child":                &VirtualMachineChildGenerator{},
		"virtual_machine_scale_set":            &ModernScaleSetGenerator{},
		"vmware":                               &VMwareGenerator{},
		"application_insights":                 &ApplicationInsightsGenerator{},
		"application_load_balancer":            &ApplicationLoadBalancerGenerator{},
		"application_security_group":           &ApplicationSecurityGroupGenerator{},
		"attestation":                          &AttestationGenerator{},
		"logic_app":                            &LogicAppGenerator{},
		"web_application_firewall_policy":      &WebApplicationFirewallPolicyGenerator{},
		"web_apps":                             &WebAppsGenerator{},
		"application_gateway":                  &ApplicationGatewayGenerator{},
		"cognitive":                            &CognitiveGenerator{},
		"communication":                        &CommunicationGenerator{},
		"confidential_ledger":                  &ConfidentialLedgerGenerator{},
		"cdn":                                  &CDNGenerator{},
		"express_route":                        &ExpressRouteGenerator{},
		"firewall":                             &FirewallGenerator{},
		"frontdoor":                            &FrontDoorGenerator{},
		"nat_gateway":                          &NatGatewayGenerator{},
		"network_basics":                       &NetworkBasicsGenerator{},
		"network_function":                     &NetworkFunctionGenerator{},
		"network_gateway":                      &NetworkGatewayGenerator{},
		"network_manager":                      &NetworkManagerGenerator{},
		"private_dns_resolver":                 &PrivateDNSResolverGenerator{},
		"traffic_manager":                      &TrafficManagerGenerator{},
		"virtual_wan":                          &VirtualWANGenerator{},
		"cosmosdb":                             &CosmosDBGenerator{},
		"cosmosdb_postgresql":                  &CosmosPostgreSQLGenerator{},
		"bot":                                  &BotGenerator{},
		"container":                            &ContainerGenerator{},
		"container_app":                        &ContainerAppGenerator{},
		"cost_management":                      &CostManagementGenerator{},
		"database":                             &DatabasesGenerator{},
		"database_flexible":                    &FlexibleServersGenerator{},
		"database_migration":                   &DatabaseMigrationGenerator{},
		"databricks":                           &DatabricksGenerator{},
		"data_factory":                         &DataFactoryGenerator{},
		"dedicated_hsm":                        &DedicatedHSMGenerator{},
		"data_protection":                      &DataProtectionGenerator{},
		"data_share":                           &DataShareGenerator{},
		"elastic":                              &ElasticGenerator{},
		"kusto":                                &KustoGenerator{},
		"mongo_cluster":                        &MongoClusterGenerator{},
		"mssql":                                &MSSQLGenerator{},
		"mssql_virtual_machine":                &MSSQLVirtualMachineGenerator{},
		"disk":                                 &DiskGenerator{},
		"dns":                                  &DNSGenerator{},
		"eventgrid":                            &EventGridGenerator{},
		"domain_services":                      &DomainServicesGenerator{},
		"eventhub":                             &EventHubGenerator{},
		"federated_identity_credential":        &FederatedIdentityCredentialGenerator{},
		"iot_security":                         &IoTSecurityGenerator{},
		"iotcentral":                           &IoTCentralGenerator{},
		"iothub":                               &IoTHubGenerator{},
		"iothub_device_update":                 &IoTDeviceUpdateGenerator{},
		"fluid_relay":                          &FluidRelayGenerator{},
		"graph_services":                       &GraphServicesGenerator{},
		"hdinsight":                            &HDInsightGenerator{},
		"keyvault":                             &KeyVaultGenerator{},
		"kubernetes":                           &KubernetesGenerator{},
		"kubernetes_fleet":                     &KubernetesFleetGenerator{},
		"lighthouse":                           &LighthouseGenerator{},
		"load_balancer":                        &LoadBalancerGenerator{},
		"management_group":                     &ManagementGroupGenerator{},
		"management_lock":                      &ManagementLockGenerator{},
		"maps":                                 &MapsGenerator{},
		"network_interface":                    &NetworkInterfaceGenerator{},
		"network_security_group":               &NetworkSecurityGroupGenerator{},
		"network_watcher":                      &NetworkWatcherGenerator{},
		"notification_hub":                     &NotificationHubGenerator{},
		"private_dns":                          &PrivateDNSGenerator{},
		"platform_misc":                        &PlatformMiscGenerator{},
		"policy":                               &PolicyGenerator{},
		"private_endpoint":                     &PrivateEndpointGenerator{},
		"public_ip":                            &PublicIPGenerator{},
		"purview":                              &PurviewGenerator{},
		"rbac":                                 &RBACGenerator{},
		"redis":                                &RedisGenerator{},
		"relay":                                &RelayGenerator{},
		"resource_group":                       &ResourceGroupGenerator{},
		"route_table":                          &RouteTableGenerator{},
		"scaleset":                             &ScaleSetGenerator{},
		"security_center":                      &SecurityCenterGenerator{},
		"security_center_contact":              &SecurityCenterContactGenerator{},
		"security_center_subscription_pricing": &SecurityCenterSubscriptionPricingGenerator{},
		"sentinel":                             &SentinelGenerator{},
		"servicebus":                           &ServiceBusGenerator{},
		"search":                               &SearchGenerator{},
		"signalr":                              &SignalRGenerator{},
		"spring_cloud":                         &SpringCloudGenerator{},
		"ssh_public_key":                       &SSHPublicKeyGenerator{},
		"storage_account":                      &StorageAccountGenerator{},
		"storage_blob":                         &StorageBlobGenerator{},
		"storage_container":                    &StorageContainerGenerator{},
		"storage_extra":                        &StorageExtraGenerator{},
		"netapp":                               &NetAppGenerator{},
		"elastic_san":                          &ElasticSANGenerator{},
		"recovery_services":                    &RecoveryServicesGenerator{},
		"site_recovery":                        &SiteRecoveryGenerator{},
		"monitor":                              &MonitorGenerator{},
		"log_analytics":                        &LogAnalyticsGenerator{},
		"dashboard_grafana":                    &DashboardGrafanaGenerator{},
		"chaos_studio":                         &ChaosStudioGenerator{},
		"healthcare":                           &HealthcareGenerator{},
		"digital_twins":                        &DigitalTwinsGenerator{},
		"storage_mover":                        &StorageMoverGenerator{},
		"storage_sync":                         &StorageSyncGenerator{},
		"storage_discovery":                    &StorageDiscoveryGenerator{},
		"stream_analytics":                     &StreamAnalyticsGenerator{},
		"synapse":                              &SynapseGenerator{},
		"subnet":                               &SubnetGenerator{},
		"template_deployment":                  &TemplateDeploymentGenerator{},
		"virtual_machine":                      &VirtualMachineGenerator{},
		"virtual_network":                      &VirtualNetworkGenerator{},
		"web_pubsub":                           &WebPubSubGenerator{},
	}
}

func (p *AzureProvider) InitService(serviceName string, verbose bool) error {
	var isSupported bool
	if _, isSupported = p.GetSupportedService()[serviceName]; !isSupported {
		return errors.New("azurerm: " + serviceName + " not supported service")
	}
	p.Service = p.GetSupportedService()[serviceName]
	p.Service.SetName(serviceName)
	p.Service.SetVerbose(verbose)
	p.Service.SetProviderName(p.GetName())
	p.Service.SetArgs(map[string]interface{}{
		"config":         p.config,
		"authorizer":     p.authorizer,
		"credential":     p.credential,
		"resource_group": p.resourceGroup,
	})
	return nil
}
