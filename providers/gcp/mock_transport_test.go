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
	"io"
	"net/http"
	"strings"
)

// cannedRoute binds a URL path substring to a canned JSON body.
type cannedRoute struct {
	substr string
	body   string
}

// cannedTransport serves canned JSON for the google.golang.org/api clients,
// so generator enumeration can be tested with no credentials and no network.
// Routes are matched in order; the first substring match wins. Unmatched
// requests return an empty JSON object.
type cannedTransport struct {
	routes []cannedRoute
	urls   []string
}

func (c *cannedTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	c.urls = append(c.urls, r.URL.String())
	body := "{}"
	for _, route := range c.routes {
		if strings.Contains(r.URL.Path, route.substr) {
			body = route.body
			break
		}
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Request:    r,
	}, nil
}

func (c *cannedTransport) client() *http.Client {
	return &http.Client{Transport: c}
}
