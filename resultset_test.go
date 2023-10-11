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
	"strconv"
	"strings"
	"testing"
	"time"

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
	if !testy.ErrorMatches(expected, err) {
		t.Errorf("Unexpected error: %s", err)
	}
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

func multiResultSet() *ResultSet {
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

	return newResultSet(context.Background(), nil, &mock.Rows{
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
		rows *ResultSet
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
		dest: &ResultSet{},
		err:  "dest must be a pointer to a slice or array",
	})
	tests.Add("0-length array", tt{
		dest: func() *[0]string { var x [0]string; return &x }(),
		err:  "0-length array passed to ScanAllDocs",
	})
	tests.Add("No docs to read", tt{
		rows: newResultSet(context.Background(), nil, &mock.Rows{}),
		dest: func() *[]string { return &[]string{} }(),
	})
	tests.Add("Success", func() interface{} {
		rows := []*driver.Row{
			{Doc: strings.NewReader(`{"foo":"bar"}`)},
		}
		return tt{
			rows: newResultSet(context.Background(), nil, &mock.Rows{
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
			rows: newResultSet(context.Background(), nil, &mock.Rows{
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
			rows: newResultSet(context.Background(), nil, &mock.Rows{
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
			rows: newResultSet(context.Background(), nil, &mock.Rows{
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
			tt.rows = newResultSet(context.Background(), nil, &mock.Rows{})
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

func TestResultSet_Next_resets_iterator_value(t *testing.T) {
	idx := 0
	rows := newResultSet(context.Background(), nil, &mock.Rows{
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

func TestResultSet_Getters(t *testing.T) {
	const id = "foo"
	key := []byte("[1234]")
	const offset = int64(2)
	const totalrows = int64(3)
	const updateseq = "asdfasdf"
	r := &ResultSet{
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
			r := newResultSet(context.Background(), nil, rowsi)

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
			r := newResultSet(context.Background(), nil, rowsi)

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
		r := &ResultSet{
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

func TestResultSet_Metadata(t *testing.T) {
	t.Run("iteration incomplete", func(t *testing.T) {
		r := newResultSet(context.Background(), nil, &mock.Rows{
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

	check := func(t *testing.T, r *ResultSet) {
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
		r := newResultSet(context.Background(), nil, &mock.Rows{
			OffsetFunc:    func() int64 { return 123 },
			TotalRowsFunc: func() int64 { return 234 },
			UpdateSeqFunc: func() string { return "seq" },
		})
		check(t, r)
	})
	t.Run("Bookmarker", func(t *testing.T) {
		expected := "test bookmark"
		r := newResultSet(context.Background(), nil, &mock.Bookmarker{
			BookmarkFunc: func() string { return expected },
		})
		check(t, r)
	})
	t.Run("Warner", func(t *testing.T) {
		const expected = "test warning"
		r := newResultSet(context.Background(), nil, &mock.RowsWarner{
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

		r := newResultSet(context.Background(), nil, &mock.Rows{
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

		r := newResultSet(context.Background(), nil, &mock.Rows{
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

func Test_bug576(t *testing.T) {
	rows := newResultSet(context.Background(), nil, &mock.Rows{
		NextFunc: func(*driver.Row) error {
			return io.EOF
		},
	})

	var result interface{}
	err := rows.ScanDoc(&result)
	const wantErr = "no results"
	wantStatus := http.StatusNotFound
	if !testy.ErrorMatches(wantErr, err) {
		t.Errorf("unexpected error: %s", err)
	}
	if status := HTTPStatus(err); status != wantStatus {
		t.Errorf("Unexpected error status: %v", status)
	}
}

func TestResultSet_single(t *testing.T) {
	const wantRev = "1-abc"
	const docContent = `{"_id":"foo"}`
	wantDoc := map[string]interface{}{"_id": "foo"}
	rows := newResultSet(context.Background(), nil, &mock.Rows{
		NextFunc: func(row *driver.Row) error {
			row.Rev = wantRev
			row.Doc = strings.NewReader(docContent)
			return nil
		},
	})

	rev, err := rows.Rev()
	if err != nil {
		t.Fatal(err)
	}
	if rev != wantRev {
		t.Errorf("Unexpected rev: %s", rev)
	}
	var doc map[string]interface{}
	if err := rows.ScanDoc(&doc); err != nil {
		t.Fatal(err)
	}
	if d := cmp.Diff(wantDoc, doc); d != "" {
		t.Error(d)
	}
	rev2, err := rows.Rev()
	if err != nil {
		t.Fatal(err)
	}
	if rev2 != wantRev {
		t.Errorf("Unexpected rev on second read: %s", rev2)
	}
}

func TestResultSet_Close_blocks(t *testing.T) {
	t.Parallel()

	const delay = 100 * time.Millisecond

	type tt struct {
		rows driver.Rows
		work func(*ResultSet)
	}

	tests := testy.NewTable()
	tests.Add("ScanDoc", tt{
		rows: &mock.Rows{
			NextFunc: func(row *driver.Row) error {
				row.Doc = io.MultiReader(
					testy.DelayReader(delay),
					strings.NewReader(`{}`),
				)
				return nil
			},
		},
		work: func(rs *ResultSet) {
			var i interface{}
			_ = rs.ScanDoc(&i)
		},
	})
	tests.Add("ScanValue", tt{
		rows: &mock.Rows{
			NextFunc: func(row *driver.Row) error {
				row.Value = io.MultiReader(
					testy.DelayReader(delay),
					strings.NewReader(`{}`),
				)
				return nil
			},
		},
		work: func(rs *ResultSet) {
			var i interface{}
			_ = rs.ScanValue(&i)
		},
	})
	tests.Add("Attachments", tt{
		rows: &mock.Rows{
			NextFunc: func(row *driver.Row) error {
				row.Attachments = &mock.Attachments{
					NextFunc: func(*driver.Attachment) error {
						time.Sleep(delay)
						return io.EOF
					},
				}
				return nil
			},
		},
		work: func(rs *ResultSet) {
			atts, err := rs.Attachments()
			if err != nil {
				t.Fatal(err)
			}
			for {
				_, _ = atts.Next()
			}
		},
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		t.Parallel()

		rs := newResultSet(context.Background(), nil, tt.rows)

		start := time.Now()
		go tt.work(rs)
		time.Sleep(delay / 2)
		_ = rs.Close()
		if elapsed := time.Since(start); elapsed < delay {
			t.Errorf("rs.Close() didn't block long enouggh (%v < %v)", elapsed, delay)
		}
	})
}
