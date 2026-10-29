
### Use with AWS

Example:

```
 terraformer import aws --resources=vpc,subnet --connect=true --regions=eu-west-1 --profile=prod
 terraformer import aws --resources=vpc,subnet --filter=vpc=vpc_id1:vpc_id2:vpc_id3 --regions=eu-west-1
```

#### Profiles support

AWS configuration including environmental variables, shared credentials file (\~/.aws/credentials), and shared config file (\~/.aws/config) will be loaded by the tool by default. To use a specific profile, you can use the following command:

```
terraformer import aws --resources=vpc,subnet --regions=eu-west-1 --profile=prod
```

You can also provide no regions when importing resources:
```
terraformer import aws --resources=cloudfront --profile=prod
```
In that case terraformer will not know with which region resources are associated with and will not assume any region. That scenario is useful in case of global resources (e.g. CloudFront distributions or Route 53 records) and when region is passed implicitly through environmental variables or metadata service.

Examples to import other resources-

 * Security Group-
```
terraformer import aws --resources=sg --regions=us-east-1
```



#### Supported services

*   `?None`
    * `aws_pinpoint_adm_channel`
    * `aws_pinpoint_apns_channel`
    * `aws_pinpoint_apns_sandbox_channel`
    * `aws_pinpoint_apns_voip_channel`
    * `aws_pinpoint_apns_voip_sandbox_channel`
    * `aws_pinpoint_baidu_channel`
    * `aws_pinpoint_email_channel`
    * `aws_pinpoint_gcm_channel`
    * `aws_pinpoint_sms_channel`
*   `?Wafv2Generator`
    * `aws_wafv2_web_acl_rule`
    * `aws_wafv2_web_acl_rule_group_association`
*   `accessanalyzer`
    * `aws_accessanalyzer_analyzer`
    * `aws_accessanalyzer_archive_rule`
*   `account`
    * `aws_account_region`
*   `accountaccess`
    * `aws_accountaccess_application`
    * `aws_accountaccess_entitlement`
*   `acm`
    * `aws_acm_certificate`
*   `acmpca`
    * `aws_acmpca_certificate_authority`
    * `aws_acmpca_permission`
    * `aws_acmpca_policy`
*   `alb` (supports ALB and NLB)
    * `aws_lb`
    * `aws_lb_listener`
    * `aws_lb_listener_certificate`
    * `aws_lb_listener_rule`
    * `aws_lb_target_group`
    * `aws_lb_target_group_attachment`
    * `aws_lb_trust_store`
    * `aws_lb_trust_store_revocation`
*   `amplify`
    * `aws_amplify_app`
    * `aws_amplify_backend_environment`
    * `aws_amplify_branch`
    * `aws_amplify_domain_association`
    * `aws_amplify_webhook`
*   `api_gateway`
    * `aws_api_gateway_api_key`
    * `aws_api_gateway_authorizer`
    * `aws_api_gateway_base_path_mapping`
    * `aws_api_gateway_client_certificate`
    * `aws_api_gateway_deployment`
    * `aws_api_gateway_documentation_part`
    * `aws_api_gateway_documentation_version`
    * `aws_api_gateway_domain_name`
    * `aws_api_gateway_domain_name_access_association`
    * `aws_api_gateway_gateway_response`
    * `aws_api_gateway_integration`
    * `aws_api_gateway_integration_response`
    * `aws_api_gateway_method`
    * `aws_api_gateway_method_response`
    * `aws_api_gateway_method_settings`
    * `aws_api_gateway_model`
    * `aws_api_gateway_request_validator`
    * `aws_api_gateway_resource`
    * `aws_api_gateway_rest_api`
    * `aws_api_gateway_rest_api_policy`
    * `aws_api_gateway_stage`
    * `aws_api_gateway_usage_plan`
    * `aws_api_gateway_usage_plan_key`
    * `aws_api_gateway_vpc_link`
*   `api_gatewayv2`
    * `aws_apigatewayv2_api`
    * `aws_apigatewayv2_api_mapping`
    * `aws_apigatewayv2_authorizer`
    * `aws_apigatewayv2_deployment`
    * `aws_apigatewayv2_domain_name`
    * `aws_apigatewayv2_integration`
    * `aws_apigatewayv2_integration_response`
    * `aws_apigatewayv2_model`
    * `aws_apigatewayv2_route`
    * `aws_apigatewayv2_route_response`
    * `aws_apigatewayv2_routing_rule`
    * `aws_apigatewayv2_stage`
    * `aws_apigatewayv2_vpc_link`
*   `appautoscaling`
    * `aws_appautoscaling_policy`
    * `aws_appautoscaling_target`
*   `appconfig`
    * `aws_appconfig_application`
    * `aws_appconfig_configuration_profile`
    * `aws_appconfig_deployment`
    * `aws_appconfig_deployment_strategy`
    * `aws_appconfig_environment`
    * `aws_appconfig_extension`
    * `aws_appconfig_extension_association`
    * `aws_appconfig_hosted_configuration_version`
*   `appfabric`
    * `aws_appfabric_app_bundle`
    * `aws_appfabric_ingestion`
*   `appflow`
    * `aws_appflow_connector_profile`
    * `aws_appflow_flow`
*   `appintegrations`
    * `aws_appintegrations_data_integration`
    * `aws_appintegrations_event_integration`
*   `applicationinsights`
    * `aws_applicationinsights_application`
*   `appmesh`
    * `aws_appmesh_gateway_route`
    * `aws_appmesh_mesh`
    * `aws_appmesh_route`
    * `aws_appmesh_virtual_gateway`
    * `aws_appmesh_virtual_node`
    * `aws_appmesh_virtual_router`
    * `aws_appmesh_virtual_service`
*   `apprunner`
    * `aws_apprunner_auto_scaling_configuration_version`
    * `aws_apprunner_connection`
    * `aws_apprunner_custom_domain_association`
    * `aws_apprunner_observability_configuration`
    * `aws_apprunner_service`
    * `aws_apprunner_vpc_connector`
    * `aws_apprunner_vpc_ingress_connection`
*   `appstream`
    * `aws_appstream_directory_config`
    * `aws_appstream_fleet`
    * `aws_appstream_image_builder`
    * `aws_appstream_stack`
    * `aws_appstream_user`
    * `aws_appstream_user_stack_association`
*   `appsync`
    * `aws_appsync_api`
    * `aws_appsync_api_cache`
    * `aws_appsync_api_key`
    * `aws_appsync_channel_namespace`
    * `aws_appsync_datasource`
    * `aws_appsync_domain_name`
    * `aws_appsync_domain_name_api_association`
    * `aws_appsync_function`
    * `aws_appsync_graphql_api`
    * `aws_appsync_resolver`
    * `aws_appsync_source_api_association`
    * `aws_appsync_type`
*   `arcregionswitch`
    * `aws_arcregionswitch_plan`
*   `arczonalshift`
    * `aws_arczonalshift_zonal_autoshift_configuration`
*   `athena`
    * `aws_athena_capacity_reservation`
    * `aws_athena_data_catalog`
    * `aws_athena_database`
    * `aws_athena_named_query`
    * `aws_athena_prepared_statement`
    * `aws_athena_workgroup`
*   `auditmanager`
    * `aws_auditmanager_assessment`
    * `aws_auditmanager_assessment_delegation`
    * `aws_auditmanager_assessment_report`
    * `aws_auditmanager_control`
    * `aws_auditmanager_framework`
    * `aws_auditmanager_framework_share`
*   `auto_scaling`
    * `aws_autoscaling_group`
    * `aws_autoscaling_lifecycle_hook`
    * `aws_autoscaling_policy`
    * `aws_autoscaling_schedule`
    * `aws_autoscalingplans_scaling_plan`
    * `aws_launch_configuration`
    * `aws_launch_template`
*   `backup`
    * `aws_backup_framework`
    * `aws_backup_logically_air_gapped_vault`
    * `aws_backup_plan`
    * `aws_backup_report_plan`
    * `aws_backup_restore_testing_plan`
    * `aws_backup_restore_testing_selection`
    * `aws_backup_selection`
    * `aws_backup_vault`
    * `aws_backup_vault_lock_configuration`
    * `aws_backup_vault_notifications`
    * `aws_backup_vault_policy`
*   `batch`
    * `aws_batch_compute_environment`
    * `aws_batch_job_definition`
    * `aws_batch_job_queue`
    * `aws_batch_scheduling_policy`
*   `bcmdataexports`
    * `aws_bcmdataexports_export`
*   `bedrock`
    * `aws_bedrock_custom_model`
    * `aws_bedrock_enforced_guardrail_configuration`
    * `aws_bedrock_evaluation_job`
    * `aws_bedrock_guardrail`
    * `aws_bedrock_inference_profile`
    * `aws_bedrock_model_invocation_job`
    * `aws_bedrock_model_invocation_logging_configuration`
    * `aws_bedrock_provisioned_model_throughput`
*   `bedrockagent`
    * `aws_bedrockagent_agent`
    * `aws_bedrockagent_agent_action_group`
    * `aws_bedrockagent_agent_alias`
    * `aws_bedrockagent_agent_collaborator`
    * `aws_bedrockagent_agent_knowledge_base_association`
    * `aws_bedrockagent_data_source`
    * `aws_bedrockagent_flow`
    * `aws_bedrockagent_knowledge_base`
    * `aws_bedrockagent_prompt`
*   `bedrockagentcore`
    * `aws_bedrockagentcore_agent_runtime`
    * `aws_bedrockagentcore_agent_runtime_endpoint`
    * `aws_bedrockagentcore_api_key_credential_provider`
    * `aws_bedrockagentcore_browser`
    * `aws_bedrockagentcore_browser_profile`
    * `aws_bedrockagentcore_code_interpreter`
    * `aws_bedrockagentcore_evaluator`
    * `aws_bedrockagentcore_gateway`
    * `aws_bedrockagentcore_gateway_rule`
    * `aws_bedrockagentcore_gateway_target`
    * `aws_bedrockagentcore_harness`
    * `aws_bedrockagentcore_memory`
    * `aws_bedrockagentcore_memory_strategy`
    * `aws_bedrockagentcore_oauth2_credential_provider`
    * `aws_bedrockagentcore_online_evaluation_config`
    * `aws_bedrockagentcore_policy`
    * `aws_bedrockagentcore_policy_engine`
    * `aws_bedrockagentcore_registry`
    * `aws_bedrockagentcore_token_vault_cmk`
    * `aws_bedrockagentcore_workload_identity`
*   `billing`
    * `aws_billing_view`
*   `budgets`
    * `aws_budgets_budget`
    * `aws_budgets_budget_action`
*   `chatbot`
    * `aws_chatbot_slack_channel_configuration`
    * `aws_chatbot_teams_channel_configuration`
*   `chime`
    * `aws_chime_voice_connector`
    * `aws_chime_voice_connector_group`
    * `aws_chime_voice_connector_logging`
    * `aws_chime_voice_connector_origination`
    * `aws_chime_voice_connector_streaming`
    * `aws_chime_voice_connector_termination`
    * `aws_chimesdkvoice_sip_media_application`
    * `aws_chimesdkvoice_sip_rule`
    * `aws_chimesdkvoice_voice_profile_domain`
*   `chimesdkmediapipelines`
    * `aws_chimesdkmediapipelines_media_insights_pipeline_configuration`
*   `cleanrooms`
    * `aws_cleanrooms_collaboration`
    * `aws_cleanrooms_configured_table`
    * `aws_cleanrooms_membership`
*   `client_vpn`
    * `aws_ec2_client_vpn_authorization_rule`
    * `aws_ec2_client_vpn_endpoint`
    * `aws_ec2_client_vpn_network_association`
    * `aws_ec2_client_vpn_route`
*   `cloud9`
    * `aws_cloud9_environment_ec2`
    * `aws_cloud9_environment_membership`
*   `cloudformation`
    * `aws_cloudformation_stack`
    * `aws_cloudformation_stack_instances`
    * `aws_cloudformation_stack_set`
    * `aws_cloudformation_stack_set_instance`
    * `aws_cloudformation_type`
*   `cloudfront`
    * `aws_cloudfront_anycast_ip_list`
    * `aws_cloudfront_cache_policy`
    * `aws_cloudfront_connection_function`
    * `aws_cloudfront_connection_group`
    * `aws_cloudfront_continuous_deployment_policy`
    * `aws_cloudfront_distribution`
    * `aws_cloudfront_distribution_tenant`
    * `aws_cloudfront_field_level_encryption_config`
    * `aws_cloudfront_field_level_encryption_profile`
    * `aws_cloudfront_function`
    * `aws_cloudfront_key_group`
    * `aws_cloudfront_key_value_store`
    * `aws_cloudfront_monitoring_subscription`
    * `aws_cloudfront_multitenant_distribution`
    * `aws_cloudfront_origin_access_control`
    * `aws_cloudfront_origin_access_identity`
    * `aws_cloudfront_origin_request_policy`
    * `aws_cloudfront_public_key`
    * `aws_cloudfront_realtime_log_config`
    * `aws_cloudfront_response_headers_policy`
    * `aws_cloudfront_trust_store`
    * `aws_cloudfront_vpc_origin`
*   `cloudfrontkeyvaluestore`
    * `aws_cloudfrontkeyvaluestore_key`
*   `cloudhsm`
    * `aws_cloudhsm_v2_cluster`
    * `aws_cloudhsm_v2_hsm`
*   `cloudsearch`
    * `aws_cloudsearch_domain`
    * `aws_cloudsearch_domain_service_access_policy`
