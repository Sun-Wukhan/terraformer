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

*   `addresses`
    * `google_compute_address`
*   `autoscalers`
    * `google_compute_autoscaler`
*   `backendBuckets`
    * `google_compute_backend_bucket`
*   `backendServices`
    * `google_compute_backend_service`
*   `bigQuery`
    * `google_bigquery_dataset`
    * `google_bigquery_table`
*   `cloudFunctions`
    * `google_cloudfunctions_function`
    * `google_cloudfunctions2_function`
*   `cloudbuild`
    * `google_cloudbuild_trigger`
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
*   `dataProc`
    * `google_dataproc_cluster`
*   `disks`
    * `google_compute_disk`
*   `dns`
    * `google_dns_managed_zone`
    * `google_dns_record_set`
*   `externalVpnGateways`
    * `google_compute_external_vpn_gateway`
*   `firewall`
    * `google_compute_firewall`
*   `forwardingRules`
    * `google_compute_forwarding_rule`
*   `gcs`
    * `google_storage_bucket`
    * `google_storage_bucket_acl`
    * `google_storage_bucket_iam_binding`
    * `google_storage_bucket_iam_member`
    * `google_storage_bucket_iam_policy`
    * `google_storage_default_object_acl`
    * `google_storage_notification`
*   `gke`
    * `google_container_cluster`
    * `google_container_node_pool`
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
*   `interconnectAttachments`
    * `google_compute_interconnect_attachment`
*   `kms`
    * `google_kms_crypto_key`
    * `google_kms_key_ring`
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
*   `memoryStore`
    * `google_redis_instance`
*   `monitoring`
    * `google_monitoring_alert_policy`
    * `google_monitoring_group`
    * `google_monitoring_notification_channel`
    * `google_monitoring_uptime_check_config`
*   `networks`
    * `google_compute_network`
*   `nodeGroups`
    * `google_compute_node_group`
*   `nodeTemplates`
    * `google_compute_node_template`
*   `packetMirrorings`
    * `google_compute_packet_mirroring`
*   `project`
    * `google_project`
*   `pubsub`
    * `google_pubsub_subscription`
    * `google_pubsub_topic`
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
*   `schedulerJobs`
    * `google_cloud_scheduler_job`
*   `securityPolicies`
    * `google_compute_security_policy`
*   `snapshots`
    * `google_compute_snapshot`
*   `sslCertificates`
    * `google_compute_managed_ssl_certificate`
*   `sslPolicies`
    * `google_compute_ssl_policy`
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
*   `vpnTunnels`
    * `google_compute_vpn_tunnel`

Your `tf` and `tfstate` files are written by default to
`generated/gcp/zone/service`.
