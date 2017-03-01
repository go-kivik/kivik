// Package memory provides a memory-backed Kivik driver, intended for testing.
package memory

import (
	"net/http"

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

// database is an in-memory database representation.
type database struct{}

type client struct {
	*common.Client
	dbs map[string]database
}

var _ driver.Client = &client{}

func (d *memDriver) NewClient(name string) (driver.Client, error) {
	return &client{
		Client: common.NewClient("0.0.1", "Kivik Memory Adaptor", "0.0.1"),
		dbs:    make(map[string]database),
	}, nil
}

func (c *client) AllDBs() ([]string, error) {
	dbs := make([]string, 0, len(c.dbs))
	for k := range c.dbs {
		dbs = append(dbs, k)
	}
	return dbs, nil
}

func (c *client) UUIDs(count int) ([]string, error) {
	uuids := make([]string, count)
	for i := 0; i < count; i++ {
		uuids[i] = uuid.New()
	}
	return uuids, nil
}

func (c *client) DBExists(dbName string) (bool, error) {
	_, ok := c.dbs[dbName]
	return ok, nil
}

func (c *client) CreateDB(dbName string) error {
	if exists, _ := c.DBExists(dbName); exists {
		return errors.Status(http.StatusPreconditionFailed, "database exists")
	}
	c.dbs[dbName] = database{}
	return nil
}

func (c *client) DestroyDB(dbName string) error {
	if exists, _ := c.DBExists(dbName); !exists {
		return errors.Status(http.StatusNotFound, "database not found")
	}
	delete(c.dbs, dbName)
	return nil
}

func (c *client) DB(dbName string) (driver.DB, error) {
	return &db{
		client: c,
		dbName: dbName,
	}, nil
}