*   `cloudtrail`
    * `aws_cloudtrail`
    * `aws_cloudtrail_event_data_store`
    * `aws_cloudtrail_organization_delegated_admin_account`
*   `cloudwatch`
    * `aws_cloudwatch_alarm_mute_rule`
    * `aws_cloudwatch_composite_alarm`
    * `aws_cloudwatch_contributor_insight_rule`
    * `aws_cloudwatch_dashboard`
    * `aws_cloudwatch_event_api_destination`
    * `aws_cloudwatch_event_archive`
    * `aws_cloudwatch_event_bus`
    * `aws_cloudwatch_event_bus_policy`
    * `aws_cloudwatch_event_connection`
    * `aws_cloudwatch_event_rule`
    * `aws_cloudwatch_event_target`
    * `aws_cloudwatch_metric_alarm`
    * `aws_cloudwatch_metric_stream`
*   `cloudwatchomni`
    * `aws_cloudwatchomni_access_grant`
    * `aws_cloudwatchomni_access_profile`
    * `aws_cloudwatchomni_dashboard`
    * `aws_cloudwatchomni_domain`
    * `aws_cloudwatchomni_space`
    * `aws_cloudwatchomni_view`
*   `codeartifact`
    * `aws_codeartifact_domain`
    * `aws_codeartifact_repository`
*   `codebuild`
    * `aws_codebuild_fleet`
    * `aws_codebuild_project`
    * `aws_codebuild_report_group`
    * `aws_codebuild_source_credential`
    * `aws_codebuild_webhook`
*   `codecommit`
    * `aws_codecommit_approval_rule_template`
    * `aws_codecommit_repository`
*   `codeconnections`
    * `aws_codeconnections_connection`
    * `aws_codeconnections_host`
*   `codedeploy`
    * `aws_codedeploy_app`
    * `aws_codedeploy_deployment_config`
    * `aws_codedeploy_deployment_group`
*   `codeguruprofiler`
    * `aws_codeguruprofiler_profiling_group`
*   `codepipeline`
    * `aws_codepipeline`
    * `aws_codepipeline_custom_action_type`
    * `aws_codepipeline_webhook`
*   `codestarnotifications`
    * `aws_codestarnotifications_notification_rule`
*   `cognito`
    * `aws_cognito_identity_pool`
    * `aws_cognito_identity_pool_roles_attachment`
    * `aws_cognito_identity_provider`
    * `aws_cognito_log_delivery_configuration`
    * `aws_cognito_managed_login_branding`
    * `aws_cognito_resource_server`
    * `aws_cognito_risk_configuration`
    * `aws_cognito_user`
    * `aws_cognito_user_group`
    * `aws_cognito_user_pool`
    * `aws_cognito_user_pool_client`
    * `aws_cognito_user_pool_domain`
*   `comprehend`
    * `aws_comprehend_document_classifier`
    * `aws_comprehend_entity_recognizer`
*   `config`
    * `aws_config_aggregate_authorization`
    * `aws_config_config_rule`
    * `aws_config_configuration_aggregator`
    * `aws_config_configuration_recorder`
    * `aws_config_configuration_recorder_status`
    * `aws_config_conformance_pack`
    * `aws_config_delivery_channel`
    * `aws_config_organization_conformance_pack`
    * `aws_config_organization_custom_policy_rule`
    * `aws_config_organization_custom_rule`
    * `aws_config_organization_managed_rule`
    * `aws_config_remediation_configuration`
    * `aws_config_retention_configuration`
*   `connect`
    * `aws_connect_bot_association`
    * `aws_connect_contact_flow`
    * `aws_connect_contact_flow_module`
    * `aws_connect_hours_of_operation`
    * `aws_connect_instance`
    * `aws_connect_instance_storage_config`
    * `aws_connect_lambda_function_association`
    * `aws_connect_phone_number`
    * `aws_connect_phone_number_contact_flow_association`
    * `aws_connect_queue`
    * `aws_connect_quick_connect`
    * `aws_connect_routing_profile`
    * `aws_connect_security_profile`
    * `aws_connect_user`
    * `aws_connect_user_hierarchy_group`
    * `aws_connect_user_hierarchy_structure`
    * `aws_connect_vocabulary`
*   `controltower`
    * `aws_controltower_baseline`
    * `aws_controltower_control`
    * `aws_controltower_landing_zone`
*   `costexplorer`
    * `aws_ce_anomaly_monitor`
    * `aws_ce_anomaly_subscription`
    * `aws_ce_cost_allocation_tag`
    * `aws_ce_cost_category`
*   `customer_gateway`
    * `aws_customer_gateway`
*   `customerprofiles`
    * `aws_customerprofiles_domain`
*   `dataexchange`
    * `aws_dataexchange_data_set`
    * `aws_dataexchange_event_action`
    * `aws_dataexchange_revision`
*   `datapipeline`
    * `aws_datapipeline_pipeline`
    * `aws_datapipeline_pipeline_definition`
*   `datasync`
    * `aws_datasync_agent`
    * `aws_datasync_location_azure_blob`
    * `aws_datasync_location_efs`
    * `aws_datasync_location_fsx_lustre_file_system`
    * `aws_datasync_location_fsx_ontap_file_system`
    * `aws_datasync_location_fsx_openzfs_file_system`
    * `aws_datasync_location_fsx_windows_file_system`
    * `aws_datasync_location_hdfs`
    * `aws_datasync_location_nfs`
    * `aws_datasync_location_object_storage`
    * `aws_datasync_location_s3`
    * `aws_datasync_location_smb`
    * `aws_datasync_task`
*   `datazone`
    * `aws_datazone_asset_type`
    * `aws_datazone_domain`
    * `aws_datazone_environment`
    * `aws_datazone_environment_blueprint_configuration`
    * `aws_datazone_environment_profile`
    * `aws_datazone_form_type`
    * `aws_datazone_glossary`
    * `aws_datazone_glossary_term`
    * `aws_datazone_project`
    * `aws_datazone_user_profile`
*   `dax`
    * `aws_dax_cluster`
    * `aws_dax_parameter_group`
    * `aws_dax_subnet_group`
*   `detective`
    * `aws_detective_graph`
    * `aws_detective_member`
*   `devicefarm`
    * `aws_devicefarm_device_pool`
    * `aws_devicefarm_instance_profile`
    * `aws_devicefarm_network_profile`
    * `aws_devicefarm_project`
    * `aws_devicefarm_test_grid_project`
    * `aws_devicefarm_upload`
*   `devopsagent`
    * `aws_devopsagent_agent_space`
    * `aws_devopsagent_asset`
*   `devopsguru`
    * `aws_devopsguru_notification_channel`
    * `aws_devopsguru_resource_collection`
*   `directoryservice`
    * `aws_directory_service_conditional_forwarder`
    * `aws_directory_service_directory`
    * `aws_directory_service_directory_setting`
    * `aws_directory_service_ip_route`
    * `aws_directory_service_log_subscription`
    * `aws_directory_service_radius_settings`
    * `aws_directory_service_region`
    * `aws_directory_service_shared_directory`
    * `aws_directory_service_trust`
    * `aws_directoryservicedata_group`
    * `aws_directoryservicedata_user`
*   `dlm`
    * `aws_dlm_lifecycle_policy`
*   `dms`
    * `aws_dms_certificate`
    * `aws_dms_data_provider`
    * `aws_dms_endpoint`
    * `aws_dms_event_subscription`
    * `aws_dms_instance_profile`
    * `aws_dms_migration_project`
    * `aws_dms_replication_config`
    * `aws_dms_replication_instance`
    * `aws_dms_replication_subnet_group`
    * `aws_dms_replication_task`
    * `aws_dms_s3_endpoint`
*   `docdb`
    * `aws_docdb_cluster`
    * `aws_docdb_cluster_instance`
    * `aws_docdb_cluster_parameter_group`
    * `aws_docdb_cluster_snapshot`
    * `aws_docdb_event_subscription`
    * `aws_docdb_global_cluster`
    * `aws_docdb_subnet_group`
    * `aws_docdbelastic_cluster`
*   `drs`
    * `aws_drs_replication_configuration_template`
*   `dsql`
    * `aws_dsql_cluster`
    * `aws_dsql_cluster_peering`
    * `aws_dsql_cluster_policy`
*   `dx`
    * `aws_dx_connection`
    * `aws_dx_gateway`
    * `aws_dx_gateway_association`
    * `aws_dx_gateway_association_proposal`
    * `aws_dx_hosted_private_virtual_interface`
    * `aws_dx_hosted_private_virtual_interface_accepter`
    * `aws_dx_hosted_public_virtual_interface`
    * `aws_dx_hosted_public_virtual_interface_accepter`
    * `aws_dx_hosted_transit_virtual_interface`
    * `aws_dx_hosted_transit_virtual_interface_accepter`
    * `aws_dx_lag`
    * `aws_dx_private_virtual_interface`
    * `aws_dx_public_virtual_interface`
    * `aws_dx_transit_virtual_interface`
*   `dynamodb`
    * `aws_dynamodb_contributor_insights`
    * `aws_dynamodb_global_secondary_index`
    * `aws_dynamodb_global_table`
    * `aws_dynamodb_kinesis_streaming_destination`
    * `aws_dynamodb_resource_policy`
    * `aws_dynamodb_table`
    * `aws_dynamodb_table_export`
*   `ebs`
    * `aws_ebs_fast_snapshot_restore`
    * `aws_ebs_snapshot`
    * `aws_ebs_volume`
    * `aws_volume_attachment`
*   `ec2_instance`
    * `aws_ami`
    * `aws_ami_launch_permission`
    * `aws_ec2_availability_zone_group`
    * `aws_ec2_capacity_reservation`
    * `aws_ec2_default_credit_specification`
    * `aws_ec2_fleet`
    * `aws_ec2_host`
    * `aws_ec2_instance_connect_endpoint`
    * `aws_ec2_managed_prefix_list`
    * `aws_ec2_managed_prefix_list_entry`
    * `aws_ec2_network_insights_access_scope`
    * `aws_ec2_network_insights_analysis`
    * `aws_ec2_network_insights_path`
    * `aws_ec2_secondary_network`
    * `aws_ec2_secondary_subnet`
    * `aws_ec2_traffic_mirror_filter`
    * `aws_ec2_traffic_mirror_filter_rule`
    * `aws_ec2_traffic_mirror_session`
    * `aws_ec2_traffic_mirror_target`
    * `aws_instance`
    * `aws_key_pair`
    * `aws_placement_group`
    * `aws_spot_datafeed_subscription`
    * `aws_spot_fleet_request`
*   `ecr`
    * `aws_ecr_lifecycle_policy`
    * `aws_ecr_pull_through_cache_rule`
    * `aws_ecr_pull_time_update_exclusion`
    * `aws_ecr_repository`
    * `aws_ecr_repository_creation_template`
    * `aws_ecr_repository_policy`
*   `ecrpublic`
    * `aws_ecrpublic_repository`
    * `aws_ecrpublic_repository_policy`
*   `ecs`
    * `aws_ecs_capacity_provider`
    * `aws_ecs_cluster`
    * `aws_ecs_cluster_capacity_providers`
    * `aws_ecs_daemon`
    * `aws_ecs_daemon_task_definition`
    * `aws_ecs_express_gateway_service`
    * `aws_ecs_service`
    * `aws_ecs_task_definition`
    * `aws_ecs_task_set`
*   `efs`
    * `aws_efs_access_point`
    * `aws_efs_backup_policy`
    * `aws_efs_file_system`
    * `aws_efs_file_system_policy`
    * `aws_efs_mount_target`
    * `aws_efs_replication_configuration`
*   `eip`
    * `aws_eip`
    * `aws_eip_association`
    * `aws_eip_domain_name`
*   `eks`
    * `aws_eks_access_entry`
    * `aws_eks_access_policy_association`
    * `aws_eks_addon`
    * `aws_eks_capability`
    * `aws_eks_cluster`
    * `aws_eks_fargate_profile`
    * `aws_eks_identity_provider_config`
    * `aws_eks_node_group`
    * `aws_eks_pod_identity_association`
*   `elastic_beanstalk`
    * `aws_elastic_beanstalk_application`
    * `aws_elastic_beanstalk_environment`
*   `elasticache`
    * `aws_elasticache_cluster`
    * `aws_elasticache_global_replication_group`
    * `aws_elasticache_parameter_group`
    * `aws_elasticache_replication_group`
    * `aws_elasticache_reserved_cache_node`
    * `aws_elasticache_serverless_cache`
    * `aws_elasticache_subnet_group`
    * `aws_elasticache_user`
    * `aws_elasticache_user_group`
    * `aws_elasticache_user_group_association`
*   `elastictranscoder`
    * `aws_elastictranscoder_pipeline`
    * `aws_elastictranscoder_preset`
*   `elb`
    * `aws_app_cookie_stickiness_policy`
    * `aws_elb`
*   `emr`
    * `aws_emr_cluster`
    * `aws_emr_instance_fleet`
    * `aws_emr_instance_group`
    * `aws_emr_managed_scaling_policy`
    * `aws_emr_security_configuration`
    * `aws_emr_studio`
    * `aws_emr_studio_session_mapping`
    * `aws_emrcontainers_job_template`
    * `aws_emrcontainers_virtual_cluster`
    * `aws_emrserverless_application`
*   `eni`
    * `aws_network_interface`
    * `aws_network_interface_attachment`
    * `aws_network_interface_permission`
    * `aws_network_interface_sg_attachment`
