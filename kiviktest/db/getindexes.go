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
	kt.Register("GetIndexes", getIndexes)
}

func getIndexes(t *testing.T, c *kt.Context) {
	t.Helper()
	c.RunAdmin(t, func(t *testing.T) {
		t.Helper()
		t.Parallel()
		roGetIndexesTests(t, c, c.Admin)
	})
	c.RunNoAuth(t, func(t *testing.T) {
		t.Helper()
		t.Parallel()
		roGetIndexesTests(t, c, c.NoAuth)
	})
	c.RunRW(t, func(t *testing.T) {
		t.Helper()
		c.RunAdmin(t, func(t *testing.T) {
			t.Helper()
			t.Parallel()
			rwGetIndexesTests(t, c, c.Admin)
		})
		c.RunNoAuth(t, func(t *testing.T) {
			t.Helper()
			t.Parallel()
			rwGetIndexesTests(t, c, c.NoAuth)
		})
	})
}

func roGetIndexesTests(t *testing.T, c *kt.Context, client *kivik.Client) { //nolint:thelper
	databases := c.MustStringSlice(t, "databases")
	for _, dbname := range databases {
		func(dbname string) {
			c.Run(t, dbname, func(t *testing.T) {
				t.Parallel()
				testGetIndexes(t, c, client, dbname, c.Interface(t, "indexes"))
			})
		}(dbname)
	}
}

func rwGetIndexesTests(t *testing.T, c *kt.Context, client *kivik.Client) { //nolint:thelper
	dbname := c.TestDB(t)
	dba := c.Admin.DB(dbname, c.Options(t, "db"))
	if err := dba.Err(); err != nil {
		t.Fatalf("Failed to open db as admin: %s", err)
	}
	if err := dba.CreateIndex(context.Background(), "foo", "bar", `{"fields":["foo"]}`); err != nil {
		t.Fatalf("Failed to create index: %s", err)
	}
	indexes := c.Interface(t, "indexes")
	if indexes == nil {
		indexes = []kivik.Index{
			kt.AllDocsIndex,
			{
				DesignDoc: "_design/foo",
				Name:      "bar",
				Type:      "json",
				Definition: map[string]any{
					"fields": []map[string]string{
						{"foo": "asc"},
					},
				},
			},
		}
		testGetIndexes(t, c, client, dbname, indexes)
	}
}

func testGetIndexes(t *testing.T, c *kt.Context, client *kivik.Client, dbname string, expected any) { //nolint:thelper
	db := client.DB(dbname, c.Options(t, "db"))
	if err := db.Err(); err != nil {
		t.Fatalf("Failed to open db: %s", err)
	}
	indexes, err := db.GetIndexes(context.Background())
	if !c.IsExpectedSuccess(t, err) {
		return
	}
	if d := testy.DiffAsJSON(expected, indexes); d != nil {
		t.Errorf("Indexes differ from expectation:\n%s\n", d)
	}
}
