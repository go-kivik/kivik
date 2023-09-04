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
	"io"
	"net/http"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal/mock"
)

func TestChangesNext(t *testing.T) {
	tests := []struct {
		name     string
		changes  *Changes
		expected bool
	}{
		{
			name: "nothing more",
			changes: &Changes{
				iter: &iter{state: stateClosed},
			},
			expected: false,
		},
		{
			name: "more",
			changes: &Changes{
				iter: &iter{
					feed: &mockIterator{
						NextFunc: func(_ interface{}) error { return nil },
					},
				},
			},
			expected: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.changes.Next()
			if result != test.expected {
				t.Errorf("Unexpected result: %v", result)
			}
		})
	}
}

func TestChangesErr(t *testing.T) {
	const expected = "foo error"
	c := &Changes{
		iter: &iter{err: errors.New(expected)},
	}
	err := c.Err()
	testy.Error(t, expected, err)
}

func TestChangesClose(t *testing.T) {
	const expected = "close error"
	c := &Changes{
		iter: &iter{
			feed: &mockIterator{CloseFunc: func() error { return errors.New(expected) }},
		},
	}
	err := c.Close()
	testy.Error(t, expected, err)
}

func TestChangesIteratorNext(t *testing.T) {
	const expected = "foo error"
	c := &changesIterator{
		Changes: &mock.Changes{
			NextFunc: func(_ *driver.Change) error { return errors.New(expected) },
		},
	}
	var i driver.Change
	err := c.Next(&i)
	testy.Error(t, expected, err)
}

func TestChangesIteratorNew(t *testing.T) {
	ch := newChanges(context.Background(), nil, &mock.Changes{})
	expected := &Changes{
		iter: &iter{
			feed: &changesIterator{
				Changes: &mock.Changes{},
			},
			curVal: &driver.Change{},
		},
		changesi: &mock.Changes{},
	}
	ch.cancel = nil // determinism
	if d := testy.DiffInterface(expected, ch); d != nil {
		t.Error(d)
	}
}

func TestChangesGetters(t *testing.T) {
	changes := []*driver.Change{
		{
			ID:      "foo",
			Deleted: true,
			Changes: []string{"1", "2", "3"},
			Seq:     "2-foo",
		},
	}
	c := newChanges(context.Background(), nil, &mock.Changes{
		NextFunc: func(c *driver.Change) error {
			if len(changes) == 0 {
				return io.EOF
			}
			change := changes[0]
			changes = changes[1:]
			*c = *change
			return nil
		},
		PendingFunc: func() int64 { return 123 },
		LastSeqFunc: func() string { return "3-bar" },
		ETagFunc:    func() string { return "etag-foo" },
	})
	_ = c.Next()

	t.Run("Changes", func(t *testing.T) {
		expected := []string{"1", "2", "3"}
		result := c.Changes()
		if d := testy.DiffInterface(expected, result); d != nil {
			t.Error(d)
		}
	})

	t.Run("Deleted", func(t *testing.T) {
		expected := true
		result := c.Deleted()
		if expected != result {
			t.Errorf("Unexpected result: %v", result)
		}
	})

	t.Run("ID", func(t *testing.T) {
		expected := "foo"
		result := c.ID()
		if expected != result {
			t.Errorf("Unexpected result: %v", result)
		}
	})
	t.Run("Seq", func(t *testing.T) {
		expected := "2-foo"
		result := c.Seq()
		if expected != result {
			t.Errorf("Unexpected result: %v", result)
		}
	})
	t.Run("ETag", func(t *testing.T) {
		expected := "etag-foo"
		result := c.ETag()
		if expected != result {
			t.Errorf("Unexpected result: %v", result)
		}
	})
	t.Run("Metadata", func(t *testing.T) {
		_ = c.Next()
		t.Run("LastSeq", func(t *testing.T) {
			expected := "3-bar"
			meta, err := c.Metadata()
			if err != nil {
				t.Fatal(err)
			}
			if expected != meta.LastSeq {
				t.Errorf("Unexpected LastSeq: %v", meta.LastSeq)
			}
		})
		t.Run("Pending", func(t *testing.T) {
			expected := int64(123)
			meta, err := c.Metadata()
			if err != nil {
				t.Fatal(err)
			}
			if expected != meta.Pending {
				t.Errorf("Unexpected Pending: %v", meta.Pending)
			}
		})
	})
}

func TestChangesScanDoc(t *testing.T) {
	tests := []struct {
		name     string
		changes  *Changes
		expected interface{}
		status   int
		err      string
	}{
		{
			name: "success",
			changes: &Changes{
				iter: &iter{
					state: stateRowReady,
					curVal: &driver.Change{
						Doc: []byte(`{"foo":123.4}`),
					},
				},
			},
			expected: map[string]interface{}{"foo": 123.4},
		},
		{
			name: "closed",
			changes: &Changes{
				iter: &iter{
					state: stateClosed,
				},
			},
			status: http.StatusBadRequest,
			err:    "kivik: Iterator is closed",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result interface{}
			err := test.changes.ScanDoc(&result)
			testy.StatusError(t, test.err, test.status, err)
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestChanges(t *testing.T) {
	tests := []struct {
		name     string
		db       *DB
		opts     Options
		expected *Changes
		status   int
		err      string
	}{
		{
			name: "db error",
			db: &DB{
				client: &Client{},
				driverDB: &mock.DB{
					ChangesFunc: func(_ context.Context, _ map[interface{}]interface{}) (driver.Changes, error) {
						return nil, errors.New("db error")
					},
				},
			},
			status: 500,
			err:    "db error",
		},
		{
			name: "success",
			db: &DB{
				client: &Client{},
				driverDB: &mock.DB{
					ChangesFunc: func(_ context.Context, opts map[interface{}]interface{}) (driver.Changes, error) {
						expectedOpts := map[interface{}]interface{}{"foo": 123.4}
						if d := testy.DiffInterface(expectedOpts, opts); d != nil {
							return nil, fmt.Errorf("Unexpected options:\n%s", d)
						}
						return &mock.Changes{}, nil
					},
				},
			},
			opts: map[interface{}]interface{}{"foo": 123.4},
			expected: &Changes{
				iter: &iter{
					feed: &changesIterator{
						Changes: &mock.Changes{},
					},
					curVal: &driver.Change{},
				},
				changesi: &mock.Changes{},
			},
		},
		{
			name: errClientClosed,
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
				err: errors.New("db error"),
			},
			status: http.StatusInternalServerError,
			err:    "db error",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.db.Changes(context.Background(), test.opts)
			testy.StatusError(t, test.err, test.status, result.Err())
			result.cancel = nil  // Determinism
			result.onClose = nil // Determinism
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
	t.Run("standalone", func(t *testing.T) {
		t.Run("after err, close doesn't block", func(t *testing.T) {
			db := &DB{
				client: &Client{},
				driverDB: &mock.DB{
					ChangesFunc: func(context.Context, map[interface{}]interface{}) (driver.Changes, error) {
						return nil, errors.New("unf")
					},
				},
			}
			rows := db.Changes(context.Background())
			if err := rows.Err(); err == nil {
				t.Fatal("expected an error, got none")
			}
			_ = db.Close() // Should not block
		})
	})
}

func TestChanges_uninitialized_should_not_panic(*testing.T) {
	// These must not panic, because they can be called before iterating
	// begins.
	c := &Changes{}
	_, _ = c.Metadata()
	_ = c.ETag()
}
