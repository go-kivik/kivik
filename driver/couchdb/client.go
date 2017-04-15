package couchdb

import (
	"context"
	"fmt"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/errors"
)

func (c *client) AllDBsContext(ctx context.Context, _ map[string]interface{}) ([]string, error) {
	var allDBs []string
	_, err := c.DoJSON(ctx, kivik.MethodGet, "/_all_dbs", nil, &allDBs)
	return allDBs, err
}

func (c *client) UUIDsContext(ctx context.Context, count int) ([]string, error) {
	var uuids struct {
		UUIDs []string `json:"uuids"`
	}
	_, err := c.DoJSON(ctx, kivik.MethodGet, fmt.Sprintf("/_uuids?count=%d", count), nil, &uuids)
	return uuids.UUIDs, err
}

func (c *client) DBExistsContext(ctx context.Context, dbName string, _ map[string]interface{}) (bool, error) {
	_, err := c.DoError(ctx, kivik.MethodHead, dbName, nil)
	if errors.StatusCode(err) == kivik.StatusNotFound {
		return false, nil
	}
	return err == nil, err
}

func (c *client) CreateDBContext(ctx context.Context, dbName string, _ map[string]interface{}) error {
	_, err := c.DoError(ctx, kivik.MethodPut, dbName, nil)
	return err
}

func (c *client) DestroyDBContext(ctx context.Context, dbName string, _ map[string]interface{}) error {
	_, err := c.DoError(ctx, kivik.MethodDelete, dbName, nil)
	return err
}
