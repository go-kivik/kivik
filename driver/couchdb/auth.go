package couchdb

import (
	"errors"

	"github.com/flimzy/kivik/driver/couchdb/chttp"
)

type authenticator interface {
	authenticate(*client) error
}

// BasicAuth provides basic HTTP Authentication services.
type BasicAuth struct {
	*chttp.BasicAuth
}

var _ authenticator = &BasicAuth{}

func (a *BasicAuth) authenticate(c *client) error {
	return c.Auth(a)
}

// CookieAuth provides CouchDB Cookie auth services as described at
// http://docs.couchdb.org/en/2.0.0/api/server/authn.html#cookie-authentication
type CookieAuth struct {
	*chttp.CookieAuth
}

var _ authenticator = &CookieAuth{}

func (a *CookieAuth) authenticate(c *client) error {
	return c.Auth(a)
}

func (c *client) Authenticate(a interface{}) error {
	if auth, ok := a.(authenticator); ok {
		return auth.authenticate(c)
	}
	return errors.New("invalid authenticator")
}
