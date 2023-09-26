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
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal/mock"
)

func TestRowsIteratorNext(t *testing.T) {
	const expected = "foo error"
	r := &rowsIterator{
		Rows: &mock.Rows{
			NextFunc: func(_ *driver.Row) error { return errors.New(expected) },
		},
	}
	var i driver.Row
	err := r.Next(&i)
	testy.Error(t, expected, err)
}

func TestNextResultSet(t *testing.T) {
	t.Run("two resultsets", func(t *testing.T) {
		r := multiResultSet()

		ids := []string{}
		for r.NextResultSet() {
			for r.Next() {
				id, _ := r.ID()
				ids = append(ids, id)
			}
		}
		if err := r.Err(); err != nil {
			t.Error(err)
		}
		want := []string{"1", "2", "3", "x", "y"}
		if d := testy.DiffInterface(want, ids); d != nil {
			t.Error(d)
		}
	})
	t.Run("called out of order", func(t *testing.T) {
		r := multiResultSet()

		if !r.Next() {
			t.Fatal("expected next to return true")
		}
		if r.NextResultSet() {
			t.Fatal("expected NextResultSet to return false")
		}

		wantErr := "must call NextResultSet before Next"
		err := r.Err()
		if !testy.ErrorMatches(wantErr, err) {
			t.Errorf("Unexpected error: %s", err)
		}
	})
	t.Run("next only", func(t *testing.T) {
		r := multiResultSet()

		ids := []string{}
		for r.Next() {
			id, _ := r.ID()
			ids = append(ids, id)
		}
		if err := r.Err(); err != nil {
			t.Error(err)
		}
		want := []string{"1", "2", "3", "x", "y"}
		if d := testy.DiffInterface(want, ids); d != nil {
			t.Error(d)
		}
	})
	t.Run("don't call NextResultSet in loop", func(t *testing.T) {
		r := multiResultSet()

		ids := []string{}
		r.NextResultSet()
		for r.Next() {
			id, _ := r.ID()
			ids = append(ids, id)
		}
		_ = r.Next() // once more to ensure it doesn't error past the end of the first RS
		if err := r.Err(); err != nil {
			t.Error(err)
		}

		// Only the first result set is processed, since NextResultSet is never
		// called a second time.
		want := []string{"1", "2", "3"}
		if d := testy.DiffInterface(want, ids); d != nil {
			t.Error(d)
		}
	})
}

func multiResultSet() *rows {
	rows := []interface{}{
		&driver.Row{ID: "1", Doc: strings.NewReader(`{"foo":"bar"}`)},
		&driver.Row{ID: "2", Doc: strings.NewReader(`{"foo":"bar"}`)},
		&driver.Row{ID: "3", Doc: strings.NewReader(`{"foo":"bar"}`)},
		int64(5),
		&driver.Row{ID: "x", Doc: strings.NewReader(`{"foo":"bar"}`)},
		&driver.Row{ID: "y", Doc: strings.NewReader(`{"foo":"bar"}`)},
		int64(2),
	}
	var offset int64

	return newRows(context.Background(), nil, &mock.Rows{
		NextFunc: func(r *driver.Row) error {
			if len(rows) == 0 {
				return io.EOF
			}
			row := rows[0]
			rows = rows[1:]
			switch t := row.(type) {
			case *driver.Row:
				*r = *t
				return nil
			case int64:
				offset = t
				return driver.EOQ
			default:
				panic("unknown type")
			}
		},
		OffsetFunc: func() int64 {
			return offset
		},
	})
}

