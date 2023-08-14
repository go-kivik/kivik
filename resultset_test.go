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
	"net/http"
	"strings"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal/mock"
)

func TestRowsNext(t *testing.T) {
	tests := []struct {
		name     string
		rows     *rows
		expected bool
	}{
		{
			name: "nothing more",
			rows: &rows{
				iter: &iter{state: stateClosed},
			},
			expected: false,
		},
		{
			name: "more",
			rows: &rows{
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
			result := test.rows.Next()
			if result != test.expected {
				t.Errorf("Unexpected result: %v", result)
			}
		})
	}
}

func TestRowsErr(t *testing.T) {
	const expected = "foo error"
	r := &rows{
		iter: &iter{err: errors.New(expected)},
	}
	err := r.Err()
	testy.Error(t, expected, err)
}

func TestRowsClose(t *testing.T) {
	const expected = "close error"
	r := &rows{
		iter: &iter{
			feed: &mockIterator{CloseFunc: func() error { return errors.New(expected) }},
		},
	}
	err := r.Close()
	testy.Error(t, expected, err)
}

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

func TestRowsScanValue(t *testing.T) {
	type tt struct {
		rows     *rows
		expected interface{}
		state    int
		status   int
		err      string
	}

	tests := testy.NewTable()
	tests.Add("success", tt{
		rows: &rows{
			iter: &iter{
				state: stateRowReady,
				curVal: &driver.Row{
					Value: strings.NewReader(`{"foo":123.4}`),
				},
			},
		},
		state:    stateRowReady,
		expected: map[string]interface{}{"foo": 123.4},
	})
	tests.Add("one item", func() interface{} {
		rowsi := &mock.Rows{
			NextFunc: func(r *driver.Row) error {
				r.Value = strings.NewReader(`"foo"`)
				return nil
			},
		}
		return tt{
			rows:     newRows(context.Background(), nil, rowsi),
			expected: "foo",
			state:    stateClosed,
		}
	})
	tests.Add("closed", tt{
		rows: &rows{
			iter: &iter{
				state: stateClosed,
				curVal: &driver.Row{
					Value: strings.NewReader(`{"foo":123.4}`),
				},
			},
		},
		expected: map[string]interface{}{"foo": 123.4},
		state:    stateClosed,
	})
	tests.Add("row error", tt{
		rows: &rows{
			iter: &iter{
				state: stateRowReady,
				curVal: &driver.Row{
					Error: errors.New("row error"),
				},
			},
		},
		status: 500,
		err:    "row error",
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		var result interface{}
		err := tt.rows.ScanValue(&result)
		testy.StatusError(t, tt.err, tt.status, err)
		if d := testy.DiffInterface(tt.expected, result); d != nil {
			t.Error(d)
		}
		if tt.state != tt.rows.state {
			t.Errorf("Unexpected state: %v", tt.rows.state)
		}
	})
}

func TestRowsScanDoc(t *testing.T) {
	type tt struct {
		rows     *rows
		expected interface{}
		state    int
		status   int
		err      string
	}

	tests := testy.NewTable()

	tests.Add("old row", tt{
		rows: &rows{
			iter: &iter{
				state: stateRowReady,
				curVal: &driver.Row{
					Doc: strings.NewReader(`{"foo":123.4}`),
				},
			},
		},
		state:    stateRowReady,
		expected: map[string]interface{}{"foo": 123.4},
	})
	tests.Add("success", tt{
		rows: &rows{
			iter: &iter{
				state: stateRowReady,
				curVal: &driver.Row{
					Doc: strings.NewReader(`{"foo":123.4}`),
				},
			},
		},
		state:    stateRowReady,
		expected: map[string]interface{}{"foo": 123.4},
	})
	tests.Add("one item", func() interface{} {
		rowsi := &mock.Rows{
			NextFunc: func(r *driver.Row) error {
				r.Doc = strings.NewReader(`{"foo":"bar"}`)
				return nil
			},
		}
		return tt{
			rows:     newRows(context.Background(), nil, rowsi),
			expected: map[string]interface{}{"foo": "bar"},
			state:    stateClosed,
		}
	})
	tests.Add("closed", tt{
		rows: &rows{
			iter: &iter{
				state: stateClosed,
				curVal: &driver.Row{
					Doc: strings.NewReader(`{"foo":123.4}`),
				},
			},
		},
		state:    stateClosed,
		expected: map[string]interface{}{"foo": 123.4},
	})
	tests.Add("nil doc", tt{
		rows: &rows{
			iter: &iter{
				state: stateRowReady,
				curVal: &driver.Row{
					Doc: nil,
				},
			},
		},
		status: http.StatusBadRequest,
		err:    "kivik: doc is nil; does the query include docs?",
	})
	tests.Add("row error", tt{
		rows: &rows{
			iter: &iter{
				state: stateRowReady,
				curVal: &driver.Row{
					Error: errors.New("row error"),
				},
			},
		},
		status: 500,
		err:    "row error",
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		var result interface{}
		err := tt.rows.ScanDoc(&result)
		testy.StatusError(t, tt.err, tt.status, err)
		if d := testy.DiffInterface(tt.expected, result); d != nil {
			t.Error(d)
		}
		if tt.state != tt.rows.state {
			t.Errorf("Unexpected state: %v", tt.rows.state)
		}
	})
}

