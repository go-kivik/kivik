package kivik

import (
	"context"

	"github.com/flimzy/kivik/config"
	"github.com/flimzy/kivik/driver"
)

// Config returns the server's configuration.
//
// DO NOT USE THIS FUNCTION.
//
// This functionality is going away soon!
func (c *Client) Config(ctx context.Context) (*config.Config, error) {
	if conf, ok := c.driverClient.(driver.Configer); ok {
		c, err := conf.Config(ctx)
		return config.New(c), err
	}
	return nil, ErrNotImplemented
}
