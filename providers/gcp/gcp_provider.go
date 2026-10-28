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

package gcp

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"google.golang.org/api/compute/v1"
)

type GCPProvider struct { //nolint
	terraformutils.Provider
	projectName  string
	region       compute.Region
	providerType string
}

func GetRegions(project string) []string {
	computeService, err := compute.NewService(context.Background())
	if err != nil {
		return []string{}
	}
	regionsList, err := computeService.Regions.List(project).Do()
	if err != nil {
		return []string{}
	}
	regions := []string{}
	for _, region := range regionsList.Items {
		regions = append(regions, region.Name)
	}
	return regions
}

func getRegion(project, regionName string) *compute.Region {
	if regionName == "global" {
		return &compute.Region{}
	}
	computeService, err := compute.NewService(context.Background())
	if err != nil {
		log.Println(err)
		return &compute.Region{}
	}
	regionsGetCall := computeService.Regions.Get(project, regionName).Fields("name", "zones")
	region, err := regionsGetCall.Do()
	if err != nil {
		log.Println(err)
		return &compute.Region{}
	}
	return region
}

// check projectName in env params
func (p *GCPProvider) Init(args []string) error {
	projectName := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if len(args) > 1 {
		projectName = args[1]
	}
	if projectName == "" {
		return errors.New("google cloud project name must be set")
	}
	p.projectName = projectName
	p.region = *getRegion(projectName, args[0])
	p.providerType = args[2]
	return nil
}

func (p *GCPProvider) GetName() string {
	if p.providerType != "" {
		return "google-" + p.providerType
	}
	return "google"
}

func (p *GCPProvider) InitService(serviceName string, verbose bool) error {
	var isSupported bool
	if _, isSupported = p.GetSupportedService()[serviceName]; !isSupported {
		return errors.New("gcp: " + serviceName + " not supported service")
	}
	p.Service = p.GetSupportedService()[serviceName]
	p.Service.SetName(serviceName)
	p.Service.SetVerbose(verbose)
	p.Service.SetProviderName(p.GetName())
	p.Service.SetArgs(map[string]interface{}{
		"region":  p.region,
		"project": p.projectName,
	})
	return nil
}

