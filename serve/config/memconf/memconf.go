// Package memconf provides an in-memory configuration backend. Changes are
// discarded on program termination.
package memconf

import (
	"context"
	"net/http"

	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/errors"
)

// Config is a configuration instance.
type Config map[string]map[string]string

var _ driver.Config = &Config{}

// New returns an empty configuration.
func New() *Config {
	return &Config{}
}

// GetAll returns the full configuration tree.
func (c Config) GetAll(_ context.Context) (map[string]map[string]string, error) {
	return c, nil
}

// Set sets a configuration value.
func (c Config) Set(_ context.Context, secName, key, value string) error {
	if _, ok := c[secName]; !ok {
		c[secName] = make(map[string]string)
	}
	c[secName][key] = value
	return nil
}

// Delete clears a configuration key.
func (c Config) Delete(_ context.Context, secName, key string) error {
	if _, ok := c[secName]; !ok {
		return errors.Status(http.StatusNotFound, "configuration section not found")
	}
	if _, ok := c[secName][key]; !ok {
		return errors.Status(http.StatusNotFound, "configuration key not found")
	}
	delete(c[secName], key)
	return nil
}
