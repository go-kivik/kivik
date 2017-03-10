package config

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/errors"
)

// Config allows reading and setting CouchDB server configuration.
type Config struct {
	driver.Config
}

// New instantiates a new configuration interface.
func New(conf driver.Config) *Config {
	return &Config{conf}
}

// GetAll calls GetAllContext with a background context.
func (c *Config) GetAll() (map[string]map[string]string, error) {
	return c.GetAllContext(context.Background())
}

// GetAllContext returns the complete server configuration.
func (c *Config) GetAllContext(ctx context.Context) (map[string]map[string]string, error) {
	return c.Config.GetAllContext(ctx)
}

// Set calls SetContext with a background context.
func (c *Config) Set(secName, key, value string) error {
	return c.SetContext(context.Background(), secName, key, value)
}

// SetContext sets the specified configuration option.
func (c *Config) SetContext(ctx context.Context, secName, key, value string) error {
	return c.Config.SetContext(ctx, secName, key, value)
}

// Delete calls DeleteContext with a background context.
func (c *Config) Delete(secName, key string) error {
	return c.DeleteContext(context.Background(), secName, key)
}

// DeleteContext deletes the specified key from the configuration.
func (c *Config) DeleteContext(ctx context.Context, secName, key string) error {
	return c.Config.DeleteContext(ctx, secName, key)
}

// GetSection calls GetSectionContext with a background context.
func (c *Config) GetSection(secName string) (map[string]string, error) {
	return c.GetSectionContext(context.Background(), secName)
}

// GetSectionContext returns a complete config section.
func (c *Config) GetSectionContext(ctx context.Context, secName string) (map[string]string, error) {
	if sectioner, ok := c.Config.(driver.ConfigSection); ok {
		sec, err := sectioner.GetSectionContext(ctx, secName)
		if errors.StatusCode(err) == http.StatusNotFound {
			err = nil
		}
		return sec, err
	}
	conf, err := c.GetAllContext(ctx)
	if err != nil {
		return nil, err
	}
	return conf[secName], nil
}

// Get calls GetContext with a background context.
func (c *Config) Get(secName, key string) (string, error) {
	return c.GetContext(context.Background(), secName, key)
}

// GetContext retrieves a specific config value.
func (c *Config) GetContext(ctx context.Context, secName, key string) (string, error) {
	if itemer, ok := c.Config.(driver.ConfigItem); ok {
		return itemer.GetContext(ctx, secName, key)
	}
	sec, err := c.GetSectionContext(ctx, secName)
	if err != nil {
		return "", err
	}
	if value, ok := sec[key]; ok {
		return value, nil
	}
	return "", errors.Status(http.StatusNotFound, "config key not found")
}

// IsSet returns true iff the requested key is set.
func (c *Config) IsSet(secName, key string) bool {
	_, err := c.Get(secName, key)
	return err == nil
}

// GetString returns the requested value as a string.
func (c *Config) GetString(secName, key string) string {
	value, _ := c.Get(secName, key)
	return value
}

// GetInt returns the requested value as an int64.
func (c *Config) GetInt(secName, key string) int64 {
	value, _ := c.Get(secName, key)
	i, _ := strconv.ParseInt(value, 10, 64)
	return i
}

// GetBool returns the requested value as a boolean. A value is considered
// true if it equals "true" (case insensitive).
func (c *Config) GetBool(secName, key string) bool {
	value, _ := c.Get(secName, key)
	return strings.ToLower(value) == "true"
}
