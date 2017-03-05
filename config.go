package kivik

import (
	"github.com/flimzy/kivik/config"
	"github.com/flimzy/kivik/driver"
)

// Config returns the server's configuration.
func (c *Client) Config() (*config.Config, error) {
	if conf, ok := c.driverClient.(driver.Configer); ok {
		c, err := conf.Config()
		return config.New(c), err
	}
	return nil, ErrNotImplemented
}
