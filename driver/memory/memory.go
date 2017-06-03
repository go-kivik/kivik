// Package memory provides a memory-backed Kivik driver, intended for testing.
package memory

import (
	"context"
	"net/http"
	"sync"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/driver/common"
	"github.com/flimzy/kivik/errors"
)

type memDriver struct{}

var _ driver.Driver = &memDriver{}

func init() {
	kivik.Register("memory", &memDriver{})
}

type client struct {
	*common.Client
	mutex sync.RWMutex
	dbs   map[string]*database
}

var _ driver.Client = &client{}

// Identifying constants
const (
	Version = "0.0.1"
	Vendor  = "Kivik Memory Adaptor"
)

func (d *memDriver) NewClient(_ context.Context, name string) (driver.Client, error) {
	return &client{
		Client: common.NewClient(Version, Vendor),
		dbs:    make(map[string]*database),
	}, nil
}

func (c *client) AllDBs(_ context.Context, _ map[string]interface{}) ([]string, error) {
	dbs := make([]string, 0, len(c.dbs))
	for k := range c.dbs {
		dbs = append(dbs, k)
	}
	return dbs, nil
}

func (c *client) DBExists(_ context.Context, dbName string, _ map[string]interface{}) (bool, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	_, ok := c.dbs[dbName]
	return ok, nil
}

func (c *client) CreateDB(ctx context.Context, dbName string, options map[string]interface{}) error {
	if exists, _ := c.DBExists(ctx, dbName, options); exists {
		return errors.Status(http.StatusPreconditionFailed, "database exists")
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.dbs[dbName] = &database{}
	return nil
}

func (c *client) DestroyDB(ctx context.Context, dbName string, options map[string]interface{}) error {
	if exists, _ := c.DBExists(ctx, dbName, options); !exists {
		return errors.Status(http.StatusNotFound, "database does not exist")
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.dbs, dbName)
	return nil
}

func (c *client) DB(_ context.Context, dbName string, options map[string]interface{}) (driver.DB, error) {
	return &db{
		client: c,
		dbName: dbName,
	}, nil
}
