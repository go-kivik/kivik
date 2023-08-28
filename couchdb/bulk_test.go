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
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"gitlab.com/flimzy/testy"
)

func TestBulkDocs(t *testing.T) {
	tests := []struct {
		name    string
		db      *db
		docs    []interface{}
		options map[string]interface{}
		status  int
		err     string
	}{
		{
			name:   "network error",
			db:     newTestDB(nil, errors.New("net error")),
			status: http.StatusBadGateway,
			err:    `Post "?http://example.com/testdb/_bulk_docs"?: net error`,
		},
		{
			name: "JSON encoding error",
			db: newTestDB(&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil),
			docs:   []interface{}{make(chan int)},
			status: http.StatusBadRequest,
			err:    `Post "?http://example.com/testdb/_bulk_docs"?: json: unsupported type: chan int`,
		},
		{
			name: "docs rejected",
			db: newTestDB(&http.Response{
				StatusCode: http.StatusExpectationFailed,
				Body:       io.NopCloser(strings.NewReader("[]")),
			}, nil),
			docs:   []interface{}{1, 2, 3},
			status: http.StatusExpectationFailed,
			err:    "Expectation Failed: one or more document was rejected",
		},
		{
			name: "error response",
			db: newTestDB(&http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil),
			docs:   []interface{}{1, 2, 3},
			status: http.StatusBadRequest,
			err:    "Bad Request",
		},
		{
			name: "invalid JSON response",
			db: newTestDB(&http.Response{
				StatusCode: http.StatusCreated,
				Body:       io.NopCloser(strings.NewReader("invalid json")),
			}, nil),
			docs:   []interface{}{1, 2, 3},
			status: http.StatusBadGateway,
			err:    "invalid character 'i' looking for beginning of value",
		},
		{
			name: "unexpected response code",
			db: newTestDB(&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("[]")),
			}, nil),
			docs: []interface{}{1, 2, 3},
		},
		{
			name:    "new_edits",
			options: map[string]interface{}{"new_edits": true},
			db: newCustomDB(func(req *http.Request) (*http.Response, error) {
				defer req.Body.Close() // nolint: errcheck
				var body struct {
					NewEdits bool `json:"new_edits"`
				}
				if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
					return nil, err
				}
				if !body.NewEdits {
					return nil, errors.New("`new_edits` not set")
				}
				return &http.Response{
					StatusCode: http.StatusCreated,
					Body:       io.NopCloser(strings.NewReader("[]")),
				}, nil
			}),
		},
		{
			name:    "full commit",
			options: map[string]interface{}{OptionFullCommit: true},
			db: newCustomDB(func(req *http.Request) (*http.Response, error) {
				defer req.Body.Close() // nolint: errcheck
				var body map[string]interface{}
				if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
					return nil, err
				}
				if _, ok := body[OptionFullCommit]; ok {
					return nil, errors.New("Full Commit key found in body")
				}
				if value := req.Header.Get("X-Couch-Full-Commit"); value != "true" {
					return nil, errors.New("X-Couch-Full-Commit not set to true")
				}
				return &http.Response{
					StatusCode: http.StatusCreated,
					Body:       io.NopCloser(strings.NewReader("[]")),
				}, nil
			}),
		},
		{
			name:    "invalid full commit type",
			db:      &db{},
			options: map[string]interface{}{OptionFullCommit: 123},
			status:  http.StatusBadRequest,
			err:     "kivik: option 'X-Couch-Full-Commit' must be bool, not int",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := test.db.BulkDocs(context.Background(), test.docs, test.options)
			testy.StatusErrorRE(t, test.err, test.status, err)
		})
	}
}

type closeTracker struct {
	closed bool
	io.ReadCloser
}

func (c *closeTracker) Close() error {
	c.closed = true
	return c.ReadCloser.Close()
}
