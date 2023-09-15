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
	"context"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"

	kivik "github.com/go-kivik/kivik/v4"
)

// cookieAuth provides CouchDB Cookie auth services as described at
// http://docs.couchdb.org/en/2.0.0/api/server/authn.html#cookie-authentication
//
// CookieAuth stores authentication state after use, so should not be re-used.
type cookieAuth struct {
	Username string `json:"name"`
	Password string `json:"password"`

	client *Client
	// transport stores the original transport that is overridden by this auth
	// mechanism
	transport http.RoundTripper
}

var (
	_ authenticator = &cookieAuth{}
	_ kivik.Option  = (*cookieAuth)(nil)
)

func (a *cookieAuth) Apply(target interface{}) {
	if auth, ok := target.(*authenticator); ok {
		// Clone this so that it's safe to re-use the same option to multiple
		// client connections. TODO: This can no doubt be refactored.
		*auth = &cookieAuth{
			Username: a.Username,
			Password: a.Password,
		}
	}
}

func (a *cookieAuth) String() string {
	return fmt.Sprintf("[CookieAuth{user:%s,pass:%s}]", a.Username, strings.Repeat("*", len(a.Password)))
}

// Authenticate initiates a session with the CouchDB server.
func (a *cookieAuth) Authenticate(c *Client) error {
	a.client = c
	a.setCookieJar()
	a.transport = c.Transport
	if a.transport == nil {
		a.transport = http.DefaultTransport
	}
	c.Transport = a
	return nil
}

// shouldAuth returns true if there is no cookie set, or if it has expired.
func (a *cookieAuth) shouldAuth(req *http.Request) bool {
	if _, err := req.Cookie(kivik.SessionCookieName); err == nil {
		return false
	}
	cookie := a.Cookie()
	if cookie == nil {
		return true
	}
	if !cookie.Expires.IsZero() {
		return cookie.Expires.Before(time.Now().Add(time.Minute))
	}
	// If we get here, it means the server did not include an expiry time in
	// the session cookie. Some CouchDB configurations do this, but rather than
	// re-authenticating for every request, we'll let the session expire. A
	// future change might be to make a client-configurable option to set the
	// re-authentication timeout.
	return false
}

// Cookie returns the current session cookie if found, or nil if not.
func (a *cookieAuth) Cookie() *http.Cookie {
	if a.client == nil {
		return nil
	}
	for _, cookie := range a.client.Jar.Cookies(a.client.dsn) {
		if cookie.Name == kivik.SessionCookieName {
			return cookie
		}
	}
	return nil
}

var authInProgress = &struct{ name string }{"in progress"}

// RoundTrip fulfills the http.RoundTripper interface. It sets
// (re-)authenticates when the cookie has expired or is not yet set.
// It also drops the auth cookie if we receive a 401 response to ensure
// that follow up requests can try to authenticate again.
func (a *cookieAuth) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := a.authenticate(req); err != nil {
		return nil, err
	}

	res, err := a.transport.RoundTrip(req)
	if err != nil {
		return res, err
	}

	if res != nil && res.StatusCode == http.StatusUnauthorized {
		if cookie := a.Cookie(); cookie != nil {
			// set to expire yesterday to allow us to ditch it
			cookie.Expires = time.Now().AddDate(0, 0, -1)
			a.client.Jar.SetCookies(a.client.dsn, []*http.Cookie{cookie})
		}
	}
	return res, nil
}

func (a *cookieAuth) authenticate(req *http.Request) error {
	ctx := req.Context()
	if inProg, _ := ctx.Value(authInProgress).(bool); inProg {
		return nil
	}
	if !a.shouldAuth(req) {
		return nil
	}
	a.client.authMU.Lock()
	defer a.client.authMU.Unlock()
	if c := a.Cookie(); c != nil {
		// In case another simultaneous process authenticated successfully first
		req.AddCookie(c)
		return nil
	}
	ctx = context.WithValue(ctx, authInProgress, true)
	opts := &Options{
		GetBody: BodyEncoder(a),
		Header: http.Header{
			HeaderIdempotencyKey: []string{},
		},
	}
	if _, err := a.client.DoError(ctx, http.MethodPost, "/_session", opts); err != nil {
		return err
	}
	if c := a.Cookie(); c != nil {
		req.AddCookie(c)
	}
	return nil
}

func (a *cookieAuth) setCookieJar() {
	// If a jar is already set, just use it
	if a.client.Jar != nil {
		return
	}
	// cookiejar.New never returns an error
	jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	a.client.Jar = jar
}
