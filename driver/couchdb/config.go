package couchdb

import (
	"context"
	"fmt"

	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/driver/couchdb/chttp"
)

func (c *client) ConfigContext(_ context.Context) (driver.Config, error) {
	return &config{client: c}, nil
}

type config struct {
	*client
}

var _ driver.Config = &config{}

func (c *config) GetAllContext(ctx context.Context) (map[string]map[string]string, error) {
	conf := map[string]map[string]string{}
	return conf, c.DoJSON(ctx, chttp.MethodGet, "/_config", nil, &conf)
}

func (c *config) SetContext(ctx context.Context, secName, key, value string) error {
	return c.DoError(ctx, chttp.MethodPut, fmt.Sprintf("/_config/%s/%s", secName, key), &chttp.Options{JSON: value})
}

func (c *config) DeleteContext(ctx context.Context, secName, key string) error {
	return c.DoError(ctx, chttp.MethodDelete, fmt.Sprintf("/_config/%s/%s", secName, key), nil)
}

func (c *config) GetSectionContext(ctx context.Context, secName string) (map[string]string, error) {
	sec := map[string]string{}
	return sec, c.DoJSON(ctx, chttp.MethodGet, fmt.Sprintf("/_config/%s", secName), nil, &sec)
}

func (c *config) GetContext(ctx context.Context, secName, key string) (string, error) {
	var value string
	return value, c.DoJSON(ctx, chttp.MethodGet, fmt.Sprintf("_config/%s/%s", secName, key), nil, &value)
}
