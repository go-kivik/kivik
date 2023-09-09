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
	"github.com/go-kivik/kivik/v4/pouchdb/bindings"
)

type pouchDriver struct{}

var _ driver.Driver = &pouchDriver{}

func init() {
	kivik.Register("pouch", &pouchDriver{})
}

// NewClient returns a PouchDB client handle. Provide a dsn only for remote
// databases. Otherwise specify ""
func (d *pouchDriver) NewClient(dsn string, _ map[string]interface{}) (driver.Client, error) {
	var u *url.URL
	var auth authenticator
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
		auth = &BasicAuth{
			Name:     user.Username(),
			Password: pass,
		}
		if err := auth.authenticate(client); err != nil {
			return nil, err
		}
	}
	return client, nil
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
func (c *client) AllDBs(ctx context.Context, _ map[string]interface{}) ([]string, error) {
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
func (c *client) DBExists(ctx context.Context, dbName string, opts map[string]interface{}) (bool, error) {
	pouchOpts := c.options(opts, Options{"skip_setup": true})
	_, err := c.pouch.New(c.dbURL(dbName), pouchOpts).Info(ctx)
	if err == nil {
		return true, nil
	}
	if kivik.HTTPStatus(err) == http.StatusNotFound {
		return false, nil
	}
	return false, err
}

func (c *client) CreateDB(ctx context.Context, dbName string, opts map[string]interface{}) error {
	if c.isRemote() {
		if exists, _ := c.DBExists(ctx, dbName, opts); exists {
			return &kivik.Error{Status: http.StatusPreconditionFailed, Message: "database exists"}
		}
	}
	pouchOpts := c.options(opts)
	_, err := c.pouch.New(c.dbURL(dbName), pouchOpts).Info(ctx)
	return err
}

func (c *client) DestroyDB(ctx context.Context, dbName string, opts map[string]interface{}) error {
	pouchOpts := c.options(opts)
	exists, err := c.DBExists(ctx, dbName, pouchOpts)
	if err != nil {
		return err
	}
	if !exists {
		// This will only ever do anything for a remote database
		return &kivik.Error{Status: http.StatusNotFound, Message: "database does not exist"}
	}
	return c.pouch.New(c.dbURL(dbName), pouchOpts).Destroy(ctx, nil)
}

func (c *client) DB(dbName string, opts map[string]interface{}) (driver.DB, error) {
	pouchOpts := c.options(opts)
	return &db{
		// TODO: #68 Consider deferring this pouch.New call until the first use,
		// so ctx can be used.
		db:     c.pouch.New(c.dbURL(dbName), pouchOpts),
		client: c,
	}, nil
}
