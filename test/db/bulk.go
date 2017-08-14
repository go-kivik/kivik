package db

import (
	"context"

	"github.com/flimzy/diff"
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
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

func testBulkDocs(ctx *kt.Context, client *kivik.Client) {
	ctx.Parallel()
	dbname := ctx.TestDB()
	defer ctx.Admin.DestroyDB(context.Background(), dbname, ctx.Options("db"))
	adb, err := ctx.Admin.DB(context.Background(), dbname, ctx.Options("db"))
	if err != nil {
		ctx.Fatalf("Failed to connect to db as admin: %s", err)
	}
	db, err := client.DB(context.Background(), dbname, ctx.Options("db"))
	if err != nil {
		ctx.Fatalf("Failed to connect to db: %s", err)
	}
	ctx.Run("group", func(ctx *kt.Context) {
		ctx.Run("Create", func(ctx *kt.Context) {
			ctx.Parallel()
			doc := map[string]string{
				"name": "Robert",
			}
			updates, err := db.BulkDocs(context.Background(), []interface{}{doc})
			if !ctx.IsExpectedSuccess(err) {
				return
			}
			for updates.Next() {
				if err := updates.UpdateErr(); err != nil {
					ctx.Errorf("Bulk create failed: %s", err)
				}
			}
			if err := updates.Err(); err != nil {
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
			updates, err := db.BulkDocs(context.Background(), []interface{}{doc})
			if !ctx.IsExpectedSuccess(err) {
				return
			}
			for updates.Next() {
				if err := updates.UpdateErr(); err != nil {
					ctx.Errorf("Bulk update failed: %s", err)
				}
			}
			if err := updates.Err(); err != nil {
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
			updates, err := db.BulkDocs(context.Background(), []interface{}{doc})
			if !ctx.IsExpectedSuccess(err) {
				return
			}
			for updates.Next() {
				if err := updates.UpdateErr(); err != nil {
					ctx.Errorf("Bulk delete failed: %s", err)
				}
			}
			if err := updates.Err(); err != nil {
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

			updates, err := db.BulkDocs(context.Background(), []interface{}{doc0, doc1, doc2, doc3})
			if !ctx.IsExpectedSuccess(err) {
				return
			}
			for updates.Next() {
				var testName string
				switch updates.ID() {
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
					ctx.CheckError(updates.UpdateErr())
				})
			}
			if err := updates.Err(); err != nil {
				ctx.Errorf("Iteration error: %s", err)
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
			updates, err := db.BulkDocs(context.Background(), docs)
			if !ctx.IsExpectedSuccess(err) {
				return
			}
			for updates.Next() {
				if err := updates.UpdateErr(); err != nil {
					ctx.Errorf("Bulk create failed: %s", err)
				}
			}
			if err := updates.Err(); err != nil {
				ctx.Errorf("Iteration error: %s", err)
			}
			ctx.Run("Retrieve", func(ctx *kt.Context) {
				row, err := db.Get(context.Background(), id2)
				if err != nil {
					ctx.Fatalf("failed to retrieve bulk-inserted document: %s", err)
				}
				var result map[string]interface{}
				if err = row.ScanDoc(&result); err != nil {
					ctx.Fatalf("failed to scan bulk-inserted document: %s", err)
				}
				expected := map[string]interface{}{
					"_id":     id2,
					"name":    "Alice",
					"the_age": 32,
					"_rev":    result["_rev"],
				}
				if d := diff.AsJSON(expected, result); d != nil {
					ctx.Errorf("Retrieved document differs:\n%s\n", d)
				}
			})
		})
	})
}