*   `es`
    * `aws_elasticsearch_domain`
    * `aws_elasticsearch_domain_policy`
    * `aws_elasticsearch_vpc_endpoint`
*   `evidently`
    * `aws_evidently_feature`
    * `aws_evidently_launch`
    * `aws_evidently_project`
    * `aws_evidently_segment`
*   `finspace`
    * `aws_finspace_kx_cluster`
    * `aws_finspace_kx_database`
    * `aws_finspace_kx_dataview`
    * `aws_finspace_kx_environment`
    * `aws_finspace_kx_scaling_group`
    * `aws_finspace_kx_user`
    * `aws_finspace_kx_volume`
*   `firehose`
    * `aws_kinesis_firehose_delivery_stream`
*   `fis`
    * `aws_fis_experiment_template`
    * `aws_fis_target_account_configuration`
*   `fms`
    * `aws_fms_policy`
    * `aws_fms_resource_set`
*   `fsx`
    * `aws_fsx_backup`
    * `aws_fsx_data_repository_association`
    * `aws_fsx_file_cache`
    * `aws_fsx_lustre_file_system`
    * `aws_fsx_ontap_file_system`
    * `aws_fsx_ontap_storage_virtual_machine`
    * `aws_fsx_ontap_volume`
    * `aws_fsx_openzfs_file_system`
    * `aws_fsx_openzfs_snapshot`
    * `aws_fsx_openzfs_volume`
    * `aws_fsx_s3_access_point_attachment`
    * `aws_fsx_windows_file_system`
*   `gamelift`
    * `aws_gamelift_alias`
    * `aws_gamelift_build`
    * `aws_gamelift_fleet`
    * `aws_gamelift_game_server_group`
    * `aws_gamelift_script`
*   `glacier`
    * `aws_glacier_vault`
    * `aws_glacier_vault_lock`
*   `globalaccelerator`
    * `aws_globalaccelerator_accelerator`
    * `aws_globalaccelerator_cross_account_attachment`
    * `aws_globalaccelerator_custom_routing_accelerator`
    * `aws_globalaccelerator_custom_routing_endpoint_group`
    * `aws_globalaccelerator_custom_routing_listener`
    * `aws_globalaccelerator_endpoint_group`
    * `aws_globalaccelerator_listener`
*   `glue`
    * `aws_glue_catalog_database`
    * `aws_glue_catalog_table`
    * `aws_glue_catalog_table_optimizer`
    * `aws_glue_classifier`
    * `aws_glue_connection`
    * `aws_glue_crawler`
    * `aws_glue_data_quality_ruleset`
    * `aws_glue_dev_endpoint`
    * `aws_glue_job`
    * `aws_glue_ml_transform`
    * `aws_glue_partition_index`
    * `aws_glue_registry`
    * `aws_glue_schema`
    * `aws_glue_security_configuration`
    * `aws_glue_trigger`
    * `aws_glue_user_defined_function`
    * `aws_glue_workflow`
*   `grafana`
    * `aws_grafana_license_association`
    * `aws_grafana_workspace`
    * `aws_grafana_workspace_saml_configuration`
    * `aws_grafana_workspace_service_account`
*   `guardduty`
    * `aws_guardduty_detector`
    * `aws_guardduty_filter`
    * `aws_guardduty_ipset`
    * `aws_guardduty_malware_protection_plan`
    * `aws_guardduty_member`
    * `aws_guardduty_publishing_destination`
    * `aws_guardduty_threatintelset`
*   `iam`
    * `aws_iam_access_key`
    * `aws_iam_account_alias`
    * `aws_iam_group`
    * `aws_iam_group_membership`
    * `aws_iam_group_policy`
    * `aws_iam_group_policy_attachment`
    * `aws_iam_instance_profile`
    * `aws_iam_openid_connect_provider`
    * `aws_iam_policy`
    * `aws_iam_role`
    * `aws_iam_role_policy`
    * `aws_iam_role_policy_attachment`
    * `aws_iam_saml_provider`
    * `aws_iam_server_certificate`
    * `aws_iam_service_linked_role`
    * `aws_iam_service_specific_credential`
    * `aws_iam_signing_certificate`
    * `aws_iam_user`
    * `aws_iam_user_group_membership`
    * `aws_iam_user_login_profile`
    * `aws_iam_user_policy`
    * `aws_iam_user_policy_attachment`
    * `aws_iam_user_ssh_key`
    * `aws_iam_virtual_mfa_device`
*   `identitystore`
    * `aws_identitystore_group`
    * `aws_identitystore_group_membership`
    * `aws_identitystore_user`
*   `igw`
    * `aws_ec2_carrier_gateway`
    * `aws_egress_only_internet_gateway`
    * `aws_internet_gateway`
    * `aws_internet_gateway_attachment`
*   `imagebuilder`
    * `aws_imagebuilder_component`
    * `aws_imagebuilder_container_recipe`
    * `aws_imagebuilder_distribution_configuration`
    * `aws_imagebuilder_image`
    * `aws_imagebuilder_image_pipeline`
    * `aws_imagebuilder_image_recipe`
    * `aws_imagebuilder_infrastructure_configuration`
    * `aws_imagebuilder_lifecycle_policy`
    * `aws_imagebuilder_workflow`
*   `inspector`
    * `aws_inspector_assessment_target`
    * `aws_inspector_assessment_template`
*   `inspector2`
    * `aws_inspector2_filter`
    * `aws_inspector2_member_association`
*   `internetmonitor`
    * `aws_internetmonitor_monitor`
*   `invoicing`
    * `aws_invoicing_invoice_unit`
*   `iot`
    * `aws_iot_authorizer`
    * `aws_iot_billing_group`
    * `aws_iot_ca_certificate`
    * `aws_iot_certificate`
    * `aws_iot_domain_configuration`
    * `aws_iot_event_configurations`
    * `aws_iot_policy`
    * `aws_iot_provisioning_template`
    * `aws_iot_role_alias`
    * `aws_iot_thing`
    * `aws_iot_thing_group`
    * `aws_iot_thing_type`
    * `aws_iot_topic_rule`
    * `aws_iot_topic_rule_destination`
*   `ivs`
    * `aws_ivs_channel`
    * `aws_ivs_playback_key_pair`
    * `aws_ivs_recording_configuration`
    * `aws_ivschat_logging_configuration`
    * `aws_ivschat_room`
*   `kendra`
    * `aws_kendra_data_source`
    * `aws_kendra_experience`
    * `aws_kendra_faq`
    * `aws_kendra_index`
    * `aws_kendra_query_suggestions_block_list`
    * `aws_kendra_thesaurus`
*   `keyspaces`
    * `aws_keyspaces_keyspace`
    * `aws_keyspaces_table`
*   `kinesis`
    * `aws_kinesis_analytics_application`
    * `aws_kinesis_resource_policy`
    * `aws_kinesis_stream`
    * `aws_kinesis_stream_consumer`
    * `aws_kinesis_video_stream`
    * `aws_kinesisanalyticsv2_application`
*   `kms`
    * `aws_kms_alias`
    * `aws_kms_custom_key_store`
    * `aws_kms_external_key`
    * `aws_kms_grant`
    * `aws_kms_key`
    * `aws_kms_key_policy`
    * `aws_kms_replica_external_key`
    * `aws_kms_replica_key`
*   `lakeformation`
    * `aws_lakeformation_data_cells_filter`
    * `aws_lakeformation_lf_tag`
    * `aws_lakeformation_lf_tag_expression`
    * `aws_lakeformation_opt_in`
    * `aws_lakeformation_resource`
*   `lambda`
    * `aws_lambda_alias`
    * `aws_lambda_capacity_provider`
    * `aws_lambda_code_signing_config`
    * `aws_lambda_event_source_mapping`
    * `aws_lambda_function`
    * `aws_lambda_function_event_invoke_config`
    * `aws_lambda_function_recursion_config`
    * `aws_lambda_function_scaling_config`
    * `aws_lambda_function_url`
    * `aws_lambda_layer_version`
    * `aws_lambda_layer_version_permission`
    * `aws_lambda_permission`
    * `aws_lambda_provisioned_concurrency_config`
    * `aws_lambda_runtime_management_config`
*   `lex`
    * `aws_lex_bot`
    * `aws_lex_bot_alias`
    * `aws_lex_intent`
    * `aws_lex_slot_type`
    * `aws_lexv2models_bot`
    * `aws_lexv2models_bot_locale`
    * `aws_lexv2models_bot_version`
    * `aws_lexv2models_intent`
    * `aws_lexv2models_slot`
    * `aws_lexv2models_slot_type`
*   `licensemanager`
    * `aws_licensemanager_grant`
    * `aws_licensemanager_license_configuration`
*   `lightsail`
    * `aws_lightsail_bucket`
    * `aws_lightsail_bucket_access_key`
    * `aws_lightsail_bucket_resource_access`
    * `aws_lightsail_certificate`
    * `aws_lightsail_container_service`
    * `aws_lightsail_container_service_deployment_version`
    * `aws_lightsail_database`
    * `aws_lightsail_disk`
    * `aws_lightsail_disk_attachment`
    * `aws_lightsail_distribution`
    * `aws_lightsail_domain`
    * `aws_lightsail_domain_entry`
    * `aws_lightsail_instance`
    * `aws_lightsail_instance_public_ports`
    * `aws_lightsail_key_pair`
    * `aws_lightsail_lb`
    * `aws_lightsail_lb_attachment`
    * `aws_lightsail_lb_certificate`
    * `aws_lightsail_lb_certificate_attachment`
    * `aws_lightsail_lb_https_redirection_policy`
    * `aws_lightsail_lb_stickiness_policy`
    * `aws_lightsail_static_ip`
    * `aws_lightsail_static_ip_attachment`
*   `local_gateway`
    * `aws_ec2_local_gateway_route`
    * `aws_ec2_local_gateway_route_table`
    * `aws_ec2_local_gateway_route_table_virtual_interface_group_association`
    * `aws_ec2_local_gateway_route_table_vpc_association`
*   `location`
    * `aws_location_geofence_collection`
    * `aws_location_map`
    * `aws_location_place_index`
    * `aws_location_route_calculator`
    * `aws_location_tracker`
*   `logs`
    * `aws_cloudwatch_log_account_policy`
    * `aws_cloudwatch_log_anomaly_detector`
    * `aws_cloudwatch_log_data_protection_policy`
    * `aws_cloudwatch_log_delivery`
    * `aws_cloudwatch_log_delivery_destination`
    * `aws_cloudwatch_log_delivery_destination_policy`
    * `aws_cloudwatch_log_delivery_source`
    * `aws_cloudwatch_log_destination`
    * `aws_cloudwatch_log_destination_policy`
    * `aws_cloudwatch_log_group`
    * `aws_cloudwatch_log_index_policy`
    * `aws_cloudwatch_log_metric_filter`
    * `aws_cloudwatch_log_resource_policy`
    * `aws_cloudwatch_log_stream`
    * `aws_cloudwatch_log_subscription_filter`
    * `aws_cloudwatch_log_transformer`
    * `aws_cloudwatch_query_definition`
*   `m2`
    * `aws_m2_application`
    * `aws_m2_environment`
*   `macie2`
    * `aws_macie2_classification_job`
    * `aws_macie2_custom_data_identifier`
    * `aws_macie2_findings_filter`
    * `aws_macie2_member`
*   `mailmanager`
    * `aws_mailmanager_archive`
    * `aws_mailmanager_ingress_point`
    * `aws_mailmanager_relay`
    * `aws_mailmanager_rule_set`
    * `aws_mailmanager_traffic_policy`
*   `media_package`
    * `aws_media_package_channel`
    * `aws_media_packagev2_channel_group`
*   `media_store`
    * `aws_media_store_container`
    * `aws_media_store_container_policy`
*   `mediaconvert`
    * `aws_media_convert_queue`
*   `medialive`
    * `aws_medialive_channel`
    * `aws_medialive_input`
    * `aws_medialive_input_security_group`
    * `aws_medialive_multiplex`
    * `aws_medialive_multiplex_program`
*   `memorydb`
    * `aws_memorydb_acl`
    * `aws_memorydb_cluster`
    * `aws_memorydb_multi_region_cluster`
    * `aws_memorydb_parameter_group`
    * `aws_memorydb_snapshot`
    * `aws_memorydb_subnet_group`
    * `aws_memorydb_user`
*   `mq`
    * `aws_mq_broker`
    * `aws_mq_configuration`
*   `msk`
    * `aws_msk_channel`
    * `aws_msk_cluster`
    * `aws_msk_cluster_policy`
    * `aws_msk_configuration`
    * `aws_msk_replicator`
    * `aws_msk_scram_secret_association`
    * `aws_msk_serverless_cluster`
    * `aws_msk_vpc_connection`
    * `aws_mskconnect_connector`
    * `aws_mskconnect_custom_plugin`
    * `aws_mskconnect_worker_configuration`
*   `mwaa`
    * `aws_mwaa_environment`
*   `nacl`
    * `aws_default_network_acl`
    * `aws_network_acl`
    * `aws_network_acl_association`
    * `aws_network_acl_rule`
*   `nat`
    * `aws_nat_gateway`
    * `aws_nat_gateway_eip_association`
*   `neptune`
    * `aws_neptune_cluster`
    * `aws_neptune_cluster_endpoint`
    * `aws_neptune_cluster_instance`
    * `aws_neptune_cluster_parameter_group`
    * `aws_neptune_cluster_snapshot`
    * `aws_neptune_event_subscription`
    * `aws_neptune_global_cluster`
    * `aws_neptune_parameter_group`
    * `aws_neptune_subnet_group`
