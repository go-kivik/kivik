package couchdb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/driver/couchdb/chttp"
)

func (c *client) Config(_ context.Context) (driver.Config, error) {
	return &config{client: c}, nil
}

type config struct {
	*client
}

var _ driver.Config = &config{}

func (c *config) GetAll(ctx context.Context) (map[string]map[string]string, error) {
	conf := map[string]map[string]string{}
	_, err := c.DoJSON(ctx, kivik.MethodGet, "/_config", nil, &conf)
	return conf, err
}

func (c *config) Set(ctx context.Context, secName, key, value string) error {
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(value); err != nil {
		return err
	}
	_, err := c.DoError(ctx, kivik.MethodPut, fmt.Sprintf("/_config/%s/%s", secName, key), &chttp.Options{Body: buf})
	return err
}

func (c *config) Delete(ctx context.Context, secName, key string) error {
	_, err := c.DoError(ctx, kivik.MethodDelete, fmt.Sprintf("/_config/%s/%s", secName, key), nil)
	return err
}

func (c *config) GetSection(ctx context.Context, secName string) (map[string]string, error) {
	sec := map[string]string{}
	_, err := c.DoJSON(ctx, kivik.MethodGet, fmt.Sprintf("/_config/%s", secName), nil, &sec)
	return sec, err
}

func (c *config) Get(ctx context.Context, secName, key string) (string, error) {
	var value string
	_, err := c.DoJSON(ctx, kivik.MethodGet, fmt.Sprintf("_config/%s/%s", secName, key), nil, &value)
	return value, err
}
