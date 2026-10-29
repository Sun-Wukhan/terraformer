// Copyright 2018 The Terraformer Authors.
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
	"os"
	"strconv"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/pkg/errors"
	"github.com/zclconf/go-cty/cty"
)

type AWSProvider struct { //nolint
	terraformutils.Provider
	region  string
	profile string
}

const GlobalRegion = "aws-global"
const MainRegionPublicPartition = "us-east-1"
const NoRegion = ""

// SupportedGlobalResources should be bound to a default region. AWS doesn't specify in which region default services are
// placed (see  https://docs.aws.amazon.com/general/latest/gr/rande.html), so we shouldn't assume any region as well
var SupportedGlobalResources = []string{
	"budgets",
	"cloudfront",
	"ecrpublic",
	"iam",
	"organization",
	"route53",
	"waf",
}

// SupportedEastOnlyResources should be bound to us-east-1 region only, and does not work in any other region.
var SupportedEastOnlyResources = []string{
	"wafv2_cloudfront",
}

func (p AWSProvider) GetResourceConnections() map[string]map[string][]string {
	return map[string]map[string][]string{
		"alb": {
			"sg":     []string{"security_groups", "id"},
			"subnet": []string{"subnets", "id"},
			"alb": []string{
				"load_balancer_arn", "id",
				"listener_arn", "id",
				// TF ALB TG attachment logic doesn't work well with references (doesn't interpolate)
			},
		},
		"auto_scaling": {
			"sg":     []string{"security_groups", "id"},
			"subnet": []string{"vpc_zone_identifier", "id"},
		},
		"ec2_instance": {
			"sg":     []string{"vpc_security_group_ids", "id"},
			"subnet": []string{"subnet_id", "id"},
			"ebs":    []string{"ebs_block_device", "id"},
		},
		"elasticache": {
			"vpc":    []string{"vpc_id", "id"},
			"subnet": []string{"subnet_ids", "id"},
			"sg":     []string{"security_group_ids", "id"},
		},
		"ebs": {
			// TF EBS attachment logic doesn't work well with references (doesn't interpolate)
		},
		"ecs": {
			// ECS is not able anymore to support references (doesn't interpolate)
			"subnet": []string{"network_configuration.subnets", "id"},
			"sg":     []string{"network_configuration.security_groups", "id"},
		},
		"eks": {
			"subnet": []string{"vpc_config.subnet_ids", "id"},
			"sg":     []string{"vpc_config.security_group_ids", "id"},
		},
		"elb": {
			"sg":     []string{"security_groups", "id"},
			"subnet": []string{"subnets", "id"},
		},
		"igw": {"vpc": []string{"vpc_id", "id"}},
		"identitystore": {
			"identitystore": []string{
				"group_id", "id",
				"member_id", "id",
			},
		},
		"msk": {
			"subnet": []string{"broker_node_group_info.client_subnets", "id"},
			"sg":     []string{"broker_node_group_info.security_groups", "id"},
		},
		"nacl": {
			"subnet": []string{"subnet_ids", "id"},
			"vpc":    []string{"vpc_id", "id"},
		},
		"organization": {
			"organization": []string{
				"policy_id", "id",
				"parent_id", "id",
				"target_id", "id",
			},
		},
		"rds": {
			"subnet": []string{"subnet_ids", "id"},
			"sg":     []string{"vpc_security_group_ids", "id"},
		},
		"route_table": {
			"route_table": []string{"route_table_id", "id"},
			"subnet":      []string{"subnet_id", "id"},
			"vpc":         []string{"vpc_id", "id"},
		},
		"sns": {
			"sns": []string{"topic_arn", "id"},
			"sqs": []string{"endpoint", "arn"},
		},
		"sg": {
			"sg": []string{
				"egress.security_groups", "id",
				"ingress.security_groups", "id",
				"security_group_id", "id",
				"source_security_group_id", "id",
			},
		},
		"subnet": {"vpc": []string{"vpc_id", "id"}},
		"transit_gateway": {
			"vpc":             []string{"vpc_id", "id"},
			"transit_gateway": []string{"transit_gateway_id", "id"},
			"subnet":          []string{"subnet_ids", "id"},
			"vpn_connection":  []string{"vpn_connection_id", "id"},
		},
		"vpn_gateway": {"vpc": []string{"vpc_id", "id"}},
		"vpn_connection": {
			"customer_gateway": []string{"customer_gateway_id", "id"},
			"vpn_gateway":      []string{"vpn_gateway_id", "id"},
		},
	}
}

