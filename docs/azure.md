# Use with Azure

## Authentication

### Supported Methods

- [Azure CLI](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/guides/azure_cli)
- [managed identities for Azure resources](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/guides/managed_service_identity)
- [Service Principal with Client Certificate](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/guides/service_principal_client_certificate)
- [Service Principal with Client Secret](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/guides/service_principal_client_secret)
- [Service Principal with Open ID Connect](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/guides/service_principal_oidc)

### Examples

``` sh
# Using Azure CLI (az login)
export ARM_SUBSCRIPTION_ID=[SUBSCRIPTION_ID]

# Using Managed identities for Azure resources
export ARM_SUBSCRIPTION_ID=[SUBSCRIPTION_ID]
export ARM_CLIENT_ID=[CLIENT_ID]  # only necessary for user assigned identity
export ARM_TENANT_ID=[TENANT_ID]
export ARM_USE_MSI=true
export ARM_MSI_ENDPOINT=[ARM_MSI_ENDPOINT] # only necessary when the msi endpoint is different than the well-known one

# Using Service Principal with Client Certificate
export ARM_SUBSCRIPTION_ID=[SUBSCRIPTION_ID]
export ARM_CLIENT_ID=[CLIENT_ID] # only necessary for user assigned identity
export ARM_TENANT_ID=[TENANT_ID]
export ARM_CLIENT_CERTIFICATE_PATH="/path/to/my/client/certificate.pfx"
export ARM_CLIENT_CERTIFICATE_PASSWORD=[CLIENT_CERTIFICATE_PASSWORD]

# Using Service Principal with Client Secret
export ARM_SUBSCRIPTION_ID=[SUBSCRIPTION_ID]
export ARM_CLIENT_ID=[CLIENT_ID]
export ARM_TENANT_ID=[TENANT_ID]
export ARM_CLIENT_SECRET=[CLIENT_SECRET]

# Using Service Principal with Open ID Connect
export ARM_SUBSCRIPTION_ID=[SUBSCRIPTION_ID]
export ARM_CLIENT_ID=[CLIENT_ID]
export ARM_TENANT_ID=[TENANT_ID]
export ARM_USE_OIDC=true

# Using deprecated ADAL authentication for throubleshooting
export ARM_USE_ADAL=true

./terraformer import azure -r resource_group
./terraformer import azure -R my_resource_group -r virtual_network,resource_group
./terraformer import azure -r resource_group --filter=resource_group=/subscriptions/<Subscription id>/resourceGroups/<RGNAME>
```

## List of supported Azure resources

*   `ai_foundry`
    * `azurerm_ai_foundry`
    * `azurerm_ai_foundry_project`
*   `analysis`
    * `azurerm_analysis_services_server`
*   `api_management`
    * `azurerm_api_management`
    * `azurerm_api_management_api`
    * `azurerm_api_management_api_diagnostic`
    * `azurerm_api_management_api_operation`
    * `azurerm_api_management_api_operation_tag`
    * `azurerm_api_management_api_release`
    * `azurerm_api_management_api_schema`
    * `azurerm_api_management_api_tag`
    * `azurerm_api_management_api_tag_description`
    * `azurerm_api_management_api_version_set`
    * `azurerm_api_management_authorization_server`
    * `azurerm_api_management_backend`
    * `azurerm_api_management_certificate`
    * `azurerm_api_management_diagnostic`
    * `azurerm_api_management_email_template`
    * `azurerm_api_management_gateway`
    * `azurerm_api_management_gateway_api`
    * `azurerm_api_management_gateway_certificate_authority`
    * `azurerm_api_management_gateway_host_name_configuration`
    * `azurerm_api_management_global_schema`
    * `azurerm_api_management_group`
    * `azurerm_api_management_group_user`
    * `azurerm_api_management_identity_provider_aad`
    * `azurerm_api_management_identity_provider_aadb2c`
    * `azurerm_api_management_identity_provider_facebook`
    * `azurerm_api_management_identity_provider_google`
    * `azurerm_api_management_identity_provider_microsoft`
    * `azurerm_api_management_identity_provider_twitter`
    * `azurerm_api_management_logger`
    * `azurerm_api_management_named_value`
    * `azurerm_api_management_notification_recipient_email`
    * `azurerm_api_management_notification_recipient_user`
    * `azurerm_api_management_openid_connect_provider`
    * `azurerm_api_management_product`
    * `azurerm_api_management_product_api`
    * `azurerm_api_management_product_group`
    * `azurerm_api_management_product_tag`
    * `azurerm_api_management_redis_cache`
    * `azurerm_api_management_subscription`
    * `azurerm_api_management_tag`
    * `azurerm_api_management_user`
*   `app_configuration`
    * `azurerm_app_configuration`
*   `app_service`
    * `azurerm_app_service`
*   `application_gateway`
    * `azurerm_application_gateway`
*   `automation`
    * `azurerm_automation_account`
    * `azurerm_automation_certificate`
    * `azurerm_automation_connection`
    * `azurerm_automation_connection_certificate`
    * `azurerm_automation_connection_classic_certificate`
    * `azurerm_automation_connection_service_principal`
    * `azurerm_automation_connection_type`
    * `azurerm_automation_credential`
    * `azurerm_automation_dsc_configuration`
    * `azurerm_automation_dsc_nodeconfiguration`
    * `azurerm_automation_hybrid_runbook_worker`
    * `azurerm_automation_hybrid_runbook_worker_group`
    * `azurerm_automation_job_schedule`
    * `azurerm_automation_module`
    * `azurerm_automation_python3_package`
    * `azurerm_automation_runbook`
    * `azurerm_automation_runtime_environment`
    * `azurerm_automation_runtime_environment_package`
    * `azurerm_automation_schedule`
    * `azurerm_automation_source_control`
    * `azurerm_automation_watcher`
    * `azurerm_automation_webhook`
*   `cognitive`
    * `azurerm_cognitive_account`
    * `azurerm_cognitive_account_connection_account_key`
    * `azurerm_cognitive_account_connection_account_managed_identity`
    * `azurerm_cognitive_account_connection_api_key`
    * `azurerm_cognitive_account_connection_custom_keys`
    * `azurerm_cognitive_account_connection_entra_id`
    * `azurerm_cognitive_account_project`
    * `azurerm_cognitive_account_rai_blocklist`
    * `azurerm_cognitive_account_rai_policy`
    * `azurerm_cognitive_deployment`
*   `communication`
    * `azurerm_communication_service`
*   `confidential_ledger`
    * `azurerm_confidential_ledger`
*   `arc`
    * `azurerm_arc_kubernetes_cluster`
    * `azurerm_arc_kubernetes_cluster_extension`
    * `azurerm_arc_kubernetes_flux_configuration`
    * `azurerm_arc_machine`
    * `azurerm_arc_machine_extension`
    * `azurerm_arc_private_link_scope`
    * `azurerm_arc_resource_bridge_appliance`
*   `automanage`
    * `azurerm_arc_machine_automanage_configuration_assignment`
    * `azurerm_automanage_configuration`
    * `azurerm_virtual_machine_automanage_configuration_assignment`
*   `batch`
    * `azurerm_batch_account`
    * `azurerm_batch_application`
    * `azurerm_batch_pool`
*   `compute`
    * `azurerm_availability_set`
    * `azurerm_capacity_reservation`
    * `azurerm_capacity_reservation_group`
    * `azurerm_dedicated_host`
    * `azurerm_dedicated_host_group`
    * `azurerm_disk_access`
    * `azurerm_disk_encryption_set`
    * `azurerm_gallery_application`
    * `azurerm_gallery_application_version`
    * `azurerm_image`
    * `azurerm_proximity_placement_group`
    * `azurerm_shared_image`
    * `azurerm_shared_image_gallery`
    * `azurerm_shared_image_version`
    * `azurerm_snapshot`
    * `azurerm_virtual_machine_restore_point_collection`
*   `cdn`
    * `azurerm_cdn_endpoint`
    * `azurerm_cdn_endpoint_custom_domain`
    * `azurerm_cdn_frontdoor_custom_domain`
    * `azurerm_cdn_frontdoor_endpoint`
    * `azurerm_cdn_frontdoor_origin`
    * `azurerm_cdn_frontdoor_origin_group`
    * `azurerm_cdn_frontdoor_profile`
    * `azurerm_cdn_frontdoor_route`
    * `azurerm_cdn_frontdoor_rule`
    * `azurerm_cdn_frontdoor_rule_set`
    * `azurerm_cdn_frontdoor_secret`
    * `azurerm_cdn_frontdoor_security_policy`
    * `azurerm_cdn_profile`
*   `application_insights`
    * `azurerm_application_insights`
    * `azurerm_application_insights_analytics_item`
    * `azurerm_application_insights_api_key`
    * `azurerm_application_insights_smart_detection_rule`
    * `azurerm_application_insights_web_test`
    * `azurerm_application_insights_workbook`
*   `application_load_balancer`
    * `azurerm_application_load_balancer`
    * `azurerm_application_load_balancer_frontend`
    * `azurerm_application_load_balancer_security_policy`
    * `azurerm_application_load_balancer_subnet_association`
*   `application_security_group`
    * `azurerm_application_security_group`
*   `attestation`
    * `azurerm_attestation_provider`
*   `bot`
    * `azurerm_bot_channel_alexa`
    * `azurerm_bot_channel_direct_line_speech`
    * `azurerm_bot_channel_directline`
    * `azurerm_bot_channel_email`
    * `azurerm_bot_channel_facebook`
    * `azurerm_bot_channel_line`
    * `azurerm_bot_channel_ms_teams`
    * `azurerm_bot_channel_slack`
    * `azurerm_bot_channel_sms`
    * `azurerm_bot_channel_web_chat`
    * `azurerm_bot_connection`
    * `azurerm_bot_service_azure_bot`
*   `container`
    * `azurerm_container_connected_registry`
    * `azurerm_container_group`
    * `azurerm_container_registry`
    * `azurerm_container_registry_agent_pool`
    * `azurerm_container_registry_cache_rule`
    * `azurerm_container_registry_credential_set`
    * `azurerm_container_registry_scope_map`
    * `azurerm_container_registry_task`
    * `azurerm_container_registry_token`
    * `azurerm_container_registry_webhook`
*   `container_app`
    * `azurerm_container_app`
    * `azurerm_container_app_custom_domain`
    * `azurerm_container_app_environment`
    * `azurerm_container_app_environment_certificate`
    * `azurerm_container_app_environment_custom_domain`
    * `azurerm_container_app_environment_dapr_component`
    * `azurerm_container_app_environment_managed_certificate`
    * `azurerm_container_app_environment_storage`
    * `azurerm_container_app_job`
