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
	kt.Register("BulkDocs", bulkDocs)
}

func bulkDocs(t *testing.T, c *kt.Context) {
	t.Helper()
	c.RunRW(t, func(t *testing.T) {
		t.Helper()
		c.RunAdmin(t, func(t *testing.T) {
			t.Helper()
			testBulkDocs(t, c, c.Admin)
		})
		c.RunNoAuth(t, func(t *testing.T) {
			t.Helper()
			testBulkDocs(t, c, c.NoAuth)
		})
	})
}

func failOnBulkErrors(t *testing.T, updates []kivik.BulkResult, op string) { //nolint:thelper
	for _, update := range updates {
		if update.Error != nil {
			t.Errorf("Bulk %s failed: %s", op, update.Error)
		}
	}
}

func testBulkDocs(t *testing.T, c *kt.Context, client *kivik.Client) { //nolint:thelper
	t.Parallel()
	dbname := c.TestDB(t)
	adb := c.Admin.DB(dbname, c.Options(t, "db"))
	if err := adb.Err(); err != nil {
		t.Fatalf("Failed to connect to db as admin: %s", err)
	}
	db := client.DB(dbname, c.Options(t, "db"))
	if err := db.Err(); err != nil {
		t.Fatalf("Failed to connect to db: %s", err)
	}
	c.Run(t, "Create", func(t *testing.T) {
		t.Parallel()
		doc := map[string]string{
			"name": "Robert",
		}
		var updates []kivik.BulkResult
		err := kt.Retry(func() error {
			var err error
			updates, err = db.BulkDocs(context.Background(), []any{doc})
			return err
		})
		if !c.IsExpectedSuccess(t, err) {
			return
		}
		failOnBulkErrors(t, updates, "create")
	})
	c.Run(t, "Update", func(t *testing.T) {
		t.Parallel()
		doc := map[string]string{
			"_id":  kt.TestDBName(t),
			"name": "Alice",
		}
		rev, err := adb.Put(context.Background(), doc["_id"], doc)
		if err != nil {
			t.Fatalf("Failed to create doc: %s", err)
		}
		doc["_rev"] = rev
		var updates []kivik.BulkResult
		err = kt.Retry(func() error {
			var err error
			updates, err = db.BulkDocs(context.Background(), []any{doc})
			return err
		})
		if !c.IsExpectedSuccess(t, err) {
			return
		}
		failOnBulkErrors(t, updates, "update")
	})
	c.Run(t, "Delete", func(t *testing.T) {
		t.Parallel()
		id := kt.TestDBName(t)
		doc := map[string]any{
			"_id":  id,
			"name": "Alice",
		}
		rev, err := adb.Put(context.Background(), id, doc)
		if err != nil {
			t.Fatalf("Failed to create doc: %s", err)
		}
		doc["_rev"] = rev
		doc["_deleted"] = true
		var updates []kivik.BulkResult
		err = kt.Retry(func() error {
			var err error
			updates, err = db.BulkDocs(context.Background(), []any{doc})
			return err
		})
		if !c.IsExpectedSuccess(t, err) {
			return
		}
		failOnBulkErrors(t, updates, "delete")
	})
	c.Run(t, "Mix", func(t *testing.T) {
		t.Parallel()

		doc0 := map[string]string{
			"name": "Fred",
		}

		id1 := kt.TestDBName(t)
		doc1 := map[string]any{
			"_id":  id1,
			"name": "Robert",
		}

		rev1, err := adb.Put(context.Background(), id1, doc1)
		if err != nil {
			t.Fatalf("Failed to create doc1: %s", err)
		}
		doc1["_rev"] = rev1

		id2 := kt.TestDBName(t)
		doc2 := map[string]any{
			"_id":  id2,
			"name": "Alice",
		}
		rev2, err := adb.Put(context.Background(), id2, doc2)
		if err != nil {
			t.Fatalf("Failed to create doc2: %s", err)
		}
		doc2["_rev"] = rev2
		doc2["_deleted"] = true

		id3 := kt.TestDBName(t)
		doc3 := map[string]string{
			"_id": id3,
		}
		_, err = adb.Put(context.Background(), id3, doc3)
		if err != nil {
			t.Fatalf("Failed to create doc2: %s", err)
		}

		var updates []kivik.BulkResult

		err = kt.Retry(func() error {
			var err error
			updates, err = db.BulkDocs(context.Background(), []any{doc0, doc1, doc2, doc3})
			return err
		})
		if !c.IsExpectedSuccess(t, err) {
			return
		}
		for _, update := range updates {
			var testName string
			switch update.ID {
			case id3:
				testName = "Conflict"
			case id1:
				testName = "Update"
			case id2:
				testName = "Delete"
			default:
				testName = "Create"
			}
			c.Run(t, testName, func(t *testing.T) {
				c.CheckError(t, update.Error)
			})
		}
	})
	c.Run(t, "NonJSON", func(t *testing.T) {
		const age = 32
		t.Parallel()
		id1 := kt.TestDBName(t)
		id2 := kt.TestDBName(t)
		docs := []any{
			struct {
				ID   string `json:"_id"`
				Name string `json:"name"`
			}{ID: id1, Name: "Robert"},
			struct {
				ID   string `json:"_id"`
				Name string `json:"name"`
				Age  int    `json:"the_age"`
			}{ID: id2, Name: "Alice", Age: age},
		}
		var updates []kivik.BulkResult
		err := kt.Retry(func() error {
			var err error
			updates, err = db.BulkDocs(context.Background(), docs)
			return err
		})
		if !c.IsExpectedSuccess(t, err) {
			return
		}
		failOnBulkErrors(t, updates, "create")
		c.Run(t, "Retrieve", func(t *testing.T) {
			var result map[string]any
			if err = db.Get(context.Background(), id2).ScanDoc(&result); err != nil {
				t.Fatalf("failed to scan bulk-inserted document: %s", err)
			}
			expected := map[string]any{
				"_id":     id2,
				"name":    "Alice",
				"the_age": age,
				"_rev":    result["_rev"],
			}
			if d := testy.DiffAsJSON(expected, result); d != nil {
				t.Errorf("Retrieved document differs:\n%s\n", d)
			}
		})
	})
}
