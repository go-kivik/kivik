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

package db

import (
	"context"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
)

func init() {
	kt.Register("Explain", explain)
}

func explain(t *testing.T, c *kt.Context) {
	t.Helper()
	c.RunAdmin(t, func(t *testing.T) {
		t.Helper()
		testExplain(t, c, c.Admin)
	})
	c.RunNoAuth(t, func(t *testing.T) {
		t.Helper()
		testExplain(t, c, c.NoAuth)
	})
	c.RunRW(t, func(t *testing.T) {
		t.Helper()
		testExplainRW(t, c)
	})
}

func testExplainRW(t *testing.T, c *kt.Context) {
	t.Helper()
	if c.Admin == nil {
		return
	}
	dbName := c.TestDB(t)
	c.RunAdmin(t, func(t *testing.T) {
		t.Helper()
		doExplainTest(t, c, c.Admin, dbName)
	})
	c.RunNoAuth(t, func(t *testing.T) {
		t.Helper()
		doExplainTest(t, c, c.NoAuth, dbName)
	})
}

func testExplain(t *testing.T, c *kt.Context, client *kivik.Client) { //nolint:thelper
	if !c.IsSet(t, "databases") {
		t.Errorf("databases not set; Did you configure this test?")
		return
	}
	for _, dbName := range c.StringSlice(t, "databases") {
		func(dbName string) {
			c.Run(t, dbName, func(t *testing.T) {
				doExplainTest(t, c, client, dbName)
			})
		}(dbName)
	}
}

func doExplainTest(t *testing.T, c *kt.Context, client *kivik.Client, dbName string) { //nolint:thelper
	const limit = 25
	t.Parallel()
	db := client.DB(dbName, c.Options(t, "db"))
	// Errors may be deferred here, so only return if we actually get
	// an error.
	if err := db.Err(); err != nil && !c.IsExpectedSuccess(t, err) {
		return
	}

	var plan *kivik.QueryPlan
	err := kt.Retry(func() error {
		var e error
		plan, e = db.Explain(context.Background(), `{"selector":{"_id":{"$gt":null}}}`)
		return e
	})
	if !c.IsExpectedSuccess(t, err) {
		return
	}
	expected := new(kivik.QueryPlan)
	if e, ok := c.Interface(t, "plan").(*kivik.QueryPlan); ok {
		*expected = *e // Make a shallow copy
	} else {
		expected = &kivik.QueryPlan{
			Index: map[string]any{
				"ddoc": nil,
				"name": "_all_docs",
				"type": "special",
				"def":  map[string]any{"fields": []any{map[string]string{"_id": "asc"}}},
			},
			Selector: map[string]any{"_id": map[string]any{"$gt": nil}},
			Options: map[string]any{
				"bookmark":  "nil",
				"conflicts": false,
				"fields":    "all_fields",
				"limit":     limit,
				"r":         []int{49},
				"skip":      0,
				"sort":      map[string]any{},
				"use_index": []any{},
			},
			Limit: limit,
			Range: map[string]any{
				"start_key": nil,
				"end_key":   "\xef\xbf\xbd",
			},
		}
	}
	expected.DBName = dbName
	if d := testy.DiffAsJSON(expected, plan); d != nil {
		t.Errorf("Unexpected plan returned:\n%s\n", d)
	}
}
