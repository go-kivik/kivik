package client

import (
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
		dbName := ctx.TestDBName()
		defer ctx.Admin.DestroyDB(dbName)
		if err := ctx.Admin.CreateDB(dbName); err != nil {
			ctx.Fatalf("Failed to create db: %s", err)
		}
		ctx.CheckError(client.DestroyDB(dbName))
	})
	ctx.Run("NonExistantDB", func(ctx *kt.Context) {
		ctx.Parallel()
		ctx.CheckError(client.DestroyDB(ctx.TestDBName()))
	})
}