*   `cost_management`
    * `azurerm_consumption_budget_management_group`
    * `azurerm_consumption_budget_resource_group`
    * `azurerm_consumption_budget_subscription`
    * `azurerm_resource_group_cost_management_export`
    * `azurerm_resource_group_cost_management_view`
    * `azurerm_subscription_cost_management_export`
    * `azurerm_subscription_cost_management_view`
*   `cosmosdb`
    * `azurerm_cosmosdb_account`
    * `azurerm_cosmosdb_cassandra_cluster`
    * `azurerm_cosmosdb_cassandra_datacenter`
    * `azurerm_cosmosdb_cassandra_keyspace`
    * `azurerm_cosmosdb_cassandra_table`
    * `azurerm_cosmosdb_gremlin_database`
    * `azurerm_cosmosdb_gremlin_graph`
    * `azurerm_cosmosdb_mongo_collection`
    * `azurerm_cosmosdb_mongo_database`
    * `azurerm_cosmosdb_sql_container`
    * `azurerm_cosmosdb_sql_database`
    * `azurerm_cosmosdb_sql_function`
    * `azurerm_cosmosdb_sql_role_assignment`
    * `azurerm_cosmosdb_sql_role_definition`
    * `azurerm_cosmosdb_sql_stored_procedure`
    * `azurerm_cosmosdb_sql_trigger`
    * `azurerm_cosmosdb_table`
*   `cosmosdb_postgresql`
    * `azurerm_cosmosdb_postgresql_cluster`
    * `azurerm_cosmosdb_postgresql_coordinator_configuration`
    * `azurerm_cosmosdb_postgresql_firewall_rule`
    * `azurerm_cosmosdb_postgresql_node_configuration`
    * `azurerm_cosmosdb_postgresql_role`
*   `data_factory`
    * `azurerm_data_factory`
    * `azurerm_data_factory_custom_dataset`
    * `azurerm_data_factory_data_flow`
    * `azurerm_data_factory_dataset_azure_blob`
    * `azurerm_data_factory_dataset_azure_sql_table`
    * `azurerm_data_factory_dataset_binary`
    * `azurerm_data_factory_dataset_cosmosdb_sqlapi`
    * `azurerm_data_factory_dataset_delimited_text`
    * `azurerm_data_factory_dataset_http`
    * `azurerm_data_factory_dataset_json`
    * `azurerm_data_factory_dataset_mysql`
    * `azurerm_data_factory_dataset_parquet`
    * `azurerm_data_factory_dataset_postgresql`
    * `azurerm_data_factory_dataset_snowflake`
    * `azurerm_data_factory_dataset_sql_server_table`
    * `azurerm_data_factory_integration_runtime_azure`
    * `azurerm_data_factory_integration_runtime_azure_ssis`
    * `azurerm_data_factory_integration_runtime_managed`
    * `azurerm_data_factory_integration_runtime_self_hosted`
    * `azurerm_data_factory_linked_custom_service`
    * `azurerm_data_factory_linked_service_azure_blob_storage`
    * `azurerm_data_factory_linked_service_azure_databricks`
    * `azurerm_data_factory_linked_service_azure_file_storage`
    * `azurerm_data_factory_linked_service_azure_function`
    * `azurerm_data_factory_linked_service_azure_search`
    * `azurerm_data_factory_linked_service_azure_sql_database`
    * `azurerm_data_factory_linked_service_azure_table_storage`
    * `azurerm_data_factory_linked_service_cosmosdb`
    * `azurerm_data_factory_linked_service_cosmosdb_mongoapi`
    * `azurerm_data_factory_linked_service_odbc`
    * `azurerm_data_factory_linked_service_data_lake_storage_gen2`
    * `azurerm_data_factory_linked_service_key_vault`
    * `azurerm_data_factory_linked_service_kusto`
    * `azurerm_data_factory_linked_service_mysql`
    * `azurerm_data_factory_linked_service_odata`
    * `azurerm_data_factory_linked_service_postgresql`
    * `azurerm_data_factory_linked_service_sftp`
    * `azurerm_data_factory_linked_service_snowflake`
    * `azurerm_data_factory_linked_service_sql_server`
    * `azurerm_data_factory_linked_service_synapse`
    * `azurerm_data_factory_linked_service_web`
    * `azurerm_data_factory_flowlet_data_flow`
    * `azurerm_data_factory_managed_private_endpoint`
    * `azurerm_data_factory_pipeline`
    * `azurerm_data_factory_trigger_blob_event`
    * `azurerm_data_factory_trigger_custom_event`
    * `azurerm_data_factory_trigger_schedule`
    * `azurerm_data_factory_trigger_tumbling_window`
*   `data_protection`
    * `azurerm_data_protection_backup_instance_blob_storage`
    * `azurerm_data_protection_backup_instance_data_lake_storage`
    * `azurerm_data_protection_backup_instance_disk`
    * `azurerm_data_protection_backup_instance_kubernetes_cluster`
    * `azurerm_data_protection_backup_instance_mysql_flexible_server`
    * `azurerm_data_protection_backup_instance_postgresql_flexible_server`
    * `azurerm_data_protection_backup_policy_blob_storage`
    * `azurerm_data_protection_backup_policy_cosmosdb_account`
    * `azurerm_data_protection_backup_policy_data_lake_storage`
    * `azurerm_data_protection_backup_policy_disk`
    * `azurerm_data_protection_backup_policy_kubernetes_cluster`
    * `azurerm_data_protection_backup_policy_mysql_flexible_server`
    * `azurerm_data_protection_backup_policy_postgresql_flexible_server`
    * `azurerm_data_protection_backup_vault`
    * `azurerm_data_protection_resource_guard`
*   `data_share`
    * `azurerm_data_share`
    * `azurerm_data_share_account`
    * `azurerm_data_share_dataset_blob_storage`
    * `azurerm_data_share_dataset_data_lake_gen2`
    * `azurerm_data_share_dataset_kusto_cluster`
    * `azurerm_data_share_dataset_kusto_database`
*   `database`
    * `azurerm_mariadb_configuration`
    * `azurerm_mariadb_database`
    * `azurerm_mariadb_firewall_rule`
    * `azurerm_mariadb_server`
    * `azurerm_mariadb_virtual_network_rule`
    * `azurerm_mysql_configuration`
    * `azurerm_mysql_database`
    * `azurerm_mysql_firewall_rule`
    * `azurerm_mysql_server`
    * `azurerm_mysql_virtual_network_rule`
    * `azurerm_postgresql_configuration`
    * `azurerm_postgresql_database`
    * `azurerm_postgresql_firewall_rule`
    * `azurerm_postgresql_server`
    * `azurerm_postgresql_virtual_network_rule`
    * `azurerm_sql_active_directory_administrator`
    * `azurerm_sql_database`
    * `azurerm_sql_elasticpool`
    * `azurerm_sql_failover_group`
    * `azurerm_sql_firewall_rule`
    * `azurerm_sql_server`
    * `azurerm_sql_virtual_network_rule`
*   `database_flexible`
    * `azurerm_mysql_flexible_database`
    * `azurerm_mysql_flexible_server`
    * `azurerm_mysql_flexible_server_configuration`
    * `azurerm_mysql_flexible_server_firewall_rule`
    * `azurerm_postgresql_flexible_server`
    * `azurerm_postgresql_flexible_server_configuration`
    * `azurerm_postgresql_flexible_server_database`
    * `azurerm_postgresql_flexible_server_firewall_rule`
*   `database_migration`
    * `azurerm_database_migration_project`
    * `azurerm_database_migration_service`
*   `elastic`
    * `azurerm_elastic_cloud_elasticsearch`
*   `kusto`
    * `azurerm_kusto_attached_database_configuration`
    * `azurerm_kusto_cluster`
    * `azurerm_kusto_cluster_managed_private_endpoint`
    * `azurerm_kusto_cluster_principal_assignment`
    * `azurerm_kusto_cosmosdb_data_connection`
    * `azurerm_kusto_database`
    * `azurerm_kusto_database_principal_assignment`
    * `azurerm_kusto_eventgrid_data_connection`
    * `azurerm_kusto_eventhub_data_connection`
    * `azurerm_kusto_iothub_data_connection`
    * `azurerm_kusto_script`
*   `mongo_cluster`
    * `azurerm_mongo_cluster`
    * `azurerm_mongo_cluster_firewall_rule`
    * `azurerm_mongo_cluster_user`
*   `mssql`
    * `azurerm_mssql_database_extended_auditing_policy`
    * `azurerm_mssql_elasticpool`
    * `azurerm_mssql_failover_group`
    * `azurerm_mssql_job`
    * `azurerm_mssql_job_agent`
    * `azurerm_mssql_job_credential`
    * `azurerm_mssql_job_step`
    * `azurerm_mssql_job_target_group`
    * `azurerm_mssql_managed_database`
    * `azurerm_mssql_managed_instance`
    * `azurerm_mssql_managed_instance_active_directory_administrator`
    * `azurerm_mssql_managed_instance_failover_group`
    * `azurerm_mssql_managed_instance_security_alert_policy`
    * `azurerm_mssql_managed_instance_transparent_data_encryption`
    * `azurerm_mssql_managed_instance_vulnerability_assessment`
    * `azurerm_mssql_outbound_firewall_rule`
    * `azurerm_mssql_server_dns_alias`
    * `azurerm_mssql_server_extended_auditing_policy`
    * `azurerm_mssql_server_microsoft_support_auditing_policy`
    * `azurerm_mssql_server_security_alert_policy`
    * `azurerm_mssql_server_transparent_data_encryption`
    * `azurerm_mssql_server_vulnerability_assessment`
    * `azurerm_mssql_virtual_network_rule`
*   `mssql_virtual_machine`
    * `azurerm_mssql_virtual_machine`
    * `azurerm_mssql_virtual_machine_availability_group_listener`
    * `azurerm_mssql_virtual_machine_group`
*   `databricks`
    * `azurerm_databricks_access_connector`
    * `azurerm_databricks_virtual_network_peering`
    * `azurerm_databricks_workspace`
*   `datadog`
    * `azurerm_datadog_monitor`
    * `azurerm_datadog_monitor_sso_configuration`
    * `azurerm_datadog_monitor_tag_rule`
*   `dynatrace`
    * `azurerm_dynatrace_monitor`
    * `azurerm_dynatrace_tag_rules`
*   `dedicated_hsm`
    * `azurerm_dedicated_hardware_security_module`
*   `dev_center`
    * `azurerm_dev_center`
    * `azurerm_dev_center_attached_network`
    * `azurerm_dev_center_catalog`
    * `azurerm_dev_center_dev_box_definition`
    * `azurerm_dev_center_environment_type`
    * `azurerm_dev_center_gallery`
    * `azurerm_dev_center_network_connection`
    * `azurerm_dev_center_project`
    * `azurerm_dev_center_project_environment_type`
    * `azurerm_dev_center_project_pool`