*   `neptunegraph`
    * `aws_neptunegraph_graph`
    * `aws_neptunegraph_private_graph_endpoint`
*   `networkfirewall`
    * `aws_networkfirewall_container_association`
    * `aws_networkfirewall_firewall`
    * `aws_networkfirewall_firewall_policy`
    * `aws_networkfirewall_logging_configuration`
    * `aws_networkfirewall_resource_policy`
    * `aws_networkfirewall_rule_group`
    * `aws_networkfirewall_tls_inspection_configuration`
    * `aws_networkfirewall_vpc_endpoint_association`
*   `networkflowmonitor`
    * `aws_networkflowmonitor_monitor`
    * `aws_networkflowmonitor_scope`
*   `networkmanager`
    * `aws_networkmanager_attachment_routing_policy_label`
    * `aws_networkmanager_connect_attachment`
    * `aws_networkmanager_connect_peer`
    * `aws_networkmanager_connection`
    * `aws_networkmanager_core_network`
    * `aws_networkmanager_core_network_policy_attachment`
    * `aws_networkmanager_customer_gateway_association`
    * `aws_networkmanager_device`
    * `aws_networkmanager_dx_gateway_attachment`
    * `aws_networkmanager_global_network`
    * `aws_networkmanager_link`
    * `aws_networkmanager_link_association`
    * `aws_networkmanager_prefix_list_association`
    * `aws_networkmanager_site`
    * `aws_networkmanager_site_to_site_vpn_attachment`
    * `aws_networkmanager_transit_gateway_connect_peer_association`
    * `aws_networkmanager_transit_gateway_peering`
    * `aws_networkmanager_transit_gateway_registration`
    * `aws_networkmanager_transit_gateway_route_table_attachment`
    * `aws_networkmanager_vpc_attachment`
*   `networkmonitor`
    * `aws_networkmonitor_monitor`
    * `aws_networkmonitor_probe`
*   `notifications`
    * `aws_notifications_channel_association`
    * `aws_notifications_event_rule`
    * `aws_notifications_managed_notification_additional_channel_association`
    * `aws_notifications_notification_configuration`
    * `aws_notifications_notification_hub`
    * `aws_notifications_organizational_unit_association`
*   `notificationscontacts`
    * `aws_notificationscontacts_email_contact`
*   `oam`
    * `aws_oam_link`
    * `aws_oam_sink`
    * `aws_oam_sink_policy`
*   `observabilityadmin`
    * `aws_observabilityadmin_centralization_rule_for_organization`
    * `aws_observabilityadmin_dataset_integration`
    * `aws_observabilityadmin_s3_table_integration`
    * `aws_observabilityadmin_telemetry_pipeline`
    * `aws_observabilityadmin_telemetry_rule`
    * `aws_observabilityadmin_telemetry_rule_for_organization`
*   `odb`
    * `aws_odb_autonomous_database`
    * `aws_odb_cloud_autonomous_vm_cluster`
    * `aws_odb_cloud_exadata_infrastructure`
    * `aws_odb_cloud_vm_cluster`
    * `aws_odb_network`
    * `aws_odb_network_peering_connection`
*   `opensearch`
    * `aws_opensearch_application`
    * `aws_opensearch_domain`
    * `aws_opensearch_domain_policy`
    * `aws_opensearch_inbound_connection_accepter`
    * `aws_opensearch_outbound_connection`
    * `aws_opensearch_package`
    * `aws_opensearch_vpc_endpoint`
    * `aws_opensearchserverless_access_policy`
    * `aws_opensearchserverless_collection`
    * `aws_opensearchserverless_collection_group`
    * `aws_opensearchserverless_lifecycle_policy`
    * `aws_opensearchserverless_security_config`
    * `aws_opensearchserverless_security_policy`
    * `aws_opensearchserverless_vpc_endpoint`
*   `opsworks`
    * `aws_opsworks_application`
    * `aws_opsworks_custom_layer`
    * `aws_opsworks_instance`
    * `aws_opsworks_java_app_layer`
    * `aws_opsworks_php_app_layer`
    * `aws_opsworks_rds_db_instance`
    * `aws_opsworks_stack`
    * `aws_opsworks_static_web_layer`
    * `aws_opsworks_user_profile`
*   `organization`
    * `aws_organizations_account`
    * `aws_organizations_aws_service_access`
    * `aws_organizations_delegated_administrator`
    * `aws_organizations_organization`
    * `aws_organizations_organizational_unit`
    * `aws_organizations_policy`
    * `aws_organizations_policy_attachment`
    * `aws_organizations_resource_policy`
*   `osis`
    * `aws_osis_pipeline`
    * `aws_osis_pipeline_endpoint`
*   `outposts`
    * `aws_outposts_capacity_task`
*   `paymentcryptography`
    * `aws_paymentcryptography_key`
    * `aws_paymentcryptography_key_alias`
*   `pinpoint`
    * `aws_pinpoint_app`
    * `aws_pinpoint_email_template`
    * `aws_pinpoint_event_stream`
*   `pinpointsmsvoicev2`
    * `aws_pinpointsmsvoicev2_configuration_set`
    * `aws_pinpointsmsvoicev2_event_destination`
    * `aws_pinpointsmsvoicev2_keyword`
    * `aws_pinpointsmsvoicev2_opt_out_list`
    * `aws_pinpointsmsvoicev2_phone_number`
    * `aws_pinpointsmsvoicev2_pool`
    * `aws_pinpointsmsvoicev2_resource_policy`
    * `aws_pinpointsmsvoicev2_sender_id`
*   `pipes`
    * `aws_pipes_pipe`
*   `pricingplanmanager`
    * `aws_pricingplanmanager_subscription`
*   `prometheus`
    * `aws_prometheus_alert_manager_definition`
    * `aws_prometheus_anomaly_detector`
    * `aws_prometheus_query_logging_configuration`
    * `aws_prometheus_resource_policy`
    * `aws_prometheus_rule_group_namespace`
    * `aws_prometheus_scraper`
    * `aws_prometheus_scraper_logging_configuration`
    * `aws_prometheus_workspace`
    * `aws_prometheus_workspace_configuration`
*   `qbusiness`
    * `aws_qbusiness_application`
*   `qldb`
    * `aws_qldb_ledger`
    * `aws_qldb_stream`
*   `quicksight`
    * `aws_quicksight_account_settings`
    * `aws_quicksight_account_subscription`
    * `aws_quicksight_analysis`
    * `aws_quicksight_custom_permissions`
    * `aws_quicksight_dashboard`
    * `aws_quicksight_data_set`
    * `aws_quicksight_data_source`
    * `aws_quicksight_folder`
    * `aws_quicksight_folder_membership`
    * `aws_quicksight_group`
    * `aws_quicksight_group_membership`
    * `aws_quicksight_iam_policy_assignment`
    * `aws_quicksight_ingestion`
    * `aws_quicksight_ip_restriction`
    * `aws_quicksight_key_registration`
    * `aws_quicksight_namespace`
    * `aws_quicksight_refresh_schedule`
    * `aws_quicksight_role_custom_permission`
    * `aws_quicksight_role_membership`
    * `aws_quicksight_template`
    * `aws_quicksight_template_alias`
    * `aws_quicksight_theme`
    * `aws_quicksight_user`
    * `aws_quicksight_user_custom_permission`
    * `aws_quicksight_vpc_connection`
*   `ram`
    * `aws_ram_permission`
    * `aws_ram_principal_association`
    * `aws_ram_resource_association`
    * `aws_ram_resource_share`
*   `rbin`
    * `aws_rbin_rule`
*   `rds`
    * `aws_db_cluster`
    * `aws_db_cluster_snapshot`
    * `aws_db_event_subscription`
    * `aws_db_instance`
    * `aws_db_instance_automated_backups_replication`
    * `aws_db_instance_role_association`
    * `aws_db_option_group`
    * `aws_db_parameter_group`
    * `aws_db_proxy`
    * `aws_db_proxy_default_target_group`
    * `aws_db_proxy_endpoint`
    * `aws_db_proxy_target`
    * `aws_db_snapshot`
    * `aws_db_snapshot_copy`
    * `aws_db_subnet_group`
    * `aws_rds_certificate`
    * `aws_rds_cluster`
    * `aws_rds_cluster_activity_stream`
    * `aws_rds_cluster_endpoint`
    * `aws_rds_cluster_instance`
    * `aws_rds_cluster_parameter_group`
    * `aws_rds_cluster_role_association`
    * `aws_rds_cluster_snapshot_copy`
    * `aws_rds_custom_db_engine_version`
    * `aws_rds_export_task`
    * `aws_rds_global_cluster`
    * `aws_rds_integration`
    * `aws_rds_reserved_instance`
    * `aws_rds_shard_group`
*   `redshift`
    * `aws_redshift_authentication_profile`
    * `aws_redshift_cluster`
    * `aws_redshift_cluster_iam_roles`
    * `aws_redshift_cluster_snapshot`
    * `aws_redshift_data_share_authorization`
    * `aws_redshift_data_share_consumer_association`
    * `aws_redshift_endpoint_access`
    * `aws_redshift_endpoint_authorization`
    * `aws_redshift_event_subscription`
    * `aws_redshift_hsm_client_certificate`
    * `aws_redshift_hsm_configuration`
    * `aws_redshift_idc_application`
    * `aws_redshift_integration`
    * `aws_redshift_parameter_group`
    * `aws_redshift_partner`
    * `aws_redshift_scheduled_action`
    * `aws_redshift_snapshot_copy_grant`
    * `aws_redshift_snapshot_schedule`
    * `aws_redshift_snapshot_schedule_association`
    * `aws_redshift_subnet_group`
    * `aws_redshift_usage_limit`
*   `redshiftserverless`
    * `aws_redshiftserverless_custom_domain_association`
    * `aws_redshiftserverless_endpoint_access`
    * `aws_redshiftserverless_namespace`
    * `aws_redshiftserverless_snapshot`
    * `aws_redshiftserverless_usage_limit`
    * `aws_redshiftserverless_workgroup`
*   `rekognition`
    * `aws_rekognition_collection`
    * `aws_rekognition_project`
    * `aws_rekognition_stream_processor`
*   `resiliencehub`
    * `aws_resiliencehub_resiliency_policy`
*   `resiliencehubv2`
    * `aws_resiliencehubv2_assertion`
    * `aws_resiliencehubv2_input_source`
    * `aws_resiliencehubv2_policy`
    * `aws_resiliencehubv2_service`
    * `aws_resiliencehubv2_service_function`
    * `aws_resiliencehubv2_system`
    * `aws_resiliencehubv2_user_journey`
*   `resourceexplorer2`
    * `aws_resourceexplorer2_index`
    * `aws_resourceexplorer2_view`
*   `resourcegroups`
    * `aws_resourcegroups_group`
    * `aws_resourcegroups_resource`
*   `rolesanywhere`
    * `aws_rolesanywhere_profile`
    * `aws_rolesanywhere_trust_anchor`
*   `route53`
    * `aws_route53_cidr_collection`
    * `aws_route53_cidr_location`
    * `aws_route53_delegation_set`
    * `aws_route53_health_check`
    * `aws_route53_hosted_zone_dnssec`
    * `aws_route53_key_signing_key`
    * `aws_route53_query_log`
    * `aws_route53_record`
    * `aws_route53_traffic_policy`
    * `aws_route53_traffic_policy_instance`
    * `aws_route53_vpc_association_authorization`
    * `aws_route53_zone`
    * `aws_route53_zone_association`
*   `route53domains`
    * `aws_route53domains_delegation_signer_record`
    * `aws_route53domains_domain`
    * `aws_route53domains_registered_domain`
*   `route53profiles`
    * `aws_route53profiles_association`
    * `aws_route53profiles_profile`
    * `aws_route53profiles_resource_association`
*   `route53recoverycontrolconfig`
    * `aws_route53recoverycontrolconfig_cluster`
    * `aws_route53recoverycontrolconfig_control_panel`
    * `aws_route53recoverycontrolconfig_routing_control`
    * `aws_route53recoverycontrolconfig_safety_rule`
*   `route53recoveryreadiness`
    * `aws_route53recoveryreadiness_cell`
    * `aws_route53recoveryreadiness_readiness_check`
    * `aws_route53recoveryreadiness_recovery_group`
    * `aws_route53recoveryreadiness_resource_set`
*   `route53resolver`
    * `aws_route53_resolver_dnssec_config`
    * `aws_route53_resolver_endpoint`
    * `aws_route53_resolver_firewall_config`
    * `aws_route53_resolver_firewall_domain_list`
    * `aws_route53_resolver_firewall_rule`
    * `aws_route53_resolver_firewall_rule_group`
    * `aws_route53_resolver_firewall_rule_group_association`
    * `aws_route53_resolver_query_log_config`
    * `aws_route53_resolver_query_log_config_association`
    * `aws_route53_resolver_rule`
    * `aws_route53_resolver_rule_association`
*   `route_table`
    * `aws_default_route_table`
    * `aws_main_route_table_association`
    * `aws_route`
    * `aws_route_table`
    * `aws_route_table_association`
*   `rum`
    * `aws_rum_app_monitor`
    * `aws_rum_metrics_destination`
