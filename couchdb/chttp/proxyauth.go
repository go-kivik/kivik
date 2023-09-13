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
	"fmt"
	"net/http"
	"strings"

	"github.com/go-kivik/kivik/v4"
)

type proxyAuth struct {
	Username string
	Secret   string
	Roles    []string
	Headers  http.Header

	transport http.RoundTripper
	token     string
}

var (
	_ authenticator = &proxyAuth{}
	_ kivik.Option  = (*proxyAuth)(nil)
)

func (a *proxyAuth) Apply(target interface{}) {
	if auth, ok := target.(*authenticator); ok {
		// Clone this so that it's safe to re-use the same option to multiple
		// client connections. TODO: This can no doubt be refactored.
		*auth = &proxyAuth{
			Username: a.Username,
			Secret:   a.Secret,
			Roles:    a.Roles,
			Headers:  a.Headers,
		}
	}
}

func (a *proxyAuth) String() string {
	return fmt.Sprintf("[ProxyAuth{username:%s,secret:%s}]", a.Username, strings.Repeat("*", len(a.Secret)))
}

func (a *proxyAuth) header(header string) string {
	if h := a.Headers.Get(header); h != "" {
		return http.CanonicalHeaderKey(h)
	}
	return header
}

func (a *proxyAuth) genToken() string {
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
func (a *proxyAuth) RoundTrip(req *http.Request) (*http.Response, error) {
	if token := a.genToken(); token != "" {
		req.Header.Set(a.header("X-Auth-CouchDB-Token"), token)
	}

	req.Header.Set(a.header("X-Auth-CouchDB-UserName"), a.Username)
	req.Header.Set(a.header("X-Auth-CouchDB-Roles"), strings.Join(a.Roles, ","))

	return a.transport.RoundTrip(req)
}

// Authenticate allows authentication via ProxyAuth.
func (a *proxyAuth) Authenticate(c *Client) error {
	a.transport = c.Transport
	if a.transport == nil {
		a.transport = http.DefaultTransport
	}
	c.Transport = a
	return nil
}
