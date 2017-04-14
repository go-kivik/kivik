package kivik

import (
	"context"

	"github.com/flimzy/kivik/driver"
)

// Find calls FindContext with a background context.
func (db *DB) Find(query interface{}) (*Rows, error) {
	return db.FindContext(context.Background(), query)
}

// FindContext executes a query using the new /_find interface. The query must
//  be JSON-marshalable to a valid query.
// See http://docs.couchdb.org/en/2.0.0/api/database/find.html#db-find
func (db *DB) FindContext(ctx context.Context, query interface{}) (*Rows, error) {
	if finder, ok := db.driverDB.(driver.Finder); ok {
		rowsi, err := finder.FindContext(ctx, query)
		if err != nil {
			return nil, err
		}
		rows := &Rows{rowsi: rowsi}
		rows.initContextClose(ctx)
		return rows, nil
	}
	return nil, ErrNotImplemented
}

// CreateIndex calls CreateIndexContext with a background context.
func (db *DB) CreateIndex(ddoc, name string, index interface{}) error {
	return db.CreateIndexContext(context.Background(), ddoc, name, index)
}

// CreateIndexContext creates an index if it doesn't already exist. ddoc and
// name may be empty, in which case they will be auto-generated.  index must be
// a valid index object, as described here:
// http://docs.couchdb.org/en/2.0.0/api/database/find.html#find-sort
func (db *DB) CreateIndexContext(ctx context.Context, ddoc, name string, index interface{}) error {
	if finder, ok := db.driverDB.(driver.Finder); ok {
		return finder.CreateIndexContext(ctx, ddoc, name, index)
	}
	return ErrNotImplemented
}

// DeleteIndex calls DeleteIndexContext with a background context.
func (db *DB) DeleteIndex(ddoc, name string) error {
	return db.DeleteIndexContext(context.Background(), ddoc, name)
}

// DeleteIndexContext deletes the requested index.
func (db *DB) DeleteIndexContext(ctx context.Context, ddoc, name string) error {
	if finder, ok := db.driverDB.(driver.Finder); ok {
		return finder.DeleteIndexContext(ctx, ddoc, name)
	}
	return ErrNotImplemented
}
