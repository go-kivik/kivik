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
	expected := "foo error"
	r := &rows{
		iter: &iter{lasterr: errors.New(expected)},
	}
	err := r.Err()
	testy.Error(t, expected, err)
}

func TestRowsClose(t *testing.T) {
	expected := "close error"
	r := &rows{
		iter: &iter{
			feed: &mockIterator{CloseFunc: func() error { return errors.New(expected) }},
		},
	}
	err := r.Close()
	testy.Error(t, expected, err)
}

func TestRowsIteratorNext(t *testing.T) {
	expected := "foo error"
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
	tests.Add("prior error", tt{
		rows: &rows{
			err: errors.New("prev"),
		},
		status: http.StatusInternalServerError,
		err:    "prev",
	})
	tests.Add("success", tt{
		rows: &rows{
			iter: &iter{
				state: stateRowReady,
				curVal: &driver.Row{
					ValueReader: strings.NewReader(`{"foo":123.4}`),
				},
			},
		},
		state:    stateRowReady,
		expected: map[string]interface{}{"foo": 123.4},
	})
	tests.Add("one item", func() interface{} {
		rowsi := &mock.Rows{
			NextFunc: func(r *driver.Row) error {
				r.Value = []byte(`"foo"`)
				return nil
			},
		}
		return tt{
			rows:     newRows(context.Background(), rowsi),
			expected: "foo",
			state:    stateClosed,
		}
	})
	tests.Add("closed", tt{
		rows: &rows{
			iter: &iter{
				state: stateClosed,
				curVal: &driver.Row{
					ValueReader: strings.NewReader(`{"foo":123.4}`),
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
					Doc: []byte(`{"foo":123.4}`),
				},
			},
		},
		state:    stateRowReady,
		expected: map[string]interface{}{"foo": 123.4},
	})
	tests.Add("prev error", tt{
		rows: &rows{
			err: errors.New("flah"),
		},
		status: http.StatusInternalServerError,
		err:    "flah",
	})
	tests.Add("success", tt{
		rows: &rows{
			iter: &iter{
				state: stateRowReady,
				curVal: &driver.Row{
					DocReader: strings.NewReader(`{"foo":123.4}`),
				},
			},
		},
		state:    stateRowReady,
		expected: map[string]interface{}{"foo": 123.4},
	})
	tests.Add("one item", func() interface{} {
		rowsi := &mock.Rows{
			NextFunc: func(r *driver.Row) error {
				r.Doc = []byte(`{"foo":"bar"}`)
				return nil
			},
		}
		return tt{
			rows:     newRows(context.Background(), rowsi),
			expected: map[string]interface{}{"foo": "bar"},
			state:    stateClosed,
		}
	})
	tests.Add("closed", tt{
		rows: &rows{
			iter: &iter{
				state: stateClosed,
				curVal: &driver.Row{
					DocReader: strings.NewReader(`{"foo":123.4}`),
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
	tests.Add("prior error", tt{
		rows: &rows{
			err: errors.New("blahblah"),
		},
		status: http.StatusInternalServerError,
		err:    "blahblah",
	})
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
			rows:     newRows(context.Background(), rowsi),
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
					Error: errors.New("row error"),
				},
			},
		},
		status: 500,
		err:    "row error",
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		var result interface{}
		err := tt.rows.ScanKey(&result)
		testy.StatusError(t, tt.err, tt.status, err)
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
		result := r.ID()
		if id != result {
			t.Errorf("Unexpected result: %v", result)
		}
	})

	t.Run("Key", func(t *testing.T) {
		result := r.Key()
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
			r := newRows(context.Background(), rowsi)

			result := r.ID()
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
			r := newRows(context.Background(), rowsi)

			result := r.Key()
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
				feed: &rowsIterator{rowsi},
			},
			rowsi: rowsi,
		}

		if err := r.Close(); err != nil {
			t.Fatal(err)
		}

		t.Run("ID", func(t *testing.T) {
			result := r.ID()
			if id != result {
				t.Errorf("Unexpected result: %v", result)
			}
		})

		t.Run("Key", func(t *testing.T) {
			result := r.Key()
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

func TestQueryIndex(t *testing.T) {
	t.Run("QueryIndexer", func(t *testing.T) {
		expected := 100
		r := newRows(context.Background(), &mock.QueryIndexer{
			QueryIndexFunc: func() int { return expected },
		})
		if i := r.QueryIndex(); i != expected {
			t.Errorf("QueryIndex\nExpected %v\n  Actual: %v", expected, i)
		}
	})

	t.Run("Non QueryIndexer", func(t *testing.T) {
		r := newRows(context.Background(), &mock.Rows{})
		expected := 0
		if i := r.QueryIndex(); i != expected {
			t.Errorf("QueryIndex\nExpected: %v\n  Actual: %v", expected, i)
		}
	})
}

func TestFinishQuery(t *testing.T) {
	check := func(t *testing.T, r ResultSet) {
		t.Helper()
		meta, err := r.FinishQuery()
		if err != nil {
			t.Fatal(err)
		}
		if d := testy.DiffInterface(testy.Snapshot(t), meta); d != nil {
			t.Error(d)
		}
	}

	t.Run("Standard", func(t *testing.T) {
		r := newRows(context.Background(), &mock.Rows{
			OffsetFunc:    func() int64 { return 123 },
			TotalRowsFunc: func() int64 { return 234 },
			UpdateSeqFunc: func() string { return "seq" },
		})
		check(t, r)
	})
	t.Run("Bookmarker", func(t *testing.T) {
		expected := "test bookmark"
		r := newRows(context.Background(), &mock.Bookmarker{
			BookmarkFunc: func() string { return expected },
		})
		check(t, r)
	})
	t.Run("Warner", func(t *testing.T) {
		expected := "test warning"
		r := newRows(context.Background(), &mock.RowsWarner{
			WarningFunc: func() string { return expected },
		})
		check(t, r)
	})
	t.Run("query in progress", func(t *testing.T) {
		rows := []interface{}{
			&driver.Row{Doc: json.RawMessage(`{"foo":"bar"}`)},
			&driver.Row{Doc: json.RawMessage(`{"foo":"bar"}`)},
			&driver.Row{Doc: json.RawMessage(`{"foo":"bar"}`)},
		}

		r := newRows(context.Background(), &mock.Rows{
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
			&driver.Row{Doc: json.RawMessage(`{"foo":"bar"}`)},
			&driver.Row{Doc: json.RawMessage(`{"foo":"bar"}`)},
			&driver.Row{Doc: json.RawMessage(`{"foo":"bar"}`)},
		}

		r := newRows(context.Background(), &mock.Rows{
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
	t.Run("followed by other query", func(t *testing.T) {
		rows := []interface{}{
			&driver.Row{Doc: json.RawMessage(`{"foo":"bar"}`)},
			&driver.Row{Doc: json.RawMessage(`{"foo":"bar"}`)},
			&driver.Row{Doc: json.RawMessage(`{"foo":"bar"}`)},
			int64(5),
			&driver.Row{ID: "x", Doc: json.RawMessage(`{"foo":"bar"}`)},
			&driver.Row{ID: "y", Doc: json.RawMessage(`{"foo":"bar"}`)},
			int64(2),
		}
		var offset int64

		r := newRows(context.Background(), &mock.Rows{
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
		check(t, r)
		ids := []string{}
		for r.Next() {
			ids = append(ids, r.ID())
		}
		want := []string{"x", "y"}
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
		rows: newRows(context.Background(), &mock.Rows{}),
		dest: func() *[]string { return &[]string{} }(),
	})
	tests.Add("Success", func() interface{} {
		rows := []*driver.Row{
			{Doc: json.RawMessage(`{"foo":"bar"}`)},
		}
		return tt{
			rows: newRows(context.Background(), &mock.Rows{
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
			{Doc: json.RawMessage(`{"foo":"bar"}`)},
		}
		return tt{
			rows: newRows(context.Background(), &mock.Rows{
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
			{Doc: json.RawMessage(`{"foo":"bar"}`)},
		}
		return tt{
			rows: newRows(context.Background(), &mock.Rows{
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
			{Doc: json.RawMessage(`{"foo":"bar"}`)},
			{Doc: json.RawMessage(`{"foo":"bar"}`)},
			{Doc: json.RawMessage(`{"foo":"bar"}`)},
		}
		return tt{
			rows: newRows(context.Background(), &mock.Rows{
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
			tt.rows = newRows(context.Background(), &mock.Rows{})
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
		rows := []interface{}{
			&driver.Row{ID: "1", Doc: json.RawMessage(`{"foo":"bar"}`)},
			&driver.Row{ID: "2", Doc: json.RawMessage(`{"foo":"bar"}`)},
			&driver.Row{ID: "3", Doc: json.RawMessage(`{"foo":"bar"}`)},
			int64(5),
			&driver.Row{ID: "x", Doc: json.RawMessage(`{"foo":"bar"}`)},
			&driver.Row{ID: "y", Doc: json.RawMessage(`{"foo":"bar"}`)},
			int64(2),
		}
		var offset int64

		r := newRows(context.Background(), &mock.Rows{
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

		ids := []string{}
		for r.NextResultSet() {
			for r.Next() {
				ids = append(ids, r.ID())
			}
		}
		want := []string{"1", "2", "3", "x", "y"}
		if d := testy.DiffInterface(want, ids); d != nil {
			t.Error(d)
		}
	})
}
