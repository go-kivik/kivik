package couchdb

import (
	"errors"

	"github.com/flimzy/kivik/driver/couchdb/chttp"
)

type authenticator interface {
	Authenticate(*chttp.Client) error
}

func (c *client) Authenticate(a interface{}) error {
	if auth, ok := a.(authenticator); ok {
		return auth.Authenticate(c.Client)
	}
	return errors.New("invalid authenticator")
}
