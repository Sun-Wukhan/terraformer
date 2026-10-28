### Use with GCP

In order to access the information from your Google Project, you need to provide authentication credentials
by setting up the environment variable `GOOGLE_APPLICATION_CREDENTIALS` with the file path of the JSON
file that contains your service account key. 

[![asciicast](https://asciinema.org/a/243961.svg)](https://asciinema.org/a/243961)

Example:

```
terraformer import google --resources=gcs,forwardingRules,httpHealthChecks --connect=true --regions=europe-west1,europe-west4 --projects=aaa,fff
terraformer import google --resources=gcs,forwardingRules,httpHealthChecks --filter=compute_firewall=rule1:rule2:rule3 --regions=europe-west1 --projects=aaa,fff
```

For google-beta provider:

```
terraformer import google --resources=gcs,forwardingRules,httpHealthChecks --regions=europe-west4 --projects=aaa --provider-type beta
```

List of supported GCP services:

*   `accessContextManager`
    * `google_access_context_manager_access_level`
    * `google_access_context_manager_access_policy`
    * `google_access_context_manager_access_policy_iam`
    * `google_access_context_manager_authorized_orgs_desc`
    * `google_access_context_manager_gcp_user_access_binding`
    * `google_access_context_manager_service_perimeter`
*   `addresses`
    * `google_compute_address`
*   `alloydb`
    * `google_alloydb_backup`
    * `google_alloydb_cluster`
    * `google_alloydb_instance`
    * `google_alloydb_user`
*   `apiGateway`
    * `google_api_gateway_api`
    * `google_api_gateway_api_config`
    * `google_api_gateway_api_config_iam_binding`
    * `google_api_gateway_api_config_iam_member`
    * `google_api_gateway_api_config_iam_policy`
    * `google_api_gateway_api_iam_binding`
    * `google_api_gateway_api_iam_member`
    * `google_api_gateway_api_iam_policy`
    * `google_api_gateway_gateway`
    * `google_api_gateway_gateway_iam_binding`
    * `google_api_gateway_gateway_iam_member`
    * `google_api_gateway_gateway_iam_policy`
*   `agentRegistry`
    * `google_agent_identity_auth_provider`
    * `google_agent_identity_auth_provider_iam_binding`
    * `google_agent_identity_auth_provider_iam_member`
    * `google_agent_identity_auth_provider_iam_policy`
    * `google_agent_registry_binding`
    * `google_agent_registry_service`
*   `agenticApplications`
    * (no types generated — `google_agentic_applications_analyst_agent_persona` is a documented skip: the API exposes no publicly documented list endpoint or discovery document)
*   `apigee`
    * `google_apigee_addons_config`
    * `google_apigee_api`
    * `google_apigee_api_deployment`
    * `google_apigee_api_product`
    * `google_apigee_app_group`
    * `google_apigee_control_plane_access`
    * `google_apigee_data_collector`
    * `google_apigee_datastore`
    * `google_apigee_developer`
    * `google_apigee_developer_app`
    * `google_apigee_dns_zone`
    * `google_apigee_endpoint_attachment`
    * `google_apigee_envgroup`
    * `google_apigee_envgroup_attachment`
    * `google_apigee_environment`
    * `google_apigee_environment_addons_config`
    * `google_apigee_environment_api_revision_deployment`
    * `google_apigee_environment_debugmask`
    * `google_apigee_environment_iam_binding`
    * `google_apigee_environment_iam_member`
    * `google_apigee_environment_iam_policy`
    * `google_apigee_flowhook`
    * `google_apigee_instance`
    * `google_apigee_instance_attachment`
    * `google_apigee_nat_address`
    * `google_apigee_organization`
    * `google_apigee_security_action`
    * `google_apigee_security_monitoring_condition`
    * `google_apigee_security_profile_v2`
    * `google_apigee_sharedflow`
    * `google_apigee_sharedflow_deployment`
    * `google_apigee_space`
    * `google_apigee_sync_authorization`
*   `apiHub`
    * `google_apihub_api_hub_instance`
    * `google_apihub_curation`
    * `google_apihub_host_project_registration`
    * `google_apihub_plugin`
    * `google_apihub_plugin_instance`
    * `google_apihub_runtime_project_attachment`
*   `appHub`
    * `google_apphub_application`
    * `google_apphub_boundary`
    * `google_apphub_service`
    * `google_apphub_service_project_attachment`
    * `google_apphub_workload`
*   `appEngine`
    * `google_app_engine_application`
    * `google_app_engine_application_url_dispatch_rules`
    * `google_app_engine_domain_mapping`
    * `google_app_engine_firewall_rule`
    * `google_app_engine_flexible_app_version`
    * `google_app_engine_service_network_settings`
    * `google_app_engine_service_split_traffic`
    * `google_app_engine_standard_app_version`
*   `artifactRegistry`
    * `google_artifact_registry_project_config`
    * `google_artifact_registry_repository`
    * `google_artifact_registry_repository_iam_binding`
    * `google_artifact_registry_repository_iam_member`
    * `google_artifact_registry_repository_iam_policy`
    * `google_artifact_registry_rule`
    * `google_artifact_registry_vpcsc_config`
*   `autoscalers`
    * `google_compute_autoscaler`
*   `biglake`
    * `google_biglake_catalog`
    * `google_biglake_database`
    * `google_biglake_hive_catalog`
    * `google_biglake_hive_catalog_iam_binding`
    * `google_biglake_hive_catalog_iam_member`
    * `google_biglake_hive_catalog_iam_policy`
    * `google_biglake_hive_database`
    * `google_biglake_hive_database_iam_binding`
    * `google_biglake_hive_database_iam_member`
    * `google_biglake_hive_database_iam_policy`
    * `google_biglake_hive_table`
    * `google_biglake_hive_table_iam_binding`
    * `google_biglake_hive_table_iam_member`
    * `google_biglake_hive_table_iam_policy`
    * `google_biglake_iceberg_catalog`
    * `google_biglake_iceberg_catalog_iam_binding`
    * `google_biglake_iceberg_catalog_iam_member`
    * `google_biglake_iceberg_catalog_iam_policy`
    * `google_biglake_iceberg_namespace`
    * `google_biglake_iceberg_namespace_iam_binding`
    * `google_biglake_iceberg_namespace_iam_member`
    * `google_biglake_iceberg_namespace_iam_policy`
    * `google_biglake_iceberg_table`
    * `google_biglake_iceberg_table_iam_binding`
    * `google_biglake_iceberg_table_iam_member`
    * `google_biglake_iceberg_table_iam_policy`
    * `google_biglake_table`
*   `backendBuckets`
    * `google_compute_backend_bucket`
*   `backendServices`
    * `google_compute_backend_service`
*   `bigtable`
    * `google_bigtable_app_profile`
    * `google_bigtable_authorized_view`
    * `google_bigtable_instance`
    * `google_bigtable_instance_iam_binding`
    * `google_bigtable_instance_iam_member`
    * `google_bigtable_instance_iam_policy`
    * `google_bigtable_logical_view`
    * `google_bigtable_materialized_view`
    * `google_bigtable_table`
    * `google_bigtable_table_iam_binding`
    * `google_bigtable_table_iam_member`
    * `google_bigtable_table_iam_policy`
*   `backupDr`
    * `google_backup_dr_backup_plan`
    * `google_backup_dr_backup_plan_association`
    * `google_backup_dr_backup_vault`
    * `google_backup_dr_management_server`
*   `beyondcorp`
    * `google_beyondcorp_security_gateway`
    * `google_beyondcorp_security_gateway_application`
    * `google_beyondcorp_security_gateway_application_iam_binding`
    * `google_beyondcorp_security_gateway_application_iam_member`
    * `google_beyondcorp_security_gateway_application_iam_policy`
    * `google_beyondcorp_security_gateway_iam_binding`
    * `google_beyondcorp_security_gateway_iam_member`
    * `google_beyondcorp_security_gateway_iam_policy`
*   `binaryAuthorization`
    * `google_binary_authorization_attestor`
    * `google_binary_authorization_attestor_iam_binding`
    * `google_binary_authorization_attestor_iam_member`
    * `google_binary_authorization_attestor_iam_policy`
    * `google_binary_authorization_policy`
*   `bigQuery`
    * `google_bigquery_analytics_hub_data_exchange`
    * `google_bigquery_analytics_hub_data_exchange_iam_binding`
    * `google_bigquery_analytics_hub_data_exchange_iam_member`
    * `google_bigquery_analytics_hub_data_exchange_iam_policy`
    * `google_bigquery_analytics_hub_data_exchange_subscription`
    * `google_bigquery_analytics_hub_listing`
    * `google_bigquery_analytics_hub_listing_iam_binding`
    * `google_bigquery_analytics_hub_listing_iam_member`
    * `google_bigquery_analytics_hub_listing_iam_policy`
    * `google_bigquery_analytics_hub_query_template`
    * `google_bigquery_bi_reservation`
    * `google_bigquery_capacity_commitment`
    * `google_bigquery_connection`
    * `google_bigquery_connection_iam_binding`
    * `google_bigquery_connection_iam_member`
    * `google_bigquery_connection_iam_policy`
    * `google_bigquery_data_transfer_config`
    * `google_bigquery_data_transfer_data_source_enrollment`
    * `google_bigquery_datapolicy_data_policy`
    * `google_bigquery_datapolicy_data_policy_iam_binding`
    * `google_bigquery_datapolicy_data_policy_iam_member`
    * `google_bigquery_datapolicy_data_policy_iam_policy`
    * `google_bigquery_dataset`
    * `google_bigquery_dataset_iam_binding`
    * `google_bigquery_dataset_iam_member`
    * `google_bigquery_dataset_iam_policy`
    * `google_bigquery_reservation`
    * `google_bigquery_reservation_assignment`
    * `google_bigquery_reservation_group`
    * `google_bigquery_routine`
    * `google_bigquery_routine_iam_binding`
    * `google_bigquery_routine_iam_member`
    * `google_bigquery_routine_iam_policy`
    * `google_bigquery_row_access_policy`
    * `google_bigquery_table`
    * `google_bigquery_table_iam_binding`
    * `google_bigquery_table_iam_member`
    * `google_bigquery_table_iam_policy`
*   `blockchainNodeEngine`
    * `google_blockchain_node_engine_blockchain_nodes`
*   `cloudCompliance`
    * `google_cloud_security_compliance_cloud_control`
    * `google_cloud_security_compliance_framework`
    * `google_cloud_security_compliance_framework_deployment`
*   `cloudDomains`
    * `google_clouddomains_registration`
*   `cloudFunctions`
    * `google_cloudfunctions_function`
    * `google_cloudfunctions2_function`
*   `cloudbuild`
    * `google_cloudbuild_trigger`
*   `cloudFunctionsIam`
    * `google_cloudfunctions_function_iam_binding`
    * `google_cloudfunctions_function_iam_member`
    * `google_cloudfunctions_function_iam_policy`
    * `google_cloudfunctions2_function_iam_binding`
    * `google_cloudfunctions2_function_iam_member`
    * `google_cloudfunctions2_function_iam_policy`
*   `cloudDeploy`
    * `google_clouddeploy_automation`
    * `google_clouddeploy_custom_target_type`
    * `google_clouddeploy_custom_target_type_iam_binding`
    * `google_clouddeploy_custom_target_type_iam_member`
    * `google_clouddeploy_custom_target_type_iam_policy`
    * `google_clouddeploy_delivery_pipeline`
    * `google_clouddeploy_delivery_pipeline_iam_binding`
    * `google_clouddeploy_delivery_pipeline_iam_member`
    * `google_clouddeploy_delivery_pipeline_iam_policy`
    * `google_clouddeploy_deploy_policy`
    * `google_clouddeploy_target`
    * `google_clouddeploy_target_iam_binding`
    * `google_clouddeploy_target_iam_member`
    * `google_clouddeploy_target_iam_policy`
*   `certificateManager`
    * `google_certificate_manager_certificate`
    * `google_certificate_manager_certificate_issuance_config`
    * `google_certificate_manager_certificate_map`
    * `google_certificate_manager_certificate_map_entry`
    * `google_certificate_manager_dns_authorization`
    * `google_certificate_manager_trust_config`
*   `cloudIds`
    * `google_cloud_ids_endpoint`
*   `cloudRun`
    * `google_cloud_run_domain_mapping`
    * `google_cloud_run_service`
    * `google_cloud_run_service_iam_binding`
    * `google_cloud_run_service_iam_member`
    * `google_cloud_run_service_iam_policy`
    * `google_cloud_run_v2_job`
    * `google_cloud_run_v2_job_iam_binding`
    * `google_cloud_run_v2_job_iam_member`
    * `google_cloud_run_v2_job_iam_policy`
    * `google_cloud_run_v2_service`
    * `google_cloud_run_v2_service_iam_binding`
    * `google_cloud_run_v2_service_iam_member`
    * `google_cloud_run_v2_service_iam_policy`
    * `google_cloud_run_v2_worker_pool`
*   `cloudsql`
    * `google_sql_database`
    * `google_sql_database_instance`
    * `google_sql_source_representation_instance`
    * `google_sql_ssl_cert`
    * `google_sql_user`
*   `chronicle`
    * (no resources — Google SecOps resources hang off a provisioned Chronicle instance, and the Chronicle API exposes no instance list (GetInstance only), so the family cannot be enumerated; see `providers/gcp/chronicle.go`)
*   `computeDepth`
    * `google_compute_cross_site_network`
    * `google_compute_future_reservation`
    * `google_compute_global_vm_extension_policy`
    * `google_compute_ha_vpn_gateway`
    * `google_compute_instant_snapshot`
    * `google_compute_interconnect`
    * `google_compute_interconnect_attachment_group`
    * `google_compute_interconnect_group`
    * `google_compute_network_attachment`
    * `google_compute_network_edge_security_service`
    * `google_compute_preview_feature`
    * `google_compute_public_advertised_prefix`
    * `google_compute_public_delegated_prefix`
    * `google_compute_region_backend_bucket`
    * `google_compute_region_commitment`
    * `google_compute_region_composite_health_check`
    * `google_compute_region_health_aggregation_policy`
    * `google_compute_region_health_source`
    * `google_compute_region_instance_template`
    * `google_compute_region_instant_snapshot`
    * `google_compute_region_network_endpoint_group`
    * `google_compute_region_ssl_policy`
    * `google_compute_region_target_tcp_proxy`
    * `google_compute_rollout_plan`
    * `google_compute_service_attachment`
    * `google_compute_storage_pool`
    * `google_compute_target_grpc_proxy`
    * `google_compute_wire_group`
    * `google_compute_zone_vm_extension_policy`
*   `computeIam`
    * `google_compute_backend_bucket_iam_binding`
    * `google_compute_backend_bucket_iam_member`
    * `google_compute_backend_bucket_iam_policy`
    * `google_compute_backend_service_iam_binding`
    * `google_compute_backend_service_iam_member`
    * `google_compute_backend_service_iam_policy`
    * `google_compute_disk_iam_binding`
    * `google_compute_disk_iam_member`
    * `google_compute_disk_iam_policy`
    * `google_compute_firewall_policy_iam_binding`
    * `google_compute_firewall_policy_iam_member`
    * `google_compute_firewall_policy_iam_policy`
    * `google_compute_image_iam_binding`
    * `google_compute_image_iam_member`
    * `google_compute_image_iam_policy`
    * `google_compute_instance_iam_binding`
    * `google_compute_instance_iam_member`
    * `google_compute_instance_iam_policy`
    * `google_compute_instance_template_iam_binding`
    * `google_compute_instance_template_iam_member`
    * `google_compute_instance_template_iam_policy`
    * `google_compute_instant_snapshot_iam_binding`
    * `google_compute_instant_snapshot_iam_member`
    * `google_compute_instant_snapshot_iam_policy`
    * `google_compute_machine_image_iam_binding`
    * `google_compute_machine_image_iam_member`
    * `google_compute_machine_image_iam_policy`
    * `google_compute_network_firewall_policy_iam_binding`
    * `google_compute_network_firewall_policy_iam_member`
    * `google_compute_network_firewall_policy_iam_policy`
    * `google_compute_region_backend_bucket_iam_binding`
    * `google_compute_region_backend_bucket_iam_member`
    * `google_compute_region_backend_bucket_iam_policy`
    * `google_compute_region_backend_service_iam_binding`
    * `google_compute_region_backend_service_iam_member`
    * `google_compute_region_backend_service_iam_policy`
    * `google_compute_region_disk_iam_binding`
    * `google_compute_region_disk_iam_member`
    * `google_compute_region_disk_iam_policy`
    * `google_compute_region_instant_snapshot_iam_binding`
    * `google_compute_region_instant_snapshot_iam_member`
    * `google_compute_region_instant_snapshot_iam_policy`
    * `google_compute_region_network_firewall_policy_iam_binding`
    * `google_compute_region_network_firewall_policy_iam_member`
    * `google_compute_region_network_firewall_policy_iam_policy`
    * `google_compute_snapshot_iam_binding`
    * `google_compute_snapshot_iam_member`
    * `google_compute_snapshot_iam_policy`
    * `google_compute_storage_pool_iam_binding`
    * `google_compute_storage_pool_iam_member`
    * `google_compute_storage_pool_iam_policy`
    * `google_compute_subnetwork_iam_binding`
    * `google_compute_subnetwork_iam_member`
    * `google_compute_subnetwork_iam_policy`
*   `computePolicies`
    * `google_compute_firewall_policy`
    * `google_compute_firewall_policy_association`
    * `google_compute_firewall_policy_rule`
    * `google_compute_network_firewall_policy`
    * `google_compute_network_firewall_policy_association`
    * `google_compute_network_firewall_policy_packet_mirroring_rule`
    * `google_compute_network_firewall_policy_rule`
    * `google_compute_organization_security_policy`
    * `google_compute_organization_security_policy_association`
    * `google_compute_organization_security_policy_rule`
    * `google_compute_region_network_firewall_policy`
    * `google_compute_region_network_firewall_policy_association`
    * `google_compute_region_network_firewall_policy_rule`
    * `google_compute_region_network_policy`
    * `google_compute_region_network_policy_traffic_classification_rule`
    * `google_compute_region_security_policy`
    * `google_compute_region_security_policy_rule`
    * `google_compute_security_policy_rule`
*   `computeSubresources`
    * `google_compute_disk_resource_policy_attachment`
    * `google_compute_instance_group_membership`
    * `google_compute_instance_group_named_port`
    * `google_compute_instance_settings`
    * `google_compute_network_peering`
    * `google_compute_organization_snapshot_recycle_bin_policy`
    * `google_compute_per_instance_config`
    * `google_compute_project_cloud_armor_tier`
    * `google_compute_project_default_network_tier`
    * `google_compute_project_metadata`
    * `google_compute_region_disk_resource_policy_attachment`
    * `google_compute_region_per_instance_config`
    * `google_compute_region_resize_request`
    * `google_compute_resize_request`
    * `google_compute_resource_policy_attachment`
    * `google_compute_shared_vpc_host_project`
    * `google_compute_shared_vpc_service_project`
    * `google_compute_snapshot_recycle_bin_policy`
    * `google_compute_snapshot_settings`
*   `dataLossPrevention`
    * `google_data_loss_prevention_deidentify_template`
    * `google_data_loss_prevention_discovery_config`
    * `google_data_loss_prevention_inspect_template`
    * `google_data_loss_prevention_job_trigger`
    * `google_data_loss_prevention_stored_info_type`
*   `containerMulticloud`
    * `google_container_aws_cluster`
    * `google_container_aws_node_pool`
    * `google_container_azure_client`
    * `google_container_azure_cluster`
    * `google_container_azure_node_pool`
*   `dataplex`
    * `google_dataplex_aspect_type`
    * `google_dataplex_aspect_type_iam_binding`
    * `google_dataplex_aspect_type_iam_member`
    * `google_dataplex_aspect_type_iam_policy`
    * `google_dataplex_asset`
    * `google_dataplex_asset_iam_binding`
    * `google_dataplex_asset_iam_member`
    * `google_dataplex_asset_iam_policy`
    * `google_dataplex_data_product`
    * `google_dataplex_data_product_data_asset`
    * `google_dataplex_data_product_iam_binding`
    * `google_dataplex_data_product_iam_member`
    * `google_dataplex_data_product_iam_policy`
    * `google_dataplex_datascan`
    * `google_dataplex_datascan_iam_binding`
    * `google_dataplex_datascan_iam_member`
    * `google_dataplex_datascan_iam_policy`
    * `google_dataplex_entry`
    * `google_dataplex_entry_group`
    * `google_dataplex_entry_group_iam_binding`
    * `google_dataplex_entry_group_iam_member`
    * `google_dataplex_entry_group_iam_policy`
    * `google_dataplex_entry_type`
    * `google_dataplex_entry_type_iam_binding`
    * `google_dataplex_entry_type_iam_member`
    * `google_dataplex_entry_type_iam_policy`
    * `google_dataplex_glossary`
    * `google_dataplex_glossary_category`
    * `google_dataplex_glossary_iam_binding`
    * `google_dataplex_glossary_iam_member`
    * `google_dataplex_glossary_iam_policy`
    * `google_dataplex_glossary_term`
    * `google_dataplex_lake`
    * `google_dataplex_lake_iam_binding`
    * `google_dataplex_lake_iam_member`
    * `google_dataplex_lake_iam_policy`
    * `google_dataplex_metadata_feed`
    * `google_dataplex_task`
    * `google_dataplex_task_iam_binding`
    * `google_dataplex_task_iam_member`
    * `google_dataplex_task_iam_policy`
    * `google_dataplex_zone`
    * `google_dataplex_zone_iam_binding`
    * `google_dataplex_zone_iam_member`
    * `google_dataplex_zone_iam_policy`
*   `dataCatalog`
    * `google_data_catalog_entry`
    * `google_data_catalog_entry_group`
    * `google_data_catalog_entry_group_iam_binding`
    * `google_data_catalog_entry_group_iam_member`
    * `google_data_catalog_entry_group_iam_policy`
    * `google_data_catalog_policy_tag`
    * `google_data_catalog_policy_tag_iam_binding`
    * `google_data_catalog_policy_tag_iam_member`
    * `google_data_catalog_policy_tag_iam_policy`
    * `google_data_catalog_tag`
    * `google_data_catalog_taxonomy`
    * `google_data_catalog_taxonomy_iam_binding`
    * `google_data_catalog_taxonomy_iam_member`
    * `google_data_catalog_taxonomy_iam_policy`
*   `dataflow`
    * (no resources — Dataflow jobs are transient execution records and neither job resource documents import of an existing job; see `providers/gcp/dataflow.go`)
*   `dataform`
    * `google_dataform_config`
    * `google_dataform_repository`
    * `google_dataform_repository_iam_binding`
    * `google_dataform_repository_iam_member`
    * `google_dataform_repository_iam_policy`
    * `google_dataform_repository_release_config`
    * `google_dataform_repository_workflow_config`
*   `dataFusion`
    * `google_data_fusion_instance`
    * `google_data_fusion_instance_iam_binding`
    * `google_data_fusion_instance_iam_member`
    * `google_data_fusion_instance_iam_policy`
*   `dataLineage`
    * (no resources — the Data Lineage v1 API has no config resource or endpoint; see `providers/gcp/datalineage.go`)
*   `dataPipeline`
    * `google_data_pipeline_pipeline`
*   `databaseMigrationService`
    * `google_database_migration_service_connection_profile`
    * `google_database_migration_service_migration_job`
    * `google_database_migration_service_private_connection`
*   `dataprocMetastore`
    * `google_dataproc_metastore_federation`
    * `google_dataproc_metastore_federation_iam_binding`
    * `google_dataproc_metastore_federation_iam_member`
    * `google_dataproc_metastore_federation_iam_policy`
    * `google_dataproc_metastore_service`
    * `google_dataproc_metastore_service_iam_binding`
    * `google_dataproc_metastore_service_iam_member`
    * `google_dataproc_metastore_service_iam_policy`
*   `datastream`
    * `google_datastream_connection_profile`
    * `google_datastream_private_connection`
    * `google_datastream_stream`
*   `ces`
    * `google_ces_agent`
    * `google_ces_app`
    * `google_ces_app_root_agent_association`
    * `google_ces_app_version`
    * `google_ces_deployment`
    * `google_ces_evaluation`
    * `google_ces_example`
    * `google_ces_guardrail`
    * `google_ces_security_settings`
    * `google_ces_tool`
    * `google_ces_toolset`
*   `colab`
    * `google_colab_notebook_execution`
    * `google_colab_runtime`
    * `google_colab_runtime_template`
    * `google_colab_runtime_template_iam_binding`
    * `google_colab_runtime_template_iam_member`
    * `google_colab_runtime_template_iam_policy`
    * `google_colab_schedule`
*   `contactCenterInsights`
    * `google_contact_center_insights_analysis_rule`
    * `google_contact_center_insights_assessment_rule`
    * `google_contact_center_insights_auto_labeling_rule`
    * `google_contact_center_insights_encryption_spec`
    * `google_contact_center_insights_qa_question`
    * `google_contact_center_insights_qa_scorecard`
    * `google_contact_center_insights_qa_scorecard_revision`
    * `google_contact_center_insights_view`
*   `dataProc`
    * `google_dataproc_autoscaling_policy`
    * `google_dataproc_autoscaling_policy_iam_binding`
    * `google_dataproc_autoscaling_policy_iam_member`
    * `google_dataproc_autoscaling_policy_iam_policy`
    * `google_dataproc_batch`
    * `google_dataproc_cluster`
    * `google_dataproc_cluster_iam_binding`
    * `google_dataproc_cluster_iam_member`
    * `google_dataproc_cluster_iam_policy`
    * `google_dataproc_session_template`
    * `google_dataproc_workflow_template`
*   `developerConnect`
    * `google_developer_connect_account_connector`
    * `google_developer_connect_connection`
    * `google_developer_connect_git_repository_link`
*   `dialogflow`
    * `google_dialogflow_agent`
    * `google_dialogflow_conversation_profile`
    * `google_dialogflow_encryption_spec`
    * `google_dialogflow_entity_type`
    * `google_dialogflow_environment`
    * `google_dialogflow_fulfillment`
    * `google_dialogflow_generator`
    * `google_dialogflow_intent`
    * `google_dialogflow_sip_trunk`
    * `google_dialogflow_tool`
    * `google_dialogflow_version`
*   `dialogflowCx`
    * `google_dialogflow_cx_agent`
    * `google_dialogflow_cx_entity_type`
    * `google_dialogflow_cx_environment`
    * `google_dialogflow_cx_flow`
    * `google_dialogflow_cx_generative_settings`
    * `google_dialogflow_cx_generator`
    * `google_dialogflow_cx_intent`
    * `google_dialogflow_cx_page`
    * `google_dialogflow_cx_playbook`
    * `google_dialogflow_cx_security_settings`
    * `google_dialogflow_cx_test_case`
    * `google_dialogflow_cx_tool`
    * `google_dialogflow_cx_tool_version`
    * `google_dialogflow_cx_version`
    * `google_dialogflow_cx_webhook`
*   `disks`
    * `google_compute_disk`
*   `documentAi`
    * `google_document_ai_processor`
    * `google_document_ai_processor_default_version`
    * `google_document_ai_schema`
    * (documented skip: `google_document_ai_schema_version` — the dataset schema is a singleton document with no addressable version collection in the v1/v1beta3 APIs)
*   `documentAiWarehouse`
    * `google_document_ai_warehouse_document_schema`
    * `google_document_ai_warehouse_location`
*   `discoveryEngine`
    * `google_discovery_engine_acl_config`
    * `google_discovery_engine_assistant`
    * `google_discovery_engine_chat_engine`
    * `google_discovery_engine_cmek_config`
    * `google_discovery_engine_control`
    * `google_discovery_engine_data_connector`
    * `google_discovery_engine_data_store`
    * `google_discovery_engine_license_config`
    * `google_discovery_engine_recommendation_engine`
    * `google_discovery_engine_schema`
    * `google_discovery_engine_search_engine`
    * `google_discovery_engine_search_engine_iam`
    * `google_discovery_engine_serving_config`
    * `google_discovery_engine_sitemap`
    * `google_discovery_engine_target_site`
    * `google_discovery_engine_user_store`
    * `google_discovery_engine_widget_config`
*   `dns`
    * `google_dns_managed_zone`
    * `google_dns_managed_zone_iam_binding`
    * `google_dns_managed_zone_iam_member`
    * `google_dns_managed_zone_iam_policy`
    * `google_dns_outbound_endpoint`
    * `google_dns_policy`
    * `google_dns_record_set`
    * `google_dns_response_policy`
    * `google_dns_response_policy_rule`
*   `endpoints`
    * `google_endpoints_service`
    * `google_endpoints_service_iam_binding`
    * `google_endpoints_service_iam_member`
    * `google_endpoints_service_iam_policy`
*   `externalVpnGateways`
    * `google_compute_external_vpn_gateway`
*   `edgeContainer`
    * `google_edgecontainer_cluster`
    * `google_edgecontainer_node_pool`
    * `google_edgecontainer_vpn_connection`
*   `edgeNetwork`
    * `google_edgenetwork_interconnect_attachment`
    * `google_edgenetwork_network`
    * `google_edgenetwork_subnet`
*   `filestore`
    * `google_filestore_backup`
    * `google_filestore_instance`
    * `google_filestore_snapshot`
*   `firebase`
    * `google_firebase_ai_logic_config`
    * `google_firebase_ai_logic_prompt_template`
    * `google_firebase_ai_logic_prompt_template_lock`
    * `google_firebase_android_app`
    * `google_firebase_app_check_app_attest_config`
    * `google_firebase_app_check_debug_token`
    * `google_firebase_app_check_device_check_config`
    * `google_firebase_app_check_play_integrity_config`
    * `google_firebase_app_check_recaptcha_enterprise_config`
    * `google_firebase_app_check_recaptcha_v3_config`
    * `google_firebase_app_check_resource_policy`
    * `google_firebase_app_check_service_config`
    * `google_firebase_app_hosting_backend`
    * `google_firebase_app_hosting_build`
    * `google_firebase_app_hosting_default_domain`
    * `google_firebase_app_hosting_domain`
    * `google_firebase_app_hosting_traffic`
    * `google_firebase_apple_app`
    * `google_firebase_data_connect_service`
    * `google_firebase_database_instance`
    * `google_firebase_extensions_instance`
    * `google_firebase_hosting_channel`
    * `google_firebase_hosting_custom_domain`
    * `google_firebase_hosting_release`
    * `google_firebase_hosting_site`
    * `google_firebase_hosting_version`
    * `google_firebase_project`
    * `google_firebase_remote_config_remote_config`
    * `google_firebase_storage_bucket`
    * `google_firebase_storage_default_bucket`
    * `google_firebase_web_app`
    * `google_firebaserules_release`
    * `google_firebaserules_ruleset`
*   `firestore`
    * `google_firestore_backup_schedule`
    * `google_firestore_database`
    * `google_firestore_field`
    * `google_firestore_index`
    * `google_firestore_user_creds`
*   `firewall`
    * `google_compute_firewall`
*   `forwardingRules`
    * `google_compute_forwarding_rule`
*   `gcs`
    * `google_storage_anywhere_cache`
    * `google_storage_batch_operations_job`
    * `google_storage_bucket`
    * `google_storage_bucket_access_control`
    * `google_storage_bucket_acl`
    * `google_storage_bucket_iam_binding`
    * `google_storage_bucket_iam_member`
    * `google_storage_bucket_iam_policy`
    * `google_storage_bucket_object`
    * `google_storage_control_folder_intelligence_config`
    * `google_storage_control_organization_intelligence_config`
    * `google_storage_control_project_feature_config`
    * `google_storage_control_project_intelligence_config`
    * `google_storage_default_object_access_control`
    * `google_storage_default_object_acl`
    * `google_storage_folder`
    * `google_storage_ftp_server`
    * `google_storage_ftp_user`
    * `google_storage_hmac_key`
    * `google_storage_managed_folder`
    * `google_storage_managed_folder_iam`
    * `google_storage_notification`
    * `google_storage_object_access_control`
    * `google_storage_object_acl`
*   `gemini`
    * `google_gemini_code_repository_index`
    * `google_gemini_code_tools_setting`
    * `google_gemini_code_tools_setting_binding`
    * `google_gemini_data_analytics_data_agent`
    * `google_gemini_data_analytics_data_agent_iam`
    * `google_gemini_data_sharing_with_google_setting`
    * `google_gemini_data_sharing_with_google_setting_binding`
    * `google_gemini_gda_observability_setting`
    * `google_gemini_gda_observability_setting_binding`
    * `google_gemini_gemini_gcp_enablement_setting`
    * `google_gemini_gemini_gcp_enablement_setting_binding`
    * `google_gemini_gibq_observability_setting`
    * `google_gemini_gibq_observability_setting_binding`
    * `google_gemini_logging_setting`
    * `google_gemini_logging_setting_binding`
    * `google_gemini_release_channel_setting`
    * `google_gemini_release_channel_setting_binding`
    * `google_gemini_repository_group`
    * `google_gemini_repository_group_iam`
*   `gke`
    * `google_container_analysis_note`
    * `google_container_analysis_note_iam_binding`
    * `google_container_analysis_note_iam_member`
    * `google_container_analysis_note_iam_policy`
    * `google_container_analysis_occurrence`
    * `google_container_attached_cluster`
    * `google_container_cluster`
    * `google_container_node_pool`
    * (not generated: `google_container_registry` — no Container Registry API resource exists to enumerate; the provider resource is a project singleton backed by a Cloud Storage bucket, which the `gcs` service already emits; see `providers/gcp/gke_depth.go`)
*   `gkeBackup`
    * `google_gke_backup_backup_channel`
    * `google_gke_backup_backup_plan`
    * `google_gke_backup_backup_plan_iam_binding`
    * `google_gke_backup_backup_plan_iam_member`
    * `google_gke_backup_backup_plan_iam_policy`
    * `google_gke_backup_restore_channel`
    * `google_gke_backup_restore_plan`
    * `google_gke_backup_restore_plan_iam_binding`
    * `google_gke_backup_restore_plan_iam_member`
    * `google_gke_backup_restore_plan_iam_policy`
*   `gkeHub`
    * `google_gke_hub_feature`
    * `google_gke_hub_feature_iam_binding`
    * `google_gke_hub_feature_iam_member`
    * `google_gke_hub_feature_iam_policy`
    * `google_gke_hub_feature_membership`
    * `google_gke_hub_fleet`
    * `google_gke_hub_membership`
    * `google_gke_hub_membership_binding`
    * `google_gke_hub_membership_iam_binding`
    * `google_gke_hub_membership_iam_member`
    * `google_gke_hub_membership_iam_policy`
    * `google_gke_hub_membership_rbac_role_binding`
    * `google_gke_hub_namespace`
    * `google_gke_hub_rollout_sequence`
    * `google_gke_hub_scope`
    * `google_gke_hub_scope_iam_binding`
    * `google_gke_hub_scope_iam_member`
    * `google_gke_hub_scope_iam_policy`
    * `google_gke_hub_scope_rbac_role_binding`
*   `gkeOnprem`
    * `google_gkeonprem_bare_metal_admin_cluster`
    * `google_gkeonprem_bare_metal_cluster`
    * `google_gkeonprem_bare_metal_node_pool`
    * `google_gkeonprem_vmware_admin_cluster`
    * `google_gkeonprem_vmware_cluster`
    * `google_gkeonprem_vmware_node_pool`
*   `globalAddresses`
    * `google_compute_global_address`
*   `globalForwardingRules`
    * `google_compute_global_forwarding_rule`
*   `globalNetworkEndpointGroups`
    * `google_compute_global_network_endpoint_group`
*   `healthChecks`
    * `google_compute_health_check`
*   `httpHealthChecks`
    * `google_compute_http_health_check`
*   `httpsHealthChecks`
    * `google_compute_https_health_check`
*   `hypercomputeCluster`
    * (no resources — the Cluster Director API has no generated client in the pinned google.golang.org/api module and no verifiable published list endpoint, so `google_hypercomputecluster_cluster` cannot be enumerated; see `providers/gcp/hypercomputecluster.go`)
*   `lustre`
    * `google_lustre_instance`
*   `iam`
    * `google_iam_oauth_client`
    * `google_iam_oauth_client_credential`
    * `google_iam_workload_identity_pool`
    * `google_iam_workload_identity_pool_iam_binding`
    * `google_iam_workload_identity_pool_iam_member`
    * `google_iam_workload_identity_pool_iam_policy`
    * `google_iam_workload_identity_pool_managed_identity`
    * `google_iam_workload_identity_pool_namespace`
    * `google_iam_workload_identity_pool_provider`
    * `google_project_iam_custom_role`
    * `google_project_iam_member`
    * `google_service_account`
    * `google_iam_workforce_pool`
    * `google_iam_workforce_pool_iam_binding`
    * `google_iam_workforce_pool_iam_member`
    * `google_iam_workforce_pool_iam_policy`
    * `google_iam_workforce_pool_provider`
    * `google_iam_workforce_pool_provider_key`
*   `identityPlatform`
    * `google_identity_platform_config`
    * `google_identity_platform_default_supported_idp_config`
    * `google_identity_platform_inbound_saml_config`
    * `google_identity_platform_oauth_idp_config`
    * `google_identity_platform_tenant`
    * `google_identity_platform_tenant_default_supported_idp_config`
    * `google_identity_platform_tenant_inbound_saml_config`
    * `google_identity_platform_tenant_oauth_idp_config`
*   `images`
    * `google_compute_image`
*   `instanceGroupManagers`
    * `google_compute_instance_group_manager`
*   `instanceGroups`
    * `google_compute_instance_group`
*   `instanceTemplates`
    * `google_compute_instance_template`
*   `instances`
    * `google_compute_instance`
*   `integrationConnectors`
    * `google_integration_connectors_connection`
    * `google_integration_connectors_endpoint_attachment`
    * `google_integration_connectors_managed_zone`
*   `integrations`
    * `google_integrations_auth_config`
*   `interconnectAttachments`
    * `google_compute_interconnect_attachment`
*   `kms`
    * `google_kms_crypto_key`
    * `google_kms_crypto_key_iam_binding`
    * `google_kms_crypto_key_iam_member`
    * `google_kms_crypto_key_iam_policy`
    * `google_kms_crypto_key_version`
    * `google_kms_ekm_connection`
    * `google_kms_ekm_connection_iam_binding`
    * `google_kms_ekm_connection_iam_member`
    * `google_kms_ekm_connection_iam_policy`
    * `google_kms_key_handle`
    * `google_kms_key_ring`
*   `looker`
    * `google_looker_instance`
    * `google_kms_key_ring_iam_binding`
    * `google_kms_key_ring_iam_member`
    * `google_kms_key_ring_iam_policy`
    * `google_kms_key_ring_import_job`
    * `google_kms_organization_kaj_policy_config`
    * `google_kms_project_autokey_config`
    * `google_kms_project_kaj_policy_config`
*   `machineImages`
    * `google_compute_machine_image`
*   `logging`
    * `google_logging_linked_dataset`
    * `google_logging_log_scope`
    * `google_logging_log_view`
    * `google_logging_log_view_iam_binding`
    * `google_logging_log_view_iam_member`
    * `google_logging_log_view_iam_policy`
    * `google_logging_metric`
    * `google_logging_project_bucket_config`
    * `google_logging_project_exclusion`
    * `google_logging_project_sink`
    * `google_logging_saved_query`
*   `memcache`
    * `google_memcache_instance`
*   `memorystore`
    * `google_memorystore_acl_policy`
    * `google_memorystore_instance`
*   `managedKafka`
    * `google_managed_kafka_acl`
    * `google_managed_kafka_cluster`
    * `google_managed_kafka_connect_cluster`
    * `google_managed_kafka_connector`
    * `google_managed_kafka_topic`
*   `memoryStore`
    * `google_redis_cluster`
    * `google_redis_cluster_acl_policy`
    * `google_redis_instance`
*   `modelArmor`
    * `google_model_armor_floorsetting`
    * `google_model_armor_template`
*   `monitoring`
    * `google_monitoring_alert_policy`
    * `google_monitoring_group`
    * `google_monitoring_notification_channel`
    * `google_monitoring_uptime_check_config`
*   `migrationCenter`
    * `google_migration_center_discovery_client`
    * `google_migration_center_group`
    * `google_migration_center_import_data_file`
    * `google_migration_center_import_job`
    * `google_migration_center_preference_set`
    * `google_migration_center_report`
    * `google_migration_center_report_config`
    * `google_migration_center_settings`
    * `google_migration_center_source`
*   `networkConnectivity`
    * `google_network_connectivity_group`
    * `google_network_connectivity_hub`
    * `google_network_connectivity_hub_iam_binding`
    * `google_network_connectivity_hub_iam_member`
    * `google_network_connectivity_hub_iam_policy`
    * `google_network_connectivity_internal_range`
    * `google_network_connectivity_policy_based_route`
    * `google_network_connectivity_regional_endpoint`
    * `google_network_connectivity_service_connection_policy`
    * `google_network_connectivity_spoke`
*   `networkSecurity`
    * `google_network_security_address_group`
    * `google_network_security_address_group_iam_binding`
    * `google_network_security_address_group_iam_member`
    * `google_network_security_address_group_iam_policy`
    * `google_network_security_authorization_policy`
    * `google_network_security_authz_policy`
    * `google_network_security_backend_authentication_config`
    * `google_network_security_client_tls_policy`
    * `google_network_security_firewall_endpoint`
    * `google_network_security_firewall_endpoint_association`
    * `google_network_security_gateway_security_policy`
    * `google_network_security_gateway_security_policy_rule`
    * `google_network_security_intercept_deployment`
    * `google_network_security_intercept_deployment_group`
    * `google_network_security_intercept_endpoint_group`
    * `google_network_security_intercept_endpoint_group_association`
    * `google_network_security_mirroring_deployment`
    * `google_network_security_mirroring_deployment_group`
    * `google_network_security_mirroring_endpoint_group`
    * `google_network_security_mirroring_endpoint_group_association`
    * `google_network_security_security_profile`
    * `google_network_security_security_profile_group`
    * `google_network_security_server_tls_policy`
    * `google_network_security_tls_inspection_policy`
    * `google_network_security_url_lists`
*   `networkServices`
    * `google_network_services_authz_extension`
    * `google_network_services_endpoint_policy`
    * `google_network_services_gateway`
    * `google_network_services_grpc_route`
    * `google_network_services_http_route`
    * `google_network_services_lb_edge_extension`
    * `google_network_services_lb_route_extension`
    * `google_network_services_lb_traffic_extension`
    * `google_network_services_mesh`
    * `google_network_services_service_binding`
    * `google_network_services_service_lb_policies`
    * `google_network_services_tcp_route`
    * `google_network_services_tls_route`
    * `google_network_services_wasm_plugin`
*   `networkManagement`
    * `google_network_management_connectivity_test`
    * `google_network_management_network_monitoring_provider`
    * `google_network_management_organization_vpc_flow_logs_config`
    * `google_network_management_vpc_flow_logs_config`
*   `networks`
    * `google_compute_network`
*   `netapp`
    * `google_netapp_active_directory`
    * `google_netapp_backup`
    * `google_netapp_backup_policy`
    * `google_netapp_backup_vault`
    * `google_netapp_host_group`
    * `google_netapp_kmsconfig`
    * `google_netapp_storage_pool`
    * `google_netapp_volume`
    * `google_netapp_volume_quota_rule`
    * `google_netapp_volume_replication`
    * `google_netapp_volume_snapshot`
*   `parallelstore`
    * `google_parallelstore_instance`
*   `nodeGroups`
    * `google_compute_node_group`
*   `nodeTemplates`
    * `google_compute_node_template`
*   `packetMirrorings`
    * `google_compute_packet_mirroring`
*   `parameterManager`
    * `google_parameter_manager_parameter`
    * `google_parameter_manager_parameter_version`
    * `google_parameter_manager_regional_parameter`
    * `google_parameter_manager_regional_parameter_version`
    * `google_parameter_manager_template`
*   `privilegedAccessManager`
    * `google_privileged_access_manager_entitlement`
*   `project`
    * `google_project`
    * `google_project_service`
    * `google_project_usage_export_bucket`
*   `workbench`
    * `google_workbench_instance`
    * `google_workbench_instance_iam_binding`
    * `google_workbench_instance_iam_member`
    * `google_workbench_instance_iam_policy`
*   `privateca`
    * `google_privateca_ca_pool`
    * `google_privateca_ca_pool_iam_binding`
    * `google_privateca_ca_pool_iam_member`
    * `google_privateca_ca_pool_iam_policy`
    * `google_privateca_certificate`
    * `google_privateca_certificate_authority`
    * `google_privateca_certificate_template`
    * `google_privateca_certificate_template_iam_binding`
    * `google_privateca_certificate_template_iam_member`
    * `google_privateca_certificate_template_iam_policy`
*   `pubsub`
    * `google_pubsub_subscription`
    * `google_pubsub_topic`
*   `recaptchaEnterprise`
    * `google_recaptcha_enterprise_key`
*   `regionAutoscalers`
    * `google_compute_region_autoscaler`
*   `regionBackendServices`
    * `google_compute_region_backend_service`
*   `regionDisks`
    * `google_compute_region_disk`
*   `regionHealthChecks`
    * `google_compute_region_health_check`
*   `regionInstanceGroupManagers`
    * `google_compute_region_instance_group_manager`
*   `regionInstanceGroups`
    * `google_compute_region_instance_group`
*   `regionSslCertificates`
    * `google_compute_region_ssl_certificate`
*   `regionTargetHttpProxies`
    * `google_compute_region_target_http_proxy`
*   `regionTargetHttpsProxies`
    * `google_compute_region_target_https_proxy`
*   `regionUrlMaps`
    * `google_compute_region_url_map`
*   `reservations`
    * `google_compute_reservation`
*   `resourcePolicies`
    * `google_compute_resource_policy`
*   `routers`
    * `google_compute_router`
*   `routerSubresources`
    * `google_compute_router_interface`
    * `google_compute_router_nat`
    * `google_compute_router_peer`
*   `routes`
    * `google_compute_route`
*   `saasRuntime`
    * `google_saas_runtime_release`
    * `google_saas_runtime_rollout_kind`
    * `google_saas_runtime_saas`
    * `google_saas_runtime_tenant`
    * `google_saas_runtime_unit`
    * `google_saas_runtime_unit_kind`
*   `schedulerJobs`
    * `google_cloud_scheduler_job`
*   `scc`
    * `google_scc_event_threat_detection_custom_module`
    * `google_scc_folder_custom_module`
    * `google_scc_folder_notification_config`
    * `google_scc_folder_scc_big_query_export`
    * `google_scc_management_folder_security_health_analytics_custom_module`
    * `google_scc_management_organization_event_threat_detection_custom_module`
    * `google_scc_management_organization_security_health_analytics_custom_module`
    * `google_scc_management_project_security_health_analytics_custom_module`
    * `google_scc_mute_config`
    * `google_scc_notification_config`
    * `google_scc_organization_custom_module`
    * `google_scc_organization_scc_big_query_export`
    * `google_scc_project_custom_module`
    * `google_scc_project_notification_config`
    * `google_scc_project_scc_big_query_export`
    * `google_scc_source`
    * `google_scc_source_iam_binding`
    * `google_scc_source_iam_member`
    * `google_scc_source_iam_policy`
    * `google_scc_v2_folder_mute_config`
    * `google_scc_v2_folder_notification_config`
    * `google_scc_v2_folder_scc_big_query_export`
    * `google_scc_v2_organization_mute_config`
    * `google_scc_v2_organization_notification_config`
    * `google_scc_v2_organization_scc_big_query_export`
    * `google_scc_v2_organization_source`
    * `google_scc_v2_organization_source_iam_binding`
    * `google_scc_v2_organization_source_iam_member`
    * `google_scc_v2_organization_source_iam_policy`
    * `google_scc_v2_project_mute_config`
    * `google_scc_v2_project_notification_config`
    * `google_scc_v2_project_scc_big_query_export`
*   `secretManager`
    * `google_secret_manager_regional_secret`
    * `google_secret_manager_regional_secret_iam_binding`
    * `google_secret_manager_regional_secret_iam_member`
    * `google_secret_manager_regional_secret_iam_policy`
    * `google_secret_manager_regional_secret_version`
    * `google_secret_manager_secret`
    * `google_secret_manager_secret_iam_binding`
    * `google_secret_manager_secret_iam_member`
    * `google_secret_manager_secret_iam_policy`
    * `google_secret_manager_secret_version`
*   `securityPolicies`
    * `google_compute_security_policy`
*   `siteVerification`
    * `google_site_verification_owner`
    * `google_site_verification_web_resource`
*   `securityPosture`
    * `google_securityposture_posture`
    * `google_securityposture_posture_deployment`
*   `serviceDirectory`
    * `google_service_directory_endpoint`
    * `google_service_directory_namespace`
    * `google_service_directory_namespace_iam_binding`
    * `google_service_directory_namespace_iam_member`
    * `google_service_directory_namespace_iam_policy`
    * `google_service_directory_service`
    * `google_service_directory_service_iam_binding`
    * `google_service_directory_service_iam_member`
    * `google_service_directory_service_iam_policy`
*   `serviceNetworking`
    * `google_service_networking_connection`
    * `google_service_networking_peered_dns_domain`
    * `google_service_networking_vpc_service_controls`
*   `snapshots`
    * `google_compute_snapshot`
*   `spanner`
    * `google_spanner_backup_schedule`
    * `google_spanner_database`
    * `google_spanner_database_iam_binding`
    * `google_spanner_database_iam_member`
    * `google_spanner_database_iam_policy`
    * `google_spanner_instance`
    * `google_spanner_instance_config`
    * `google_spanner_instance_iam_binding`
    * `google_spanner_instance_iam_member`
    * `google_spanner_instance_iam_policy`
    * `google_spanner_instance_partition`
*   `sslCertificates`
    * `google_compute_managed_ssl_certificate`
    * `google_compute_ssl_certificate`
*   `sslPolicies`
    * `google_compute_ssl_policy`
*   `storageInsights`
    * `google_storage_insights_dataset_config`
    * `google_storage_insights_report_config`
*   `storageTransfer`
    * `google_storage_transfer_agent_pool`
*   `subnetworks`
    * `google_compute_subnetwork`
*   `targetHttpProxies`
    * `google_compute_target_http_proxy`
*   `targetHttpsProxies`
    * `google_compute_target_https_proxy`
*   `targetInstances`
    * `google_compute_target_instance`
*   `targetPools`
    * `google_compute_target_pool`
*   `targetSslProxies`
    * `google_compute_target_ssl_proxy`
*   `targetTcpProxies`
    * `google_compute_target_tcp_proxy`
*   `targetVpnGateways`
    * `google_compute_vpn_gateway`
*   `transcoder`
    * `google_transcoder_job`
    * `google_transcoder_job_template`
*   `urlMaps`
    * `google_compute_url_map`
*   `vertexAi`
    * `google_vertex_ai_cache_config`
    * `google_vertex_ai_custom_job`
    * `google_vertex_ai_dataset`
    * `google_vertex_ai_deployment_resource_pool`
    * `google_vertex_ai_endpoint`
    * `google_vertex_ai_feature_group`
    * `google_vertex_ai_feature_group_feature`
    * `google_vertex_ai_feature_group_iam_binding`
    * `google_vertex_ai_feature_group_iam_member`
    * `google_vertex_ai_feature_group_iam_policy`
    * `google_vertex_ai_feature_online_store`
    * `google_vertex_ai_feature_online_store_featureview`
    * `google_vertex_ai_feature_online_store_featureview_iam_binding`
    * `google_vertex_ai_feature_online_store_featureview_iam_member`
    * `google_vertex_ai_feature_online_store_featureview_iam_policy`
    * `google_vertex_ai_feature_online_store_iam_binding`
    * `google_vertex_ai_feature_online_store_iam_member`
    * `google_vertex_ai_feature_online_store_iam_policy`
    * `google_vertex_ai_featurestore`
    * `google_vertex_ai_featurestore_entitytype`
    * `google_vertex_ai_featurestore_entitytype_feature`
    * `google_vertex_ai_featurestore_entitytype_iam_binding`
    * `google_vertex_ai_featurestore_entitytype_iam_member`
    * `google_vertex_ai_featurestore_entitytype_iam_policy`
    * `google_vertex_ai_featurestore_iam_binding`
    * `google_vertex_ai_featurestore_iam_member`
    * `google_vertex_ai_featurestore_iam_policy`
    * `google_vertex_ai_index`
    * `google_vertex_ai_index_endpoint`
    * `google_vertex_ai_index_endpoint_deployed_index`
    * `google_vertex_ai_metadata_store`
    * `google_vertex_ai_persistent_resource`
    * `google_vertex_ai_pipeline_job`
    * `google_vertex_ai_rag_corpus`
    * `google_vertex_ai_rag_file`
    * `google_vertex_ai_reasoning_engine`
    * `google_vertex_ai_tensorboard`
    * `google_vertex_ai_tensorboard_experiment`
    * `google_vertex_ai_tensorboard_run`
    * `google_vertex_ai_endpoint_iam_binding`
    * `google_vertex_ai_endpoint_iam_member`
    * `google_vertex_ai_endpoint_iam_policy`
    * `google_vertex_ai_rag_engine_config`
    * `google_vertex_ai_semantic_governance_policy`
    * `google_vertex_ai_semantic_governance_policy_engine`
    * (documented skips: `google_vertex_ai_agent_anomaly_detection_scope`, `google_vertex_ai_evaluation_metric`, `google_vertex_ai_model_development_cluster`, `google_vertex_ai_model_garden_enable_model`, `google_vertex_ai_online_evaluator` — no list/get collection in aiplatform v1/v1beta1; `google_vertex_ai_reasoning_engine_iam` — no getIamPolicy in v1/v1beta1; `google_vertex_ai_endpoint_with_model_garden_deployment` — duplicate of `google_vertex_ai_endpoint` over the same collection)
*   `tpuV2`
    * `google_tpu_v2_queued_resource`
    * `google_tpu_v2_vm`
*   `vmwareEngine`
    * `google_vmwareengine_cluster`
    * `google_vmwareengine_datastore`
    * `google_vmwareengine_external_access_rule`
    * `google_vmwareengine_external_address`
    * `google_vmwareengine_network`
    * `google_vmwareengine_network_peering`
    * `google_vmwareengine_network_policy`
    * `google_vmwareengine_private_cloud`
    * `google_vmwareengine_subnet`
*   `workstations`
    * `google_workstations_workstation`
    * `google_workstations_workstation_cluster`
    * `google_workstations_workstation_config`
    * `google_workstations_workstation_config_iam_binding`
    * `google_workstations_workstation_config_iam_member`
    * `google_workstations_workstation_config_iam_policy`
    * `google_workstations_workstation_iam_binding`
    * `google_workstations_workstation_iam_member`
    * `google_workstations_workstation_iam_policy`
*   `vectorSearch`
    * `google_vector_search_collection`
    * `google_vector_search_data_object`
    * `google_vector_search_index`
*   `vpnTunnels`
    * `google_compute_vpn_tunnel`
*   `activeDirectory`
    * `google_active_directory_domain`
    * `google_active_directory_domain_trust`
    * `google_active_directory_peering`
*   `cloudAsset`
    * `google_cloud_asset_folder_feed`
    * `google_cloud_asset_organization_feed`
    * `google_cloud_asset_project_feed`
*   `cloudIdentity`
    * `google_cloud_identity_group`
    * `google_cloud_identity_group_membership`
    * `google_cloud_identity_policy`
*   `cloudbuildDepth`
    * `google_cloudbuild_bitbucket_server_config`
    * `google_cloudbuild_worker_pool`
    * `google_cloudbuildv2_connection`
    * `google_cloudbuildv2_connection_iam`
    * `google_cloudbuildv2_repository`
*   `composer`
    * `google_composer_environment`
    * `google_composer_user_workloads_config_map`
    * `google_composer_user_workloads_secret`
*   `deploymentManager`
    * `google_deployment_manager_deployment`
*   `eventarc`
    * `google_eventarc_channel`
    * `google_eventarc_enrollment`
    * `google_eventarc_google_api_source`
    * `google_eventarc_google_channel_config`
    * `google_eventarc_message_bus`
    * `google_eventarc_pipeline`
    * `google_eventarc_pipeline_iam`
    * `google_eventarc_trigger`
*   `healthcare`
    * `google_healthcare_consent_store`
    * `google_healthcare_consent_store_iam`
    * `google_healthcare_dataset`
    * `google_healthcare_dataset_iam`
    * `google_healthcare_dicom_store`
    * `google_healthcare_dicom_store_iam`
    * `google_healthcare_fhir_store`
    * `google_healthcare_fhir_store_iam`
    * `google_healthcare_hl7_v2_store`
    * `google_healthcare_hl7_v2_store_iam`
    * `google_healthcare_pipeline_job`
    * `google_healthcare_workspace`
*   `infraManager`
    * `google_config_deployment`
*   `monitoringDepth`
    * `google_monitoring_custom_service`
    * `google_monitoring_dashboard`
    * `google_monitoring_metric_descriptor`
    * `google_monitoring_monitored_project`
    * `google_monitoring_service`
    * `google_monitoring_slo`
    * `google_monitoring_snooze`
*   `observability`
    * `google_observability_bucket`
    * `google_observability_folder_settings`
    * `google_observability_link`
    * `google_observability_organization_settings`
    * `google_observability_project_settings`
    * `google_observability_trace_scope`
*   `oracleDatabase`
    * `google_oracle_database_autonomous_database`
    * `google_oracle_database_cloud_exadata_infrastructure`
    * `google_oracle_database_cloud_vm_cluster`
    * `google_oracle_database_db_system`
    * `google_oracle_database_exadb_vm_cluster`
    * `google_oracle_database_exascale_db_storage_vault`
    * `google_oracle_database_goldengate_connection`
    * `google_oracle_database_goldengate_connection_assignment`
    * `google_oracle_database_goldengate_deployment`
    * `google_oracle_database_odb_network`
    * `google_oracle_database_odb_subnet`
*   `osConfig`
    * `google_os_config_os_policy_assignment`
    * `google_os_config_patch_deployment`
    * `google_os_config_v2_policy_orchestrator`
    * `google_os_config_v2_policy_orchestrator_for_folder`
    * `google_os_config_v2_policy_orchestrator_for_organization`
*   `osLogin`
    * (no resources — `google_os_login_ssh_public_key` is a per-user collection with no project-level list API; see `providers/gcp/oslogin.go`)
*   `pubsubDepth`
    * `google_pubsub_lite_reservation`
    * `google_pubsub_lite_subscription`
    * `google_pubsub_lite_topic`
    * `google_pubsub_schema`
    * `google_pubsub_schema_iam`
    * `google_pubsub_subscription_iam`
    * `google_pubsub_topic_iam`
*   `runtimeConfig`
    * `google_runtimeconfig_config`
    * `google_runtimeconfig_config_iam`
    * `google_runtimeconfig_variable`
*   `sourceRepo`
    * `google_sourcerepo_repository`
    * `google_sourcerepo_repository_iam`
*   `workflows`
    * `google_workflows_workflow`

Additionally, the `logging` service now also emits the folder and
organization variants (`google_logging_folder_sink`,
`google_logging_folder_exclusion`, `google_logging_folder_bucket_config`,
`google_logging_folder_settings`, `google_logging_organization_sink`,
`google_logging_organization_exclusion`,
`google_logging_organization_bucket_config` and
`google_logging_organization_settings`) discovered through project
ancestry, and the `cloudtasks` service emits
`google_cloud_tasks_queue_iam` for queues carrying IAM bindings.

*   `billing`
    * `google_billing_account_iam_binding`
    * `google_billing_account_iam_member`
    * `google_billing_account_iam_policy`
    * `google_billing_budget`
    * `google_billing_project_info`
    * `google_billing_subaccount`
*   `iamOrg`
    * `google_folder`
    * `google_folder_access_approval_settings`
    * `google_folder_iam_binding`
    * `google_folder_iam_member`
    * `google_folder_iam_policy`
    * `google_folder_organization_policy`
    * `google_essential_contacts_contact`
    * `google_iam_deny_policy`
    * `google_iam_folders_policy_binding`
    * `google_iam_organizations_policy_binding`
    * `google_iam_principal_access_boundary_policy`
    * `google_iam_projects_policy_binding`
    * `google_org_policy_custom_constraint`
    * `google_org_policy_policy`
    * `google_organization_access_approval_settings`
    * `google_organization_iam_binding`
    * `google_organization_iam_custom_role`
    * `google_organization_iam_member`
    * `google_organization_iam_policy`
    * `google_organization_policy`
    * `google_project_access_approval_settings`
    * `google_project_iam_binding`
    * `google_project_iam_member`
    * `google_project_iam_policy`
    * `google_project_organization_policy`
    * `google_resource_manager_lien`
    * `google_service_account_iam_binding`
    * `google_service_account_iam_member`
    * `google_service_account_iam_policy`
    * `google_service_account_key`
    * `google_tags_tag_binding`
    * `google_tags_tag_key`
    * `google_tags_tag_key_iam_binding`
    * `google_tags_tag_key_iam_member`
    * `google_tags_tag_key_iam_policy`
    * `google_tags_tag_value`
    * `google_tags_tag_value_iam_binding`
    * `google_tags_tag_value_iam_member`
    * `google_tags_tag_value_iam_policy`
*   `iap`
    * `google_iap_app_engine_service_iam_binding`
    * `google_iap_app_engine_service_iam_member`
    * `google_iap_app_engine_service_iam_policy`
    * `google_iap_app_engine_version_iam_binding`
    * `google_iap_app_engine_version_iam_member`
    * `google_iap_app_engine_version_iam_policy`
    * `google_iap_settings`
    * `google_iap_tunnel_dest_group`
    * `google_iap_tunnel_dest_group_iam_binding`
    * `google_iap_tunnel_dest_group_iam_member`
    * `google_iap_tunnel_dest_group_iam_policy`
    * `google_iap_tunnel_iam_binding`
    * `google_iap_tunnel_iam_member`
    * `google_iap_tunnel_iam_policy`
    * `google_iap_tunnel_instance_iam_binding`
    * `google_iap_tunnel_instance_iam_member`
    * `google_iap_tunnel_instance_iam_policy`
    * `google_iap_web_backend_service_iam_binding`
    * `google_iap_web_backend_service_iam_member`
    * `google_iap_web_backend_service_iam_policy`
    * `google_iap_web_cloud_run_service_iam_binding`
    * `google_iap_web_cloud_run_service_iam_member`
    * `google_iap_web_cloud_run_service_iam_policy`
    * `google_iap_web_forwarding_rule_service_iam_binding`
    * `google_iap_web_forwarding_rule_service_iam_member`
    * `google_iap_web_forwarding_rule_service_iam_policy`
    * `google_iap_web_iam_binding`
    * `google_iap_web_iam_member`
    * `google_iap_web_iam_policy`
    * `google_iap_web_region_backend_service_iam_binding`
    * `google_iap_web_region_backend_service_iam_member`
    * `google_iap_web_region_backend_service_iam_policy`
    * `google_iap_web_region_forwarding_rule_service_iam_binding`
    * `google_iap_web_region_forwarding_rule_service_iam_member`
    * `google_iap_web_region_forwarding_rule_service_iam_policy`
    * `google_iap_web_type_app_engine_iam_binding`
    * `google_iap_web_type_app_engine_iam_member`
    * `google_iap_web_type_app_engine_iam_policy`
    * `google_iap_web_type_compute_iam_binding`
    * `google_iap_web_type_compute_iam_member`
    * `google_iap_web_type_compute_iam_policy`
*   `orgPlatform`
    * `google_apikeys_key`
    * `google_assured_workloads_workload`
    * `google_cloud_quotas_quota_adjuster_settings`
    * `google_cloud_quotas_quota_preference`
    * `google_license_manager_configuration`
    * `google_public_ca_external_account_key`
    * `google_secure_source_manager_branch_rule`
    * `google_secure_source_manager_hook`
    * `google_secure_source_manager_instance`
    * `google_secure_source_manager_instance_iam_binding`
    * `google_secure_source_manager_instance_iam_member`
    * `google_secure_source_manager_instance_iam_policy`
    * `google_secure_source_manager_repository`
    * `google_secure_source_manager_repository_iam_binding`
    * `google_secure_source_manager_repository_iam_member`
    * `google_secure_source_manager_repository_iam_policy`
    * `google_security_scanner_scan_config`
*   `vpcAccess`
    * `google_vpc_access_connector`

Your `tf` and `tfstate` files are written by default to
`generated/gcp/zone/service`.


### Known limits — not importable

Terraformer's GCP coverage is complete for every google provider resource type that can be
discovered by enumeration. The 180 types below are the entire remaining gap: each is provably
not importable by enumeration, for the reason given. The same reasons are documented in the
generator source files under `providers/gcp/`.

* `google_access_context_manager_access_level_condition` — The provider documents "does not support import"; a condition has no ID of its own (its ID is the parent access level's name) and is embedded in the access level.
* `google_access_context_manager_access_levels` — Plural bulk resource duplicating the singular `google_access_context_manager_access_level` over the same API objects.
* `google_access_context_manager_egress_policy` — Embedded in the perimeter's spec/status with the perimeter's own name as ID; does not support import.
* `google_access_context_manager_ingress_policy` — Embedded in the perimeter's spec/status with the perimeter's own name as ID; does not support import.
* `google_access_context_manager_service_perimeter_dry_run_egress_policy` — Embedded in the perimeter's dryRunSpec with the perimeter's own name as ID; does not support import.
* `google_access_context_manager_service_perimeter_dry_run_ingress_policy` — Embedded in the perimeter's dryRunSpec with the perimeter's own name as ID; does not support import.
* `google_access_context_manager_service_perimeter_dry_run_resource` — Embedded in the perimeter's dryRunSpec with the perimeter's own name as ID; does not support import.
* `google_access_context_manager_service_perimeter_egress_policy` — Embedded in the perimeter's spec/status with the perimeter's own name as ID; does not support import.
* `google_access_context_manager_service_perimeter_ingress_policy` — Embedded in the perimeter's spec/status with the perimeter's own name as ID; does not support import.
* `google_access_context_manager_service_perimeter_resource` — Embedded in the perimeter's spec/status with the perimeter's own name as ID; does not support import.
* `google_access_context_manager_service_perimeters` — Plural bulk resource duplicating the singular `google_access_context_manager_service_perimeter` over the same API objects.
* `google_agentic_applications_analyst_agent_persona` — The resource exists only in google-beta; agenticapplications.googleapis.com has no publicly documented discovery document or list endpoint to enumerate personas.
* `google_apigee_env_keystore` — Apigee v1 exposes create/get/delete only for keystores; no list method exists in the generated client or the discovery document.
* `google_apigee_env_references` — Create/get/delete/update only; no list method in the client or the discovery document.
* `google_apigee_environment_keyvaluemaps` — No list method in the client or the discovery document.
* `google_apigee_environment_keyvaluemaps_entries` — Entries have a list call, but their parent map names are undiscoverable because the maps themselves are unlistable.
* `google_apigee_keystores_aliases_key_cert_file` — Hangs off the unlistable keystores.
* `google_apigee_keystores_aliases_pkcs12` — Hangs off the unlistable keystores.
* `google_apigee_keystores_aliases_self_signed_cert` — Hangs off the unlistable keystores.
* `google_apigee_security_feedback` — No such collection exists in the API at all.
* `google_apigee_target_server` — No list method in the client or the discovery document.
* `google_backup_dr_auto_protection_policy` — No autoProtectionPolicies collection exists in the client or the discovery document.
* `google_backup_dr_restore_workload` — A transient `:restore` action on backups; there is no collection.
* `google_backup_dr_service_config` — The API exposes only `initialize` on the singleton; no get or list exists in the client or the discovery document.
* `google_bigquery_analytics_hub_listing_subscription` — Same API object and ID as the emitted `google_bigquery_analytics_hub_data_exchange_subscription` and indistinguishable when enumerated; subscriptions are emitted once, as data exchange subscriptions.
* `google_bigquery_datapolicyv2_data_policy` — The v2 resource addresses the same data policy objects (same names and IDs) as the emitted v1 `google_bigquery_datapolicy_data_policy`; emitting both would double-manage them.
* `google_bigquery_datapolicyv2_data_policy_iam` — Same data policy objects and IDs as the emitted v1 data policy IAM resources; emitting both would double-manage them.
* `google_bigquery_dataset_access` — Access entries have no unique ID (the provider ID is the dataset's own) and the resource does not support import; the entries remain part of the dataset resource.
* `google_bigquery_job` — Transient execution records; jobs.list is run history, not configuration.
* `google_bigtable_gc_policy` — The GC rule is embedded in the table's column families; there is no separate API object or list.
* `google_bigtable_memory_layer` — No memoryLayers surface exists in the pinned bigtableadmin v2 client or discovery document.
* `google_bigtable_schema_bundle` — No schemaBundles surface exists in the pinned client or discovery document.
* `google_chronicle_big_query_export` — SecOps instance not discoverable: the Chronicle client exposes GetInstance only (no ListInstances exists anywhere), the instance ID is an opaque onboarding UUID that cannot be derived from project/region, and the provider has no instance resource to anchor enumeration (see `providers/gcp/chronicle.go`).
* `google_chronicle_case_close_definition` — SecOps instance not discoverable: the Chronicle client exposes GetInstance only (no ListInstances exists anywhere), the instance ID is an opaque onboarding UUID that cannot be derived from project/region, and the provider has no instance resource to anchor enumeration (see `providers/gcp/chronicle.go`).
* `google_chronicle_case_stage_definition` — SecOps instance not discoverable: the Chronicle client exposes GetInstance only (no ListInstances exists anywhere), the instance ID is an opaque onboarding UUID that cannot be derived from project/region, and the provider has no instance resource to anchor enumeration (see `providers/gcp/chronicle.go`).
* `google_chronicle_case_tag_definition` — SecOps instance not discoverable: the Chronicle client exposes GetInstance only (no ListInstances exists anywhere), the instance ID is an opaque onboarding UUID that cannot be derived from project/region, and the provider has no instance resource to anchor enumeration (see `providers/gcp/chronicle.go`).
* `google_chronicle_custom_list` — SecOps instance not discoverable: the Chronicle client exposes GetInstance only (no ListInstances exists anywhere), the instance ID is an opaque onboarding UUID that cannot be derived from project/region, and the provider has no instance resource to anchor enumeration (see `providers/gcp/chronicle.go`).
* `google_chronicle_dashboard_chart` — SecOps instance not discoverable: the Chronicle client exposes GetInstance only (no ListInstances exists anywhere), the instance ID is an opaque onboarding UUID that cannot be derived from project/region, and the provider has no instance resource to anchor enumeration (see `providers/gcp/chronicle.go`).
* `google_chronicle_data_access_label` — SecOps instance not discoverable: the Chronicle client exposes GetInstance only (no ListInstances exists anywhere), the instance ID is an opaque onboarding UUID that cannot be derived from project/region, and the provider has no instance resource to anchor enumeration (see `providers/gcp/chronicle.go`).
* `google_chronicle_data_access_scope` — SecOps instance not discoverable: the Chronicle client exposes GetInstance only (no ListInstances exists anywhere), the instance ID is an opaque onboarding UUID that cannot be derived from project/region, and the provider has no instance resource to anchor enumeration (see `providers/gcp/chronicle.go`).
* `google_chronicle_data_export` — SecOps instance not discoverable: the Chronicle client exposes GetInstance only (no ListInstances exists anywhere), the instance ID is an opaque onboarding UUID that cannot be derived from project/region, and the provider has no instance resource to anchor enumeration (see `providers/gcp/chronicle.go`).
* `google_chronicle_data_table` — SecOps instance not discoverable: the Chronicle client exposes GetInstance only (no ListInstances exists anywhere), the instance ID is an opaque onboarding UUID that cannot be derived from project/region, and the provider has no instance resource to anchor enumeration (see `providers/gcp/chronicle.go`).
* `google_chronicle_data_table_row` — SecOps instance not discoverable: the Chronicle client exposes GetInstance only (no ListInstances exists anywhere), the instance ID is an opaque onboarding UUID that cannot be derived from project/region, and the provider has no instance resource to anchor enumeration (see `providers/gcp/chronicle.go`).
* `google_chronicle_environment` — SecOps instance not discoverable: the Chronicle client exposes GetInstance only (no ListInstances exists anywhere), the instance ID is an opaque onboarding UUID that cannot be derived from project/region, and the provider has no instance resource to anchor enumeration (see `providers/gcp/chronicle.go`).
* `google_chronicle_environment_group` — SecOps instance not discoverable: the Chronicle client exposes GetInstance only (no ListInstances exists anywhere), the instance ID is an opaque onboarding UUID that cannot be derived from project/region, and the provider has no instance resource to anchor enumeration (see `providers/gcp/chronicle.go`).
* `google_chronicle_feed` — SecOps instance not discoverable: the Chronicle client exposes GetInstance only (no ListInstances exists anywhere), the instance ID is an opaque onboarding UUID that cannot be derived from project/region, and the provider has no instance resource to anchor enumeration (see `providers/gcp/chronicle.go`).
* `google_chronicle_findings_refinement` — SecOps instance not discoverable: the Chronicle client exposes GetInstance only (no ListInstances exists anywhere), the instance ID is an opaque onboarding UUID that cannot be derived from project/region, and the provider has no instance resource to anchor enumeration (see `providers/gcp/chronicle.go`).
* `google_chronicle_findings_refinement_deployment` — SecOps instance not discoverable: the Chronicle client exposes GetInstance only (no ListInstances exists anywhere), the instance ID is an opaque onboarding UUID that cannot be derived from project/region, and the provider has no instance resource to anchor enumeration (see `providers/gcp/chronicle.go`).
* `google_chronicle_native_dashboard` — SecOps instance not discoverable: the Chronicle client exposes GetInstance only (no ListInstances exists anywhere), the instance ID is an opaque onboarding UUID that cannot be derived from project/region, and the provider has no instance resource to anchor enumeration (see `providers/gcp/chronicle.go`).
* `google_chronicle_parser` — SecOps instance not discoverable: the Chronicle client exposes GetInstance only (no ListInstances exists anywhere), the instance ID is an opaque onboarding UUID that cannot be derived from project/region, and the provider has no instance resource to anchor enumeration (see `providers/gcp/chronicle.go`).
* `google_chronicle_parser_extension` — SecOps instance not discoverable: the Chronicle client exposes GetInstance only (no ListInstances exists anywhere), the instance ID is an opaque onboarding UUID that cannot be derived from project/region, and the provider has no instance resource to anchor enumeration (see `providers/gcp/chronicle.go`).
* `google_chronicle_reference_list` — SecOps instance not discoverable: the Chronicle client exposes GetInstance only (no ListInstances exists anywhere), the instance ID is an opaque onboarding UUID that cannot be derived from project/region, and the provider has no instance resource to anchor enumeration (see `providers/gcp/chronicle.go`).
* `google_chronicle_retrohunt` — SecOps instance not discoverable: the Chronicle client exposes GetInstance only (no ListInstances exists anywhere), the instance ID is an opaque onboarding UUID that cannot be derived from project/region, and the provider has no instance resource to anchor enumeration (see `providers/gcp/chronicle.go`).
* `google_chronicle_rule` — SecOps instance not discoverable: the Chronicle client exposes GetInstance only (no ListInstances exists anywhere), the instance ID is an opaque onboarding UUID that cannot be derived from project/region, and the provider has no instance resource to anchor enumeration (see `providers/gcp/chronicle.go`).
* `google_chronicle_rule_deployment` — SecOps instance not discoverable: the Chronicle client exposes GetInstance only (no ListInstances exists anywhere), the instance ID is an opaque onboarding UUID that cannot be derived from project/region, and the provider has no instance resource to anchor enumeration (see `providers/gcp/chronicle.go`).
* `google_chronicle_soar_domain` — SecOps instance not discoverable: the Chronicle client exposes GetInstance only (no ListInstances exists anywhere), the instance ID is an opaque onboarding UUID that cannot be derived from project/region, and the provider has no instance resource to anchor enumeration (see `providers/gcp/chronicle.go`).
* `google_chronicle_soar_network` — SecOps instance not discoverable: the Chronicle client exposes GetInstance only (no ListInstances exists anywhere), the instance ID is an opaque onboarding UUID that cannot be derived from project/region, and the provider has no instance resource to anchor enumeration (see `providers/gcp/chronicle.go`).
* `google_chronicle_watchlist` — SecOps instance not discoverable: the Chronicle client exposes GetInstance only (no ListInstances exists anywhere), the instance ID is an opaque onboarding UUID that cannot be derived from project/region, and the provider has no instance resource to anchor enumeration (see `providers/gcp/chronicle.go`).
* `google_cloud_run_v2_instance` — The GA run/v2 client has no instances collection (re-verified against the pinned client; no Instances service).
* `google_cloud_run_v2_instance_iam` — The GA run/v2 client has no instances collection, so there is no parent to attach IAM enumeration to.
* `google_cloud_run_v2_worker_pool_iam` — WorkerPools exposes no GetIamPolicy in the pinned client (re-verified).
* `google_cloud_support_support_event_subscription` — The Cloud Support discovery documents (v2/v2beta) expose only cases, attachments and comments; there is no subscription collection.
* `google_compute_attached_disk` — Facet of the emitted `google_compute_instance` resource; it has no separate identity.
* `google_compute_backend_bucket_signed_url_key` — The provider documents no import support; the key material is secret and returned only at creation.
* `google_compute_backend_service_signed_url_key` — The provider documents no import support; the key material is secret and returned only at creation.
* `google_compute_bulk_per_instance_config` — Bulk view over the same objects emitted individually as `google_compute_per_instance_config`; its provider ID is the managed instance group itself.
* `google_compute_disk_async_replication` — The provider documents no import support; the replication pairing lives in fields on the two emitted disk resources.
* `google_compute_firewall_policy_with_rules` — Same API object and provider ID as the emitted `google_compute_firewall_policy` (the inline-rules alternate management form).
* `google_compute_global_network_endpoint` — Members are identified by IP/port with no stable name; the parent network endpoint groups are emitted.
* `google_compute_instance_from_machine_image` — Creation wrapper; the resulting object is the emitted `google_compute_instance`.
* `google_compute_instance_from_template` — Creation wrapper; the resulting object is the emitted `google_compute_instance`.
* `google_compute_network_endpoint` — Members are identified by IP/port with no stable name; the parent network endpoint groups are emitted.
* `google_compute_network_endpoints` — Members are identified by IP/port with no stable name; the parent network endpoint groups are emitted.
* `google_compute_network_firewall_policy_with_rules` — Same API object and provider ID as the emitted `google_compute_network_firewall_policy` (the inline-rules alternate management form).
* `google_compute_network_peering_routes_config` — Same object and ID as the emitted `google_compute_network_peering`, with overlapping route fields.
* `google_compute_project_metadata_item` — Per-key facet that the provider documents as conflicting with the emitted whole-set `google_compute_project_metadata`.
* `google_compute_region_network_endpoint` — Members are identified by IP/port with no stable name; the parent network endpoint groups are emitted.
* `google_compute_region_network_firewall_policy_with_rules` — Same API object and provider ID as the emitted `google_compute_region_network_firewall_policy` (the inline-rules alternate management form).
* `google_compute_router_named_set` — Not exposed on the GA Router resource in the pinned client (re-verified in Batch 5).
* `google_compute_router_nat_address` — Fields of the emitted router NAT configuration; no separate object exists.
* `google_compute_router_route_policy` — Not exposed on the GA Router resource in the pinned client (re-verified in Batch 5).
* `google_container_registry` — No Container Registry API resource/list/get surface exists; the provider resource is a project singleton whose backing object is the registry's Cloud Storage bucket (already emitted by the `gcs` service), and emitting it unconditionally would claim a registry that may not exist.
* `google_data_catalog_tag_template` — Data Catalog v1 has no tag template list method (get/IAM on a known name only), in the client and the discovery document.
* `google_data_catalog_tag_template_iam` — The parent tag template is undiscoverable (no list method), so its IAM cannot be enumerated.
* `google_data_lineage_config` — Data Lineage v1 exposes only processes, runs, lineage events and link search; no config resource or endpoint exists to read it through.
* `google_data_loss_prevention_content_policy` — No contentPolicy collection exists in the DLP discovery document under projects or organizations.
* `google_dataflow_flex_template_job` — A transient job launch; jobs.list is run history, the provider resource does not document import of an existing job, and the pipeline definition is not recoverable from the job record.
* `google_dataflow_job` — A transient job launch; jobs.list is run history, the provider resource does not document import of an existing job, and the pipeline definition is not recoverable from the job record.
* `google_dataform_folder` — No folders resource exists in the Dataform REST discovery document or pinned client (GAPIC-only surface).
* `google_dataform_team_folder` — No teamFolders resource exists in the Dataform REST discovery document or pinned client (GAPIC-only surface).
* `google_dataplex_data_asset` — Deprecated alias of `google_dataplex_data_product_data_asset` over the same API object and ID.
* `google_dataplex_entry_link` — The Dataplex API has no entry-link list method (create/get/patch/delete only); lookupEntryLinks is a per-entry reverse lookup, not enumeration.
* `google_dataproc_gdc_application_environment` — Dataproc on air-gapped Google Distributed Cloud: a separate endpoint with no public client or discovery document, undiscoverable from cloud project scope.
* `google_dataproc_gdc_service_instance` — Dataproc on air-gapped Google Distributed Cloud: a separate endpoint with no public client or discovery document, undiscoverable from cloud project scope.
* `google_dataproc_gdc_spark_application` — Dataproc on air-gapped Google Distributed Cloud: a separate endpoint with no public client or discovery document, undiscoverable from cloud project scope.
* `google_dataproc_job_iam` — Jobs are transient execution records; jobs.list is run history.
* `google_dataproc_metastore_database_iam` — The Metastore API exposes only getIamPolicy/setIamPolicy on databases; there is no list or get for the parent databases in any version, so they are undiscoverable.
* `google_dataproc_metastore_table_iam` — The Metastore API exposes only getIamPolicy/setIamPolicy on tables; there is no list or get for the parent tables in any version, so they are undiscoverable.
* `google_developer_connect_insights_config` — No insightsConfigs collection exists in the Developer Connect discovery document or client.
* `google_document_ai_schema_version` — The dataset schema is a single versioned document with no addressable version collection (no list/get) in the v1/v1beta3 discovery documents.
* `google_endpoints_service_consumers_iam` — The Service Management API exposes only getIamPolicy/setIamPolicy/testIamPermissions on a service's consumers; there is no list method for a service's consumers, so the consumer projects holding IAM bindings cannot be enumerated.
* `google_firestore_change_stream` — No such resource or collection exists in the Firestore Admin v1 client or discovery document.
* `google_firestore_document` — Data-plane user records: unbounded application data, not configuration.
* `google_folder_service_identity` — Provisioned through the generateServiceIdentity POST; no listable or gettable object exists, and reading it would mutate state.
* `google_hypercomputecluster_cluster` — Pre-GA Cluster Director API: no generated client in the pinned google.golang.org/api and no published discovery document exposing a verifiable list endpoint or provider ID form.
* `google_iam_access_boundary_policy` — Beta access-policy attachments have no list call in iam/v2/v2beta/v3beta and are addressed by undiscoverable user-chosen names; the listable policy definitions and policyBindings are generated.
* `google_iam_folder_access_policy` — Beta access-policy attachments have no list call in iam/v2/v2beta/v3beta and are addressed by undiscoverable user-chosen names; the listable policy definitions and policyBindings are generated.
* `google_iam_organization_access_policy` — Beta access-policy attachments have no list call in iam/v2/v2beta/v3beta and are addressed by undiscoverable user-chosen names; the listable policy definitions and policyBindings are generated.
* `google_iam_project_access_policy` — Beta access-policy attachments have no list call in iam/v2/v2beta/v3beta and are addressed by undiscoverable user-chosen names; the listable policy definitions and policyBindings are generated.
* `google_iam_workforce_pool_provider_scim_tenant` — No SCIM service exists in the pinned iam/v1 client.
* `google_iam_workforce_pool_provider_scim_token` — No SCIM service exists in the pinned iam/v1 client; tokens are also transient creation-time credentials.
* `google_iap_agent_registry_agent_iam` — No Agent Registry service exists in the pinned iap/v1 or v1beta1 clients and no list endpoint exists in the IAP discovery surface.
* `google_iap_agent_registry_endpoint_iam` — No Agent Registry service exists in the pinned iap/v1 or v1beta1 clients and no list endpoint exists in the IAP discovery surface.
* `google_iap_agent_registry_iam` — No Agent Registry service exists in the pinned iap/v1 or v1beta1 clients and no list endpoint exists in the IAP discovery surface.
* `google_iap_agent_registry_mcp_server_iam` — No Agent Registry service exists in the pinned iap/v1 or v1beta1 clients and no list endpoint exists in the IAP discovery surface.
* `google_iap_location_web_iam` — Per-location web policy; iap/v1 exposes no project location list with which to discover locations.
* `google_integrations_client` — The clients resource exposes only provision/deprovision; there is no get or list.
* `google_kms_autokey_config` — The folder-scoped Autokey singleton (`folders/{folder}/autokeyConfig`); a project's folders are not discoverable from the project scope available to this import.
* `google_kms_folder_kaj_policy_config` — The folder-scoped KAJ policy singleton; a project's folders are not discoverable from the project scope available to this import.
* `google_kms_secret_ciphertext` — A local-only helper resource that encrypts a plaintext at apply time; it has no API object, no Cloud KMS ID and no import support — nothing exists server-side to enumerate.
* `google_logging_billing_account_bucket_config` — Billing-account-scoped; a project's ancestry contains folders and an organization, never a billing account, so it cannot be discovered from a project-scoped import.
* `google_logging_billing_account_exclusion` — Billing-account-scoped; a project's ancestry contains folders and an organization, never a billing account, so it cannot be discovered from a project-scoped import.
* `google_logging_billing_account_sink` — Billing-account-scoped; a project's ancestry contains folders and an organization, never a billing account, so it cannot be discovered from a project-scoped import.
* `google_memorystore_instance_desired_user_created_endpoints` — A field of the instance object; the provider resource is an overlay carrying the instance's own ID.
* `google_migration_center_assets_export_job` — No assetsExportJobs collection exists in the pinned migrationcenter v1 client or discovery document.
* `google_network_connectivity_destination` — No multicloud data transfer service exists in the v1 or v1alpha1 generated clients.
* `google_network_connectivity_gateway_advertised_route` — No gateways service exists in the v1 or v1alpha1 generated clients.
* `google_network_connectivity_multicloud_data_transfer_config` — No multicloud data transfer service exists in the v1 or v1alpha1 generated clients.
* `google_network_connectivity_transport` — No transports service exists in the v1 or v1alpha1 generated clients.
* `google_network_security_dns_threat_detector` — No DNS threat detector service exists in the v1 or v1beta1 clients.
* `google_network_security_mirroring_endpoint` — Only mirroring endpoint groups and endpoint group associations exist; there is no standalone endpoint collection.
* `google_network_security_sac_attachment` — No Secure Access Connect services exist in either client.
* `google_network_security_sac_realm` — No Secure Access Connect services exist in either client.
* `google_network_security_ull_mirroring_collector` — No ULL mirroring services exist in either client.
* `google_network_security_ull_mirroring_collector_rule` — No ULL mirroring services exist in either client.
* `google_network_security_ull_mirroring_engine` — No ULL mirroring services exist in either client.
* `google_network_services_agent_connectivity_template` — No agent connectivity services exist in the v1 or v1beta1 generated clients.
* `google_network_services_agent_gateway` — No agent connectivity services exist in the v1 or v1beta1 generated clients.
* `google_network_services_cdn_edge_extension` — No CDN edge extension service exists in either generated client (only `lb_edge_extension`, in v1beta1).
* `google_network_services_edge_cache_keyset` — The EdgeCache services expose only IAM methods (get/set/test IAM policy); there is no list or get surface to enumerate from.
* `google_network_services_edge_cache_origin` — The EdgeCache services expose only IAM methods (get/set/test IAM policy); there is no list or get surface to enumerate from.
* `google_network_services_edge_cache_service` — The EdgeCache services expose only IAM methods (get/set/test IAM policy); there is no list or get surface to enumerate from.
* `google_network_services_multicast_consumer_association` — No multicast services exist in either generated client.
* `google_network_services_multicast_domain` — No multicast services exist in either generated client.
* `google_network_services_multicast_domain_activation` — No multicast services exist in either generated client.
* `google_network_services_multicast_domain_group` — No multicast services exist in either generated client.
* `google_network_services_multicast_group_consumer_activation` — No multicast services exist in either generated client.
* `google_network_services_multicast_group_producer_activation` — No multicast services exist in either generated client.
* `google_network_services_multicast_group_range` — No multicast services exist in either generated client.
* `google_network_services_multicast_group_range_activation` — No multicast services exist in either generated client.
* `google_network_services_multicast_producer_association` — No multicast services exist in either generated client.
* `google_network_services_telemetry_policy` — No telemetry policy service exists in either generated client.
* `google_oracle_database_cloud_exadata_infrastructure_exascale_config` — Not a separate API object: a section of the infrastructure resource, with no dedicated get/list endpoint in the discovery document.
* `google_organization_service_identity` — Provisioned through the generateServiceIdentity POST; no listable or gettable object exists, and reading it would mutate state.
* `google_os_config_guest_policies` — Deprecated plural resource: a project-level aggregate whose provider ID is the project itself, superseded by `google_os_config_os_policy_assignment`; emitting it would double-manage the same settings.
* `google_os_login_ssh_public_key` — A per-user collection (`users/{user}/sshPublicKeys`); oslogin v1 has no project-level list and the user set is not enumerable from a project import.
* `google_privileged_access_manager_settings` — A per-location singleton with no list or get surface in any generated client (none exists for this API).
* `google_project_default_service_accounts` — An action-only helper resource; it has no API object and no provider import.
* `google_project_iam_member_remove` — An action-only helper resource; it has no API object.
* `google_project_service_identity` — Provisioned through the generateServiceIdentity POST; no listable or gettable object exists, and reading it would mutate state.
* `google_redis_cluster_user_created_connections` — Entries of the cluster's own PSC connection configuration; an overlay on `google_redis_cluster` carrying the cluster's ID, with no separate collection.
* `google_resource_manager_boundary` — No Boundaries service exists in the Cloud Resource Manager v1/v3 clients and no raw-REST list endpoint exists in the CRM discovery documents.
* `google_resource_manager_boundary_config` — No Boundaries service exists in the Cloud Resource Manager v1/v3 clients and no raw-REST list endpoint exists in the CRM discovery documents.
* `google_resource_manager_capability` — No Capabilities service exists in the Cloud Resource Manager v1/v3 clients and no raw-REST list endpoint exists in the CRM discovery documents.
* `google_resource_manager_capability_config` — No Capabilities service exists in the Cloud Resource Manager v1/v3 clients and no raw-REST list endpoint exists in the CRM discovery documents.
* `google_saas_runtime_unit_operation` — A transient execution record, not durable configuration.
* `google_scc_notification_service_account` — Backed by a Service Usage generateServiceIdentity provisioning POST; no listable or gettable object exists, and reading it would mutate state.
* `google_scc_v2_organization_scc_big_query_exports` — Deprecated plural duplicate of the singular resource over the same API objects and ID format; emitting both would double-manage them.
* `google_service_usage_consumer_quota_override` — Overrides exist only under fully-qualified metric/limit parents; serviceusage/v1 has no project- or service-level list call.
* `google_service_usage_v2_consumer_policy` — Per-service get-only in v2; there is no list endpoint and no v2 client.
* `google_sql_provision_script` — Not an API object: the provider executes the script text through the Data API at apply time; nothing is stored or listable.
* `google_tags_location_tag_binding` — Location bindings list only per-location through regional CRM endpoints; the pinned CRM v3 client exposes only the global TagBindings list and no location enumeration.
* `google_vertex_ai_agent_anomaly_detection_scope` — No list/get collection exists in aiplatform v1 or v1beta1.
* `google_vertex_ai_endpoint_with_model_garden_deployment` — Duplicate over the same endpoints collection that `google_vertex_ai_endpoint` already enumerates.
* `google_vertex_ai_evaluation_metric` — No list/get collection exists in aiplatform v1 or v1beta1.
* `google_vertex_ai_model_development_cluster` — No list/get collection exists in aiplatform v1 or v1beta1.
* `google_vertex_ai_model_garden_enable_model` — No list/get collection exists in aiplatform v1 or v1beta1.
* `google_vertex_ai_online_evaluator` — No list/get collection exists in aiplatform v1 or v1beta1.
* `google_vertex_ai_reasoning_engine_iam` — No getIamPolicy exists for reasoning engines in v1 or v1beta1 (endpoints only).
* `google_workload_identity_service_agent` — Provisioned through the generateServiceIdentity POST; no listable or gettable object exists, and reading it would mutate state.