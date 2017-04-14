package couchdb

import (
	"context"
	"errors"

	"github.com/flimzy/kivik/driver/couchdb/chttp"
)

func (c *client) AuthenticateContext(ctx context.Context, a interface{}) error {
	if auth, ok := a.(chttp.Authenticator); ok {
		return auth.Authenticate(ctx, c.Client)
	}
	return errors.New("invalid authenticator")
}
