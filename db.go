package kivik

// DB is a handle to a specific database.
type DB struct{}

// AllDocs returns a list of all documents in the database.
func (db *DB) AllDocs() error {
	return nil
}
