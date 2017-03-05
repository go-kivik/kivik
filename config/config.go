package config

import (
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

// GetAll returns the complete server configuration.
func (c *Config) GetAll() (map[string]map[string]string, error) {
	return c.Config.GetAll()
}

// Set sets the specified configuration option.
func (c *Config) Set(secName, key, value string) error {
	return c.Config.Set(secName, key, value)
}

// Delete deletes the specified key from the configuration.
func (c *Config) Delete(secName, key string) error {
	return c.Config.Delete(secName, key)
}

// GetSection returns a complete config section.
func (c *Config) GetSection(secName string) (map[string]string, error) {
	if sectioner, ok := c.Config.(driver.ConfigSection); ok {
		return sectioner.GetSection(secName)
	}
	conf, err := c.GetAll()
	if err != nil {
		return nil, err
	}
	if sec, ok := conf[secName]; ok {
		return sec, nil
	}
	return nil, errors.Statusf(http.StatusNotFound, "configuration section '%s' not found", secName)
}

// Get retrieves a specific config value.
func (c *Config) Get(secName, key string) (string, error) {
	if itemer, ok := c.Config.(driver.ConfigItem); ok {
		return itemer.Get(secName, key)
	}
	sec, err := c.GetSection(secName)
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