func (p AWSProvider) GetProviderData(arg ...string) map[string]interface{} {
	awsConfig := map[string]interface{}{}

	if p.region == GlobalRegion {
		awsConfig["region"] = MainRegionPublicPartition // For TF to workaround terraform-providers/terraform-provider-aws#1043
	} else if p.region != NoRegion {
		awsConfig["region"] = p.region
	}

	return map[string]interface{}{
		"provider": map[string]interface{}{
			"aws": awsConfig,
		},
	}
}

func (p *AWSProvider) GetConfig() cty.Value {
	if p.region != GlobalRegion {
		return cty.ObjectVal(map[string]cty.Value{
			"region":                 cty.StringVal(p.region),
			"skip_region_validation": cty.True,
		})
	}
	return cty.ObjectVal(map[string]cty.Value{
		"region":                 cty.StringVal(""),
		"skip_region_validation": cty.True,
	})
}

func (p *AWSProvider) GetBasicConfig() cty.Value {
	return p.GetConfig()
}

// check projectName in env params
func (p *AWSProvider) Init(args []string) error {
	p.region = args[0]
	p.profile = args[1]

	// Terraformer accepts region and profile configuration, so we must detect what env variables to adjust to make Go SDK rely on them. AWS_SDK_LOAD_CONFIG here must be checked to determine correct variable to set.
	enableSharedConfig, _ := strconv.ParseBool(os.Getenv("AWS_SDK_LOAD_CONFIG"))
	var err error
	if p.region != GlobalRegion && p.region != NoRegion {
		if enableSharedConfig {
			err = os.Setenv("AWS_DEFAULT_REGION", p.region)
		} else {
			err = os.Setenv("AWS_REGION", p.region)
		}
		if err != nil {
			return err
		}
	}

	if p.profile != "" && p.profile != "default" {
		envVar := "AWS_PROFILE"
		if enableSharedConfig {
			envVar = "AWS_DEFAULT_PROFILE"
		}

		if err := os.Setenv(envVar, p.profile); err != nil {
			return err
		}
	}
	return nil
}

func (p *AWSProvider) GetName() string {
	return "aws"
}

func (p *AWSProvider) InitService(serviceName string, verbose bool) error {
	var isSupported bool
	if _, isSupported = p.GetSupportedService()[serviceName]; !isSupported {
		return errors.New("aws: " + serviceName + " not supported service")
	}
	p.Service = p.GetSupportedService()[serviceName]
	p.Service.SetName(serviceName)
	p.Service.SetVerbose(verbose)
	p.Service.SetProviderName(p.GetName())
	p.Service.SetArgs(map[string]interface{}{
		"region":                 p.region,
		"profile":                p.profile,
		"skip_region_validation": true,
	})
	return nil
}

