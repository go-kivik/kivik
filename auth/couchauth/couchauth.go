// Package couchauth provides auth services to a remote CouchDB server.
package couchauth

import (
	"context"
	"net/url"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/auth"
	"github.com/flimzy/kivik/driver/couchdb/chttp"
	"github.com/flimzy/kivik/errors"
)

type client struct {
	*chttp.Client
}

var _ auth.Handler = &client{}

// New returns a new auth handler, which authenticates users against a remote
// CouchDB server.
func New(dsn string) (auth.Handler, error) {
	p, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}
	if p.User != nil {
		return nil, errors.New("DSN must not contain authentication credentials")
	}
	c, err := chttp.New(dsn)
	return &client{c}, err
}

func (c *client) Validate(ctx context.Context, username, password string) (bool, error) {
	req, err := c.NewRequest(ctx, chttp.MethodGet, "/_session", nil)
	if err != nil {
		return false, err
	}
	req.SetBasicAuth(username, password)
	resp, err := c.Do(req)
	if err != nil {
		return false, err
	}
	err = chttp.ResponseError(resp)
	if errors.StatusCode(err) == kivik.StatusUnauthorized {
		return false, nil
	}
	return err == nil, err
}

func (c *client) Roles(ctx context.Context, username string) ([]string, error) {
	// var result struct {
	// 	Ctx struct {
	// 		Roles []string `json:"roles"`
	// 	} `json:"userCtx"`
	// }
	// return result.Ctx.Roles, c.DoJSON()
	return nil, nil
}
