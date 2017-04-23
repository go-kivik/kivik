package client

import (
	"context"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	kt.Register("DestroyDB", destroyDB)
}

func destroyDB(ctx *kt.Context) {
	// All DestroyDB tests are RW by nature.
	ctx.RunRW(func(ctx *kt.Context) {
		ctx.RunAdmin(func(ctx *kt.Context) {
			ctx.Parallel()
			testDestroy(ctx, ctx.Admin)
		})
		ctx.RunNoAuth(func(ctx *kt.Context) {
			ctx.Parallel()
			testDestroy(ctx, ctx.NoAuth)
		})
	})
}

func testDestroy(ctx *kt.Context, client *kivik.Client) {
	ctx.Run("ExistingDB", func(ctx *kt.Context) {
		ctx.Parallel()
		dbName := ctx.TestDB()
		defer ctx.Admin.DestroyDB(context.Background(), dbName, ctx.Options("db"))
		ctx.CheckError(client.DestroyDB(context.Background(), dbName, ctx.Options("db")))
	})
	ctx.Run("NonExistantDB", func(ctx *kt.Context) {
		ctx.Parallel()
		ctx.CheckError(client.DestroyDB(context.Background(), ctx.TestDBName(), ctx.Options("db")))
	})
}
