package couchdb

import (
	"context"
	"fmt"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver/couchdb/chttp"
	"github.com/flimzy/kivik/errors"
)

func (c *client) AllDBsContext(ctx context.Context) ([]string, error) {
	var allDBs []string
	return allDBs, c.DoJSON(ctx, chttp.MethodGet, "/_all_dbs", nil, &allDBs)
}

func (c *client) UUIDsContext(ctx context.Context, count int) ([]string, error) {
	var uuids struct {
		UUIDs []string `json:"uuids"`
	}
	return uuids.UUIDs, c.DoJSON(ctx, chttp.MethodGet, fmt.Sprintf("/_uuids?count=%d", count), nil, &uuids)
}

// MembershipContext returns membership information. As a special case, if Couch
// 1.6 compatibility mode is enabled, this method returns Not Implemented
// immediately, rather than making the HTTP request.
func (c *client) MembershipContext(ctx context.Context) ([]string, []string, error) {
	if c.Compat == CompatCouch16 {
		return nil, nil, kivik.ErrNotImplemented
	}
	var membership struct {
		All     []string `json:"all_nodes"`
		Cluster []string `json:"cluster_nodes"`
	}
	return membership.All, membership.Cluster, c.DoJSON(ctx, chttp.MethodGet, "/_membership", nil, &membership)
}

func (c *client) DBExistsContext(ctx context.Context, dbName string) (bool, error) {
	err := c.DoError(ctx, chttp.MethodHead, dbName, nil)
	if errors.StatusCode(err) == kivik.StatusNotFound {
		return false, nil
	}
	return err == nil, err
}

func (c *client) CreateDBContext(ctx context.Context, dbName string) error {
	return c.DoError(ctx, chttp.MethodPut, dbName, nil)
}

func (c *client) DestroyDBContext(ctx context.Context, dbName string) error {
	return c.DoError(ctx, chttp.MethodDelete, dbName, nil)
}
