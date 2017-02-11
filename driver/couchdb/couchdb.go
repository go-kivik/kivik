// Package couchdb is a driver for connecting with a CouchDB server over HTTP.
package couchdb

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver"
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
	baseURL  *url.URL
	authUser string
	authPass string
	client   *http.Client
}

var _ driver.Client = &client{}

func (c *client) url(path string, query url.Values) string {
	myURL := *c.baseURL // Make a copy
	myURL.Path = myURL.Path + strings.TrimLeft(path, "/")
	if query != nil {
		myURL.RawQuery = query.Encode()
	}
	return myURL.String()
}

// NewClient establishes a new connection to a CouchDB server instance. If
// auth credentials are included in the URL, they are used to make the
// connection.
func (c *Couch) NewClient(urlstring string) (driver.Client, error) {
	u, err := url.Parse(urlstring)
	if err != nil {
		return nil, err
	}
	var user, pass string
	if u.User != nil {
		user = u.User.Username()
		pass, _ = u.User.Password()
	}
	// Copy all the values, so multiple connections don't fight with each other.
	httpClient := &http.Client{
		Transport:     c.HTTPClient.Transport,
		CheckRedirect: c.HTTPClient.CheckRedirect,
		Jar:           c.HTTPClient.Jar,
		Timeout:       c.HTTPClient.Timeout,
	}
	u.RawQuery, u.Fragment = "", ""
	return &client{
		baseURL:  u,
		client:   httpClient,
		authUser: user,
		authPass: pass,
	}, nil
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

// ServerInfo returns the server's version info.
func (c *client) ServerInfo() (driver.ServerInfo, error) {
	i := &info{}
	return i, c.getJSON("/", i, nil)
}

func (c *client) AllDBs() ([]string, error) {
	var allDBs []string
	return allDBs, c.getJSON("/_all_dbs", &allDBs, nil)
}

func (c *client) UUIDs(count int) ([]string, error) {
	var uuids struct {
		UUIDs []string `json:"uuids"`
	}
	err := c.getJSON("/_uuids", &uuids, url.Values{
		"count": []string{strconv.Itoa(count)},
	})
	return uuids.UUIDs, err
}

func (c *client) Membership() ([]string, []string, error) {
	var membership struct {
		All     []string `json:"all_nodes"`
		Cluster []string `json:"cluster_nodes"`
	}
	err := c.getJSON("/_membership", &membership, nil)
	return membership.All, membership.Cluster, err
}

func (c *client) Log(buf []byte, offset int) (int, error) {
	err := c.getText("_log", buf, url.Values{
		"offset": []string{strconv.Itoa(offset)},
		"bytes":  []string{strconv.Itoa(len(buf))},
	})
	return len(buf), err
}

func (c *client) DBExists(dbName string) (bool, error) {
	err := c.head(dbName, nil)
	if StatusCode(err) == http.StatusNotFound {
		return false, nil
	}
	return err == nil, err
}
