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

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kiviktest/v4/kt"
)

func init() {
	kt.Register("BulkDocs", bulkDocs)
}

func bulkDocs(ctx *kt.Context) {
	ctx.RunRW(func(ctx *kt.Context) {
		ctx.RunAdmin(func(ctx *kt.Context) {
			testBulkDocs(ctx, ctx.Admin)
		})
		ctx.RunNoAuth(func(ctx *kt.Context) {
			testBulkDocs(ctx, ctx.NoAuth)
		})
	})
}

func testBulkDocs(ctx *kt.Context, client *kivik.Client) { // nolint: gocyclo
	ctx.Parallel()
	dbname := ctx.TestDB()
	defer ctx.DestroyDB(dbname)
	adb := ctx.Admin.DB(dbname, ctx.Options("db"))
	if err := adb.Err(); err != nil {
		ctx.Fatalf("Failed to connect to db as admin: %s", err)
	}
	db := client.DB(dbname, ctx.Options("db"))
	if err := db.Err(); err != nil {
		ctx.Fatalf("Failed to connect to db: %s", err)
	}
	ctx.Run("group", func(ctx *kt.Context) {
		ctx.Run("Create", func(ctx *kt.Context) {
			ctx.Parallel()
			doc := map[string]string{
				"name": "Robert",
			}
			var updates []kivik.BulkResult
			err := kt.Retry(func() error {
				var err error
				updates, err = db.BulkDocs(context.Background(), []interface{}{doc})
				return err
			})
			if !ctx.IsExpectedSuccess(err) {
				return
			}
			for _, update := range updates {
				if update.Error != nil {
					ctx.Errorf("Bulk create failed: %s", update.Error)
				}
			}
			if err != nil {
				ctx.Errorf("Iteration error: %s", err)
			}
		})
		ctx.Run("Update", func(ctx *kt.Context) {
			ctx.Parallel()
			doc := map[string]string{
				"_id":  ctx.TestDBName(),
				"name": "Alice",
			}
			rev, err := adb.Put(context.Background(), doc["_id"], doc)
			if err != nil {
				ctx.Fatalf("Failed to create doc: %s", err)
			}
			doc["_rev"] = rev
			var updates []kivik.BulkResult
			err = kt.Retry(func() error {
				var err error
				updates, err = db.BulkDocs(context.Background(), []interface{}{doc})
				return err
			})
			if !ctx.IsExpectedSuccess(err) {
				return
			}
			for _, update := range updates {
				if update.Error != nil {
					ctx.Errorf("Bulk delete failed: %s", update.Error)
				}
			}
			if err != nil {
				ctx.Errorf("Iteration error: %s", err)
			}
		})
		ctx.Run("Delete", func(ctx *kt.Context) {
			ctx.Parallel()
			id := ctx.TestDBName()
			doc := map[string]interface{}{
				"_id":  id,
				"name": "Alice",
			}
			rev, err := adb.Put(context.Background(), id, doc)
			if err != nil {
				ctx.Fatalf("Failed to create doc: %s", err)
			}
			doc["_rev"] = rev
			doc["_deleted"] = true
			var updates []kivik.BulkResult
			err = kt.Retry(func() error {
				var err error
				updates, err = db.BulkDocs(context.Background(), []interface{}{doc})
				return err
			})
			if !ctx.IsExpectedSuccess(err) {
				return
			}
			for _, update := range updates {
				if update.Error != nil {
					ctx.Errorf("Bulk update failed: %s", update.Error)
				}
			}
			if err != nil {
				ctx.Errorf("Iteration error: %s", err)
			}
		})
		ctx.Run("Mix", func(ctx *kt.Context) {
			ctx.Parallel()

			doc0 := map[string]string{
				"name": "Fred",
			}

			id1 := ctx.TestDBName()
			doc1 := map[string]interface{}{
				"_id":  id1,
				"name": "Robert",
			}

			rev1, err := adb.Put(context.Background(), id1, doc1)
			if err != nil {
				ctx.Fatalf("Failed to create doc1: %s", err)
			}
			doc1["_rev"] = rev1

			id2 := ctx.TestDBName()
			doc2 := map[string]interface{}{
				"_id":  id2,
				"name": "Alice",
			}
			rev2, err := adb.Put(context.Background(), id2, doc2)
			if err != nil {
				ctx.Fatalf("Failed to create doc2: %s", err)
			}
			doc2["_rev"] = rev2
			doc2["_deleted"] = true

			id3 := ctx.TestDBName()
			doc3 := map[string]string{
				"_id": id3,
			}
			_, err = adb.Put(context.Background(), id3, doc3)
			if err != nil {
				ctx.Fatalf("Failed to create doc2: %s", err)
			}

			var updates []kivik.BulkResult

			err = kt.Retry(func() error {
				var err error
				updates, err = db.BulkDocs(context.Background(), []interface{}{doc0, doc1, doc2, doc3})
				return err
			})
			if !ctx.IsExpectedSuccess(err) {
				return
			}
			if err != nil {
				ctx.Errorf("Iteration error: %s", err)
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
				ctx.Run(testName, func(ctx *kt.Context) {
					ctx.CheckError(update.Error)
				})
			}
		})
		ctx.Run("NonJSON", func(ctx *kt.Context) {
			ctx.Parallel()
			id1 := ctx.TestDBName()
			id2 := ctx.TestDBName()
			docs := []interface{}{
				struct {
					ID   string `json:"_id"`
					Name string `json:"name"`
				}{ID: id1, Name: "Robert"},
				struct {
					ID   string `json:"_id"`
					Name string `json:"name"`
					Age  int    `json:"the_age"`
				}{ID: id2, Name: "Alice", Age: 32},
			}
			var updates []kivik.BulkResult
			err := kt.Retry(func() error {
				var err error
				updates, err = db.BulkDocs(context.Background(), docs)
				return err
			})
			if !ctx.IsExpectedSuccess(err) {
				return
			}
			if err != nil {
				ctx.Errorf("Iteration error: %s", err)
			}
			for _, update := range updates {
				if e := update.Error; e != nil {
					ctx.Errorf("Bulk create failed: %s", e)
				}
			}
			ctx.Run("Retrieve", func(ctx *kt.Context) {
				var result map[string]interface{}
				if err = db.Get(context.Background(), id2).ScanDoc(&result); err != nil {
					ctx.Fatalf("failed to scan bulk-inserted document: %s", err)
				}
				expected := map[string]interface{}{
					"_id":     id2,
					"name":    "Alice",
					"the_age": 32,
					"_rev":    result["_rev"],
				}
				if d := testy.DiffAsJSON(expected, result); d != nil {
					ctx.Errorf("Retrieved document differs:\n%s\n", d)
				}
			})
		})
	})
}
