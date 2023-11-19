// Package basic provides HTTP Basic Auth services.
package basic

import (
	"net/http"

	"github.com/go-kivik/kivikd/v4"
	"github.com/go-kivik/kivikd/v4/auth"
	"github.com/go-kivik/kivikd/v4/authdb"
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
func (a *HTTPBasicAuth) Authenticate(_ http.ResponseWriter, r *http.Request) (*authdb.UserContext, error) {
	store := kivikd.GetService(r).UserStore
	username, password, ok := r.BasicAuth()
	if !ok {
		return nil, nil
	}
	return store.Validate(r.Context(), username, password)
}
