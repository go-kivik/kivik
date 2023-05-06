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
				driverDB: &mock.DB{},
			},
			status: http.StatusNotImplemented,
			err:    "kivik: driver does not support Find interface",
		},
		{
			name: "db error",
			db: &DB{
				driverDB: &mock.OptsFinder{
					FindFunc: func(_ context.Context, _ interface{}, _ map[string]interface{}) (driver.Rows, error) {
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
				driverDB: &mock.OptsFinder{
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
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rs := test.db.Find(context.Background(), test.query)
			testy.StatusError(t, test.err, test.status, rs.Err())
			if r, ok := rs.(*rows); ok {
				r.cancel = nil // Determinism
			}
			if d := testy.DiffInterface(test.expected, rs); d != nil {
				t.Error(d)
			}
		})
	}
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
				driverDB: &mock.DB{},
			},
			status: http.StatusNotImplemented,
			err:    "kivik: driver does not support Find interface",
		},
		{
			testName: "db error",
			db: &DB{
				driverDB: &mock.OptsFinder{
					CreateIndexFunc: func(_ context.Context, _, _ string, _ interface{}, _ map[string]interface{}) error {
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
				driverDB: &mock.OptsFinder{
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
				driverDB: &mock.DB{},
			},
			status: http.StatusNotImplemented,
			err:    "kivik: driver does not support Find interface",
		},
		{
			testName: "db error",
			db: &DB{
				driverDB: &mock.OptsFinder{
					DeleteIndexFunc: func(_ context.Context, _, _ string, _ map[string]interface{}) error {
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
				driverDB: &mock.OptsFinder{
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
				driverDB: &mock.DB{},
			},
			status: http.StatusNotImplemented,
			err:    "kivik: driver does not support Find interface",
		},
		{
			testName: "db error",
			db: &DB{
				driverDB: &mock.OptsFinder{
					GetIndexesFunc: func(_ context.Context, _ map[string]interface{}) ([]driver.Index, error) {
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
				driverDB: &mock.OptsFinder{
					GetIndexesFunc: func(_ context.Context, _ map[string]interface{}) ([]driver.Index, error) {
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
	tests := []struct {
		name     string
		db       driver.DB
		query    interface{}
		expected *QueryPlan
		status   int
		err      string
	}{
		{
			name:   "non-finder",
			db:     &mock.DB{},
			status: http.StatusNotImplemented,
			err:    "kivik: driver does not support Find interface",
		},
		{
			name: "explain error",
			db: &mock.OptsFinder{
				ExplainFunc: func(_ context.Context, _ interface{}, _ map[string]interface{}) (*driver.QueryPlan, error) {
					return nil, errors.New("explain error")
				},
			},
			status: http.StatusInternalServerError,
			err:    "explain error",
		},
		{
			name: "success",
			db: &mock.OptsFinder{
				ExplainFunc: func(_ context.Context, query interface{}, _ map[string]interface{}) (*driver.QueryPlan, error) {
					expectedQuery := int(3)
					if d := testy.DiffInterface(expectedQuery, query); d != nil {
						return nil, fmt.Errorf("Unexpected query:\n%s", d)
					}
					return &driver.QueryPlan{DBName: "foo"}, nil
				},
			},
			query:    int(3),
			expected: &QueryPlan{DBName: "foo"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := &DB{driverDB: test.db}
			result, err := db.Explain(context.Background(), test.query)
			testy.StatusError(t, test.err, test.status, err)
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}
