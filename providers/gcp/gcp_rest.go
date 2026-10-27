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

package gcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"google.golang.org/api/option"
	htransport "google.golang.org/api/transport/http"
)

// iamBinding is a package-neutral view of a single IAM policy binding,
// extracted from whichever generated client's Policy type the caller used.
type iamBinding struct {
	Role    string
	Members []string
}

// newResourceFunc builds a terraformer resource for a generator; each
// generator passes its own constructor so the allow-empty-values and
// additional-fields conventions of its family are preserved.
type newResourceFunc func(id, resourceType string, attrs map[string]string) terraformutils.Resource

// expandIamPolicy expands a live IAM policy into the google provider's
// _iam_policy, _iam_binding and _iam_member resources for a parent
// resource, following the pattern established by the Network Security
// address groups: every emitted resource carries the parent's resource
// name as its ID, and the binding/member resources additionally carry
// the role (and member) in their attributes.
func expandIamPolicy(bindings []iamBinding, parentID, policyType, bindingType, memberType string, parentAttrs map[string]string, newResource newResourceFunc) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	resources = append(resources, newResource(parentID, policyType, parentAttrs))
	for _, binding := range bindings {
		bindingAttrs := map[string]string{"role": binding.Role}
		for k, v := range parentAttrs {
			bindingAttrs[k] = v
		}
		resources = append(resources, newResource(parentID, bindingType, bindingAttrs))
		for _, member := range binding.Members {
			memberAttrs := map[string]string{"role": binding.Role, "member": member}
			for k, v := range parentAttrs {
				memberAttrs[k] = v
			}
			resources = append(resources, newResource(parentID, memberType, memberAttrs))
		}
	}
	return resources
}

// newAuthenticatedHTTPClient builds the OAuth2 HTTP client used for the
// raw REST calls of APIs that have no generated client surface for the
// collections terraformer needs (Security Command Center v2 and
// Management APIs, parts of Dataplex and Analytics Hub).
func newAuthenticatedHTTPClient(ctx context.Context) (*http.Client, error) {
	client, _, err := htransport.NewClient(ctx, option.WithScopes("https://www.googleapis.com/auth/cloud-platform"))
	return client, err
}

// restResource is one item of a raw REST list response. AncestorModule is
// populated only by the Security Command Center custom module APIs,
// where a list at a folder or project also returns modules inherited
// from ancestors; those are owned (and emitted) at the ancestor scope.
type restResource struct {
	Name           string `json:"name"`
	AncestorModule string `json:"ancestorModule"`
}

// restListResources pages through a Google API list endpoint and returns
// every item's resource name (and ancestor module, when present). The
// collection field name differs per API, so the response is decoded
// generically: the single array-valued field is the collection.
func restListResources(ctx context.Context, client *http.Client, url string) ([]restResource, error) {
	resources := []restResource{}
	pageToken := ""
	for {
		pageURL := url
		if pageToken != "" {
			separator := "?"
			if strings.Contains(pageURL, "?") {
				separator = "&"
			}
			pageURL += separator + "pageToken=" + pageToken
		}
		body, err := restDo(ctx, client, http.MethodGet, pageURL, nil)
		if err != nil {
			return nil, err
		}
		var page map[string]json.RawMessage
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		nextPageToken := ""
		for field, raw := range page {
			if field == "nextPageToken" {
				_ = json.Unmarshal(raw, &nextPageToken)
				continue
			}
			var items []restResource
			if err := json.Unmarshal(raw, &items); err == nil {
				resources = append(resources, items...)
			}
		}
		if nextPageToken == "" {
			return resources, nil
		}
		pageToken = nextPageToken
	}
}

// restGetIamPolicy calls a google.iam.v1 :getIamPolicy endpoint (POST
// with an empty GetIamPolicyRequest body) and returns its bindings.
func restGetIamPolicy(ctx context.Context, client *http.Client, url string) ([]iamBinding, error) {
	body, err := restDo(ctx, client, http.MethodPost, url, strings.NewReader("{}"))
	if err != nil {
		return nil, err
	}
	var policy struct {
		Bindings []iamBinding `json:"bindings"`
	}
	if err := json.Unmarshal(body, &policy); err != nil {
		return nil, err
	}
	return policy.Bindings, nil
}

func restDo(ctx context.Context, client *http.Client, method, url string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("%s %s: %s: %s", method, url, resp.Status, strings.TrimSpace(string(responseBody)))
	}
	return responseBody, nil
}
