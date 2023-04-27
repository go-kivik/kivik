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

func TestChangesNext(t *testing.T) {
	tests := []struct {
		name     string
		changes  *Changes
		expected bool
	}{
		{
			name: "nothing more",
			changes: &Changes{
				iter: &iter{closed: true},
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
	expected := "foo error" // nolint: goconst
	c := &Changes{
		iter: &iter{lasterr: errors.New(expected)},
	}
	err := c.Err()
	testy.Error(t, expected, err)
}

func TestChangesClose(t *testing.T) {
	expected := "close error"
	c := &Changes{
		iter: &iter{
			feed: &mockIterator{CloseFunc: func() error { return errors.New(expected) }},
		},
	}
	err := c.Close()
	testy.Error(t, expected, err)
}

func TestChangesIteratorNext(t *testing.T) {
	expected := "foo error"
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
	ch := newChanges(context.Background(), &mock.Changes{})
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
	c := &Changes{
		iter: &iter{
			curVal: &driver.Change{
				ID:      "foo",
				Deleted: true,
				Changes: []string{"1", "2", "3"},
				Seq:     "2-foo",
			},
		},
		changesi: &mock.Changes{
			PendingFunc: func() int64 { return 123 },
			LastSeqFunc: func() string { return "3-bar" },
			ETagFunc:    func() string { return "etag-foo" },
		},
	}

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
	t.Run("LastSeq", func(t *testing.T) {
		expected := "3-bar"
		result := c.LastSeq()
		if expected != result {
			t.Errorf("Unexpected result: %v", result)
		}
	})
	t.Run("Pending", func(t *testing.T) {
		expected := int64(123)
		result := c.Pending()
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
					closed: true,
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
				driverDB: &mock.DB{
					ChangesFunc: func(_ context.Context, _ map[string]interface{}) (driver.Changes, error) {
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
				driverDB: &mock.DB{
					ChangesFunc: func(_ context.Context, opts map[string]interface{}) (driver.Changes, error) {
						expectedOpts := map[string]interface{}{"foo": 123.4}
						if d := testy.DiffInterface(expectedOpts, opts); d != nil {
							return nil, fmt.Errorf("Unexpected options:\n%s", d)
						}
						return &mock.Changes{}, nil
					},
				},
			},
			opts: map[string]interface{}{"foo": 123.4},
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
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.db.Changes(context.Background(), test.opts)
			testy.StatusError(t, test.err, test.status, err)
			result.cancel = nil // Determinism
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestChanges_uninitialized_should_not_panic(*testing.T) {
	// These must not panic, because they can be called before iterating
	// begins.
	c := &Changes{}
	_ = c.LastSeq()
	_ = c.Pending()
	_ = c.ETag()
}
