package db

import (
	"context"

	"github.com/flimzy/diff"
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	kt.Register("Get", get)
}

type testDoc struct {
	ID   string `json:"_id"`
	Rev  string `json:"_rev,omitempty"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func get(ctx *kt.Context) {
	ctx.RunRW(func(ctx *kt.Context) {
		dbName := ctx.TestDB()
		defer ctx.Admin.DestroyDB(context.Background(), dbName, ctx.Options("db"))
		db, err := ctx.Admin.DB(context.Background(), dbName, ctx.Options("db"))
		if err != nil {
			ctx.Fatalf("Failed to connect to test db: %s", err)
		}

		doc := &testDoc{
			ID:   "bob",
			Name: "Robert",
			Age:  32,
		}
		rev, err := db.Put(context.Background(), doc.ID, doc)
		if err != nil {
			ctx.Fatalf("Failed to create doc in test db: %s", err)
		}
		doc.Rev = rev

		ddoc := &testDoc{
			ID:   "_design/foo",
			Name: "Designer",
		}
		rev, err = db.Put(context.Background(), ddoc.ID, ddoc)
		if err != nil {
			ctx.Fatalf("Failed to create design doc in test db: %s", err)
		}
		ddoc.Rev = rev

		local := &testDoc{
			ID:   "_local/foo",
			Name: "Designer",
		}
		rev, err = db.Put(context.Background(), local.ID, local)
		if err != nil {
			ctx.Fatalf("Failed to create local doc in test db: %s", err)
		}
		local.Rev = rev

		ctx.Run("group", func(ctx *kt.Context) {
			ctx.RunAdmin(func(ctx *kt.Context) {
				ctx.Parallel()
				db, err := ctx.Admin.DB(context.Background(), dbName, ctx.Options("db"))
				if !ctx.IsExpectedSuccess(err) {
					return
				}
				testGet(ctx, db, doc)
				testGet(ctx, db, ddoc)
				testGet(ctx, db, local)
				testGet(ctx, db, &testDoc{ID: "bogus"})
			})
			ctx.RunNoAuth(func(ctx *kt.Context) {
				ctx.Parallel()
				db, err := ctx.NoAuth.DB(context.Background(), dbName, ctx.Options("db"))
				if !ctx.IsExpectedSuccess(err) {
					return
				}
				testGet(ctx, db, doc)
				testGet(ctx, db, ddoc)
				testGet(ctx, db, local)
				testGet(ctx, db, &testDoc{ID: "bogus"})
			})
		})
	})
}

func testGet(ctx *kt.Context, db *kivik.DB, expectedDoc *testDoc) {
	ctx.Run(expectedDoc.ID, func(ctx *kt.Context) {
		ctx.Parallel()
		row, err := db.Get(context.Background(), expectedDoc.ID)
		if !ctx.IsExpectedSuccess(err) {
			return
		}
		doc := &testDoc{}
		if err = row.ScanDoc(&doc); err != nil {
			ctx.Fatalf("Failed to scan doc: %s", err)
		}
		if d := diff.AsJSON(expectedDoc, doc); d != nil {
			ctx.Errorf("Fetched document not as expected:\n%s\n", d)
		}
	})
}