*   `s3`
    * `aws_s3_bucket`
    * `aws_s3_bucket_abac`
    * `aws_s3_bucket_accelerate_configuration`
    * `aws_s3_bucket_acl`
    * `aws_s3_bucket_analytics_configuration`
    * `aws_s3_bucket_cors_configuration`
    * `aws_s3_bucket_intelligent_tiering_configuration`
    * `aws_s3_bucket_inventory`
    * `aws_s3_bucket_lifecycle_configuration`
    * `aws_s3_bucket_logging`
    * `aws_s3_bucket_metadata_configuration`
    * `aws_s3_bucket_metric`
    * `aws_s3_bucket_notification`
    * `aws_s3_bucket_object_lock_configuration`
    * `aws_s3_bucket_ownership_controls`
    * `aws_s3_bucket_policy`
    * `aws_s3_bucket_public_access_block`
    * `aws_s3_bucket_replication_configuration`
    * `aws_s3_bucket_request_payment_configuration`
    * `aws_s3_bucket_server_side_encryption_configuration`
    * `aws_s3_bucket_versioning`
    * `aws_s3_bucket_website_configuration`
    * `aws_s3_directory_bucket`
*   `s3control`
    * `aws_s3_access_point`
    * `aws_s3control_access_grant`
    * `aws_s3control_access_grants_instance`
    * `aws_s3control_access_grants_instance_resource_policy`
    * `aws_s3control_access_grants_location`
    * `aws_s3control_access_point_policy`
    * `aws_s3control_bucket`
    * `aws_s3control_bucket_lifecycle_configuration`
    * `aws_s3control_bucket_policy`
    * `aws_s3control_directory_bucket_access_point_scope`
    * `aws_s3control_multi_region_access_point`
    * `aws_s3control_multi_region_access_point_policy`
    * `aws_s3control_multi_region_access_point_routes`
    * `aws_s3control_object_lambda_access_point`
    * `aws_s3control_object_lambda_access_point_policy`
    * `aws_s3control_storage_lens_configuration`
*   `s3files`
    * `aws_s3files_access_point`
    * `aws_s3files_file_system`
    * `aws_s3files_file_system_policy`
    * `aws_s3files_mount_target`
    * `aws_s3files_synchronization_configuration`
*   `s3outposts`
    * `aws_s3outposts_endpoint`
*   `s3tables`
    * `aws_s3tables_namespace`
    * `aws_s3tables_table`
    * `aws_s3tables_table_bucket`
    * `aws_s3tables_table_bucket_policy`
    * `aws_s3tables_table_bucket_replication`
    * `aws_s3tables_table_policy`
    * `aws_s3tables_table_replication`
*   `s3vectors`
    * `aws_s3vectors_index`
    * `aws_s3vectors_vector_bucket`
    * `aws_s3vectors_vector_bucket_policy`
*   `sagemaker`
    * `aws_sagemaker_algorithm`
    * `aws_sagemaker_app`
    * `aws_sagemaker_app_image_config`
    * `aws_sagemaker_code_repository`
    * `aws_sagemaker_data_quality_job_definition`
    * `aws_sagemaker_device`
    * `aws_sagemaker_device_fleet`
    * `aws_sagemaker_domain`
    * `aws_sagemaker_endpoint`
    * `aws_sagemaker_endpoint_configuration`
    * `aws_sagemaker_feature_group`
    * `aws_sagemaker_flow_definition`
    * `aws_sagemaker_hub`
    * `aws_sagemaker_hub_content_reference`
    * `aws_sagemaker_human_task_ui`
    * `aws_sagemaker_hyper_parameter_tuning_job`
    * `aws_sagemaker_image`
    * `aws_sagemaker_image_version`
    * `aws_sagemaker_inference_component`
    * `aws_sagemaker_labeling_job`
    * `aws_sagemaker_mlflow_app`
    * `aws_sagemaker_mlflow_tracking_server`
    * `aws_sagemaker_model`
    * `aws_sagemaker_model_card`
    * `aws_sagemaker_model_card_export_job`
    * `aws_sagemaker_model_package_group`
    * `aws_sagemaker_model_package_group_policy`
    * `aws_sagemaker_monitoring_schedule`
    * `aws_sagemaker_notebook_instance`
    * `aws_sagemaker_notebook_instance_lifecycle_configuration`
    * `aws_sagemaker_pipeline`
    * `aws_sagemaker_project`
    * `aws_sagemaker_servicecatalog_portfolio_status`
    * `aws_sagemaker_space`
    * `aws_sagemaker_studio_lifecycle_config`
    * `aws_sagemaker_training_job`
    * `aws_sagemaker_user_profile`
    * `aws_sagemaker_workforce`
    * `aws_sagemaker_workteam`
*   `savingsplans`
    * `aws_savingsplans_savings_plan`
*   `scheduler`
    * `aws_scheduler_schedule`
    * `aws_scheduler_schedule_group`
*   `schemas`
    * `aws_schemas_discoverer`
    * `aws_schemas_registry`
    * `aws_schemas_registry_policy`
    * `aws_schemas_schema`
*   `secretsmanager`
    * `aws_secretsmanager_secret`
    * `aws_secretsmanager_secret_policy`
    * `aws_secretsmanager_secret_rotation`
    * `aws_secretsmanager_secret_version`
*   `securityhub`
    * `aws_securityhub_account`
    * `aws_securityhub_action_target`
    * `aws_securityhub_aggregator_v2`
    * `aws_securityhub_automation_rule`
    * `aws_securityhub_automation_rule_v2`
    * `aws_securityhub_configuration_policy`
    * `aws_securityhub_connector_v2`
    * `aws_securityhub_finding_aggregator`
    * `aws_securityhub_insight`
    * `aws_securityhub_member`
    * `aws_securityhub_product_subscription`
    * `aws_securityhub_standards_control`
    * `aws_securityhub_standards_subscription`
*   `securitylake`
    * `aws_securitylake_custom_log_source`
    * `aws_securitylake_data_lake`
    * `aws_securitylake_subscriber`
*   `servicecatalog`
    * `aws_servicecatalog_constraint`
    * `aws_servicecatalog_portfolio`
    * `aws_servicecatalog_portfolio_share`
    * `aws_servicecatalog_principal_portfolio_association`
    * `aws_servicecatalog_product`
    * `aws_servicecatalog_product_portfolio_association`
    * `aws_servicecatalog_provisioned_product`
    * `aws_servicecatalog_provisioning_artifact`
    * `aws_servicecatalog_service_action`
    * `aws_servicecatalog_tag_option`
    * `aws_servicecatalog_tag_option_resource_association`
*   `servicecatalogappregistry`
    * `aws_servicecatalogappregistry_application`
    * `aws_servicecatalogappregistry_attribute_group`
    * `aws_servicecatalogappregistry_attribute_group_association`
*   `servicediscovery`
    * `aws_service_discovery_http_namespace`
    * `aws_service_discovery_instance`
    * `aws_service_discovery_private_dns_namespace`
    * `aws_service_discovery_public_dns_namespace`
    * `aws_service_discovery_service`
*   `servicequotas`
    * `aws_servicequotas_service_quota`
*   `ses`
    * `aws_ses_configuration_set`
    * `aws_ses_domain_dkim`
    * `aws_ses_domain_identity`
    * `aws_ses_domain_mail_from`
    * `aws_ses_email_identity`
    * `aws_ses_event_destination`
    * `aws_ses_identity_notification_topic`
    * `aws_ses_identity_policy`
    * `aws_ses_receipt_filter`
    * `aws_ses_receipt_rule`
    * `aws_ses_receipt_rule_set`
    * `aws_ses_template`
*   `sesv2`
    * `aws_sesv2_configuration_set`
    * `aws_sesv2_configuration_set_event_destination`
    * `aws_sesv2_contact_list`
    * `aws_sesv2_dedicated_ip_assignment`
    * `aws_sesv2_dedicated_ip_pool`
    * `aws_sesv2_email_identity`
    * `aws_sesv2_email_identity_feedback_attributes`
    * `aws_sesv2_email_identity_mail_from_attributes`
    * `aws_sesv2_email_identity_policy`
    * `aws_sesv2_tenant`
    * `aws_sesv2_tenant_resource_association`
*   `sfn`
    * `aws_sfn_activity`
    * `aws_sfn_alias`
    * `aws_sfn_state_machine`
*   `sg`
    * `aws_default_security_group`
    * `aws_security_group`
    * `aws_security_group_rule`
    * `aws_vpc_security_group_egress_rule`
    * `aws_vpc_security_group_ingress_rule`
    * `aws_vpc_security_group_rules_exclusive`
    * `aws_vpc_security_group_vpc_association`
*   `shield`
    * `aws_shield_drt_access_log_bucket_association`
    * `aws_shield_protection`
    * `aws_shield_protection_group`
    * `aws_shield_protection_health_check_association`
*   `signer`
    * `aws_signer_signing_profile`
    * `aws_signer_signing_profile_permission`
*   `sns`
    * `aws_sns_platform_application`
    * `aws_sns_topic`
    * `aws_sns_topic_data_protection_policy`
    * `aws_sns_topic_policy`
    * `aws_sns_topic_subscription`
*   `sqs`
    * `aws_sqs_queue`
    * `aws_sqs_queue_policy`
    * `aws_sqs_queue_redrive_allow_policy`
    * `aws_sqs_queue_redrive_policy`
*   `ssm`
    * `aws_ssm_activation`
    * `aws_ssm_association`
    * `aws_ssm_document`
    * `aws_ssm_maintenance_window`
    * `aws_ssm_maintenance_window_target`
    * `aws_ssm_maintenance_window_task`
    * `aws_ssm_parameter`
    * `aws_ssm_patch_baseline`
    * `aws_ssm_patch_group`
    * `aws_ssm_resource_data_sync`
*   `ssmcontacts`
    * `aws_ssmcontacts_contact`
    * `aws_ssmcontacts_contact_channel`
    * `aws_ssmcontacts_plan`
    * `aws_ssmcontacts_rotation`
*   `ssmincidents`
    * `aws_ssmincidents_response_plan`
*   `ssmquicksetup`
    * `aws_ssmquicksetup_configuration_manager`
*   `ssoadmin`
    * `aws_ssoadmin_account_assignment`
    * `aws_ssoadmin_application`
    * `aws_ssoadmin_application_access_scope`
    * `aws_ssoadmin_application_assignment`
    * `aws_ssoadmin_application_assignment_configuration`
    * `aws_ssoadmin_customer_managed_policy_attachment`
    * `aws_ssoadmin_instance_access_control_attributes`
    * `aws_ssoadmin_managed_policy_attachment`
    * `aws_ssoadmin_permission_set`
    * `aws_ssoadmin_permission_set_inline_policy`
    * `aws_ssoadmin_permissions_boundary_attachment`
    * `aws_ssoadmin_region`
    * `aws_ssoadmin_trusted_token_issuer`
*   `storagegateway`
    * `aws_storagegateway_cache`
    * `aws_storagegateway_cached_iscsi_volume`
    * `aws_storagegateway_file_system_association`
    * `aws_storagegateway_gateway`
    * `aws_storagegateway_nfs_file_share`
    * `aws_storagegateway_smb_file_share`
    * `aws_storagegateway_stored_iscsi_volume`
    * `aws_storagegateway_tape_pool`
    * `aws_storagegateway_upload_buffer`
    * `aws_storagegateway_working_storage`
*   `subnet`
    * `aws_default_subnet`
    * `aws_ec2_subnet_cidr_reservation`
    * `aws_subnet`
*   `swf`
    * `aws_swf_domain`
*   `synthetics`
    * `aws_synthetics_canary`
    * `aws_synthetics_group`
    * `aws_synthetics_group_association`
*   `timestreaminfluxdb`
    * `aws_timestreaminfluxdb_db_cluster`
    * `aws_timestreaminfluxdb_db_instance`
*   `timestreamquery`
    * `aws_timestreamquery_scheduled_query`
*   `timestreamwrite`
    * `aws_timestreamwrite_database`
    * `aws_timestreamwrite_table`
*   `transcribe`
    * `aws_transcribe_language_model`
    * `aws_transcribe_medical_vocabulary`
    * `aws_transcribe_vocabulary`
    * `aws_transcribe_vocabulary_filter`
*   `transfer`
    * `aws_transfer_access`
    * `aws_transfer_agreement`
    * `aws_transfer_certificate`
    * `aws_transfer_connector`
    * `aws_transfer_host_key`
    * `aws_transfer_profile`
    * `aws_transfer_server`
    * `aws_transfer_ssh_key`
    * `aws_transfer_user`
    * `aws_transfer_web_app`
    * `aws_transfer_web_app_customization`
    * `aws_transfer_workflow`
*   `transit_gateway`
    * `aws_ec2_transit_gateway`
    * `aws_ec2_transit_gateway_connect`
    * `aws_ec2_transit_gateway_connect_peer`
    * `aws_ec2_transit_gateway_metering_policy`
    * `aws_ec2_transit_gateway_metering_policy_entry`
    * `aws_ec2_transit_gateway_multicast_domain`
    * `aws_ec2_transit_gateway_peering_attachment`
    * `aws_ec2_transit_gateway_policy_table`
    * `aws_ec2_transit_gateway_policy_table_association`
    * `aws_ec2_transit_gateway_policy_table_entry`
    * `aws_ec2_transit_gateway_prefix_list_reference`
    * `aws_ec2_transit_gateway_route`
    * `aws_ec2_transit_gateway_route_table`
    * `aws_ec2_transit_gateway_route_table_association`
    * `aws_ec2_transit_gateway_route_table_propagation`
    * `aws_ec2_transit_gateway_vpc_attachment`
    * `aws_vpn_concentrator`
