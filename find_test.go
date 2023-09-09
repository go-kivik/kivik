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
	"fmt"
	"net/http"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal/mock"
)

func TestFind(t *testing.T) {
	tests := []struct {
		name     string
		db       *DB
		query    interface{}
		expected *rows
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
					FindFunc: func(context.Context, interface{}, map[string]interface{}) (driver.Rows, error) {
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
					FindFunc: func(_ context.Context, query interface{}, _ map[string]interface{}) (driver.Rows, error) {
						expectedQuery := int(3)
						if d := testy.DiffInterface(expectedQuery, query); d != nil {
							return nil, fmt.Errorf("Unexpected query:\n%s", d)
						}
						return &mock.Rows{ID: "a"}, nil
					},
				},
			},
			query: int(3),
			expected: &rows{
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
			name: errClientClosed,
			db: &DB{
				client: &Client{
					closed: 1,
				},
				driverDB: &mock.Finder{},
			},
			status: http.StatusServiceUnavailable,
			err:    errClientClosed,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rs := test.db.Find(context.Background(), test.query)
			testy.StatusError(t, test.err, test.status, rs.Err())
			if r, ok := rs.underlying.(*rows); ok {
				r.cancel = nil  // Determinism
				r.onClose = nil // Determinism
			}
			if d := testy.DiffInterface(&ResultSet{underlying: test.expected}, rs); d != nil {
				t.Error(d)
			}
		})
	}
	t.Run("standalone", func(t *testing.T) {
		t.Run("after err, close doesn't block", func(t *testing.T) {
			db := &DB{
				client: &Client{},
				driverDB: &mock.Finder{
					FindFunc: func(context.Context, interface{}, map[string]interface{}) (driver.Rows, error) {
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
					CreateIndexFunc: func(context.Context, string, string, interface{}, map[string]interface{}) error {
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
					CreateIndexFunc: func(_ context.Context, ddoc, name string, index interface{}, _ map[string]interface{}) error {
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
					closed: 1,
				},
			},
			status: http.StatusServiceUnavailable,
			err:    errClientClosed,
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
			testy.StatusError(t, test.err, test.status, err)
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
					DeleteIndexFunc: func(context.Context, string, string, map[string]interface{}) error {
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
					DeleteIndexFunc: func(_ context.Context, ddoc, name string, _ map[string]interface{}) error {
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
			testName: errClientClosed,
			db: &DB{
				client: &Client{
					closed: 1,
				},
			},
			status: http.StatusServiceUnavailable,
			err:    errClientClosed,
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
			testy.StatusError(t, test.err, test.status, err)
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
					GetIndexesFunc: func(context.Context, map[string]interface{}) ([]driver.Index, error) {
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
					GetIndexesFunc: func(context.Context, map[string]interface{}) ([]driver.Index, error) {
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
			testName: errClientClosed,
			db: &DB{
				client: &Client{
					closed: 1,
				},
			},
			status: http.StatusServiceUnavailable,
			err:    errClientClosed,
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
			testy.StatusError(t, test.err, test.status, err)
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
				ExplainFunc: func(context.Context, interface{}, map[string]interface{}) (*driver.QueryPlan, error) {
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
				ExplainFunc: func(_ context.Context, query interface{}, _ map[string]interface{}) (*driver.QueryPlan, error) {
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
	tests.Add(errClientClosed, tt{
		db: &DB{
			client: &Client{
				closed: 1,
			},
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

	tests.Run(t, func(t *testing.T, tt tt) {
		result, err := tt.db.Explain(context.Background(), tt.query)
		testy.StatusError(t, tt.err, tt.status, err)
		if d := testy.DiffInterface(tt.expected, result); d != nil {
			t.Error(d)
		}
	})
}
