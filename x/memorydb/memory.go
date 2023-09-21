// Package memorydb provides a memory-backed Kivik driver, intended for testing.
package memorydb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sync"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
)

type memDriver struct{}

var _ driver.Driver = &memDriver{}

func init() {
	kivik.Register("memory", &memDriver{})
}

type client struct {
	version *driver.Version
	mutex   sync.RWMutex
	dbs     map[string]*database
}

var _ driver.Client = &client{}

// Identifying constants
const (
	Version = "0.0.1"
	Vendor  = "Kivik Memory Adaptor"
)

func (d *memDriver) NewClient(name string, _ driver.Options) (driver.Client, error) {
	return &client{
		version: &driver.Version{
			Version:     Version,
			Vendor:      Vendor,
			RawResponse: json.RawMessage(fmt.Sprintf(`{"version":"%s","vendor":{"name":"%s"}}`, Version, Vendor)),
		},
		dbs: make(map[string]*database),
	}, nil
}

func (c *client) AllDBs(context.Context, driver.Options) ([]string, error) {
	dbs := make([]string, 0, len(c.dbs))
	for k := range c.dbs {
		dbs = append(dbs, k)
	}
	return dbs, nil
}

func (c *client) DBExists(_ context.Context, dbName string, _ driver.Options) (bool, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	_, ok := c.dbs[dbName]
	return ok, nil
}

// Copied verbatim from http://docs.couchdb.org/en/2.0.0/api/database/common.html#head--db
var (
	validDBName = regexp.MustCompile("^[a-z][a-z0-9_$()+/-]*$")
	validNames  = map[string]struct{}{
		"_users":      {},
		"_replicator": {},
	}
)

func (c *client) CreateDB(ctx context.Context, dbName string, options driver.Options) error {
	if exists, _ := c.DBExists(ctx, dbName, options); exists {
		return statusError{status: http.StatusPreconditionFailed, error: errors.New("database exists")}
	}
	if _, ok := validNames[dbName]; !ok {
		if !validDBName.MatchString(dbName) {
			return statusError{status: http.StatusBadRequest, error: errors.New("invalid database name")}
		}
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.dbs[dbName] = &database{
		docs:     make(map[string]*document),
		security: &driver.Security{},
	}
	return nil
}

func (c *client) DestroyDB(ctx context.Context, dbName string, options driver.Options) error {
	if exists, _ := c.DBExists(ctx, dbName, options); !exists {
		return statusError{status: http.StatusNotFound, error: errors.New("database does not exist")}
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.dbs[dbName].mu.Lock()
	defer c.dbs[dbName].mu.Unlock()
	c.dbs[dbName].deleted = true // To invalidate any outstanding db handles
	delete(c.dbs, dbName)
	return nil
}

func (c *client) DB(dbName string, options driver.Options) (driver.DB, error) {
	return &db{
		client: c,
		dbName: dbName,
		db:     c.dbs[dbName],
	}, nil
}

// Version returns the configured server info.
func (c *client) Version(_ context.Context) (*driver.Version, error) {
	return c.version, nil
}
