// Package basic provides HTTP Basic Auth services.
package basic

import (
	"net/http"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/auth"
	"github.com/flimzy/kivik/authdb"
)

type basic struct {
	store authdb.UserStore
}

var _ auth.Handler = &basic{}

// New returns a new HTTP Basic Auth handler.
func New(store authdb.UserStore) auth.Handler {
	return &basic{
		store: store,
	}
}

func (a *basic) MethodName() string {
	return "default" // For compatibility with the name used by CouchDB
}

func (a *basic) Authenticate(r *http.Request) (*authdb.UserContext, error) {
	username, password, ok := r.BasicAuth()
	if !ok {
		return nil, kivik.ErrUnauthorized
	}
	return a.store.Validate(r.Context(), username, password)
}
