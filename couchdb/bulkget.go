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

package couchdb

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-kivik/kivik/v4/couchdb/chttp"
	"github.com/go-kivik/kivik/v4/driver"
)

func (d *db) BulkGet(ctx context.Context, docs []driver.BulkGetReference, options driver.Options) (driver.Rows, error) {
	opts := map[string]interface{}{}
	options.Apply(opts)
	query, err := optionsToParams(opts)
	if err != nil {
		return nil, err
	}
	body := map[string]interface{}{
		"docs": docs,
	}
	chttpOpts := &chttp.Options{
		Query:   query,
		GetBody: chttp.BodyEncoder(body),
		Header: http.Header{
			chttp.HeaderIdempotencyKey: []string{},
		},
	}
	resp, err := d.Client.DoReq(ctx, http.MethodPost, d.path("_bulk_get"), chttpOpts)
	if err != nil {
		return nil, err
	}
	if err = chttp.ResponseError(resp); err != nil {
		return nil, err
	}
	return newBulkGetRows(ctx, resp.Body), nil
}

// bulkGetError represents an error for a single document returned by a
// GetBulk call.
type bulkGetError struct {
	ID     string `json:"id"`
	Rev    string `json:"rev"`
	Err    string `json:"error"`
	Reason string `json:"reason"`
}

var _ error = &bulkGetError{}

func (e *bulkGetError) Error() string {
	return fmt.Sprintf("%s: %s", e.Err, e.Reason)
}

type bulkResultDoc struct {
	Doc   json.RawMessage `json:"ok,omitempty"`
	Error *bulkGetError   `json:"error,omitempty"`
}

type bulkResult struct {
	ID   string          `json:"id"`
	Docs []bulkResultDoc `json:"docs"`
}
