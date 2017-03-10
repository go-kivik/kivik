// Package memory provides a memory-backed Kivik driver, intended for testing.
package memory

import (
	"context"
	"net/http"
	"sync"

	"github.com/pborman/uuid"

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

func (d *memDriver) NewClient(name string) (driver.Client, error) {
	return &client{
		Client: common.NewClient(Version, Vendor, Version),
		dbs:    make(map[string]*database),
	}, nil
}

func (c *client) AllDBsContext(_ context.Context) ([]string, error) {
	dbs := make([]string, 0, len(c.dbs))
	for k := range c.dbs {
		dbs = append(dbs, k)
	}
	return dbs, nil
}

func (c *client) UUIDsContext(_ context.Context, count int) ([]string, error) {
	uuids := make([]string, count)
	for i := 0; i < count; i++ {
		uuids[i] = uuid.New()
	}
	return uuids, nil
}

func (c *client) DBExistsContext(_ context.Context, dbName string) (bool, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	_, ok := c.dbs[dbName]
	return ok, nil
}

func (c *client) CreateDBContext(ctx context.Context, dbName string) error {
	if exists, _ := c.DBExistsContext(ctx, dbName); exists {
		return errors.Status(http.StatusPreconditionFailed, "database exists")
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.dbs[dbName] = &database{}
	return nil
}

func (c *client) DestroyDBContext(ctx context.Context, dbName string) error {
	if exists, _ := c.DBExistsContext(ctx, dbName); !exists {
		return errors.Status(http.StatusNotFound, "database not found")
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.dbs, dbName)
	return nil
}

func (c *client) DBContext(_ context.Context, dbName string) (driver.DB, error) {
	return &db{
		client: c,
		dbName: dbName,
	}, nil
}
