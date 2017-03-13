// Package cookie provides standard CouchDB cookie auth as described at
// http://docs.couchdb.org/en/2.0.0/api/server/authn.html#cookie-authentication
package cookie

import (
	"net/http"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/auth"
	"github.com/flimzy/kivik/authdb"
	"github.com/flimzy/kivik/serve"
)

// Auth provides CouchDB Cookie authentication.
type Auth struct{}

var _ auth.Handler = &Auth{}

// MethodName returns "cookie"
func (a *Auth) MethodName() string {
	return "cookie" // For compatibility with the name used by CouchDB
}

// Authenticate authenticates a request with cookie auth against the user store.
func (a *Auth) Authenticate(w http.ResponseWriter, r *http.Request) (*authdb.UserContext, error) {
	store := serve.GetService(r)
	cookie, err := r.Cookie(kivik.SessionCookieName)
	if err != nil {
		return nil, nil
	}
	name, _, err := serve.DecodeCookie(cookie.Value)
	if err != nil {
		return nil, nil
	}
	user, err := store.UserCtx(r.Context(), name)
	if err != nil {
		// Failed to look up the user
		return nil, nil
	}
	s := serve.GetService(r)
	valid, err := s.ValidateCookie(r.Context(), user, cookie.Value)
	if err != nil {
		return nil, nil
	}
	if !valid {
		return nil, kivik.ErrUnauthorized
	}
	return user, nil
}
