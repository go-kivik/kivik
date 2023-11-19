// Package couchauth provides auth services to a remote CouchDB server.
package couchauth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/go-kivik/kivik/v4/couchdb/chttp"
	"github.com/go-kivik/kivikd/v4/authdb"
)

type client struct {
	*chttp.Client
}

var _ authdb.UserStore = &client{}

// New returns a new auth user store, which authenticates users against a remote
// CouchDB server.
func New(dsn string) (authdb.UserStore, error) {
	p, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}
	if p.User != nil {
		return nil, errors.New("DSN must not contain authentication credentials")
	}
	c, err := chttp.New(&http.Client{}, dsn, nil)
	return &client{c}, err
}

func (c *client) Validate(ctx context.Context, username, password string) (*authdb.UserContext, error) {
	req, err := c.NewRequest(ctx, http.MethodGet, "/_session", nil, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(username, password)
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if err = chttp.ResponseError(resp); err != nil {
		return nil, err
	}
	result := struct {
		Ctx struct {
			Name  string   `json:"name"`
			Roles []string `json:"roles"`
		}
	}{}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	if err = dec.Decode(&result); err != nil {
		return nil, err
	}
	return &authdb.UserContext{
		Name:  result.Ctx.Name,
		Roles: result.Ctx.Roles,
		Salt:  "", // FIXME
	}, nil
}

func (c *client) UserCtx(context.Context, string) (*authdb.UserContext, error) {
	// var result struct {
	// 	Ctx struct {
	// 		Roles []string `json:"roles"`
	// 	} `json:"userCtx"`
	// }
	// return result.Ctx.Roles, c.DoJSON()
	return nil, nil
}