func TestRowsScanKey(t *testing.T) {
	type tt struct {
		rows     *rows
		expected interface{}
		state    int
		status   int
		err      string
	}

	tests := testy.NewTable()
	tests.Add("success", tt{
		rows: &rows{
			iter: &iter{
				state: stateRowReady,
				curVal: &driver.Row{
					Key: []byte(`{"foo":123.4}`),
				},
			},
		},
		state:    stateRowReady,
		expected: map[string]interface{}{"foo": 123.4},
	})
	tests.Add("one item", func() interface{} {
		rowsi := &mock.Rows{
			NextFunc: func(r *driver.Row) error {
				r.Key = []byte(`"foo"`)
				return nil
			},
		}
		return tt{
			rows:     newRows(context.Background(), nil, rowsi),
			expected: "foo",
			state:    stateClosed,
		}
	})
	tests.Add("closed", tt{
		rows: &rows{
			iter: &iter{
				state: stateClosed,
				curVal: &driver.Row{
					Key: []byte(`"foo"`),
				},
			},
		},
		state:    stateClosed,
		expected: "foo",
	})
	tests.Add("row error", tt{
		rows: &rows{
			iter: &iter{
				state: stateRowReady,
				curVal: &driver.Row{
					Key:   json.RawMessage(`"id"`),
					Error: errors.New("row error"),
				},
			},
		},
		expected: "id",
		state:    stateRowReady,
		status:   http.StatusInternalServerError,
		err:      "row error",
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		var result interface{}
		err := tt.rows.ScanKey(&result)
		if !testy.ErrorMatches(tt.err, err) {
			t.Errorf("unexpected error: %s", err)
		}
		if status := HTTPStatus(err); status != tt.status {
			t.Errorf("Unexpected error status: %v", status)
		}
		if d := testy.DiffInterface(tt.expected, result); d != nil {
			t.Error(d)
		}
		if tt.state != tt.rows.state {
			t.Errorf("Unexpected state: %v", tt.rows.state)
		}
	})
}

func TestRowsGetters(t *testing.T) {
	id := "foo"
	key := []byte("[1234]")
	offset := int64(2)
	totalrows := int64(3)
	updateseq := "asdfasdf"
	r := &rows{
		iter: &iter{
			state: stateRowReady,
			curVal: &driver.Row{
				ID:  id,
				Key: key,
			},
		},
		rowsi: &mock.Rows{
			OffsetFunc:    func() int64 { return offset },
			TotalRowsFunc: func() int64 { return totalrows },
			UpdateSeqFunc: func() string { return updateseq },
		},
	}

	t.Run("ID", func(t *testing.T) {
		result, _ := r.ID()
		if id != result {
			t.Errorf("Unexpected result: %v", result)
		}
	})

	t.Run("Key", func(t *testing.T) {
		result, _ := r.Key()
		if string(key) != result {
			t.Errorf("Unexpected result: %v", result)
		}
	})

	t.Run("Not Ready", func(t *testing.T) {
		t.Run("ID", func(t *testing.T) {
			rowsi := &mock.Rows{
				NextFunc: func(r *driver.Row) error {
					r.ID = id
					return nil
				},
			}
			r := newRows(context.Background(), nil, rowsi)

			result, _ := r.ID()
			if result != id {
				t.Errorf("Unexpected result: %v", result)
			}
		})

		t.Run("Key", func(t *testing.T) {
			rowsi := &mock.Rows{
				NextFunc: func(r *driver.Row) error {
					r.Key = key
					return nil
				},
			}
			r := newRows(context.Background(), nil, rowsi)

			result, _ := r.Key()
			if result != string(key) {
				t.Errorf("Unexpected result: %v", result)
			}
		})
	})

	t.Run("after close", func(t *testing.T) {
		rowsi := &mock.Rows{
			OffsetFunc:    func() int64 { return offset },
			TotalRowsFunc: func() int64 { return totalrows },
			UpdateSeqFunc: func() string { return updateseq },
		}
		r := &rows{
			iter: &iter{
				state: stateRowReady,
				curVal: &driver.Row{
					ID:  id,
					Key: key,
				},
				feed: &rowsIterator{Rows: rowsi},
			},
			rowsi: rowsi,
		}

		if err := r.Close(); err != nil {
			t.Fatal(err)
		}

		t.Run("ID", func(t *testing.T) {
			result, _ := r.ID()
			if id != result {
				t.Errorf("Unexpected result: %v", result)
			}
		})

		t.Run("Key", func(t *testing.T) {
			result, _ := r.Key()
			if string(key) != result {
				t.Errorf("Unexpected result: %v", result)
			}
		})

		t.Run("ScanKey", func(t *testing.T) {
			var result json.RawMessage
			err := r.ScanKey(&result)
			if err != nil {
				t.Fatal(err)
			}
			if string(result) != string(key) {
				t.Errorf("Unexpected result: %v", result)
			}
		})
	})
}

