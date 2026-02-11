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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
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
	tests.Add("field equality", func(t *testing.T) any {
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
	tests.Add("limit", func(t *testing.T) any {
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
	tests.Add("skip", func(t *testing.T) any {
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
	tests.Add("fields", func(t *testing.T) any {
		d := newDB(t)
		_ = d.tPut("foo", map[string]any{
			"foo": "bar",
			"baz": "qux",
			"deeply": map[string]any{
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
	tests.Add("_attachments field ", func(t *testing.T) any {
		d := newDB(t)
		_ = d.tPut("foo", map[string]any{
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
	tests.Add("_conflicts field ", func(t *testing.T) any {
		d := newDB(t)
		_ = d.tPut("foo", map[string]any{"_rev": "1-foo"}, kivik.Param("new_edits", false))
		_ = d.tPut("foo", map[string]any{"_rev": "1-bar"}, kivik.Param("new_edits", false))

		return test{
			db:    d,
			query: `{"selector":{},"fields":["_conflicts"],"conflicts":true}`,
			want: []rowResult{
				{Doc: `{"_conflicts":["1-bar"]}`},
			},
		}
	})
	tests.Add("bookmark", func(t *testing.T) any {
		d := newDB(t)
		_ = d.tPut("a", map[string]any{})
		_ = d.tPut("b", map[string]any{})
		revC := d.tPut("c", map[string]any{})
		_ = d.tPut("d", map[string]any{})

		rows, err := d.Find(context.Background(), json.RawMessage(`{"selector":{},"limit":1,"skip":1}`), mock.NilOption)
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
	tests.Add("sort without matching index", func(t *testing.T) any {
		d := newDB(t)

		return test{
			db:         d,
			query:      `{"selector":{},"sort":["name"]}`,
			wantStatus: http.StatusBadRequest,
			wantErr:    "no index exists for this sort, try indexing by the sort fields",
		}
	})
	tests.Add("sort ascending with index", func(t *testing.T) any {
		d := newDB(t)
		err := d.CreateIndex(context.Background(), "_design/idx", "byName", json.RawMessage(`{"fields":["name"]}`), mock.NilOption)
		if err != nil {
			t.Fatalf("CreateIndex failed: %s", err)
		}
		revAlice := d.tPut("alice", map[string]string{"name": "Charlie"})
		revBob := d.tPut("bob", map[string]string{"name": "Alice"})
		revCharlie := d.tPut("charlie", map[string]string{"name": "Bob"})

		return test{
			db:    d,
			query: `{"selector":{},"sort":["name"]}`,
			want: []rowResult{
				{Doc: `{"_id":"bob","_rev":"` + revBob + `","name":"Alice"}`},
				{Doc: `{"_id":"charlie","_rev":"` + revCharlie + `","name":"Bob"}`},
				{Doc: `{"_id":"alice","_rev":"` + revAlice + `","name":"Charlie"}`},
			},
		}
	})
	tests.Add("sort descending with index", func(t *testing.T) any {
		d := newDB(t)
		err := d.CreateIndex(context.Background(), "_design/idx", "byName", json.RawMessage(`{"fields":["name"]}`), mock.NilOption)
		if err != nil {
			t.Fatalf("CreateIndex failed: %s", err)
		}
		revAlice := d.tPut("alice", map[string]string{"name": "Charlie"})
		revBob := d.tPut("bob", map[string]string{"name": "Alice"})
		revCharlie := d.tPut("charlie", map[string]string{"name": "Bob"})

		return test{
			db:    d,
			query: `{"selector":{},"sort":[{"name":"desc"}]}`,
			want: []rowResult{
				{Doc: `{"_id":"alice","_rev":"` + revAlice + `","name":"Charlie"}`},
				{Doc: `{"_id":"charlie","_rev":"` + revCharlie + `","name":"Bob"}`},
				{Doc: `{"_id":"bob","_rev":"` + revBob + `","name":"Alice"}`},
			},
		}
	})
	tests.Add("sort mixed directions rejects without matching index", func(t *testing.T) any {
		d := newDB(t)
		err := d.CreateIndex(context.Background(), "", "", json.RawMessage(`{"fields":["a","b"]}`), mock.NilOption)
		if err != nil {
			t.Fatalf("CreateIndex failed: %s", err)
		}
		d.tPut("doc1", map[string]string{"a": "x", "b": "y"})

		return test{
			db:         d,
			query:      `{"selector":{},"sort":[{"a":"asc"},{"b":"desc"}]}`,
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
	tests.Add("sort, invalid direction", test{
		query:      `{"selector":{},"sort":[{"name":"foo"}]}`,
		wantStatus: http.StatusBadRequest,
		wantErr:    "invalid sort direction",
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
	tests.Add("find with index", func(t *testing.T) any {
		d := newDB(t)
		err := d.CreateIndex(context.Background(), "_design/idx", "idx", json.RawMessage(`{"fields":["name"]}`), mock.NilOption)
		if err != nil {
			t.Fatalf("CreateIndex failed: %s", err)
		}
		_ = d.tPut("alice", map[string]string{"name": "Alice"})
		revBob := d.tPut("bob", map[string]string{"name": "Bob"})
		_ = d.tPut("charlie", map[string]string{"name": "Charlie"})

		return test{
			db:    d,
			query: `{"selector":{"name":"Bob"}}`,
			want: []rowResult{
				{Doc: `{"_id":"bob","_rev":"` + revBob + `","name":"Bob"}`},
			},
		}
	})
	tests.Add("use_index", func(t *testing.T) any {
		d := newDB(t)
		err := d.CreateIndex(context.Background(), "_design/myidx", "byName", json.RawMessage(`{"fields":["name"]}`), mock.NilOption)
		if err != nil {
			t.Fatalf("CreateIndex failed: %s", err)
		}
		revBob := d.tPut("bob", map[string]string{"name": "Bob"})
		_ = d.tPut("alice", map[string]string{"name": "Alice"})

		return test{
			db:    d,
			query: `{"selector":{"name":"Bob"},"use_index":"_design/myidx"}`,
			want: []rowResult{
				{Doc: `{"_id":"bob","_rev":"` + revBob + `","name":"Bob"}`},
			},
		}
	})
	tests.Add("use_index, not found", test{
		query:      `{"selector":{},"use_index":"_design/nonexistent"}`,
		wantStatus: http.StatusBadRequest,
		wantErr:    `index "_design/nonexistent" not found`,
	})
	tests.Add("use_index, array form", func(t *testing.T) any {
		d := newDB(t)
		err := d.CreateIndex(context.Background(), "_design/myidx", "byName", json.RawMessage(`{"fields":["name"]}`), mock.NilOption)
		if err != nil {
			t.Fatalf("CreateIndex failed: %s", err)
		}
		revBob := d.tPut("bob", map[string]string{"name": "Bob"})
		_ = d.tPut("alice", map[string]string{"name": "Alice"})

		return test{
			db:    d,
			query: `{"selector":{"name":"Bob"},"use_index":["_design/myidx","byName"]}`,
			want: []rowResult{
				{Doc: `{"_id":"bob","_rev":"` + revBob + `","name":"Bob"}`},
			},
		}
	})
	tests.Add("use_index with sort, wrong index", func(t *testing.T) any {
		d := newDB(t)
		err := d.CreateIndex(context.Background(), "_design/nameIdx", "byName", json.RawMessage(`{"fields":["name"]}`), mock.NilOption)
		if err != nil {
			t.Fatalf("CreateIndex failed: %s", err)
		}
		err = d.CreateIndex(context.Background(), "_design/ageIdx", "byAge", json.RawMessage(`{"fields":["age"]}`), mock.NilOption)
		if err != nil {
			t.Fatalf("CreateIndex failed: %s", err)
		}
		d.tPut("doc1", map[string]any{"name": "Alice", "age": 30})

		return test{
			db:         d,
			query:      `{"selector":{},"sort":["name"],"use_index":"_design/ageIdx"}`,
			wantStatus: http.StatusBadRequest,
			wantErr:    "no index exists for this sort, try indexing by the sort fields",
		}
	})
	tests.Add("field name with single quote", func(t *testing.T) any {
		d := newDB(t)
		err := d.CreateIndex(context.Background(), "_design/idx", "byQuote", json.RawMessage(`{"fields":["o'brian"]}`), mock.NilOption)
		if err != nil {
			t.Fatalf("CreateIndex failed: %s", err)
		}
		rev := d.tPut("doc1", map[string]any{"o'brian": "yes"})

		return test{
			db:    d,
			query: `{"selector":{"o'brian":"yes"}}`,
			want: []rowResult{
				{Doc: `{"_id":"doc1","_rev":"` + rev + `","o'brian":"yes"}`},
			},
		}
	})

	tests.Run(t, func(t *testing.T, tt test) {
		t.Parallel()
		db := tt.db
		if db == nil {
			db = newDB(t)
		}
		rows, err := db.Find(context.Background(), json.RawMessage(tt.query), mock.NilOption)
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

func TestFindUsesIndex(t *testing.T) {
	t.Parallel()
	d := newDB(t)

	err := d.CreateIndex(context.Background(), "_design/idx", "idx", json.RawMessage(`{"fields":["name"]}`), mock.NilOption)
	if err != nil {
		t.Fatalf("CreateIndex failed: %s", err)
	}

	for i := 0; i < 100; i++ {
		id := fmt.Sprintf("doc%03d", i)
		_ = d.tPut(id, map[string]string{"name": fmt.Sprintf("Name%03d", i)})
	}

	underlying := d.underlying()
	tableName := `"kivik$test"`
	indexName := mangoIndexName("test", "_design/idx", "idx")

	// Verify the index exists.
	var count int
	err = underlying.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name=` + indexName).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query sqlite_master: %s", err)
	}
	if count != 1 {
		t.Fatalf("Expected 1 index, got %d", count)
	}

	// Run EXPLAIN QUERY PLAN on a query that mirrors what leavesCTE generates.
	query := `EXPLAIN QUERY PLAN
		SELECT *
		FROM ` + tableName + ` AS doc
		WHERE json_extract(doc.doc, '$.name') = 'Name050'`

	rows, err := underlying.Query(query)
	if err != nil {
		t.Fatalf("EXPLAIN QUERY PLAN failed: %s", err)
	}
	defer rows.Close()

	var found bool
	for rows.Next() {
		var id, parent, notused int
		var detail string
		if err := rows.Scan(&id, &parent, &notused, &detail); err != nil {
			t.Fatalf("Scan failed: %s", err)
		}
		t.Logf("EXPLAIN: %s", detail)
		if strings.Contains(detail, "USING INDEX") && strings.Contains(detail, "mango") {
			found = true
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if !found {
		t.Error("Expression index was not used by query plan")
	}
}

func Test_selectorToSQL(t *testing.T) {
	t.Parallel()

	type test struct {
		selector  json.RawMessage
		argOffset int
		wantConds []string
		wantArgs  []any
	}

	tests := testy.NewTable()

	tests.Add("implicit eq", test{
		selector:  json.RawMessage(`{"name": "Bob"}`),
		argOffset: 0,
		wantConds: []string{`json_extract(doc.doc, "$.name") = $1`},
		wantArgs:  []any{"Bob"},
	})
	tests.Add("$gt with number", test{
		selector:  json.RawMessage(`{"age": {"$gt": 21}}`),
		argOffset: 0,
		wantConds: []string{`json_type(doc.doc, "$.age") NOT IN ('integer', 'real') OR json_extract(doc.doc, "$.age") > $1`},
		wantArgs:  []any{float64(21)},
	})
	tests.Add("$lt with string", test{
		selector:  json.RawMessage(`{"name": {"$lt": "M"}}`),
		argOffset: 0,
		wantConds: []string{`json_type(doc.doc, "$.name") != 'text' OR json_extract(doc.doc, "$.name") < $1`},
		wantArgs:  []any{"M"},
	})
	tests.Add("$gte with number", test{
		selector:  json.RawMessage(`{"score": {"$gte": 90.5}}`),
		argOffset: 0,
		wantConds: []string{`json_type(doc.doc, "$.score") NOT IN ('integer', 'real') OR json_extract(doc.doc, "$.score") >= $1`},
		wantArgs:  []any{float64(90.5)},
	})
	tests.Add("$exists true", test{
		selector:  json.RawMessage(`{"name": {"$exists": true}}`),
		argOffset: 0,
		wantConds: []string{`json_extract(doc.doc, "$.name") IS NOT NULL`},
		wantArgs:  nil,
	})
	tests.Add("$in", test{
		selector:  json.RawMessage(`{"status": {"$in": ["active", "pending"]}}`),
		argOffset: 0,
		wantConds: []string{`json_extract(doc.doc, "$.status") IN ($1, $2)`},
		wantArgs:  []any{"active", "pending"},
	})
	tests.Add("argOffset", test{
		selector:  json.RawMessage(`{"name": "Bob"}`),
		argOffset: 5,
		wantConds: []string{`json_extract(doc.doc, "$.name") = $6`},
		wantArgs:  []any{"Bob"},
	})
	tests.Add("unsupported operator skipped", test{
		selector:  json.RawMessage(`{"name": {"$regex": "^B"}}`),
		argOffset: 0,
		wantConds: nil,
		wantArgs:  nil,
	})
	tests.Add("$and", test{
		selector:  json.RawMessage(`{"$and": [{"name": "Bob"}, {"age": {"$eq": 21}}]}`),
		argOffset: 0,
		wantConds: []string{`json_extract(doc.doc, "$.name") = $1 AND json_extract(doc.doc, "$.age") = $2`},
		wantArgs:  []any{"Bob", float64(21)},
	})
	tests.Add("$or", test{
		selector:  json.RawMessage(`{"$or": [{"name": "Bob"}, {"name": "Alice"}]}`),
		argOffset: 0,
		wantConds: []string{`(json_extract(doc.doc, "$.name") = $1 OR json_extract(doc.doc, "$.name") = $2)`},
		wantArgs:  []any{"Bob", "Alice"},
	})
	tests.Add("null eq", test{
		selector:  json.RawMessage(`{"name": {"$eq": null}}`),
		argOffset: 0,
		wantConds: []string{`json_extract(doc.doc, "$.name") IS NULL`},
		wantArgs:  nil,
	})
	tests.Add("boolean true", test{
		selector:  json.RawMessage(`{"active": true}`),
		argOffset: 0,
		wantConds: []string{`json_extract(doc.doc, "$.active") = $1`},
		wantArgs:  []any{1},
	})

	tests.Run(t, func(t *testing.T, tt test) {
		t.Parallel()
		gotConds, gotArgs := selectorToSQL(tt.selector, tt.argOffset)
		if d := cmp.Diff(tt.wantConds, gotConds); d != "" {
			t.Errorf("conditions mismatch (-want +got):\n%s", d)
		}
		if d := cmp.Diff(tt.wantArgs, gotArgs); d != "" {
			t.Errorf("args mismatch (-want +got):\n%s", d)
		}
	})
}
