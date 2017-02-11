// Package memory provides a memory-backed Kivik driver, intended for testing.
package memory

import (
	"encoding/json"
	"net/http"

	"github.com/pborman/uuid"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver"
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
	dbs map[string]database
}

var _ driver.Client = &client{}

func (d *memDriver) NewClient(name string) (driver.Client, error) {
	return &client{
		dbs: make(map[string]database),
	}, nil
}

type serverInfo struct {
}

var _ driver.ServerInfo = &serverInfo{}

func (i *serverInfo) Response() json.RawMessage {
	return []byte(`{"couchdb":"Welcome","uuid":"a176f89954c3ddba7aa592d712c25140","version":"0.0.1","vendor":{"name":"Kivik Memory Adaptor","version":"0.0.1"}}`)
}

func (i *serverInfo) Vendor() string {
	return "Kivik"
}

func (i *serverInfo) Version() string {
	return "0.0.1"
}

// ServerInfo returns the server info for this driver.
func (c *client) ServerInfo() (driver.ServerInfo, error) {
	return &serverInfo{}, nil
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