func TestMetadata(t *testing.T) {
	t.Run("iteration incomplete", func(t *testing.T) {
		r := newRows(context.Background(), nil, &mock.Rows{
			OffsetFunc:    func() int64 { return 123 },
			TotalRowsFunc: func() int64 { return 234 },
			UpdateSeqFunc: func() string { return "seq" },
		})
		_, err := r.Metadata()
		wantErr := "Metadata must not be called until result set iteration is complete"
		if !testy.ErrorMatches(wantErr, err) {
			t.Errorf("Unexpected error: %s", err)
		}
	})

	check := func(t *testing.T, r ResultSetX) {
		t.Helper()
		for r.Next() { //nolint:revive // Consume all rows
		}
		meta, err := r.Metadata()
		if err != nil {
			t.Fatal(err)
		}
		if d := testy.DiffInterface(testy.Snapshot(t), meta); d != nil {
			t.Error(d)
		}
	}

	t.Run("Standard", func(t *testing.T) {
		r := newRows(context.Background(), nil, &mock.Rows{
			OffsetFunc:    func() int64 { return 123 },
			TotalRowsFunc: func() int64 { return 234 },
			UpdateSeqFunc: func() string { return "seq" },
		})
		check(t, r)
	})
	t.Run("Bookmarker", func(t *testing.T) {
		expected := "test bookmark"
		r := newRows(context.Background(), nil, &mock.Bookmarker{
			BookmarkFunc: func() string { return expected },
		})
		check(t, r)
	})
	t.Run("Warner", func(t *testing.T) {
		expected := "test warning"
		r := newRows(context.Background(), nil, &mock.RowsWarner{
			WarningFunc: func() string { return expected },
		})
		check(t, r)
	})
	t.Run("query in progress", func(t *testing.T) {
		rows := []interface{}{
			&driver.Row{Doc: strings.NewReader(`{"foo":"bar"}`)},
			&driver.Row{Doc: strings.NewReader(`{"foo":"bar"}`)},
			&driver.Row{Doc: strings.NewReader(`{"foo":"bar"}`)},
		}

		r := newRows(context.Background(), nil, &mock.Rows{
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
			OffsetFunc: func() int64 {
				return 5
			},
		})
		var i int
		for r.Next() {
			i++
			if i > 10 {
				panic(i)
			}
		}
		check(t, r)
	})
	t.Run("no query in progress", func(t *testing.T) {
		rows := []interface{}{
			&driver.Row{Doc: strings.NewReader(`{"foo":"bar"}`)},
			&driver.Row{Doc: strings.NewReader(`{"foo":"bar"}`)},
			&driver.Row{Doc: strings.NewReader(`{"foo":"bar"}`)},
		}

		r := newRows(context.Background(), nil, &mock.Rows{
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
			OffsetFunc: func() int64 {
				return 5
			},
		})
		check(t, r)
	})
	t.Run("followed by other query in resultset mode", func(t *testing.T) {
		r := multiResultSet()

		_ = r.NextResultSet()
		check(t, r)
		ids := []string{}
		for r.Next() {
			id, _ := r.ID()
			ids = append(ids, id)
		}
		want := []string{"x", "y"}
		if d := testy.DiffInterface(want, ids); d != nil {
			t.Error(d)
		}
		t.Run("second query", func(t *testing.T) {
			check(t, r)
		})
	})
	t.Run("followed by other query in row mode", func(t *testing.T) {
		r := multiResultSet()

		check(t, r)
		ids := []string{}
		for r.Next() {
			id, _ := r.ID()
			ids = append(ids, id)
		}
		want := []string{}
		if d := testy.DiffInterface(want, ids); d != nil {
			t.Error(d)
		}
		t.Run("second query", func(t *testing.T) {
			check(t, r)
		})
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
		err := ScanAllDocs(tt.rows, tt.dest)
		if !testy.ErrorMatches(tt.err, err) {
			t.Errorf("Unexpected error: %s", err)
		}
		if d := testy.DiffAsJSON(testy.Snapshot(t), tt.dest); d != nil {
			t.Error(d)
		}
	})
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

func multiResultSet() ResultSetX {
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

func Test_bug576(t *testing.T) {
	rows := newRows(context.Background(), nil, &mock.Rows{
		NextFunc: func(*driver.Row) error {
			return io.EOF
		},
	})

	var result interface{}
	err := rows.ScanDoc(&result)
	wantErr := "no results"
	wantStatus := http.StatusNotFound
	if !testy.ErrorMatches(wantErr, err) {
		t.Errorf("unexpected error: %s", err)
	}
	if status := HTTPStatus(err); status != wantStatus {
		t.Errorf("Unexpected error status: %v", status)
	}
}
