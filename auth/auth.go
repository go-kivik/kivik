// Package auth provides a common API for communicating with CouchDB
// authentication services. This package is used for implementing CouchDB-like
// servers, and is not used for standard client access to CouchDB.
package auth

import "github.com/flimzy/kivik/driver"

// Client represents a client connection to a user authentication backend.
type Client struct {
	driverClient driver.UserStore
}

func New(driverName, datasource string) (*Client, error) {

}