*   `dev_test`
    * `azurerm_dev_test_global_vm_shutdown_schedule`
    * `azurerm_dev_test_lab`
    * `azurerm_dev_test_linux_virtual_machine`
    * `azurerm_dev_test_policy`
    * `azurerm_dev_test_schedule`
    * `azurerm_dev_test_virtual_network`
    * `azurerm_dev_test_windows_virtual_machine`
*   `disk`
    * `azurerm_managed_disk`
*   `dns`
    * `azurerm_dns_a_record`
    * `azurerm_dns_aaaa_record`
    * `azurerm_dns_caa_record`
    * `azurerm_dns_cname_record`
    * `azurerm_dns_mx_record`
    * `azurerm_dns_ns_record`
    * `azurerm_dns_ptr_record`
    * `azurerm_dns_srv_record`
    * `azurerm_dns_txt_record`
    * `azurerm_dns_zone`
*   `eventgrid`
    * `azurerm_eventgrid_domain`
    * `azurerm_eventgrid_domain_topic`
    * `azurerm_eventgrid_event_subscription`
    * `azurerm_eventgrid_system_topic`
    * `azurerm_eventgrid_system_topic_event_subscription`
    * `azurerm_eventgrid_topic`
*   `domain_services`
    * `azurerm_active_directory_domain_service`
    * `azurerm_active_directory_domain_service_replica_set`
*   `eventhub`
    * `azurerm_eventhub`
    * `azurerm_eventhub_authorization_rule`
    * `azurerm_eventhub_cluster`
    * `azurerm_eventhub_consumer_group`
    * `azurerm_eventhub_namespace`
    * `azurerm_eventhub_namespace_authorization_rule`
    * `azurerm_eventhub_namespace_disaster_recovery_config`
    * `azurerm_eventhub_namespace_schema_group`
*   `federated_identity_credential`
    * `azurerm_federated_identity_credential`
*   `iot_security`
    * `azurerm_iot_security_device_group`
    * `azurerm_iot_security_solution`
*   `iotcentral`
    * `azurerm_iotcentral_application`
*   `iothub`
    * `azurerm_iothub`
    * `azurerm_iothub_certificate`
    * `azurerm_iothub_consumer_group`
    * `azurerm_iothub_shared_access_policy`
*   `iothub_device_update`
    * `azurerm_iothub_device_update_account`
    * `azurerm_iothub_device_update_instance`
*   `keyvault`
    * `azurerm_key_vault`
    * `azurerm_key_vault_managed_hardware_security_module`
*   `express_route`
    * `azurerm_express_route_circuit`
    * `azurerm_express_route_circuit_authorization`
    * `azurerm_express_route_circuit_connection`
    * `azurerm_express_route_circuit_peering`
    * `azurerm_express_route_connection`
    * `azurerm_express_route_gateway`
    * `azurerm_express_route_port`
    * `azurerm_express_route_port_authorization`
*   `firewall`
    * `azurerm_firewall`
    * `azurerm_firewall_application_rule_collection`
    * `azurerm_firewall_nat_rule_collection`
    * `azurerm_firewall_network_rule_collection`
    * `azurerm_firewall_policy`
    * `azurerm_firewall_policy_rule_collection_group`
    * `azurerm_ip_group`
    * `azurerm_ip_group_cidr`
*   `frontdoor`
    * `azurerm_cdn_frontdoor_firewall_policy`
    * `azurerm_frontdoor`
    * `azurerm_frontdoor_custom_https_configuration`
    * `azurerm_frontdoor_firewall_policy`
    * `azurerm_frontdoor_rules_engine`
*   `fluid_relay`
    * `azurerm_fluid_relay_server`
*   `graph_services`
    * `azurerm_graph_services_account`
*   `hdinsight`
    * `azurerm_hdinsight_hadoop_cluster`
    * `azurerm_hdinsight_hbase_cluster`
    * `azurerm_hdinsight_interactive_query_cluster`
    * `azurerm_hdinsight_kafka_cluster`
    * `azurerm_hdinsight_spark_cluster`
*   `kubernetes`
    * `azurerm_kubernetes_automatic_cluster`
    * `azurerm_kubernetes_cluster`
    * `azurerm_kubernetes_cluster_deployment_safeguard`
    * `azurerm_kubernetes_cluster_extension`
    * `azurerm_kubernetes_cluster_node_pool`
    * `azurerm_kubernetes_cluster_trusted_access_role_binding`
    * `azurerm_kubernetes_flux_configuration`
*   `kubernetes_fleet`
    * `azurerm_kubernetes_fleet_manager`
    * `azurerm_kubernetes_fleet_member`
    * `azurerm_kubernetes_fleet_update_run`
    * `azurerm_kubernetes_fleet_update_strategy`
*   `lighthouse`
    * `azurerm_lighthouse_assignment`
    * `azurerm_lighthouse_definition`
*   `load_balancer`
    * `azurerm_lb`
    * `azurerm_lb_backend_address_pool`
    * `azurerm_lb_backend_address_pool_address`
    * `azurerm_lb_nat_pool`
    * `azurerm_lb_nat_rule`
    * `azurerm_lb_outbound_rule`
    * `azurerm_lb_probe`
*   `machine_learning`
    * `azurerm_machine_learning_compute_cluster`
    * `azurerm_machine_learning_compute_instance`
    * `azurerm_machine_learning_inference_cluster`
    * `azurerm_machine_learning_synapse_spark`
    * `azurerm_machine_learning_workspace`
*   `maintenance`
    * `azurerm_maintenance_assignment_dedicated_host`
    * `azurerm_maintenance_assignment_dynamic_scope`
    * `azurerm_maintenance_assignment_virtual_machine`
    * `azurerm_maintenance_assignment_virtual_machine_scale_set`
    * `azurerm_maintenance_configuration`
*   `managed_compute`
    * `azurerm_managed_application`
    * `azurerm_managed_application_definition`
    * `azurerm_managed_devops_pool`
    * `azurerm_managed_redis`
    * `azurerm_redhat_openshift_cluster`
    * `azurerm_lb_rule`
*   `nat_gateway`
    * `azurerm_nat_gateway`
    * `azurerm_nat_gateway_public_ip_association`
    * `azurerm_nat_gateway_public_ip_prefix_association`
*   `network_basics`
    * `azurerm_bastion_host`
    * `azurerm_custom_ip_prefix`
    * `azurerm_network_connection_monitor`
    * `azurerm_network_ddos_protection_plan`
    * `azurerm_network_profile`
*   `network_function`
    * `azurerm_network_function_azure_traffic_collector`
    * `azurerm_network_function_collector_policy`
*   `network_gateway`
    * `azurerm_local_network_gateway`
    * `azurerm_virtual_network_gateway`
    * `azurerm_virtual_network_gateway_connection`
    * `azurerm_virtual_network_gateway_nat_rule`
*   `network_manager`
    * `azurerm_network_manager`
    * `azurerm_network_manager_admin_rule`
    * `azurerm_network_manager_admin_rule_collection`
    * `azurerm_network_manager_connectivity_configuration`
    * `azurerm_network_manager_network_group`
    * `azurerm_network_manager_scope_connection`
    * `azurerm_network_manager_security_admin_configuration`
    * `azurerm_network_manager_static_member`
    * `azurerm_network_manager_subscription_connection`
*   `logic_app`
    * `azurerm_logic_app_action_custom`
    * `azurerm_logic_app_action_http`
    * `azurerm_logic_app_integration_account`
    * `azurerm_logic_app_integration_account_agreement`
    * `azurerm_logic_app_integration_account_assembly`
    * `azurerm_logic_app_integration_account_batch_configuration`
    * `azurerm_logic_app_integration_account_certificate`
    * `azurerm_logic_app_integration_account_map`
    * `azurerm_logic_app_integration_account_partner`
    * `azurerm_logic_app_integration_account_schema`
    * `azurerm_logic_app_integration_account_session`
    * `azurerm_logic_app_trigger_custom`
    * `azurerm_logic_app_trigger_http_request`
    * `azurerm_logic_app_trigger_recurrence`
    * `azurerm_logic_app_workflow`
*   `maps`
    * `azurerm_maps_account`
*   `management_group`
    * `azurerm_management_group`
    * `azurerm_management_group_subscription_association`
*   `network_interface`
    * `azurerm_network_interface`
    * `azurerm_network_interface_application_gateway_backend_address_pool_association`
    * `azurerm_network_interface_application_security_group_association`
    * `azurerm_network_interface_backend_address_pool_association`
    * `azurerm_network_interface_nat_rule_association`
    * `azurerm_network_interface_security_group_association`
*   `network_security_group`
    * `azurerm_network_security_group`
    * `azurerm_network_security_rule`
*   `network_watcher`
    * `azurerm_network_packet_capture`
    * `azurerm_network_watcher`
    * `azurerm_network_watcher_flow_log`
*   `new_relic`
    * `azurerm_new_relic_monitor`
    * `azurerm_new_relic_tag_rule`
*   `nginx`
    * `azurerm_nginx_certificate`
    * `azurerm_nginx_configuration`
    * `azurerm_nginx_deployment`
*   `oracle`
    * `azurerm_oracle_autonomous_database`
    * `azurerm_oracle_autonomous_database_backup`
    * `azurerm_oracle_autonomous_database_clone_from_backup`
    * `azurerm_oracle_autonomous_database_clone_from_database`
    * `azurerm_oracle_autonomous_database_cross_region_disaster_recovery`
    * `azurerm_oracle_cloud_vm_cluster`
    * `azurerm_oracle_exadata_infrastructure`
    * `azurerm_oracle_exascale_database_storage_vault`
*   `palo_alto`
    * `azurerm_palo_alto_local_rulestack`
    * `azurerm_palo_alto_local_rulestack_certificate`
    * `azurerm_palo_alto_local_rulestack_fqdn_list`
    * `azurerm_palo_alto_local_rulestack_prefix_list`
    * `azurerm_palo_alto_local_rulestack_rule`
    * `azurerm_palo_alto_next_generation_firewall_virtual_hub_local_rulestack`
    * `azurerm_palo_alto_next_generation_firewall_virtual_hub_panorama`
    * `azurerm_palo_alto_next_generation_firewall_virtual_hub_strata_cloud_manager`
    * `azurerm_palo_alto_next_generation_firewall_virtual_network_local_rulestack`
    * `azurerm_palo_alto_next_generation_firewall_virtual_network_panorama`
    * `azurerm_palo_alto_next_generation_firewall_virtual_network_strata_cloud_manager`
