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

//go:build js
// +build js

package pouchdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal"
	"github.com/go-kivik/kivik/v4/pouchdb/bindings"
)

type pouchDriver struct{}

var _ driver.Driver = &pouchDriver{}

func init() {
	kivik.Register("pouch", &pouchDriver{})
}

// NewClient returns a PouchDB client handle. Provide a dsn only for remote
// databases. Otherwise specify ""
func (d *pouchDriver) NewClient(dsn string, options driver.Options) (driver.Client, error) {
	var u *url.URL
	var user *url.Userinfo
	if dsn != "" {
		var err error
		u, err = url.Parse(dsn)
		if err != nil {
			return nil, fmt.Errorf("Invalid DSN URL '%s' provided: %s", dsn, err)
		}
		user = u.User
		u.User = nil
	}
	pouch := bindings.GlobalPouchDB()
	client := &client{
		dsn:   u,
		pouch: pouch,
		opts:  make(map[string]Options),
	}
	if user != nil {
		pass, _ := user.Password()
		client.setAuth(user.Username(), pass)
	}
	options.Apply(client)
	return client, nil
}

func (c *client) setAuth(username, password string) {
	c.opts["authenticator"] = Options{
		"auth": map[string]interface{}{
			"username": username,
			"password": password,
		},
	}
}

type client struct {
	dsn   *url.URL
	opts  map[string]Options
	pouch *bindings.PouchDB

	// This mantains a list of running replications
	replications   []*replication
	replicationsMU sync.RWMutex
}

var _ driver.Client = &client{}

// AllDBs returns the list of all existing databases. This function depends on
// the pouchdb-all-dbs plugin being loaded.
func (c *client) AllDBs(ctx context.Context, _ driver.Options) ([]string, error) {
	if c.dsn == nil {
		return c.pouch.AllDBs(ctx)
	}
	return nil, errors.New("AllDBs() not implemented for remote PouchDB databases")
}

func (c *client) Version(context.Context) (*driver.Version, error) {
	ver := c.pouch.Version()
	return &driver.Version{
		Version:     ver,
		Vendor:      "PouchDB",
		RawResponse: json.RawMessage(fmt.Sprintf(`{"version":"%s","vendor":{"name":"PouchDB"}}`, ver)),
	}, nil
}

func (c *client) dbURL(db string) string {
	if c.dsn == nil {
		// No transformation for local databases
		return db
	}
	myURL := *c.dsn // Make a copy
	myURL.Path += strings.TrimLeft(db, "/")
	return myURL.String()
}

// Options is a struct of options, as documented in the PouchDB API.
type Options map[string]interface{}

func (c *client) options(options ...Options) Options {
	o := Options{}
	for _, defOpts := range c.opts {
		for k, v := range defOpts {
			o[k] = v
		}
	}
	for _, opts := range options {
		for k, v := range opts {
			o[k] = v
		}
	}
	return o
}

func (c *client) isRemote() bool {
	return c.dsn != nil
}

// DBExists returns true if the requested DB exists. This function only works
// for remote databases. For local databases, it creates the database.
// Silly PouchDB.
func (c *client) DBExists(ctx context.Context, dbName string, options driver.Options) (bool, error) {
	pouchOpts := map[string]interface{}{"skip_setup": true}
	options.Apply(pouchOpts)
	_, err := c.pouch.New(c.dbURL(dbName), pouchOpts).Info(ctx)
	if err == nil {
		return true, nil
	}
	if kivik.HTTPStatus(err) == http.StatusNotFound {
		return false, nil
	}
	return false, err
}

func (c *client) CreateDB(ctx context.Context, dbName string, options driver.Options) error {
	if c.isRemote() {
		if exists, _ := c.DBExists(ctx, dbName, options); exists {
			return &internal.Error{Status: http.StatusPreconditionFailed, Message: "database exists"}
		}
	}
	pouchOpts := map[string]interface{}{}
	options.Apply(pouchOpts)
	_, err := c.pouch.New(c.dbURL(dbName), pouchOpts).Info(ctx)
	return err
}

func (c *client) DestroyDB(ctx context.Context, dbName string, options driver.Options) error {
	exists, err := c.DBExists(ctx, dbName, options)
	if err != nil {
		return err
	}
	if !exists {
		// This will only ever do anything for a remote database
		return &internal.Error{Status: http.StatusNotFound, Message: "database does not exist"}
	}
	pouchOpts := map[string]interface{}{}
	options.Apply(pouchOpts)
	return c.pouch.New(c.dbURL(dbName), pouchOpts).Destroy(ctx, nil)
}

func (c *client) DB(dbName string, options driver.Options) (driver.DB, error) {
	pouchOpts := map[string]interface{}{}
	options.Apply(pouchOpts)
	return &db{
		// TODO: #68 Consider deferring this pouch.New call until the first use,
		// so ctx can be used.
		db:     c.pouch.New(c.dbURL(dbName), pouchOpts),
		client: c,
	}, nil
}
