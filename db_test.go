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
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4/driver"
	internal "github.com/go-kivik/kivik/v4/int/errors"
	"github.com/go-kivik/kivik/v4/int/mock"
)

func TestClient(t *testing.T) {
	client := &Client{}
	db := &DB{client: client}
	result := db.Client()
	if result != client {
		t.Errorf("Unexpected result. Expected %p, got %p", client, result)
	}
}

func TestName(t *testing.T) {
	dbName := "foo"
	db := &DB{name: dbName}
	result := db.Name()
	if result != dbName {
		t.Errorf("Unexpected result. Expected %s, got %s", dbName, result)
	}
}

func TestAllDocs(t *testing.T) {
	tests := []struct {
		name     string
		db       *DB
		options  Option
		expected *ResultSet
		status   int
		err      string
	}{
		{
			name: "db error",
			db: &DB{
				client: &Client{},
				driverDB: &mock.DB{
					AllDocsFunc: func(context.Context, driver.Options) (driver.Rows, error) {
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
				driverDB: &mock.DB{
					AllDocsFunc: func(_ context.Context, options driver.Options) (driver.Rows, error) {
						gotOpts := map[string]interface{}{}
						options.Apply(gotOpts)
						if d := testy.DiffInterface(testOptions, gotOpts); d != nil {
							return nil, fmt.Errorf("Unexpected options: %s", d)
						}
						return &mock.Rows{ID: "a"}, nil
					},
				},
			},
			options: Params(testOptions),
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
			name: "client closed",
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
				err: errors.New("db error"),
			},
			status: http.StatusInternalServerError,
			err:    "db error",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rs := test.db.AllDocs(context.Background(), test.options)
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
				driverDB: &mock.DB{
					AllDocsFunc: func(context.Context, driver.Options) (driver.Rows, error) {
						return nil, errors.New("unf")
					},
				},
			}
			rows := db.AllDocs(context.Background())
			if err := rows.Err(); err == nil {
				t.Fatal("expected an error, got none")
			}
			_ = db.Close() // Should not block
		})
		t.Run("missing ids", func(t *testing.T) {
			rows := []*driver.Row{
				{
					ID:    "i-exist",
					Key:   json.RawMessage("i-exist"),
					Value: strings.NewReader(`{"rev":"1-967a00dff5e02add41819138abb3284d"}`),
				},
				{
					Key:   json.RawMessage("i-dont"),
					Error: errors.New("not found"),
				},
			}
			db := &DB{
				client: &Client{},
				driverDB: &mock.DB{
					AllDocsFunc: func(context.Context, driver.Options) (driver.Rows, error) {
						return &mock.Rows{
							NextFunc: func(r *driver.Row) error {
								if len(rows) == 0 {
									return io.EOF
								}
								row := rows[0]
								rows = rows[1:]
								*r = *row
								return nil
							},
						}, nil
					},
				},
			}
			rs := db.AllDocs(context.Background(), Params(map[string]interface{}{
				"include_docs": true,
				"keys":         []string{"i-exist", "i-dont"},
			}))
			type row struct {
				ID    string
				Key   string
				Value string
				Doc   string
				Error string
			}
			want := []row{
				{
					ID:    "i-exist",
					Key:   "i-exist",
					Value: `{"rev":"1-967a00dff5e02add41819138abb3284d"}`,
				},
				{
					Key:   "i-dont",
					Error: "not found",
				},
			}
			var got []row
			for rs.Next() {
				var doc, value json.RawMessage
				_ = rs.ScanDoc(&doc)
				_ = rs.ScanValue(&value)
				var errStr string
				id, err := rs.ID()
				key, _ := rs.Key()
				if err != nil {
					errStr = err.Error()
				}
				got = append(got, row{
					ID:    id,
					Key:   key,
					Doc:   string(doc),
					Value: string(value),
					Error: errStr,
				})
			}
			if d := cmp.Diff(want, got, cmp.Transformer("Error", func(t error) string {
				if t == nil {
					return ""
				}
				return t.Error()
			})); d != "" {
				t.Error(d)
			}
		})
	})
}

func TestDesignDocs(t *testing.T) {
	tests := []struct {
		name     string
		db       *DB
		options  Option
		expected *ResultSet
		status   int
		err      string
	}{
		{
			name: "db error",
			db: &DB{
				client: &Client{},
				driverDB: &mock.DesignDocer{
					DesignDocsFunc: func(context.Context, driver.Options) (driver.Rows, error) {
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
				driverDB: &mock.DesignDocer{
					DesignDocsFunc: func(_ context.Context, options driver.Options) (driver.Rows, error) {
						opts := map[string]interface{}{}
						options.Apply(opts)
						if d := testy.DiffInterface(testOptions, opts); d != nil {
							return nil, fmt.Errorf("Unexpected options: %s", d)
						}
						return &mock.Rows{ID: "a"}, nil
					},
				},
			},
			options: Params(testOptions),
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
			name: "not supported",
			db: &DB{
				client:   &Client{},
				driverDB: &mock.DB{},
			},
			status: http.StatusNotImplemented,
			err:    "kivik: design doc view not supported by driver",
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
				driverDB: &mock.DesignDocer{},
			},
			status: http.StatusServiceUnavailable,
			err:    "kivik: client closed",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rs := test.db.DesignDocs(context.Background(), test.options)
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
				driverDB: &mock.DesignDocer{
					DesignDocsFunc: func(context.Context, driver.Options) (driver.Rows, error) {
						return nil, errors.New("unf")
					},
				},
			}
			rows := db.DesignDocs(context.Background())
			if err := rows.Err(); err == nil {
				t.Fatal("expected an error, got none")
			}
			_ = db.Close() // Should not block
		})
	})
}

func TestLocalDocs(t *testing.T) {
	tests := []struct {
		name     string
		db       *DB
		options  Option
		expected *ResultSet
		status   int
		err      string
	}{
		{
			name: "error",
			db: &DB{
				client: &Client{},
				driverDB: &mock.LocalDocer{
					LocalDocsFunc: func(context.Context, driver.Options) (driver.Rows, error) {
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
				driverDB: &mock.LocalDocer{
					LocalDocsFunc: func(_ context.Context, options driver.Options) (driver.Rows, error) {
						opts := map[string]interface{}{}
						options.Apply(opts)
						if d := testy.DiffInterface(testOptions, opts); d != nil {
							return nil, fmt.Errorf("Unexpected options: %s", d)
						}
						return &mock.Rows{ID: "a"}, nil
					},
				},
			},
			options: Params(testOptions),
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
			name: "not supported",
			db: &DB{
				client:   &Client{},
				driverDB: &mock.DB{},
			},
			status: http.StatusNotImplemented,
			err:    "kivik: local doc view not supported by driver",
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
				driverDB: &mock.LocalDocer{},
			},
			status: http.StatusServiceUnavailable,
			err:    "kivik: client closed",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rs := test.db.LocalDocs(context.Background(), test.options)
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
				driverDB: &mock.LocalDocer{
					LocalDocsFunc: func(context.Context, driver.Options) (driver.Rows, error) {
						return nil, errors.New("unf")
					},
				},
			}
			rows := db.LocalDocs(context.Background())
			if err := rows.Err(); err == nil {
				t.Fatal("expected an error, got none")
			}
			_ = db.Close() // Should not block
		})
	})
}

