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
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
)

// APIManagementGenerator imports API Management services and their child
// resources via the track-2 armapimanagement SDK.
//
// Skipped azurerm types in this family (no usable List surface in
// armapimanagement v1.1.1):
//   - azurerm_api_management_api_operation_policy, azurerm_api_management_api_policy,
//     azurerm_api_management_policy, azurerm_api_management_product_policy:
//     singleton policy documents; the SDK policy clients expose only
//     Get/CreateOrUpdate/Delete, no List operation.
//   - azurerm_api_management_custom_domain: no SDK client; custom domains are
//     hostnameConfiguration properties embedded in the service resource.
//   - azurerm_api_management_policy_fragment and all azurerm_api_management_workspace*
//     types: armapimanagement v1.1.1 has no PolicyFragment or Workspace clients.
//   - azurerm_api_management_standalone_gateway: the SDK GatewayClient only
//     lists gateways under a service (ListByService); there is no
//     resource-group/subscription List for Microsoft.ApiManagement/gateways.
type APIManagementGenerator struct {
	AzureService
}

type apimServiceRef struct {
	resourceGroup string
	name          string
	id            string
}

// drainPager collects every page of a track-2 pager and maps each item.
func drainPager[R any, I any](ctx context.Context, pager *runtime.Pager[R], values func(R) []*I, mapItem func(*I) (terraformutils.Resource, bool)) ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Println(err)
			return resources, err
		}
		for _, item := range values(page) {
			if res, ok := mapItem(item); ok {
				resources = append(resources, res)
			}
		}
	}
	return resources, nil
}

func (g *APIManagementGenerator) simple(id, name, resourceType string) (terraformutils.Resource, bool) {
	if id == "" || name == "" {
		return terraformutils.Resource{}, false
	}
	return terraformutils.NewSimpleResource(id, name, resourceType, g.ProviderName, []string{}), true
}

func (g *APIManagementGenerator) listServices(ctx context.Context) ([]apimServiceRef, []terraformutils.Resource, error) {
	client, err := armapimanagement.NewServiceClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, nil, err
	}
	var refs []apimServiceRef
	mapItem := func(item *armapimanagement.ServiceResource) (terraformutils.Resource, bool) {
		if item.ID == nil || item.Name == nil {
			return terraformutils.Resource{}, false
		}
		parsed, err := ParseAzureResourceID(*item.ID)
		if err != nil {
			return terraformutils.Resource{}, false
		}
		refs = append(refs, apimServiceRef{resourceGroup: parsed.ResourceGroup, name: *item.Name, id: *item.ID})
		return g.simple(*item.ID, *item.Name, "azurerm_api_management")
	}
	var resources []terraformutils.Resource
	if rg := g.getResourceGroup(); rg != "" {
		resources, err = drainPager(ctx, client.NewListByResourceGroupPager(rg, nil),
			func(page armapimanagement.ServiceClientListByResourceGroupResponse) []*armapimanagement.ServiceResource {
				return page.Value
			}, mapItem)
	} else {
		resources, err = drainPager(ctx, client.NewListPager(nil),
			func(page armapimanagement.ServiceClientListResponse) []*armapimanagement.ServiceResource {
				return page.Value
			}, mapItem)
	}
	return refs, resources, err
}

// perService runs fn for every APIM service in scope and collects resources.
func (g *APIManagementGenerator) perService(ctx context.Context, refs []apimServiceRef, fn func(ctx context.Context, svc apimServiceRef) ([]terraformutils.Resource, error)) ([]terraformutils.Resource, error) {
	var resources []terraformutils.Resource
	for _, svc := range refs {
		res, err := fn(ctx, svc)
		if err != nil {
			return resources, err
		}
		resources = append(resources, res...)
	}
	return resources, nil
}

func idSegmentAfter(id, marker string) string {
	idx := strings.Index(id, marker)
	if idx < 0 {
		return id
	}
	rest := id[idx+len(marker):]
	if slash := strings.Index(rest, "/"); slash >= 0 {
		return rest[:slash]
	}
	return rest
}

// listAPIs returns the APIs of a service (also used as parents for API children).
func (g *APIManagementGenerator) listAPIs(ctx context.Context, svc apimServiceRef) ([]*armapimanagement.APIContract, error) {
	client, err := armapimanagement.NewAPIClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	var apis []*armapimanagement.APIContract
	pager := client.NewListByServicePager(svc.resourceGroup, svc.name, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return apis, err
		}
		apis = append(apis, page.Value...)
	}
	return apis, nil
}

