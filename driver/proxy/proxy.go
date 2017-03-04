package proxy

import (
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver"
)

// NewClient wraps an existing *kivik.Client connection, allowing it to be used
// as a driver.Client
func NewClient(c *kivik.Client) driver.Client {
	return &client{c}
}

type client struct {
	*kivik.Client
}

var _ driver.Client = &client{}

func (c *client) DB(name string) (driver.DB, error) {
	d, err := c.Client.DB(name)
	return &db{d}, err
}

type db struct {
	*kivik.DB
}

var _ driver.DB = &db{}

func (d *db) AllDocs(i interface{}, opts map[string]interface{}) (int, int, string, error) {
	return d.DB.AllDocs(i, opts)
}

func (d *db) Get(id string, i interface{}, opts map[string]interface{}) error {
	return d.DB.Get(id, i, opts)
}
