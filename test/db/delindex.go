package db

import (
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	kt.Register("DeleteIndex", delindex)
}

func delindex(ctx *kt.Context) {
	ctx.RunRW(func(ctx *kt.Context) {
		ctx.RunAdmin(func(ctx *kt.Context) {
			ctx.Parallel()
			testDelIndex(ctx, ctx.Admin)
		})
		ctx.RunNoAuth(func(ctx *kt.Context) {
			ctx.Parallel()
			testDelIndex(ctx, ctx.NoAuth)
		})
	})
}

func testDelIndex(ctx *kt.Context, client *kivik.Client) {
	dbname := ctx.TestDBName()
	defer ctx.Admin.DestroyDB(dbname)
	if err := ctx.Admin.CreateDB(dbname); err != nil {
		ctx.Fatalf("Failed to create db: %s", err)
	}
	dba, err := ctx.Admin.DB(dbname)
	if err != nil {
		ctx.Fatalf("Failed to open db as admin: %s", err)
	}
	if err = dba.CreateIndex("foo", "bar", `{"fields":["foo"]}`); err != nil {
		ctx.Fatalf("Failed to create index: %s", err)
	}
	db, err := client.DB(dbname)
	if err != nil {
		ctx.Fatalf("Failed to open db: %s", err)
	}
	ctx.Run("group", func(ctx *kt.Context) {
		ctx.Run("ValidIndex", func(ctx *kt.Context) {
			ctx.Parallel()
			ctx.CheckError(db.DeleteIndex("foo", "bar"))
		})
		ctx.Run("NotFoundDdoc", func(ctx *kt.Context) {
			ctx.Parallel()
			ctx.CheckError(db.DeleteIndex("notFound", "bar"))
		})
		ctx.Run("NotFoundName", func(ctx *kt.Context) {
			ctx.Parallel()
			ctx.CheckError(db.DeleteIndex("foo", "notFound"))
		})
	})
}
