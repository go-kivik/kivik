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
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal/mock"
)

type rowResult struct {
	ID    string
	Rev   string
	Doc   string
	Value string
	Error string
}

func TestDBAllDocs(t *testing.T) {
	t.Parallel()
	type test struct {
		setup      func(*testing.T, driver.DB)
		options    driver.Options
		want       []rowResult
		wantStatus int
		wantErr    string
	}
	tests := testy.NewTable()
	tests.Add("no docs in db", test{
		want: nil,
	})
	tests.Add("single doc", test{
		setup: func(t *testing.T, db driver.DB) {
			_, err := db.Put(context.Background(), "foo", map[string]string{"cat": "meow"}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
		},
		want: []rowResult{
			{
				ID:    "foo",
				Rev:   "1-274558516009acbe973682d27a58b598",
				Value: `{"value":{"rev":"1-274558516009acbe973682d27a58b598"}}` + "\n",
			},
		},
	})
	tests.Add("include_docs=true", test{
		setup: func(t *testing.T, db driver.DB) {
			_, err := db.Put(context.Background(), "foo", map[string]string{"cat": "meow"}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
		},
		options: kivik.Param("include_docs", true),
		want: []rowResult{
			{
				ID:    "foo",
				Rev:   "1-274558516009acbe973682d27a58b598",
				Value: `{"value":{"rev":"1-274558516009acbe973682d27a58b598"}}` + "\n",
				Doc:   `{"_id":"foo","_rev":"1-274558516009acbe973682d27a58b598","cat":"meow"}`,
			},
		},
	})
	tests.Add("single doc multiple revisions", test{
		setup: func(t *testing.T, db driver.DB) {
			_, err := db.Put(context.Background(), "foo", map[string]string{"cat": "meow"}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
			_, err = db.Put(context.Background(), "foo", map[string]string{"cat": "purr"}, kivik.Rev("1-274558516009acbe973682d27a58b598"))
			if err != nil {
				t.Fatal(err)
			}
		},
		want: []rowResult{
			{
				ID:    "foo",
				Rev:   "2-c1f7f9ed8874502b095381186a35af4b",
				Value: `{"value":{"rev":"2-c1f7f9ed8874502b095381186a35af4b"}}` + "\n",
			},
		},
	})
	tests.Add("conflicting document, select winning rev", test{
		setup: func(t *testing.T, db driver.DB) {
			_, err := db.Put(context.Background(), "foo", map[string]string{
				"cat":  "meow",
				"_rev": "1-xxx",
			}, kivik.Param("new_edits", false))
			if err != nil {
				t.Fatal(err)
			}
			_, err = db.Put(context.Background(), "foo", map[string]string{
				"cat":  "purr",
				"_rev": "1-aaa",
			}, kivik.Param("new_edits", false))
			if err != nil {
				t.Fatal(err)
			}
		},
		want: []rowResult{
			{
				ID:    "foo",
				Rev:   "1-xxx",
				Value: `{"value":{"rev":"1-xxx"}}` + "\n",
			},
		},
	})
	tests.Add("deleted doc", test{
		setup: func(t *testing.T, db driver.DB) {
			_, err := db.Put(context.Background(), "foo", map[string]string{"cat": "meow"}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
			_, err = db.Delete(context.Background(), "foo", kivik.Rev("1-274558516009acbe973682d27a58b598"))
			if err != nil {
				t.Fatal(err)
			}
		},
		want: nil,
	})
	tests.Add("select lower revision number when higher rev in winning branch has been deleted", test{
		setup: func(t *testing.T, db driver.DB) {
			_, err := db.Put(context.Background(), "foo", map[string]string{
				"cat":  "meow",
				"_rev": "1-xxx",
			}, kivik.Param("new_edits", false))
			if err != nil {
				t.Fatal(err)
			}
			_, err = db.Put(context.Background(), "foo", map[string]string{
				"cat":  "purr",
				"_rev": "1-aaa",
			}, kivik.Param("new_edits", false))
			if err != nil {
				t.Fatal(err)
			}
			_, err = db.Delete(context.Background(), "foo", kivik.Rev("1-aaa"))
			if err != nil {
				t.Fatal(err)
			}
		},
		want: []rowResult{
			{
				ID:    "foo",
				Rev:   "1-xxx",
				Value: `{"value":{"rev":"1-xxx"}}` + "\n",
			},
		},
	})
	tests.Add("conflicts=true", test{
		setup: func(t *testing.T, db driver.DB) {
			_, err := db.Put(context.Background(), "foo", map[string]string{
				"cat":  "meow",
				"_rev": "1-xxx",
			}, kivik.Param("new_edits", false))
			if err != nil {
				t.Fatal(err)
			}
			_, err = db.Put(context.Background(), "foo", map[string]string{
				"cat":  "purr",
				"_rev": "1-aaa",
			}, kivik.Param("new_edits", false))
			if err != nil {
				t.Fatal(err)
			}
		},
		options: kivik.Params(map[string]interface{}{
			"conflicts":    true,
			"include_docs": true,
		}),
		want: []rowResult{
			{
				ID:    "foo",
				Rev:   "1-xxx",
				Value: `{"value":{"rev":"1-xxx"}}` + "\n",
				Doc:   `{"_id":"foo","_rev":"1-xxx","cat":"meow","_conflicts":["1-aaa"]}`,
			},
		},
	})
	tests.Add("conflicts=true ignored without include_docs", test{
		setup: func(t *testing.T, db driver.DB) {
			_, err := db.Put(context.Background(), "foo", map[string]string{
				"cat":  "meow",
				"_rev": "1-xxx",
			}, kivik.Param("new_edits", false))
			if err != nil {
				t.Fatal(err)
			}
			_, err = db.Put(context.Background(), "foo", map[string]string{
				"cat":  "purr",
				"_rev": "1-aaa",
			}, kivik.Param("new_edits", false))
			if err != nil {
				t.Fatal(err)
			}
		},
		options: kivik.Params(map[string]interface{}{
			"conflicts": true,
		}),
		want: []rowResult{
			{
				ID:    "foo",
				Rev:   "1-xxx",
				Value: `{"value":{"rev":"1-xxx"}}` + "\n",
			},
		},
	})
	tests.Add("default sorting", test{
		setup: func(t *testing.T, db driver.DB) {
			_, err := db.Put(context.Background(), "cat", map[string]string{
				"cat": "meow",
			}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
			_, err = db.Put(context.Background(), "dog", map[string]string{
				"dog": "woof",
			}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
			_, err = db.Put(context.Background(), "cow", map[string]string{
				"cow": "moo",
			}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
		},
		want: []rowResult{
			{
				ID:    "cat",
				Rev:   "1-274558516009acbe973682d27a58b598",
				Value: `{"value":{"rev":"1-274558516009acbe973682d27a58b598"}}` + "\n",
			},
			{
				ID:    "cow",
				Rev:   "1-80b1ed11e92f08613f0007cc2b2f486d",
				Value: `{"value":{"rev":"1-80b1ed11e92f08613f0007cc2b2f486d"}}` + "\n",
			},
			{
				ID:    "dog",
				Rev:   "1-a5f1dc478532231c6252f63fa94f433a",
				Value: `{"value":{"rev":"1-a5f1dc478532231c6252f63fa94f433a"}}` + "\n",
			},
		},
	})
	tests.Add("descending=true", test{
		setup: func(t *testing.T, db driver.DB) {
			_, err := db.Put(context.Background(), "cat", map[string]string{
				"cat": "meow",
			}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
			_, err = db.Put(context.Background(), "dog", map[string]string{
				"dog": "woof",
			}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
			_, err = db.Put(context.Background(), "cow", map[string]string{
				"cow": "moo",
			}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
		},
		options: kivik.Param("descending", true),
		want: []rowResult{
			{
				ID:    "dog",
				Rev:   "1-a5f1dc478532231c6252f63fa94f433a",
				Value: `{"value":{"rev":"1-a5f1dc478532231c6252f63fa94f433a"}}` + "\n",
			},
			{
				ID:    "cow",
				Rev:   "1-80b1ed11e92f08613f0007cc2b2f486d",
				Value: `{"value":{"rev":"1-80b1ed11e92f08613f0007cc2b2f486d"}}` + "\n",
			},
			{
				ID:    "cat",
				Rev:   "1-274558516009acbe973682d27a58b598",
				Value: `{"value":{"rev":"1-274558516009acbe973682d27a58b598"}}` + "\n",
			},
		},
	})
	tests.Add("endkey", test{
		setup: func(t *testing.T, db driver.DB) {
			_, err := db.Put(context.Background(), "cat", map[string]string{
				"cat": "meow",
			}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
			_, err = db.Put(context.Background(), "dog", map[string]string{
				"dog": "woof",
			}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
			_, err = db.Put(context.Background(), "cow", map[string]string{
				"cow": "moo",
			}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
		},
		options: kivik.Param("endkey", "cow"),
		want: []rowResult{
			{
				ID:    "cat",
				Rev:   "1-274558516009acbe973682d27a58b598",
				Value: `{"value":{"rev":"1-274558516009acbe973682d27a58b598"}}` + "\n",
			},
			{
				ID:    "cow",
				Rev:   "1-80b1ed11e92f08613f0007cc2b2f486d",
				Value: `{"value":{"rev":"1-80b1ed11e92f08613f0007cc2b2f486d"}}` + "\n",
			},
		},
	})
	tests.Add("descending=true, endkey", test{
		setup: func(t *testing.T, db driver.DB) {
			_, err := db.Put(context.Background(), "cat", map[string]string{
				"cat": "meow",
			}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
			_, err = db.Put(context.Background(), "dog", map[string]string{
				"dog": "woof",
			}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
			_, err = db.Put(context.Background(), "cow", map[string]string{
				"cow": "moo",
			}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
		},
		options: kivik.Params(map[string]interface{}{
			"endkey":     "cow",
			"descending": true,
		}),
		want: []rowResult{
			{
				ID:    "dog",
				Rev:   "1-a5f1dc478532231c6252f63fa94f433a",
				Value: `{"value":{"rev":"1-a5f1dc478532231c6252f63fa94f433a"}}` + "\n",
			},
			{
				ID:    "cow",
				Rev:   "1-80b1ed11e92f08613f0007cc2b2f486d",
				Value: `{"value":{"rev":"1-80b1ed11e92f08613f0007cc2b2f486d"}}` + "\n",
			},
		},
	})
	tests.Add("end_key", test{
		setup: func(t *testing.T, db driver.DB) {
			_, err := db.Put(context.Background(), "cat", map[string]string{
				"cat": "meow",
			}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
			_, err = db.Put(context.Background(), "dog", map[string]string{
				"dog": "woof",
			}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
			_, err = db.Put(context.Background(), "cow", map[string]string{
				"cow": "moo",
			}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
		},
		options: kivik.Param("end_key", "cow"),
		want: []rowResult{
			{
				ID:    "cat",
				Rev:   "1-274558516009acbe973682d27a58b598",
				Value: `{"value":{"rev":"1-274558516009acbe973682d27a58b598"}}` + "\n",
			},
			{
				ID:    "cow",
				Rev:   "1-80b1ed11e92f08613f0007cc2b2f486d",
				Value: `{"value":{"rev":"1-80b1ed11e92f08613f0007cc2b2f486d"}}` + "\n",
			},
		},
	})

	/*
		TODO:
		- deleted doc
		- Order return values
		- Options:
			- endkey_docid
			- end_key_doc_id
			- group
			- group_level
			- include_docs
			- attachments
			- att_encoding_infio
			- inclusive_end
			- key
			- keys
			- limit
			- reduce
			- skip
			- sorted
			- stable
			- statle
			- startkey
			- start_key
			- startkey_docid
			- start_key_doc_id
			- update
			- update_seq
		- AllDocs() called for DB that doesn't exit
		- UpdateSeq() called on rows
		- Offset() called on rows
		- TotalRows() called on rows
	*/

	tests.Run(t, func(t *testing.T, tt test) {
		t.Parallel()
		db := newDB(t)
		opts := tt.options
		if tt.setup != nil {
			tt.setup(t, db)
		}
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
			value, err := io.ReadAll(row.Value)
			if err != nil {
				t.Fatal(err)
			}
			var doc []byte
			if row.Doc != nil {
				doc, err = io.ReadAll(row.Doc)
				if err != nil {
					t.Fatal(err)
				}
			}
			got = append(got, rowResult{
				ID:    row.ID,
				Rev:   row.Rev,
				Value: string(value),
				Doc:   string(doc),
				Error: errMsg,
			})
		}
		if d := cmp.Diff(tt.want, got); d != "" {
			t.Errorf("Unexpected rows:\n%s", d)
		}
	})
}
