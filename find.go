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
	"encoding/json"
	"net/http"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/int/errors"
)

// Find executes a query using the [_find interface]. The query must be a
// string, []byte, or [encoding/json.RawMessage] value, or JSON-marshalable to a
// valid valid query. The options are merged with the query, and will overwrite
// any values in the query.
//
// This arguments this method accepts will change in Kivik 5.x, to be more
// consistent with the rest of the Kivik API. See [issue #1014] for details.
//
// [_find interface]: https://docs.couchdb.org/en/stable/api/database/find.html
// [issue #1014]: https://github.com/go-kivik/kivik/issues/1014
func (db *DB) Find(ctx context.Context, query interface{}, options ...Option) *ResultSet {
	if db.err != nil {
		return &ResultSet{iter: errIterator(db.err)}
	}
	finder, ok := db.driverDB.(driver.Finder)
	if !ok {
		return &ResultSet{iter: errIterator(errFindNotImplemented)}
	}

	jsonQuery, err := toQuery(query, options...)
	if err != nil {
		return &ResultSet{iter: errIterator(err)}
	}

	endQuery, err := db.startQuery()
	if err != nil {
		return &ResultSet{iter: errIterator(err)}
	}
	rowsi, err := finder.Find(ctx, jsonQuery, multiOptions(options))
	if err != nil {
		endQuery()
		return &ResultSet{iter: errIterator(err)}
	}
	return newResultSet(ctx, endQuery, rowsi)
}

// toQuery combines query and options into a final JSON query to be sent to the
// driver.
func toQuery(query interface{}, options ...Option) (json.RawMessage, error) {
	var queryJSON []byte
	switch t := query.(type) {
	case string:
		queryJSON = []byte(t)
	case []byte:
		queryJSON = t
	case json.RawMessage:
		queryJSON = t
	default:
		var err error
		queryJSON, err = json.Marshal(query)
		if err != nil {
			return nil, &errors.Error{Status: http.StatusBadRequest, Err: err}
		}
	}
	var queryObject map[string]interface{}
	if err := json.Unmarshal(queryJSON, &queryObject); err != nil {
		return nil, &errors.Error{Status: http.StatusBadRequest, Err: err}
	}

	opts := map[string]interface{}{}
	multiOptions(options).Apply(opts)

	for k, v := range opts {
		queryObject[k] = v
	}

	return json.Marshal(queryObject)
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
	endQuery, err := db.startQuery()
	if err != nil {
		return err
	}
	defer endQuery()
	if finder, ok := db.driverDB.(driver.Finder); ok {
		return finder.CreateIndex(ctx, ddoc, name, index, multiOptions(options))
	}
	return errFindNotImplemented
}

// DeleteIndex deletes the requested index.
func (db *DB) DeleteIndex(ctx context.Context, ddoc, name string, options ...Option) error {
	if db.err != nil {
		return db.err
	}
	endQuery, err := db.startQuery()
	if err != nil {
		return err
	}
	defer endQuery()
	if finder, ok := db.driverDB.(driver.Finder); ok {
		return finder.DeleteIndex(ctx, ddoc, name, multiOptions(options))
	}
	return errFindNotImplemented
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
	endQuery, err := db.startQuery()
	if err != nil {
		return nil, err
	}
	defer endQuery()
	if finder, ok := db.driverDB.(driver.Finder); ok {
		dIndexes, err := finder.GetIndexes(ctx, multiOptions(options))
		indexes := make([]Index, len(dIndexes))
		for i, index := range dIndexes {
			indexes[i] = Index(index)
		}
		return indexes, err
	}
	return nil, errFindNotImplemented
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
	if explainer, ok := db.driverDB.(driver.Finder); ok {
		endQuery, err := db.startQuery()
		if err != nil {
			return nil, err
		}
		defer endQuery()
		plan, err := explainer.Explain(ctx, query, multiOptions(options))
		if err != nil {
			return nil, err
		}
		qp := QueryPlan(*plan)
		return &qp, nil
	}
	return nil, errFindNotImplemented
}