func (g *APIManagementGenerator) InitResources() error {
	ctx := context.Background()
	refs, resources, err := g.listServices(ctx)
	if err != nil {
		return err
	}
	g.Resources = append(g.Resources, resources...)

	steps := []func(context.Context, []apimServiceRef) ([]terraformutils.Resource, error){
		g.listAPIsAndChildren,
		g.listVersionSets,
		g.listAuthorizationServers,
		g.listBackends,
		g.listCertificates,
		g.listDiagnostics,
		g.listEmailTemplates,
		g.listGatewaysAndChildren,
		g.listGlobalSchemas,
		g.listGroupsAndUsers,
		g.listIdentityProviders,
		g.listLoggers,
		g.listNamedValues,
		g.listNotificationRecipients,
		g.listOpenIDConnectProviders,
		g.listProductsAndChildren,
		g.listRedisCaches,
		g.listSubscriptions,
		g.listTags,
		g.listUsers,
	}
	for _, step := range steps {
		res, err := step(ctx, refs)
		if err != nil {
			return err
		}
		g.Resources = append(g.Resources, res...)
	}
	return nil
}

func (g *APIManagementGenerator) listAPIsAndChildren(ctx context.Context, refs []apimServiceRef) ([]terraformutils.Resource, error) {
	return g.perService(ctx, refs, func(ctx context.Context, svc apimServiceRef) ([]terraformutils.Resource, error) {
		var resources []terraformutils.Resource
		apis, err := g.listAPIs(ctx, svc)
		if err != nil {
			return resources, err
		}
		diagClient, err := armapimanagement.NewAPIDiagnosticClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
		if err != nil {
			return resources, err
		}
		opClient, err := armapimanagement.NewAPIOperationClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
		if err != nil {
			return resources, err
		}
		releaseClient, err := armapimanagement.NewAPIReleaseClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
		if err != nil {
			return resources, err
		}
		schemaClient, err := armapimanagement.NewAPISchemaClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
		if err != nil {
			return resources, err
		}
		tagClient, err := armapimanagement.NewTagClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
		if err != nil {
			return resources, err
		}
		tagDescClient, err := armapimanagement.NewAPITagDescriptionClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
		if err != nil {
			return resources, err
		}
		for _, api := range apis {
			if api.ID == nil || api.Name == nil {
				continue
			}
			if res, ok := g.simple(*api.ID, *api.Name, "azurerm_api_management_api"); ok {
				resources = append(resources, res)
			}
			apiID := idSegmentAfter(*api.ID, "/apis/")

			diags, err := drainPager(ctx, diagClient.NewListByServicePager(svc.resourceGroup, svc.name, apiID, nil),
				func(p armapimanagement.APIDiagnosticClientListByServiceResponse) []*armapimanagement.DiagnosticContract {
					return p.Value
				},
				func(item *armapimanagement.DiagnosticContract) (terraformutils.Resource, bool) {
					if item.ID == nil || item.Name == nil {
						return terraformutils.Resource{}, false
					}
					return g.simple(*item.ID, *item.Name, "azurerm_api_management_api_diagnostic")
				})
			if err != nil {
				return resources, err
			}
			resources = append(resources, diags...)

			ops, err := drainPager(ctx, opClient.NewListByAPIPager(svc.resourceGroup, svc.name, apiID, nil),
				func(p armapimanagement.APIOperationClientListByAPIResponse) []*armapimanagement.OperationContract {
					return p.Value
				},
				func(item *armapimanagement.OperationContract) (terraformutils.Resource, bool) {
					if item.ID == nil || item.Name == nil {
						return terraformutils.Resource{}, false
					}
					return g.simple(*item.ID, *item.Name, "azurerm_api_management_api_operation")
				})
			if err != nil {
				return resources, err
			}
			resources = append(resources, ops...)
			for _, op := range ops {
				opID := idSegmentAfter(op.InstanceState.ID, "/operations/")
				opTags, err := drainPager(ctx, tagClient.NewListByOperationPager(svc.resourceGroup, svc.name, apiID, opID, nil),
					func(p armapimanagement.TagClientListByOperationResponse) []*armapimanagement.TagContract {
						return p.Value
					},
					func(item *armapimanagement.TagContract) (terraformutils.Resource, bool) {
						if item.ID == nil || item.Name == nil {
							return terraformutils.Resource{}, false
						}
						return g.simple(*item.ID, *item.Name, "azurerm_api_management_api_operation_tag")
					})
				if err != nil {
					return resources, err
				}
				resources = append(resources, opTags...)
			}

			releases, err := drainPager(ctx, releaseClient.NewListByServicePager(svc.resourceGroup, svc.name, apiID, nil),
				func(p armapimanagement.APIReleaseClientListByServiceResponse) []*armapimanagement.APIReleaseContract {
					return p.Value
				},
				func(item *armapimanagement.APIReleaseContract) (terraformutils.Resource, bool) {
					if item.ID == nil || item.Name == nil {
						return terraformutils.Resource{}, false
					}
					return g.simple(*item.ID, *item.Name, "azurerm_api_management_api_release")
				})
			if err != nil {
				return resources, err
			}
			resources = append(resources, releases...)

			schemas, err := drainPager(ctx, schemaClient.NewListByAPIPager(svc.resourceGroup, svc.name, apiID, nil),
				func(p armapimanagement.APISchemaClientListByAPIResponse) []*armapimanagement.SchemaContract {
					return p.Value
				},
				func(item *armapimanagement.SchemaContract) (terraformutils.Resource, bool) {
					if item.ID == nil || item.Name == nil {
						return terraformutils.Resource{}, false
					}
					return g.simple(*item.ID, *item.Name, "azurerm_api_management_api_schema")
				})
			if err != nil {
				return resources, err
			}
			resources = append(resources, schemas...)

			apiTags, err := drainPager(ctx, tagClient.NewListByAPIPager(svc.resourceGroup, svc.name, apiID, nil),
				func(p armapimanagement.TagClientListByAPIResponse) []*armapimanagement.TagContract { return p.Value },
				func(item *armapimanagement.TagContract) (terraformutils.Resource, bool) {
					if item.ID == nil || item.Name == nil {
						return terraformutils.Resource{}, false
					}
					return g.simple(*item.ID, *item.Name, "azurerm_api_management_api_tag")
				})
			if err != nil {
				return resources, err
			}
			resources = append(resources, apiTags...)

			tagDescs, err := drainPager(ctx, tagDescClient.NewListByServicePager(svc.resourceGroup, svc.name, apiID, nil),
				func(p armapimanagement.APITagDescriptionClientListByServiceResponse) []*armapimanagement.TagDescriptionContract {
					return p.Value
				},
				func(item *armapimanagement.TagDescriptionContract) (terraformutils.Resource, bool) {
					if item.ID == nil || item.Name == nil {
						return terraformutils.Resource{}, false
					}
					return g.simple(*item.ID, *item.Name, "azurerm_api_management_api_tag_description")
				})
			if err != nil {
				return resources, err
			}
			resources = append(resources, tagDescs...)
		}
		return resources, nil
	})
}

