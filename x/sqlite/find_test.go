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

package sqlite

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/int/mock"
)

func TestFind(t *testing.T) {
	t.Parallel()
	type test struct {
		db         *testDB
		query      string
		want       []rowResult
		wantStatus int
		wantErr    string
	}

	tests := testy.NewTable()
	tests.Add("no docs in db", test{
		query: `{"selector":{}}`,
		want:  nil,
	})
	tests.Add("query is invalid json", test{
		query:      "invalid json",
		wantStatus: http.StatusBadRequest,
		wantErr:    "invalid character 'i' looking for beginning of value",
	})
	tests.Add("field equality", func(t *testing.T) interface{} {
		d := newDB(t)
		rev := d.tPut("foo", map[string]string{"foo": "bar"})
		_ = d.tPut("bar", map[string]string{"bar": "baz"})

		return test{
			db:    d,
			query: `{"selector":{"foo":"bar"}}`,
			want: []rowResult{
				{Doc: `{"_id":"foo","_rev":"` + rev + `","foo":"bar"}`},
			},
		}
	})
	tests.Add("limit", func(t *testing.T) interface{} {
		d := newDB(t)
		rev := d.tPut("foo", map[string]string{"foo": "bar"})
		_ = d.tPut("bar", map[string]string{"bar": "baz"})
		rev2 := d.tPut("foo2", map[string]string{"foo": "bar"})
		_ = d.tPut("foo3", map[string]string{"foo": "bar"})

		return test{
			db:    d,
			query: `{"selector":{"foo": "bar"}, "limit": 2}`,
			want: []rowResult{
				{Doc: `{"_id":"foo","_rev":"` + rev + `","foo":"bar"}`},
				{Doc: `{"_id":"foo2","_rev":"` + rev2 + `","foo":"bar"}`},
			},
		}
	})
	tests.Add("skip", func(t *testing.T) interface{} {
		d := newDB(t)
		_ = d.tPut("foo", map[string]string{"foo": "bar"})
		_ = d.tPut("bar", map[string]string{"bar": "baz"})
		_ = d.tPut("foo2", map[string]string{"foo": "bar"})
		rev3 := d.tPut("foo3", map[string]string{"foo": "bar"})

		return test{
			db:    d,
			query: `{"selector":{"foo": "bar"}, "skip": 2}`,
			want: []rowResult{
				{Doc: `{"_id":"foo3","_rev":"` + rev3 + `","foo":"bar"}`},
			},
		}
	})
	tests.Add("fields", func(t *testing.T) interface{} {
		d := newDB(t)
		_ = d.tPut("foo", map[string]interface{}{
			"foo": "bar",
			"baz": "qux",
			"deeply": map[string]interface{}{
				"nested": "value",
				"other":  "value",
				"yet":    "more",
			},
		})

		return test{
			db:    d,
			query: `{"selector":{"foo": "bar"},"fields":["foo","deeply.nested","deeply.yet"]}`,
			want: []rowResult{
				{Doc: `{"deeply":{"nested":"value","yet":"more"},"foo":"bar"}`},
			},
		}
	})
	tests.Add("_attachments field ", func(t *testing.T) interface{} {
		d := newDB(t)
		_ = d.tPut("foo", map[string]interface{}{
			"foo":          "bar",
			"_attachments": newAttachments().add("foo.txt", "foo"),
		})

		return test{
			db:    d,
			query: `{"selector":{"foo": "bar"},"fields":["_attachments"]}`,
			want: []rowResult{
				{Doc: `{"_attachments":{"foo.txt":{"content_type":"text/plain","digest":"md5-rL0Y20zC+Fzt72VPzMSk2A==","length":3,"revpos":1,"stub":true}}}`},
			},
		}
	})
	tests.Add("_conflicts field ", func(t *testing.T) interface{} {
		d := newDB(t)
		_ = d.tPut("foo", map[string]interface{}{"_rev": "1-foo"}, kivik.Param("new_edits", false))
		_ = d.tPut("foo", map[string]interface{}{"_rev": "1-bar"}, kivik.Param("new_edits", false))

		return test{
			db:    d,
			query: `{"selector":{},"fields":["_conflicts"],"conflicts":true}`,
			want: []rowResult{
				{Doc: `{"_conflicts":["1-bar"]}`},
			},
		}
	})
	tests.Add("bookmark", func(t *testing.T) interface{} {
		d := newDB(t)
		_ = d.tPut("a", map[string]interface{}{})
		_ = d.tPut("b", map[string]interface{}{})
		revC := d.tPut("c", map[string]interface{}{})
		_ = d.tPut("d", map[string]interface{}{})

		rows, err := d.Find(t.Context(), json.RawMessage(`{"selector":{},"limit":1,"skip":1}`), mock.NilOption)
		if err != nil {
			t.Fatalf("Failed to get bookmark: %s", err)
		}
		defer rows.Close()
		var row driver.Row
		for {
			err := rows.Next(&row)
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
		}
		bookmark := rows.(driver.Bookmarker).Bookmark()

		return test{
			db:    d,
			query: `{"selector":{},"bookmark":"` + bookmark + `","limit":1}`,
			want: []rowResult{
				{Doc: `{"_id":"c","_rev":"` + revC + `"}`},
			},
		}
	})
	tests.Add("non-string bookmark", test{
		query:      `{"selector":{},"bookmark":true}`,
		wantStatus: http.StatusBadRequest,
		wantErr:    "invalid value for 'bookmark': true",
	})
	tests.Add("invalid bookmark", test{
		query:      `{"selector":{},"bookmark":"moo"}`,
		wantStatus: http.StatusBadRequest,
		wantErr:    "invalid value for 'bookmark': moo",
	})
	tests.Add("sort", func(t *testing.T) interface{} {
		d := newDB(t)
		// revA := d.tPut("a", map[string]interface{}{"name": "Bob"})
		// revB := d.tPut("b", map[string]interface{}{"name": "Alice"})
		// revC := d.tPut("c", map[string]interface{}{"name": "Charlie"})
		// revD := d.tPut("d", map[string]interface{}{"name": "Dick"})

		return test{
			db:    d,
			query: `{"selector":{},"sort":["name"]}`,
			// TODO: Support sorting
			wantStatus: http.StatusBadRequest,
			wantErr:    "no index exists for this sort, try indexing by the sort fields",
		}
	})
	tests.Add("sort, non-array", test{
		query:      `{"selector":{},"sort":"x"}`,
		wantStatus: http.StatusBadRequest,
		wantErr:    "invalid value for 'sort': x",
	})
	tests.Add("sort, invalid field", test{
		query:      `{"selector":{},"sort":["x",3]}`,
		wantStatus: http.StatusBadRequest,
		wantErr:    "invalid 'sort' field: 3",
	})
	tests.Add("fields, non-array", test{
		query:      `{"selector":{},"fields":"x"}`,
		wantStatus: http.StatusBadRequest,
		wantErr:    "invalid value for 'fields': x",
	})
	tests.Add("fields, invalid field", test{
		query:      `{"selector":{},"fields":["x",3]}`,
		wantStatus: http.StatusBadRequest,
		wantErr:    "invalid 'fields' field: 3",
	})

	/*
		TODO:
		- sort
		- stable
		- update
		- stale
		- use_index
		- execution_stats -- Not currently supported by Kivik
	*/

	tests.Run(t, func(t *testing.T, tt test) {
		t.Parallel()
		db := tt.db
		if db == nil {
			db = newDB(t)
		}
		rows, err := db.Find(t.Context(), json.RawMessage(tt.query), mock.NilOption)
		if !testy.ErrorMatchesRE(tt.wantErr, err) {
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