// GetAWSSupportService return map of support service for AWS
func (p *AWSProvider) GetSupportedService() map[string]terraformutils.ServiceGenerator {
	return map[string]terraformutils.ServiceGenerator{
		"accessanalyzer":               &AwsFacade{service: &AccessAnalyzerGenerator{}},
		"acm":                          &AwsFacade{service: &ACMGenerator{}},
		"alb":                          &AwsFacade{service: &AlbGenerator{}},
		"api_gateway":                  &AwsFacade{service: &APIGatewayGenerator{}},
		"api_gatewayv2":                &AwsFacade{service: &APIGatewayV2Generator{}},
		"appsync":                      &AwsFacade{service: &AppSyncGenerator{}},
		"auto_scaling":                 &AwsFacade{service: &AutoScalingGenerator{}},
		"batch":                        &AwsFacade{service: &BatchGenerator{}},
		"bedrock":                      &AwsFacade{service: &BedrockGenerator{}},
		"bedrockagent":                 &AwsFacade{service: &BedrockAgentGenerator{}},
		"bedrockagentcore":             &AwsFacade{service: &BedrockAgentCoreGenerator{}},
		"budgets":                      &AwsFacade{service: &BudgetsGenerator{}},
		"cloud9":                       &AwsFacade{service: &Cloud9Generator{}},
		"cloudformation":               &AwsFacade{service: &CloudFormationGenerator{}},
		"cloudfront":                   &AwsFacade{service: &CloudFrontGenerator{}},
		"cloudhsm":                     &AwsFacade{service: &CloudHsmGenerator{}},
		"cloudtrail":                   &AwsFacade{service: &CloudTrailGenerator{}},
		"cloudwatch":                   &AwsFacade{service: &CloudWatchGenerator{}},
		"codebuild":                    &AwsFacade{service: &CodeBuildGenerator{}},
		"codecommit":                   &AwsFacade{service: &CodeCommitGenerator{}},
		"codedeploy":                   &AwsFacade{service: &CodeDeployGenerator{}},
		"codepipeline":                 &AwsFacade{service: &CodePipelineGenerator{}},
		"cognito":                      &AwsFacade{service: &CognitoGenerator{}},
		"config":                       &AwsFacade{service: &ConfigGenerator{}},
		"connect":                      &AwsFacade{service: &ConnectGenerator{}},
		"customer_gateway":             &AwsFacade{service: &CustomerGatewayGenerator{}},
		"datapipeline":                 &AwsFacade{service: &DataPipelineGenerator{}},
		"devicefarm":                   &AwsFacade{service: &DeviceFarmGenerator{}},
		"docdb":                        &AwsFacade{service: &DocDBGenerator{}},
		"dx":                           &AwsFacade{service: &DirectConnectGenerator{}},
		"dynamodb":                     &AwsFacade{service: &DynamoDbGenerator{}},
		"ebs":                          &AwsFacade{service: &EbsGenerator{}},
		"ec2_instance":                 &AwsFacade{service: &Ec2Generator{}},
		"ecr":                          &AwsFacade{service: &EcrGenerator{}},
		"ecrpublic":                    &AwsFacade{service: &EcrPublicGenerator{}},
		"ecs":                          &AwsFacade{service: &EcsGenerator{}},
		"efs":                          &AwsFacade{service: &EfsGenerator{}},
		"eks":                          &AwsFacade{service: &EksGenerator{}},
		"eip":                          &AwsFacade{service: &ElasticIPGenerator{}},
		"elasticache":                  &AwsFacade{service: &ElastiCacheGenerator{}},
		"elastic_beanstalk":            &AwsFacade{service: &BeanstalkGenerator{}},
		"elb":                          &AwsFacade{service: &ElbGenerator{}},
		"emr":                          &AwsFacade{service: &EmrGenerator{}},
		"eni":                          &AwsFacade{service: &EniGenerator{}},
		"es":                           &AwsFacade{service: &EsGenerator{}},
		"firehose":                     &AwsFacade{service: &FirehoseGenerator{}},
		"glue":                         &AwsFacade{service: &GlueGenerator{}},
		"iam":                          &AwsFacade{service: &IamGenerator{}},
		"identitystore":                &AwsFacade{service: &IdentityStoreGenerator{}},
		"igw":                          &AwsFacade{service: &IgwGenerator{}},
		"iot":                          &AwsFacade{service: &IotGenerator{}},
		"kinesis":                      &AwsFacade{service: &KinesisGenerator{}},
		"kms":                          &AwsFacade{service: &KmsGenerator{}},
		"lambda":                       &AwsFacade{service: &LambdaGenerator{}},
		"lightsail":                    &AwsFacade{service: &LightsailGenerator{}},
		"logs":                         &AwsFacade{service: &LogsGenerator{}},
		"media_package":                &AwsFacade{service: &MediaPackageGenerator{}},
		"media_store":                  &AwsFacade{service: &MediaStoreGenerator{}},
		"medialive":                    &AwsFacade{service: &MediaLiveGenerator{}},
		"mq":                           &AwsFacade{service: &MQGenerator{}},
		"msk":                          &AwsFacade{service: &MskGenerator{}},
		"nacl":                         &AwsFacade{service: &NaclGenerator{}},
		"nat":                          &AwsFacade{service: &NatGatewayGenerator{}},
		"networkmanager":               &AwsFacade{service: &NetworkManagerGenerator{}},
		"opsworks":                     &AwsFacade{service: &OpsworksGenerator{}},
		"organization":                 &AwsFacade{service: &OrganizationGenerator{}},
		"qbusiness":                    &AwsFacade{service: &QBusinessGenerator{}},
		"qldb":                         &AwsFacade{service: &QLDBGenerator{}},
		"quicksight":                   &AwsFacade{service: &QuickSightGenerator{}},
		"rds":                          &AwsFacade{service: &RDSGenerator{}},
		"redshift":                     &AwsFacade{service: &RedshiftGenerator{}},
		"resourcegroups":               &AwsFacade{service: &ResourceGroupsGenerator{}},
		"route53":                      &AwsFacade{service: &Route53Generator{}},
		"route_table":                  &AwsFacade{service: &RouteTableGenerator{}},
		"s3":                           &AwsFacade{service: &S3Generator{}},
		"sagemaker":                    &AwsFacade{service: &SageMakerGenerator{}},
		"secretsmanager":               &AwsFacade{service: &SecretsManagerGenerator{}},
		"securityhub":                  &AwsFacade{service: &SecurityhubGenerator{}},
		"servicecatalog":               &AwsFacade{service: &ServiceCatalogGenerator{}},
		"ses":                          &AwsFacade{service: &SesGenerator{}},
		"sfn":                          &AwsFacade{service: &SfnGenerator{}},
		"sg":                           &AwsFacade{service: &SecurityGenerator{}},
		"sqs":                          &AwsFacade{service: &SqsGenerator{}},
		"sns":                          &AwsFacade{service: &SnsGenerator{}},
		"ssm":                          &AwsFacade{service: &SsmGenerator{}},
		"ssoadmin":                     &AwsFacade{service: &SSOAdminGenerator{}},
		"subnet":                       &AwsFacade{service: &SubnetGenerator{}},
		"swf":                          &AwsFacade{service: &SWFGenerator{}},
		"transit_gateway":              &AwsFacade{service: &TransitGatewayGenerator{}},
		"waf":                          &AwsFacade{service: &WafGenerator{}},
		"waf_regional":                 &AwsFacade{service: &WafRegionalGenerator{}},
		"wafv2_cloudfront":             &AwsFacade{service: NewWafv2CloudfrontGenerator()},
		"wafv2_regional":               &AwsFacade{service: NewWafv2RegionalGenerator()},
		"vpc":                          &AwsFacade{service: &VpcGenerator{}},
		"vpc_endpoint":                 &AwsFacade{service: &VpcEndpointGenerator{}},
		"vpc_peering":                  &AwsFacade{service: &VpcPeeringConnectionGenerator{}},
		"vpclattice":                   &AwsFacade{service: &VPCLatticeGenerator{}},
		"vpn_connection":               &AwsFacade{service: &VpnConnectionGenerator{}},
		"vpn_gateway":                  &AwsFacade{service: &VpnGatewayGenerator{}},
		"workspaces":                   &AwsFacade{service: &WorkspacesGenerator{}},
		"workspacesweb":                &AwsFacade{service: &WorkspacesWebGenerator{}},
		"xray":                         &AwsFacade{service: &XrayGenerator{}},
		"account":                      &AwsFacade{service: &AccountGenerator{}},
		"accountaccess":                &AwsFacade{service: &AccountAccessGenerator{}},
		"acmpca":                       &AwsFacade{service: &AcmpcaGenerator{}},
		"amplify":                      &AwsFacade{service: &AmplifyGenerator{}},
		"appautoscaling":               &AwsFacade{service: &AppAutoScalingGenerator{}},
		"appconfig":                    &AwsFacade{service: &AppConfigGenerator{}},
		"appfabric":                    &AwsFacade{service: &AppFabricGenerator{}},
		"appflow":                      &AwsFacade{service: &AppFlowGenerator{}},
		"appintegrations":              &AwsFacade{service: &AppIntegrationsGenerator{}},
		"applicationinsights":          &AwsFacade{service: &ApplicationInsightsGenerator{}},
		"appmesh":                      &AwsFacade{service: &AppMeshGenerator{}},
		"apprunner":                    &AwsFacade{service: &AppRunnerGenerator{}},
		"appstream":                    &AwsFacade{service: &AppStreamGenerator{}},
		"arcregionswitch":              &AwsFacade{service: &ArcRegionSwitchGenerator{}},
		"arczonalshift":                &AwsFacade{service: &ArcZonalShiftGenerator{}},
		"athena":                       &AwsFacade{service: &AthenaGenerator{}},
		"auditmanager":                 &AwsFacade{service: &AuditManagerGenerator{}},
		"backup":                       &AwsFacade{service: &BackupGenerator{}},
		"bcmdataexports":               &AwsFacade{service: &BcmDataExportsGenerator{}},
		"billing":                      &AwsFacade{service: &BillingGenerator{}},
		"chatbot":                      &AwsFacade{service: &ChatbotGenerator{}},
		"chime":                        &AwsFacade{service: &ChimeGenerator{}},
		"chimesdkmediapipelines":       &AwsFacade{service: &ChimeSdkMediaPipelinesGenerator{}},
		"cleanrooms":                   &AwsFacade{service: &CleanRoomsGenerator{}},
		"client_vpn":                   &AwsFacade{service: &ClientVpnGenerator{}},
		"cloudfrontkeyvaluestore":      &AwsFacade{service: &CloudFrontKeyValueStoreGenerator{}},
		"cloudsearch":                  &AwsFacade{service: &CloudSearchGenerator{}},
		"cloudwatchomni":               &AwsFacade{service: &CloudWatchOmniGenerator{}},
		"codeartifact":                 &AwsFacade{service: &CodeArtifactGenerator{}},
		"codeconnections":              &AwsFacade{service: &CodeConnectionsGenerator{}},
		"codeguruprofiler":             &AwsFacade{service: &CodeGuruProfilerGenerator{}},
		"codestarnotifications":        &AwsFacade{service: &CodeStarNotificationsGenerator{}},
		"comprehend":                   &AwsFacade{service: &ComprehendGenerator{}},
		"controltower":                 &AwsFacade{service: &ControlTowerGenerator{}},
		"costexplorer":                 &AwsFacade{service: &CostExplorerGenerator{}},
		"customerprofiles":             &AwsFacade{service: &CustomerProfilesGenerator{}},
		"dataexchange":                 &AwsFacade{service: &DataExchangeGenerator{}},
		"datasync":                     &AwsFacade{service: &DataSyncGenerator{}},
		"datazone":                     &AwsFacade{service: &DataZoneGenerator{}},
		"dax":                          &AwsFacade{service: &DAXGenerator{}},
		"detective":                    &AwsFacade{service: &DetectiveGenerator{}},
		"devopsagent":                  &AwsFacade{service: &DevOpsAgentGenerator{}},
		"devopsguru":                   &AwsFacade{service: &DevOpsGuruGenerator{}},
		"directoryservice":             &AwsFacade{service: &DirectoryServiceGenerator{}},
		"dlm":                          &AwsFacade{service: &DlmGenerator{}},
		"dms":                          &AwsFacade{service: &DMSGenerator{}},
		"drs":                          &AwsFacade{service: &DrsGenerator{}},
		"dsql":                         &AwsFacade{service: &DSQLGenerator{}},
		"elastictranscoder":            &AwsFacade{service: &ElasticTranscoderGenerator{}},
		"evidently":                    &AwsFacade{service: &EvidentlyGenerator{}},
		"finspace":                     &AwsFacade{service: &FinSpaceGenerator{}},
		"fis":                          &AwsFacade{service: &FisGenerator{}},
		"fms":                          &AwsFacade{service: &FmsGenerator{}},
		"fsx":                          &AwsFacade{service: &FsxGenerator{}},
		"gamelift":                     &AwsFacade{service: &GameLiftGenerator{}},
		"glacier":                      &AwsFacade{service: &GlacierGenerator{}},
		"globalaccelerator":            &AwsFacade{service: &GlobalAcceleratorGenerator{}},
		"grafana":                      &AwsFacade{service: &GrafanaGenerator{}},
		"guardduty":                    &AwsFacade{service: &GuardDutyGenerator{}},
		"imagebuilder":                 &AwsFacade{service: &ImageBuilderGenerator{}},
		"inspector":                    &AwsFacade{service: &InspectorGenerator{}},
		"inspector2":                   &AwsFacade{service: &Inspector2Generator{}},
		"internetmonitor":              &AwsFacade{service: &InternetMonitorGenerator{}},
		"invoicing":                    &AwsFacade{service: &InvoicingGenerator{}},
		"ivs":                          &AwsFacade{service: &IvsGenerator{}},
		"kendra":                       &AwsFacade{service: &KendraGenerator{}},
		"keyspaces":                    &AwsFacade{service: &KeyspacesGenerator{}},
		"lakeformation":                &AwsFacade{service: &LakeFormationGenerator{}},
		"lex":                          &AwsFacade{service: &LexGenerator{}},
		"licensemanager":               &AwsFacade{service: &LicenseManagerGenerator{}},
		"local_gateway":                &AwsFacade{service: &LocalGatewayGenerator{}},
		"location":                     &AwsFacade{service: &LocationGenerator{}},
		"m2":                           &AwsFacade{service: &M2Generator{}},
		"macie2":                       &AwsFacade{service: &Macie2Generator{}},
		"mailmanager":                  &AwsFacade{service: &MailManagerGenerator{}},
		"mediaconvert":                 &AwsFacade{service: &MediaConvertGenerator{}},
		"memorydb":                     &AwsFacade{service: &MemoryDBGenerator{}},
		"mwaa":                         &AwsFacade{service: &MwaaGenerator{}},
		"neptune":                      &AwsFacade{service: &NeptuneGenerator{}},
		"neptunegraph":                 &AwsFacade{service: &NeptuneGraphGenerator{}},
		"networkfirewall":              &AwsFacade{service: &NetworkFirewallGenerator{}},
		"networkflowmonitor":           &AwsFacade{service: &NetworkFlowMonitorGenerator{}},
		"networkmonitor":               &AwsFacade{service: &NetworkMonitorGenerator{}},
		"notifications":                &AwsFacade{service: &NotificationsGenerator{}},
		"notificationscontacts":        &AwsFacade{service: &NotificationsContactsGenerator{}},
		"oam":                          &AwsFacade{service: &OamGenerator{}},
		"observabilityadmin":           &AwsFacade{service: &ObservabilityAdminGenerator{}},
		"odb":                          &AwsFacade{service: &ODBGenerator{}},
		"opensearch":                   &AwsFacade{service: &OpenSearchGenerator{}},
		"osis":                         &AwsFacade{service: &OsisGenerator{}},
		"outposts":                     &AwsFacade{service: &OutpostsGenerator{}},
		"paymentcryptography":          &AwsFacade{service: &PaymentCryptographyGenerator{}},
		"pinpoint":                     &AwsFacade{service: &PinpointGenerator{}},
		"pinpointsmsvoicev2":           &AwsFacade{service: &PinpointSmsVoiceV2Generator{}},
		"pipes":                        &AwsFacade{service: &PipesGenerator{}},
		"pricingplanmanager":           &AwsFacade{service: &PricingPlanManagerGenerator{}},
		"prometheus":                   &AwsFacade{service: &PrometheusGenerator{}},
		"ram":                          &AwsFacade{service: &RamGenerator{}},
		"rbin":                         &AwsFacade{service: &RbinGenerator{}},
		"redshiftserverless":           &AwsFacade{service: &RedshiftServerlessGenerator{}},
		"rekognition":                  &AwsFacade{service: &RekognitionGenerator{}},
		"resiliencehub":                &AwsFacade{service: &ResilienceHubGenerator{}},
		"resiliencehubv2":              &AwsFacade{service: &ResilienceHubV2Generator{}},
		"resourceexplorer2":            &AwsFacade{service: &ResourceExplorer2Generator{}},
		"rolesanywhere":                &AwsFacade{service: &RolesAnywhereGenerator{}},
		"route53domains":               &AwsFacade{service: &Route53DomainsGenerator{}},
		"route53profiles":              &AwsFacade{service: &Route53ProfilesGenerator{}},
		"route53recoverycontrolconfig": &AwsFacade{service: &Route53RecoveryControlConfigGenerator{}},
		"route53recoveryreadiness":     &AwsFacade{service: &Route53RecoveryReadinessGenerator{}},
		"route53resolver":              &AwsFacade{service: &Route53ResolverGenerator{}},
		"rum":                          &AwsFacade{service: &RumGenerator{}},
		"s3control":                    &AwsFacade{service: &S3ControlGenerator{}},
		"s3files":                      &AwsFacade{service: &S3FilesGenerator{}},
		"s3outposts":                   &AwsFacade{service: &S3OutpostsGenerator{}},
		"s3tables":                     &AwsFacade{service: &S3TablesGenerator{}},
		"s3vectors":                    &AwsFacade{service: &S3VectorsGenerator{}},
		"savingsplans":                 &AwsFacade{service: &SavingsPlansGenerator{}},
		"scheduler":                    &AwsFacade{service: &SchedulerGenerator{}},
		"schemas":                      &AwsFacade{service: &SchemasGenerator{}},
		"securitylake":                 &AwsFacade{service: &SecurityLakeGenerator{}},
		"servicecatalogappregistry":    &AwsFacade{service: &ServiceCatalogAppRegistryGenerator{}},
		"servicediscovery":             &AwsFacade{service: &ServiceDiscoveryGenerator{}},
		"servicequotas":                &AwsFacade{service: &ServiceQuotasGenerator{}},
		"sesv2":                        &AwsFacade{service: &SesV2Generator{}},
		"shield":                       &AwsFacade{service: &ShieldGenerator{}},
		"signer":                       &AwsFacade{service: &SignerGenerator{}},
		"ssmcontacts":                  &AwsFacade{service: &SsmContactsGenerator{}},
		"ssmincidents":                 &AwsFacade{service: &SsmIncidentsGenerator{}},
		"ssmquicksetup":                &AwsFacade{service: &SsmQuickSetupGenerator{}},
		"storagegateway":               &AwsFacade{service: &StorageGatewayGenerator{}},
		"synthetics":                   &AwsFacade{service: &SyntheticsGenerator{}},
		"timestreaminfluxdb":           &AwsFacade{service: &TimestreamInfluxDBGenerator{}},
		"timestreamquery":              &AwsFacade{service: &TimestreamQueryGenerator{}},
		"timestreamwrite":              &AwsFacade{service: &TimestreamWriteGenerator{}},
		"transcribe":                   &AwsFacade{service: &TranscribeGenerator{}},
		"transfer":                     &AwsFacade{service: &TransferGenerator{}},
		"verifiedaccess":               &AwsFacade{service: &VerifiedAccessGenerator{}},
		"verifiedpermissions":          &AwsFacade{service: &VerifiedPermissionsGenerator{}},
		"workmail":                     &AwsFacade{service: &WorkMailGenerator{}},
	}
}

func StringValue(value *string) string {
	if value != nil {
		return *value
	}
	return ""
}
