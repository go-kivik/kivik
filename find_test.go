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
	internal "github.com/go-kivik/kivik/v4/int/errors"
	"github.com/go-kivik/kivik/v4/int/mock"
)

func TestFind(t *testing.T) {
	tests := []struct {
		name     string
		db       *DB
		query    interface{}
		options  []Option
		expected *ResultSet
		status   int
		err      string
	}{
		{
			name: "non-finder",
			db: &DB{
				client:   &Client{},
				driverDB: &mock.DB{},
			},
			status: http.StatusNotImplemented,
			err:    "kivik: driver does not support Find interface",
		},
		{
			name: "db error",
			db: &DB{
				client: &Client{},
				driverDB: &mock.Finder{
					FindFunc: func(context.Context, interface{}, driver.Options) (driver.Rows, error) {
						return nil, errors.New("db error")
					},
				},
			},
			status: http.StatusInternalServerError,
			err:    "db error",
		},
		{
			name: "success",
			db: &DB{
				client: &Client{},
				driverDB: &mock.Finder{
					FindFunc: func(_ context.Context, query interface{}, _ driver.Options) (driver.Rows, error) {
						expectedQuery := json.RawMessage(`{"limit":3,"selector":{"foo":"bar"},"skip":10}`)
						if d := testy.DiffInterface(expectedQuery, query); d != nil {
							return nil, fmt.Errorf("Unexpected query:\n%s", d)
						}
						return &mock.Rows{ID: "a"}, nil
					},
				},
			},
			query: map[string]interface{}{"selector": map[string]interface{}{"foo": "bar"}},
			options: []Option{
				Param("limit", 3),
				Param("skip", 10),
			},
			expected: &ResultSet{
				iter: &iter{
					feed: &rowsIterator{
						Rows: &mock.Rows{ID: "a"},
					},
					curVal: &driver.Row{},
				},
				rowsi: &mock.Rows{ID: "a"},
			},
		},
		{
			name: "db error",
			db: &DB{
				err: errors.New("db error"),
			},
			status: http.StatusInternalServerError,
			err:    "db error",
		},
		{
			name: "client closed",
			db: &DB{
				client: &Client{
					closed: true,
				},
				driverDB: &mock.Finder{},
			},
			status: http.StatusServiceUnavailable,
			err:    "kivik: client closed",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rs := test.db.Find(context.Background(), test.query, test.options...)
			err := rs.Err()
			if d := internal.StatusErrorDiff(test.err, test.status, err); d != "" {
				t.Error(d)
			}
			if err != nil {
				return
			}
			rs.cancel = nil  // Determinism
			rs.onClose = nil // Determinism
			if d := testy.DiffInterface(test.expected, rs); d != nil {
				t.Error(d)
			}
		})
	}
	t.Run("standalone", func(t *testing.T) {
		t.Run("after err, close doesn't block", func(t *testing.T) {
			db := &DB{
				client: &Client{},
				driverDB: &mock.Finder{
					FindFunc: func(context.Context, interface{}, driver.Options) (driver.Rows, error) {
						return nil, errors.New("sdfsdf")
					},
				},
			}
			rows := db.Find(context.Background(), nil)
			if err := rows.Err(); err == nil {
				t.Fatal("expected an error, got none")
			}
			_ = db.Close() // Should not block
		})
		t.Run("not finder, close doesn't block", func(t *testing.T) {
			db := &DB{
				client:   &Client{},
				driverDB: &mock.DB{},
			}
			rows := db.Find(context.Background(), nil)
			if err := rows.Err(); err == nil {
				t.Fatal("expected an error, got none")
			}
			_ = db.Close() // Should not block
		})
	})
}

func TestCreateIndex(t *testing.T) {
	tests := []struct {
		testName   string
		db         *DB
		ddoc, name string
		index      interface{}
		status     int
		err        string
	}{
		{
			testName: "non-finder",
			db: &DB{
				client:   &Client{},
				driverDB: &mock.DB{},
			},
			status: http.StatusNotImplemented,
			err:    "kivik: driver does not support Find interface",
		},
		{
			testName: "db error",
			db: &DB{
				client: &Client{},
				driverDB: &mock.Finder{
					CreateIndexFunc: func(context.Context, string, string, interface{}, driver.Options) error {
						return errors.New("db error")
					},
				},
			},
			status: http.StatusInternalServerError,
			err:    "db error",
		},
		{
			testName: "success",
			db: &DB{
				client: &Client{},
				driverDB: &mock.Finder{
					CreateIndexFunc: func(_ context.Context, ddoc, name string, index interface{}, _ driver.Options) error {
						expectedDdoc := "foo"
						expectedName := "bar"
						expectedIndex := int(3)
						if expectedDdoc != ddoc {
							return fmt.Errorf("Unexpected ddoc: %s", ddoc)
						}
						if expectedName != name {
							return fmt.Errorf("Unexpected name: %s", name)
						}
						if d := testy.DiffInterface(expectedIndex, index); d != nil {
							return fmt.Errorf("Unexpected index:\n%s", d)
						}
						return nil
					},
				},
			},
			ddoc:  "foo",
			name:  "bar",
			index: int(3),
		},
		{
			name: "closed",
			db: &DB{
				client: &Client{
					closed: true,
				},
			},
			status: http.StatusServiceUnavailable,
			err:    "kivik: client closed",
		},
		{
			name: "db error",
			db: &DB{
				client: &Client{},
				err:    errors.New("db error"),
			},
			status: http.StatusInternalServerError,
			err:    "db error",
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			err := test.db.CreateIndex(context.Background(), test.ddoc, test.name, test.index)
			if d := internal.StatusErrorDiff(test.err, test.status, err); d != "" {
				t.Error(d)
			}
		})
	}
}

