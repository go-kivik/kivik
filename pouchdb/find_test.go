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

//go:build js

package pouchdb

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/gopherjs/gopherjs/js"
	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/pouchdb/bindings"
	"github.com/go-kivik/kivik/v4/pouchdb/internal"
)

func init() {
	memPouch := js.Global.Get("PouchDB").Call("defaults", map[string]interface{}{
		"db": js.Global.Call("require", "memdown"),
	})
	js.Global.Set("PouchDB", memPouch)
}

func TestBuildIndex(t *testing.T) {
	tests := []struct {
		Ddoc     string
		Name     string
		Index    interface{}
		Expected string
	}{
		{Expected: `{}`},
		{Index: `{"fields":["foo"]}`, Expected: `{"fields":["foo"]}`},
		{Index: `{"fields":["foo"]}`, Name: "test", Expected: `{"fields":["foo"],"name":"test"}`},
		{Index: `{"fields":["foo"]}`, Name: "test", Ddoc: "_foo", Expected: `{"fields":["foo"],"name":"test","ddoc":"_foo"}`},
	}
	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			result, err := buildIndex(test.Ddoc, test.Name, test.Index)
			if err != nil {
				t.Errorf("Build Index failed: %s", err)
			}
			r := js.Global.Get("JSON").Call("stringify", result).String()
			if d := testy.DiffJSON([]byte(test.Expected), []byte(r)); d != nil {
				t.Errorf("BuildIndex result differs:\n%s\n", d)
			}
		})
	}
}

func TestExplain(t *testing.T) {
	defaultLimit := int64(0)
	if ver := internal.PouchDBVersion(t); strings.HasPrefix(ver, "9") {
		defaultLimit = 25
	}
	type test struct {
		db       *db
		query    interface{}
		expected *driver.QueryPlan
		err      string
	}
	tests := testy.NewTable()
	tests.Add("query error", test{
		db:    &db{db: bindings.GlobalPouchDB().New("foo", nil)},
		query: nil,
		err:   "TypeError: Cannot read propert",
	})
	tests.Add("simple selector", func(t *testing.T) interface{} {
		options := map[string]interface{}{
			"bookmark":  "nil",
			"conflicts": false,
			"r":         []interface{}{49},
			"sort":      map[string]interface{}{},
			"use_index": []interface{}{},
		}
		if defaultLimit > 0 {
			options["limit"] = defaultLimit
		}

		return test{
			db:    &db{db: bindings.GlobalPouchDB().New("foo", nil)},
			query: map[string]interface{}{"selector": map[string]interface{}{"_id": "foo"}},
			expected: &driver.QueryPlan{
				DBName: "foo",
				Limit:  defaultLimit,
				Index: map[string]interface{}{
					"ddoc": nil,
					"def": map[string]interface{}{
						"fields": []interface{}{map[string]interface{}{"_id": "asc"}},
					},
					"name": "_all_docs",
					"type": "special",
				},
				Options:  options,
				Selector: map[string]interface{}{"_id": map[string]interface{}{"$eq": "foo"}},
				Fields:   nil,
				Range:    map[string]interface{}{},
			},
		}
	})
	tests.Add("fields list", func(t *testing.T) interface{} {
		options := map[string]interface{}{
			"bookmark":  "nil",
			"conflicts": false,
			"fields":    []interface{}{"_id", map[string]interface{}{"type": "desc"}},
			"r":         []interface{}{49},
			"sort":      map[string]interface{}{},
			"use_index": []interface{}{},
		}
		if defaultLimit > 0 {
			options["limit"] = defaultLimit
		}

		return test{
			db: &db{db: bindings.GlobalPouchDB().New("foo", nil)},
			query: map[string]interface{}{
				"selector": map[string]interface{}{"_id": "foo"},
				"fields":   []interface{}{"_id", map[string]interface{}{"type": "desc"}},
			},
			expected: &driver.QueryPlan{
				DBName: "foo",
				Limit:  defaultLimit,
				Index: map[string]interface{}{
					"ddoc": nil,
					"def": map[string]interface{}{
						"fields": []interface{}{map[string]interface{}{"_id": "asc"}},
					},
					"name": "_all_docs",
					"type": "special",
				},
				Options:  options,
				Selector: map[string]interface{}{"_id": map[string]interface{}{"$eq": "foo"}},
				Fields:   []interface{}{"_id", map[string]interface{}{"type": "desc"}},
				Range:    map[string]interface{}{},
			},
		}
	})
	tests.Run(t, func(t *testing.T, tt test) {
		result, err := tt.db.Explain(context.Background(), tt.query, nil)
		if !testy.ErrorMatchesRE(tt.err, err) {
			t.Errorf("Unexpected error: %s", err)
		}
		if err != nil {
			return
		}
		if d := testy.DiffAsJSON(tt.expected, result); d != nil {
			t.Error(d)
		}
	})
}

func TestUnmarshalQueryPlan(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *queryPlan
		err      string
	}{
		{
			name:  "non-array",
			input: `{"fields":{}}`,
			err:   "json: cannot unmarshal object into Go",
		},
		{
			name:     "all_fields",
			input:    `{"fields":"all_fields","dbname":"foo"}`,
			expected: &queryPlan{DBName: "foo"},
		},
		{
			name:     "simple field list",
			input:    `{"fields":["foo","bar"],"dbname":"foo"}`,
			expected: &queryPlan{Fields: []interface{}{"foo", "bar"}, DBName: "foo"},
		},
		{
			name:  "complex field list",
			input: `{"dbname":"foo", "fields":[{"foo":"asc"},{"bar":"desc"}]}`,
			expected: &queryPlan{
				DBName: "foo",
				Fields: []interface{}{
					map[string]interface{}{"foo": "asc"},
					map[string]interface{}{"bar": "desc"},
				},
			},
		},
		{
			name:  "invalid bare string",
			input: `{"fields":"not_all_fields"}`,
			err:   "json: cannot unmarshal string into Go",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := new(queryPlan)
			err := json.Unmarshal([]byte(test.input), &result)
			if !testy.ErrorMatchesRE(test.err, err) {
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
