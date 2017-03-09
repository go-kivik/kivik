// Package couchdb is a driver for connecting with a CouchDB server over HTTP.
package couchdb

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/driver/couchdb/chttp"
	"github.com/flimzy/kivik/errors"
)

const (
	typeJSON  = "application/json"
	typeText  = "text/plain"
	typeMixed = "multipart/mixed"
)

// Couch represents the parent driver instance. The default driver uses a
// default http.Client. To change the timeout or other attributes, register
// a new instance with a custom HTTPClient value.
//
//    func init() {
//        kivik.Register("myCouch", &couchdb.Couch{
//            HTTPClient: &http.Client{Timeout: 15},
//        })
//    }
//    // ... then later
//    client, err := kivik("myCouch", ...)
//
type Couch struct {
	HTTPClient *http.Client
}

var _ driver.Driver = &Couch{}

func init() {
	kivik.Register("couch", &Couch{
		HTTPClient: &http.Client{},
	})
}

type client struct {
	*chttp.Client
}

var _ driver.Client = &client{}

// NewClient establishes a new connection to a CouchDB server instance. If
// auth credentials are included in the URL, they are used to authenticate using
// CookieAuth (or BasicAuth if compiled with GopherJS). If you wish to use a
// different auth mechanism, do not specify credentials here, and instead call
// Authenticate() later.
func (c *Couch) NewClient(dsn string) (driver.Client, error) {
	chttpClient, err := chttp.New(dsn)
	return &client{
		Client: chttpClient,
	}, err
}

type info struct {
	Data json.RawMessage
	Ver  string `json:"version"`
	Vend struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"vendor"`
}

var _ driver.ServerInfo = &info{}

func (i *info) UnmarshalJSON(data []byte) error {
	type alias info
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	i.Data = data
	i.Ver = a.Ver
	i.Vend = a.Vend
	return nil
}

func (i *info) Response() json.RawMessage { return i.Data }
func (i *info) Version() string           { return i.Ver }
func (i *info) Vendor() string            { return i.Vend.Name }
func (i *info) VendorVersion() string     { return i.Vend.Version }

// ServerInfo returns the server's version info.
func (c *client) ServerInfo() (driver.ServerInfo, error) {
	i := &info{}
	return i, c.DoJSON(chttp.MethodGet, "/", nil, i)
}

func (c *client) AllDBs() ([]string, error) {
	var allDBs []string
	return allDBs, c.DoJSON(chttp.MethodGet, "/_all_dbs", nil, &allDBs)
}

func (c *client) UUIDs(count int) ([]string, error) {
	var uuids struct {
		UUIDs []string `json:"uuids"`
	}
	return uuids.UUIDs, c.DoJSON(chttp.MethodGet, fmt.Sprintf("/_uuids?count=%d", count), nil, &uuids)
}

func (c *client) Membership() ([]string, []string, error) {
	var membership struct {
		All     []string `json:"all_nodes"`
		Cluster []string `json:"cluster_nodes"`
	}
	return membership.All, membership.Cluster, c.DoJSON(chttp.MethodGet, "/_membership", nil, &membership)
}

func (c *client) DBExists(dbName string) (bool, error) {
	err := c.DoError(chttp.MethodHead, dbName, nil)
	if errors.StatusCode(err) == kivik.StatusNotFound {
		return false, nil
	}
	return err == nil, err
}

func (c *client) CreateDB(dbName string) error {
	return c.DoError(chttp.MethodPut, dbName, nil)
}

func (c *client) DestroyDB(dbName string) error {
	return c.DoError(chttp.MethodDelete, dbName, nil)
}

func (c *client) DB(dbName string) (driver.DB, error) {
	return &db{
		client: c,
		dbName: dbName,
	}, nil
}

type putResponse struct {
	ID  string `json:"id"`
	OK  bool   `json:"ok"`
	Rev string `json:"rev"`
}
