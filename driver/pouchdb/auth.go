package pouchdb

import "errors"

// Authenticator is an authentication interface, which may be implemented by
// any PouchDB-centric authentication type..
type Authenticator interface {
	Authenticate(Options) error
}

// BasicAuth handles HTTP Basic Auth for remote PouchDB connections. This
// is the only auth support built directly into PouchDB, so this is a very
// thin wrapper.
type BasicAuth struct {
	Name     string
	Password string
}

// Authenticate sets the HTTP Basic Auth parameters.
func (a *BasicAuth) Authenticate(opts Options) error {
	if _, ok := opts["auth"]; !ok {
		opts["auth"] = make(map[string]interface{})
	}
	auth, ok := opts["auth"].(map[string]interface{})
	if !ok {
		panic("unexpected type for options.auth")
	}
	auth["username"] = a.Name
	if a.Password != "" {
		auth["password"] = a.Password
	}
	return nil
}

func (c *client) SetAuth(a interface{}) error {
	if a == nil {
		c.auth = nil
		return nil
	}
	authenticator, ok := a.(Authenticator)
	if !ok {
		return errors.New("invalid authenticator")
	}
	c.auth = authenticator
	return nil
}
