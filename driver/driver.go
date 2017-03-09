package driver

import (
	"encoding/json"
	"time"
)

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
	// VendorVersion should return the vendor version number.
	VendorVersion() string
}

// Client is a connection to a database server.
type Client interface {
	// VersionInfo returns the server implementation's details.
	ServerInfo() (ServerInfo, error)
	AllDBs() ([]string, error)
	// DBExists returns true if the database exists.
	DBExists(dbName string) (bool, error)
	// CreateDB creates the requested DB. The dbName is validated as a valid
	// CouchDB database name prior to calling this function, so the driver can
	// assume a valid name.
	CreateDB(dbName string) error
	// DestroyDB deletes the requested DB.
	DestroyDB(dbName string) error
	// DB returns a handleto the requested database
	DB(dbName string) (DB, error)
}

// Authenticator is an optional interface that may be implemented by a Client
// that supports authenitcated connections.
type Authenticator interface {
	// Authenticate attempts to authenticate the client using an authenticator.
	// If the authenticator is not known to the client, an error should be
	// returned.
	Authenticate(authenticator interface{}) error
}

// UUIDer is an optional interface that may be implemented by a Client. Generally,
// this should not be used, but it is part of the CouchDB spec, so it is included
// for completeness.
type UUIDer interface {
	UUIDs(count int) ([]string, error)
}

// Logger is an optional interface that may be implemented by a Client. When
// implemented, the method should fill the passed buf []byte array with the most
// recent server logs. If offset is present, offset bytes should be skipped at
// the end of the log. The return value is the number of bytes read, up to
// len(buf).
type Logger interface {
	Log(buf []byte, offset int) (int, error)
}

// Cluster is an optional interface that may be implemented by a Client for
// servers that support clustering operations (specifically CouchDB 2.0)
type Cluster interface {
	Membership() (allNodes []string, clusterNodes []string, err error)
}

// DB is a database handle.
type DB interface {
	AllDocs(docs interface{}, options map[string]interface{}) (offset, totalrows int, seq string, err error)
	// BulkDocs()
	// Get fetches the requested document from the database, and unmarshals it
	// into doc.
	Get(docID string, doc interface{}, options map[string]interface{}) error
	// CreateDoc creates a new doc, with a server-generated ID.
	CreateDoc(doc interface{}) (docID, rev string, err error)
	// Put writes the document in the database.
	Put(docID string, doc interface{}) (rev string, err error)
	// Delete marks the specified document as deleted.
	Delete(docID, rev string) (newRev string, err error)
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

// DBFlusher is an optional interface that may be implemented by a database
// that can force a flush of the database backend file(s) to disk or other
// permanent storage.
type DBFlusher interface {
	// Flush requests a flush of disk cache to disk or other permanent storage.
	// The response a timestamp when the database backend opened the storage
	// backend.
	//
	// See http://docs.couchdb.org/en/2.0.0/api/database/compact.html#db-ensure-full-commit
	Flush() (time.Time, error)
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
	Config() (Config, error)
}

// Config is the minimal interface that a Config backend must implement.
type Config interface {
	GetAll() (config map[string]map[string]string, err error)
	Set(secName, key, value string) error
	Delete(secName, key string) error
}

// ConfigSection is an optional interface that may be implemented by a Config
// backend. If not implemented, it will be emulated with GetAll() and SetAll().
// The only reason for a config backend to implement this interface is if
// reading a config section alone can be more efficient than reading the entire
// configuration for the specific storage backend.
type ConfigSection interface {
	GetSection(secName string) (section map[string]string, err error)
}

// ConfigItem is an optional interface that may be implemented by a Config
// backend. If not implemented, it will be emulated with GetAll() and SetAll().
// The only reason for a config backend to implement this interface is if
// reading a single config value alone can be more efficient than reading the
// entire configuration for the specific storage backend.
type ConfigItem interface {
	Get(secName, key string) (value string, err error)
}