func TestScanAllDocs(t *testing.T) {
	type tt struct {
		rows *rows
		dest interface{}
		err  string
	}

	tests := testy.NewTable()
	tests.Add("non-pointer dest", tt{
		dest: "string",
		err:  "must pass a pointer to ScanAllDocs",
	})
	tests.Add("nil pointer dest", tt{
		dest: (*string)(nil),
		err:  "nil pointer passed to ScanAllDocs",
	})
	tests.Add("not a slice or array", tt{
		dest: &rows{},
		err:  "dest must be a pointer to a slice or array",
	})
	tests.Add("0-length array", tt{
		dest: func() *[0]string { var x [0]string; return &x }(),
		err:  "0-length array passed to ScanAllDocs",
	})
	tests.Add("No docs to read", tt{
		rows: newRows(context.Background(), nil, &mock.Rows{}),
		dest: func() *[]string { return &[]string{} }(),
	})
	tests.Add("Success", func() interface{} {
		rows := []*driver.Row{
			{Doc: strings.NewReader(`{"foo":"bar"}`)},
		}
		return tt{
			rows: newRows(context.Background(), nil, &mock.Rows{
				NextFunc: func(r *driver.Row) error {
					if len(rows) == 0 {
						return io.EOF
					}
					*r = *rows[0]
					rows = rows[1:]
					return nil
				},
			}),
			dest: func() *[]json.RawMessage { return &[]json.RawMessage{} }(),
		}
	})
	tests.Add("Success, slice of pointers", func() interface{} {
		rows := []*driver.Row{
			{Doc: strings.NewReader(`{"foo":"bar"}`)},
		}
		return tt{
			rows: newRows(context.Background(), nil, &mock.Rows{
				NextFunc: func(r *driver.Row) error {
					if len(rows) == 0 {
						return io.EOF
					}
					*r = *rows[0]
					rows = rows[1:]
					return nil
				},
			}),
			dest: func() *[]*json.RawMessage { return &[]*json.RawMessage{} }(),
		}
	})
	tests.Add("Success, long array", func() interface{} {
		rows := []*driver.Row{
			{Doc: strings.NewReader(`{"foo":"bar"}`)},
		}
		return tt{
			rows: newRows(context.Background(), nil, &mock.Rows{
				NextFunc: func(r *driver.Row) error {
					if len(rows) == 0 {
						return io.EOF
					}
					*r = *rows[0]
					rows = rows[1:]
					return nil
				},
			}),
			dest: func() *[5]*json.RawMessage { return &[5]*json.RawMessage{} }(),
		}
	})
	tests.Add("Success, short array", func() interface{} {
		rows := []*driver.Row{
			{Doc: strings.NewReader(`{"foo":"bar"}`)},
			{Doc: strings.NewReader(`{"foo":"bar"}`)},
			{Doc: strings.NewReader(`{"foo":"bar"}`)},
		}
		return tt{
			rows: newRows(context.Background(), nil, &mock.Rows{
				NextFunc: func(r *driver.Row) error {
					if len(rows) == 0 {
						return io.EOF
					}
					*r = *rows[0]
					rows = rows[1:]
					return nil
				},
			}),
			dest: func() *[1]*json.RawMessage { return &[1]*json.RawMessage{} }(),
		}
	})
	tests.Run(t, func(t *testing.T, tt tt) {
		if tt.rows == nil {
			tt.rows = newRows(context.Background(), nil, &mock.Rows{})
		}
		rs := &ResultSet{
			underlying: tt.rows,
		}
		err := ScanAllDocs(rs, tt.dest)
		if !testy.ErrorMatches(tt.err, err) {
			t.Errorf("Unexpected error: %s", err)
		}
		if d := testy.DiffAsJSON(testy.Snapshot(t), tt.dest); d != nil {
			t.Error(d)
		}
	})
}

func TestResultSet_Next_resets_iterator_value(t *testing.T) {
	idx := 0
	rows := newRows(context.Background(), nil, &mock.Rows{
		NextFunc: func(r *driver.Row) error {
			idx++
			switch idx {
			case 1:
				r.ID = strconv.Itoa(idx)
				return nil
			case 2:
				return nil
			}
			return io.EOF
		},
	})

	wantIDs := []string{"1", ""}
	gotIDs := []string{}
	for rows.Next() {
		id, err := rows.ID()
		if err != nil {
			t.Fatal(err)
		}
		gotIDs = append(gotIDs, id)
	}
	if d := cmp.Diff(wantIDs, gotIDs); d != "" {
		t.Error(d)
	}
}
