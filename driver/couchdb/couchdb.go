// Package couchdb is a driver for connecting with a CouchDB server over HTTP.
package couchdb

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver"
)

type couch struct{}

var _ driver.Driver = &couch{}

func init() {
	kivik.Register("couch", &couch{})
}

type server struct {
	url string
}

var _ driver.Conn = &server{}

func (c *couch) Open(name string) (driver.Conn, error) {
	return &server{
		url: name,
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
func (s *server) ServerInfo() (driver.ServerInfo, error) {
	req, err := http.NewRequest("GET", s.url, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	req.Header.Add("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected response: %d (%s)", resp.StatusCode, resp.Status)
	}
	if resp.Header.Get("Content-Type") != "application/json" {
		return nil, fmt.Errorf("Got unexpected response content type of '%s'", resp.Header.Get("Content-Type"))
	}
	dec := json.NewDecoder(resp.Body)
	defer resp.Body.Close()
	i := &info{}
	if err := dec.Decode(&i); err != nil {
		return nil, err
	}
	return i, nil
}
