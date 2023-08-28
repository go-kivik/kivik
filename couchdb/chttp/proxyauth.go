// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

package chttp

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"net/http"
	"strings"
)

// ProxyAuth provides support for CouchDB proxy authentication.
type ProxyAuth struct {
	Username string
	Secret   string
	Roles    []string
	Headers  http.Header

	transport http.RoundTripper
	token     string
}

var _ Authenticator = &ProxyAuth{}

func (a *ProxyAuth) header(header string) string {
	if h := a.Headers.Get(header); h != "" {
		return http.CanonicalHeaderKey(h)
	}
	return header
}

func (a *ProxyAuth) genToken() string {
	if a.Secret == "" {
		return ""
	}
	if a.token != "" {
		return a.token
	}
	// Generate auth token
	// https://docs.couchdb.org/en/stable/config/auth.html#couch_httpd_auth/x_auth_token
	h := hmac.New(sha1.New, []byte(a.Secret))
	_, _ = h.Write([]byte(a.Username))
	a.token = hex.EncodeToString(h.Sum(nil))
	return a.token
}

// RoundTrip implements the http.RoundTripper interface.
func (a *ProxyAuth) RoundTrip(req *http.Request) (*http.Response, error) {
	if token := a.genToken(); token != "" {
		req.Header.Set(a.header("X-Auth-CouchDB-Token"), token)
	}

	req.Header.Set(a.header("X-Auth-CouchDB-UserName"), a.Username)
	req.Header.Set(a.header("X-Auth-CouchDB-Roles"), strings.Join(a.Roles, ","))

	return a.transport.RoundTrip(req)
}

// Authenticate allows authentication via ProxyAuth.
func (a *ProxyAuth) Authenticate(c *Client) error {
	a.transport = c.Transport
	if a.transport == nil {
		a.transport = http.DefaultTransport
	}
	c.Transport = a
	return nil
}
