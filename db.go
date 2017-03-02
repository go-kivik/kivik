package kivik

import "github.com/flimzy/kivik/driver"

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
	return db.driverDB.Put(docID, doc)
}
