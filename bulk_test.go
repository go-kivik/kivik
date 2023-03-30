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
	"io"
	"net/http"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal/mock"
)

func TestBulkNext(t *testing.T) {
	tests := []struct {
		name     string
		r        *BulkResults
		expected bool
	}{
		{
			name: "true",
			r: &BulkResults{
				iter: &iter{
					feed:   &TestFeed{max: 1},
					curVal: new(int64),
				},
			},
			expected: true,
		},
		{
			name: "false",
			r: &BulkResults{
				iter: &iter{
					feed:   &TestFeed{max: 0},
					curVal: new(int64),
				},
			},
			expected: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.r.Next()
			if result != test.expected {
				t.Errorf("Unexpected result: %v", result)
			}
		})
	}
}

func TestBulkErr(t *testing.T) {
	expected := "bulk error"
	r := &BulkResults{
		iter: &iter{lasterr: errors.New(expected)},
	}
	err := r.Err()
	testy.Error(t, expected, err)
}

func TestBulkClose(t *testing.T) {
	expected := "close error" // nolint: goconst
	r := &BulkResults{
		iter: &iter{
			feed: &TestFeed{closeErr: errors.New(expected)},
		},
	}
	err := r.Close()
	testy.Error(t, expected, err)
}

func TestBulkIteratorNext(t *testing.T) {
	tests := []struct {
		name     string
		r        *bulkIterator
		err      string
		expected *driver.BulkResult
	}{
		{
			name: "error",
			r: &bulkIterator{&mock.BulkResults{
				NextFunc: func(_ *driver.BulkResult) error {
					return errors.New("iter error")
				},
			}},
			err: "iter error",
		},
		{
			name: "success",
			r: &bulkIterator{&mock.BulkResults{
				NextFunc: func(result *driver.BulkResult) error {
					*result = driver.BulkResult{ID: "foo"}
					return nil
				},
			}},
			expected: &driver.BulkResult{ID: "foo"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := new(driver.BulkResult)
			err := test.r.Next(result)
			testy.Error(t, test.err, err)
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

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
			iter: &iter{closed: true},
			err:  "kivik: Iterator is closed",
		},
		{
			name: "success",
			iter: &iter{ready: true},
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
	type bdTest struct {
		name     string
		dbDriver driver.DB
		docs     []interface{}
		options  Options
		expected *BulkResults
		status   int
		err      string
	}
	tests := []bdTest{
		{
			name: "invalid JSON",
			dbDriver: &mock.BulkDocer{
				BulkDocsFunc: func(_ context.Context, docs []interface{}, _ map[string]interface{}) (driver.BulkResults, error) {
					_, err := json.Marshal(docs)
					return nil, err
				},
			},
			docs:   []interface{}{json.RawMessage("invalid json")},
			status: http.StatusInternalServerError,
			err:    "json: error calling MarshalJSON for type json.RawMessage: invalid character 'i' looking for beginning of value",
		},
		{
			name: "emulated BulkDocs support",
			dbDriver: &mock.DB{
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
			docs: []interface{}{
				map[string]string{"_id": "foo"},
				123,
				map[string]string{"_id": "error"},
			},
			options: testOptions,
			expected: &BulkResults{
				iter: &iter{
					feed: &bulkIterator{
						BulkResults: &emulatedBulkResults{
							results: []driver.BulkResult{
								{ID: "foo", Rev: "2-xxx"},
								{ID: "newDocID", Rev: "1-xxx"},
								{ID: "error", Error: errors.New("error")},
							},
						},
					},
					curVal: &driver.BulkResult{},
				},
				bulki: &emulatedBulkResults{
					results: []driver.BulkResult{
						{ID: "foo", Rev: "2-xxx"},
						{ID: "newDocID", Rev: "1-xxx"},
						{ID: "error", Error: errors.New("error")},
					},
				},
			},
		},
		{
			name: "new_edits",
			dbDriver: &mock.BulkDocer{
				BulkDocsFunc: func(_ context.Context, docs []interface{}, opts map[string]interface{}) (driver.BulkResults, error) {
					expectedDocs := []interface{}{map[string]string{"_id": "foo"}, 123}
					expectedOpts := map[string]interface{}{"new_edits": true}
					if d := testy.DiffInterface(expectedDocs, docs); d != nil {
						return nil, fmt.Errorf("Unexpected docs:\n%s", d)
					}
					if d := testy.DiffInterface(expectedOpts, opts); d != nil {
						return nil, fmt.Errorf("Unexpected opts:\n%s", d)
					}
					return &mock.BulkResults{ID: "foo"}, nil
				},
			},
			docs: []interface{}{
				map[string]string{"_id": "foo"},
				123,
			},
			options: Options{"new_edits": true},
			expected: &BulkResults{
				iter: &iter{
					feed: &bulkIterator{
						BulkResults: &mock.BulkResults{ID: "foo"},
					},
					curVal: &driver.BulkResult{},
				},
				bulki: &mock.BulkResults{ID: "foo"},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := &DB{driverDB: test.dbDriver}
			result, err := db.BulkDocs(context.Background(), test.docs, test.options)
			testy.StatusError(t, test.err, test.status, err)
			result.cancel = nil // Determinism
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestEmulatedBulkResults(t *testing.T) {
	results := []driver.BulkResult{
		{
			ID:    "chicken",
			Rev:   "foo",
			Error: nil,
		},
		{
			ID:    "duck",
			Rev:   "bar",
			Error: errors.New("fail"),
		},
		{
			ID:    "dog",
			Rev:   "baz",
			Error: nil,
		},
	}
	br := &emulatedBulkResults{results}
	result := &driver.BulkResult{}
	if err := br.Next(result); err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if d := testy.DiffInterface(&results[0], result); d != nil {
		t.Error(d)
	}
	if err := br.Next(result); err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if d := testy.DiffInterface(&results[1], result); d != nil {
		t.Error(d)
	}
	if err := br.Close(); err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if err := br.Next(result); err != io.EOF {
		t.Error("Expected EOF")
	}
}

func TestBulkResultsGetters(t *testing.T) {
	id := "foo"
	rev := "3-xxx"
	err := "update error"
	r := &BulkResults{
		iter: &iter{
			ready: true,
			curVal: &driver.BulkResult{
				ID:    id,
				Rev:   rev,
				Error: errors.New(err),
			},
		},
	}

	t.Run("ID", func(t *testing.T) {
		result := r.ID()
		if result != id {
			t.Errorf("Unexpected ID: %v", result)
		}
	})

	t.Run("Rev", func(t *testing.T) {
		result := r.Rev()
		if result != rev {
			t.Errorf("Unexpected Rev: %v", result)
		}
	})

	t.Run("UpdateErr", func(t *testing.T) {
		result := r.UpdateErr()
		testy.Error(t, err, result)
	})

	t.Run("Not ready", func(t *testing.T) {
		r.ready = false

		t.Run("ID", func(t *testing.T) {
			result := r.ID()
			if result != "" {
				t.Errorf("Unexpected ID: %v", result)
			}
		})

		t.Run("Rev", func(t *testing.T) {
			result := r.Rev()
			if result != "" {
				t.Errorf("Unexpected Rev: %v", result)
			}
		})

		t.Run("UpdateErr", func(t *testing.T) {
			result := r.UpdateErr()
			testy.Error(t, "", result)
		})
	})
}
