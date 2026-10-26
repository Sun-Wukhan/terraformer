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
	"strconv"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
)

var lightsailAllowEmptyValues = []string{"tags."}

type LightsailGenerator struct {
	AWSService
}

// InitResources enumerates the Amazon Lightsail account in the configured
// region: buckets with their access keys and resource access grants,
// certificates, container services and their deployment versions, relational
// databases, disks and their attachments, distributions, domains and their
// DNS entries, instances and their public ports, key pairs, load balancers
// with their attachments, certificates and policies, and static IPs and
// their attachments.
//
// The Lightsail list APIs are Get-style calls with optional page tokens
// rather than SDK paginators; the token loops here follow them manually.
func (g *LightsailGenerator) InitResources() error {
	config, e := g.generateConfig()
	if e != nil {
		return e
	}
	client := lightsail.NewFromConfig(config)

	for _, add := range []func(*lightsail.Client) error{
		g.addBuckets,
		g.addCertificates,
		g.addContainerServices,
		g.addDatabases,
		g.addDisks,
		g.addDistributions,
		g.addDomains,
		g.addInstances,
		g.addKeyPairs,
		g.addLoadBalancers,
		g.addStaticIps,
	} {
		if err := add(client); err != nil {
			return err
		}
	}
	return nil
}

func (g *LightsailGenerator) addBuckets(client *lightsail.Client) error {
	input := &lightsail.GetBucketsInput{
		IncludeConnectedResources: aws.Bool(true),
	}
	for {
		page, err := client.GetBuckets(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, bucket := range page.Buckets {
			bucketName := *bucket.Name
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				bucketName,
				bucketName,
				"aws_lightsail_bucket",
				"aws",
				lightsailAllowEmptyValues,
			))
			if err := g.addBucketAccessKeys(client, bucketName); err != nil {
				return err
			}
			for _, access := range bucket.ResourcesReceivingAccess {
				// Provider ID format: bucket name, resource name.
				g.Resources = append(g.Resources, terraformutils.NewResource(
					bucketName+","+*access.Name,
					*access.Name,
					"aws_lightsail_bucket_resource_access",
					"aws",
					map[string]string{"bucket_name": bucketName},
					lightsailAllowEmptyValues,
					map[string]interface{}{},
				))
			}
		}
		if page.NextPageToken == nil || *page.NextPageToken == "" {
			return nil
		}
		input.PageToken = page.NextPageToken
	}
}

// Provider ID format: bucket name, access key ID.
func (g *LightsailGenerator) addBucketAccessKeys(client *lightsail.Client, bucketName string) error {
	out, err := client.GetBucketAccessKeys(context.TODO(), &lightsail.GetBucketAccessKeysInput{
		BucketName: aws.String(bucketName),
	})
	if err != nil {
		return err
	}
	for _, key := range out.AccessKeys {
		g.Resources = append(g.Resources, terraformutils.NewResource(
			bucketName+","+*key.AccessKeyId,
			*key.AccessKeyId,
			"aws_lightsail_bucket_access_key",
			"aws",
			map[string]string{"bucket_name": bucketName},
			lightsailAllowEmptyValues,
			map[string]interface{}{},
		))
	}
	return nil
}

func (g *LightsailGenerator) addCertificates(client *lightsail.Client) error {
	input := &lightsail.GetCertificatesInput{}
	for {
		page, err := client.GetCertificates(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, certificate := range page.Certificates {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*certificate.CertificateName,
				*certificate.CertificateName,
				"aws_lightsail_certificate",
				"aws",
				lightsailAllowEmptyValues,
			))
		}
		if page.NextPageToken == nil || *page.NextPageToken == "" {
			return nil
		}
		input.PageToken = page.NextPageToken
	}
}

func (g *LightsailGenerator) addContainerServices(client *lightsail.Client) error {
	out, err := client.GetContainerServices(context.TODO(), &lightsail.GetContainerServicesInput{})
	if err != nil {
		return err
	}
	for _, service := range out.ContainerServices {
		serviceName := *service.ContainerServiceName
		g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
			serviceName,
			serviceName,
			"aws_lightsail_container_service",
			"aws",
			lightsailAllowEmptyValues,
		))
		if err := g.addContainerServiceDeploymentVersions(client, serviceName); err != nil {
			return err
		}
	}
	return nil
}

// Provider ID format: service name, deployment version (slash separated).
func (g *LightsailGenerator) addContainerServiceDeploymentVersions(client *lightsail.Client, serviceName string) error {
	out, err := client.GetContainerServiceDeployments(context.TODO(), &lightsail.GetContainerServiceDeploymentsInput{
		ServiceName: aws.String(serviceName),
	})
	if err != nil {
		return err
	}
	for _, deployment := range out.Deployments {
		version := strconv.FormatInt(int64(*deployment.Version), 10)
		g.Resources = append(g.Resources, terraformutils.NewResource(
			serviceName+"/"+version,
			serviceName+"_"+version,
			"aws_lightsail_container_service_deployment_version",
			"aws",
			map[string]string{"service_name": serviceName},
			lightsailAllowEmptyValues,
			map[string]interface{}{},
		))
	}
	return nil
}