func (g *APIManagementGenerator) listVersionSets(ctx context.Context, refs []apimServiceRef) ([]terraformutils.Resource, error) {
	client, err := armapimanagement.NewAPIVersionSetClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	return g.perService(ctx, refs, func(ctx context.Context, svc apimServiceRef) ([]terraformutils.Resource, error) {
		return drainPager(ctx, client.NewListByServicePager(svc.resourceGroup, svc.name, nil),
			func(p armapimanagement.APIVersionSetClientListByServiceResponse) []*armapimanagement.APIVersionSetContract {
				return p.Value
			},
			func(item *armapimanagement.APIVersionSetContract) (terraformutils.Resource, bool) {
				if item.ID == nil || item.Name == nil {
					return terraformutils.Resource{}, false
				}
				return g.simple(*item.ID, *item.Name, "azurerm_api_management_api_version_set")
			})
	})
}

func (g *APIManagementGenerator) listAuthorizationServers(ctx context.Context, refs []apimServiceRef) ([]terraformutils.Resource, error) {
	client, err := armapimanagement.NewAuthorizationServerClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	return g.perService(ctx, refs, func(ctx context.Context, svc apimServiceRef) ([]terraformutils.Resource, error) {
		return drainPager(ctx, client.NewListByServicePager(svc.resourceGroup, svc.name, nil),
			func(p armapimanagement.AuthorizationServerClientListByServiceResponse) []*armapimanagement.AuthorizationServerContract {
				return p.Value
			},
			func(item *armapimanagement.AuthorizationServerContract) (terraformutils.Resource, bool) {
				if item.ID == nil || item.Name == nil {
					return terraformutils.Resource{}, false
				}
				return g.simple(*item.ID, *item.Name, "azurerm_api_management_authorization_server")
			})
	})
}