*   `verifiedaccess`
    * `aws_verifiedaccess_endpoint`
    * `aws_verifiedaccess_instance`
    * `aws_verifiedaccess_instance_logging_configuration`
    * `aws_verifiedaccess_instance_trust_provider_attachment`
    * `aws_verifiedaccess_trust_provider`
*   `verifiedpermissions`
    * `aws_verifiedpermissions_identity_source`
    * `aws_verifiedpermissions_policy`
    * `aws_verifiedpermissions_policy_store`
    * `aws_verifiedpermissions_policy_template`
    * `aws_verifiedpermissions_schema`
*   `vpc`
    * `aws_default_vpc`
    * `aws_default_vpc_dhcp_options`
    * `aws_flow_log`
    * `aws_vpc`
    * `aws_vpc_block_public_access_exclusion`
    * `aws_vpc_dhcp_options`
    * `aws_vpc_dhcp_options_association`
    * `aws_vpc_encryption_control`
    * `aws_vpc_ipam`
    * `aws_vpc_ipam_organization_admin_account`
    * `aws_vpc_ipam_pool`
    * `aws_vpc_ipam_pool_cidr`
    * `aws_vpc_ipam_pool_cidr_allocation`
    * `aws_vpc_ipam_resource_discovery`
    * `aws_vpc_ipam_resource_discovery_association`
    * `aws_vpc_ipam_scope`
    * `aws_vpc_ipv4_cidr_block_association`
    * `aws_vpc_ipv6_cidr_block_association`
    * `aws_vpc_route_server`
    * `aws_vpc_route_server_endpoint`
    * `aws_vpc_route_server_peer`
    * `aws_vpc_route_server_propagation`
    * `aws_vpc_route_server_vpc_association`
*   `vpc_endpoint`
    * `aws_vpc_endpoint`
    * `aws_vpc_endpoint_connection_accepter`
    * `aws_vpc_endpoint_connection_notification`
    * `aws_vpc_endpoint_policy`
    * `aws_vpc_endpoint_private_dns`
    * `aws_vpc_endpoint_route_table_association`
    * `aws_vpc_endpoint_security_group_association`
    * `aws_vpc_endpoint_service`
    * `aws_vpc_endpoint_subnet_association`
*   `vpc_peering`
    * `aws_vpc_peering_connection`
    * `aws_vpc_peering_connection_accepter`
    * `aws_vpc_peering_connection_options`
*   `vpclattice`
    * `aws_vpclattice_access_log_subscription`
    * `aws_vpclattice_auth_policy`
    * `aws_vpclattice_domain_verification`
    * `aws_vpclattice_listener`
    * `aws_vpclattice_listener_rule`
    * `aws_vpclattice_resource_configuration`
    * `aws_vpclattice_resource_gateway`
    * `aws_vpclattice_resource_policy`
    * `aws_vpclattice_service`
    * `aws_vpclattice_service_network`
    * `aws_vpclattice_service_network_resource_association`
    * `aws_vpclattice_service_network_service_association`
    * `aws_vpclattice_service_network_vpc_association`
    * `aws_vpclattice_target_group`
*   `vpn_connection`
    * `aws_vpn_connection`
*   `vpn_gateway`
    * `aws_vpn_gateway`
*   `waf`
    * `aws_waf_byte_match_set`
    * `aws_waf_geo_match_set`
    * `aws_waf_ipset`
    * `aws_waf_rate_based_rule`
    * `aws_waf_regex_match_set`
    * `aws_waf_regex_pattern_set`
    * `aws_waf_rule`
    * `aws_waf_rule_group`
    * `aws_waf_size_constraint_set`
    * `aws_waf_sql_injection_match_set`
    * `aws_waf_web_acl`
    * `aws_waf_xss_match_set`
*   `waf_regional`
    * `aws_wafregional_byte_match_set`
    * `aws_wafregional_geo_match_set`
    * `aws_wafregional_ipset`
    * `aws_wafregional_rate_based_rule`
    * `aws_wafregional_regex_match_set`
    * `aws_wafregional_regex_pattern_set`
    * `aws_wafregional_rule`
    * `aws_wafregional_rule_group`
    * `aws_wafregional_size_constraint_set`
    * `aws_wafregional_sql_injection_match_set`
    * `aws_wafregional_web_acl`
    * `aws_wafregional_web_acl_association`
    * `aws_wafregional_xss_match_set`
*   `wafv2_cloudfront`
    * `aws_wafv2_ip_set`
    * `aws_wafv2_regex_pattern_set`
    * `aws_wafv2_rule_group`
    * `aws_wafv2_web_acl`
    * `aws_wafv2_web_acl_logging_configuration`
*   `wafv2_regional`
    * `aws_wafv2_ip_set`
    * `aws_wafv2_regex_pattern_set`
    * `aws_wafv2_rule_group`
    * `aws_wafv2_web_acl`
    * `aws_wafv2_web_acl_association`
    * `aws_wafv2_web_acl_logging_configuration`
*   `workmail`
    * `aws_workmail_domain`
    * `aws_workmail_group`
    * `aws_workmail_organization`
    * `aws_workmail_user`
*   `workspaces`
    * `aws_workspaces_connection_alias`
    * `aws_workspaces_directory`
    * `aws_workspaces_ip_group`
    * `aws_workspaces_pool`
    * `aws_workspaces_workspace`
*   `workspacesweb`
    * `aws_workspacesweb_browser_settings`
    * `aws_workspacesweb_browser_settings_association`
    * `aws_workspacesweb_data_protection_settings`
    * `aws_workspacesweb_data_protection_settings_association`
    * `aws_workspacesweb_identity_provider`
    * `aws_workspacesweb_ip_access_settings`
    * `aws_workspacesweb_ip_access_settings_association`
    * `aws_workspacesweb_network_settings`
    * `aws_workspacesweb_network_settings_association`
    * `aws_workspacesweb_portal`
    * `aws_workspacesweb_session_logger`
    * `aws_workspacesweb_session_logger_association`
    * `aws_workspacesweb_trust_store`
    * `aws_workspacesweb_trust_store_association`
    * `aws_workspacesweb_user_access_logging_settings`
    * `aws_workspacesweb_user_access_logging_settings_association`
    * `aws_workspacesweb_user_settings`
    * `aws_workspacesweb_user_settings_association`
*   `xray`
    * `aws_xray_group`
    * `aws_xray_indexing_rule`
    * `aws_xray_resource_policy`
    * `aws_xray_sampling_rule`


#### Known limits — not importable

The following provider types are deliberately not generated. Each was verified during the
AWS coverage sweep: it has no list/describe API in aws-sdk-go-v2, is a singleton
account/Region-level settings object, is derived/version/association glue with no
independent enumeration, has no importer in the Terraform AWS provider, or can only be
enumerated through write-path APIs.

