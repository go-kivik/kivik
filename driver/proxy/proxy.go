package proxy

import (
	"context"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver"
)

// CompleteClient is a composite of all compulsory and optional driver.* client
// interfaces.
type CompleteClient interface {
	driver.Client
	driver.Authenticator
	driver.UUIDer
	driver.LogReader
	driver.Cluster
	driver.Configer
}

// NewClient wraps an existing *kivik.Client connection, allowing it to be used
// as a driver.Client
func NewClient(c *kivik.Client) CompleteClient {
	return &client{c}
}

type client struct {
	*kivik.Client
}

var _ CompleteClient = &client{}

func (c *client) DBContext(ctx context.Context, name string) (driver.DB, error) {
	d, err := c.Client.DBContext(ctx, name)
	return &db{d}, err
}

func (c *client) ConfigContext(ctx context.Context) (driver.Config, error) {
	return c.ConfigContext(ctx)
}

type db struct {
	*kivik.DB
}

var _ driver.DB = &db{}

func (d *db) AllDocsContext(ctx context.Context, i interface{}, opts map[string]interface{}) (int, int, string, error) {
	return d.DB.AllDocsContext(ctx, i, opts)
}

func (d *db) GetContext(ctx context.Context, id string, i interface{}, opts map[string]interface{}) error {
	return d.DB.GetContext(ctx, id, i, opts)
}

func (d *db) InfoContext(ctx context.Context) (*driver.DBInfo, error) {
	i, err := d.DB.InfoContext(ctx)
	// For Go 1.7
	return &driver.DBInfo{
		Name:         i.Name,
		DocCount:     i.DocCount,
		DeletedCount: i.DeletedCount,
		UpdateSeq:    i.UpdateSeq,
		DiskSize:     i.DiskSize,
		ActiveSize:   i.ActiveSize,
		ExternalSize: i.ExternalSize,
	}, err
}
