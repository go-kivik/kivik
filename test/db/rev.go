package db

import (
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	kt.Register("Rev", rev)
}

func rev(ctx *kt.Context) {
	ctx.RunRW(func(ctx *kt.Context) {
		dbName := ctx.TestDBName()
		defer ctx.Admin.DestroyDB(dbName)
		if err := ctx.Admin.CreateDB(dbName); err != nil {
			ctx.Fatalf("Failed to create test db: %s", err)
		}
		db, err := ctx.Admin.DB(dbName)
		if err != nil {
			ctx.Fatalf("Failed to connect to test db: %s", err)
		}
		doc := &testDoc{
			ID: "bob",
		}
		rev, err := db.Put(doc.ID, doc)
		if err != nil {
			ctx.Fatalf("Failed to create doc in test db: %s", err)
		}
		doc.Rev = rev
		ctx.Run("group", func(ctx *kt.Context) {
			ctx.RunAdmin(func(ctx *kt.Context) {
				ctx.Parallel()
				db, err := ctx.Admin.DB(dbName)
				if !ctx.IsExpectedSuccess(err) {
					return
				}
				testRev(ctx, db, doc)
				testRev(ctx, db, &testDoc{ID: "bogus"})
			})
			ctx.RunNoAuth(func(ctx *kt.Context) {
				ctx.Parallel()
				db, err := ctx.NoAuth.DB(dbName)
				if !ctx.IsExpectedSuccess(err) {
					return
				}
				testRev(ctx, db, doc)
				testRev(ctx, db, &testDoc{ID: "bogus"})
			})
		})
	})
}

func testRev(ctx *kt.Context, db *kivik.DB, expectedDoc *testDoc) {
	ctx.Run(expectedDoc.ID, func(ctx *kt.Context) {
		ctx.Parallel()
		rev, err := db.Rev(expectedDoc.ID)
		if !ctx.IsExpectedSuccess(err) {
			return
		}
		doc := &testDoc{}
		if err = db.Get(expectedDoc.ID, &doc, nil); err != nil {
			ctx.Fatalf("Failed to get doc: %s", err)
		}
		if rev != doc.Rev {
			ctx.Errorf("Unexpected rev. Expected: %s, Actual: %s", doc.Rev, rev)
		}
	})
}
