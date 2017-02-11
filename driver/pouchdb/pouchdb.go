package pouchdb

import (
	"encoding/json"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/driver/pouchdb/bindings"
)

type pouchDriver struct{}

var _ driver.Driver = &pouchDriver{}

func init() {
	kivik.Register("pouchdb", &pouchDriver{})
}

// NewClient returns a PouchDB client handle. The connection string is ignored.
func (d *pouchDriver) NewClient(_ string) (driver.Client, error) {
	return &pouchClient{}, nil
}

type pouchClient struct{}

var _ driver.Client = &pouchClient{}

// AllDBs returns the list of all existing databases. This function depends on
// the pouchdb-all-dbs plugin being loaded.
func (c *pouchClient) AllDBs() ([]string, error) {
	return bindings.AllDBs()
}

type pouchInfo struct{}

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
func (i *pouchInfo) Version() string { return bindings.Version() }

func (c *pouchClient) ServerInfo() (driver.ServerInfo, error) {
	return &pouchInfo{}, nil
}