// GetGCPSupportService return map of support service for GCP
func (p *GCPProvider) GetSupportedService() map[string]terraformutils.ServiceGenerator {
	services := ComputeServices
	services["bigQuery"] = &GCPFacade{service: &BigQueryGenerator{}}
	services["cloudFunctions"] = &GCPFacade{service: &CloudFunctionsGenerator{}}
	services["cloudsql"] = &GCPFacade{service: &CloudSQLGenerator{}}
	services["cloudtasks"] = &GCPFacade{service: &CloudTaskGenerator{}}
	services["dataProc"] = &GCPFacade{service: &DataprocGenerator{}}
	services["dns"] = &GCPFacade{service: &CloudDNSGenerator{}}
	services["gcs"] = &GCPFacade{service: &GcsGenerator{}}
	services["gke"] = &GCPFacade{service: &GkeGenerator{}}
	services["iam"] = &GCPFacade{service: &IamGenerator{}}
	services["kms"] = &GCPFacade{service: &KmsGenerator{}}
	services["logging"] = &GCPFacade{service: &LoggingGenerator{}}
	services["memoryStore"] = &GCPFacade{service: &MemoryStoreGenerator{}}
	services["monitoring"] = &GCPFacade{service: &MonitoringGenerator{}}
	services["project"] = &GCPFacade{service: &ProjectGenerator{}}
	services["instances"] = &GCPFacade{service: &InstancesGenerator{}}
	services["pubsub"] = &GCPFacade{service: &PubsubGenerator{}}
	services["schedulerJobs"] = &GCPFacade{service: &SchedulerJobsGenerator{}}
	services["cloudbuild"] = &GCPFacade{service: &CloudBuildGenerator{}}
	services["vertexAi"] = &GCPFacade{service: &VertexAiGenerator{}}
	services["apigee"] = &GCPFacade{service: &ApigeeGenerator{}}
	services["networkSecurity"] = &GCPFacade{service: &NetworkSecurityGenerator{}}
	services["networkServices"] = &GCPFacade{service: &NetworkServicesGenerator{}}
	services["networkConnectivity"] = &GCPFacade{service: &NetworkConnectivityGenerator{}}
	services["cloudRun"] = &GCPFacade{service: &CloudRunGenerator{}}
	services["scc"] = &GCPFacade{service: &SccGenerator{}}
	services["dataplex"] = &GCPFacade{service: &DataplexGenerator{}}
	services["chronicle"] = &GCPFacade{service: &ChronicleGenerator{}}
	services["firebase"] = &GCPFacade{service: &FirebaseGenerator{}}
	services["gemini"] = &GCPFacade{service: &GeminiGenerator{}}
	services["accessContextManager"] = &GCPFacade{service: &AccessContextManagerGenerator{}}
	services["discoveryEngine"] = &GCPFacade{service: &DiscoveryEngineGenerator{}}
	services["routerSubresources"] = &GCPFacade{service: &RouterSubresourcesGenerator{}}
	services["computeDepth"] = &GCPFacade{service: &ComputeDepthGenerator{}}
	services["computePolicies"] = &GCPFacade{service: &ComputePoliciesGenerator{}}
	services["computeSubresources"] = &GCPFacade{service: &ComputeSubresourcesGenerator{}}
	services["computeIam"] = &GCPFacade{service: &ComputeIamGenerator{}}
	services["bigtable"] = &GCPFacade{service: &BigtableGenerator{}}
	services["spanner"] = &GCPFacade{service: &SpannerGenerator{}}
	services["firestore"] = &GCPFacade{service: &FirestoreGenerator{}}
	services["alloydb"] = &GCPFacade{service: &AlloydbGenerator{}}
	services["memorystore"] = &GCPFacade{service: &MemorystoreGenerator{}}
	services["memcache"] = &GCPFacade{service: &MemcacheGenerator{}}
	services["dataprocMetastore"] = &GCPFacade{service: &DataprocMetastoreGenerator{}}
	services["dataCatalog"] = &GCPFacade{service: &DataCatalogGenerator{}}
	services["dataform"] = &GCPFacade{service: &DataformGenerator{}}
	services["dataFusion"] = &GCPFacade{service: &DataFusionGenerator{}}
	services["dataflow"] = &GCPFacade{service: &DataflowGenerator{}}
	services["datastream"] = &GCPFacade{service: &DatastreamGenerator{}}
	services["migrationCenter"] = &GCPFacade{service: &MigrationCenterGenerator{}}
	services["databaseMigrationService"] = &GCPFacade{service: &DatabaseMigrationServiceGenerator{}}
	services["dataPipeline"] = &GCPFacade{service: &DataPipelineGenerator{}}
	services["dataLineage"] = &GCPFacade{service: &DataLineageGenerator{}}
	services["iamOrg"] = &GCPFacade{service: &IamOrgGenerator{}}
	services["billing"] = &GCPFacade{service: &BillingGenerator{}}
	services["iap"] = &GCPFacade{service: &IapGenerator{}}
	services["orgPlatform"] = &GCPFacade{service: &OrgPlatformGenerator{}}
	services["monitoringDepth"] = &GCPFacade{service: &MonitoringDepthGenerator{}}
	services["observability"] = &GCPFacade{service: &ObservabilityGenerator{}}
	services["osConfig"] = &GCPFacade{service: &OsConfigGenerator{}}
	services["eventarc"] = &GCPFacade{service: &EventarcGenerator{}}
	services["pubsubDepth"] = &GCPFacade{service: &PubsubDepthGenerator{}}
	services["workflows"] = &GCPFacade{service: &WorkflowsGenerator{}}
	services["composer"] = &GCPFacade{service: &ComposerGenerator{}}
	services["cloudbuildDepth"] = &GCPFacade{service: &CloudBuildDepthGenerator{}}
	services["sourceRepo"] = &GCPFacade{service: &SourceRepoGenerator{}}
	services["runtimeConfig"] = &GCPFacade{service: &RuntimeConfigGenerator{}}
	services["infraManager"] = &GCPFacade{service: &InfraManagerGenerator{}}
	services["deploymentManager"] = &GCPFacade{service: &DeploymentManagerGenerator{}}
	services["cloudAsset"] = &GCPFacade{service: &CloudAssetGenerator{}}
	services["cloudIdentity"] = &GCPFacade{service: &CloudIdentityGenerator{}}
	services["activeDirectory"] = &GCPFacade{service: &ActiveDirectoryGenerator{}}
	services["healthcare"] = &GCPFacade{service: &HealthcareGenerator{}}
	services["oracleDatabase"] = &GCPFacade{service: &OracleDatabaseGenerator{}}
	services["osLogin"] = &GCPFacade{service: &OsLoginGenerator{}}
	services["secretManager"] = &GCPFacade{service: &SecretManagerGenerator{}}
	services["apiGateway"] = &GCPFacade{service: &ApiGatewayGenerator{}}
	services["apiHub"] = &GCPFacade{service: &ApiHubGenerator{}}
	services["appHub"] = &GCPFacade{service: &AppHubGenerator{}}
	services["backupDr"] = &GCPFacade{service: &BackupDrGenerator{}}
	services["blockchainNodeEngine"] = &GCPFacade{service: &BlockchainNodeEngineGenerator{}}
	services["cloudCompliance"] = &GCPFacade{service: &CloudComplianceGenerator{}}
	services["cloudFunctionsIam"] = &GCPFacade{service: &CloudFunctionsIamGenerator{}}
	services["cloudDomains"] = &GCPFacade{service: &CloudDomainsGenerator{}}
	services["dataLossPrevention"] = &GCPFacade{service: &DlpGenerator{}}
	services["developerConnect"] = &GCPFacade{service: &DeveloperConnectGenerator{}}
	services["identityPlatform"] = &GCPFacade{service: &IdentityPlatformGenerator{}}
	services["integrationConnectors"] = &GCPFacade{service: &IntegrationConnectorsGenerator{}}
	services["integrations"] = &GCPFacade{service: &IntegrationsGenerator{}}
	services["looker"] = &GCPFacade{service: &LookerGenerator{}}
	services["managedKafka"] = &GCPFacade{service: &ManagedKafkaGenerator{}}
	services["parameterManager"] = &GCPFacade{service: &ParameterManagerGenerator{}}
	services["privilegedAccessManager"] = &GCPFacade{service: &PrivilegedAccessGenerator{}}
	services["saasRuntime"] = &GCPFacade{service: &SaasRuntimeGenerator{}}
	services["siteVerification"] = &GCPFacade{service: &SiteVerificationGenerator{}}
	services["storageInsights"] = &GCPFacade{service: &StorageInsightsGenerator{}}
	services["storageTransfer"] = &GCPFacade{service: &StorageTransferGenerator{}}
	services["transcoder"] = &GCPFacade{service: &TranscoderGenerator{}}
	services["gkeHub"] = &GCPFacade{service: &GkeHubGenerator{}}
	services["gkeBackup"] = &GCPFacade{service: &GkeBackupGenerator{}}
	services["gkeOnprem"] = &GCPFacade{service: &GkeOnpremGenerator{}}
	services["containerMulticloud"] = &GCPFacade{service: &ContainerMulticloudGenerator{}}
	services["vmwareEngine"] = &GCPFacade{service: &VmwareEngineGenerator{}}
	services["edgeContainer"] = &GCPFacade{service: &EdgeContainerGenerator{}}
	services["edgeNetwork"] = &GCPFacade{service: &EdgeNetworkGenerator{}}
	services["netapp"] = &GCPFacade{service: &NetappGenerator{}}
	services["filestore"] = &GCPFacade{service: &FilestoreGenerator{}}
	services["parallelstore"] = &GCPFacade{service: &ParallelstoreGenerator{}}
	services["lustre"] = &GCPFacade{service: &LustreGenerator{}}
	services["hypercomputeCluster"] = &GCPFacade{service: &HypercomputeClusterGenerator{}}
	services["tpuV2"] = &GCPFacade{service: &TpuGenerator{}}
	services["workstations"] = &GCPFacade{service: &WorkstationsGenerator{}}
	services["cloudDeploy"] = &GCPFacade{service: &CloudDeployGenerator{}}
	services["artifactRegistry"] = &GCPFacade{service: &ArtifactRegistryGenerator{}}
	services["appEngine"] = &GCPFacade{service: &AppEngineGenerator{}}
	services["biglake"] = &GCPFacade{service: &BiglakeGenerator{}}
	services["dialogflowCx"] = &GCPFacade{service: &DialogflowCxGenerator{}}
	services["dialogflow"] = &GCPFacade{service: &DialogflowGenerator{}}
	services["ces"] = &GCPFacade{service: &CesGenerator{}}
	services["contactCenterInsights"] = &GCPFacade{service: &ContactCenterInsightsGenerator{}}
	services["documentAi"] = &GCPFacade{service: &DocumentAiGenerator{}}
	services["documentAiWarehouse"] = &GCPFacade{service: &DocumentAiWarehouseGenerator{}}
	services["modelArmor"] = &GCPFacade{service: &ModelArmorGenerator{}}
	services["colab"] = &GCPFacade{service: &ColabGenerator{}}
	services["workbench"] = &GCPFacade{service: &WorkbenchGenerator{}}
	services["agentRegistry"] = &GCPFacade{service: &AgentGenerator{}}
	services["agenticApplications"] = &GCPFacade{service: &AgenticApplicationsGenerator{}}
	services["vectorSearch"] = &GCPFacade{service: &VectorSearchGenerator{}}
	services["privateca"] = &GCPFacade{service: &PrivateCaGenerator{}}
	services["certificateManager"] = &GCPFacade{service: &CertificateManagerGenerator{}}
	services["beyondcorp"] = &GCPFacade{service: &BeyondcorpGenerator{}}
	services["binaryAuthorization"] = &GCPFacade{service: &BinaryAuthorizationGenerator{}}
	services["cloudIds"] = &GCPFacade{service: &CloudIdsGenerator{}}
	services["recaptchaEnterprise"] = &GCPFacade{service: &RecaptchaEnterpriseGenerator{}}
	services["securityPosture"] = &GCPFacade{service: &SecurityPostureGenerator{}}
	services["vpcAccess"] = &GCPFacade{service: &VpcAccessGenerator{}}
	services["networkManagement"] = &GCPFacade{service: &NetworkManagementGenerator{}}
	services["serviceNetworking"] = &GCPFacade{service: &ServiceNetworkingGenerator{}}
	services["endpoints"] = &GCPFacade{service: &EndpointsGenerator{}}
	services["serviceDirectory"] = &GCPFacade{service: &ServiceDirectoryGenerator{}}
	return services
}