func (g *APIManagementGenerator) listBackends(ctx context.Context, refs []apimServiceRef) ([]terraformutils.Resource, error) {
	client, err := armapimanagement.NewBackendClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	return g.perService(ctx, refs, func(ctx context.Context, svc apimServiceRef) ([]terraformutils.Resource, error) {
		return drainPager(ctx, client.NewListByServicePager(svc.resourceGroup, svc.name, nil),
			func(p armapimanagement.BackendClientListByServiceResponse) []*armapimanagement.BackendContract {
				return p.Value
			},
			func(item *armapimanagement.BackendContract) (terraformutils.Resource, bool) {
				if item.ID == nil || item.Name == nil {
					return terraformutils.Resource{}, false
				}
				return g.simple(*item.ID, *item.Name, "azurerm_api_management_backend")
			})
	})
}

func (g *APIManagementGenerator) listCertificates(ctx context.Context, refs []apimServiceRef) ([]terraformutils.Resource, error) {
	client, err := armapimanagement.NewCertificateClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	return g.perService(ctx, refs, func(ctx context.Context, svc apimServiceRef) ([]terraformutils.Resource, error) {
		return drainPager(ctx, client.NewListByServicePager(svc.resourceGroup, svc.name, nil),
			func(p armapimanagement.CertificateClientListByServiceResponse) []*armapimanagement.CertificateContract {
				return p.Value
			},
			func(item *armapimanagement.CertificateContract) (terraformutils.Resource, bool) {
				if item.ID == nil || item.Name == nil {
					return terraformutils.Resource{}, false
				}
				return g.simple(*item.ID, *item.Name, "azurerm_api_management_certificate")
			})
	})
}

func (g *APIManagementGenerator) listDiagnostics(ctx context.Context, refs []apimServiceRef) ([]terraformutils.Resource, error) {
	client, err := armapimanagement.NewDiagnosticClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	return g.perService(ctx, refs, func(ctx context.Context, svc apimServiceRef) ([]terraformutils.Resource, error) {
		return drainPager(ctx, client.NewListByServicePager(svc.resourceGroup, svc.name, nil),
			func(p armapimanagement.DiagnosticClientListByServiceResponse) []*armapimanagement.DiagnosticContract {
				return p.Value
			},
			func(item *armapimanagement.DiagnosticContract) (terraformutils.Resource, bool) {
				if item.ID == nil || item.Name == nil {
					return terraformutils.Resource{}, false
				}
				return g.simple(*item.ID, *item.Name, "azurerm_api_management_diagnostic")
			})
	})
}

func (g *APIManagementGenerator) listEmailTemplates(ctx context.Context, refs []apimServiceRef) ([]terraformutils.Resource, error) {
	client, err := armapimanagement.NewEmailTemplateClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	return g.perService(ctx, refs, func(ctx context.Context, svc apimServiceRef) ([]terraformutils.Resource, error) {
		return drainPager(ctx, client.NewListByServicePager(svc.resourceGroup, svc.name, nil),
			func(p armapimanagement.EmailTemplateClientListByServiceResponse) []*armapimanagement.EmailTemplateContract {
				return p.Value
			},
			func(item *armapimanagement.EmailTemplateContract) (terraformutils.Resource, bool) {
				if item.ID == nil || item.Name == nil {
					return terraformutils.Resource{}, false
				}
				return g.simple(*item.ID, *item.Name, "azurerm_api_management_email_template")
			})
	})
}

