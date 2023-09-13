// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

package kivik

import (
	"context"
	"net/http"

	"github.com/go-kivik/kivik/v4/driver"
)

var findNotImplemented = &Error{Status: http.StatusNotImplemented, Message: "kivik: driver does not support Find interface"}

// Find executes a query using the new /_find interface. The query must be
// JSON-marshalable to a valid query.
//
// See the [CouchDB documentation].
//
// [CouchDB documentation]: https://docs.couchdb.org/en/stable/api/database/find.html
func (db *DB) Find(ctx context.Context, query interface{}, options ...Option) *ResultSet {
	if db.err != nil {
		return &ResultSet{err: db.err}
	}
	if finder, ok := db.driverDB.(driver.Finder); ok {
		if err := db.startQuery(); err != nil {
			return &ResultSet{err: err}
		}
		rowsi, err := finder.Find(ctx, query, allOptions(options))
		if err != nil {
			db.endQuery()
			return &ResultSet{err: err}
		}
		return &ResultSet{underlying: newRows(ctx, db.endQuery, rowsi)}
	}
	return &ResultSet{err: findNotImplemented}
}

// CreateIndex creates an index if it doesn't already exist. ddoc and name may
// be empty, in which case they will be auto-generated.  index must be
// marshalable to a valid index object, as described in the [CouchDB documentation].
//
// [CouchDB documentation]: http://docs.couchdb.org/en/stable/api/database/find.html#db-index
func (db *DB) CreateIndex(ctx context.Context, ddoc, name string, index interface{}, options ...Option) error {
	if db.err != nil {
		return db.err
	}
	if err := db.startQuery(); err != nil {
		return err
	}
	defer db.endQuery()
	if finder, ok := db.driverDB.(driver.Finder); ok {
		return finder.CreateIndex(ctx, ddoc, name, index, allOptions(options))
	}
	return findNotImplemented
}

// DeleteIndex deletes the requested index.
func (db *DB) DeleteIndex(ctx context.Context, ddoc, name string, options ...Option) error {
	if db.err != nil {
		return db.err
	}
	if err := db.startQuery(); err != nil {
		return err
	}
	defer db.endQuery()
	if finder, ok := db.driverDB.(driver.Finder); ok {
		return finder.DeleteIndex(ctx, ddoc, name, allOptions(options))
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
func (db *DB) GetIndexes(ctx context.Context, options ...Option) ([]Index, error) {
	if db.err != nil {
		return nil, db.err
	}
	if err := db.startQuery(); err != nil {
		return nil, err
	}
	defer db.endQuery()
	if finder, ok := db.driverDB.(driver.Finder); ok {
		dIndexes, err := finder.GetIndexes(ctx, allOptions(options))
		indexes := make([]Index, len(dIndexes))
		for i, index := range dIndexes {
			indexes[i] = Index(index)
		}
		return indexes, err
	}
	return nil, findNotImplemented
}

// QueryPlan is the query execution plan for a query, as returned by
// [DB.Explain].
type QueryPlan struct {
	DBName   string                 `json:"dbname"`
	Index    map[string]interface{} `json:"index"`
	Selector map[string]interface{} `json:"selector"`
	Options  map[string]interface{} `json:"opts"`
	Limit    int64                  `json:"limit"`
	Skip     int64                  `json:"skip"`

	// Fields is the list of fields to be returned in the result set, or
	// an empty list if all fields are to be returned.
	Fields []interface{}          `json:"fields"`
	Range  map[string]interface{} `json:"range"`
}

// Explain returns the query plan for a given query. Explain takes the same
// arguments as [DB.Find].
func (db *DB) Explain(ctx context.Context, query interface{}, options ...Option) (*QueryPlan, error) {
	if db.err != nil {
		return nil, db.err
	}
	if err := db.startQuery(); err != nil {
		return nil, err
	}
	defer db.endQuery()
	if explainer, ok := db.driverDB.(driver.Finder); ok {
		plan, err := explainer.Explain(ctx, query, allOptions(options))
		if err != nil {
			return nil, err
		}
		qp := QueryPlan(*plan)
		return &qp, nil
	}
	return nil, findNotImplemented
}
