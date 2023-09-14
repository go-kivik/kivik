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

package couchdb

import (
	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/couchdb/chttp"
)

// BasicAuth provides support for HTTP Basic authentication.  Pass this option
// to [github.com/go-kivik/kivik/v4.New] to use Basic Authentication.
func BasicAuth(username, password string) kivik.Option {
	return chttp.BasicAuth(username, password)
}

// CookieAuth provides CouchDB [Cookie auth]. Cookie Auth is the default
// authentication method if credentials are included in the connection URL
// passed to [github.com/go-kivik/kivik/v4.New]. You may also pass this option
// as an argument to the same function, if you need to provide your auth
// credentials outside of the URL.
//
// [Cookie auth]: http://docs.couchdb.org/en/2.0.0/api/server/authn.html#cookie-authentication
func CookieAuth(username, password string) kivik.Option {
	return chttp.CookieAuth(username, password)
}

// JWTAuth provides support for CouchDB JWT-based authentication. Kivik does
// no validation on the JWT token; it is passed verbatim to the server.
//
// See https://docs.couchdb.org/en/latest/api/server/authn.html#jwt-authentication
func JWTAuth(token string) kivik.Option {
	return chttp.JWTAuth(token)
}

// ProxyAuth provides support for CouchDB's [proxy authentication]. Pass this
// option to [github.com/go-kivik/kivik/v4.New] to use proxy authentication.
//
// The `secret` argument represents the [couch_httpd_auth/secret] value
// configured on the CouchDB server.
//
// If `secret` is the empty string, the X-Auth-CouchDB-Token header will not be
// set, to support disabling the [proxy_use_secret] server setting.
//
// The optional `headers` map may be passed to use non-standard header names.
// For instance, to use `X-User` in place of the `X-Auth-CouchDB-Username`
// header, pass a value of {"X-Auth-CouchDB-UserName": "X-User"}.
// The relevant headers are X-Auth-CouchDB-UserName, X-Auth-CouchDB-Roles, and
// X-Auth-CouchDB-Token.
//
// [proxy authentication]: https://docs.couchdb.org/en/stable/api/server/authn.html?highlight=proxy%20auth#proxy-authentication
// [couch_httpd_auth/secret]: https://docs.couchdb.org/en/stable/config/auth.html#couch_httpd_auth/secret
// [proxy_use_secret]: https://docs.couchdb.org/en/stable/config/auth.html#couch_httpd_auth/proxy_use_secret
func ProxyAuth(user, secret string, roles []string, headers ...map[string]string) kivik.Option {
	return chttp.ProxyAuth(user, secret, roles, headers...)
}