func TestQuery(t *testing.T) {
	tests := []struct {
		name       string
		db         *DB
		ddoc, view string
		options    Option
		expected   *ResultSet
		status     int
		err        string
	}{
		{
			name: "db error",
			db: &DB{
				client: &Client{},
				driverDB: &mock.DB{
					QueryFunc: func(context.Context, string, string, driver.Options) (driver.Rows, error) {
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
				driverDB: &mock.DB{
					QueryFunc: func(_ context.Context, ddoc, view string, options driver.Options) (driver.Rows, error) {
						expectedDdoc := "foo"
						expectedView := "bar" // nolint: goconst
						if ddoc != expectedDdoc {
							return nil, fmt.Errorf("Unexpected ddoc: %s", ddoc)
						}
						if view != expectedView {
							return nil, fmt.Errorf("Unexpected view: %s", view)
						}
						gotOpts := map[string]interface{}{}
						options.Apply(gotOpts)
						if d := testy.DiffInterface(testOptions, gotOpts); d != nil {
							return nil, fmt.Errorf("Unexpected options: %s", d)
						}
						return &mock.Rows{ID: "a"}, nil
					},
				},
			},
			ddoc:    "foo",
			view:    "bar",
			options: Params(testOptions),
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
			},
			status: http.StatusServiceUnavailable,
			err:    "kivik: client closed",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rs := test.db.Query(context.Background(), test.ddoc, test.view, test.options)
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
}

func TestGet(t *testing.T) {
	type tt struct {
		db       *DB
		docID    string
		options  Option
		expected string
		status   int
		err      string
	}

	tests := testy.NewTable()
	tests.Add("db error", tt{
		db: &DB{
			client: &Client{},
			driverDB: &mock.DB{
				GetFunc: func(context.Context, string, driver.Options) (*driver.Document, error) {
					return nil, fmt.Errorf("db error")
				},
			},
		},
		status: http.StatusInternalServerError,
		err:    "db error",
	})
	tests.Add("success", tt{
		db: &DB{
			client: &Client{},
			driverDB: &mock.DB{
				GetFunc: func(_ context.Context, docID string, options driver.Options) (*driver.Document, error) {
					expectedDocID := "foo"
					if docID != expectedDocID {
						return nil, fmt.Errorf("Unexpected docID: %s", docID)
					}
					gotOpts := map[string]interface{}{}
					options.Apply(gotOpts)
					if d := testy.DiffInterface(testOptions, gotOpts); d != nil {
						return nil, fmt.Errorf("Unexpected options:\n%s", d)
					}
					return &driver.Document{
						Rev:  "1-xxx",
						Body: body(`{"_id":"foo"}`),
					}, nil
				},
			},
		},
		docID:    "foo",
		options:  Params(testOptions),
		expected: `{"_id":"foo"}`,
	})
	tests.Add("streaming attachments", tt{
		db: &DB{
			client: &Client{},
			driverDB: &mock.DB{
				GetFunc: func(_ context.Context, docID string, options driver.Options) (*driver.Document, error) {
					expectedDocID := "foo"
					gotOpts := map[string]interface{}{}
					options.Apply(gotOpts)
					wantOpts := map[string]interface{}{"include_docs": true}
					if docID != expectedDocID {
						return nil, fmt.Errorf("Unexpected docID: %s", docID)
					}
					if d := testy.DiffInterface(wantOpts, gotOpts); d != nil {
						return nil, fmt.Errorf("Unexpected options:\n%s", d)
					}
					return &driver.Document{
						Rev:         "1-xxx",
						Body:        body(`{"_id":"foo"}`),
						Attachments: &mock.Attachments{ID: "asdf"},
					}, nil
				},
			},
		},
		docID:    "foo",
		options:  IncludeDocs(),
		expected: `{"_id":"foo"}`,
	})
	tests.Add("client closed", tt{
		db: &DB{
			client: &Client{
				closed: true,
			},
		},
		status: http.StatusServiceUnavailable,
		err:    "kivik: client closed",
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		var doc json.RawMessage
		err := tt.db.Get(context.Background(), tt.docID, tt.options).ScanDoc(&doc)
		if !testy.ErrorMatches(tt.err, err) {
			t.Errorf("Unexpected error: %s", err)
		}
		if status := HTTPStatus(err); status != tt.status {
			t.Errorf("Unexpected error status: %v", status)
		}
		if d := testy.DiffJSON([]byte(tt.expected), []byte(doc)); d != nil {
			t.Error(d)
		}
	})
}

func TestOpenRevs(t *testing.T) {
	tests := []struct {
		name     string
		db       *DB
		ddoc     string
		revs     []string
		options  Option
		expected *ResultSet
		status   int
		err      string
	}{
		{
			name: "db error",
			db: &DB{
				client: &Client{},
				driverDB: &mock.OpenRever{
					OpenRevsFunc: func(context.Context, string, []string, driver.Options) (driver.Rows, error) {
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
				driverDB: &mock.OpenRever{
					OpenRevsFunc: func(_ context.Context, ddoc string, revs []string, options driver.Options) (driver.Rows, error) {
						const expectedDdoc = "foo"
						expectedRevs := []string{"all"}
						if ddoc != expectedDdoc {
							return nil, fmt.Errorf("Unexpected ddoc: %s", ddoc)
						}
						if d := cmp.Diff(expectedRevs, revs); d != "" {
							return nil, fmt.Errorf("Unexpected revs: %s", d)
						}
						gotOpts := map[string]interface{}{}
						options.Apply(gotOpts)
						if d := testy.DiffInterface(testOptions, gotOpts); d != nil {
							return nil, fmt.Errorf("Unexpected options: %s", d)
						}
						return &mock.Rows{ID: "a"}, nil
					},
				},
			},
			ddoc:    "foo",
			revs:    []string{"all"},
			options: Params(testOptions),
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
			name: "unsupported by driver",
			db: &DB{
				client: &Client{
					closed: true,
				},
			},
			status: http.StatusNotImplemented,
			err:    "kivik: driver does not support OpenRevs interface",
		},
		{
			name: "client closed",
			db: &DB{
				driverDB: &mock.OpenRever{},
				client: &Client{
					closed: true,
				},
			},
			status: http.StatusServiceUnavailable,
			err:    "kivik: client closed",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rs := test.db.OpenRevs(context.Background(), test.ddoc, test.revs, test.options)
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
}

func TestFlush(t *testing.T) {
	tests := []struct {
		name   string
		db     *DB
		status int
		err    string
	}{
		{
			name: "non-Flusher",
			db: &DB{
				client:   &Client{},
				driverDB: &mock.DB{},
			},
			status: http.StatusNotImplemented,
			err:    "kivik: flush not supported by driver",
		},
		{
			name: "db error",
			db: &DB{
				client: &Client{},
				driverDB: &mock.Flusher{
					FlushFunc: func(context.Context) error {
						return &internal.Error{Status: http.StatusBadGateway, Err: errors.New("flush error")}
					},
				},
			},
			status: http.StatusBadGateway,
			err:    "flush error",
		},
		{
			name: "success",
			db: &DB{
				client: &Client{},
				driverDB: &mock.Flusher{
					FlushFunc: func(context.Context) error {
						return nil
					},
				},
			},
		},
		{
			name: "client closed",
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
				err: errors.New("db error"),
			},
			status: http.StatusInternalServerError,
			err:    "db error",
		},
		{
			name: "database closed",
			db: &DB{
				closed: true,
				client: &Client{
					closed: true,
				},
			},
			status: http.StatusServiceUnavailable,
			err:    "kivik: database closed",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.db.Flush(context.Background())
			if d := internal.StatusErrorDiff(test.err, test.status, err); d != "" {
				t.Error(d)
			}
		})
	}
}

func TestStats(t *testing.T) {
	tests := []struct {
		name     string
		db       *DB
		expected *DBStats
		status   int
		err      string
	}{
		{
			name: "stats error",
			db: &DB{
				client: &Client{},
				driverDB: &mock.DB{
					StatsFunc: func(context.Context) (*driver.DBStats, error) {
						return nil, &internal.Error{Status: http.StatusBadGateway, Err: errors.New("stats error")}
					},
				},
			},
			status: http.StatusBadGateway,
			err:    "stats error",
		},
		{
			name: "success",
			db: &DB{
				client: &Client{},
				driverDB: &mock.DB{
					StatsFunc: func(context.Context) (*driver.DBStats, error) {
						return &driver.DBStats{
							Name:           "foo",
							CompactRunning: true,
							DocCount:       1,
							DeletedCount:   2,
							UpdateSeq:      "abc",
							DiskSize:       3,
							ActiveSize:     4,
							ExternalSize:   5,
							Cluster: &driver.ClusterStats{
								Replicas:    6,
								Shards:      7,
								ReadQuorum:  8,
								WriteQuorum: 9,
							},
							RawResponse: []byte("foo"),
						}, nil
					},
				},
			},
			expected: &DBStats{
				Name:           "foo",
				CompactRunning: true,
				DocCount:       1,
				DeletedCount:   2,
				UpdateSeq:      "abc",
				DiskSize:       3,
				ActiveSize:     4,
				ExternalSize:   5,
				Cluster: &ClusterConfig{
					Replicas:    6,
					Shards:      7,
					ReadQuorum:  8,
					WriteQuorum: 9,
				},
				RawResponse: []byte("foo"),
			},
		},
		{
			name: "client closed",
			db: &DB{
				client: &Client{
					closed: true,
				},
			},
			status: http.StatusServiceUnavailable,
			err:    "kivik: client closed",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.db.Stats(context.Background())
			if d := internal.StatusErrorDiff(test.err, test.status, err); d != "" {
				t.Error(d)
			}
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestCompact(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		expected := "compact error"
		db := &DB{
			client: &Client{},
			driverDB: &mock.DB{
				CompactFunc: func(context.Context) error {
					return &internal.Error{Status: http.StatusBadRequest, Err: errors.New(expected)}
				},
			},
		}
		err := db.Compact(context.Background())
		if d := internal.StatusErrorDiff(expected, http.StatusBadRequest, err); d != "" {
			t.Error(d)
		}
	})
	t.Run("closed", func(t *testing.T) {
		const expected = "kivik: client closed"
		db := &DB{
			client: &Client{
				closed: true,
			},
		}
		err := db.Compact(context.Background())
		if d := internal.StatusErrorDiff(expected, http.StatusServiceUnavailable, err); d != "" {
			t.Error(d)
		}
	})
	t.Run("db error", func(t *testing.T) {
		db := &DB{
			client: &Client{},
			err:    errors.New("db error"),
		}
		err := db.Compact(context.Background())
		if !testy.ErrorMatches("db error", err) {
			t.Errorf("Unexpected error: %s", err)
		}
	})
}

func TestCompactView(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		expectedDDocID := "foo"
		expected := "compact view error"
		db := &DB{
			client: &Client{},
			driverDB: &mock.DB{
				CompactViewFunc: func(_ context.Context, ddocID string) error {
					if ddocID != expectedDDocID {
						return fmt.Errorf("Unexpected ddocID: %s", ddocID)
					}
					return &internal.Error{Status: http.StatusBadRequest, Err: errors.New(expected)}
				},
			},
		}
		err := db.CompactView(context.Background(), expectedDDocID)
		if d := internal.StatusErrorDiff(expected, http.StatusBadRequest, err); d != "" {
			t.Error(d)
		}
	})
	t.Run("closed", func(t *testing.T) {
		const expected = "kivik: client closed"
		db := &DB{
			client: &Client{
				closed: true,
			},
		}
		err := db.CompactView(context.Background(), "")
		if d := internal.StatusErrorDiff(expected, http.StatusServiceUnavailable, err); d != "" {
			t.Error(d)
		}
	})
	t.Run("db error", func(t *testing.T) {
		db := &DB{
			client: &Client{},
			err:    errors.New("db error"),
		}
		err := db.CompactView(context.Background(), "")
		if !testy.ErrorMatches("db error", err) {
			t.Errorf("Unexpected error: %s", err)
		}
	})
}

func TestViewCleanup(t *testing.T) {
	t.Run("compact error", func(t *testing.T) {
		expected := "compact error"
		db := &DB{
			client: &Client{},
			driverDB: &mock.DB{
				ViewCleanupFunc: func(context.Context) error {
					return &internal.Error{Status: http.StatusBadRequest, Err: errors.New(expected)}
				},
			},
		}
		err := db.ViewCleanup(context.Background())
		if d := internal.StatusErrorDiff(expected, http.StatusBadRequest, err); d != "" {
			t.Error(d)
		}
	})
	t.Run("client closed", func(t *testing.T) {
		const expected = "kivik: client closed"
		db := &DB{
			client: &Client{
				closed: true,
			},
		}
		err := db.ViewCleanup(context.Background())
		if d := internal.StatusErrorDiff(expected, http.StatusServiceUnavailable, err); d != "" {
			t.Error(d)
		}
	})
	t.Run("db error", func(t *testing.T) {
		const expected = "db error"
		db := &DB{
			err: errors.New(expected),
		}
		err := db.ViewCleanup(context.Background())
		if d := internal.StatusErrorDiff(expected, http.StatusInternalServerError, err); d != "" {
			t.Error(d)
		}
	})
}

func TestSecurity(t *testing.T) {
	tests := []struct {
		name     string
		db       *DB
		expected *Security
		status   int
		err      string
	}{
		{
			name: "security error",
			db: &DB{
				client: &Client{},
				driverDB: &mock.SecurityDB{
					SecurityFunc: func(context.Context) (*driver.Security, error) {
						return nil, &internal.Error{Status: http.StatusBadGateway, Err: errors.New("security error")}
					},
				},
			},
			status: http.StatusBadGateway,
			err:    "security error",
		},
		{
			name: "success",
			db: &DB{
				client: &Client{},
				driverDB: &mock.SecurityDB{
					SecurityFunc: func(context.Context) (*driver.Security, error) {
						return &driver.Security{
							Admins: driver.Members{
								Names: []string{"a"},
								Roles: []string{"b"},
							},
							Members: driver.Members{
								Names: []string{"c"},
								Roles: []string{"d"},
							},
						}, nil
					},
				},
			},
			expected: &Security{
				Admins: Members{
					Names: []string{"a"},
					Roles: []string{"b"},
				},
				Members: Members{
					Names: []string{"c"},
					Roles: []string{"d"},
				},
			},
		},
		{
			name: "client closed",
			db: &DB{
				client: &Client{
					closed: true,
				},
				driverDB: &mock.SecurityDB{},
			},
			status: http.StatusServiceUnavailable,
			err:    "kivik: client closed",
		},
		{
			name: "db error",
			db: &DB{
				err: errors.New("db error"),
			},
			status: http.StatusInternalServerError,
			err:    "db error",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.db.Security(context.Background())
			if d := internal.StatusErrorDiff(test.err, test.status, err); d != "" {
				t.Error(d)
			}
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestSetSecurity(t *testing.T) {
	tests := []struct {
		name     string
		db       *DB
		security *Security
		status   int
		err      string
	}{
		{
			name: "nil security",
			db: &DB{
				client:   &Client{},
				driverDB: &mock.SecurityDB{},
			},
			status: http.StatusBadRequest,
			err:    "kivik: security required",
		},
		{
			name: "set error",
			db: &DB{
				client: &Client{},
				driverDB: &mock.SecurityDB{
					SetSecurityFunc: func(context.Context, *driver.Security) error {
						return &internal.Error{Status: http.StatusBadGateway, Err: errors.New("set security error")}
					},
				},
			},
			security: &Security{},
			status:   http.StatusBadGateway,
			err:      "set security error",
		},
		{
			name: "success",
			db: &DB{
				client: &Client{},
				driverDB: &mock.SecurityDB{
					SetSecurityFunc: func(_ context.Context, security *driver.Security) error {
						expectedSecurity := &driver.Security{
							Admins: driver.Members{
								Names: []string{"a"},
								Roles: []string{"b"},
							},
							Members: driver.Members{
								Names: []string{"c"},
								Roles: []string{"d"},
							},
						}
						if d := testy.DiffInterface(expectedSecurity, security); d != nil {
							return fmt.Errorf("Unexpected security:\n%s", d)
						}
						return nil
					},
				},
			},
			security: &Security{
				Admins: Members{
					Names: []string{"a"},
					Roles: []string{"b"},
				},
				Members: Members{
					Names: []string{"c"},
					Roles: []string{"d"},
				},
			},
		},
		{
			name: "client closed",
			db: &DB{
				client: &Client{
					closed: true,
				},
				driverDB: &mock.SecurityDB{},
			},
			security: &Security{},
			status:   http.StatusServiceUnavailable,
			err:      "kivik: client closed",
		},
		{
			name: "db error",
			db: &DB{
				err: errors.New("db error"),
			},
			status: http.StatusInternalServerError,
			err:    "db error",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.db.SetSecurity(context.Background(), test.security)
			if d := internal.StatusErrorDiff(test.err, test.status, err); d != "" {
				t.Error(d)
			}
		})
	}
}

func TestGetRev(t *testing.T) { // nolint: gocyclo
	tests := []struct {
		name    string
		db      *DB
		docID   string
		rev     string
		options Option
		status  int
		err     string
	}{
		{
			name: "meta getter error",
			db: &DB{
				client: &Client{},
				driverDB: &mock.RevGetter{
					GetRevFunc: func(context.Context, string, driver.Options) (string, error) {
						return "", &internal.Error{Status: http.StatusBadGateway, Err: errors.New("get meta error")}
					},
				},
			},
			status: http.StatusBadGateway,
			err:    "get meta error",
		},
		{
			name: "meta getter success",
			db: &DB{
				client: &Client{},
				driverDB: &mock.RevGetter{
					GetRevFunc: func(_ context.Context, docID string, options driver.Options) (string, error) {
						expectedDocID := "foo"
						if docID != expectedDocID {
							return "", fmt.Errorf("Unexpected docID: %s", docID)
						}
						gotOpts := map[string]interface{}{}
						options.Apply(gotOpts)
						if d := testy.DiffInterface(testOptions, gotOpts); d != nil {
							return "", fmt.Errorf("Unexpected options:\n%s", d)
						}
						return "1-xxx", nil
					},
				},
			},
			docID:   "foo",
			options: Params(testOptions),
			rev:     "1-xxx",
		},
		{
			name: "non-meta getter error",
			db: &DB{
				client: &Client{},
				driverDB: &mock.DB{
					GetFunc: func(context.Context, string, driver.Options) (*driver.Document, error) {
						return nil, &internal.Error{Status: http.StatusBadGateway, Err: errors.New("get error")}
					},
				},
			},
			status: http.StatusBadGateway,
			err:    "get error",
		},
		{
			name: "non-meta getter success with rev",
			db: &DB{
				client: &Client{},
				driverDB: &mock.DB{
					GetFunc: func(_ context.Context, docID string, _ driver.Options) (*driver.Document, error) {
						expectedDocID := "foo"
						if docID != expectedDocID {
							return nil, fmt.Errorf("Unexpected docID: %s", docID)
						}
						return &driver.Document{
							Rev:  "1-xxx",
							Body: body(`{"_rev":"1-xxx"}`),
						}, nil
					},
				},
			},
			docID: "foo",
			rev:   "1-xxx",
		},
		{
			name: "non-meta getter success without rev",
			db: &DB{
				client: &Client{},
				driverDB: &mock.DB{
					GetFunc: func(_ context.Context, docID string, _ driver.Options) (*driver.Document, error) {
						expectedDocID := "foo"
						if docID != expectedDocID {
							return nil, fmt.Errorf("Unexpected docID: %s", docID)
						}
						return &driver.Document{
							Body: body(`{"_rev":"1-xxx"}`),
						}, nil
					},
				},
			},
			docID: "foo",
			rev:   "1-xxx",
		},
		{
			name: "non-meta getter success without rev, invalid json",
			db: &DB{
				client: &Client{},
				driverDB: &mock.DB{
					GetFunc: func(_ context.Context, docID string, _ driver.Options) (*driver.Document, error) {
						expectedDocID := "foo"
						if docID != expectedDocID {
							return nil, fmt.Errorf("Unexpected docID: %s", docID)
						}
						return &driver.Document{
							Body: body(`invalid json`),
						}, nil
					},
				},
			},
			docID:  "foo",
			status: http.StatusInternalServerError,
			err:    "invalid character 'i' looking for beginning of value",
		},
		{
			name: "client closed",
			db: &DB{
				client: &Client{
					closed: true,
				},
				driverDB: &mock.RevGetter{},
			},
			status: http.StatusServiceUnavailable,
			err:    "kivik: client closed",
		},
		{
			name: "db error",
			db: &DB{
				err: errors.New("db error"),
			},
			status: http.StatusInternalServerError,
			err:    "db error",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rev, err := test.db.GetRev(context.Background(), test.docID, test.options)
			if d := internal.StatusErrorDiff(test.err, test.status, err); d != "" {
				t.Error(d)
			}
			if rev != test.rev {
				t.Errorf("Unexpected rev: %v", rev)
			}
		})
	}
}

func TestCopy(t *testing.T) {
	tests := []struct {
		name           string
		db             *DB
		target, source string
		options        Option
		expected       string
		status         int
		err            string
	}{
		{
			name: "missing target",
			db: &DB{
				client: &Client{},
			},
			status: http.StatusBadRequest,
			err:    "kivik: targetID required",
		},
		{
			name: "missing source",
			db: &DB{
				client: &Client{},
			},
			target: "foo",
			status: http.StatusBadRequest,
			err:    "kivik: sourceID required",
		},
		{
			name: "copier error",
			db: &DB{
				client: &Client{},
				driverDB: &mock.Copier{
					CopyFunc: func(context.Context, string, string, driver.Options) (string, error) {
						return "", &internal.Error{Status: http.StatusBadRequest, Err: errors.New("copy error")}
					},
				},
			},
			target: "foo",
			source: "bar",
			status: http.StatusBadRequest,
			err:    "copy error",
		},
		{
			name: "copier success",
			db: &DB{
				client: &Client{},
				driverDB: &mock.Copier{
					CopyFunc: func(_ context.Context, target, source string, options driver.Options) (string, error) {
						expectedTarget := "foo"
						expectedSource := "bar"
						if target != expectedTarget {
							return "", fmt.Errorf("Unexpected target: %s", target)
						}
						if source != expectedSource {
							return "", fmt.Errorf("Unexpected source: %s", source)
						}
						gotOpts := map[string]interface{}{}
						options.Apply(gotOpts)
						if d := testy.DiffInterface(testOptions, gotOpts); d != nil {
							return "", fmt.Errorf("Unexpected options:\n%s", d)
						}
						return "1-xxx", nil
					},
				},
			},
			target:   "foo",
			source:   "bar",
			options:  Params(testOptions),
			expected: "1-xxx",
		},
		{
			name: "non-copier get error",
			db: &DB{
				client: &Client{},
				driverDB: &mock.DB{
					GetFunc: func(context.Context, string, driver.Options) (*driver.Document, error) {
						return nil, &internal.Error{Status: http.StatusBadGateway, Err: errors.New("get error")}
					},
				},
			},
			target: "foo",
			source: "bar",
			status: http.StatusBadGateway,
			err:    "get error",
		},
		{
			name: "non-copier invalid JSON",
			db: &DB{
				client: &Client{},
				driverDB: &mock.DB{
					GetFunc: func(context.Context, string, driver.Options) (*driver.Document, error) {
						return &driver.Document{
							Body: body("invalid json"),
						}, nil
					},
				},
			},
			target: "foo",
			source: "bar",
			status: http.StatusInternalServerError,
			err:    "invalid character 'i' looking for beginning of value",
		},
		{
			name: "non-copier put error",
			db: &DB{
				client: &Client{},
				driverDB: &mock.DB{
					GetFunc: func(context.Context, string, driver.Options) (*driver.Document, error) {
						return &driver.Document{
							Body: body(`{"_id":"foo","_rev":"1-xxx"}`),
						}, nil
					},
					PutFunc: func(context.Context, string, interface{}, driver.Options) (string, error) {
						return "", &internal.Error{Status: http.StatusBadGateway, Err: errors.New("put error")}
					},
				},
			},
			target: "foo",
			source: "bar",
			status: http.StatusBadGateway,
			err:    "put error",
		},
		{
			name: "success",
			db: &DB{
				client: &Client{},
				driverDB: &mock.DB{
					GetFunc: func(_ context.Context, docID string, _ driver.Options) (*driver.Document, error) {
						expectedDocID := "bar"
						if docID != expectedDocID {
							return nil, fmt.Errorf("Unexpected get docID: %s", docID)
						}
						return &driver.Document{
							Body: body(`{"_id":"bar","_rev":"1-xxx","foo":123.4}`),
						}, nil
					},
					PutFunc: func(_ context.Context, docID string, doc interface{}, options driver.Options) (string, error) {
						expectedDocID := "foo"
						expectedDoc := map[string]interface{}{"_id": "foo", "foo": 123.4}
						expectedOpts := map[string]interface{}{"batch": true}
						if docID != expectedDocID {
							return "", fmt.Errorf("Unexpected put docID: %s", docID)
						}
						if d := testy.DiffInterface(expectedDoc, doc); d != nil {
							return "", fmt.Errorf("Unexpected doc:\n%s", doc)
						}
						gotOpts := map[string]interface{}{}
						options.Apply(gotOpts)
						if d := testy.DiffInterface(expectedOpts, gotOpts); d != nil {
							return "", fmt.Errorf("Unexpected opts:\n%s", d)
						}
						return "1-xxx", nil
					},
				},
			},
			target:   "foo",
			source:   "bar",
			options:  Params(map[string]interface{}{"rev": "1-xxx", "batch": true}),
			expected: "1-xxx",
		},
		{
			name: "closed",
			db: &DB{
				client: &Client{
					closed: true,
				},
			},
			target: "x",
			source: "y",
			status: http.StatusServiceUnavailable,
			err:    "kivik: client closed",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.db.Copy(context.Background(), test.target, test.source, test.options)
			if d := internal.StatusErrorDiff(test.err, test.status, err); d != "" {
				t.Error(d)
			}
			if result != test.expected {
				t.Errorf("Unexpected result: %s", result)
			}
		})
	}
}

type errorReader struct{}

var _ io.Reader = &errorReader{}

func (r *errorReader) Read(_ []byte) (int, error) {
	return 0, errors.New("errorReader")
}

func TestNormalizeFromJSON(t *testing.T) {
	type njTest struct {
		Name     string
		Input    interface{}
		Expected interface{}
		Status   int
		Error    string
	}
	tests := []njTest{
		{
			Name:     "Interface",
			Input:    int(5),
			Expected: int(5),
		},
		{
			Name:     "RawMessage",
			Input:    json.RawMessage(`{"foo":"bar"}`),
			Expected: map[string]interface{}{"foo": "bar"},
		},
		{
			Name:     "ioReader",
			Input:    strings.NewReader(`{"foo":"bar"}`),
			Expected: map[string]interface{}{"foo": "bar"},
		},
		{
			Name:   "ErrorReader",
			Input:  &errorReader{},
			Status: http.StatusBadRequest,
			Error:  "errorReader",
		},
	}
	for _, test := range tests {
		func(test njTest) {
			t.Run(test.Name, func(t *testing.T) {
				result, err := normalizeFromJSON(test.Input)
				if d := internal.StatusErrorDiff(test.Error, test.Status, err); d != "" {
					t.Error(d)
				}
				if d := testy.DiffAsJSON(test.Expected, result); d != nil {
					t.Error(d)
				}
			})
		}(test)
	}
}

func TestPut(t *testing.T) {
	putFunc := func(_ context.Context, docID string, doc interface{}, options driver.Options) (string, error) {
		expectedDocID := "foo"
		expectedDoc := map[string]interface{}{"foo": "bar"}
		if expectedDocID != docID {
			return "", fmt.Errorf("Unexpected docID: %s", docID)
		}
		if d := testy.DiffAsJSON(expectedDoc, doc); d != nil {
			return "", fmt.Errorf("Unexpected doc: %s", d)
		}
		gotOpts := map[string]interface{}{}
		options.Apply(gotOpts)
		if d := testy.DiffInterface(testOptions, gotOpts); d != nil {
			return "", fmt.Errorf("Unexpected opts: %s", d)
		}
		return "1-xxx", nil
	}

	type test struct {
		db      *DB
		docID   string
		input   interface{}
		options Option
		status  int
		err     string
		newRev  string
	}

	tests := testy.NewTable()
	tests.Add("no docID", test{
		db: &DB{
			client: &Client{},
		},
		status: http.StatusBadRequest,
		err:    "kivik: docID required",
	})
	tests.Add("error", test{
		db: &DB{
			client: &Client{},
			driverDB: &mock.DB{
				PutFunc: func(context.Context, string, interface{}, driver.Options) (string, error) {
					return "", &internal.Error{Status: http.StatusBadRequest, Err: errors.New("db error")}
				},
			},
		},
		docID:  "foo",
		status: http.StatusBadRequest,
		err:    "db error",
	})
	tests.Add("Interface", test{
		db: &DB{
			client: &Client{},
			driverDB: &mock.DB{
				PutFunc: putFunc,
			},
		},
		docID:   "foo",
		input:   map[string]interface{}{"foo": "bar"},
		options: Params(testOptions),
		newRev:  "1-xxx",
	})
	tests.Add("InvalidJSON", test{
		db: &DB{
			client: &Client{},
			driverDB: &mock.DB{
				PutFunc: putFunc,
			},
		},
		docID:  "foo",
		input:  json.RawMessage("Something bogus"),
		status: http.StatusInternalServerError,
		err:    "Unexpected doc: failed to marshal actual value: invalid character 'S' looking for beginning of value",
	})
	tests.Add("Bytes", test{
		db: &DB{
			client: &Client{},
			driverDB: &mock.DB{
				PutFunc: putFunc,
			},
		},
		docID:   "foo",
		input:   []byte(`{"foo":"bar"}`),
		options: Params(testOptions),
		newRev:  "1-xxx",
	})
	tests.Add("RawMessage", test{
		db: &DB{
			client: &Client{},
			driverDB: &mock.DB{
				PutFunc: putFunc,
			},
		},
		docID:   "foo",
		input:   json.RawMessage(`{"foo":"bar"}`),
		options: Params(testOptions),
		newRev:  "1-xxx",
	})
	tests.Add("Reader", test{
		db: &DB{
			client: &Client{},
			driverDB: &mock.DB{
				PutFunc: putFunc,
			},
		},
		docID:   "foo",
		input:   strings.NewReader(`{"foo":"bar"}`),
		options: Params(testOptions),
		newRev:  "1-xxx",
	})
	tests.Add("ErrorReader", test{
		db: &DB{
			client: &Client{},
		},
		docID:  "foo",
		input:  &errorReader{},
		status: http.StatusBadRequest,
		err:    "errorReader",
	})
	tests.Add("client closed", test{
		db: &DB{
			client: &Client{
				closed: true,
			},
		},
		docID:  "foo",
		status: http.StatusServiceUnavailable,
		err:    "kivik: client closed",
	})
	tests.Add("db error", test{
		db: &DB{
			err: errors.New("db error"),
		},
		status: http.StatusInternalServerError,
		err:    "db error",
	})
	tests.Run(t, func(t *testing.T, tt test) {
		newRev, err := tt.db.Put(context.Background(), tt.docID, tt.input, tt.options)
		if d := internal.StatusErrorDiff(tt.err, tt.status, err); d != "" {
			t.Error(d)
		}
		if newRev != tt.newRev {
			t.Errorf("Unexpected new rev: %s", newRev)
		}
	})
}

func TestExtractDocID(t *testing.T) {
	type ediTest struct {
		name     string
		i        interface{}
		id       string
		expected bool
	}
	tests := []ediTest{
		{
			name: "nil",
		},
		{
			name: "string/interface map, no id",
			i: map[string]interface{}{
				"value": "foo",
			},
		},
		{
			name: "string/interface map, with id",
			i: map[string]interface{}{
				"_id": "foo",
			},
			id:       "foo",
			expected: true,
		},
		{
			name: "string/string map, with id",
			i: map[string]string{
				"_id": "foo",
			},
			id:       "foo",
			expected: true,
		},
		{
			name: "invalid JSON",
			i:    make(chan int),
		},
		{
			name: "valid JSON",
			i: struct {
				ID string `json:"_id"`
			}{ID: "oink"},
			id:       "oink",
			expected: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			id, ok := extractDocID(test.i)
			if ok != test.expected || test.id != id {
				t.Errorf("Expected %t/%s, got %t/%s", test.expected, test.id, ok, id)
			}
		})
	}
}

func TestCreateDoc(t *testing.T) {
	tests := []struct {
		name       string
		db         *DB
		doc        interface{}
		options    Option
		docID, rev string
		status     int
		err        string
	}{
		{
			name: "error",
			db: &DB{
				client: &Client{},
				driverDB: &mock.DocCreator{
					CreateDocFunc: func(context.Context, interface{}, driver.Options) (string, string, error) {
						return "", "", &internal.Error{Status: http.StatusBadRequest, Err: errors.New("create error")}
					},
				},
			},
			status: http.StatusBadRequest,
			err:    "create error",
		},
		{
			name: "success",
			db: &DB{
				client: &Client{},
				driverDB: &mock.DocCreator{
					CreateDocFunc: func(_ context.Context, doc interface{}, options driver.Options) (string, string, error) {
						gotOpts := map[string]interface{}{}
						options.Apply(gotOpts)
						expectedDoc := map[string]string{"type": "test"}
						if d := testy.DiffInterface(expectedDoc, doc); d != nil {
							return "", "", fmt.Errorf("Unexpected doc:\n%s", d)
						}
						if d := testy.DiffInterface(testOptions, gotOpts); d != nil {
							return "", "", fmt.Errorf("Unexpected options:\n%s", d)
						}
						return "foo", "1-xxx", nil
					},
				},
			},
			doc:     map[string]string{"type": "test"},
			options: Params(testOptions),
			docID:   "foo",
			rev:     "1-xxx",
		},
		{
			name: "closed",
			db: &DB{
				client: &Client{
					closed: true,
				},
				driverDB: &mock.DocCreator{},
			},
			status: http.StatusServiceUnavailable,
			err:    "kivik: client closed",
		},
		{
			name: "emulated with docID",
			db: &DB{
				client: &Client{},
				driverDB: &mock.DB{
					PutFunc: func(_ context.Context, docID string, doc interface{}, options driver.Options) (string, error) {
						gotOpts := map[string]interface{}{}
						options.Apply(gotOpts)
						expectedDoc := map[string]string{"_id": "foo", "type": "test"}
						if docID != "foo" {
							return "", fmt.Errorf("Unexpected docID: %s", docID)
						}
						if d := testy.DiffInterface(expectedDoc, doc); d != nil {
							return "", fmt.Errorf("Unexpected doc:\n%s", d)
						}
						if d := testy.DiffInterface(testOptions, gotOpts); d != nil {
							return "", fmt.Errorf("Unexpected options:\n%s", d)
						}
						return "1-xxx", nil
					},
				},
			},
			doc:     map[string]string{"type": "test", "_id": "foo"},
			options: Params(testOptions),
			docID:   "foo",
			rev:     "1-xxx",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			docID, rev, err := test.db.CreateDoc(context.Background(), test.doc, test.options)
			if d := internal.StatusErrorDiff(test.err, test.status, err); d != "" {
				t.Error(d)
			}
			if docID != test.docID || rev != test.rev {
				t.Errorf("Unexpected result: %s / %s", docID, rev)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name       string
		db         *DB
		docID, rev string
		options    Option
		newRev     string
		status     int
		err        string
	}{
		{
			name: "no doc ID",
			db: &DB{
				client: &Client{},
			},
			status: http.StatusBadRequest,
			err:    "kivik: docID required",
		},
		{
			name: "error",
			db: &DB{
				client: &Client{},
				driverDB: &mock.DB{
					DeleteFunc: func(context.Context, string, driver.Options) (string, error) {
						return "", &internal.Error{Status: http.StatusBadRequest, Err: errors.New("delete error")}
					},
				},
			},
			docID:  "foo",
			status: http.StatusBadRequest,
			err:    "delete error",
		},
		{
			name: "rev in opts",
			db: &DB{
				client: &Client{},
				driverDB: &mock.DB{
					DeleteFunc: func(_ context.Context, docID string, options driver.Options) (string, error) {
						const expectedDocID = "foo"
						if docID != expectedDocID {
							return "", fmt.Errorf("Unexpected docID: %s", docID)
						}
						gotOpts := map[string]interface{}{}
						options.Apply(gotOpts)
						wantOpts := map[string]interface{}{"rev": "1-xxx"}
						if d := testy.DiffInterface(wantOpts, gotOpts); d != nil {
							return "", fmt.Errorf("Unexpected options:\n%s", d)
						}
						return "2-xxx", nil
					},
				},
			},
			docID:   "foo",
			options: Rev("1-xxx"),
			newRev:  "2-xxx",
		},
		{
			name: "success",
			db: &DB{
				client: &Client{},
				driverDB: &mock.DB{
					DeleteFunc: func(_ context.Context, docID string, options driver.Options) (string, error) {
						const expectedDocID = "foo"
						if docID != expectedDocID {
							return "", fmt.Errorf("Unexpected docID: %s", docID)
						}
						wantOpts := map[string]interface{}{
							"foo": 123,
							"rev": "1-xxx",
						}
						gotOpts := map[string]interface{}{}
						options.Apply(gotOpts)
						if d := testy.DiffInterface(wantOpts, gotOpts); d != nil {
							return "", fmt.Errorf("Unexpected options:\n%s", d)
						}
						return "2-xxx", nil
					},
				},
			},
			docID:   "foo",
			rev:     "1-xxx",
			options: Params(testOptions),
			newRev:  "2-xxx",
		},
		{
			name: "success, no opts",
			db: &DB{
				client: &Client{},
				driverDB: &mock.DB{
					DeleteFunc: func(_ context.Context, docID string, options driver.Options) (string, error) {
						const expectedDocID = "foo"
						if docID != expectedDocID {
							return "", fmt.Errorf("Unexpected docID: %s", docID)
						}
						wantOpts := map[string]interface{}{
							"rev": "1-xxx",
						}
						gotOpts := map[string]interface{}{}
						options.Apply(gotOpts)
						if d := testy.DiffInterface(wantOpts, gotOpts); d != nil {
							return "", fmt.Errorf("Unexpected options:\n%s", d)
						}
						return "2-xxx", nil
					},
				},
			},
			docID:  "foo",
			rev:    "1-xxx",
			newRev: "2-xxx",
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
				err: errors.New("db error"),
			},
			status: http.StatusInternalServerError,
			err:    "db error",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			newRev, err := test.db.Delete(context.Background(), test.docID, test.rev, test.options)
			if d := internal.StatusErrorDiff(test.err, test.status, err); d != "" {
				t.Error(d)
			}
			if newRev != test.newRev {
				t.Errorf("Unexpected newRev: %s", newRev)
			}
		})
	}
}

func TestPutAttachment(t *testing.T) {
	tests := []struct {
		name    string
		db      *DB
		docID   string
		att     *Attachment
		options Option
		newRev  string
		status  int
		err     string

		body string
	}{
		{
			name:  "db error",
			docID: "foo",
			db: &DB{
				client: &Client{},
				driverDB: &mock.DB{
					PutAttachmentFunc: func(context.Context, string, *driver.Attachment, driver.Options) (string, error) {
						return "", &internal.Error{Status: http.StatusBadRequest, Err: errors.New("db error")}
					},
				},
			},
			att: &Attachment{
				Filename: "foo.txt",
				Content:  io.NopCloser(strings.NewReader("")),
			},
			status: http.StatusBadRequest,
			err:    "db error",
		},
		{
			name: "no doc id",
			db: &DB{
				client: &Client{},
			},
			status: http.StatusBadRequest,
			err:    "kivik: docID required",
		},
		{
			name: "no filename",
			db: &DB{
				client: &Client{},
			},
			docID:  "foo",
			att:    &Attachment{},
			status: http.StatusBadRequest,
			err:    "kivik: filename required",
		},
		{
			name:  "success",
			docID: "foo",
			db: &DB{
				client: &Client{},
				driverDB: &mock.DB{
					PutAttachmentFunc: func(_ context.Context, docID string, att *driver.Attachment, options driver.Options) (string, error) {
						const expectedDocID = "foo"
						const expectedContent = "Test file"
						expectedAtt := &driver.Attachment{
							Filename:    "foo.txt",
							ContentType: "text/plain",
						}
						if docID != expectedDocID {
							return "", fmt.Errorf("Unexpected docID: %s", docID)
						}
						content, err := io.ReadAll(att.Content)
						if err != nil {
							t.Fatal(err)
						}
						if d := testy.DiffText(expectedContent, string(content)); d != nil {
							return "", fmt.Errorf("Unexpected content:\n%s", string(content))
						}
						att.Content = nil
						if d := testy.DiffInterface(expectedAtt, att); d != nil {
							return "", fmt.Errorf("Unexpected attachment:\n%s", d)
						}
						wantOpts := map[string]interface{}{"rev": "1-xxx"}
						gotOpts := map[string]interface{}{}
						options.Apply(gotOpts)
						if d := testy.DiffInterface(wantOpts, gotOpts); d != nil {
							return "", fmt.Errorf("Unexpected options:\n%s", d)
						}
						return "2-xxx", nil
					},
				},
			},
			att: &Attachment{
				Filename:    "foo.txt",
				ContentType: "text/plain",
				Content:     io.NopCloser(strings.NewReader("Test file")),
			},
			options: Rev("1-xxx"),
			newRev:  "2-xxx",
			body:    "Test file",
		},
		{
			name: "nil attachment",
			db: &DB{
				client: &Client{},
			},
			docID:  "foo",
			status: http.StatusBadRequest,
			err:    "kivik: attachment required",
		},
		{
			name: "client closed",
			db: &DB{
				client: &Client{
					closed: true,
				},
			},
			docID: "foo",
			att: &Attachment{
				Filename: "foo.txt",
			},
			status: http.StatusServiceUnavailable,
			err:    "kivik: client closed",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			newRev, err := test.db.PutAttachment(context.Background(), test.docID, test.att, test.options)
			if d := internal.StatusErrorDiff(test.err, test.status, err); d != "" {
				t.Error(d)
			}
			if newRev != test.newRev {
				t.Errorf("Unexpected newRev: %s", newRev)
			}
		})
	}
}

func TestDeleteAttachment(t *testing.T) {
	const (
		expectedDocID    = "foo"
		expectedRev      = "1-xxx"
		expectedFilename = "foo.txt"
	)

	type tt struct {
		db                   *DB
		docID, rev, filename string
		options              Option

		newRev string
		status int
		err    string
	}

	tests := testy.NewTable()
	tests.Add("missing doc id", tt{
		db: &DB{
			client: &Client{},
		},
		status: http.StatusBadRequest,
		err:    "kivik: docID required",
	})
	tests.Add("missing filename", tt{
		db: &DB{
			client: &Client{},
		},
		docID:  "foo",
		status: http.StatusBadRequest,
		err:    "kivik: filename required",
	})
	tests.Add("error", tt{
		docID:    "foo",
		filename: expectedFilename,
		db: &DB{
			client: &Client{},
			driverDB: &mock.DB{
				DeleteAttachmentFunc: func(context.Context, string, string, driver.Options) (string, error) {
					return "", &internal.Error{Status: http.StatusBadRequest, Err: errors.New("db error")}
				},
			},
		},
		status: http.StatusBadRequest,
		err:    "db error",
	})
	tests.Add("rev in options", func(t *testing.T) interface{} {
		return tt{
			docID:    expectedDocID,
			filename: expectedFilename,
			options:  Rev(expectedRev),
			db: &DB{
				client: &Client{},
				driverDB: &mock.DB{
					DeleteAttachmentFunc: func(_ context.Context, docID, filename string, options driver.Options) (string, error) {
						if docID != expectedDocID {
							t.Errorf("Unexpected docID: %s", docID)
						}
						if filename != expectedFilename {
							t.Errorf("Unexpected filename: %s", filename)
						}
						wantOpts := map[string]interface{}{"rev": expectedRev}
						gotOpts := map[string]interface{}{}
						options.Apply(gotOpts)
						if d := testy.DiffInterface(wantOpts, gotOpts); d != nil {
							t.Errorf("Unexpected options:\n%s", d)
						}
						return "2-xxx", nil
					},
				},
			},
			newRev: "2-xxx",
		}
	})
	tests.Add("success", func(t *testing.T) interface{} {
		return tt{
			docID:    expectedDocID,
			rev:      expectedRev,
			filename: expectedFilename,
			options:  Params(testOptions),
			newRev:   "2-xxx",
			db: &DB{
				client: &Client{},
				driverDB: &mock.DB{
					DeleteAttachmentFunc: func(_ context.Context, docID, filename string, options driver.Options) (string, error) {
						if docID != expectedDocID {
							t.Errorf("Unexpected docID: %s", docID)
						}
						if filename != expectedFilename {
							t.Errorf("Unexpected filename: %s", filename)
						}
						wantOpts := map[string]interface{}{
							"foo": 123,
							"rev": "1-xxx",
						}
						gotOpts := map[string]interface{}{}
						options.Apply(gotOpts)
						if d := testy.DiffInterface(wantOpts, gotOpts); d != nil {
							t.Errorf("Unexpected options:\n%s", d)
						}
						return "2-xxx", nil
					},
				},
			},
		}
	})
	tests.Add("closed", tt{
		db: &DB{
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
		newRev, err := tt.db.DeleteAttachment(context.Background(), tt.docID, tt.rev, tt.filename, tt.options)
		if d := internal.StatusErrorDiff(tt.err, tt.status, err); d != "" {
			t.Error(d)
		}
		if newRev != tt.newRev {
			t.Errorf("Unexpected new rev: %s", newRev)
		}
	})
}

func TestGetAttachment(t *testing.T) {
	expectedDocID, expectedFilename := "foo", "foo.txt"
	type tt struct {
		db              *DB
		docID, filename string
		options         Option

		content  string
		expected *Attachment
		status   int
		err      string
	}

	tests := testy.NewTable()
	tests.Add("error", tt{
		db: &DB{
			client: &Client{},
			driverDB: &mock.DB{
				GetAttachmentFunc: func(context.Context, string, string, driver.Options) (*driver.Attachment, error) {
					return nil, errors.New("fail")
				},
			},
		},
		docID:    expectedDocID,
		filename: expectedFilename,
		status:   500,
		err:      "fail",
	})
	tests.Add("success", func(t *testing.T) interface{} {
		return tt{
			docID:    expectedDocID,
			filename: expectedFilename,
			options:  Params(testOptions),
			content:  "Test",
			expected: &Attachment{
				Filename:    expectedFilename,
				ContentType: "text/plain",
				Size:        4,
				Digest:      "md5-foo",
			},
			db: &DB{
				client: &Client{},
				driverDB: &mock.DB{
					GetAttachmentFunc: func(_ context.Context, docID, filename string, options driver.Options) (*driver.Attachment, error) {
						if docID != expectedDocID {
							t.Errorf("Unexpected docID: %s", docID)
						}
						if filename != expectedFilename {
							t.Errorf("Unexpected filename: %s", filename)
						}
						gotOpts := map[string]interface{}{}
						options.Apply(gotOpts)
						if d := testy.DiffInterface(testOptions, gotOpts); d != nil {
							t.Errorf("Unexpected options:\n%s", d)
						}
						return &driver.Attachment{
							Filename:    expectedFilename,
							ContentType: "text/plain",
							Digest:      "md5-foo",
							Size:        4,
							Content:     body("Test"),
						}, nil
					},
				},
			},
		}
	})
	tests.Add("no docID", tt{
		db: &DB{
			client: &Client{},
		},
		status: http.StatusBadRequest,
		err:    "kivik: docID required",
	})
	tests.Add("no filename", tt{
		db: &DB{
			client: &Client{},
		},
		docID:  "foo",
		status: http.StatusBadRequest,
		err:    "kivik: filename required",
	})
	tests.Add("client closed", tt{
		db: &DB{
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
		result, err := tt.db.GetAttachment(context.Background(), tt.docID, tt.filename, tt.options)
		if d := internal.StatusErrorDiff(tt.err, tt.status, err); d != "" {
			t.Error(d)
		}
		if err != nil {
			return
		}
		content, err := io.ReadAll(result.Content)
		if err != nil {
			t.Fatal(err)
		}
		if d := testy.DiffText(tt.content, string(content)); d != nil {
			t.Errorf("Unexpected content:\n%s", d)
		}
		_ = result.Content.Close()
		result.Content = nil
		if d := testy.DiffInterface(tt.expected, result); d != nil {
			t.Error(d)
		}
	})
}

func TestGetAttachmentMeta(t *testing.T) { // nolint: gocyclo
	const expectedDocID, expectedFilename = "foo", "foo.txt"
	tests := []struct {
		name            string
		db              *DB
		docID, filename string
		options         Option

		expected *Attachment
		status   int
		err      string
	}{
		{
			name: "plain db, error",
			db: &DB{
				client: &Client{},
				driverDB: &mock.DB{
					GetAttachmentFunc: func(context.Context, string, string, driver.Options) (*driver.Attachment, error) {
						return nil, errors.New("fail")
					},
				},
			},
			docID:    "foo",
			filename: expectedFilename,
			status:   500,
			err:      "fail",
		},
		{
			name: "plain db, success",
			db: &DB{
				client: &Client{},
				driverDB: &mock.DB{
					GetAttachmentFunc: func(_ context.Context, docID, filename string, options driver.Options) (*driver.Attachment, error) {
						if docID != expectedDocID {
							return nil, fmt.Errorf("Unexpected docID: %s", docID)
						}
						if filename != expectedFilename {
							return nil, fmt.Errorf("Unexpected filename: %s", filename)
						}
						gotOpts := map[string]interface{}{}
						options.Apply(gotOpts)
						if d := testy.DiffInterface(testOptions, gotOpts); d != nil {
							return nil, fmt.Errorf("Unexpected options:\n%s", d)
						}
						return &driver.Attachment{
							Filename:    "foo.txt",
							ContentType: "text/plain",
							Digest:      "md5-foo",
							Size:        4,
							Content:     body("Test"),
						}, nil
					},
				},
			},
			docID:    "foo",
			filename: "foo.txt",
			options:  Params(testOptions),
			expected: &Attachment{
				Filename:    "foo.txt",
				ContentType: "text/plain",
				Digest:      "md5-foo",
				Size:        4,
				Content:     nilContent,
			},
		},
		{
			name: "error",
			db: &DB{
				client: &Client{},
				driverDB: &mock.AttachmentMetaGetter{
					GetAttachmentMetaFunc: func(context.Context, string, string, driver.Options) (*driver.Attachment, error) {
						return nil, errors.New("fail")
					},
				},
			},
			docID:    "foo",
			filename: "foo.txt",
			status:   500,
			err:      "fail",
		},
		{
			name: "success",
			db: &DB{
				client: &Client{},
				driverDB: &mock.AttachmentMetaGetter{
					GetAttachmentMetaFunc: func(_ context.Context, docID, filename string, options driver.Options) (*driver.Attachment, error) {
						expectedDocID, expectedFilename := "foo", "foo.txt"
						if docID != expectedDocID {
							return nil, fmt.Errorf("Unexpected docID: %s", docID)
						}
						if filename != expectedFilename {
							return nil, fmt.Errorf("Unexpected filename: %s", filename)
						}
						gotOpts := map[string]interface{}{}
						options.Apply(gotOpts)
						if d := testy.DiffInterface(testOptions, gotOpts); d != nil {
							return nil, fmt.Errorf("Unexpected options:\n%s", d)
						}
						return &driver.Attachment{
							Filename:    "foo.txt",
							ContentType: "text/plain",
							Digest:      "md5-foo",
							Size:        4,
						}, nil
					},
				},
			},
			docID:    "foo",
			filename: "foo.txt",
			options:  Params(testOptions),
			expected: &Attachment{
				Filename:    "foo.txt",
				ContentType: "text/plain",
				Digest:      "md5-foo",
				Size:        4,
				Content:     nilContent,
			},
		},
		{
			name: "no doc id",
			db: &DB{
				client: &Client{},
			},
			status: http.StatusBadRequest,
			err:    "kivik: docID required",
		},
		{
			name: "no filename",
			db: &DB{
				client: &Client{},
			},
			docID:  "foo",
			status: http.StatusBadRequest,
			err:    "kivik: filename required",
		},
		{
			name: "client closed",
			db: &DB{
				client: &Client{
					closed: true,
				},
			},
			docID:    "foo",
			filename: "bar.txt",
			status:   http.StatusServiceUnavailable,
			err:      "kivik: client closed",
		},
		{
			name: "db error",
			db: &DB{
				err: errors.New("db error"),
			},
			status: http.StatusInternalServerError,
			err:    "db error",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.db.GetAttachmentMeta(context.Background(), test.docID, test.filename, test.options)
			if d := internal.StatusErrorDiff(test.err, test.status, err); d != "" {
				t.Error(d)
			}
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestPurge(t *testing.T) {
	type purgeTest struct {
		name   string
		db     *DB
		docMap map[string][]string

		expected *PurgeResult
		status   int
		err      string
	}

	docMap := map[string][]string{
		"foo": {"1-abc", "2-xyz"},
	}

	tests := []purgeTest{
		{
			name: "success, nothing purged",
			db: &DB{
				client: &Client{},
				driverDB: &mock.Purger{
					PurgeFunc: func(_ context.Context, dm map[string][]string) (*driver.PurgeResult, error) {
						if d := testy.DiffInterface(docMap, dm); d != nil {
							return nil, fmt.Errorf("Unexpected docmap: %s", d)
						}
						return &driver.PurgeResult{Seq: 2}, nil
					},
				},
			},
			docMap: docMap,
			expected: &PurgeResult{
				Seq: 2,
			},
		},
		{
			name: "success, all purged",
			db: &DB{
				client: &Client{},
				driverDB: &mock.Purger{
					PurgeFunc: func(_ context.Context, dm map[string][]string) (*driver.PurgeResult, error) {
						if d := testy.DiffInterface(docMap, dm); d != nil {
							return nil, fmt.Errorf("Unexpected docmap: %s", d)
						}
						return &driver.PurgeResult{Seq: 2, Purged: docMap}, nil
					},
				},
			},
			docMap: docMap,
			expected: &PurgeResult{
				Seq:    2,
				Purged: docMap,
			},
		},
		{
			name: "non-purger",
			db: &DB{
				client:   &Client{},
				driverDB: &mock.DB{},
			},
			status: http.StatusNotImplemented,
			err:    "kivik: purge not supported by driver",
		},
		{
			name: "couch 2.0-2.1 example",
			db: &DB{
				client: &Client{},
				driverDB: &mock.Purger{
					PurgeFunc: func(context.Context, map[string][]string) (*driver.PurgeResult, error) {
						return nil, &internal.Error{Status: http.StatusNotImplemented, Message: "this feature is not yet implemented"}
					},
				},
			},
			status: http.StatusNotImplemented,
			err:    "this feature is not yet implemented",
		},
		{
			name: "client closed",
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
				err: errors.New("db error"),
			},
			status: http.StatusInternalServerError,
			err:    "db error",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.db.Purge(context.Background(), test.docMap)
			if d := internal.StatusErrorDiff(test.err, test.status, err); d != "" {
				t.Error(d)
			}
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestBulkGet(t *testing.T) {
	type bulkGetTest struct {
		name    string
		db      *DB
		docs    []BulkGetReference
		options Option

		expected *ResultSet
		status   int
		err      string
	}

	tests := []bulkGetTest{
		{
			name: "non-bulkGetter",
			db: &DB{
				client:   &Client{},
				driverDB: &mock.DB{},
			},
			status: http.StatusNotImplemented,
			err:    "kivik: bulk get not supported by driver",
		},
		{
			name: "query error",
			db: &DB{
				client: &Client{},
				driverDB: &mock.BulkGetter{
					BulkGetFunc: func(context.Context, []driver.BulkGetReference, driver.Options) (driver.Rows, error) {
						return nil, errors.New("query error")
					},
				},
			},
			status: http.StatusInternalServerError,
			err:    "query error",
		},
		{
			name: "success",
			db: &DB{
				client: &Client{},
				driverDB: &mock.BulkGetter{
					BulkGetFunc: func(context.Context, []driver.BulkGetReference, driver.Options) (driver.Rows, error) {
						return &mock.Rows{ID: "bulkGet1"}, nil
					},
				},
			},
			expected: &ResultSet{
				iter: &iter{
					feed: &rowsIterator{
						Rows: &mock.Rows{ID: "bulkGet1"},
					},
					curVal: &driver.Row{},
				},
				rowsi: &mock.Rows{ID: "bulkGet1"},
			},
		},
		{
			name: "client closed",
			db: &DB{
				client: &Client{
					closed: true,
				},
				driverDB: &mock.BulkGetter{},
			},
			status: http.StatusServiceUnavailable,
			err:    "kivik: client closed",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rs := test.db.BulkGet(context.Background(), test.docs, test.options)
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
				driverDB: &mock.BulkGetter{
					BulkGetFunc: func(context.Context, []driver.BulkGetReference, driver.Options) (driver.Rows, error) {
						return nil, errors.New("unf")
					},
				},
			}
			rows := db.BulkGet(context.Background(), nil)
			if err := rows.Err(); err == nil {
				t.Fatal("expected an error, got none")
			}
			_ = db.Close() // Should not block
		})
	})
}

func newDB(db driver.DB) *DB {
	client := &Client{}
	client.wg.Add(1)
	return &DB{
		client:   client,
		driverDB: db,
	}
}

func TestDBClose(t *testing.T) {
	t.Parallel()

	type tst struct {
		db  *DB
		err string
	}
	tests := testy.NewTable()
	tests.Add("error", tst{
		db: newDB(&mock.DB{
			CloseFunc: func() error {
				return errors.New("close err")
			},
		}),
		err: "close err",
	})
	tests.Add("success", tst{
		db: newDB(&mock.DB{
			CloseFunc: func() error {
				return nil
			},
		}),
	})

	tests.Run(t, func(t *testing.T, test tst) {
		err := test.db.Close()
		if !testy.ErrorMatches(test.err, err) {
			t.Errorf("Unexpected error: %s", err)
		}
	})

	t.Run("blocks", func(t *testing.T) {
		t.Parallel()

		const delay = 100 * time.Millisecond

		type tt struct {
			db   driver.DB
			work func(*testing.T, *DB)
		}

		tests := testy.NewTable()
		tests.Add("Flush", tt{
			db: &mock.Flusher{
				FlushFunc: func(context.Context) error {
					time.Sleep(delay)
					return nil
				},
			},
			work: func(_ *testing.T, db *DB) {
				_ = db.Flush(context.Background())
			},
		})
		tests.Add("AllDocs", tt{
			db: &mock.DB{
				AllDocsFunc: func(context.Context, driver.Options) (driver.Rows, error) {
					return &mock.Rows{
						NextFunc: func(*driver.Row) error {
							time.Sleep(delay)
							return io.EOF
						},
					}, nil
				},
			},
			work: func(t *testing.T, db *DB) { //nolint:thelper // Not a helper
				u := db.AllDocs(context.Background())
				for u.Next() { //nolint:revive // intentional empty block
				}
				if u.Err() != nil {
					t.Fatal(u.Err())
				}
			},
		})
		tests.Add("BulkDocs", tt{
			db: &mock.BulkDocer{
				BulkDocsFunc: func(context.Context, []interface{}, driver.Options) ([]driver.BulkResult, error) {
					time.Sleep(delay)
					return []driver.BulkResult{}, nil
				},
			},
			work: func(t *testing.T, db *DB) { //nolint:thelper // Not a helper
				_, err := db.BulkDocs(context.Background(), []interface{}{
					map[string]string{"_id": "foo"},
				})
				if err != nil {
					t.Fatal(err)
				}
			},
		})

		tests.Run(t, func(t *testing.T, tt tt) {
			t.Parallel()

			db := &DB{
				client:   &Client{},
				driverDB: tt.db,
			}

			start := time.Now()
			tt.work(t, db)
			time.Sleep(delay / 2)
			_ = db.Close()
			if elapsed := time.Since(start); elapsed < delay {
				t.Errorf("db.Close() didn't block long enough (%v < %v)", elapsed, delay)
			}
		})
	})
}

func TestRevsDiff(t *testing.T) {
	type tt struct {
		db       *DB
		revMap   interface{}
		status   int
		err      string
		expected *ResultSet
	}
	tests := testy.NewTable()
	tests.Add("non-DBReplicator", tt{
		db: &DB{
			client:   &Client{},
			driverDB: &mock.DB{},
		},
		status: http.StatusNotImplemented,
		err:    "kivik: _revs_diff not supported by driver",
	})
	tests.Add("network error", tt{
		db: &DB{
			client: &Client{},
			driverDB: &mock.RevsDiffer{
				RevsDiffFunc: func(context.Context, interface{}) (driver.Rows, error) {
					return nil, errors.New("net error")
				},
			},
		},
		status: http.StatusInternalServerError,
		err:    "net error",
	})
	tests.Add("success", tt{
		db: &DB{
			client: &Client{},
			driverDB: &mock.RevsDiffer{
				RevsDiffFunc: func(context.Context, interface{}) (driver.Rows, error) {
					return &mock.Rows{ID: "a"}, nil
				},
			},
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
	})
	tests.Add("client closed", tt{
		db: &DB{
			client: &Client{
				closed: true,
			},
			driverDB: &mock.RevsDiffer{},
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
		rs := tt.db.RevsDiff(context.Background(), tt.revMap)
		err := rs.Err()
		if d := internal.StatusErrorDiff(tt.err, tt.status, err); d != "" {
			t.Error(d)
		}
		if err != nil {
			return
		}
		rs.cancel = nil  // Determinism
		rs.onClose = nil // Determinism
		if d := testy.DiffInterface(tt.expected, rs); d != nil {
			t.Error(d)
		}
	})
}

func TestPartitionStats(t *testing.T) {
	type tt struct {
		db     *DB
		name   string
		status int
		err    string
	}
	tests := testy.NewTable()
	tests.Add("non-PartitionedDB", tt{
		db: &DB{
			client:   &Client{},
			driverDB: &mock.DB{},
		},
		status: http.StatusNotImplemented,
		err:    "kivik: partitions not supported by driver",
	})
	tests.Add("error", tt{
		db: &DB{
			client: &Client{},
			driverDB: &mock.PartitionedDB{
				PartitionStatsFunc: func(context.Context, string) (*driver.PartitionStats, error) {
					return nil, &internal.Error{Status: http.StatusBadGateway, Err: errors.New("stats error")}
				},
			},
		},
		status: http.StatusBadGateway,
		err:    "stats error",
	})
	tests.Add("success", tt{
		db: &DB{
			client: &Client{},
			driverDB: &mock.PartitionedDB{
				PartitionStatsFunc: func(_ context.Context, name string) (*driver.PartitionStats, error) {
					if name != "partXX" {
						return nil, fmt.Errorf("Unexpected name: %s", name)
					}
					return &driver.PartitionStats{
						DBName:    "dbXX",
						Partition: name,
						DocCount:  123,
					}, nil
				},
			},
		},
		name: "partXX",
	})
	tests.Add("client closed", tt{
		db: &DB{
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
		result, err := tt.db.PartitionStats(context.Background(), tt.name)
		if d := internal.StatusErrorDiff(tt.err, tt.status, err); d != "" {
			t.Error(d)
		}
		if err != nil {
			return
		}
		if d := testy.DiffInterface(testy.Snapshot(t), result); d != nil {
			t.Error(d)
		}
	})
}
