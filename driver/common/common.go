// Package common contains logic and data structures shared by various kivik
// backends. It is not meant to be used directly.
package common

import (
	"context"
	"encoding/json"

	"github.com/flimzy/kivik/driver"
)

// Client is a shared core between Kivik backends.
type Client struct {
	serverInfo driver.ServerInfo
}

// NewClient returns a new client core.
func NewClient(version, vendorName, vendorVersion string) *Client {
	return &Client{
		serverInfo: &serverInfo{
			version:       version,
			vendorName:    vendorName,
			vendorVersion: vendorVersion,
		},
	}
}

// ServerInfoContext returns the configured server info.
func (c *Client) ServerInfoContext(_ context.Context, _ map[string]interface{}) (driver.ServerInfo, error) {
	return c.serverInfo, nil
}

// serverInfo is a simple implementation of the driver.ServerInfo interface.
type serverInfo struct {
	version       string
	vendorName    string
	vendorVersion string
}

var _ driver.ServerInfo = &serverInfo{}

// Vendor returns the vendor name string
func (i *serverInfo) Vendor() string {
	return i.vendorName
}

// Version returns the version string
func (i *serverInfo) Version() string {
	return i.version
}

// VendorVersion returns the vendor version string
func (i *serverInfo) VendorVersion() string {
	return i.vendorVersion
}

// Response returns a full JSON response
func (i *serverInfo) Response() json.RawMessage {
	data, _ := json.Marshal(map[string]interface{}{
		"couchdb": "Welcome",
		"version": i.Version(),
		"vendor": map[string]interface{}{
			"name":    i.Vendor(),
			"version": i.VendorVersion(),
		},
	})
	return data
}
