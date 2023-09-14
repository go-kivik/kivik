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
	"fmt"
	"net/http"
	"strings"

	"github.com/go-kivik/kivik/v4"
)

// BasicAuth provides HTTP Basic Auth for a client.
type basicAuth struct {
	Username string
	Password string

	// transport stores the original transport that is overridden by this auth
	// mechanism
	transport http.RoundTripper
}

var (
	_ authenticator = &basicAuth{}
	_ kivik.Option  = (*basicAuth)(nil)
)

func (a *basicAuth) Apply(target interface{}) {
	if auth, ok := target.(*authenticator); ok {
		// Clone this so that it's safe to re-use the same option to multiple
		// client connections. TODO: This can no doubt be refactored.
		*auth = &basicAuth{
			Username: a.Username,
			Password: a.Password,
		}
	}
}

func (a *basicAuth) String() string {
	return fmt.Sprintf("[BasicAuth{user:%s,pass:%s}]", a.Username, strings.Repeat("*", len(a.Password)))
}

// RoundTrip fulfills the http.RoundTripper interface. It sets HTTP Basic Auth
// on outbound requests.
func (a *basicAuth) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(a.Username, a.Password)
	return a.transport.RoundTrip(req)
}

// Authenticate sets HTTP Basic Auth headers for the client.
func (a *basicAuth) Authenticate(c *Client) error {
	a.transport = c.Transport
	if a.transport == nil {
		a.transport = http.DefaultTransport
	}
	c.Transport = a
	return nil
}

type jwtAuth struct {
	Token string

	transport http.RoundTripper
}

var _ kivik.Option = (*jwtAuth)(nil)

func (a *jwtAuth) Apply(target interface{}) {
	if auth, ok := target.(*authenticator); ok {
		// Clone this so that it's safe to re-use the same option to multiple
		// client connections. TODO: This can no doubt be refactored.
		*auth = &jwtAuth{
			Token: a.Token,
		}
	}
}

func (a *jwtAuth) String() string {
	token := a.Token
	const unmaskedLen = 3
	if len(token) > unmaskedLen {
		token = token[:unmaskedLen] + strings.Repeat("*", len(token)-unmaskedLen)
	}
	return fmt.Sprintf("[JWTAuth{token:%s}]", token)
}

// RoundTrip satisfies the http.RoundTripper interface.
func (a *jwtAuth) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+a.Token)
	return a.transport.RoundTrip(req)
}

// Authenticate performs authentication against CouchDB.
func (a *jwtAuth) Authenticate(c *Client) error {
	a.transport = c.Transport
	if a.transport == nil {
		a.transport = http.DefaultTransport
	}
	c.Transport = a
	return nil
}
