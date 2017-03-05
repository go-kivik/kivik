package kivik

import (
	"net/http"

	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/errors"
)

// Config allows reading and setting CouchDB server configuration.
type Config struct {
	driver.Config
}

// Config returns the server's configuration.
func (c *Client) Config() (*Config, error) {
	if conf, ok := c.driverClient.(driver.Configer); ok {
		c, err := conf.Config()
		return &Config{c}, err
	}
	return nil, ErrNotImplemented
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
	return nil, errors.Status(http.StatusNotFound, "section not found")
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
