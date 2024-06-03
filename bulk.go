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
	"errors"
	"net/http"

	"github.com/go-kivik/kivik/v4/driver"
	internal "github.com/go-kivik/kivik/v4/int/errors"
)

// BulkResult is the result of a single BulkDoc update.
type BulkResult struct {
	ID    string `json:"id"`
	Rev   string `json:"rev"`
	Error error
}

// BulkDocs allows you to create and update multiple documents at the same time
// within a single request. This function returns an iterator over the results
// of the bulk operation.
//
// See the [CouchDB documentation].
//
// As with [DB.Put], each individual document may be a JSON-marshable object, or
// a raw JSON string in a [encoding/json.RawMessage], or [io.Reader].
//
// [CouchDB documentation]: https://docs.couchdb.org/en/stable/api/database/bulk-api.html#db-bulk-docs
func (db *DB) BulkDocs(ctx context.Context, docs []interface{}, options ...Option) ([]BulkResult, error) {
	if db.err != nil {
		return nil, db.err
	}
	docsi, err := docsInterfaceSlice(docs)
	if err != nil {
		return nil, err
	}
	if len(docsi) == 0 {
		return nil, &internal.Error{Status: http.StatusBadRequest, Err: errors.New("kivik: no documents provided")}
	}
	endQuery, err := db.startQuery()
	if err != nil {
		return nil, err
	}
	defer endQuery()
	opts := multiOptions(options)
	if bulkDocer, ok := db.driverDB.(driver.BulkDocer); ok {
		bulki, err := bulkDocer.BulkDocs(ctx, docsi, opts)
		if err != nil {
			return nil, err
		}
		results := make([]BulkResult, len(bulki))
		for i, result := range bulki {
			results[i] = BulkResult(result)
		}
		return results, nil
	}
	results := make([]BulkResult, 0, len(docsi))
	for _, doc := range docsi {
		var err error
		var id, rev string
		if docID, ok := extractDocID(doc); ok {
			id = docID
			rev, err = db.Put(ctx, id, doc, opts)
		} else {
			id, rev, err = db.CreateDoc(ctx, doc, opts)
		}
		results = append(results, BulkResult{
			ID:    id,
			Rev:   rev,
			Error: err,
		})
	}
	return results, nil
}

func docsInterfaceSlice(docsi []interface{}) ([]interface{}, error) {
	for i, doc := range docsi {
		x, err := normalizeFromJSON(doc)
		if err != nil {
			return nil, &internal.Error{Status: http.StatusBadRequest, Err: err}
		}
		docsi[i] = x
	}
	return docsi, nil
}
