package pouchdb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/driver/pouchdb/bindings"
	"github.com/flimzy/kivik/errors"
	"github.com/gopherjs/gopherjs/js"
)

// Driver represents the configuration for a PouchDB driver. You may specify
// custom configuration PouchDB by registering your own instance of this driver.
// For example, to use the memdown driver (https://github.com/level/memdown) in
// Node.js for unit tests:
//
//     func init() {
//         kivik.Register("memdown", pouchdb.Driver{
//             Defaults: map[string]interface{}{
//                 "db": js.Global.Call("require", "memdown"),
//             },
//         })
//    }
//
//    func main() {
//        db := kivik.NewClient("memdown", "")
//        // ...
//    }
//
type Driver struct {
	// Options is a map of default options to pass to the PouchDB constructor.
	// See https://pouchdb.com/api.html#defaults
	Defaults map[string]interface{}
}

var _ driver.Driver = &Driver{}

func init() {
	kivik.Register("pouch", &Driver{})
}

// NewClient returns a PouchDB client handle. Provide a dsn only for remote
// databases. Otherwise specify ""
func (d *Driver) NewClient(dsn string) (driver.Client, error) {
	var u *url.URL
	var auth Authenticator
	if dsn != "" {
		var err error
		u, err = url.Parse(dsn)
		if err != nil {
			return nil, fmt.Errorf("Invalid DSN URL '%s' provided: %s", dsn, err)
		}
		if u.User != nil {
			pass, _ := u.User.Password()
			auth = &BasicAuth{
				Name:     u.User.Username(),
				Password: pass,
			}
		}
		u.User = nil
	}
	var pouch *bindings.PouchDB
	if d.Defaults == nil {
		pouch = bindings.GlobalPouchDB()
	} else {
		pouch = bindings.Defaults(d.Defaults)
	}
	return &client{
		dsn:   u,
		pouch: pouch,
		auth:  auth,
	}, nil
}

type client struct {
	dsn   *url.URL
	auth  Authenticator
	pouch *bindings.PouchDB
}

var _ driver.Client = &client{}

// AllDBs returns the list of all existing databases. This function depends on
// the pouchdb-all-dbs plugin being loaded.
func (c *client) AllDBs() ([]string, error) {
	if c.dsn == nil {
		return c.pouch.AllDBs()
	}
	return nil, errors.New("AllDBs() not implemented for remote PouchDB databases")
}

type pouchInfo struct {
	vers string
}

var _ driver.ServerInfo = &pouchInfo{}

func (i *pouchInfo) Response() json.RawMessage {
	data, _ := json.Marshal(map[string]interface{}{
		"couchdb": "Welcome",
		"version": i.Version(),
		"vendor": map[string]interface{}{
			"name":    i.Vendor(),
			"version": i.Version(),
		},
	})
	return data
}
func (i *pouchInfo) Vendor() string        { return "PouchDB" }
func (i *pouchInfo) Version() string       { return i.vers }
func (i *pouchInfo) VendorVersion() string { return i.vers }

func (c *client) ServerInfo() (driver.ServerInfo, error) {
	return &pouchInfo{
		vers: c.pouch.Version(),
	}, nil
}

func (c *client) dbURL(db string) string {
	if c.dsn == nil {
		// No transformation for local databases
		return db
	}
	myURL := *c.dsn // Make a copy
	myURL.Path = myURL.Path + strings.TrimLeft(db, "/")
	return myURL.String()
}

// Options is a struct of options, as documented in the PouchDB API.
type Options map[string]interface{}

func (c *client) options(opts Options) (Options, error) {
	if c.auth != nil {
		if err := c.auth.Authenticate(opts); err != nil {
			return nil, err
		}
	}
	return opts, nil
}

// DBExists returns true if the requested DB exists. This function only works
// for remote databases. For local databases, it creates the database. Silly
// PouchDB.
func (c *client) DBExists(dbName string) (bool, error) {
	opts, err := c.options(Options{
		"skip_setup": true,
	})
	if err != nil {
		return false, err
	}
	_, err = c.pouch.New(c.dbURL(dbName), opts).Info()
	if err == nil {
		return true, nil
	}
	if jsObs, ok := err.(*js.Error); ok {
		if jsObs.Get("status").Int() == http.StatusNotFound {
			return false, nil
		}
	}
	return false, err
}

func (c *client) CreateDB(dbName string) error {
	opts, err := c.options(Options{})
	if err != nil {
		return err
	}
	_, err = c.pouch.New(c.dbURL(dbName), opts).Info()
	return err
}

func (c *client) DestroyDB(dbName string) error {
	exists, err := c.DBExists(dbName)
	if err != nil {
		return err
	}
	if !exists {
		// This will only ever do anything for a remote database
		return errors.Status(http.StatusNotFound, "database does not exist")
	}
	opts, err := c.options(Options{})
	if err != nil {
		return err
	}
	return c.pouch.New(c.dbURL(dbName), opts).Destroy(nil)
}
