// Package cookie provides standard CouchDB cookie auth as described at
// http://docs.couchdb.org/en/2.0.0/api/server/authn.html#cookie-authentication
package cookie

import (
	"net/http"

	"github.com/flimzy/kivik/auth"
	"github.com/flimzy/kivik/authdb"
)

// CookieAuth provides CouchDB Cookie authentication.
type CookieAuth struct{}

var _ auth.Handler = &CookieAuth{}

// MethodName returns "cookie"
func (a *CookieAuth) MethodName() string {
	return "cookie" // For compatibility with the name used by CouchDB
}

// Authenticate authenticates a request with cookie auth against the user store.
func (a *CookieAuth) Authenticate(r *http.Request, store authdb.UserStore) (*authdb.UserContext, error) {
	return nil, nil
}
