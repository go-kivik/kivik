package db

import (
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	kt.Register("CreateIndex", createIndex)
}

func createIndex(ctx *kt.Context) {
	ctx.RunRW(func(ctx *kt.Context) {
		ctx.RunAdmin(func(ctx *kt.Context) {
			ctx.Parallel()
			testCreateIndex(ctx, ctx.Admin)
		})
		ctx.RunNoAuth(func(ctx *kt.Context) {
			ctx.Parallel()
			testCreateIndex(ctx, ctx.NoAuth)
		})
	})
}

func testCreateIndex(ctx *kt.Context, client *kivik.Client) {
	dbname := ctx.TestDBName()
	defer ctx.Admin.DestroyDB(dbname)
	if err := ctx.Admin.CreateDB(dbname); err != nil {
		ctx.Fatalf("Failed to create db: %s", err)
	}
	db, err := client.DB(dbname)
	if err != nil {
		ctx.Fatalf("Failed to open db: %s", err)
	}
	ctx.Run("group", func(ctx *kt.Context) {
		ctx.Run("Valid", func(ctx *kt.Context) {
			ctx.Parallel()
			ctx.CheckError(db.CreateIndex("", "", `{"fields":["foo"]}`))
		})
		ctx.Run("NilIndex", func(ctx *kt.Context) {
			ctx.Parallel()
			ctx.CheckError(db.CreateIndex("", "", nil))
		})
		ctx.Run("BlankIndex", func(ctx *kt.Context) {
			ctx.Parallel()
			ctx.CheckError(db.CreateIndex("", "", ""))
		})
		ctx.Run("EmptyIndex", func(ctx *kt.Context) {
			ctx.Parallel()
			ctx.CheckError(db.CreateIndex("", "", "{}"))
		})
		ctx.Run("InvalidIndex", func(ctx *kt.Context) {
			ctx.Parallel()
			ctx.CheckError(db.CreateIndex("", "", `{"oink":true}`))
		})
		ctx.Run("InvalidJSON", func(ctx *kt.Context) {
			ctx.Parallel()
			ctx.CheckError(db.CreateIndex("", "", `chicken`))
		})
	})
}
