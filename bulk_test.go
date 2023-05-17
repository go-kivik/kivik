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
	"errors"
	"fmt"
	"net/http"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal/mock"
)

func TestRLOCK(t *testing.T) {
	tests := []struct {
		name string
		iter *iter
		err  string
	}{
		{
			name: "not ready",
			iter: &iter{},
			err:  "kivik: Iterator access before calling Next",
		},
		{
			name: "closed",
			iter: &iter{state: stateClosed},
			err:  "kivik: Iterator is closed",
		},
		{
			name: "success",
			iter: &iter{state: stateRowReady},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			closeFn, err := test.iter.rlock()
			testy.Error(t, test.err, err)
			if closeFn == nil {
				t.Fatal("close is nil")
			}
			closeFn()
		})
	}
}

func TestDocsInterfaceSlice(t *testing.T) {
	type diTest struct {
		name     string
		input    []interface{}
		expected interface{}
		status   int
		err      string
	}
	tests := []diTest{
		{
			name:     "InterfaceSlice",
			input:    []interface{}{map[string]string{"foo": "bar"}},
			expected: []interface{}{map[string]string{"foo": "bar"}},
		},
		{
			name: "JSONDoc",
			input: []interface{}{
				map[string]string{"foo": "bar"},
				json.RawMessage(`{"foo":"bar"}`),
			},
			expected: []interface{}{
				map[string]string{"foo": "bar"},
				map[string]string{"foo": "bar"},
			},
		},
	}
	for _, test := range tests {
		func(test diTest) {
			t.Run(test.name, func(t *testing.T) {
				result, err := docsInterfaceSlice(test.input)
				testy.StatusError(t, test.err, test.status, err)
				if d := testy.DiffAsJSON(test.expected, result); d != nil {
					t.Errorf("%s", d)
				}
			})
		}(test)
	}
}

func TestBulkDocs(t *testing.T) { // nolint: gocyclo
	type tt struct {
		db       *DB
		docs     []interface{}
		options  Options
		expected []BulkResult
		status   int
		err      string
	}

	tests := testy.NewTable()
	tests.Add("invalid JSON", tt{
		db: &DB{
			client: &Client{},
			driverDB: &mock.BulkDocer{
				BulkDocsFunc: func(_ context.Context, docs []interface{}, _ map[string]interface{}) ([]driver.BulkResult, error) {
					_, err := json.Marshal(docs)
					return nil, err
				},
			},
		},
		docs:   []interface{}{json.RawMessage("invalid json")},
		status: http.StatusInternalServerError,
		err:    "json: error calling MarshalJSON for type json.RawMessage: invalid character 'i' looking for beginning of value",
	})
	tests.Add("emulated BulkDocs support", tt{
		db: &DB{
			client: &Client{},
			driverDB: &mock.DB{
				PutFunc: func(_ context.Context, docID string, doc interface{}, opts map[string]interface{}) (string, error) {
					if docID == "error" {
						return "", errors.New("error")
					}
					if docID != "foo" { // nolint: goconst
						return "", fmt.Errorf("Unexpected docID: %s", docID)
					}
					expectedDoc := map[string]string{"_id": "foo"}
					if d := testy.DiffInterface(expectedDoc, doc); d != nil {
						return "", fmt.Errorf("Unexpected doc:\n%s", d)
					}
					if d := testy.DiffInterface(testOptions, opts); d != nil {
						return "", fmt.Errorf("Unexpected opts:\n%s", d)
					}
					return "2-xxx", nil // nolint: goconst
				},
				CreateDocFunc: func(_ context.Context, doc interface{}, opts map[string]interface{}) (string, string, error) {
					expectedDoc := int(123)
					if d := testy.DiffInterface(expectedDoc, doc); d != nil {
						return "", "", fmt.Errorf("Unexpected doc:\n%s", d)
					}
					if d := testy.DiffInterface(testOptions, opts); d != nil {
						return "", "", fmt.Errorf("Unexpected opts:\n%s", d)
					}
					return "newDocID", "1-xxx", nil // nolint: goconst
				},
			},
		},
		docs: []interface{}{
			map[string]string{"_id": "foo"},
			123,
			map[string]string{"_id": "error"},
		},
		options: testOptions,
		expected: []BulkResult{
			{ID: "foo", Rev: "2-xxx"},
			{ID: "newDocID", Rev: "1-xxx"},
			{ID: "error", Error: errors.New("error")},
		},
	})
	tests.Add("new_edits", tt{
		db: &DB{
			client: &Client{},
			driverDB: &mock.BulkDocer{
				BulkDocsFunc: func(_ context.Context, docs []interface{}, opts map[string]interface{}) ([]driver.BulkResult, error) {
					expectedDocs := []interface{}{map[string]string{"_id": "foo"}, 123}
					expectedOpts := map[string]interface{}{"new_edits": true}
					if d := testy.DiffInterface(expectedDocs, docs); d != nil {
						return nil, fmt.Errorf("Unexpected docs:\n%s", d)
					}
					if d := testy.DiffInterface(expectedOpts, opts); d != nil {
						return nil, fmt.Errorf("Unexpected opts:\n%s", d)
					}
					return []driver.BulkResult{
						{ID: "foo"},
					}, nil
				},
			},
		},
		docs: []interface{}{
			map[string]string{"_id": "foo"},
			123,
		},
		options: Options{"new_edits": true},
		expected: []BulkResult{
			{ID: "foo"},
		},
	})
	tests.Add(errClientClosed, tt{
		db: &DB{
			client: &Client{
				closed: 1,
			},
		},
		docs: []interface{}{
			map[string]string{"_id": "foo"},
		},
		status: http.StatusServiceUnavailable,
		err:    errClientClosed,
	})
	tests.Add("db error", tt{
		db: &DB{
			err: errors.New("db error"),
		},
		status: http.StatusInternalServerError,
		err:    "db error",
	})
	tests.Add("unreadable doc", tt{
		db: &DB{
			client: &Client{},
			driverDB: &mock.BulkDocer{
				BulkDocsFunc: func(_ context.Context, docs []interface{}, _ map[string]interface{}) ([]driver.BulkResult, error) {
					_, err := json.Marshal(docs)
					return nil, err
				},
			},
		},
		docs:   []interface{}{testy.ErrorReader("", errors.New("read error"))},
		status: http.StatusBadRequest,
		err:    "read error",
	})
	tests.Run(t, func(t *testing.T, tt tt) {
		result, err := tt.db.BulkDocs(context.Background(), tt.docs, tt.options)
		testy.StatusError(t, tt.err, tt.status, err)
		if d := testy.DiffInterface(tt.expected, result); d != nil {
			t.Error(d)
		}
	})
}