func (g *APIManagementGenerator) listGatewaysAndChildren(ctx context.Context, refs []apimServiceRef) ([]terraformutils.Resource, error) {
	gatewayClient, err := armapimanagement.NewGatewayClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	apiClient, err := armapimanagement.NewGatewayAPIClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	caClient, err := armapimanagement.NewGatewayCertificateAuthorityClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	hnClient, err := armapimanagement.NewGatewayHostnameConfigurationClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	return g.perService(ctx, refs, func(ctx context.Context, svc apimServiceRef) ([]terraformutils.Resource, error) {
		var gateways []*armapimanagement.GatewayContract
		pager := gatewayClient.NewListByServicePager(svc.resourceGroup, svc.name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return nil, err
			}
			gateways = append(gateways, page.Value...)
		}
		var resources []terraformutils.Resource
		for _, gw := range gateways {
			if gw.ID == nil || gw.Name == nil {
				continue
			}
			if res, ok := g.simple(*gw.ID, *gw.Name, "azurerm_api_management_gateway"); ok {
				resources = append(resources, res)
			}
			gwID := *gw.Name
			apis, err := drainPager(ctx, apiClient.NewListByServicePager(svc.resourceGroup, svc.name, gwID, nil),
				func(p armapimanagement.GatewayAPIClientListByServiceResponse) []*armapimanagement.APIContract {
					return p.Value
				},
				func(item *armapimanagement.APIContract) (terraformutils.Resource, bool) {
					if item.ID == nil || item.Name == nil {
						return terraformutils.Resource{}, false
					}
					return g.simple(*item.ID, *item.Name, "azurerm_api_management_gateway_api")
				})
			if err != nil {
				return resources, err
			}
			resources = append(resources, apis...)
			cas, err := drainPager(ctx, caClient.NewListByServicePager(svc.resourceGroup, svc.name, gwID, nil),
				func(p armapimanagement.GatewayCertificateAuthorityClientListByServiceResponse) []*armapimanagement.GatewayCertificateAuthorityContract {
					return p.Value
				},
				func(item *armapimanagement.GatewayCertificateAuthorityContract) (terraformutils.Resource, bool) {
					if item.ID == nil || item.Name == nil {
						return terraformutils.Resource{}, false
					}
					return g.simple(*item.ID, *item.Name, "azurerm_api_management_gateway_certificate_authority")
				})
			if err != nil {
				return resources, err
			}
			resources = append(resources, cas...)
			hns, err := drainPager(ctx, hnClient.NewListByServicePager(svc.resourceGroup, svc.name, gwID, nil),
				func(p armapimanagement.GatewayHostnameConfigurationClientListByServiceResponse) []*armapimanagement.GatewayHostnameConfigurationContract {
					return p.Value
				},
				func(item *armapimanagement.GatewayHostnameConfigurationContract) (terraformutils.Resource, bool) {
					if item.ID == nil || item.Name == nil {
						return terraformutils.Resource{}, false
					}
					return g.simple(*item.ID, *item.Name, "azurerm_api_management_gateway_host_name_configuration")
				})
			if err != nil {
				return resources, err
			}
			resources = append(resources, hns...)
		}
		return resources, nil
	})
}

func (g *APIManagementGenerator) listGlobalSchemas(ctx context.Context, refs []apimServiceRef) ([]terraformutils.Resource, error) {
	client, err := armapimanagement.NewGlobalSchemaClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	return g.perService(ctx, refs, func(ctx context.Context, svc apimServiceRef) ([]terraformutils.Resource, error) {
		return drainPager(ctx, client.NewListByServicePager(svc.resourceGroup, svc.name, nil),
			func(p armapimanagement.GlobalSchemaClientListByServiceResponse) []*armapimanagement.GlobalSchemaContract {
				return p.Value
			},
			func(item *armapimanagement.GlobalSchemaContract) (terraformutils.Resource, bool) {
				if item.ID == nil || item.Name == nil {
					return terraformutils.Resource{}, false
				}
				return g.simple(*item.ID, *item.Name, "azurerm_api_management_global_schema")
			})
	})
}

func (g *APIManagementGenerator) listGroupsAndUsers(ctx context.Context, refs []apimServiceRef) ([]terraformutils.Resource, error) {
	groupClient, err := armapimanagement.NewGroupClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	groupUserClient, err := armapimanagement.NewGroupUserClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	return g.perService(ctx, refs, func(ctx context.Context, svc apimServiceRef) ([]terraformutils.Resource, error) {
		var groups []*armapimanagement.GroupContract
		pager := groupClient.NewListByServicePager(svc.resourceGroup, svc.name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return nil, err
			}
			groups = append(groups, page.Value...)
		}
		var resources []terraformutils.Resource
		for _, grp := range groups {
			if grp.ID == nil || grp.Name == nil {
				continue
			}
			if res, ok := g.simple(*grp.ID, *grp.Name, "azurerm_api_management_group"); ok {
				resources = append(resources, res)
			}
			users, err := drainPager(ctx, groupUserClient.NewListPager(svc.resourceGroup, svc.name, *grp.Name, nil),
				func(p armapimanagement.GroupUserClientListResponse) []*armapimanagement.UserContract { return p.Value },
				func(item *armapimanagement.UserContract) (terraformutils.Resource, bool) {
					if item.ID == nil || item.Name == nil {
						return terraformutils.Resource{}, false
					}
					// GroupUser list returns the user's service-scope ID; the
					// group_user importer ID is nested under the group.
					id := svc.id + "/groups/" + *grp.Name + "/users/" + idSegmentAfter(*item.ID, "/users/")
					return g.simple(id, *item.Name, "azurerm_api_management_group_user")
				})
			if err != nil {
				return resources, err
			}
			resources = append(resources, users...)
		}
		return resources, nil
	})
}

