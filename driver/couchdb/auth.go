package couchdb

import (
	"errors"
	"net/http"
)

// Authenticator is an authentication interface, which may be implemented by
// any number of HTTP-centric authentication types.
type Authenticator interface {
	SetAuth(*http.Request) error
}

// BasicAuth provides basic HTTP Authentication services.
type BasicAuth struct {
	Name     string
	Password string
}

// SetAuth sets HTTP Basic Auth on the request.
func (a *BasicAuth) SetAuth(req *http.Request) error {
	req.SetBasicAuth(a.Name, a.Password)
	return nil
}

func (c *client) SetAuth(a interface{}) error {
	_, ok := a.(Authenticator)
	if !ok {
		return errors.New("invalid authenticator")
	}
	return nil
}
