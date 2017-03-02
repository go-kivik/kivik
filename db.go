package kivik

import "github.com/flimzy/kivik/driver"

// DB is a handle to a specific database.
type DB struct {
	driverDB driver.DB
}

// AllDocs returns a list of all documents in the database.
func (db *DB) AllDocs(docs interface{}, options Options) (offset, totalrows int, err error) {
	return db.driverDB.AllDocs(docs, options)
}

// Get fetches the requested document.
func (db *DB) Get(docID string, doc interface{}, options Options) error {
	return db.driverDB.Get(docID, doc, options)
}
