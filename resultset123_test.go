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

//go:build go1.23

package kivik

import (
	"errors"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/int/mock"
)

func TestResultSetIterator(t *testing.T) {
	t.Parallel()

	rows := []interface{}{
		&driver.Row{ID: "a"},
		&driver.Row{ID: "b"},
		&driver.Row{ID: "c"},
	}

	r := newResultSet(t.Context(), nil, &mock.Rows{
		NextFunc: func(r *driver.Row) error {
			if len(rows) == 0 {
				return io.EOF
			}
			if dr, ok := rows[0].(*driver.Row); ok {
				rows = rows[1:]
				*r = *dr
				return nil
			}
			return driver.EOQ
		},
	})

	want := []string{"a", "b", "c"}
	ids := []string{}

	for row, err := range r.Iterator() {
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		id, err := row.ID()
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		ids = append(ids, id)
	}
	if d := cmp.Diff(want, ids); d != "" {
		t.Errorf("Unexpected IDs: %s", d)
	}
}

func TestResultSetIteratorError(t *testing.T) {
	t.Parallel()

	r := newResultSet(t.Context(), nil, &mock.Rows{
		NextFunc: func(*driver.Row) error {
			return errors.New("failure")
		},
	})

	for _, err := range r.Iterator() {
		if err == nil {
			t.Fatal("expected error")
		}
		return
	}
	t.Fatal("Expected an error during iteration")
}

func TestResultSetIteratorBreak(t *testing.T) {
	t.Parallel()

	r := newResultSet(t.Context(), nil, &mock.Rows{
		NextFunc: func(*driver.Row) error {
			return nil
		},
	})

	for _, err := range r.Iterator() {
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		break
	}
	if r.state != stateClosed {
		t.Errorf("Expected iterator to be closed")
	}
}

func TestResultSetNextIterator(t *testing.T) {
	t.Parallel()
	r := multiResultSet()

	ids := []string{}
	for range r.NextIterator() {
		for row, err := range r.Iterator() {
			if err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}
			id, err := row.ID()
			if err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}
			ids = append(ids, id)
		}
	}

	if err := r.Err(); err != nil {
		t.Error(err)
	}
	want := []string{"1", "2", "3", "x", "y"}
	if d := cmp.Diff(want, ids); d != "" {
		t.Error(d)
	}

	for row, err := range r.Iterator() {
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		id, err := row.ID()
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		ids = append(ids, id)
	}
	if d := cmp.Diff(want, ids); d != "" {
		t.Errorf("Unexpected IDs: %s", d)
	}
}

func TestResultSetNextIteratorBreak(t *testing.T) {
	t.Parallel()

	r := newResultSet(t.Context(), nil, &mock.Rows{
		NextFunc: func(*driver.Row) error {
			return errors.New("failure")
		},
	})

	for range r.NextIterator() {
		break
	}

	if r.state != stateClosed {
		t.Errorf("Expected iterator to be closed")
	}
}