*   `notification_hub`
    * `azurerm_notification_hub`
    * `azurerm_notification_hub_authorization_rule`
    * `azurerm_notification_hub_namespace`
*   `private_dns`
    * `azurerm_private_dns_a_record`
    * `azurerm_private_dns_aaaa_record`
    * `azurerm_private_dns_cname_record`
    * `azurerm_private_dns_mx_record`
    * `azurerm_private_dns_ptr_record`
    * `azurerm_private_dns_srv_record`
    * `azurerm_private_dns_txt_record`
    * `azurerm_private_dns_zone`
    * `azurerm_private_dns_zone_virtual_network_link`
*   `private_dns_resolver`
    * `azurerm_private_dns_resolver`
    * `azurerm_private_dns_resolver_dns_forwarding_ruleset`
    * `azurerm_private_dns_resolver_forwarding_rule`
    * `azurerm_private_dns_resolver_inbound_endpoint`
    * `azurerm_private_dns_resolver_outbound_endpoint`
    * `azurerm_private_dns_resolver_virtual_network_link`
*   `platform_misc`
    * `azurerm_advisor_suppression`
    * `azurerm_blueprint_assignment`
    * `azurerm_custom_provider`
    * `azurerm_databox_edge_device`
    * `azurerm_extended_location_custom_location`
    * `azurerm_fabric_capacity`
    * `azurerm_healthbot`
    * `azurerm_load_test`
    * `azurerm_marketplace_agreement`
    * `azurerm_playwright_workspace`
    * `azurerm_policy_virtual_machine_configuration_assignment`
    * `azurerm_portal_dashboard`
    * `azurerm_portal_tenant_configuration`
    * `azurerm_powerbi_embedded`
    * `azurerm_resource_deployment_script_azure_cli`
    * `azurerm_resource_deployment_script_azure_power_shell`
    * `azurerm_trusted_signing_account`
    * `azurerm_user_assigned_identity`
    * `azurerm_workloads_sap_discovery_virtual_instance`
    * `azurerm_workloads_sap_single_node_virtual_instance`
    * `azurerm_workloads_sap_three_tier_virtual_instance`
*   `policy`
    * `azurerm_management_group_policy_assignment`
    * `azurerm_management_group_policy_remediation`
    * `azurerm_management_group_policy_set_definition`
    * `azurerm_policy_definition`
    * `azurerm_policy_set_definition`
    * `azurerm_resource_group_policy_assignment`
    * `azurerm_resource_group_policy_remediation`
    * `azurerm_subscription_policy_assignment`
    * `azurerm_subscription_policy_remediation`
*   `private_endpoint`
    * `azurerm_private_endpoint`
    * `azurerm_private_endpoint_application_security_group_association`
    * `azurerm_private_link_service`
*   `public_ip`
    * `azurerm_public_ip`
    * `azurerm_public_ip_prefix`
*   `purview`
    * `azurerm_purview_account`
*   `rbac`
    * `azurerm_marketplace_role_assignment`
    * `azurerm_pim_active_role_assignment`
    * `azurerm_pim_eligible_role_assignment`
    * `azurerm_role_assignment`
    * `azurerm_role_definition`
    * `azurerm_role_management_policy`
*   `redis`
    * `azurerm_redis_cache`
    * `azurerm_redis_firewall_rule`
    * `azurerm_redis_linked_server`
*   `relay`
    * `azurerm_relay_hybrid_connection`
    * `azurerm_relay_hybrid_connection_authorization_rule`
    * `azurerm_relay_namespace`
    * `azurerm_relay_namespace_authorization_rule`
*   `resource_group`
    * `azurerm_management_lock`
    * `azurerm_resource_group`
*   `route_table`
    * `azurerm_route`
    * `azurerm_route_filter`
    * `azurerm_route_table`
*   `scaleset`
    * `azurerm_virtual_machine_scale_set`
*   `scvmm`
    * `azurerm_system_center_virtual_machine_manager_availability_set`
    * `azurerm_system_center_virtual_machine_manager_cloud`
    * `azurerm_system_center_virtual_machine_manager_server`
    * `azurerm_system_center_virtual_machine_manager_virtual_machine_instance`
    * `azurerm_system_center_virtual_machine_manager_virtual_machine_instance_guest_agent`
    * `azurerm_system_center_virtual_machine_manager_virtual_machine_template`
    * `azurerm_system_center_virtual_machine_manager_virtual_network`
*   `service_fabric`
    * `azurerm_service_fabric_cluster`
    * `azurerm_service_fabric_managed_cluster`
*   `stack_hci`
    * `azurerm_stack_hci_cluster`
    * `azurerm_stack_hci_deployment_setting`
    * `azurerm_stack_hci_extension`
    * `azurerm_stack_hci_logical_network`
    * `azurerm_stack_hci_marketplace_gallery_image`
    * `azurerm_stack_hci_network_interface`
    * `azurerm_stack_hci_storage_path`
    * `azurerm_stack_hci_virtual_hard_disk`
*   `search`
    * `azurerm_search_service`
    * `azurerm_search_shared_private_link_service`
*   `security_center`
    * `azurerm_security_center_assessment`
    * `azurerm_security_center_assessment_policy`
    * `azurerm_security_center_automation`
    * `azurerm_security_center_contact`
    * `azurerm_security_center_server_vulnerability_assessments_setting`
    * `azurerm_security_center_setting`
    * `azurerm_security_center_subscription_pricing`
    * `azurerm_security_center_workspace`
*   `sentinel`
    * `azurerm_sentinel_alert_rule_fusion`
    * `azurerm_sentinel_alert_rule_machine_learning_behavior_analytics`
    * `azurerm_sentinel_alert_rule_ms_security_incident`
    * `azurerm_sentinel_alert_rule_nrt`
    * `azurerm_sentinel_alert_rule_scheduled`
    * `azurerm_sentinel_alert_rule_threat_intelligence`
    * `azurerm_sentinel_automation_rule`
    * `azurerm_sentinel_data_connector_aws_cloud_trail`
    * `azurerm_sentinel_data_connector_aws_s3`
    * `azurerm_sentinel_data_connector_azure_active_directory`
    * `azurerm_sentinel_data_connector_azure_advanced_threat_protection`
    * `azurerm_sentinel_data_connector_azure_security_center`
    * `azurerm_sentinel_data_connector_dynamics_365`
    * `azurerm_sentinel_data_connector_iot`
    * `azurerm_sentinel_data_connector_microsoft_cloud_app_security`
    * `azurerm_sentinel_data_connector_microsoft_defender_advanced_threat_protection`
    * `azurerm_sentinel_data_connector_microsoft_threat_intelligence`
    * `azurerm_sentinel_data_connector_microsoft_threat_protection`
    * `azurerm_sentinel_data_connector_office_365`
    * `azurerm_sentinel_data_connector_office_365_project`
    * `azurerm_sentinel_data_connector_office_atp`
    * `azurerm_sentinel_data_connector_office_irm`
    * `azurerm_sentinel_data_connector_office_power_bi`
    * `azurerm_sentinel_data_connector_threat_intelligence`
    * `azurerm_sentinel_data_connector_threat_intelligence_taxii`
    * `azurerm_sentinel_log_analytics_workspace_onboarding`
    * `azurerm_sentinel_threat_intelligence_indicator`
    * `azurerm_sentinel_watchlist`
    * `azurerm_sentinel_watchlist_item`
*   `servicebus`
    * `azurerm_servicebus_namespace`
    * `azurerm_servicebus_namespace_authorization_rule`
    * `azurerm_servicebus_namespace_disaster_recovery_config`
    * `azurerm_servicebus_queue`
    * `azurerm_servicebus_queue_authorization_rule`
    * `azurerm_servicebus_subscription`
    * `azurerm_servicebus_subscription_rule`
    * `azurerm_servicebus_topic`
    * `azurerm_servicebus_topic_authorization_rule`
*   `signalr`
    * `azurerm_signalr_service`
    * `azurerm_signalr_service_custom_certificate`
    * `azurerm_signalr_service_custom_domain`
    * `azurerm_signalr_shared_private_link_resource`
*   `spring_cloud`
    * `azurerm_spring_cloud_accelerator`
    * `azurerm_spring_cloud_active_deployment`
    * `azurerm_spring_cloud_api_portal`
    * `azurerm_spring_cloud_api_portal_custom_domain`
    * `azurerm_spring_cloud_app`
    * `azurerm_spring_cloud_app_cosmosdb_association`
    * `azurerm_spring_cloud_app_dynamics_application_performance_monitoring`
    * `azurerm_spring_cloud_app_mysql_association`
    * `azurerm_spring_cloud_app_redis_association`
    * `azurerm_spring_cloud_application_insights_application_performance_monitoring`
    * `azurerm_spring_cloud_application_live_view`
    * `azurerm_spring_cloud_build_deployment`
    * `azurerm_spring_cloud_build_pack_binding`
    * `azurerm_spring_cloud_builder`
    * `azurerm_spring_cloud_certificate`
    * `azurerm_spring_cloud_configuration_service`
    * `azurerm_spring_cloud_connection`
    * `azurerm_spring_cloud_container_deployment`
    * `azurerm_spring_cloud_custom_domain`
    * `azurerm_spring_cloud_customized_accelerator`
    * `azurerm_spring_cloud_dev_tool_portal`
    * `azurerm_spring_cloud_dynatrace_application_performance_monitoring`
    * `azurerm_spring_cloud_elastic_application_performance_monitoring`
    * `azurerm_spring_cloud_gateway`
    * `azurerm_spring_cloud_gateway_custom_domain`
    * `azurerm_spring_cloud_gateway_route_config`
    * `azurerm_spring_cloud_java_deployment`
    * `azurerm_spring_cloud_new_relic_application_performance_monitoring`
    * `azurerm_spring_cloud_service`
    * `azurerm_spring_cloud_storage`
*   `storage_account`
    * `azurerm_storage_account`
    * `azurerm_storage_blob`
    * `azurerm_storage_container`
*   `stream_analytics`
    * `azurerm_stream_analytics_cluster`
    * `azurerm_stream_analytics_function_javascript_uda`
    * `azurerm_stream_analytics_function_javascript_udf`
    * `azurerm_stream_analytics_job`
    * `azurerm_stream_analytics_job_storage_account`
    * `azurerm_stream_analytics_managed_private_endpoint`
    * `azurerm_stream_analytics_output_blob`
    * `azurerm_stream_analytics_output_cosmosdb`
    * `azurerm_stream_analytics_output_eventhub`
    * `azurerm_stream_analytics_output_function`
    * `azurerm_stream_analytics_output_mssql`
    * `azurerm_stream_analytics_output_powerbi`
    * `azurerm_stream_analytics_output_servicebus_queue`
    * `azurerm_stream_analytics_output_servicebus_topic`
    * `azurerm_stream_analytics_output_synapse`
    * `azurerm_stream_analytics_output_table`
    * `azurerm_stream_analytics_reference_input_blob`
    * `azurerm_stream_analytics_reference_input_mssql`
    * `azurerm_stream_analytics_stream_input_blob`
    * `azurerm_stream_analytics_stream_input_eventhub`
    * `azurerm_stream_analytics_stream_input_eventhub_v2`
    * `azurerm_stream_analytics_stream_input_iothub`