func (GCPProvider) GetResourceConnections() map[string]map[string][]string {
	return map[string]map[string][]string{
		"backendBuckets": {"gcs": []string{"bucket_name", "name"}},
		"firewall":       {"networks": []string{"network", "self_link"}},
		"gke": {
			"networks":    []string{"network", "self_link"},
			"subnetworks": []string{"subnetwork", "self_link"},
		},
		"instanceTemplates": {
			"networks":    []string{"network", "self_link"},
			"subnetworks": []string{"subnetworks", "self_link"},
		},
		"regionInstanceGroupManagers": {"instanceTemplates": []string{"version.instance_template", "self_link"}},
		"instanceGroups":              {"instanceTemplates": []string{"version.instance_template", "self_link"}},
		"routes":                      {"networks": []string{"network", "self_link"}},
		"subnetworks":                 {"networks": []string{"network", "self_link"}},
		"forwardingRules": {
			"regionBackendServices": []string{"backend_service", "self_link"},
			"networks":              []string{"network", "self_link"},
		},
		"globalForwardingRules": {
			"targetHttpsProxies": []string{"target", "self_link"},
			"targetHttpProxies":  []string{"target", "self_link"},
			"targetSslProxies":   []string{"target", "self_link"},
		},
		"targetHttpsProxies": {
			"urlMaps": []string{"url_map", "self_link"},
		},
		"targetHttpProxies": {
			"urlMaps": []string{"url_map", "self_link"},
		},
		"targetSslProxies": {
			"backendServices": []string{"backend_service", "self_link"},
		},
		"backendServices": {
			"regionInstanceGroupManagers": []string{"backend.group", "instance_group"},
			"instanceGroupManagers":       []string{"backend.group", "instance_group"},
			"healthChecks":                []string{"health_checks", "self_link"},
		},
		"regionBackendServices": {
			"regionInstanceGroupManagers": []string{"backend.group", "instance_group"},
			"instanceGroupManagers":       []string{"backend.group", "instance_group"},
			"healthChecks":                []string{"health_checks", "self_link"},
		},
		"urlMaps": {
			"backendServices": []string{
				"default_service", "self_link",
				"path_matcher.default_service", "self_link",
				"path_matcher.path_rule.service", "self_link",
			},
			"regionBackendServices": []string{
				"default_service", "self_link",
				"path_matcher.default_service", "self_link",
				"path_matcher.path_rule.service", "self_link",
			},
		},
	}
}
func (p GCPProvider) GetProviderData(arg ...string) map[string]interface{} {
	return map[string]interface{}{
		"provider": map[string]interface{}{
			p.GetName(): map[string]interface{}{
				"project": p.projectName,
			},
		},
	}
}
