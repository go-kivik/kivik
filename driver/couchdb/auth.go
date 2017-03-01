package couchdb

import "errors"

type authenticator interface {
	authenticate(*client) error
}

func (c *client) Authenticate(a interface{}) error {
	if auth, ok := a.(authenticator); ok {
		return auth.authenticate(c)
	}
	return errors.New("invalid authenticator")
}
