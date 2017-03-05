package kivik

import "github.com/flimzy/kivik/driver"

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
