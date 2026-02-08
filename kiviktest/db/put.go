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

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
)

func init() {
	kt.Register("Put", put)
}

func put(t *testing.T, c *kt.Context) {
	t.Helper()
	c.RunRW(t, func(t *testing.T) {
		t.Helper()
		c.RunAdmin(t, func(t *testing.T) {
			t.Helper()
			testPut(t, c, c.Admin)
		})
		c.RunNoAuth(t, func(t *testing.T) {
			t.Helper()
			testPut(t, c, c.NoAuth)
		})
	})
}

func testPut(t *testing.T, c *kt.Context, client *kivik.Client) { //nolint:thelper
	t.Parallel()
	dbName := c.TestDB(t)
	db := client.DB(dbName, c.Options(t, "db"))
	if err := db.Err(); !c.IsExpectedSuccess(t, err) {
		return
	}
	c.Run(t, "Create", func(t *testing.T) {
		const age = 32
		t.Parallel()

		doc := &testDoc{
			ID:   kt.TestDBName(t),
			Name: "Alberto",
			Age:  age,
		}
		var rev string
		err := kt.Retry(func() error {
			var e error
			rev, e = db.Put(context.Background(), doc.ID, doc)
			return e
		})
		if !c.IsExpectedSuccess(t, err) {
			return
		}
		doc.Rev = rev
		doc.Age = 40
		c.Run(t, "Update", func(t *testing.T) {
			err := kt.Retry(func() error {
				_, e := db.Put(context.Background(), doc.ID, doc)
				return e
			})
			c.CheckError(t, err)
		})
	})
	c.Run(t, "DesignDoc", func(t *testing.T) {
		t.Parallel()
		doc := map[string]any{
			"_id":      "_design/testddoc",
			"language": "javascript",
			"views": map[string]any{
				"testview": map[string]any{
					"map": `function(doc) {
			                if (doc.include) {
			                    emit(doc._id, doc.index);
			                }
			            }`,
				},
			},
		}
		err := kt.Retry(func() error {
			_, err := db.Put(context.Background(), doc["_id"].(string), doc)
			return err
		})
		c.CheckError(t, err)
	})
	c.Run(t, "Local", func(t *testing.T) {
		t.Parallel()
		doc := map[string]any{
			"_id":  "_local/foo",
			"name": "Bob",
		}
		err := kt.Retry(func() error {
			_, err := db.Put(context.Background(), doc["_id"].(string), doc)
			return err
		})
		c.CheckError(t, err)
	})
	c.Run(t, "LeadingUnderscoreInID", func(t *testing.T) {
		t.Parallel()
		doc := map[string]any{
			"_id":  "_badid",
			"name": "Bob",
		}
		err := kt.Retry(func() error {
			_, err := db.Put(context.Background(), doc["_id"].(string), doc)
			return err
		})
		c.CheckError(t, err)
	})
	c.Run(t, "HeavilyEscapedID", func(t *testing.T) {
		t.Parallel()
		doc := map[string]any{
			"_id":  "foo+bar & spáces ?!*,",
			"name": "Bob",
		}
		err := kt.Retry(func() error {
			_, err := db.Put(context.Background(), doc["_id"].(string), doc)
			return err
		})
		c.CheckError(t, err)
	})
	c.Run(t, "SlashInID", func(t *testing.T) {
		t.Parallel()
		doc := map[string]any{
			"_id":  "foo/bar",
			"name": "Bob",
		}
		err := kt.Retry(func() error {
			_, err := db.Put(context.Background(), doc["_id"].(string), doc)
			return err
		})
		c.CheckError(t, err)
	})
	c.Run(t, "Conflict", func(t *testing.T) {
		t.Parallel()
		const id = "duplicate"
		doc := map[string]any{
			"_id":  id,
			"name": "Bob",
		}
		err := kt.Retry(func() error {
			_, err := c.Admin.DB(dbName).Put(context.Background(), id, doc)
			return err
		})
		if err != nil {
			t.Fatalf("Failed to create document for duplicate test: %s", err)
		}
		err = kt.Retry(func() error {
			_, err = db.Put(context.Background(), id, doc)
			return err
		})
		c.CheckError(t, err)
	})
}
