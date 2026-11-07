package vault

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	vault "github.com/hashicorp/vault/api"
)

type ServiceGenerator struct { //nolint
	terraformutils.Service
	client    *vault.Client
	mountType string
	resource  string
}

func (g *ServiceGenerator) setVaultClient() error {
	client, err := vault.NewClient(&vault.Config{Address: g.Args["address"].(string)})
	if err != nil {
		return err
	}
	if g.Args["token"] != "" {
		client.SetToken(g.Args["token"].(string))
	}
	g.client = client
	return nil
}

func (g *ServiceGenerator) InitResources() error {
	switch g.resource {
	case "secret_backend":
		return g.createSecretBackendResources()
	case "secret_backend_role":
		return g.createSecretBackendRoleResources()
	case "auth_backend":
		return g.createAuthBackendResources()
	case "auth_backend_role":
		return g.createAuthBackendEntityResources("role", "role")
	case "auth_backend_user":
		return g.createAuthBackendEntityResources("users", "user")
	case "auth_backend_group":
		return g.createAuthBackendEntityResources("groups", "group")
	case "policy":
		return g.createPolicyResources()
	case "generic_secret":
		return g.createGenericSecretResources()
	case "mount":
		return g.createMountResources()
	default:
		return errors.New("unsupported service type. shouldn't ever reach here")
	}
}

func (g *ServiceGenerator) createSecretBackendResources() error {
	mounts, err := g.mountsByType()
	if err != nil {
		return err
	}
	for _, mount := range mounts {
		g.Resources = append(g.Resources,
			terraformutils.NewSimpleResource(
				mount,
				mount,
				fmt.Sprintf("vault_%s_secret_backend", g.mountType),
				g.ProviderName,
				[]string{}))
	}
	return nil
}

func (g *ServiceGenerator) createSecretBackendRoleResources() error {
	mounts, err := g.mountsByType()
	if err != nil {
		return err
	}
	for _, mount := range mounts {
		path := fmt.Sprintf("%s/roles", mount)
		s, err := g.client.Logical().List(path)
		if err != nil {
			log.Printf("error calling path %s: %s", path, err)
			continue
		}
		if s == nil {
			log.Printf("call to %s returned nil result", path)
			continue
		}
		roles, ok := s.Data["keys"]
		if !ok {
			log.Printf("no keys in call to %s", path)
			continue
		}
		for _, role := range roles.([]interface{}) {
			g.Resources = append(g.Resources,
				terraformutils.NewSimpleResource(
					fmt.Sprintf("%s/roles/%s", mount, role),
					fmt.Sprintf("%s_%s", mount, role),
					fmt.Sprintf("vault_%s_secret_backend_role", g.mountType),
					g.ProviderName,
					[]string{}))
		}
	}
	return nil
}

func (g *ServiceGenerator) mountsByType() ([]string, error) {
	mounts, err := g.client.Sys().ListMounts()
	if err != nil {
		return nil, err
	}
	var typeMounts []string
	for name, mount := range mounts {
		if g.mountType == "" || mount.Type == g.mountType {
			id := strings.ReplaceAll(name, "/", "")
			typeMounts = append(typeMounts, id)
		}
	}
	return typeMounts, nil
}

func (g *ServiceGenerator) createAuthBackendResources() error {
	backends, err := g.backendsByType()
	if err != nil {
		return err
	}
	for _, backend := range backends {
		g.Resources = append(g.Resources,
			terraformutils.NewSimpleResource(
				backend,
				backend,
				fmt.Sprintf("vault_%s_auth_backend", g.mountType),
				g.ProviderName,
				[]string{}))
	}
	return nil
}

func (g *ServiceGenerator) createAuthBackendEntityResources(apiEntity, tfEntity string) error {
	backends, err := g.backendsByType()
	if err != nil {
		return err
	}
	for _, backend := range backends {
		path := fmt.Sprintf("/auth/%s/%s", backend, apiEntity)
		s, err := g.client.Logical().List(path)
		if err != nil {
			log.Printf("error calling path %s: %s", path, err)
			continue
		}
		if s == nil {
			log.Printf("call to %s returned nil result", path)
			continue
		}
		names, ok := s.Data["keys"]
		if !ok {
			log.Printf("no keys in call to %s", path)
			continue
		}
		for _, name := range names.([]interface{}) {
			g.Resources = append(g.Resources,
				terraformutils.NewSimpleResource(
					fmt.Sprintf("auth/%s/%s/%s", backend, apiEntity, name),
					fmt.Sprintf("%s_%s", backend, name),
					fmt.Sprintf("vault_%s_auth_backend_%s", g.mountType, tfEntity),
					g.ProviderName,
					[]string{}))
		}
	}
	return nil
}

func (g *ServiceGenerator) backendsByType() ([]string, error) {
	authBackends, err := g.client.Sys().ListAuth()
	if err != nil {
		return nil, err
	}
	var typeBackends []string
	for name, authBackend := range authBackends {
		if authBackend.Type != g.mountType {
			continue
		}
		id := strings.ReplaceAll(name, "/", "")
		typeBackends = append(typeBackends, id)
	}
	return typeBackends, nil
}

