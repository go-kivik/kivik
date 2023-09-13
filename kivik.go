// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

package kivik

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal/registry"
)

// Client is a client connection handle to a CouchDB-like server.
type Client struct {
	dsn          string
	driverName   string
	driverClient driver.Client

	// closed will be non-0 when the client has been closed
	closed int32
	mu     sync.Mutex
	wg     sync.WaitGroup
}

// Register makes a database driver available by the provided name. If Register
// is called twice with the same name or if driver is nil, it panics.
func Register(name string, driver driver.Driver) {
	registry.Register(name, driver)
}

// New creates a new client object specified by its database driver name
// and a driver-specific data source name.
//
// The use of options is driver-specific, so consult with the documentation for
// your driver for supported options.
func New(driverName, dataSourceName string, options ...Option) (*Client, error) {
	driveri := registry.Driver(driverName)
	if driveri == nil {
		return nil, &Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("kivik: unknown driver %q (forgotten import?)", driverName)}
	}
	client, err := driveri.NewClient(dataSourceName, allOptions(options))
	if err != nil {
		return nil, err
	}
	return &Client{
		dsn:          dataSourceName,
		driverName:   driverName,
		driverClient: client,
	}, nil
}

// Driver returns the name of the driver string used to connect this client.
func (c *Client) Driver() string {
	return c.driverName
}

// DSN returns the data source name used to connect this client.
func (c *Client) DSN() string {
	return c.dsn
}

// Version represents a server version response.
type Version struct {
	// Version is the version number reported by the server or backend.
	Version string
	// Vendor is the vendor string reported by the server or backend.
	Vendor string
	// Features is a list of enabled, optional features.  This was added in
	// CouchDB 2.1.0, and can be expected to be empty for older versions.
	Features []string
	// RawResponse is the raw response body returned by the server, useful if
	// you need additional backend-specific information.  Refer to the
	// [CouchDB documentation] for format details.
	//
	// [CouchDB documentation]: http://docs.couchdb.org/en/2.0.0/api/server/common.html#get
	RawResponse json.RawMessage
}

func (c *Client) startQuery() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if atomic.LoadInt32(&c.closed) > 0 {
		return ErrClientClosed
	}
	c.wg.Add(1)
	return nil
}

func (c *Client) endQuery() {
	c.mu.Lock()
	c.wg.Done()
	c.mu.Unlock()
}

// Version returns version and vendor info about the backend.
func (c *Client) Version(ctx context.Context) (*Version, error) {
	if err := c.startQuery(); err != nil {
		return nil, err
	}
	defer c.endQuery()
	ver, err := c.driverClient.Version(ctx)
	if err != nil {
		return nil, err
	}
	v := &Version{}
	*v = Version(*ver)
	return v, nil
}

// DB returns a handle to the requested database. Any errors encountered during
// initiation of the DB object is deferred until the first method call, or may
// be checked directly with [DB.Err].
func (c *Client) DB(dbName string, options ...Option) *DB {
	db, err := c.driverClient.DB(dbName, allOptions(options))
	return &DB{
		client:   c,
		name:     dbName,
		driverDB: db,
		err:      err,
	}
}

// AllDBs returns a list of all databases.
func (c *Client) AllDBs(ctx context.Context, options ...Option) ([]string, error) {
	if err := c.startQuery(); err != nil {
		return nil, err
	}
	defer c.endQuery()
	return c.driverClient.AllDBs(ctx, allOptions(options))
}

// DBExists returns true if the specified database exists.
func (c *Client) DBExists(ctx context.Context, dbName string, options ...Option) (bool, error) {
	if err := c.startQuery(); err != nil {
		return false, err
	}
	defer c.endQuery()
	return c.driverClient.DBExists(ctx, dbName, allOptions(options))
}

// CreateDB creates a DB of the requested name.
func (c *Client) CreateDB(ctx context.Context, dbName string, options ...Option) error {
	if err := c.startQuery(); err != nil {
		return err
	}
	defer c.endQuery()
	return c.driverClient.CreateDB(ctx, dbName, allOptions(options))
}

// DestroyDB deletes the requested DB.
func (c *Client) DestroyDB(ctx context.Context, dbName string, options ...Option) error {
	if err := c.startQuery(); err != nil {
		return err
	}
	defer c.endQuery()
	return c.driverClient.DestroyDB(ctx, dbName, allOptions(options))
}

func missingArg(arg string) error {
	return &Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("kivik: %s required", arg)}
}

// DBsStats returns database statistics about one or more databases.
func (c *Client) DBsStats(ctx context.Context, dbnames []string) ([]*DBStats, error) {
	if err := c.startQuery(); err != nil {
		return nil, err
	}
	defer c.endQuery()
	dbstats, err := c.nativeDBsStats(ctx, dbnames)
	switch HTTPStatus(err) {
	case http.StatusNotFound, http.StatusNotImplemented:
		return c.fallbackDBsStats(ctx, dbnames)
	}
	return dbstats, err
}

func (c *Client) fallbackDBsStats(ctx context.Context, dbnames []string) ([]*DBStats, error) {
	dbstats := make([]*DBStats, len(dbnames))
	for i, dbname := range dbnames {
		stat, err := c.DB(dbname).Stats(ctx)
		if err != nil {
			return nil, err
		}
		dbstats[i] = stat
	}
	return dbstats, nil
}

func (c *Client) nativeDBsStats(ctx context.Context, dbnames []string) ([]*DBStats, error) {
	statser, ok := c.driverClient.(driver.DBsStatser)
	if !ok {
		return nil, &Error{Status: http.StatusNotImplemented, Message: "kivik: not supported by driver"}
	}
	stats, err := statser.DBsStats(ctx, dbnames)
	if err != nil {
		return nil, err
	}
	dbstats := make([]*DBStats, len(stats))
	for i, stat := range stats {
		dbstats[i] = driverStats2kivikStats(stat)
	}
	return dbstats, nil
}

// Ping returns true if the database is online and available for requests.
func (c *Client) Ping(ctx context.Context) (bool, error) {
	if err := c.startQuery(); err != nil {
		return false, err
	}
	defer c.endQuery()
	if pinger, ok := c.driverClient.(driver.Pinger); ok {
		return pinger.Ping(ctx)
	}
	_, err := c.driverClient.Version(ctx)
	return err == nil, err
}

// Close cleans up any resources used by the client. Close is safe to call
// concurrently with other operations and will block until all other operations
// finish. After calling Close, any other client operations will return
// [ErrClientClosed].
func (c *Client) Close() error {
	c.mu.Lock()
	atomic.StoreInt32(&c.closed, 1)
	c.mu.Unlock()
	c.wg.Wait()
	if closer, ok := c.driverClient.(driver.ClientCloser); ok {
		return closer.Close()
	}
	return nil
}