*   `subnet`
    * `azurerm_subnet`
    * `azurerm_subnet_nat_gateway_association`
    * `azurerm_subnet_network_security_group_association`
    * `azurerm_subnet_route_table_association`
    * `azurerm_subnet_service_endpoint_storage_policy`
*   `synapse`
    * `azurerm_synapse_firewall_rule`
    * `azurerm_synapse_integration_runtime_azure`
    * `azurerm_synapse_integration_runtime_self_hosted`
    * `azurerm_synapse_managed_private_endpoint`
    * `azurerm_synapse_private_link_hub`
    * `azurerm_synapse_spark_pool`
    * `azurerm_synapse_sql_pool`
    * `azurerm_synapse_sql_pool_extended_auditing_policy`
    * `azurerm_synapse_sql_pool_security_alert_policy`
    * `azurerm_synapse_sql_pool_vulnerability_assessment`
    * `azurerm_synapse_sql_pool_workload_classifier`
    * `azurerm_synapse_sql_pool_workload_group`
    * `azurerm_synapse_workspace`
    * `azurerm_synapse_workspace_extended_auditing_policy`
    * `azurerm_synapse_workspace_key`
    * `azurerm_synapse_workspace_security_alert_policy`
    * `azurerm_synapse_workspace_vulnerability_assessment`
*   `traffic_manager`
    * `azurerm_traffic_manager_azure_endpoint`
    * `azurerm_traffic_manager_external_endpoint`
    * `azurerm_traffic_manager_nested_endpoint`
    * `azurerm_traffic_manager_profile`
*   `template_deployment`
    * `azurerm_management_group_template_deployment`
    * `azurerm_resource_group_template_deployment`
    * `azurerm_subscription_template_deployment`
    * `azurerm_tenant_template_deployment`
*   `virtual_machine`
    * `azurerm_ssh_public_key`
    * `azurerm_virtual_machine`
*   `virtual_desktop`
    * `azurerm_virtual_desktop_application`
    * `azurerm_virtual_desktop_application_group`
    * `azurerm_virtual_desktop_host_pool`
    * `azurerm_virtual_desktop_scaling_plan`
    * `azurerm_virtual_desktop_workspace`
*   `virtual_machine_child`
    * `azurerm_virtual_machine_data_disk_attachment`
    * `azurerm_virtual_machine_extension`
    * `azurerm_virtual_machine_implicit_data_disk_from_source`
    * `azurerm_virtual_machine_run_command`
    * `azurerm_virtual_machine_scale_set_extension`
*   `virtual_machine_scale_set`
    * `azurerm_linux_virtual_machine_scale_set`
    * `azurerm_orchestrated_virtual_machine_scale_set`
    * `azurerm_virtual_machine_scale_set_standby_pool`
    * `azurerm_windows_virtual_machine_scale_set`
*   `virtual_network`
    * `azurerm_virtual_network`
*   `vmware`
    * `azurerm_vmware_cluster`
    * `azurerm_vmware_express_route_authorization`
    * `azurerm_vmware_netapp_volume_attachment`
    * `azurerm_vmware_private_cloud`
*   `storage_extra`
    * `azurerm_storage_account_local_user`
    * `azurerm_storage_blob_inventory_policy`
    * `azurerm_storage_container_immutability_policy`
    * `azurerm_storage_encryption_scope`
    * `azurerm_storage_management_policy`
    * `azurerm_storage_queue`
    * `azurerm_storage_share`
    * `azurerm_storage_table`
*   `netapp`
    * `azurerm_netapp_account`
    * `azurerm_netapp_backup_policy`
    * `azurerm_netapp_backup_vault`
    * `azurerm_netapp_pool`
    * `azurerm_netapp_snapshot`
    * `azurerm_netapp_snapshot_policy`
    * `azurerm_netapp_volume`
    * `azurerm_netapp_volume_group_oracle`
    * `azurerm_netapp_volume_group_sap_hana`
*   `elastic_san`
    * `azurerm_elastic_san`
    * `azurerm_elastic_san_volume`
    * `azurerm_elastic_san_volume_group`
*   `recovery_services`
    * `azurerm_backup_container_storage_account`
    * `azurerm_backup_policy_file_share`
    * `azurerm_backup_policy_vm`
    * `azurerm_backup_policy_vm_workload`
    * `azurerm_backup_protected_file_share`
    * `azurerm_backup_protected_vm`
    * `azurerm_recovery_services_vault`
    * `azurerm_recovery_services_vault_resource_guard_association`
*   `site_recovery`
    * `azurerm_site_recovery_fabric`
    * `azurerm_site_recovery_hyperv_replication_policy`
    * `azurerm_site_recovery_network_mapping`
    * `azurerm_site_recovery_protection_container`
    * `azurerm_site_recovery_protection_container_mapping`
    * `azurerm_site_recovery_replicated_vm`
    * `azurerm_site_recovery_replication_policy`
    * `azurerm_site_recovery_replication_recovery_plan`
    * `azurerm_site_recovery_vmware_replicated_vm`
    * `azurerm_site_recovery_vmware_replication_policy`
*   `monitor`
    * `azurerm_monitor_action_group`
    * `azurerm_monitor_activity_log_alert`
    * `azurerm_monitor_alert_processing_rule_action_group`
    * `azurerm_monitor_alert_processing_rule_suppression`
    * `azurerm_monitor_autoscale_setting`
    * `azurerm_monitor_data_collection_endpoint`
    * `azurerm_monitor_data_collection_rule`
    * `azurerm_monitor_data_collection_rule_association`
    * `azurerm_monitor_metric_alert`
    * `azurerm_monitor_private_link_scope`
    * `azurerm_monitor_private_link_scoped_service`
    * `azurerm_monitor_scheduled_query_rules_alert`
    * `azurerm_monitor_scheduled_query_rules_alert_v2`
    * `azurerm_monitor_scheduled_query_rules_log`
*   `log_analytics`
    * `azurerm_log_analytics_cluster`
    * `azurerm_log_analytics_data_export_rule`
    * `azurerm_log_analytics_datasource_windows_event`
    * `azurerm_log_analytics_datasource_windows_performance_counter`
    * `azurerm_log_analytics_linked_service`
    * `azurerm_log_analytics_linked_storage_account`
    * `azurerm_log_analytics_saved_search`
    * `azurerm_log_analytics_storage_insights`
    * `azurerm_log_analytics_workspace`
    * `azurerm_log_analytics_workspace_table_custom_log`
*   `dashboard_grafana`
    * `azurerm_dashboard_grafana`
    * `azurerm_dashboard_grafana_managed_private_endpoint`
*   `chaos_studio`
    * `azurerm_chaos_studio_experiment`
*   `healthcare`
    * `azurerm_healthcare_dicom_service`
    * `azurerm_healthcare_fhir_service`
    * `azurerm_healthcare_medtech_service`
    * `azurerm_healthcare_medtech_service_fhir_destination`
    * `azurerm_healthcare_service`
    * `azurerm_healthcare_workspace`
*   `digital_twins`
    * `azurerm_digital_twins_endpoint_eventgrid`
    * `azurerm_digital_twins_endpoint_eventhub`
    * `azurerm_digital_twins_endpoint_servicebus`
    * `azurerm_digital_twins_instance`
    * `azurerm_digital_twins_time_series_database_connection`
*   `storage_mover`
    * `azurerm_storage_mover`
    * `azurerm_storage_mover_agent`
    * `azurerm_storage_mover_job_definition`
    * `azurerm_storage_mover_project`
*   `storage_sync`
    * `azurerm_storage_sync`
    * `azurerm_storage_sync_cloud_endpoint`
    * `azurerm_storage_sync_group`
    * `azurerm_storage_sync_server_endpoint`
*   `storage_discovery`
    * `azurerm_storage_discovery_workspace`

### Known limits — not importable

The following azurerm types are not generated, for the concrete reasons
stated. This list is mechanically reconciled with the provider type
universe: every azurerm resource type is either generated or listed here.


**API Management, App Service and integration**

