package kivik

import (
	"context"

	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/errors"
)

var findNotImplemented = errors.Status(StatusNotImplemented, "kivik: driver does not support Find interface")

// Find executes a query using the new /_find interface. The query must be
// JSON-marshalable to a valid query.
// See http://docs.couchdb.org/en/2.0.0/api/database/find.html#db-find
func (db *DB) Find(ctx context.Context, query interface{}) (*Rows, error) {
	if finder, ok := db.driverDB.(driver.Finder); ok {
		rowsi, err := finder.Find(ctx, query)
		if err != nil {
			return nil, err
		}
		return newRows(ctx, rowsi), nil
	}
	return nil, findNotImplemented
}

// CreateIndex creates an index if it doesn't already exist. ddoc and name may
// be empty, in which case they will be auto-generated.  index must be a valid
// index object, as described here:
// http://docs.couchdb.org/en/2.0.0/api/database/find.html#find-sort
func (db *DB) CreateIndex(ctx context.Context, ddoc, name string, index interface{}) error {
	if finder, ok := db.driverDB.(driver.Finder); ok {
		return finder.CreateIndex(ctx, ddoc, name, index)
	}
	return findNotImplemented
}

// DeleteIndex deletes the requested index.
func (db *DB) DeleteIndex(ctx context.Context, ddoc, name string) error {
	if finder, ok := db.driverDB.(driver.Finder); ok {
		return finder.DeleteIndex(ctx, ddoc, name)
	}
	return findNotImplemented
}

// Index is a MonboDB-style index definition.
type Index struct {
	DesignDoc  string      `json:"ddoc,omitempty"`
	Name       string      `json:"name"`
	Type       string      `json:"type"`
	Definition interface{} `json:"def"`
}

// GetIndexes returns the indexes defined on the current database.
func (db *DB) GetIndexes(ctx context.Context) ([]Index, error) {
	if finder, ok := db.driverDB.(driver.Finder); ok {
		dIndexes, err := finder.GetIndexes(ctx)
		indexes := make([]Index, len(dIndexes))
		for i, index := range dIndexes {
			indexes[i] = Index(index)
		}
		return indexes, err
	}
	return nil, findNotImplemented
}
