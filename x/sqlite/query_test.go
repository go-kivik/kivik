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
// +build !js

package sqlite

import (
	"context"
	"net/http"
	"testing"
	"time"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/x/sqlite/v4/internal/mock"
)

func TestDBQuery(t *testing.T) {
	t.Parallel()
	type test struct {
		db         *testDB
		ddoc, view string
		options    driver.Options
		want       []rowResult
		wantStatus int
		wantErr    string
		wantLogs   []string
	}
	tests := testy.NewTable()
	tests.Add("ddoc does not exist", test{
		ddoc:       "_design/foo",
		wantErr:    "missing",
		wantStatus: http.StatusNotFound,
	})
	tests.Add("ddoc does exist but view does not", func(t *testing.T) interface{} {
		d := newDB(t)
		_ = d.tPut("_design/foo", map[string]string{"cat": "meow"})

		return test{
			db:         d,
			ddoc:       "_design/foo",
			view:       "_view/bar",
			wantErr:    "missing named view",
			wantStatus: http.StatusNotFound,
		}
	})
	tests.Add("simple view with a single document", func(t *testing.T) interface{} {
		d := newDB(t)
		_ = d.tPut("_design/foo", map[string]interface{}{
			"views": map[string]interface{}{
				"bar": map[string]string{
					"map": `function(doc) { emit(doc._id, null); }`,
				},
			},
		})
		_ = d.tPut("foo", map[string]string{"_id": "foo"})

		return test{
			db:   d,
			ddoc: "_design/foo",
			view: "_view/bar",
			want: []rowResult{
				{ID: "_design/foo", Key: `"_design/foo"`, Value: "null"},
				{ID: "foo", Key: `"foo"`, Value: "null"},
			},
		}
	})
	tests.Add("invalid update value", test{
		options:    kivik.Param("update", "foo"),
		wantErr:    "invalid value for `update`",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("with update=false", func(t *testing.T) interface{} {
		d := newDB(t)
		_ = d.tPut("_design/foo", map[string]interface{}{
			"views": map[string]interface{}{
				"bar": map[string]string{
					"map": `function(doc) { emit(doc._id, null); }`,
				},
			},
		})
		_ = d.tPut("foo", map[string]string{"_id": "foo"})

		return test{
			db:      d,
			ddoc:    "_design/foo",
			view:    "_view/bar",
			options: kivik.Param("update", false),
			want:    nil,
		}
	})
	tests.Add("explicit update=true", func(t *testing.T) interface{} {
		d := newDB(t)
		_ = d.tPut("_design/foo", map[string]interface{}{
			"views": map[string]interface{}{
				"bar": map[string]string{
					"map": `function(doc) { emit(doc._id, null); }`,
				},
			},
		})
		_ = d.tPut("foo", map[string]string{"_id": "foo"})

		return test{
			db:      d,
			ddoc:    "_design/foo",
			view:    "_view/bar",
			options: kivik.Param("update", true),
			want: []rowResult{
				{ID: "_design/foo", Key: `"_design/foo"`, Value: "null"},
				{ID: "foo", Key: `"foo"`, Value: "null"},
			},
		}
	})
	tests.Add("Updating ddoc renders index obsolete", func(t *testing.T) interface{} {
		d := newDB(t)
		rev := d.tPut("_design/foo", map[string]interface{}{
			"views": map[string]interface{}{
				"bar": map[string]string{
					"map": `function(doc) { emit(doc._id, null); }`,
				},
			},
		})
		_ = d.tPut("foo", map[string]string{"_id": "foo"})
		// Ensure the index is built
		rows, err := d.Query(context.Background(), "_design/foo", "_view/bar", mock.NilOption)
		if err != nil {
			t.Fatalf("Failed to query view: %s", err)
		}
		_ = rows.Close()

		// Update the ddoc
		_ = d.tPut("_design/foo", map[string]interface{}{
			"views": map[string]interface{}{
				"bar": map[string]string{
					"map": `function(x) { emit(x._id, null); }`,
				},
			},
		}, kivik.Rev(rev))

		return test{
			db:      d,
			ddoc:    "_design/foo",
			view:    "_view/bar",
			options: kivik.Param("update", false),
			want:    nil,
		}
	})
	tests.Add("incremental update", func(t *testing.T) interface{} {
		d := newDB(t)
		_ = d.tPut("_design/foo", map[string]interface{}{
			"views": map[string]interface{}{
				"bar": map[string]string{
					"map": `function(doc) { emit(doc._id, null); }`,
				},
			},
		})
		_ = d.tPut("foo", map[string]string{"_id": "foo"})
		// Ensure the index is built
		rows, err := d.Query(context.Background(), "_design/foo", "_view/bar", mock.NilOption)
		if err != nil {
			t.Fatalf("Failed to query view: %s", err)
		}
		_ = rows.Close()

		// Add a new doc to trigger an incremental update
		_ = d.tPut("bar", map[string]interface{}{"_id": "bar"})

		return test{
			db:   d,
			ddoc: "_design/foo",
			view: "_view/bar",
			want: []rowResult{
				{ID: "_design/foo", Key: `"_design/foo"`, Value: "null"},
				{ID: "bar", Key: `"bar"`, Value: "null"},
				{ID: "foo", Key: `"foo"`, Value: "null"},
			},
		}
	})
	tests.Add("ddoc does not exist with update=false", test{
		ddoc:       "_design/foo",
		options:    kivik.Param("update", false),
		wantErr:    "missing",
		wantStatus: http.StatusNotFound,
	})
	tests.Add("ddoc does exist but view does not with update=false", func(t *testing.T) interface{} {
		d := newDB(t)
		_ = d.tPut("_design/foo", map[string]string{"cat": "meow"})

		return test{
			db:         d,
			ddoc:       "_design/foo",
			view:       "_view/bar",
			options:    kivik.Param("update", false),
			wantErr:    "missing named view",
			wantStatus: http.StatusNotFound,
		}
	})
	tests.Add("deleting doc deindexes it", func(t *testing.T) interface{} {
		d := newDB(t)
		rev := d.tPut("foo", map[string]string{"cat": "meow"})

		_ = d.tPut("_design/foo", map[string]interface{}{
			"views": map[string]interface{}{
				"bar": map[string]string{
					"map": `function(doc) { emit(doc._id, null); }`,
				},
			},
		})
		// Ensure the index is built
		rows, err := d.Query(context.Background(), "_design/foo", "_view/bar", mock.NilOption)
		if err != nil {
			t.Fatalf("Failed to query view: %s", err)
		}
		_ = rows.Close()

		// Add a new doc to trigger an incremental update
		_ = d.tDelete("foo", kivik.Rev(rev))

		return test{
			db:   d,
			ddoc: "_design/foo",
			view: "_view/bar",
			want: []rowResult{
				{ID: "_design/foo", Key: `"_design/foo"`, Value: "null"},
			},
		}
	})
	tests.Add("doc deleted before index creation", func(t *testing.T) interface{} {
		d := newDB(t)
		_ = d.tPut("_design/foo", map[string]interface{}{
			"views": map[string]interface{}{
				"bar": map[string]string{
					"map": `function(doc) { emit(doc._id, null); }`,
				},
			},
		})
		rev := d.tPut("foo", map[string]string{"_id": "foo"})
		_ = d.tDelete("foo", kivik.Rev(rev))

		return test{
			db:   d,
			ddoc: "_design/foo",
			view: "_view/bar",
			want: []rowResult{
				{ID: "_design/foo", Key: `"_design/foo"`, Value: "null"},
			},
		}
	})
	tests.Add("map function throws exception", func(t *testing.T) interface{} {
		d := newDB(t)
		_ = d.tPut("foo", map[string]string{"cat": "meow"})
		_ = d.tPut("_design/foo", map[string]interface{}{
			"views": map[string]interface{}{
				"bar": map[string]string{
					"map": `function(doc) {
						emit(doc._id, null);
						throw new Error("broken");
					}`,
				},
			},
		})
		return test{
			db:   d,
			ddoc: "_design/foo",
			view: "_view/bar",
			want: nil,
			wantLogs: []string{
				"map function threw exception for foo: Error: broken",
				"\tat map (<eval>:3:13(9))",
				"map function threw exception for _design/foo: Error: broken",
				"\tat map (<eval>:3:13(9))",
			},
		}
	})
	tests.Add("emit function throws exception", func(t *testing.T) interface{} {
		d := newDB(t)
		_ = d.tPut("foo", map[string]string{"cat": "meow"})
		_ = d.tPut("_design/foo", map[string]interface{}{
			"views": map[string]interface{}{
				"bar": map[string]string{
					"map": `function(doc) {
						emit(doc._id, function() {});
					}`,
				},
			},
		})
		return test{
			db:   d,
			ddoc: "_design/foo",
			view: "_view/bar",
			want: nil,
			wantLogs: []string{
				"map function threw exception for foo: json: unsupported type: func(goja.FunctionCall) goja.Value",
				"\tat github.com/go-kivik/kivik/v4/x/sqlite.(*db).updateIndex.(*db).updateIndex.func1.func2 (native)",
				"\tat map (<eval>:2:11(5))",
				"map function threw exception for _design/foo: json: unsupported type: func(goja.FunctionCall) goja.Value",
				"\tat github.com/go-kivik/kivik/v4/x/sqlite.(*db).updateIndex.(*db).updateIndex.func1.func2 (native)",
				"\tat map (<eval>:2:11(5))",
			},
		}
	})
	tests.Add("map that references attachments", func(t *testing.T) interface{} {
		d := newDB(t)
		_ = d.tPut("_design/foo", map[string]interface{}{
			"views": map[string]interface{}{
				"bar": map[string]string{
					"map": `function(doc) {
						if (doc._attachments) { // Check if there are attachments
							for (var filename in doc._attachments) { // Loop over all attachments
								emit(filename); // Emit the attachment filename
							}
						}
					}`,
				},
			},
		})
		_ = d.tPut("no_attachments", map[string]string{"foo": "bar"})
		_ = d.tPut("with_attachments", map[string]interface{}{
			"_attachments": newAttachments().
				add("foo.txt", "Hello, World!").
				add("bar.txt", "Goodbye, World!"),
		})

		return test{
			db:   d,
			ddoc: "_design/foo",
			view: "_view/bar",
			want: []rowResult{
				{ID: "with_attachments", Key: `"bar.txt"`, Value: "null"},
				{ID: "with_attachments", Key: `"foo.txt"`, Value: "null"},
			},
		}
	})
	/*
		TODO:
		- Are conflicts or other metadata exposed to map function?
		- custom/standard CouchDB collation https://pkg.go.dev/modernc.org/sqlite#RegisterCollationUtf8
		- built-in reduce functions: _sum, _count
		- Options:
			- conflicts
			- descending
			- endkey
			- end_key
			- endkey_docid
			- end_key_doc_id
			- group
			- group_level
			- include_docs
			- inclusive_end
			- key
			- keys
			- limit
			- reduce
			- skip
			- sorted
			- stable // N/A only for clusters
			- stale // deprecated
			- startkey
			- start_key
			- startkey_docid
			- start_key_doc_id
			- update_seq
		- map function takes too long

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
		rows, err := db.Query(context.Background(), tt.ddoc, tt.view, opts)
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
		db.checkLogs(tt.wantLogs)
	})
}

func TestDBQuery_update_lazy(t *testing.T) {
	t.Parallel()
	d := newDB(t)

	_ = d.tPut("_design/foo", map[string]interface{}{
		"views": map[string]interface{}{
			"bar": map[string]string{
				"map": `function(doc) { emit(doc._id, null); }`,
			},
		},
	})
	_ = d.tPut("foo", map[string]string{"_id": "foo"})

	// Separate func for defer, to ensure it runs after the query.
	func() {
		// Do a query, with update=lazy, which should return nothing, but trigger
		// an index build.
		rows, err := d.Query(context.Background(), "_design/foo", "_view/bar", kivik.Param("update", "lazy"))
		if err != nil {
			t.Fatalf("Failed to query view: %s", err)
		}
		defer rows.Close()
		checkRows(t, rows, nil)
	}()

	// Now wait a moment, and query again to see if there are any results
	time.Sleep(500 * time.Millisecond)

	rows, err := d.Query(context.Background(), "_design/foo", "_view/bar", kivik.Param("update", "false"))
	if err != nil {
		t.Fatalf("Failed to query view: %s", err)
	}
	defer rows.Close()
	checkRows(t, rows, []rowResult{
		{ID: "_design/foo", Key: `"_design/foo"`, Value: "null"},
		{ID: "foo", Key: `"foo"`, Value: "null"},
	})
}
