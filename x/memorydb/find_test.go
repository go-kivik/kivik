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

package memorydb

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4/driver"
)

func TestIndexSpecUnmarshalJSON(t *testing.T) {
	type isuTest struct {
		name     string
		input    string
		expected *indexSpec
		err      string
	}
	tests := []isuTest{
		{
			name:     "ddoc only",
			input:    `"foo"`,
			expected: &indexSpec{ddoc: "foo"},
		},
		{
			name:     "ddoc and index",
			input:    `["foo","bar"]`,
			expected: &indexSpec{ddoc: "foo", index: "bar"},
		},
		{
			name:  "invalid json",
			input: "asdf",
			err:   "invalid character 'a' looking for beginning of value",
		},
		{
			name:  "extra fields",
			input: `["foo","bar","baz"]`,
			err:   "invalid index specification",
		},
		{
			name:     "One field",
			input:    `["foo"]`,
			expected: &indexSpec{ddoc: "foo"},
		},
		{
			name:  "Empty array",
			input: `[]`,
			err:   "invalid index specification",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := &indexSpec{}
			err := result.UnmarshalJSON([]byte(test.input))
			var msg string
			if err != nil {
				msg = err.Error()
			}
			if msg != test.err {
				t.Errorf("Unexpected error: %s", err)
			}
			if err != nil {
				return
			}
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestCreateIndex(t *testing.T) {
	d := &db{}
	err := d.CreateIndex(t.Context(), "foo", "bar", "baz", nil)
	if err != errFindNotImplemented {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestGetIndexes(t *testing.T) {
	d := &db{}
	_, err := d.GetIndexes(t.Context(), nil)
	if err != errFindNotImplemented {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestDeleteIndex(t *testing.T) {
	d := &db{}
	err := d.DeleteIndex(t.Context(), "foo", "bar", nil)
	if err != errFindNotImplemented {
		t.Errorf("Unexpected error: %s", err)
	}
}

// TestFind tests selectors, to see that the proper doc IDs are returned.
func TestFind(t *testing.T) {
	type findTest struct {
		name        string
		db          *db
		query       interface{}
		expectedIDs []string
		err         string
		rowsErr     string
	}
	tests := []findTest{
		{
			name:  "invalid query",
			query: make(chan int),
			err:   "json: unsupported type: chan int",
		},
		{
			name:  "Invalid JSON query",
			query: "asdf",
			err:   "invalid character 'a' looking for beginning of value",
		},
		{
			name: "No query",
			err:  "missing required key: selector",
		},
		{
			name:  "empty selector",
			query: `{"selector":{}}`,
			db: func() *db {
				db := setupDB(t)
				for _, id := range []string{"a", "c", "z", "q", "chicken"} {
					if _, err := db.Put(t.Context(), id, map[string]string{"value": id}, nil); err != nil {
						t.Fatal(err)
					}
				}
				return db
			}(),
			expectedIDs: []string{"a", "c", "chicken", "q", "z"},
		},
		{
			name:  "simple selector",
			query: `{"selector":{"value":"chicken"}}`,
			db: func() *db {
				db := setupDB(t)
				for _, id := range []string{"a", "c", "z", "q", "chicken"} {
					if _, err := db.Put(t.Context(), id, map[string]string{"value": id}, nil); err != nil {
						t.Fatal(err)
					}
				}
				return db
			}(),
			expectedIDs: []string{"chicken"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := test.db
			if db == nil {
				db = setupDB(t)
			}
			rows, err := db.Find(t.Context(), test.query, nil)
			var msg string
			if err != nil {
				msg = err.Error()
			}
			if msg != test.err {
				t.Errorf("Unexpected error: %s", err)
			}
			if err != nil {
				return
			}
			checkRows(t, rows, test.expectedIDs, test.rowsErr)
		})
	}
}

// TestFindDoc is the same as Testfind, but assumes only a single result
// (ignores any others), and compares the entire document.
func TestFindDoc(t *testing.T) {
	type fdTest struct {
		name     string
		db       *db
		query    interface{}
		expected interface{}
	}
	tests := []fdTest{
		{
			name:  "simple selector",
			query: `{"selector":{}}`,
			db: func() *db {
				db := setupDB(t)
				id := "chicken"
				if _, err := db.Put(t.Context(), id, map[string]string{"value": id}, nil); err != nil {
					t.Fatal(err)
				}
				return db
			}(),
			expected: map[string]interface{}{
				"_id":   "chicken",
				"_rev":  "1-xxx",
				"value": "chicken",
			},
		},
		{
			name:  "fields",
			query: `{"selector":{}, "fields":["value","_rev"]}`,
			db: func() *db {
				db := setupDB(t)
				if _, err := db.Put(t.Context(), "foo", map[string]string{"value": "foo"}, nil); err != nil {
					t.Fatal(err)
				}
				return db
			}(),
			expected: map[string]interface{}{
				"value": "foo",
				"_rev":  "1-xxx",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := test.db
			if db == nil {
				db = setupDB(t)
			}
			rows, err := db.Find(t.Context(), test.query, nil)
			if err != nil {
				t.Fatal(err)
			}
			var row driver.Row
			if e := rows.Next(&row); e != nil {
				t.Fatal(e)
			}
			_ = rows.Close()
			var result map[string]interface{}
			if e := json.NewDecoder(row.Doc).Decode(&result); e != nil {
				t.Fatal(e)
			}
			parts := strings.Split(result["_rev"].(string), "-")
			result["_rev"] = parts[0] + "-xxx"
			if d := testy.DiffAsJSON(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestResultWarning(t *testing.T) {
	rows := &findResults{}
	expected := "no matching index found, create an index to optimize query time"
	if w := rows.Warning(); w != expected {
		t.Errorf("Unexpected warning: %s", w)
	}
}

func TestFilterDoc(t *testing.T) {
	type fdTest struct {
		name     string
		rows     *findResults
		data     string
		expected string
		err      string
	}
	tests := []fdTest{
		{
			name:     "no filter",
			rows:     &findResults{},
			data:     `{"foo":"bar"}`,
			expected: `{"foo":"bar"}`,
		},
		{
			name:     "with filter",
			rows:     &findResults{fields: map[string]struct{}{"foo": {}}},
			data:     `{"foo":"bar", "baz":"qux"}`,
			expected: `{"foo":"bar"}`,
		},
		{
			name: "invalid json",
			rows: &findResults{fields: map[string]struct{}{"foo": {}}},
			data: `{"foo":"bar", "baz":"qux}`,
			err:  "unexpected end of JSON input",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.rows.filterDoc([]byte(test.data))
			var msg string
			if err != nil {
				msg = err.Error()
			}
			if msg != test.err {
				t.Errorf("Unexpected error: %s", msg)
			}
			if err != nil {
				return
			}
			if d := testy.DiffJSON([]byte(test.expected), result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestToJSON(t *testing.T) {
	type tjTest struct {
		Name     string
		Input    interface{}
		Expected string
	}
	tests := []tjTest{
		{
			Name:     "Null",
			Expected: "null",
		},
		{
			Name:     "String",
			Input:    `{"foo":"bar"}`,
			Expected: `{"foo":"bar"}`,
		},
		{
			Name:     "ByteSlice",
			Input:    []byte(`{"foo":"bar"}`),
			Expected: `{"foo":"bar"}`,
		},
		{
			Name:     "RawMessage",
			Input:    json.RawMessage(`{"foo":"bar"}`),
			Expected: `{"foo":"bar"}`,
		},
		{
			Name:     "Interface",
			Input:    map[string]string{"foo": "bar"},
			Expected: `{"foo":"bar"}`,
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			r, err := toJSON(test.Input)
			if err != nil {
				t.Fatalf("jsonify failed: %s", err)
			}
			buf := &bytes.Buffer{}
			_, _ = buf.ReadFrom(r)
			result := strings.TrimSpace(buf.String())
			if result != test.Expected {
				t.Errorf("Expected: `%s`\n  Actual: `%s`", test.Expected, result)
			}
		})
	}
}