func TestDeleteIndex(t *testing.T) {
	tests := []struct {
		testName   string
		db         *DB
		ddoc, name string
		status     int
		err        string
	}{
		{
			testName: "non-finder",
			db: &DB{
				client:   &Client{},
				driverDB: &mock.DB{},
			},
			status: http.StatusNotImplemented,
			err:    "kivik: driver does not support Find interface",
		},
		{
			testName: "db error",
			db: &DB{
				client: &Client{},
				driverDB: &mock.Finder{
					DeleteIndexFunc: func(context.Context, string, string, driver.Options) error {
						return errors.New("db error")
					},
				},
			},
			status: http.StatusInternalServerError,
			err:    "db error",
		},
		{
			testName: "success",
			db: &DB{
				client: &Client{},
				driverDB: &mock.Finder{
					DeleteIndexFunc: func(_ context.Context, ddoc, name string, _ driver.Options) error {
						expectedDdoc := "foo"
						expectedName := "bar"
						if expectedDdoc != ddoc {
							return fmt.Errorf("Unexpected ddoc: %s", ddoc)
						}
						if expectedName != name {
							return fmt.Errorf("Unexpected name: %s", name)
						}
						return nil
					},
				},
			},
			ddoc: "foo",
			name: "bar",
		},
		{
			testName: "client closed",
			db: &DB{
				client: &Client{
					closed: true,
				},
			},
			status: http.StatusServiceUnavailable,
			err:    "kivik: client closed",
		},
		{
			testName: "db error",
			db: &DB{
				err: errors.New("db error"),
			},
			status: http.StatusInternalServerError,
			err:    "db error",
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			err := test.db.DeleteIndex(context.Background(), test.ddoc, test.name)
			if d := internal.StatusErrorDiff(test.err, test.status, err); d != "" {
				t.Error(d)
			}
		})
	}
}

func TestGetIndexes(t *testing.T) {
	tests := []struct {
		testName string
		db       *DB
		expected []Index
		status   int
		err      string
	}{
		{
			testName: "non-finder",
			db: &DB{
				client:   &Client{},
				driverDB: &mock.DB{},
			},
			status: http.StatusNotImplemented,
			err:    "kivik: driver does not support Find interface",
		},
		{
			testName: "db error",
			db: &DB{
				client: &Client{},
				driverDB: &mock.Finder{
					GetIndexesFunc: func(context.Context, driver.Options) ([]driver.Index, error) {
						return nil, errors.New("db error")
					},
				},
			},
			status: http.StatusInternalServerError,
			err:    "db error",
		},
		{
			testName: "success",
			db: &DB{
				client: &Client{},
				driverDB: &mock.Finder{
					GetIndexesFunc: func(context.Context, driver.Options) ([]driver.Index, error) {
						return []driver.Index{
							{Name: "a"},
							{Name: "b"},
						}, nil
					},
				},
			},
			expected: []Index{
				{
					Name: "a",
				},
				{
					Name: "b",
				},
			},
		},
		{
			testName: "client closed",
			db: &DB{
				client: &Client{
					closed: true,
				},
			},
			status: http.StatusServiceUnavailable,
			err:    "kivik: client closed",
		},
		{
			testName: "db error",
			db: &DB{
				err: errors.New("db error"),
			},
			status: http.StatusInternalServerError,
			err:    "db error",
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			result, err := test.db.GetIndexes(context.Background())
			if d := internal.StatusErrorDiff(test.err, test.status, err); d != "" {
				t.Error(d)
			}
			if err != nil {
				return
			}
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestExplain(t *testing.T) {
	type tt struct {
		db       *DB
		query    interface{}
		expected *QueryPlan
		status   int
		err      string
	}

	tests := testy.NewTable()
	tests.Add("non-finder", tt{
		db: &DB{
			client:   &Client{},
			driverDB: &mock.DB{},
		},
		status: http.StatusNotImplemented,
		err:    "kivik: driver does not support Find interface",
	})
	tests.Add("explain error", tt{
		db: &DB{
			client: &Client{},
			driverDB: &mock.Finder{
				ExplainFunc: func(context.Context, interface{}, driver.Options) (*driver.QueryPlan, error) {
					return nil, errors.New("explain error")
				},
			},
		},
		status: http.StatusInternalServerError,
		err:    "explain error",
	})
	tests.Add("success", tt{
		db: &DB{
			client: &Client{},
			driverDB: &mock.Finder{
				ExplainFunc: func(_ context.Context, query interface{}, _ driver.Options) (*driver.QueryPlan, error) {
					expectedQuery := int(3)
					if d := testy.DiffInterface(expectedQuery, query); d != nil {
						return nil, fmt.Errorf("Unexpected query:\n%s", d)
					}
					return &driver.QueryPlan{DBName: "foo"}, nil
				},
			},
		},
		query:    int(3),
		expected: &QueryPlan{DBName: "foo"},
	})
	tests.Add("client closed", tt{
		db: &DB{
			driverDB: &mock.Finder{},
			client: &Client{
				closed: true,
			},
		},
		status: http.StatusServiceUnavailable,
		err:    "kivik: client closed",
	})
	tests.Add("db error", tt{
		db: &DB{
			err: errors.New("db error"),
		},
		status: http.StatusInternalServerError,
		err:    "db error",
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		result, err := tt.db.Explain(context.Background(), tt.query)
		if d := internal.StatusErrorDiff(tt.err, tt.status, err); d != "" {
			t.Error(d)
		}
		if d := testy.DiffInterface(tt.expected, result); d != nil {
			t.Error(d)
		}
	})
}
