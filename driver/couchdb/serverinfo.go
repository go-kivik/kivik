package couchdb

import (
	"context"
	"encoding/json"

	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/driver/couchdb/chttp"
)

// ServerInfo returns the server's version info.
func (c *client) ServerInfoContext(ctx context.Context) (driver.ServerInfo, error) {
	i := &info{}
	_, err := c.DoJSON(ctx, chttp.MethodGet, "/", nil, i)
	return i, err
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