var apimIdentityProviderTypes = map[string]string{
	"aad":       "azurerm_api_management_identity_provider_aad",
	"aadb2c":    "azurerm_api_management_identity_provider_aadb2c",
	"facebook":  "azurerm_api_management_identity_provider_facebook",
	"google":    "azurerm_api_management_identity_provider_google",
	"microsoft": "azurerm_api_management_identity_provider_microsoft",
	"twitter":   "azurerm_api_management_identity_provider_twitter",
}

func (g *APIManagementGenerator) listIdentityProviders(ctx context.Context, refs []apimServiceRef) ([]terraformutils.Resource, error) {
	client, err := armapimanagement.NewIdentityProviderClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	return g.perService(ctx, refs, func(ctx context.Context, svc apimServiceRef) ([]terraformutils.Resource, error) {
		return drainPager(ctx, client.NewListByServicePager(svc.resourceGroup, svc.name, nil),
			func(p armapimanagement.IdentityProviderClientListByServiceResponse) []*armapimanagement.IdentityProviderContract {
				return p.Value
			},
			func(item *armapimanagement.IdentityProviderContract) (terraformutils.Resource, bool) {
				if item.ID == nil || item.Name == nil {
					return terraformutils.Resource{}, false
				}
				resourceType, ok := apimIdentityProviderTypes[strings.ToLower(*item.Name)]
				if !ok {
					return terraformutils.Resource{}, false
				}
				return g.simple(*item.ID, *item.Name, resourceType)
			})
	})
}

func (g *APIManagementGenerator) listLoggers(ctx context.Context, refs []apimServiceRef) ([]terraformutils.Resource, error) {
	client, err := armapimanagement.NewLoggerClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	return g.perService(ctx, refs, func(ctx context.Context, svc apimServiceRef) ([]terraformutils.Resource, error) {
		return drainPager(ctx, client.NewListByServicePager(svc.resourceGroup, svc.name, nil),
			func(p armapimanagement.LoggerClientListByServiceResponse) []*armapimanagement.LoggerContract {
				return p.Value
			},
			func(item *armapimanagement.LoggerContract) (terraformutils.Resource, bool) {
				if item.ID == nil || item.Name == nil {
					return terraformutils.Resource{}, false
				}
				return g.simple(*item.ID, *item.Name, "azurerm_api_management_logger")
			})
	})
}

func (g *APIManagementGenerator) listNamedValues(ctx context.Context, refs []apimServiceRef) ([]terraformutils.Resource, error) {
	client, err := armapimanagement.NewNamedValueClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	return g.perService(ctx, refs, func(ctx context.Context, svc apimServiceRef) ([]terraformutils.Resource, error) {
		return drainPager(ctx, client.NewListByServicePager(svc.resourceGroup, svc.name, nil),
			func(p armapimanagement.NamedValueClientListByServiceResponse) []*armapimanagement.NamedValueContract {
				return p.Value
			},
			func(item *armapimanagement.NamedValueContract) (terraformutils.Resource, bool) {
				if item.ID == nil || item.Name == nil {
					return terraformutils.Resource{}, false
				}
				return g.simple(*item.ID, *item.Name, "azurerm_api_management_named_value")
			})
	})
}

