package client

import (
	"context"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	kt.Register("CreateDB", createDB)
}

func createDB(ctx *kt.Context) {
	ctx.RunRW(func(ctx *kt.Context) {
		ctx.RunAdmin(func(ctx *kt.Context) {
			testCreateDB(ctx, ctx.Admin)
		})
		ctx.RunNoAuth(func(ctx *kt.Context) {
			testCreateDB(ctx, ctx.NoAuth)
		})
	})
}

func testCreateDB(ctx *kt.Context, client *kivik.Client) {
	ctx.Parallel()
	dbName := ctx.TestDBName()
	defer ctx.Admin.DestroyDB(context.Background(), dbName, ctx.Options("db"))
	if !ctx.IsExpectedSuccess(client.CreateDB(context.Background(), dbName, ctx.Options("db"))) {
		return
	}
	ctx.Run("Recreate", func(ctx *kt.Context) {
		ctx.CheckError(client.CreateDB(context.Background(), dbName, ctx.Options("db")))
	})
}