func (g *LightsailGenerator) addDatabases(client *lightsail.Client) error {
	input := &lightsail.GetRelationalDatabasesInput{}
	for {
		page, err := client.GetRelationalDatabases(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, database := range page.RelationalDatabases {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*database.Name,
				*database.Name,
				"aws_lightsail_database",
				"aws",
				lightsailAllowEmptyValues,
			))
		}
		if page.NextPageToken == nil || *page.NextPageToken == "" {
			return nil
		}
		input.PageToken = page.NextPageToken
	}
}

func (g *LightsailGenerator) addDisks(client *lightsail.Client) error {
	input := &lightsail.GetDisksInput{}
	for {
		page, err := client.GetDisks(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, disk := range page.Disks {
			diskName := *disk.Name
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				diskName,
				diskName,
				"aws_lightsail_disk",
				"aws",
				lightsailAllowEmptyValues,
			))
			if disk.AttachedTo == nil || *disk.AttachedTo == "" {
				continue
			}
			// Provider ID format: disk name, instance name.
			g.Resources = append(g.Resources, terraformutils.NewResource(
				diskName+","+*disk.AttachedTo,
				diskName,
				"aws_lightsail_disk_attachment",
				"aws",
				map[string]string{
					"disk_name":     diskName,
					"instance_name": *disk.AttachedTo,
				},
				lightsailAllowEmptyValues,
				map[string]interface{}{},
			))
		}
		if page.NextPageToken == nil || *page.NextPageToken == "" {
			return nil
		}
		input.PageToken = page.NextPageToken
	}
}

func (g *LightsailGenerator) addDistributions(client *lightsail.Client) error {
	input := &lightsail.GetDistributionsInput{}
	for {
		page, err := client.GetDistributions(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, distribution := range page.Distributions {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*distribution.Name,
				*distribution.Name,
				"aws_lightsail_distribution",
				"aws",
				lightsailAllowEmptyValues,
			))
		}
		if page.NextPageToken == nil || *page.NextPageToken == "" {
			return nil
		}
		input.PageToken = page.NextPageToken
	}
}

func (g *LightsailGenerator) addDomains(client *lightsail.Client) error {
	input := &lightsail.GetDomainsInput{}
	for {
		page, err := client.GetDomains(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, domain := range page.Domains {
			domainName := *domain.Name
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				domainName,
				domainName,
				"aws_lightsail_domain",
				"aws",
				lightsailAllowEmptyValues,
			))
			if err := g.addDomainEntries(client, domainName); err != nil {
				return err
			}
		}
		if page.NextPageToken == nil || *page.NextPageToken == "" {
			return nil
		}
		input.PageToken = page.NextPageToken
	}
}

// Provider ID format: entry name, domain name, entry type, entry target.
func (g *LightsailGenerator) addDomainEntries(client *lightsail.Client, domainName string) error {
	out, err := client.GetDomain(context.TODO(), &lightsail.GetDomainInput{
		DomainName: aws.String(domainName),
	})
	if err != nil {
		return err
	}
	if out.Domain == nil {
		return nil
	}
	for _, entry := range out.Domain.DomainEntries {
		g.Resources = append(g.Resources, terraformutils.NewResource(
			*entry.Name+","+domainName+","+*entry.Type+","+*entry.Target,
			*entry.Name,
			"aws_lightsail_domain_entry",
			"aws",
			map[string]string{"domain_name": domainName},
			lightsailAllowEmptyValues,
			map[string]interface{}{},
		))
	}
	return nil
}

func (g *LightsailGenerator) addInstances(client *lightsail.Client) error {
	input := &lightsail.GetInstancesInput{}
	for {
		page, err := client.GetInstances(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, instance := range page.Instances {
			instanceName := *instance.Name
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				instanceName,
				instanceName,
				"aws_lightsail_instance",
				"aws",
				lightsailAllowEmptyValues,
			))
			if err := g.addInstancePublicPorts(client, instanceName); err != nil {
				return err
			}
		}
		if page.NextPageToken == nil || *page.NextPageToken == "" {
			return nil
		}
		input.PageToken = page.NextPageToken
	}
}

// The public ports resource is keyed by instance name; it only exists for
// an instance when ports have been configured for it.
func (g *LightsailGenerator) addInstancePublicPorts(client *lightsail.Client, instanceName string) error {
	out, err := client.GetInstancePortStates(context.TODO(), &lightsail.GetInstancePortStatesInput{
		InstanceName: aws.String(instanceName),
	})
	if err != nil {
		return err
	}
	if len(out.PortStates) == 0 {
		return nil
	}
	g.Resources = append(g.Resources, terraformutils.NewResource(
		instanceName,
		instanceName,
		"aws_lightsail_instance_public_ports",
		"aws",
		map[string]string{"instance_name": instanceName},
		lightsailAllowEmptyValues,
		map[string]interface{}{},
	))
	return nil
}

