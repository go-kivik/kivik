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
				iter: &iter{closed: true},
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
	tests := []struct {
		name     string
		rows     *rows
		expected interface{}
		status   int
		err      string
	}{
		{
			name: "success",
			rows: &rows{
				iter: &iter{
					ready: true,
					curVal: &driver.Row{
						ValueReader: strings.NewReader(`{"foo":123.4}`),
					},
				},
			},
			expected: map[string]interface{}{"foo": 123.4},
		},
		{
			name: "closed",
			rows: &rows{
				iter: &iter{
					closed: true,
					ready:  true,
					curVal: &driver.Row{
						ValueReader: strings.NewReader(`{"foo":123.4}`),
					},
				},
			},
			expected: map[string]interface{}{"foo": 123.4},
		},
		{
			name: "row error",
			rows: &rows{
				iter: &iter{
					ready: true,
					curVal: &driver.Row{
						Error: errors.New("row error"),
					},
				},
			},
			status: 500,
			err:    "row error",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result interface{}
			err := test.rows.ScanValue(&result)
			testy.StatusError(t, test.err, test.status, err)
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestRowsScanDoc(t *testing.T) {
	tests := []struct {
		name     string
		rows     *rows
		expected interface{}
		status   int
		err      string
	}{
		{
			name: "old row",
			rows: &rows{
				iter: &iter{
					ready: true,
					curVal: &driver.Row{
						Doc: []byte(`{"foo":123.4}`),
					},
				},
			},
			expected: map[string]interface{}{"foo": 123.4},
		},
		{
			name: "success",
			rows: &rows{
				iter: &iter{
					ready: true,
					curVal: &driver.Row{
						DocReader: strings.NewReader(`{"foo":123.4}`),
					},
				},
			},
			expected: map[string]interface{}{"foo": 123.4},
		},
		{
			name: "closed",
			rows: &rows{
				iter: &iter{
					closed: true,
					ready:  true,
					curVal: &driver.Row{
						DocReader: strings.NewReader(`{"foo":123.4}`),
					},
				},
			},
			expected: map[string]interface{}{"foo": 123.4},
		},
		{
			name: "nil doc",
			rows: &rows{
				iter: &iter{
					ready: true,
					curVal: &driver.Row{
						Doc: nil,
					},
				},
			},
			status: http.StatusBadRequest,
			err:    "kivik: doc is nil; does the query include docs?",
		},
		{
			name: "row error",
			rows: &rows{
				iter: &iter{
					ready: true,
					curVal: &driver.Row{
						Error: errors.New("row error"),
					},
				},
			},
			status: 500,
			err:    "row error",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result interface{}
			err := test.rows.ScanDoc(&result)
			testy.StatusError(t, test.err, test.status, err)
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestRowsScanKey(t *testing.T) {
	tests := []struct {
		name     string
		rows     *rows
		expected interface{}
		status   int
		err      string
	}{
		{
			name: "success",
			rows: &rows{
				iter: &iter{
					ready: true,
					curVal: &driver.Row{
						Key: []byte(`{"foo":123.4}`),
					},
				},
			},
			expected: map[string]interface{}{"foo": 123.4},
		},
		{
			name: "closed",
			rows: &rows{
				iter: &iter{
					closed: true,
					ready:  true,
					curVal: &driver.Row{
						Key: []byte(`"foo"`),
					},
				},
			},
			expected: "foo",
		},
		{
			name: "row error",
			rows: &rows{
				iter: &iter{
					ready: true,
					curVal: &driver.Row{
						Error: errors.New("row error"),
					},
				},
			},
			status: 500,
			err:    "row error",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result interface{}
			err := test.rows.ScanKey(&result)
			testy.StatusError(t, test.err, test.status, err)
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestRowsGetters(t *testing.T) {
	id := "foo"
	key := []byte("[1234]")
	offset := int64(2)
	totalrows := int64(3)
	updateseq := "asdfasdf"
	r := &rows{
		iter: &iter{
			ready: true,
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

	t.Run("Offset", func(t *testing.T) {
		result := r.Offset()
		if offset != result {
			t.Errorf("Unexpected result: %v", result)
		}
	})

	t.Run("UpdateSeq", func(t *testing.T) {
		result := r.UpdateSeq()
		if updateseq != result {
			t.Errorf("Unexpected result: %v", result)
		}
	})

	t.Run("Not Ready", func(t *testing.T) {
		r := &rows{
			iter: &iter{
				ready: false,
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
			if result != "" {
				t.Errorf("Unexpected result: %v", result)
			}
		})

		t.Run("Key", func(t *testing.T) {
			result := r.Key()
			if result != "" {
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
				ready: true,
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

		t.Run("Offset", func(t *testing.T) {
			result := r.Offset()
			if offset != result {
				t.Errorf("Unexpected result: %v", result)
			}
		})

		t.Run("UpdateSeq", func(t *testing.T) {
			result := r.UpdateSeq()
			if updateseq != result {
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

func TestFinish(t *testing.T) {
	check := func(t *testing.T, r Rows) {
		t.Helper()
		meta, err := r.Finish()
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
		err := tt.rows.ScanAllDocs(tt.dest)
		if !testy.ErrorMatches(tt.err, err) {
			t.Errorf("Unexpected error: %s", err)
		}
		if d := testy.DiffAsJSON(testy.Snapshot(t), tt.dest); d != nil {
			t.Error(d)
		}
	})
}