func (g *ServiceGenerator) createPolicyResources() error {
	policies, err := g.client.Sys().ListPolicies()
	if err != nil {
		return err
	}
	for _, policy := range policies {
		if policy == "root" {
			continue
		}
		g.Resources = append(g.Resources,
			terraformutils.NewSimpleResource(
				policy,
				policy,
				"vault_policy",
				g.ProviderName,
				[]string{}))
	}
	return nil
}

// maxKVWalkDepth bounds the recursive secret discovery so a pathological
// or cyclic folder layout cannot loop forever.
const maxKVWalkDepth = 32

// kvMounts returns the kv mounts visible to the token, keyed by mount path
// (no trailing slash), with each mount's kv version ("1" or "2") taken from
// the mount options. A mount with no version option is kv-v1.
func (g *ServiceGenerator) kvMounts() (map[string]string, error) {
	mounts, err := g.client.Sys().ListMounts()
	if err != nil {
		return nil, err
	}
	out := map[string]string{}
	for name, mount := range mounts {
		if mount.Type != "kv" {
			continue
		}
		version := "1"
		if mount.Options != nil && mount.Options["version"] == "2" {
			version = "2"
		}
		out[strings.TrimSuffix(name, "/")] = version
	}
	return out, nil
}

// walkKVKeys recursively lists secret keys under listPath, which must end
// in "/". Keys ending in "/" are folders and are descended into; anything
// else is a leaf secret. Returned paths are relative to listPath. A path
// with no keys (or a 404, which the client reports as a nil secret) yields
// no leaves and no error.
func (g *ServiceGenerator) walkKVKeys(listPath string, depth int) ([]string, error) {
	if depth > maxKVWalkDepth {
		return nil, fmt.Errorf("secret discovery exceeded max depth %d at %s", maxKVWalkDepth, listPath)
	}
	s, err := g.client.Logical().List(listPath)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, nil
	}
	rawKeys, ok := s.Data["keys"]
	if !ok {
		return nil, nil
	}
	keys, ok := rawKeys.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected keys payload at %s", listPath)
	}
	var leaves []string
	for _, raw := range keys {
		key, ok := raw.(string)
		if !ok {
			continue
		}
		if strings.HasSuffix(key, "/") {
			sub, err := g.walkKVKeys(listPath+key, depth+1)
			if err != nil {
				return nil, err
			}
			for _, leaf := range sub {
				leaves = append(leaves, key+leaf)
			}
			continue
		}
		leaves = append(leaves, key)
	}
	return leaves, nil
}

func (g *ServiceGenerator) createGenericSecretResources() error {
	mounts, err := g.kvMounts()
	if err != nil {
		return err
	}
	for mount, version := range mounts {
		// kv-v1 exposes keys directly under the mount; kv-v2 lists them
		// under metadata/. Either way the resource ID is the logical path
		// (<mount>/<key>): the vault provider resolves kv-v2 mounts to
		// their data/ paths itself, and an ID that already contains the
		// data/ segment makes it report the object as non-existent.
		listRoot := mount + "/"
		idPrefix := mount + "/"
		if version == "2" {
			listRoot = mount + "/metadata/"
		}
		leaves, err := g.walkKVKeys(listRoot, 0)
		if err != nil {
			log.Printf("error listing secrets under %s: %s", listRoot, err)
			continue
		}
		for _, leaf := range leaves {
			g.Resources = append(g.Resources,
				terraformutils.NewSimpleResource(
					idPrefix+leaf,
					mount+"_"+strings.ReplaceAll(leaf, "/", "_"),
					"vault_generic_secret",
					g.ProviderName,
					[]string{}))
		}
	}
	return nil
}

func (g *ServiceGenerator) createMountResources() error {
	mounts, err := g.mountsByType()
	if err != nil {
		return err
	}
	for _, mount := range mounts {
		g.Resources = append(g.Resources,
			terraformutils.NewSimpleResource(
				mount,
				mount,
				"vault_mount",
				g.ProviderName,
				[]string{}))
	}
	return nil
}

func (g *ServiceGenerator) PostConvertHook() error {
	for _, resource := range g.Resources {
		switch resource.InstanceInfo.Type {
		case "vault_aws_secret_backend_role":
			if policyDocument, ok := resource.Item["policy_document"]; ok {
				// borrowed from providers/aws/aws_service.go
				sanitizedPolicy := regexp.MustCompile(`(\${[0-9A-Za-z:]+})`).
					ReplaceAllString(policyDocument.(string), "$$$1")
				resource.Item["policy_document"] = fmt.Sprintf(`<<POLICY
%s
POLICY`, sanitizedPolicy)
			}
		case "vault_ldap_auth_backend_group":
			if policies, ok := resource.Item["policies"]; ok {
				var strPolicies []string
				for _, policy := range policies.([]interface{}) {
					strPolicies = append(strPolicies, policy.(string))
				}
				sort.Strings(strPolicies)
				resource.Item["policies"] = strPolicies
			}
		}
	}
	return nil
}
