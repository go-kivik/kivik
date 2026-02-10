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

//go:build !js

package sqlite

import (
	"context"
	"io"
	"net/http"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/int/mock"
)

type rowResult struct {
	ID    string
	Rev   string
	Key   string
	Value string
	Doc   string
	Error string
}

func TestDBAllDocs(t *testing.T) {
	t.Parallel()
	type test struct {
		db         driver.DB
		options    driver.Options
		want       []rowResult
		wantStatus int
		wantErr    string
	}
	tests := testy.NewTable()
	tests.Add("no docs in db", test{
		want: nil,
	})
	tests.Add("single doc", func(t *testing.T) any {
		db := newDB(t)
		rev := db.tPut("foo", map[string]string{"cat": "meow"})

		return test{
			db: db,
			want: []rowResult{
				{
					ID:    "foo",
					Key:   `"foo"`,
					Value: `{"rev":"` + rev + `"}`,
				},
			},
		}
	})
	tests.Add("include_docs=true", func(t *testing.T) any {
		db := newDB(t)
		rev := db.tPut("foo", map[string]string{"cat": "meow"})

		return test{
			db:      db,
			options: kivik.Param("include_docs", true),
			want: []rowResult{
				{
					ID:    "foo",
					Key:   `"foo"`,
					Value: `{"rev":"` + rev + `"}`,
					Doc:   `{"_id":"foo","_rev":"` + rev + `","cat":"meow"}`,
				},
			},
		}
	})
	tests.Add("single doc multiple revisions", func(t *testing.T) any {
		db := newDB(t)
		rev := db.tPut("foo", map[string]string{"cat": "meow"})
		rev2 := db.tPut("foo", map[string]string{"cat": "purr"}, kivik.Rev(rev))

		return test{
			db: db,
			want: []rowResult{
				{
					ID:    "foo",
					Key:   `"foo"`,
					Value: `{"rev":"` + rev2 + `"}`,
				},
			},
		}
	})
	tests.Add("conflicting document, select winning rev", func(t *testing.T) any {
		db := newDB(t)
		_ = db.tPut("foo", map[string]string{
			"cat":  "meow",
			"_rev": "1-xxx",
		}, kivik.Param("new_edits", false))
		_ = db.tPut("foo", map[string]string{
			"cat":  "purr",
			"_rev": "1-aaa",
		}, kivik.Param("new_edits", false))

		return test{
			db: db,
			want: []rowResult{
				{
					ID:    "foo",
					Key:   `"foo"`,
					Value: `{"rev":"1-xxx"}`,
				},
			},
		}
	})
	tests.Add("deleted doc", func(t *testing.T) any {
		db := newDB(t)
		rev := db.tPut("foo", map[string]string{"cat": "meow"})
		_ = db.tDelete("foo", kivik.Rev(rev))

		return test{
			db:   db,
			want: nil,
		}
	})
	tests.Add("select lower revision number when higher rev in winning branch has been deleted", func(t *testing.T) any {
		db := newDB(t)
		_ = db.tPut("foo", map[string]string{
			"cat":  "meow",
			"_rev": "1-xxx",
		}, kivik.Param("new_edits", false))
		_ = db.tPut("foo", map[string]string{
			"cat":  "purr",
			"_rev": "1-aaa",
		}, kivik.Param("new_edits", false))
		_ = db.tDelete("foo", kivik.Rev("1-aaa"))

		return test{
			db: db,
			want: []rowResult{
				{
					ID:    "foo",
					Key:   `"foo"`,
					Value: `{"rev":"1-xxx"}`,
				},
			},
		}
	})
	tests.Add("conflicts=true", func(t *testing.T) any {
		db := newDB(t)
		_ = db.tPut("foo", map[string]string{
			"cat":  "meow",
			"_rev": "1-xxx",
		}, kivik.Param("new_edits", false))
		_ = db.tPut("foo", map[string]string{
			"cat":  "purr",
			"_rev": "1-aaa",
		}, kivik.Param("new_edits", false))

		return test{
			db: db,
			options: kivik.Params(map[string]any{
				"conflicts":    true,
				"include_docs": true,
			}),
			want: []rowResult{
				{
					ID:    "foo",
					Key:   `"foo"`,
					Value: `{"rev":"1-xxx"}`,
					Doc:   `{"_id":"foo","_rev":"1-xxx","cat":"meow","_conflicts":["1-aaa"]}`,
				},
			},
		}
	})
	tests.Add("conflicts=true ignored without include_docs", func(t *testing.T) any {
		db := newDB(t)
		_ = db.tPut("foo", map[string]string{
			"cat":  "meow",
			"_rev": "1-xxx",
		}, kivik.Param("new_edits", false))
		_ = db.tPut("foo", map[string]string{
			"cat":  "purr",
			"_rev": "1-aaa",
		}, kivik.Param("new_edits", false))

		return test{
			db: db,
			options: kivik.Params(map[string]any{
				"conflicts": true,
			}),
			want: []rowResult{
				{
					ID:    "foo",
					Key:   `"foo"`,
					Value: `{"rev":"1-xxx"}`,
				},
			},
		}
	})
	tests.Add("default sorting", func(t *testing.T) any {
		db := newDB(t)
		rev1 := db.tPut("cat", map[string]string{
			"cat": "meow",
		})
		rev2 := db.tPut("dog", map[string]string{
			"dog": "woof",
		})
		rev3 := db.tPut("cow", map[string]string{
			"cow": "moo",
		})

		return test{
			db: db,
			want: []rowResult{
				{
					ID:    "cat",
					Key:   `"cat"`,
					Value: `{"rev":"` + rev1 + `"}`,
				},
				{
					ID:    "cow",
					Key:   `"cow"`,
					Value: `{"rev":"` + rev3 + `"}`,
				},
				{
					ID:    "dog",
					Key:   `"dog"`,
					Value: `{"rev":"` + rev2 + `"}`,
				},
			},
		}
	})
	tests.Add("descending=true", func(t *testing.T) any {
		db := newDB(t)
		rev1 := db.tPut("cat", map[string]string{
			"cat": "meow",
		})
		rev2 := db.tPut("dog", map[string]string{
			"dog": "woof",
		})
		rev3 := db.tPut("cow", map[string]string{
			"cow": "moo",
		})

		return test{
			db:      db,
			options: kivik.Param("descending", true),
			want: []rowResult{
				{
					ID:    "dog",
					Key:   `"dog"`,
					Value: `{"rev":"` + rev2 + `"}`,
				},
				{
					ID:    "cow",
					Key:   `"cow"`,
					Value: `{"rev":"` + rev3 + `"}`,
				},
				{
					ID:    "cat",
					Key:   `"cat"`,
					Value: `{"rev":"` + rev1 + `"}`,
				},
			},
		}
	})
	tests.Add("endkey", func(t *testing.T) any {
		db := newDB(t)
		rev1 := db.tPut("cat", map[string]string{
			"cat": "meow",
		})
		_ = db.tPut("dog", map[string]string{
			"dog": "woof",
		})
		rev3 := db.tPut("cow", map[string]string{
			"cow": "moo",
		})

		return test{
			db:      db,
			options: kivik.Param("endkey", "cow"),
			want: []rowResult{
				{
					ID:    "cat",
					Key:   `"cat"`,
					Value: `{"rev":"` + rev1 + `"}`,
				},
				{
					ID:    "cow",
					Key:   `"cow"`,
					Value: `{"rev":"` + rev3 + `"}`,
				},
			},
		}
	})
	tests.Add("descending=true, endkey", func(t *testing.T) any {
		db := newDB(t)
		_ = db.tPut("cat", map[string]string{
			"cat": "meow",
		})
		rev2 := db.tPut("dog", map[string]string{
			"dog": "woof",
		})
		rev3 := db.tPut("cow", map[string]string{
			"cow": "moo",
		})

		return test{
			db: db,
			options: kivik.Params(map[string]any{
				"endkey":     "cow",
				"descending": true,
			}),
			want: []rowResult{
				{
					ID:    "dog",
					Key:   `"dog"`,
					Value: `{"rev":"` + rev2 + `"}`,
				},
				{
					ID:    "cow",
					Key:   `"cow"`,
					Value: `{"rev":"` + rev3 + `"}`,
				},
			},
		}
	})
	tests.Add("end_key", func(t *testing.T) any {
		db := newDB(t)
		rev1 := db.tPut("cat", map[string]string{
			"cat": "meow",
		})
		_ = db.tPut("dog", map[string]string{
			"dog": "woof",
		})
		rev3 := db.tPut("cow", map[string]string{
			"cow": "moo",
		})

		return test{
			db:      db,
			options: kivik.Param("end_key", "cow"),
			want: []rowResult{
				{
					ID:    "cat",
					Key:   `"cat"`,
					Value: `{"rev":"` + rev1 + `"}`,
				},
				{
					ID:    "cow",
					Key:   `"cow"`,
					Value: `{"rev":"` + rev3 + `"}`,
				},
			},
		}
	})
	tests.Add("endkey, inclusive_end=false", func(t *testing.T) any {
		db := newDB(t)
		rev1 := db.tPut("cat", map[string]string{
			"cat": "meow",
		})
		_ = db.tPut("dog", map[string]string{
			"dog": "woof",
		})
		_ = db.tPut("cow", map[string]string{
			"cow": "moo",
		})

		return test{
			db: db,
			options: kivik.Params(map[string]any{
				"endkey":        "cow",
				"inclusive_end": false,
			}),
			want: []rowResult{
				{
					ID:    "cat",
					Key:   `"cat"`,
					Value: `{"rev":"` + rev1 + `"}`,
				},
			},
		}
	})
	tests.Add("startkey", func(t *testing.T) any {
		db := newDB(t)
		_ = db.tPut("cat", map[string]string{
			"cat": "meow",
		})
		rev2 := db.tPut("dog", map[string]string{
			"dog": "woof",
		})
		rev3 := db.tPut("cow", map[string]string{
			"cow": "moo",
		})

		return test{
			db:      db,
			options: kivik.Param("startkey", "cow"),
			want: []rowResult{
				{
					ID:    "cow",
					Key:   `"cow"`,
					Value: `{"rev":"` + rev3 + `"}`,
				},
				{
					ID:    "dog",
					Key:   `"dog"`,
					Value: `{"rev":"` + rev2 + `"}`,
				},
			},
		}
	})
	tests.Add("start_key", func(t *testing.T) any {
		db := newDB(t)
		_ = db.tPut("cat", map[string]string{
			"cat": "meow",
		})
		rev2 := db.tPut("dog", map[string]string{
			"dog": "woof",
		})
		rev3 := db.tPut("cow", map[string]string{
			"cow": "moo",
		})

		return test{
			db:      db,
			options: kivik.Param("start_key", "cow"),
			want: []rowResult{
				{
					ID:    "cow",
					Key:   `"cow"`,
					Value: `{"rev":"` + rev3 + `"}`,
				},
				{
					ID:    "dog",
					Key:   `"dog"`,
					Value: `{"rev":"` + rev2 + `"}`,
				},
			},
		}
	})
	tests.Add("startkey, descending", func(t *testing.T) any {
		db := newDB(t)
		rev1 := db.tPut("cat", map[string]string{
			"cat": "meow",
		})
		_ = db.tPut("dog", map[string]string{
			"dog": "woof",
		})
		rev3 := db.tPut("cow", map[string]string{
			"cow": "moo",
		})

		return test{
			db: db,
			options: kivik.Params(map[string]any{
				"startkey":   "cow",
				"descending": true,
			}),
			want: []rowResult{
				{
					ID:    "cow",
					Key:   `"cow"`,
					Value: `{"rev":"` + rev3 + `"}`,
				},
				{
					ID:    "cat",
					Key:   `"cat"`,
					Value: `{"rev":"` + rev1 + `"}`,
				},
			},
		}
	})
	tests.Add("limit=2 returns first two documents only", func(t *testing.T) any {
		d := newDB(t)
		rev1 := d.tPut("cat", map[string]string{"cat": "meow"})
		_ = d.tPut("dog", map[string]string{"dog": "woof"})
		rev3 := d.tPut("cow", map[string]string{"cow": "moo"})

		return test{
			db:      d,
			options: kivik.Param("limit", 2),
			want: []rowResult{
				{
					ID:    "cat",
					Key:   `"cat"`,
					Value: `{"rev":"` + rev1 + `"}`,
				},
				{
					ID:    "cow",
					Key:   `"cow"`,
					Value: `{"rev":"` + rev3 + `"}`,
				},
			},
		}
	})
	tests.Add("skip=2 skips first two documents", func(t *testing.T) any {
		d := newDB(t)
		_ = d.tPut("cat", map[string]string{"cat": "meow"})
		rev2 := d.tPut("dog", map[string]string{"dog": "woof"})
		_ = d.tPut("cow", map[string]string{"cow": "moo"})

		return test{
			db:      d,
			options: kivik.Param("skip", 2),
			want: []rowResult{
				{
					ID:    "dog",
					Key:   `"dog"`,
					Value: `{"rev":"` + rev2 + `"}`,
				},
			},
		}
	})
	tests.Add("limit=1,skip=1 skips 1, limits 1", func(t *testing.T) any {
		d := newDB(t)
		_ = d.tPut("cat", map[string]string{"cat": "meow"})
		_ = d.tPut("dog", map[string]string{"dog": "woof"})
		rev3 := d.tPut("cow", map[string]string{"cow": "moo"})

		return test{
			db:      d,
			options: kivik.Params(map[string]any{"limit": 1, "skip": 1}),
			want: []rowResult{
				{
					ID:    "cow",
					Key:   `"cow"`,
					Value: `{"rev":"` + rev3 + `"}`,
				},
			},
		}
	})
	tests.Add("local docs excluded", func(t *testing.T) any {
		d := newDB(t)
		rev := d.tPut("cat", map[string]string{"cat": "meow"})
		_ = d.tPut("_local/dog", map[string]string{"dog": "woof"})
		rev3 := d.tPut("cow", map[string]string{"cow": "moo"})

		return test{
			db: d,
			want: []rowResult{
				{
					ID:    "cat",
					Key:   `"cat"`,
					Value: `{"rev":"` + rev + `"}`,
				},
				{
					ID:    "cow",
					Key:   `"cow"`,
					Value: `{"rev":"` + rev3 + `"}`,
				},
			},
		}
	})
	tests.Add("invalid limit value", test{
		options:    kivik.Params(map[string]any{"limit": "chicken"}),
		wantErr:    "invalid value for 'limit': chicken",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("invalid skip value", test{
		options:    kivik.Params(map[string]any{"skip": "chicken"}),
		wantErr:    "invalid value for 'skip': chicken",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("reduce not allowed", test{
		options:    kivik.Param("reduce", true),
		wantErr:    "reduce is invalid for map-only views",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("test collation order", func(t *testing.T) any {
		d := newDB(t)
		rev := d.tPut("~", map[string]string{})
		rev2 := d.tPut("a", map[string]string{})

		return test{
			db: d,
			want: []rowResult{
				{
					ID:    "~",
					Key:   `"~"`,
					Value: `{"rev":"` + rev + `"}`,
				},
				{
					ID:    "a",
					Key:   `"a"`,
					Value: `{"rev":"` + rev2 + `"}`,
				},
			},
		}
	})
	tests.Add("return unsorted results", func(t *testing.T) any {
		d := newDB(t)
		// Returned results are implicitly ordered, due to the ordering of
		// the index on the `id` column for joins. So we need to insert
		// a key that sorts differently implicitly (with ASCII ordering) than
		// explicitly (with CouchDB's UCI ordering). Thus the `~` id/key.  In
		// ASCII, ~ comes after the alphabet, in UCI it comes first.  So we
		// expect it to come after, with implicit ordering, or after, with
		// explicit. Comment out the options line below to see the difference.
		rev1 := d.tPut("~", map[string]any{"key": "~", "value": 3})
		rev2 := d.tPut("b", map[string]any{"key": "b", "value": 2})
		rev3 := d.tPut("a", map[string]any{"key": "a", "value": 1})

		return test{
			db:      d,
			options: kivik.Param("sorted", false),
			want: []rowResult{
				{ID: "a", Key: `"a"`, Value: `{"rev":"` + rev3 + `"}`},
				{ID: "b", Key: `"b"`, Value: `{"rev":"` + rev2 + `"}`},
				{ID: "~", Key: `"~"`, Value: `{"rev":"` + rev1 + `"}`},
			},
		}
	})
	tests.Add("support fetching specific key", func(t *testing.T) any {
		d := newDB(t)
		rev2 := d.tPut("b", map[string]any{"key": "b", "value": 2})
		_ = d.tPut("a", map[string]any{"key": "a", "value": 1})

		return test{
			db:      d,
			options: kivik.Param("key", "b"),
			want: []rowResult{
				{ID: "b", Key: `"b"`, Value: `{"rev":"` + rev2 + `"}`},
			},
		}
	})
	tests.Add("support fetching multiple specific keys", func(t *testing.T) any {
		d := newDB(t)
		rev1 := d.tPut("a", map[string]any{"key": "a", "value": 1})
		rev2 := d.tPut("b", map[string]any{"key": "b", "value": 2})
		_ = d.tPut("c", map[string]any{"key": "c", "value": 3})

		return test{
			db:      d,
			options: kivik.Param("keys", []string{"a", "b"}),
			want: []rowResult{
				{ID: "a", Key: `"a"`, Value: `{"rev":"` + rev1 + `"}`},
				{ID: "b", Key: `"b"`, Value: `{"rev":"` + rev2 + `"}`},
			},
		}
	})
	tests.Add("group not allowed", test{
		options:    kivik.Param("group", true),
		wantErr:    "group is invalid for map-only views",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("group_level not allowed", test{
		options:    kivik.Param("group_level", 3),
		wantErr:    "group_level is invalid for map-only views",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("fetch attachments", func(t *testing.T) any {
		d := newDB(t)
		rev1 := d.tPut("a", map[string]any{
			"_attachments": newAttachments().add("foo.txt", "This is a base64 encoding"),
		})

		return test{
			db: d,
			options: kivik.Params(map[string]any{
				"include_docs": true,
				"attachments":  true,
			}),
			want: []rowResult{
				{
					ID:    "a",
					Key:   `"a"`,
					Value: `{"rev":"` + rev1 + `"}`,
					Doc:   `{"_id":"a","_rev":"` + rev1 + `","_attachments":{"foo.txt":{"content_type":"text/plain","digest":"md5-TmfHxaRgUrE9l3tkAn4s0Q==","length":25,"revpos":1,"data":"VGhpcyBpcyBhIGJhc2U2NCBlbmNvZGluZw=="}}}`,
				},
			},
		}
	})
	tests.Add("document has attachment, but attachments=false", func(t *testing.T) any {
		d := newDB(t)
		rev1 := d.tPut("a", map[string]any{
			"_attachments": newAttachments().add("foo.txt", "This is a base64 encoding"),
		})

		return test{
			db: d,
			options: kivik.Params(map[string]any{
				"include_docs": true,
				"attachments":  false,
			}),
			want: []rowResult{
				{
					ID:    "a",
					Key:   `"a"`,
					Value: `{"rev":"` + rev1 + `"}`,
					Doc:   `{"_id":"a","_rev":"` + rev1 + `","_attachments":{"foo.txt":{"content_type":"text/plain","digest":"md5-TmfHxaRgUrE9l3tkAn4s0Q==","length":25,"revpos":1,"stub":true}}}`,
				},
			},
		}
	})
	tests.Add("doc with two attachments", func(t *testing.T) any {
		d := newDB(t)
		rev1 := d.tPut("a", map[string]any{
			"_attachments": newAttachments().
				add("foo.txt", "This is a base64 encoding").
				add("bar.txt", "This is also base64 encoded"),
		})

		return test{
			db: d,
			options: kivik.Params(map[string]any{
				"include_docs": true,
			}),
			want: []rowResult{
				{
					ID:    "a",
					Key:   `"a"`,
					Value: `{"rev":"` + rev1 + `"}`,
					Doc:   `{"_id":"a","_rev":"` + rev1 + `","_attachments":{"bar.txt":{"content_type":"text/plain","digest":"md5-uLHEKNY+WmubFxerYl5gvA==","length":27,"revpos":1,"stub":true},"foo.txt":{"content_type":"text/plain","digest":"md5-TmfHxaRgUrE9l3tkAn4s0Q==","length":25,"revpos":1,"stub":true}}}`,
				},
			},
		}
	})
	/*
		TODO:
		- Options:
			- att_encoding_infio
		- AllDocs() called for DB that doesn't exit
		- Offset() called on rows
		- TotalRows() called on rows
	*/

	tests.Run(t, func(t *testing.T, tt test) {
		t.Parallel()
		db := tt.db
		if db == nil {
			db = newDB(t)
		}
		opts := tt.options
		if opts == nil {
			opts = mock.NilOption
		}
		rows, err := db.AllDocs(context.Background(), opts)
		if !testy.ErrorMatches(tt.wantErr, err) {
			t.Errorf("Unexpected error: %s", err)
		}
		if status := kivik.HTTPStatus(err); status != tt.wantStatus {
			t.Errorf("Unexpected status: %d", status)
		}
		if err != nil {
			return
		}

		checkRows(t, rows, tt.want)
	})
}

func readRows(t *testing.T, rows driver.Rows) []rowResult {
	t.Helper()
	// iterate over rows
	var got []rowResult

loop:
	for {
		row := driver.Row{}
		err := rows.Next(&row)
		switch err {
		case io.EOF:
			break loop
		case driver.EOQ:
			continue
		case nil:
			// continue
		default:
			t.Fatalf("Next() returned error: %s", err)
		}
		var errMsg string
		if row.Error != nil {
			errMsg = row.Error.Error()
		}
		var value, doc []byte
		if row.Value != nil {
			value, err = io.ReadAll(row.Value)
			if err != nil {
				t.Fatal(err)
			}
		}

		if row.Doc != nil {
			doc, err = io.ReadAll(row.Doc)
			if err != nil {
				t.Fatal(err)
			}
		}
		got = append(got, rowResult{
			ID:    row.ID,
			Rev:   row.Rev,
			Key:   string(row.Key),
			Value: string(value),
			Doc:   string(doc),
			Error: errMsg,
		})
	}
	return got
}

func checkUnorderedRows(t *testing.T, rows driver.Rows, want []rowResult) {
	t.Helper()

	got := readRows(t, rows)
	sort.Slice(got, func(i, j int) bool {
		if r := strings.Compare(got[i].ID, got[j].ID); r != 0 {
			return r < 0
		}
		return strings.Compare(got[i].Key, got[j].Key) < 0
	})
	sort.Slice(want, func(i, j int) bool {
		if r := strings.Compare(want[i].ID, want[j].ID); r != 0 {
			return r < 0
		}
		return strings.Compare(want[i].Key, want[j].Key) < 0
	})
	if d := cmp.Diff(want, got); d != "" {
		t.Errorf("Unexpected rows:\n%s", d)
	}
}

func checkRows(t *testing.T, rows driver.Rows, want []rowResult) {
	t.Helper()

	got := readRows(t, rows)
	if d := cmp.Diff(want, got); d != "" {
		t.Errorf("Unexpected rows:\n%s", d)
	}
}

func TestDBAllDocs_total_rows(t *testing.T) {
	t.Parallel()
	d := newDB(t)

	_ = d.tPut("a", map[string]string{"foo": "bar"})
	_ = d.tPut("b", map[string]string{"foo": "baz"})
	_ = d.tPut("c", map[string]string{"foo": "qux"})

	rows, err := d.AllDocs(context.Background(), mock.NilOption)
	if err != nil {
		t.Fatalf("Failed to query AllDocs: %s", err)
	}
	defer rows.Close()

	// Consume all rows
	for {
		row := driver.Row{}
		if err := rows.Next(&row); err != nil {
			break
		}
	}

	want := int64(3)
	got := rows.TotalRows()
	if got != want {
		t.Errorf("Unexpected TotalRows: got %d, want %d", got, want)
	}
}

func TestDBAllDocs_update_seq(t *testing.T) {
	t.Parallel()
	d := newDB(t)

	_ = d.tPut("foo", map[string]string{"_id": "foo"})

	rows, err := d.AllDocs(context.Background(), kivik.Param("update_seq", true))
	if err != nil {
		t.Fatalf("Failed to query view: %s", err)
	}
	want := "1"
	got := rows.UpdateSeq()
	if got != want {
		t.Errorf("Unexpected update seq: %s", got)
	}
	_ = rows.Close()
}

func TestDBAllDocs_no_update_seq(t *testing.T) {
	t.Parallel()
	d := newDB(t)

	_ = d.tPut("foo", map[string]string{"_id": "foo"})

	rows, err := d.AllDocs(context.Background(), mock.NilOption)
	if err != nil {
		t.Fatalf("Failed to query view: %s", err)
	}
	want := ""
	got := rows.UpdateSeq()
	if got != want {
		t.Errorf("Unexpected update seq: %s", got)
	}
	_ = rows.Close()
}
