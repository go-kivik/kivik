package kivik

import (
	"context"

	"github.com/flimzy/kivik/config"
	"github.com/flimzy/kivik/driver"
)

// Config calls ConfigContext with a background context.
func (c *Client) Config() (*config.Config, error) {
	return c.ConfigContext(context.Background())
}

// ConfigContext returns the server's configuration.
func (c *Client) ConfigContext(ctx context.Context) (*config.Config, error) {
	if conf, ok := c.driverClient.(driver.Configer); ok {
		c, err := conf.ConfigContext(ctx)
		return config.New(c), err
	}
	return nil, ErrNotImplemented
}
