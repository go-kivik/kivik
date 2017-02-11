package pouchdb

import (
	"encoding/json"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/driver/pouchdb/bindings"
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

// NewClient returns a PouchDB client handle. The connection string is ignored.
func (d *Driver) NewClient(_ string) (driver.Client, error) {
	var pouch *bindings.PouchDB
	if d.Defaults == nil {
		pouch = bindings.GlobalPouchDB()
	} else {
		pouch = bindings.Defaults(d.Defaults)
	}
	return &client{
		pouch: pouch,
	}, nil
}

type client struct {
	pouch *bindings.PouchDB
}

var _ driver.Client = &client{}

// AllDBs returns the list of all existing databases. This function depends on
// the pouchdb-all-dbs plugin being loaded.
func (c *client) AllDBs() ([]string, error) {
	return c.pouch.AllDBs()
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
func (i *pouchInfo) Vendor() string  { return "PouchDB" }
func (i *pouchInfo) Version() string { return i.vers }

func (c *client) ServerInfo() (driver.ServerInfo, error) {
	return &pouchInfo{
		vers: c.pouch.Version(),
	}, nil
}

// DBExists returns true if the requested DB exists.
func (c *client) DBExists(dbName string) (bool, error) {
	info, err := c.pouch.New(dbName, map[string]interface{}{"skip_setup": true}).Info()
	spew.Dump(info)
	fmt.Printf("err = %s\n", err)
	return err == nil, err
}