* `azurerm_api_connection` — no Connections client or List operation for Microsoft.Web/connections in track-2 armappservice v1.0.0.
* `azurerm_api_management_api_operation_policy` — singleton policy document for an operation; APIM policy clients in armapimanagement v1.1.1 expose only Get/CreateOrUpdate/Delete, no List.
* `azurerm_api_management_api_policy` — singleton policy document for an API; APIPolicyClient is Get-only (no List) in armapimanagement v1.1.1.
* `azurerm_api_management_custom_domain` — no SDK client; custom domains are hostnameConfiguration properties embedded in the service resource, not an enumerable child resource.
* `azurerm_api_management_policy` — singleton policy document for the service (importer ID is the service ID itself); PolicyClient is Get-only (no List) in armapimanagement v1.1.1.
* `azurerm_api_management_policy_fragment` — no PolicyFragment client exists in armapimanagement v1.1.1 (latest stable).
* `azurerm_api_management_product_policy` — singleton policy document for a product; ProductPolicyClient is Get-only (no List) in armapimanagement v1.1.1.
* `azurerm_api_management_standalone_gateway` — SDK GatewayClient only lists gateways under a service (ListByService); no resource-group/subscription List exists for the top-level Microsoft.ApiManagement/gateways resource in armapimanagement v1.1.1.
* `azurerm_api_management_workspace` — _policy — no Workspace policy client exists in armapimanagement v1.1.1; workspace policies are also singleton Get-only documents.
* `azurerm_api_management_workspace_api_version_set` — no Workspace API version set client exists in armapimanagement v1.1.1.
* `azurerm_api_management_workspace_certificate` — no Workspace certificate client exists in armapimanagement v1.1.1.
* `azurerm_api_management_workspace_named_value` — no Workspace named value client exists in armapimanagement v1.1.1.
* `azurerm_api_management_workspace_policy` — no Workspace policy client exists in armapimanagement v1.1.1; workspace policies are also singleton Get-only documents.
* `azurerm_api_management_workspace_policy_fragment` — no Workspace policy fragment client exists in armapimanagement v1.1.1.
* `azurerm_app_configuration_feature` — data-plane object; provider importer ID is the store data-plane endpoint URL (https://<store>.azconfig.io/kv/.appconfig.featureflag%2F<key>?label=<label>), not an ARM resource ID.
* `azurerm_app_configuration_key` — data-plane object; provider importer ID is the store data-plane endpoint URL (https://<store>.azconfig.io/kv/<key>?label=<label>), not an ARM resource ID.
* `azurerm_app_service_certificate_binding` — association glue with a composite importer ID ("<hostNameBinding id>|<certificate id>"); no independent ARM resource exists to enumerate.
* `azurerm_app_service_managed_certificate` — managed certificates are the same Microsoft.Web/certificates ARM objects enumerated as azurerm_app_service_certificate; the track-2 list exposes no field distinguishing them.
* `azurerm_app_service_slot_virtual_network_swift_connection` — singleton slot config (config/virtualNetwork); SDK exposes only GetSwiftVirtualNetworkConnectionSlot, no List.
* `azurerm_app_service_source_control` — _slot — singleton per slot; SDK exposes only GetSourceControlSlot, no List, and the importer ID is the slot ID itself.
* `azurerm_app_service_source_control_slot` — singleton per slot; SDK exposes only GetSourceControlSlot, no List, and the importer ID is the slot ID itself.
* `azurerm_app_service_virtual_network_swift_connection` — singleton site config (config/virtualNetwork); SDK exposes only GetSwiftVirtualNetworkConnection, no List.
* `azurerm_application_insights_standard_web_test` — standard and classic web tests share the Microsoft.Insights/webTests ARM type and armapplicationinsights v1.2.0 (API 2020-02-02) exposes no distinguishing field (WebTestKind enum is only ping/multistep); enumerated once as azurerm_application_insights_web_test.
* `azurerm_application_insights_workbook_template` — no workbook-template client exists in armapplicationinsights v1.2.0.
* `azurerm_function_app_active_slot` — singleton per function app; importer ID is the app ID itself and the SDK exposes no List operation for an active-slot resource.
* `azurerm_function_app_flex_consumption` — armappservice v1.0.0 (API 2022-03-01) predates Flex Consumption; its Site model has no functionAppConfig field or other signal to distinguish flex apps, which are enumerated as azurerm_linux_function_app.
* `azurerm_function_app_hybrid_connection` — armappservice v1.0.0 generates ListHybridConnections/ListRelayServiceConnections with single-object (non-collection) response types, so site hybrid connections cannot be enumerated; the only hybrid-connection pager (PlansClient.ListHybridConnections) returns plan-scope IDs, not the per-site importer IDs.
* `azurerm_source_control_token` — subscription-level singleton (/providers/Microsoft.Web/sourceControls/GitHub); the SDK ListSourceControls operation lists source-control types, not user tokens.
* `azurerm_web_app_active_slot` — singleton per web app; importer ID is the app ID itself and the SDK exposes no List operation for an active-slot resource.
* `azurerm_web_app_hybrid_connection` — same SDK limitation as azurerm_function_app_hybrid_connection: ListHybridConnections/ListRelayServiceConnections have single-object response types in armappservice v1.0.0; plan-scope pager IDs do not match the per-site importer ID.

**Networking**

* `azurerm_cdn_frontdoor_batch_rule_set` — — same ARM object and importer ID (Microsoft.Cdn/profiles/{profile}/ruleSets/{name}) as azurerm_cdn_frontdoor_rule_set; batch mode is a provider-side write strategy that the ruleSets List cannot distinguish, so it would duplicate every rule set.
* `azurerm_cdn_frontdoor_custom_domain_association` — — association glue with no listable ARM counterpart: armcdn v1.1.1 has no Associations client/List; the association is stored as the customDomains references on routes.
* `azurerm_network_manager_deployment` — — side-effect-only deployment commit: ManagerCommitsClient in armnetwork exposes only BeginPost (no List/Get), and its importer ID is the synthetic {managerId}/commit|{location}|{scopeAccess}.
* `azurerm_network_manager_ipam_pool` — — no IPAM pool client/List exists in the track-2 armnetwork SDKs available (v1.1.0, v2.2.1, v4.3.0 all checked).
* `azurerm_network_manager_ipam_pool_static_cidr` — — child of the IPAM pool above; likewise no static-CIDR client/List in armnetwork v1/v2/v4.
* `azurerm_network_manager_management_group_connection` — — its only List (ManagementGroupNetworkManagerConnectionsClient) is keyed by a management group ID — a tenant-scope parent that cannot be enumerated from this subscription-scoped generator; no subscription/RG-scope List exists.
* `azurerm_network_manager_routing_configuration` — — no routing configuration client/List exists in armnetwork v1/v2/v4.
* `azurerm_network_manager_routing_rule` — _collection — child of the routing configuration above; no client/List in armnetwork v1/v2/v4.
* `azurerm_network_manager_routing_rule_collection` — — child of the routing configuration above; no client/List in armnetwork v1/v2/v4.
* `azurerm_network_manager_verifier_workspace` — _reachability_analysis_intent — child of the verifier workspace above; no client/List in armnetwork v1/v2/v4.
* `azurerm_network_manager_verifier_workspace_reachability_analysis_intent` — — child of the verifier workspace above; no client/List in armnetwork v1/v2/v4.
* `azurerm_network_security_perimeter` — — no track-2 SDK surface available: the armnetworksecurityperimeter module is not published in the module proxy (go get reports the package does not exist) and NSP clients are absent from armnetwork v1/v2/v4.
* `azurerm_network_security_perimeter_access_rule` — — child of an NSP profile; same missing SDK surface.
* `azurerm_network_security_perimeter_association` — — association child of a perimeter; same missing SDK surface.
* `azurerm_network_security_perimeter_profile` — — child of the perimeter above; same missing SDK surface.
* `azurerm_virtual_network_dns_servers` — — PUT-only sub-resource of the VNet (the dnsServers array on the virtual network object itself); no independent ARM object or List operation in armnetwork.

**Compute and hybrid**

* `azurerm_arc_kubernetes_provisioned_cluster` — importer ID is the same Microsoft.Kubernetes/connectedClusters ARM ID as azurerm_arc_kubernetes_cluster; no separate List surface, emitting both would duplicate every connected cluster.
* `azurerm_batch_job` — data-plane-only: armbatch v1.2.1 has no Jobs client (only Account/Application/Pool/Certificate); jobs live in the Batch data plane.
* `azurerm_managed_disk_sas_token` — side-effect-only resource; importer ID is the managed disk's own ARM ID, no ARM object and no List surface (mints a SAS token on create).
* `azurerm_managed_lustre_file_system` — no track-2 List surface: armstoragecache v1.0.0 (only track-2 StorageCache module) has no AmlFilesystems client, only Caches/StorageTargets.
* `azurerm_managed_redis_access_policy_assignment` — armredisenterprise v1.2.0 exposes clusters and databases only; no access-policy-assignment List/Get surface.
* `azurerm_managed_redis_geo_replication` — importer ID is the redisEnterprise cluster's own ARM ID (same object emitted as azurerm_managed_redis); geo-replication linking is an action, not an independent listable object.
* `azurerm_virtual_desktop_host_pool_registration_info` — singleton registrationInfo/default child; armdesktopvirtualization exposes only Get/Update, no List.
* `azurerm_virtual_desktop_scaling_plan_host_pool_association` — composite "{scalingPlanId}|{hostPoolId}" association ID; no ARM object, no List surface.
* `azurerm_virtual_desktop_workspace_application_group_association` — composite "{workspaceId}|{applicationGroupId}" association ID; no ARM object, no List surface.
* `azurerm_virtual_machine` — _packet_capture — importer ID is the Microsoft.Network networkWatchers/packetCaptures ARM ID, the same object already emitted as azurerm_network_packet_capture by the network_watcher service; no distinct List surface.
* `azurerm_virtual_machine_gallery_application_assignment` — composite "{vmId}|{galleryApplicationVersionId}" association ID; no ARM object, no List surface.
* `azurerm_virtual_machine_packet_capture` — importer ID is the Microsoft.Network networkWatchers/packetCaptures ARM ID, the same object already emitted as azurerm_network_packet_capture by the network_watcher service; no distinct List surface.
* `azurerm_virtual_machine_restore_point` — pinned armcompute v1.0.0 RestorePointsClient exposes only Get/Create/Delete, no List pager (collections are implemented).
* `azurerm_virtual_machine_scale_set_packet_capture` — same networkWatchers/packetCaptures ARM object as azurerm_network_packet_capture; no distinct List surface.

**Data services**

* `azurerm_cosmosdb_mongo_role_definition` — armcosmos v1.0.0 MongoDBResourcesClient has no role-definition operations; definitions live behind the Cosmos MongoDB data-plane endpoint — no ARM List.
* `azurerm_cosmosdb_mongo_user_definition` — Same as above: no user-definition operations in armcosmos v1.0.0; data-plane only.
* `azurerm_cosmosdb_sql_dedicated_gateway` — armcosmos v1.0.0 has no ServicesClient; account services (services/SqlDedicatedGateway) is Get-only in track-2, no List.
* `azurerm_data_factory_credential_service_principal` — armdatafactory v1.3.0 has no CredentialsClient — credentials are not exposed by the track-2 SDK, no List.
* `azurerm_data_factory_credential_user_managed_identity` — Same as above (shared credentials collection, no CredentialsClient).
* `azurerm_data_factory_customer_managed_key` — Singleton setting on the factory itself; provider importer ID is the factory ID — no separate ARM resource.
* `azurerm_data_factory_linked_service_sql_managed_instance` — armdatafactory v1.3.0 defines no SqlManagedInstance linked-service type/discriminator, so the generic LinkedServices List cannot produce it.
* `azurerm_data_protection_backup_vault_customer_managed_key` — Singleton setting on the vault itself; provider importer ID is the vault ID — no separate ARM resource.
* `azurerm_kusto_cluster_customer_managed_key` — Singleton setting on the cluster itself; provider importer ID is the cluster ID — armkusto has no separate CMK resource/List.
* `azurerm_mssql_database_vulnerability_assessment_rule_baseline` — armsql DatabaseVulnerabilityAssessmentRuleBaselinesClient is Get-only (per rule + baseline name), no List pager.
* `azurerm_mssql_job_schedule` — Schedule is embedded in the parent job (provider importer ID is the job ID); armsql has no JobSchedulesClient — no separate ARM resource.
* `azurerm_mssql_managed_instance_start_stop_schedule` — armsql v1.2.0 has no StartStopSchedulesClient at all (singleton "default" schedule, Get-only where the API exists).
* `azurerm_mysql_flexible_server_active_directory_administrator` — armmysqlflexibleservers v1.2.0 contains no AdministratorsClient — the flexibleServers administrators API is not shipped in the track-2 Go SDK.
* `azurerm_postgresql_flexible_server_active_directory_administrator` — armpostgresqlflexibleservers v1.1.0 contains no AdministratorsClient — same SDK gap.
* `azurerm_postgresql_flexible_server_backup` — No BackupsClient in armpostgresqlflexibleservers v1.1.0.
* `azurerm_postgresql_flexible_server_virtual_endpoint` — No VirtualEndpointsClient in armpostgresqlflexibleservers v1.1.0; importer ID is also a composite of two server IDs.
* `azurerm_redis_cache_access_policy` — armredis v1.0.0 has no AccessPoliciesClient — the redis accessPolicies API is not shipped in the track-2 Go SDK.
* `azurerm_redis_cache_access_policy_assignment` — armredis v1.0.0 has no AccessPolicyAssignmentsClient — same SDK gap.
* `azurerm_synapse_linked_service` — Data-plane only: managed via the workspace Synapse artifacts endpoint ({workspace}.dev.azuresynapse.net); armsynapse has no ARM linked-services client.
* `azurerm_synapse_role_assignment` — Data-plane only: Synapse RBAC via the workspace data-plane endpoint; armsynapse has no ARM role-assignment client, and the provider importer ID is a composite workspace-ID\|role-assignment-ID.
* `azurerm_synapse_sql_pool_vulnerability_assessment_baseline` — armsynapse SQLPoolVulnerabilityAssessmentRuleBaselinesClient is Get-only per rule/baseline name, no List pager.
* `azurerm_synapse_workspace_aad_admin` — Singleton setting; armsynapse WorkspaceAADAdminsClient is Get-only (fixed name "activeDirectory"), no List pager.
* `azurerm_synapse_workspace_sql_aad_admin` — Singleton setting; armsynapse WorkspaceSQLAADAdminsClient is Get-only (fixed name "activeDirectory"), no List pager.

**Security, identity and IoT**

* `azurerm_cognitive_account_customer_managed_key` — embedded property (properties.encryption) of the account itself; no independent ARM object or List.
* `azurerm_communication_service_email_domain_association` — synthetic composite importer ID (`<communicationServiceId>|<emailDomainId>`, verified in provider docs); no independent ARM object or List.
* `azurerm_email_communication_service` — no EmailServices client/List in the track-2 SDK: armcommunication v1.0.0 covers only CommunicationServices and no armemail module is published for Go.
* `azurerm_email_communication_service_domain` — _sender_username — same as above: no domains/senderUsernames client/List in the track-2 SDK.
* `azurerm_email_communication_service_domain_sender_username` — same as above: no domains/senderUsernames client/List in the track-2 SDK.
* `azurerm_eventgrid_namespace` — no Namespaces client/List in track-2 armeventgrid v1.0.0 (module predates Event Grid namespaces).
* `azurerm_eventgrid_namespace_topic` — no NamespaceTopics client/List in track-2 armeventgrid v1.0.0.
* `azurerm_eventgrid_partner_configuration` — no PartnerConfigurations client/List in track-2 armeventgrid v1.0.0.
* `azurerm_eventgrid_partner_namespace` — no PartnerNamespaces client/List in track-2 armeventgrid v1.0.0.
* `azurerm_eventgrid_partner_registration` — no PartnerRegistrations client/List in track-2 armeventgrid v1.0.0.
* `azurerm_eventhub_namespace_customer_managed_key` — embedded property (properties.encryption) of the namespace itself; no independent ARM object or List.
* `azurerm_iotcentral_application_network_rule_set` — embedded property (properties.networkRuleSets) of the application itself; no independent ARM object or List.
* `azurerm_iotcentral_organization` — organizations live behind the IoT Central data plane (per-application endpoint); no ARM List in track-2 armiotcentral.
* `azurerm_iothub_dps` — no Device Provisioning Service (Microsoft.Devices/provisioningServices) client or List in track-2 armiothub v1.3.0 (verified: module contains only IotHubs, Certificates, Operations, PrivateEndpointConnections, PrivateLinkResources clients).
* `azurerm_iothub_dps_certificate` — same as above: no provisioningServices/certificates List in armiothub v1.3.0.
* `azurerm_iothub_dps_shared_access_policy` — same as above: no provisioningServices/keys List in armiothub v1.3.0.
* `azurerm_iothub_endpoint_cosmosdb_account` — embedded array in the hub object (properties.routing.endpoints.cosmosDBSqlContainers); no separate ARM List.
* `azurerm_iothub_endpoint_eventhub` — embedded array in the hub object (properties.routing.endpoints.eventHubs); no separate ARM List.
* `azurerm_iothub_endpoint_servicebus_queue` — embedded array in the hub object (properties.routing.endpoints.serviceBusQueues); no separate ARM List.
* `azurerm_iothub_endpoint_servicebus_topic` — embedded array in the hub object (properties.routing.endpoints.serviceBusTopics); no separate ARM List.
* `azurerm_iothub_endpoint_storage_container` — embedded array in the hub object (properties.routing.endpoints.storageContainers); no separate ARM List.
* `azurerm_iothub_enrichment` — embedded array in the hub object (properties.routing.enrichments); no separate ARM List.
* `azurerm_iothub_fallback_route` — embedded property of the hub object (properties.routing.fallbackRoute); no separate ARM List.
* `azurerm_iothub_file_upload` — embedded property of the hub object (properties.storageEndpoints / messagingEndpoints fileUpload notifications); no separate ARM List.
* `azurerm_iothub_route` — embedded array in the hub object (properties.routing.routes); no separate ARM List.
* `azurerm_key_vault_access_policy` — embedded property (properties.accessPolicies) of the vault with a synthetic composite importer ID; no independent ARM object/List.
* `azurerm_key_vault_certificate` — _contacts — data-plane resource; enumeration requires per-vault data-plane clients, outside the ARM generator pattern.
* `azurerm_key_vault_certificate_contacts` — data-plane resource; enumeration requires per-vault data-plane clients, outside the ARM generator pattern.
* `azurerm_key_vault_certificate_issuer` — data-plane resource; enumeration requires per-vault data-plane clients, outside the ARM generator pattern.
* `azurerm_key_vault_key` — data-plane resource; enumeration requires per-vault data-plane clients, outside the ARM generator pattern. (Provider importer ID is the data-plane key URL, verified in provider docs.).
* `azurerm_key_vault_managed_hardware_security_module_key` — _rotation_policy — data-plane resource; enumeration requires per-vault data-plane clients, outside the ARM generator pattern.
* `azurerm_key_vault_managed_hardware_security_module_key_rotation_policy` — data-plane resource; enumeration requires per-vault data-plane clients, outside the ARM generator pattern.
* `azurerm_key_vault_managed_hardware_security_module_role_assignment` — data-plane resource; enumeration requires per-vault data-plane clients, outside the ARM generator pattern. (Provider importer ID is the HSM data-plane URL `https://<name>.managedhsm.azure.net///RoleAssignment/<id>`, verified in provider docs.).
* `azurerm_key_vault_managed_hardware_security_module_role_definition` — data-plane resource; enumeration requires per-vault data-plane clients, outside the ARM generator pattern.
* `azurerm_key_vault_managed_storage_account` — _sas_token_definition — data-plane resource; enumeration requires per-vault data-plane clients, outside the ARM generator pattern.
* `azurerm_key_vault_managed_storage_account_sas_token_definition` — data-plane resource; enumeration requires per-vault data-plane clients, outside the ARM generator pattern.
* `azurerm_key_vault_secret` — data-plane resource; enumeration requires per-vault data-plane clients, outside the ARM generator pattern.
* `azurerm_security_center_server_vulnerability_assessment_virtual_machine` — armsecurity exposes it only per extended resource (per virtual machine: ListByExtendedResource/Get); no subscription- or resource-group-level List.
* `azurerm_security_center_storage_defender` — armsecurity exposes DefenderForStorage settings only scoped to a single resource ID (per storage account); no subscription- or resource-group-level List.
* `azurerm_sentinel_alert_rule_anomaly_built_in` — a securityMLAnalyticsSettings resource (importer ID verified in provider docs); no client/List for that collection in track-2 armsecurityinsights v1.2.0.
* `azurerm_sentinel_alert_rule_anomaly_duplicate` — same securityMLAnalyticsSettings collection as above; no client/List in track-2 armsecurityinsights v1.2.0.
* `azurerm_sentinel_metadata` — no Metadata client/List in track-2 armsecurityinsights v1.2.0 (importer ID is the ARM `.../Microsoft.SecurityInsights/metadata/<name>` format, verified, but the SDK exposes no way to enumerate it).
* `azurerm_servicebus_namespace_customer_managed_key` — embedded property (properties.encryption) of the namespace itself; no independent ARM object or List.

**Storage, protection and monitoring**

* `azurerm_chaos_studio_capability` — armchaos lists capabilities only per target (parent resource coordinates); no subscription/RG List.
* `azurerm_chaos_studio_target` — armchaos lists targets only per parent resource; no subscription/RG List; provider ID is scoped under an arbitrary resource.
* `azurerm_log_analytics_cluster_customer_managed_key` — imports by the cluster ID itself; singleton association on the cluster object.
* `azurerm_log_analytics_query_pack` — _query — armoperationalinsights v1.2.0 has no query packs client/List.
* `azurerm_log_analytics_query_pack_query` — armoperationalinsights v1.2.0 has no query packs client/List.
* `azurerm_log_analytics_solution` — Microsoft.OperationsManagement resource; armoperationalinsights v1.2.0 has no Solutions client.
* `azurerm_log_analytics_workspace_table` — _microsoft — armoperationalinsights Tables list returns legacy table objects with no plan/schema marker distinguishing built-in Microsoft tables.
* `azurerm_log_analytics_workspace_table_microsoft` — armoperationalinsights Tables list returns legacy table objects with no plan/schema marker distinguishing built-in Microsoft tables.
* `azurerm_monitor_aad_diagnostic_setting` — tenant-level Microsoft.AADIAM settings; no client in armmonitor/armalertsmanagement.
* `azurerm_monitor_alert_prometheus_rule_group` — no PrometheusRuleGroups client/List in track-2 armmonitor v0.13.0 or armalertsmanagement.
* `azurerm_monitor_diagnostic_setting` — armmonitor lists diagnostic settings only per target resource URI; provider ID is composite "{targetResourceId}|{name}"; no subscription/RG List.
* `azurerm_monitor_smart_detector_alert_rule` — no SmartDetectorAlertRules client in armalertsmanagement (checked v0.4–v0.11) or armmonitor.
* `azurerm_monitor_workspace` — Microsoft.Monitor/accounts has no client/List in armmonitor v0.13.0.
* `azurerm_netapp_account_encryption` — imports by the NetApp account ID itself; singleton property of the account object.
* `azurerm_netapp_volume_bucket` — _with_server — armnetapp v1.0.0 (latest) has no Buckets client/List.
* `azurerm_netapp_volume_bucket_with_server` — armnetapp v1.0.0 (latest) has no Buckets client/List.
* `azurerm_netapp_volume_quota_rule` — armnetapp v1.0.0 (latest) has no VolumeQuotaRules client/List.
* `azurerm_qumulo_file_system` — no track-2 SDK module published for Qumulo.Storage (no armqumulo package exists).
* `azurerm_site_recovery_hyperv_network_mapping` — same ARM replicationNetworkMappings object imported as azurerm_site_recovery_network_mapping; no distinct type/List in armrecoveryservicessiterecovery.
* `azurerm_site_recovery_hyperv_replication_policy_association` — association glue between a container mapping and a policy; no independent ARM resource/List.
* `azurerm_site_recovery_services_vault_hyperv_site` — Hyper-V sites are the same ARM replicationFabrics objects imported as azurerm_site_recovery_fabric; no distinct type/List.
* `azurerm_site_recovery_vmware_replication_policy_association` — association glue between a container mapping and a policy; no independent ARM resource/List.
* `azurerm_storage_account_customer_managed_key` — singleton property of the storage account object (Get only, no separate ARM ID/List).
* `azurerm_storage_account_network_rules` — singleton property of the storage account object (Get only, no separate ARM ID/List).
* `azurerm_storage_account_queue_properties` — singleton properties of the queue service object (Get only, no separate ARM ID/List).
* `azurerm_storage_account_static_website` — singleton property of the blob service object (Get only, no separate ARM ID/List).
* `azurerm_storage_account_table_properties` — singleton properties of the table service object (Get only, no separate ARM ID/List).
* `azurerm_storage_data_lake_gen2_filesystem` — data-plane object (imports by https://<account>.dfs.core.windows.net/<filesystem>); no ARM List in armstorage.
* `azurerm_storage_data_lake_gen2_path` — data-plane object; no ARM List in armstorage.
* `azurerm_storage_mover_source_endpoint` — one ARM endpoints collection in armstoragemover with no source/target discriminator; provider type differs only by job-definition usage.
* `azurerm_storage_mover_target_endpoint` — one ARM endpoints collection in armstoragemover with no source/target discriminator; provider type differs only by job-definition usage.
* `azurerm_storage_object_replication` — provider importer requires a composite source;destination policy ID pair; a single enumerated policy cannot form it.
* `azurerm_storage_share_directory` — data-plane object (imports by https://<account>.file.core.windows.net/<share>/<directory>); no ARM List in armstorage.
* `azurerm_storage_share_file` — data-plane object; no ARM List in armstorage.
* `azurerm_storage_table_entity` — data-plane object; no ARM List in armstorage.

**Containers and messaging**

* `azurerm_bot_channels_registration` — legacy kind `registration` bot: the track-2 armbotservice `Kind` enum (azurebot, bot, designer, function, sdk) has no registration kind and BotsClient has no separate List for registrations, so registrations cannot be distinguished from Azure bots.
* `azurerm_bot_web_app` — legacy kind `webapp` bot: same — no `webapp` kind in the track-2 armbotservice Kind enum and no separate List surface in armbotservice.
* `azurerm_container_registry_task_schedule_run_now` — action artefact: scheduling a run-now is a POST action on a task, not an ARM resource; no List in armcontainerregistry.
* `azurerm_container_registry_token_password` — credential artefact: the token password is generated on demand via the tokens `generateCredentials` action; there is no List surface for it in armcontainerregistry.
* `azurerm_databricks_workspace_root_dbfs_customer_managed_key` — get-only singleton: the root DBFS CMK is an encryption configuration block on the workspace itself (Terraform ID = workspace ID); armdatabricks WorkspacesClient exposes only Get on the workspace, no List.
* `azurerm_signalr_service_network_acl` — get-only singleton: the network ACL is a property block on the SignalR service (Terraform ID = service ID), returned with defaults for every service; armsignalr has no NetworkACL client/List.
* `azurerm_stream_analytics_job_schedule` — side-effect-only: creating it starts the job via the streamingJobs start action under a synthetic `.../schedule/default` ID; armstreamanalytics has no Schedule client with a List/Get surface.
* `azurerm_video_indexer_account` — no track-2 SDK: there is no `armvideoindexer` module in the Azure SDK for Go for Microsoft.VideoIndexer, so there is no List surface to enumerate from.
* `azurerm_web_pubsub_network_acl` — get-only singleton: the network ACL is a property block on the Web PubSub service (Terraform ID = service ID), returned with defaults for every service; armwebpubsub has no NetworkACL client/List.

**Governance and management**

* `azurerm_aadb2c_directory` — no track-2 SDK module exists for Microsoft.AzureActiveDirectory (probed resourcemanager/azuread, b2c and azureactivedirectory module paths on the Go module proxy: all 404); the directory is a tenant-level object.
* `azurerm_active_directory_domain_service_trust` — forest trusts are embedded in the domain service's forestTrusts property; armdomainservices v1.2.0 has no Trusts client or list call.
* `azurerm_advanced_threat_protection` — armsecurity AdvancedThreatProtectionClient exposes only Create/Get against a caller-supplied target resource ID; no List exists and the settings live under arbitrary individual resources, so enumeration is infeasible generically.
* `azurerm_billing_account_cost_management_export` — its scope is a billing account (/providers/Microsoft.Billing/billingAccounts/<id>), which cannot be derived from the subscription terraformer operates in.
* `azurerm_cost_anomaly_alert` — backed by the Microsoft.CostManagement/scheduledActions API (InsightAlert kind); armcostmanagement v1.1.1 has no ScheduledActions client.
* `azurerm_cost_management_scheduled_action` — same scheduledActions API; no ScheduledActions client in armcostmanagement v1.1.1.
* `azurerm_management_group_policy_exemption` — the track-2 armpolicy module (v1.0.0, latest) has no PolicyExemptions client.
* `azurerm_resource_group_policy_exemption` — same: no PolicyExemptions client in armpolicy v1.0.0.
* `azurerm_resource_management_private_link` — _association — no PrivateLinkAssociations client exists in track-2 armauthorization v1.0.0 or armresources v1.2.0.
* `azurerm_resource_management_private_link_association` — no PrivateLinkAssociations client exists in track-2 armauthorization v1.0.0 or armresources v1.2.0.
* `azurerm_resource_policy_assignment` — AssignmentsClient.NewListForResourcePager requires a specific target resource (group, provider, type, name); resource-scope assignments cannot be enumerated generically.
* `azurerm_resource_policy_exemption` — same missing client, and additionally resource-scope enumeration is infeasible generically.
* `azurerm_resource_policy_remediation` — RemediationsClient.NewListForResourcePager likewise requires a specific target resource ID; cannot be enumerated generically.
* `azurerm_resource_provider_feature_registration` — side-effect-only subscription state toggle, and armresources v1.2.0 has no Features client to list registrations.
* `azurerm_resource_provider_registration` — side-effect-only state toggle (provider register/unregister), not an enumerable resource object.
* `azurerm_subscription` — singleton: it is the target subscription terraformer itself operates in, and its importer ID is a Microsoft.Subscription alias ID (/providers/Microsoft.Subscription/aliases/<name>), not the subscription's ARM ID.
* `azurerm_subscription_policy_exemption` — same: no PolicyExemptions client in armpolicy v1.0.0.

**Automation, ML and partner services**

* `azurerm_automation_powershell72_module` — no List in track-2 SDK: armautomation v1.0.0 has no PowerShell72Modules client for the provider's powerShell72Modules segment.
* `azurerm_automation_variable_bool` — discriminator-indistinguishable: all five variable types share ARM variables/{name}; armautomation VariableProperties exposes only Value *string with no type discriminator.
* `azurerm_automation_variable_datetime` — discriminator-indistinguishable: same ARM variables/{name} resource, Value *string only, no type discriminator.
* `azurerm_automation_variable_int` — discriminator-indistinguishable: same ARM variables/{name} resource, Value *string only, no type discriminator.
* `azurerm_automation_variable_object` — discriminator-indistinguishable: same ARM variables/{name} resource, Value *string only, no type discriminator.
* `azurerm_automation_variable_string` — discriminator-indistinguishable: same ARM variables/{name} resource, Value *string only, no type discriminator.
* `azurerm_machine_learning_datastore_blobstorage` — no List in track-2 SDK: armmachinelearning v1.0.0 has no Datastores client.
* `azurerm_machine_learning_datastore_datalake_gen2` — no List in track-2 SDK: armmachinelearning v1.0.0 has no Datastores client.
* `azurerm_machine_learning_datastore_fileshare` — no List in track-2 SDK: armmachinelearning v1.0.0 has no Datastores client.
* `azurerm_machine_learning_workspace_network_outbound_rule_fqdn` — no List in track-2 SDK: armmachinelearning v1.0.0 has no outbound-rules/networking client.
* `azurerm_machine_learning_workspace_network_outbound_rule_private_endpoint` — no List in track-2 SDK: armmachinelearning v1.0.0 has no outbound-rules/networking client.
* `azurerm_machine_learning_workspace_network_outbound_rule_service_tag` — no List in track-2 SDK: armmachinelearning v1.0.0 has no outbound-rules/networking client.
* `azurerm_nginx_api_key` — no List in track-2 SDK: armnginx v1.0.1 has no API keys client (only Deployments/Certificates/Configurations).
* `azurerm_oracle_resource_anchor` — no List in track-2 SDK: armoracledatabase v1.1.0 has no ResourceAnchors client.
* `azurerm_palo_alto_local_rulestack_outbound_trust_certificate_association` — embedded/singleton child: armpanngfw v1.1.0 has no association client/List; it is the local rulestack's outboundTrustCertificate property and the provider ID is the rulestack ID.
* `azurerm_palo_alto_local_rulestack_outbound_untrust_certificate_association` — embedded/singleton child: armpanngfw v1.1.0 has no association client/List; it is the local rulestack's outboundUnTrustCertificate property and the provider ID is the rulestack ID.
* `azurerm_palo_alto_virtual_network_appliance` — discriminator-indistinguishable: shares Microsoft.Network/networkVirtualAppliances with the generic network virtual appliance type; armnetwork exposes no vendor discriminator for Palo Alto appliances.

## Notes

### Virtual networks and subnets

Terraformer will import `azurerm_virtual_network` config with inlined subnet information swipped, in order to avoid any potential circular dependencies. To import the subnet information, please also import `azurerm_subnet`.