*   `aws_account_alternate_contact` — Fixed set of 3 contact types (BILLING/OPERATIONS/SECURITY) on the singleton account object; no list operation in the account SDK (only GetAlternateContact per type) — account-level settings object.
*   `aws_account_primary_contact` — Singleton account-level contact object; provider ID is the account ID itself; account SDK has no list API (only GetContactInformation for the caller's account) — existence not enumerable.
*   `aws_acm_certificate_validation` — derived validation-wait helper over aws_acm_certificate (already generated); no independent ACM API object.
*   `aws_acmpca_certificate` — _authority_certificate — singleton-per-CA installation of the CA's own certificate chain; an attribute of the CA itself (GetCertificateAuthorityCertificate), not an independent object.
*   `aws_acmpca_certificate_authority_certificate` — singleton-per-CA installation of the CA's own certificate chain; an attribute of the CA itself (GetCertificateAuthorityCertificate), not an independent object.
*   `aws_agentregistry_registry` — the standalone agentregistry Go SDK (v1.7.7, in go.mod) is a data-plane discovery client (List/Search/BatchGet of discoverable registry *records* only); it has no list/describe API for registries, and the provider's replacement resource for the deprecated bedrockagentcore registry is not yet established.
*   `aws_ami_copy` — no provider importer.
*   `aws_ami_from_instance` — no provider importer.
*   `aws_api_gateway_account` — singleton per-region account settings.
*   `aws_api_gateway_rest_api_put` — write-path-only action resource; its import ID is the rest_api_id itself and its body/triggers are not enumerable; would duplicate `aws_api_gateway_rest_api`.
*   `aws_appautoscaling_scheduled_action` — Provider docs have no Import section (no provider importer).
*   `aws_appfabric_app_authorization` — _connection` — provider docs have no Import section (no provider importer).
*   `aws_appfabric_app_authorization_connection` — provider docs have no Import section (no provider importer).
*   `aws_appfabric_ingestion_destination` — provider docs have no Import section (no provider importer).
*   `aws_apprunner_default_auto_scaling_configuration_version` — Singleton per-region default pointer.
*   `aws_apprunner_deployment` — Deployment operation record; provider docs have no Import section (no provider importer).
*   `aws_appstream_fleet_stack_association` — Association glue; the SDK exposes no list/describe operation enumerating fleet-stack pairs.
*   `aws_arczonalshift_autoshift_observer_notification_status` — Singleton account/region-level observer notification setting (GetAutoshiftObserverNotificationStatus only).
*   `aws_auditmanager_account_registration` — Account/Region singleton; provider ID is the Region name; only Get-status/settings APIs.
*   `aws_auditmanager_organization_admin_account_registration` — At most one delegated admin per organization; SDK exposes only `GetOrganizationAdminAccount` (no list API).
*   `aws_autoscaling_attachment` — Provider docs have no Import section (no provider importer).
*   `aws_autoscaling_group_tag` — Tag glue on the ASG.
*   `aws_autoscaling_notification` — Provider docs have no Import section (no provider importer).
*   `aws_autoscaling_traffic_source_attachment` — Association glue between an ASG and a traffic source; no independent enumeration.
*   `aws_backup_global_settings` — account-level singleton (a Describe exists but there is exactly one implicit object; not enumerable) — singleton precedent. (`backup.go`).
*   `aws_backup_region_settings` — region-level singleton, same reasoning. (`backup.go`).
*   `aws_bedrock_foundation_model_agreement` — non-enumerable: foundation model agreement offers are listed, accepted agreements are not (already documented in bedrock.go; re-verified disposition).
*   `aws_bedrock_guardrail_version` — no ListGuardrailVersions API; versions are created imperatively (already documented in bedrock.go).
*   `aws_bedrock_use_case_for_model_access` — account form submission, not a listable resource (already documented in bedrock.go).
*   `aws_bedrockagentcore_resource_policy` — only GetResourcePolicy for a known ARN; no list API (already documented in bedrock_agentcore.go).
*   `aws_chime_voice_connector_termination_credentials` — credentials are write-only; the API returns only usernames, so the resource cannot be faithfully imported.
*   `aws_chimesdkvoice_global_settings` — singleton account-level settings.
*   `aws_cloudcontrolapi_resource` — No provider importer (provider docs have no Import section); Cloud Control ListResources requires a caller-supplied resource type name, so there is no global inventory to enumerate.
*   `aws_cloudfrontkeyvaluestore_keys_exclusive` — Exclusive-management resource whose state is the store's full key set; no independent API object, would fight with the individual key resources.
*   `aws_cloudwatch_contributor_managed_insight_rule` — No enumeration API: `ListManagedInsightRules` requires a caller-supplied ResourceARN; nothing lists the resources to query.
*   `aws_cloudwatch_event_endpoint` — cloudwatchevents SDK (v1.41.7) exposes no endpoint operations at all (no Create/List/Describe endpoint APIs).
*   `aws_cloudwatch_event_permission` — No list API; permissions exist only inside the event bus policy JSON (covered by aws_cloudwatch_event_bus_policy).
*   `aws_cloudwatch_log_s3_table_integration_source` — Association glue under an S3 Tables integration ARN; no standalone list API.
*   `aws_cloudwatch_log_storage_tier_policy` — Account singleton; `GetStorageTierPolicy` takes no input and returns the single account policy.
*   `aws_cloudwatch_otel_enrichment` — Account singleton; Get-only API (`GetOtelEnrichmentStatus`-style), not enumerable.
*   `aws_cloudwatch_resource_metrics_configuration` — Get-only, keyed by a caller-supplied resource ARN; no enumeration API.
*   `aws_cloudwatchomni_organization_domain` — Organization singleton; only `GetDomainForOrganization` keyed by a caller-supplied organization ID; no list API.
*   `aws_codeartifact_domain_permissions_policy` — Policy glue; only a per-parent GetDomainPermissionsPolicy exists.
*   `aws_codeartifact_repository_permissions_policy` — Policy glue; only a per-parent GetRepositoryPermissionsPolicy exists.
*   `aws_codebuild_resource_policy` — Resource-policy glue; only a per-ARN GetResourcePolicy exists, no enumeration.
*   `aws_codecatalyst_dev_environment` — CodeCatalyst is closed to new customers and has no aws-sdk-go-v2 enumeration path used by this sweep; additionally no provider importer.
*   `aws_codecatalyst_project` — CodeCatalyst is closed to new customers; no aws-sdk-go-v2 enumeration path used by this sweep.
*   `aws_codecatalyst_source_repository` — CodeCatalyst is closed to new customers; no aws-sdk-go-v2 enumeration path used by this sweep.
*   `aws_codecommit_approval_rule_template_association` — Association glue between an approval rule template and a repository; no independent enumeration (per-repository association listing only).
*   `aws_codecommit_trigger` — Provider docs have no Import section (no provider importer).
*   `aws_codegurureviewer_repository_association` — Provider docs have no Import section (no provider importer).
*   `aws_codestarconnections_connection` — Deprecated legacy alias of aws_codeconnections_connection over the same underlying connections.
*   `aws_codestarconnections_host` — Deprecated legacy alias of aws_codeconnections_host over the same underlying hosts.
*   `aws_cognito_identity_pool_provider_principal_tag` — per-(pool, identity provider) principal-tag map; only GetPrincipalTagAttributeMap for a known provider, no list API.
*   `aws_cognito_managed_user_pool_client` — template-managed clients are indistinguishable from ordinary clients in ListUserPoolClients output, so these resources cannot be reconstructed from enumeration.
*   `aws_cognito_user_in_group` — user↔group membership association glue.
*   `aws_cognito_user_pool_ui_customization` — per-(pool, client) UI customization settings; only GetUICustomization for a known client, no list API.
*   `aws_computeoptimizer_enrollment_status` — Singleton account-level enrollment object; provider ID is the account ID; no list API.
*   `aws_computeoptimizer_recommendation_preferences` — Singleton account-level preferences object; provider ID is a scope tuple anchored on the account ID; no list API.
*   `aws_costoptimizationhub_enrollment_status` — Singleton account-level enrollment object; provider ID is the account ID; no list API.
*   `aws_costoptimizationhub_preferences` — Singleton account-level preferences object; provider ID is the account ID; no list API.
*   `aws_cur_report_definition` — No aws-sdk-go-v2 client exists for Cost and Usage Reports (DescribeReportDefinitions is SDK v1 / other SDKs only), so it cannot be enumerated with this codebase's SDK generation.
*   `aws_customerprofiles_profile` — Customer Profiles SDK has no ListProfiles operation; profiles are only reachable via SearchProfiles, which requires a caller-supplied key name + value — inventory not enumerable.
*   `aws_dataexchange_revision_assets` — No provider importer (provider docs have no Import section); write-path bundle that creates assets inside a revision, not an enumerable inventory object.
*   `aws_datazone_policy_grant` — grants are enumerable only per entity via.
*   `aws_detective_invitation_accepter` — accepter side of a member invitation.
*   `aws_detective_organization_admin_account` — singleton delegated-administrator designation.
*   `aws_detective_organization_configuration` — singleton per-graph organization configuration (DescribeOrganizationConfiguration).
*   `aws_devopsguru_event_sources_config` — Region singleton settings; provider ID is the Region name; Get-only API.
*   `aws_devopsguru_service_integration` — Region singleton settings; provider ID is the Region name; Get-only API.
*   `aws_directory_service_ip_routes_exclusive` — exclusive; conflicts with aws_directory_service_ip_route generated here.
*   `aws_directory_service_shared_directory_accepter` — accepter side of a directory share, in the receiving account.
*   `aws_dx_bgp_peer` — No Import section in provider docs — no provider importer.
*   `aws_dx_connection_association` — No Import section in provider docs — no provider importer.
*   `aws_dx_connection_confirmation` — No provider importer; action-style confirmation with no independent API object.
*   `aws_dx_hosted_connection` — No Import section in provider docs — no provider importer.
*   `aws_dx_macsec_key_association` — Provider docs give no Import section / documented ID format.
*   `aws_dynamodb_table_item` — data-plane item; items are table data, not enumerable.
*   `aws_dynamodb_table_replica` — importer ID is `"<table>:<main region>"`, but the.
*   `aws_dynamodb_tag` — single-tag association glue; tags are attributes of the table.
*   `aws_ebs_default_kms_key` — region-level singleton setting; importable by region/account in the provider but not enumerable (no list call; exactly one implicit object per region) — singleton precedent. (`ebs_sweep.go`).
*   `aws_ebs_encryption_by_default` — region-level singleton setting; same reasoning. (`ebs_sweep.go`).
*   `aws_ebs_snapshot_block_public_access` — region-level singleton setting; same reasoning. (`ebs_sweep.go`).
*   `aws_ebs_snapshot_copy` — one-off copy operation whose result is an ordinary snapshot (already covered by aws_ebs_snapshot); no API enumerates resources *as* copies, so it cannot be discovered distinctly. (`ebs_sweep.go`).
*   `aws_ebs_snapshot_import` — one-off import operation whose result is an ordinary snapshot (already covered by aws_ebs_snapshot); same reasoning. (`ebs_sweep.go`).
*   `aws_ebs_volume_copy` — one-off copy operation whose result is an ordinary volume (already covered by aws_ebs_volume); same reasoning. (`ebs_sweep.go`).
*   `aws_ec2_allowed_images_settings` — singleton account/region settings.
*   `aws_ec2_capacity_block_reservation` — no provider importer.
*   `aws_ec2_image_block_public_access` — singleton account/region settings.
*   `aws_ec2_instance_metadata_defaults` — provider does not support import.
*   `aws_ec2_instance_state` — overlay on `aws_instance` with no independent identity.
*   `aws_ec2_serial_console_access` — singleton account/region settings.
*   `aws_ec2_tag` — tags are attributes of their parent resource, already imported inline; standalone enumeration would duplicate every tag.
*   `aws_ec2_transit_gateway_default_route_table_association` — no provider importer.
*   `aws_ec2_transit_gateway_default_route_table_propagation` — no provider importer.
*   `aws_ec2_transit_gateway_multicast_domain_association` — no provider importer.
*   `aws_ec2_transit_gateway_multicast_group_member` — no provider importer.
*   `aws_ec2_transit_gateway_multicast_group_source` — no provider importer.
*   `aws_ec2_transit_gateway_peering_attachment_accepter` — accepter-side view of an attachment already enumerated by this generator; only meaningful under the accepter account's provider.
*   `aws_ec2_transit_gateway_vpc_attachment_accepter` — accepter-side view of an attachment already enumerated by this generator; only meaningful under the accepter account's provider.
*   `aws_ecr_account_setting` — Singleton registry setting; provider ID is a fixed setting name (e.g. BASIC_SCAN_TYPE_VERSION); no list API (single GetAccountSetting per name).
*   `aws_ecr_registry_policy` — Singleton registry object; provider ID is the registry ID (= account ID); no list API.
*   `aws_ecr_registry_scanning_configuration` — Singleton registry object; provider ID is the registry ID; no list API.
*   `aws_ecr_replication_configuration` — Singleton registry object; provider ID is the registry ID; no list API.
*   `aws_ecs_account_setting_default` — Singleton account settings; no independent resource identity.
*   `aws_ecs_tag` — Tag glue on other resources.
*   `aws_elastic_beanstalk_application_version` — Provider docs have no Import section (no provider importer).
*   `aws_elastic_beanstalk_configuration_template` — Provider docs have no Import section (no provider importer).
*   `aws_elasticsearch_domain_saml_options` — same as OpenSearch SAML options: no get/list API in the Elasticsearch SDK.
*   `aws_elb_attachment` — no provider importer — the provider docs for these Classic ELB policy/attachment types carry no Import section (verified in elb_depth.go).
*   `aws_emr_block_public_access_configuration` — Singleton account configuration; ID is the literal "current".
*   `aws_fis_safety_lever_state` — Exactly one safety lever per account/Region and it always exists; provider ID is the Region name; SDK has only `GetSafetyLever`.
*   `aws_fms_admin_account` — singleton administrator designation (GetAdminAccount).
*   `aws_gamelift_game_session_queue` — Finisher re-disposition.** The GameLift API has no list operation for game session queues — DescribeGameSessionQueues requires queue names/ARNs up front (verified against the pinned gamelift SDK v1.68.3 operation set), so there is no enumeration path.
*   `aws_glue_catalog` — the Glue Data Catalog itself is a singleton per.
*   `aws_glue_data_catalog_encryption_settings` — singleton settings object for the.
*   `aws_glue_partition` — partitions are data-plane slices of a table; a table can.
*   `aws_glue_resource_policy` — singleton policy object for the account's catalog.
*   `aws_grafana_role_association` — No provider importer (provider docs list no Import section).
*   `aws_grafana_workspace_api_key` — No provider importer; also write-only secret with no SDK list operation.
*   `aws_grafana_workspace_service_account_token` — No provider importer; write-only secret token.
*   `aws_guardduty_detector_feature` — no provider importer (provider docs page has no Import section; verified).
*   `aws_guardduty_invite_accepter` — accepter side of a member invitation, in the invited account.
*   `aws_guardduty_member_detector_feature` — per-member feature configuration glue under a member account; not an independent enumerable object.
*   `aws_guardduty_organization_admin_account` — singleton delegated-administrator designation.
*   `aws_guardduty_organization_configuration` — _feature — singleton organization feature configuration of the administrator account.
*   `aws_guardduty_organization_configuration_feature` — singleton organization feature configuration of the administrator account.
*   `aws_iam_account_password_policy` — singleton account-level password policy (GetAccountPasswordPolicy); no enumerable object.
*   `aws_iam_group_policies_exclusive` — exclusive resource; conflicts with aws_iam_group_policy already generated (would double-manage).
*   `aws_iam_group_policy_attachments_exclusive` — exclusive; conflicts with aws_iam_group_policy_attachment already generated.
*   `aws_iam_organizations_features` — singleton organization feature enablement; no enumerable object.
*   `aws_iam_outbound_web_identity_federation` — singleton account-level federation setting; no enumerable object.
*   `aws_iam_policy_attachment` — deprecated combined user/role/group attachment glue; the same attachments are already generated per principal, and a combined resource cannot be reconstructed unambiguously from per-principal attachments.
*   `aws_iam_role_policies_exclusive` — exclusive; conflicts with aws_iam_role_policy already generated.
*   `aws_iam_role_policy_attachments_exclusive` — exclusive; conflicts with aws_iam_role_policy_attachment already generated.
*   `aws_iam_security_token_service_preferences` — singleton account setting (global endpoint token version).
*   `aws_iam_user_policies_exclusive` — exclusive; conflicts with aws_iam_user_policy already generated.
*   `aws_iam_user_policy_attachments_exclusive` — exclusive; conflicts with aws_iam_user_policy_attachment already generated.
*   `aws_inspector2_delegated_admin_account` — singleton delegated-administrator designation.
*   `aws_inspector2_enabler` — singleton Inspector enablement for the calling account/region.
*   `aws_inspector2_organization_configuration` — singleton organization configuration (DescribeOrganizationConfiguration).
*   `aws_inspector_resource_group` — no list/describe enumeration entry point in the SDK (only DescribeResourceGroups by ARN, referenced from targets/templates).
*   `aws_iot_indexing_configuration` — provider docs carry **no Import section** for this singleton (verified: doc ends at Attribute Reference) → no provider importer.
*   `aws_iot_logging_options` — singleton account logging settings; no list API (only GetLoggingOptions for the account singleton).
*   `aws_iot_policy_attachment` — policy↔principal association glue.
*   `aws_iot_thing_group_membership` — thing↔thing-group association glue.
*   `aws_iot_thing_principal_attachment` — thing↔principal (certificate) association glue.
*   `aws_kinesis_account_settings` — account-level singleton settings; no list API, nothing to enumerate.
*   `aws_kinesisanalyticsv2_application_snapshot` — snapshot is a point-in-time version artifact of an application (derived/version glue).
*   `aws_kms_ciphertext` — result of a one-off Encrypt operation; not a stored object, no provider importer, nothing to enumerate.
*   `aws_lakeformation_data_lake_settings` — singleton settings object for the.
*   `aws_lakeformation_identity_center_configuration` — singleton per catalog.
*   `aws_lakeformation_permissions` — a permission grant; the provider ID is a hash of.
*   `aws_lakeformation_resource_lf_tag` — LF-tag-to-resource association glue; no.
*   `aws_lakeformation_resource_lf_tags` — same association-glue reason as.
*   `aws_lambda_invocation` — Action/invocation result, not a managed resource.
*   `aws_lambda_resource_policy` — Policy glue; already emitted as aws_lambda_permission from GetPolicy statements.
*   `aws_lambdacore_network_connector` — No aws-sdk-go-v2 service module exists for the lambdacore API.
*   `aws_lambdamicrovms_image` — No aws-sdk-go-v2 service module exists for the lambdamicrovms API.
*   `aws_lambdamicrovms_microvm` — No aws-sdk-go-v2 service module exists for the lambdamicrovms API.
*   `aws_lb_cookie_stickiness_policy` — no provider importer — the provider docs for these Classic ELB policy/attachment types carry no Import section (verified in elb_depth.go).
*   `aws_lb_ssl_negotiation_policy` — no provider importer — the provider docs for these Classic ELB policy/attachment types carry no Import section (verified in elb_depth.go).
*   `aws_licensemanager_association` — association glue between a license configuration and a resource; created implicitly on attach and the enumerated form carries no association ID of its own.
*   `aws_licensemanager_grant_accepter` — accepter side of a grant, in the receiving account.
*   `aws_load_balancer_backend_server_policy` — no provider importer — the provider docs for these Classic ELB policy/attachment types carry no Import section (verified in elb_depth.go).
*   `aws_load_balancer_listener_policy` — no provider importer — the provider docs for these Classic ELB policy/attachment types carry no Import section (verified in elb_depth.go).
*   `aws_load_balancer_policy` — no provider importer — the provider docs for these Classic ELB policy/attachment types carry no Import section (verified in elb_depth.go).
*   `aws_location_tracker_association` — tracker, consumer ARN: association glue, discoverable only per tracker via ListTrackerConsumers.
*   `aws_m2_deployment` — Deployment operation record per application, not a standing configuration resource.
*   `aws_macie2_account` — singleton Macie enablement for the account/region.
*   `aws_macie2_classification_export_configuration` — singleton account-level export configuration (GetClassificationExportConfiguration).
*   `aws_macie2_invitation_accepter` — accepter side of a member invitation.
*   `aws_macie2_organization_admin_account` — singleton delegated-administrator designation.
*   `aws_macie2_organization_configuration` — singleton organization configuration (DescribeOrganizationConfiguration).
*   `aws_msk_single_scram_secret_association` — single-secret association glue between one cluster and one secret; the whole-cluster association is implemented as aws_msk_scram_secret_association.
*   `aws_msk_topic` — topics live in the Kafka data plane; the MSK control-plane SDK has no ListTopics/DescribeTopic API (verified op inventory; provider format would be cluster_arn,name ✓ but there is no AWS API to enumerate topics with).
*   `aws_networkfirewall_firewall_transit_gateway_attachment_accepter` — Action-style accepter resource; no provider importer and no independent listable object (attachment is an attribute of the firewall).
*   `aws_networkmanager_attachment_accepter` — Action-style resource with no provider importer and no independent existence in the Network Manager API (already documented in `networkmanager.go`).
*   `aws_notifications_managed_notification_account_contact_association` — Association glue with no list API in the notifications SDK (only Associate/Disassociate operations exist) — existing associations are not enumerable.
*   `aws_notifications_organizations_access` — Singleton account-level Organizations-access toggle; provider ID is the account ID; no list API.
*   `aws_observabilityadmin_telemetry_enrichment` — Account/Region singleton; provider ID is the Region name; only a Get-status API exists.
*   `aws_observabilityadmin_telemetry_evaluation` — _for_organization | Organization singleton; provider ID is the Region name; only a Get-status API exists.
*   `aws_observabilityadmin_telemetry_evaluation_for_organization` — Organization singleton; provider ID is the Region name; only a Get-status API exists.
*   `aws_odb_autonomous_database_secrets_manager_integration` — no list API; the.
*   `aws_odb_iam_role_association` — IAM role associations are attributes of the.
*   `aws_opensearch_authorize_vpc_endpoint_access` — authorization grant (domain, principal) association glue; no standalone list API.
*   `aws_opensearch_domain_saml_options` — per-domain SAML settings subresource; the OpenSearch SDK exposes no get/list API for SAML options, so existence cannot be enumerated.
*   `aws_opensearch_package_association` — package↔domain association glue.
*   `aws_organizations_tag` — Single tag key/value pair on an arbitrary Organizations resource; no standalone tag inventory (tags are only readable per-resource via ListTagsForResource).
*   `aws_osis_resource_policy` — only GetResourcePolicy for a known ARN; no list/discovery API (same precedent as aws_bedrockagentcore_resource_policy).
*   `aws_proxy_protocol_policy` — no provider importer — the provider docs for these Classic ELB policy/attachment types carry no Import section (verified in elb_depth.go).
*   `aws_ram_resource_share_accepter` — accepter side of a share invitation, in the receiving account.
*   `aws_ram_resource_share_associations_exclusive` — exclusive; conflicts with the individual association resources for the same share.
*   `aws_ram_sharing_with_organization` — singleton organization sharing enablement; SDK exposes no get/list for its state.
*   `aws_rds_instance_state` — state-control action resource (start/stop) with no.
*   `aws_redshift_logging` — singleton per-cluster settings object (bucket / log.
*   `aws_redshift_namespace_registration` — no list/describe API in the Redshift SDK.
*   `aws_redshift_resource_policy` — policies attach to individual resources.
*   `aws_redshift_snapshot_copy` — singleton per-cluster settings object (cross-region.
*   `aws_redshiftdata_statement` — Transient Redshift Data API statement execution records; ListStatements returns short-lived query executions (ephemeral runtime state), not managed infrastructure.
*   `aws_redshiftserverless_resource_policy` — policies attach to individual resources.
*   `aws_route53_records_exclusive` — Exclusive-management resource whose state is exactly the zone record set already generated as `aws_route53_record`; no independent API object, and generating both would fight over the same records.
*   `aws_route53_resolver_config` — Singleton account/region-level Resolver settings object (one ListResolverConfigs result); not an enumerable object.
*   `aws_s3_account_public_access_block` — account-level singleton; the provider imports it by account ID, but it is not enumerable (exactly one per account, no list/discovery call) — singleton precedent. (`s3_subconfig.go`).
*   `aws_s3_bucket_object` — deprecated alias of aws_s3_object; same data-plane reasoning. (`s3_subconfig.go`).
*   `aws_s3_object` — data-plane user object, not bucket configuration; enumerating it would import application data, not infrastructure. (`s3_subconfig.go`).
*   `aws_s3_object_copy` — data-plane user object / copy operation, not bucket configuration; same reasoning. (`s3_subconfig.go`).
*   `aws_secretsmanager_tag` — per-tag glue; tags are attributes of aws_secretsmanager_secret already generated.
*   `aws_securityhub_account_v2` — singleton Security Hub V2 enablement for the account/region.
*   `aws_securityhub_configuration_policy_association` — association glue between a configuration policy and an account/OU/root target; the ListConfigurationPolicyAssociations summary carries no association ID of its own, and the association is fully described by the policy + target pair.
*   `aws_securityhub_feature_v2` — per-feature enablement toggle; SDK exposes only Enable/DisableSecurityHubFeatureV2, no list/get enumeration.
*   `aws_securityhub_invite_accepter` — accepter side of a member invitation, in the invited account.
*   `aws_securityhub_organization_admin_account` — singleton delegated-administrator designation.
*   `aws_securityhub_organization_configuration` — singleton organization configuration (DescribeOrganizationConfiguration).
*   `aws_securityhub_standards_control_association` — newer association form of the same standard/control enablement already covered by aws_securityhub_standards_control; emitting both would double-manage.
*   `aws_securitylake_aws_log_source` — per-account/region enablement configuration for a named AWS log source; ListLogSources entries are source configurations, not independent objects with their own IDs.
*   `aws_securitylake_subscriber_notification` — SDK exposes no separate get/list for a subscriber's notification configuration; it is part of the subscriber object.
*   `aws_serverlessapplicationrepository_cloudformation_stack` — A CloudFormation stack created by SAR; the SAR API exposes no list operation for deployed stacks, so no independent enumeration distinct from CloudFormation itself.
*   `aws_servicecatalog_budget_resource_association` — No global list; Service Catalog exposes only the reverse lookup ListBudgetsForResource for a caller-supplied resource ID.
*   `aws_servicecatalog_organizations_access` — No provider importer (provider docs have no Import section); singleton account-level Organizations-access toggle (GetAWSOrganizationsAccessStatus).
*   `aws_servicequotas_auto_management` — singleton account-level configuration (GetAutoManagementConfiguration).
*   `aws_servicequotas_template` — _association — singleton association state of that template (GetAssociationForServiceQuotaTemplate).
*   `aws_servicequotas_template_association` — singleton association state of that template (GetAssociationForServiceQuotaTemplate).
*   `aws_ses_active_receipt_rule_set` — singleton pointer naming the one active rule set; the rule set itself is already emitted by the existing generator.
*   `aws_ses_domain_identity_verification` — verification-waiter glue with no configuration of its own beyond the domain identity (already emitted).
*   `aws_sesv2_account_suppression_attributes` — singleton account-level settings.
*   `aws_sesv2_account_vdm_attributes` — singleton account-level settings.
*   `aws_sesv2_multi_region_endpoint` — provider docs have no Import section (no provider importer).
*   `aws_shield_application_layer_automatic_response` — No Import section in provider docs — no provider importer (configuration is visible per-protection in ListProtections, but cannot be imported).
*   `aws_shield_drt_access_role_arn_association` — Single account-level setting (the RoleArn field of DescribeDRTAccess); no list API.
*   `aws_shield_proactive_engagement` — Single account-level settings object; no list API.
*   `aws_shield_subscription` — Single account-level object (DescribeSubscription only); no list API.
*   `aws_signer_signing_job` — one-off signing operation; ListSigningJobs enumerates historical job executions, not durable manageable objects.
*   `aws_snapshot_create_volume_permission` — no provider importer.
*   `aws_sns_sms_preferences` — singleton account-level SMS settings.
*   `aws_spot_instance_request` — no provider importer.
*   `aws_ssm_default_patch_baseline` — Derived per-operating-system pointer to whichever baseline is registered default (RegisterDefaultPatchBaseline) over the fixed OS enum; not an independent object.
*   `aws_ssm_service_setting` — Singleton account/region service-settings object; provider ID is the setting ARN from a fixed, non-enumerable set of setting IDs; SSM exposes only GetServiceSetting per ID.
*   `aws_ssmincidents_replication_set` — Singleton per account/region; provider import ID is the literal string "import"; at most one replication set exists.
*   `aws_ssoadmin_customer_managed_policy_attachments_exclusive` — exclusive; conflicts with aws_ssoadmin_customer_managed_policy_attachment already generated in ssoadmin.go (also documented in that file).
*   `aws_ssoadmin_managed_policy_attachments_exclusive` — exclusive; conflicts with aws_ssoadmin_managed_policy_attachment already generated in ssoadmin.go (also documented in that file).
*   `aws_transfer_tag` — tag-assignment glue resource (resource ARN + key/value); tags are attributes of their parent resources, not an enumerable standalone object, and the resource has no list API of its own. (`transfer.go`).
*   `aws_uxc_account_customizations` — Singleton account-level customization object; provider ID is the account ID; uxc SDK exposes only GetAccountCustomizations for the caller's account.
*   `aws_verifiedaccess_group` — no provider importer.
*   `aws_vpc_block_public_access_options` — singleton account/region settings object.
*   `aws_vpc_endpoint_service_allowed_principal` — no provider importer.
*   `aws_vpc_endpoint_service_private_dns_verification` — provider explicitly does not support import.
*   `aws_vpc_ipam_preview_next_cidr` — computed preview, no provider importer.
*   `aws_vpc_network_performance_metric_subscription` — no provider importer.
*   `aws_vpclattice_target_group_attachment` — No provider importer (already documented in `vpclattice.go`); state ID is a create-time composite and Read derives the target from configuration, not from any listable object.
*   `aws_vpn_connection_route` — no provider importer: the provider docs page for `vpn_connection_route` has no Import section at all (verified by the finisher against the provider docs page). *(finisher).
*   `aws_vpn_gateway_attachment` — no provider importer (the VPN gateway itself is covered by VpnGatewayGenerator).
*   `aws_vpn_gateway_route_propagation` — no provider importer.
*   `aws_wafv2_api_key` — Import ID is the generated key value itself (`api_key,scope`); ListAPIKeys returns only summaries and the key value is never returned by any list/describe call (GetDecryptedAPIKey requires the encrypted key as input) — ID cannot be enumerated.
*   `aws_workmail_default_domain` — per-organization singleton pointer to the default domain; the domain itself is emitted.
*   `aws_xray_encryption_config` — Region singleton; provider ID is the Region name; Get-only API.
*   `aws_xray_trace_segment_destination` — Region singleton; provider ID is the Region name; Get-only API.

#### Global services

AWS services that are global will be imported without specified region even if several regions will be passed. It is to ensure only one representation of an AWS resource is imported.

List of global AWS services:
*   `budgets`
*   `cloudfront`
*   `ecrpublic`
*   `iam`
*   `organization`
*   `route53`
*   `waf`

#### Attribute filters

Attribute filters allow filtering across different resource types by its attributes.

```
terraformer import aws --resources=ec2_instance,ebs --filter="Name=tags.costCenter;Value=20000:'20001:1'" --regions=eu-west-1
```
Will only import AWS EC2 instances along with EBS volumes annotated with tag `costCenter` with values `20000` or `20001:1`. Attribute filters are by default applicable to all resource types although it's possible to specify to what resource type a given filter should be applicable to by providing `Type=<type>` parameter. For example:
```
terraformer import aws --resources=ec2_instance,ebs --filter=Type=ec2_instance;Name=tags.costCenter;Value=20000:'20001:1' --regions=eu-west-1
```
Will work as same as example above with a change the filter will be applicable only to `ec2_instance` resources.

Few more examples - How to import ec2 instance based on instance name and id
```
terraformer import aws --resources=ec2_instance --filter="Name=tags.Name;Value=Terraformer" --regions=us-east-1
```
This command imports ec2 instance having name as Terraformer.
```
terraformer import aws --resources=ec2_instance --filter="Name=id;Value=i-0xxxxxxxxx" --regions=us-east-1
```
This command imports ec2 instance having instance-id as i-0xxxxxxxxx.

Due to fact API Gateway generates a lot of resources, it's possible to issue a filtering query to retrieve resources related to a given REST API by tags. To fetch resources related to a REST API resource with a tag `STAGE` and value `dev`, add parameter `--filter="Type=api_gateway_rest_api;Name=tags.STAGE;Value=dev"`.

#### SQS queues retrieval

Terraformer uses AWS [ListQueues](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/API_ListQueues.html) API call to fetch available queues. The API is able to return only up to 1000 queues and an additional name prefix should be passed to filter the list results. It's possible to pass `QueueNamePrefix` parameter by environmental variable `SQS_PREFIX`.

#### Security groups and rules

Terraformer by default will try to keep rules in security groups as long as no circular dependencies are detected. This approach is implemented to keep the rules as tidy as possible but there can be cases when this behaviour is not desirable (see [GoogleCloudPlatform/terraformer#493](https://github.com/GoogleCloudPlatform/terraformer/issues/493)). To make Terraformer split rules from security groups, add `SPLIT_SG_RULES` environmental variable with any value.
