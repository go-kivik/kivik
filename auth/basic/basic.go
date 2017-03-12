// Package basic provides HTTP Basic Auth services.
package basic

import (
	"net/http"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/auth"
	"github.com/flimzy/kivik/authdb"
)

// HTTPBasicAuth provides HTTP Basic Auth
type HTTPBasicAuth struct{}

var _ auth.Handler = &HTTPBasicAuth{}

// MethodName returns "default"
func (a *HTTPBasicAuth) MethodName() string {
	return "default" // For compatibility with the name used by CouchDB
}

// Authenticate authenticates a request against a user store using HTTP Basic
// Auth.
func (a *HTTPBasicAuth) Authenticate(r *http.Request, store authdb.UserStore) (*authdb.UserContext, error) {
	username, password, ok := r.BasicAuth()
	if !ok {
		return nil, kivik.ErrUnauthorized
	}
	return store.Validate(r.Context(), username, password)
}