func (g *APIManagementGenerator) listNotificationRecipients(ctx context.Context, refs []apimServiceRef) ([]terraformutils.Resource, error) {
	notificationClient, err := armapimanagement.NewNotificationClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	emailClient, err := armapimanagement.NewNotificationRecipientEmailClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	userClient, err := armapimanagement.NewNotificationRecipientUserClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	return g.perService(ctx, refs, func(ctx context.Context, svc apimServiceRef) ([]terraformutils.Resource, error) {
		var resources []terraformutils.Resource
		pager := notificationClient.NewListByServicePager(svc.resourceGroup, svc.name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return resources, err
			}
			for _, notification := range page.Value {
				if notification.Name == nil {
					continue
				}
				name := armapimanagement.NotificationName(*notification.Name)
				emails, err := emailClient.ListByNotification(ctx, svc.resourceGroup, svc.name, name, nil)
				if err != nil {
					return resources, err
				}
				for _, item := range emails.Value {
					if item.ID == nil || item.Name == nil {
						continue
					}
					if res, ok := g.simple(*item.ID, *item.Name, "azurerm_api_management_notification_recipient_email"); ok {
						resources = append(resources, res)
					}
				}
				users, err := userClient.ListByNotification(ctx, svc.resourceGroup, svc.name, name, nil)
				if err != nil {
					return resources, err
				}
				for _, item := range users.Value {
					if item.ID == nil || item.Name == nil {
						continue
					}
					if res, ok := g.simple(*item.ID, *item.Name, "azurerm_api_management_notification_recipient_user"); ok {
						resources = append(resources, res)
					}
				}
			}
		}
		return resources, nil
	})
}

func (g *APIManagementGenerator) listOpenIDConnectProviders(ctx context.Context, refs []apimServiceRef) ([]terraformutils.Resource, error) {
	client, err := armapimanagement.NewOpenIDConnectProviderClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	return g.perService(ctx, refs, func(ctx context.Context, svc apimServiceRef) ([]terraformutils.Resource, error) {
		return drainPager(ctx, client.NewListByServicePager(svc.resourceGroup, svc.name, nil),
			func(p armapimanagement.OpenIDConnectProviderClientListByServiceResponse) []*armapimanagement.OpenidConnectProviderContract {
				return p.Value
			},
			func(item *armapimanagement.OpenidConnectProviderContract) (terraformutils.Resource, bool) {
				if item.ID == nil || item.Name == nil {
					return terraformutils.Resource{}, false
				}
				return g.simple(*item.ID, *item.Name, "azurerm_api_management_openid_connect_provider")
			})
	})
}

func (g *APIManagementGenerator) listProductsAndChildren(ctx context.Context, refs []apimServiceRef) ([]terraformutils.Resource, error) {
	productClient, err := armapimanagement.NewProductClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	apiClient, err := armapimanagement.NewProductAPIClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	groupClient, err := armapimanagement.NewProductGroupClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	tagClient, err := armapimanagement.NewTagClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	return g.perService(ctx, refs, func(ctx context.Context, svc apimServiceRef) ([]terraformutils.Resource, error) {
		var products []*armapimanagement.ProductContract
		pager := productClient.NewListByServicePager(svc.resourceGroup, svc.name, nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				return nil, err
			}
			products = append(products, page.Value...)
		}
		var resources []terraformutils.Resource
		for _, product := range products {
			if product.ID == nil || product.Name == nil {
				continue
			}
			if res, ok := g.simple(*product.ID, *product.Name, "azurerm_api_management_product"); ok {
				resources = append(resources, res)
			}
			productID := *product.Name
			apis, err := drainPager(ctx, apiClient.NewListByProductPager(svc.resourceGroup, svc.name, productID, nil),
				func(p armapimanagement.ProductAPIClientListByProductResponse) []*armapimanagement.APIContract {
					return p.Value
				},
				func(item *armapimanagement.APIContract) (terraformutils.Resource, bool) {
					if item.ID == nil || item.Name == nil {
						return terraformutils.Resource{}, false
					}
					// ProductAPI list returns service-scope API IDs; the
					// product_api importer ID is nested under the product.
					id := svc.id + "/products/" + productID + "/apis/" + idSegmentAfter(*item.ID, "/apis/")
					return g.simple(id, *item.Name, "azurerm_api_management_product_api")
				})
			if err != nil {
				return resources, err
			}
			resources = append(resources, apis...)
			groups, err := drainPager(ctx, groupClient.NewListByProductPager(svc.resourceGroup, svc.name, productID, nil),
				func(p armapimanagement.ProductGroupClientListByProductResponse) []*armapimanagement.GroupContract {
					return p.Value
				},
				func(item *armapimanagement.GroupContract) (terraformutils.Resource, bool) {
					if item.ID == nil || item.Name == nil {
						return terraformutils.Resource{}, false
					}
					id := svc.id + "/products/" + productID + "/groups/" + idSegmentAfter(*item.ID, "/groups/")
					return g.simple(id, *item.Name, "azurerm_api_management_product_group")
				})
			if err != nil {
				return resources, err
			}
			resources = append(resources, groups...)
			tags, err := drainPager(ctx, tagClient.NewListByProductPager(svc.resourceGroup, svc.name, productID, nil),
				func(p armapimanagement.TagClientListByProductResponse) []*armapimanagement.TagContract {
					return p.Value
				},
				func(item *armapimanagement.TagContract) (terraformutils.Resource, bool) {
					if item.ID == nil || item.Name == nil {
						return terraformutils.Resource{}, false
					}
					return g.simple(*item.ID, *item.Name, "azurerm_api_management_product_tag")
				})
			if err != nil {
				return resources, err
			}
			resources = append(resources, tags...)
		}
		return resources, nil
	})
}