func (g *LightsailGenerator) addKeyPairs(client *lightsail.Client) error {
	input := &lightsail.GetKeyPairsInput{}
	for {
		page, err := client.GetKeyPairs(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, keyPair := range page.KeyPairs {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				*keyPair.Name,
				*keyPair.Name,
				"aws_lightsail_key_pair",
				"aws",
				lightsailAllowEmptyValues,
			))
		}
		if page.NextPageToken == nil || *page.NextPageToken == "" {
			return nil
		}
		input.PageToken = page.NextPageToken
	}
}

func (g *LightsailGenerator) addLoadBalancers(client *lightsail.Client) error {
	input := &lightsail.GetLoadBalancersInput{}
	for {
		page, err := client.GetLoadBalancers(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, lb := range page.LoadBalancers {
			lbName := *lb.Name
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				lbName,
				lbName,
				"aws_lightsail_lb",
				"aws",
				lightsailAllowEmptyValues,
			))
			for _, health := range lb.InstanceHealthSummary {
				// Provider ID format: load balancer name, instance name.
				g.Resources = append(g.Resources, terraformutils.NewResource(
					lbName+","+*health.InstanceName,
					*health.InstanceName,
					"aws_lightsail_lb_attachment",
					"aws",
					map[string]string{"lb_name": lbName},
					lightsailAllowEmptyValues,
					map[string]interface{}{},
				))
			}
			// Both LB policies are per-load-balancer singletons keyed by
			// load balancer name.
			g.Resources = append(g.Resources, terraformutils.NewResource(
				lbName,
				lbName,
				"aws_lightsail_lb_https_redirection_policy",
				"aws",
				map[string]string{"lb_name": lbName},
				lightsailAllowEmptyValues,
				map[string]interface{}{},
			))
			g.Resources = append(g.Resources, terraformutils.NewResource(
				lbName,
				lbName,
				"aws_lightsail_lb_stickiness_policy",
				"aws",
				map[string]string{"lb_name": lbName},
				lightsailAllowEmptyValues,
				map[string]interface{}{},
			))
			if err := g.addLoadBalancerCertificates(client, lbName); err != nil {
				return err
			}
		}
		if page.NextPageToken == nil || *page.NextPageToken == "" {
			return nil
		}
		input.PageToken = page.NextPageToken
	}
}

// Provider ID format: load balancer name, certificate name. A certificate
// that is attached to the load balancer also yields the attachment
// resource, which uses the same ID format.
func (g *LightsailGenerator) addLoadBalancerCertificates(client *lightsail.Client, lbName string) error {
	out, err := client.GetLoadBalancerTlsCertificates(context.TODO(), &lightsail.GetLoadBalancerTlsCertificatesInput{
		LoadBalancerName: aws.String(lbName),
	})
	if err != nil {
		return err
	}
	for _, certificate := range out.TlsCertificates {
		certificateName := *certificate.Name
		g.Resources = append(g.Resources, terraformutils.NewResource(
			lbName+","+certificateName,
			certificateName,
			"aws_lightsail_lb_certificate",
			"aws",
			map[string]string{"lb_name": lbName},
			lightsailAllowEmptyValues,
			map[string]interface{}{},
		))
		if certificate.IsAttached == nil || !*certificate.IsAttached {
			continue
		}
		g.Resources = append(g.Resources, terraformutils.NewResource(
			lbName+","+certificateName,
			certificateName,
			"aws_lightsail_lb_certificate_attachment",
			"aws",
			map[string]string{"lb_name": lbName},
			lightsailAllowEmptyValues,
			map[string]interface{}{},
		))
	}
	return nil
}

func (g *LightsailGenerator) addStaticIps(client *lightsail.Client) error {
	input := &lightsail.GetStaticIpsInput{}
	for {
		page, err := client.GetStaticIps(context.TODO(), input)
		if err != nil {
			return err
		}
		for _, staticIP := range page.StaticIps {
			staticIPName := *staticIP.Name
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				staticIPName,
				staticIPName,
				"aws_lightsail_static_ip",
				"aws",
				lightsailAllowEmptyValues,
			))
			if staticIP.AttachedTo == nil || *staticIP.AttachedTo == "" {
				continue
			}
			// The attachment resource is keyed by static IP name alone.
			g.Resources = append(g.Resources, terraformutils.NewResource(
				staticIPName,
				staticIPName,
				"aws_lightsail_static_ip_attachment",
				"aws",
				map[string]string{"static_ip_name": staticIPName},
				lightsailAllowEmptyValues,
				map[string]interface{}{},
			))
		}
		if page.NextPageToken == nil || *page.NextPageToken == "" {
			return nil
		}
		input.PageToken = page.NextPageToken
	}
}
