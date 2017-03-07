package kivik

import (
	"strings"
	"time"

	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/errors"
)

// DB is a handle to a specific database.
type DB struct {
	driverDB driver.DB
}

// AllDocs returns a list of all documents in the database.
func (db *DB) AllDocs(docs interface{}, options Options) (offset, totalrows int, seq string, err error) {
	return db.driverDB.AllDocs(docs, options)
}

// Get fetches the requested document.
func (db *DB) Get(docID string, doc interface{}, options Options) error {
	return db.driverDB.Get(docID, doc, options)
}

// CreateDoc creates a new doc with an auto-generated unique ID. The generated
// docID and new rev are returned.
func (db *DB) CreateDoc(doc interface{}) (docID, rev string, err error) {
	return db.driverDB.CreateDoc(doc)
}

// Put creates a new doc or updates an existing one, with the specified docID.
// If the document already exists, the current revision must be included in doc,
// with JSON key '_rev', otherwise a conflict will occur. The new rev is
// returned.
func (db *DB) Put(docID string, doc interface{}) (rev string, err error) {
	// The '/' char is only permitted in the case of '_design/', so check that here
	if designDoc := strings.TrimPrefix(docID, "_design/"); strings.Contains(designDoc, "/") {
		return "", errors.Status(errors.StatusBadRequest, "invalid document ID")
	}
	return db.driverDB.Put(docID, doc)
}

// Flush requests a flush of disk cache to disk or other permanent storage.
// The response a timestamp when the database backend opened the storage
// backend.
//
// See http://docs.couchdb.org/en/2.0.0/api/database/compact.html#db-ensure-full-commit
func (db *DB) Flush() (time.Time, error) {
	if flusher, ok := db.driverDB.(driver.DBFlusher); ok {
		return flusher.Flush()
	}
	return time.Time{}, ErrNotImplemented
}
