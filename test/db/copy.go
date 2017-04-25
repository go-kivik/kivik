package db

import (
	"context"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	kt.Register("Copy", copy)
}

func copy(ctx *kt.Context) {
	ctx.RunRW(func(ctx *kt.Context) {
		dbname := ctx.TestDB()
		defer ctx.Admin.DestroyDB(context.Background(), dbname, ctx.Options("db"))
		db, err := ctx.Admin.DB(context.Background(), dbname, ctx.Options("db"))
		if err != nil {
			ctx.Fatalf("Failed to open db: %s", err)
		}

		doc := map[string]string{
			"_id":  "foo",
			"name": "Robert",
		}
		rev, err := db.Put(context.Background(), doc["_id"], doc)
		if err != nil {
			ctx.Fatalf("Failed to create source doc: %s", err)
		}
		doc["_rev"] = rev

		ddoc := map[string]string{
			"_id":  "_design/foo",
			"name": "Robert",
		}
		rev, err = db.Put(context.Background(), ddoc["_id"], ddoc)
		if err != nil {
			ctx.Fatalf("Failed to create source design doc: %s", err)
		}
		ddoc["_rev"] = rev

		local := map[string]string{
			"_id":  "_local/foo",
			"name": "Robert",
		}
		rev, err = db.Put(context.Background(), local["_id"], local)
		if err != nil {
			ctx.Fatalf("Failed to create source design doc: %s", err)
		}
		local["_rev"] = rev

		ctx.Run("group", func(ctx *kt.Context) {
			ctx.RunAdmin(func(ctx *kt.Context) {
				copyTest(ctx, ctx.Admin, dbname, doc)
				copyTest(ctx, ctx.Admin, dbname, ddoc)
				copyTest(ctx, ctx.Admin, dbname, local)
			})
			ctx.RunNoAuth(func(ctx *kt.Context) {
				copyTest(ctx, ctx.NoAuth, dbname, doc)
				copyTest(ctx, ctx.NoAuth, dbname, ddoc)
				copyTest(ctx, ctx.NoAuth, dbname, local)
			})
		})
	})
}

func copyTest(ctx *kt.Context, client *kivik.Client, dbname string, source map[string]string) {
	ctx.Run(source["_id"], func(ctx *kt.Context) {
		ctx.Parallel()
		db, err := client.DB(context.Background(), dbname, ctx.Options("db"))
		if err != nil {
			ctx.Fatalf("Failed to open db: %s", err)
		}
		targetID := ctx.TestDBName()
		rev, err := db.Copy(context.Background(), targetID, source["_id"])
		if !ctx.IsExpectedSuccess(err) {
			return
		}
		ctx.Run("RevCopy", func(ctx *kt.Context) {
			copy := map[string]string{
				"_id":  targetID,
				"name": "Bob",
				"_rev": rev,
			}
			if _, err := db.Put(context.Background(), targetID, copy); err != nil {
				ctx.Fatalf("Failed to update copy: %s", err)
			}
			targetID2 := ctx.TestDBName()
			if _, err := db.Copy(context.Background(), targetID2, targetID, kivik.Options{"rev": rev}); err != nil {
				ctx.Fatalf("Failed to copy doc with rev option: %s", err)
			}
			row, err := db.Get(context.Background(), targetID2)
			if err != nil {
				ctx.Fatalf("Failed to read copy: %s", err)
			}
			var readCopy map[string]string
			if err = row.ScanDoc(&readCopy); err != nil {
				ctx.Fatalf("Failed to scan copy: %s", err)
			}
			if readCopy["name"] != "Robert" {
				ctx.Errorf("Copy-with-rev failed. Name = %s, expected %s", readCopy["name"], "Robert")
			}
		})
	})
}
