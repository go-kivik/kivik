package driver

import "encoding/json"

// Driver is the interface that must be implemented by a database driver.
type Driver interface {
	// NewClient returns a connection handle to the database. The name is in a
	// driver-specific format.
	NewClient(name string) (Client, error)
}

// ServerInfo represents the response a server gives witha GET request to '/'.
type ServerInfo interface {
	// Response is the full response, unparsed.
	Response() json.RawMessage
	// Version should return the version string from the top level of the response.
	Version() string
	// Vendor should return the name of the vendor.
	Vendor() string
}

// Client is a connection to a database server.
type Client interface {
	// VersionInfo returns the server implementation's details.
	ServerInfo() (ServerInfo, error)
	AllDBs() ([]string, error)
	// OpenDB(name string) (DB, error)
}

// UUIDer is an optional interface that may be implemented by a Client. Generally,
// this should not be used, but it is part of the CouchDB spec, so it is included
// for completeness.
type UUIDer interface {
	UUIDs(count int) ([]string, error)
}

// DB is a database handle.
type DB interface {
	// AllDocs()
	// BulkDocs()
	// Get()
	// GetAttachment()
	// Compact()
	// CompactDDoc(ddoc string)
	// ViewCleanup()
	// GetSecurity()
	// SetSecurity()
	// // TempView() // No longer supported in 2.0
	// Purge()       // Optional?
	// MissingRevs() // Optional?
	// RevsDiff()    // Optional?
	// RevsLimit()   // Optional?
	// Changes()
	// // Close invalidates and potentially stops any pending queries.
	// Close() error
}

type DBFlusher interface {
	// Flush() // _ensure_full_commit
}

// Header is an optional interface that a DB may implement. If it is not
// impemented, Get() will be used instead.
type Header interface {
	// Head()
}

// Copier is an optional interface that may be implemented by a DB.
//
// If a DB does implement Copier, Copy() functions will use it.  If a DB does
// not implement the Copier interface, or if a call to Copy() returns an
// http.StatusUnimplemented, the driver will emulate a copy by doing
// a GET followed by PUT.
type Copier interface {
	// Copy() error
}

// Sessioner is an optional interface that may be implemented by a Driver.
//
// If Sessioner is not implemented, authenticated connections will not be
// possible.
type Sessioner interface {
	// Authenticate(user, pass string) (Session, error)
	// GET
	// DELETE
}

// Session is an interface for an authentication session.
type Session interface {
	// Name() string
	// Cookie() string
	// Roles() []string
	// OpenDB(name string) (DB, error)
}

// Configer is an optional interface that may be implemented by a Client.
//
// If a Client does implement Configer, it allows backend configuration
// to be queried and modified via the API.
type Configer interface {
	// GetConfig() (map[string]interface{}, error)
	// SetConfig(key string, value interface{}) error
	// ClearConfig(key string) error
}