func (g *APIManagementGenerator) listRedisCaches(ctx context.Context, refs []apimServiceRef) ([]terraformutils.Resource, error) {
	client, err := armapimanagement.NewCacheClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	return g.perService(ctx, refs, func(ctx context.Context, svc apimServiceRef) ([]terraformutils.Resource, error) {
		return drainPager(ctx, client.NewListByServicePager(svc.resourceGroup, svc.name, nil),
			func(p armapimanagement.CacheClientListByServiceResponse) []*armapimanagement.CacheContract {
				return p.Value
			},
			func(item *armapimanagement.CacheContract) (terraformutils.Resource, bool) {
				if item.ID == nil || item.Name == nil {
					return terraformutils.Resource{}, false
				}
				return g.simple(*item.ID, *item.Name, "azurerm_api_management_redis_cache")
			})
	})
}

func (g *APIManagementGenerator) listSubscriptions(ctx context.Context, refs []apimServiceRef) ([]terraformutils.Resource, error) {
	client, err := armapimanagement.NewSubscriptionClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	return g.perService(ctx, refs, func(ctx context.Context, svc apimServiceRef) ([]terraformutils.Resource, error) {
		return drainPager(ctx, client.NewListPager(svc.resourceGroup, svc.name, nil),
			func(p armapimanagement.SubscriptionClientListResponse) []*armapimanagement.SubscriptionContract {
				return p.Value
			},
			func(item *armapimanagement.SubscriptionContract) (terraformutils.Resource, bool) {
				if item.ID == nil || item.Name == nil {
					return terraformutils.Resource{}, false
				}
				return g.simple(*item.ID, *item.Name, "azurerm_api_management_subscription")
			})
	})
}

func (g *APIManagementGenerator) listTags(ctx context.Context, refs []apimServiceRef) ([]terraformutils.Resource, error) {
	client, err := armapimanagement.NewTagClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	return g.perService(ctx, refs, func(ctx context.Context, svc apimServiceRef) ([]terraformutils.Resource, error) {
		return drainPager(ctx, client.NewListByServicePager(svc.resourceGroup, svc.name, nil),
			func(p armapimanagement.TagClientListByServiceResponse) []*armapimanagement.TagContract {
				return p.Value
			},
			func(item *armapimanagement.TagContract) (terraformutils.Resource, bool) {
				if item.ID == nil || item.Name == nil {
					return terraformutils.Resource{}, false
				}
				return g.simple(*item.ID, *item.Name, "azurerm_api_management_tag")
			})
	})
}

func (g *APIManagementGenerator) listUsers(ctx context.Context, refs []apimServiceRef) ([]terraformutils.Resource, error) {
	client, err := armapimanagement.NewUserClient(g.getSubscriptionID(), g.getCredential(), g.armClientOptions())
	if err != nil {
		return nil, err
	}
	return g.perService(ctx, refs, func(ctx context.Context, svc apimServiceRef) ([]terraformutils.Resource, error) {
		return drainPager(ctx, client.NewListByServicePager(svc.resourceGroup, svc.name, nil),
			func(p armapimanagement.UserClientListByServiceResponse) []*armapimanagement.UserContract {
				return p.Value
			},
			func(item *armapimanagement.UserContract) (terraformutils.Resource, bool) {
				if item.ID == nil || item.Name == nil {
					return terraformutils.Resource{}, false
				}
				return g.simple(*item.ID, *item.Name, "azurerm_api_management_user")
			})
	})
}
