package db

import (
	"context"

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
	dbname := ctx.TestDB()
	defer ctx.Admin.DestroyDB(context.Background(), dbname, ctx.Options("db"))
	db, err := client.DB(context.Background(), dbname, ctx.Options("db"))
	if err != nil {
		ctx.Fatalf("Failed to open db: %s", err)
	}
	ctx.Run("group", func(ctx *kt.Context) {
		ctx.Run("Valid", func(ctx *kt.Context) {
			ctx.Parallel()
			ctx.CheckError(db.CreateIndex(context.Background(), "", "", `{"fields":["foo"]}`))
		})
		ctx.Run("NilIndex", func(ctx *kt.Context) {
			ctx.Parallel()
			ctx.CheckError(db.CreateIndex(context.Background(), "", "", nil))
		})
		ctx.Run("BlankIndex", func(ctx *kt.Context) {
			ctx.Parallel()
			ctx.CheckError(db.CreateIndex(context.Background(), "", "", ""))
		})
		ctx.Run("EmptyIndex", func(ctx *kt.Context) {
			ctx.Parallel()
			ctx.CheckError(db.CreateIndex(context.Background(), "", "", "{}"))
		})
		ctx.Run("InvalidIndex", func(ctx *kt.Context) {
			ctx.Parallel()
			ctx.CheckError(db.CreateIndex(context.Background(), "", "", `{"oink":true}`))
		})
		ctx.Run("InvalidJSON", func(ctx *kt.Context) {
			ctx.Parallel()
			ctx.CheckError(db.CreateIndex(context.Background(), "", "", `chicken`))
		})
	})
}
