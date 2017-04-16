package client

import (
	"context"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	kt.Register("DBExists", dbExists)
}

func dbExists(ctx *kt.Context) {
	ctx.RunAdmin(func(ctx *kt.Context) {
		for _, dbName := range ctx.MustStringSlice("databases") {
			checkDBExists(ctx, ctx.Admin, dbName)
		}
	})
	ctx.RunNoAuth(func(ctx *kt.Context) {
		for _, dbName := range ctx.MustStringSlice("databases") {
			checkDBExists(ctx, ctx.NoAuth, dbName)
		}
	})
	ctx.RunRW(func(ctx *kt.Context) {
		dbName := ctx.TestDBName()
		defer ctx.Admin.DestroyDB(context.Background(), dbName)
		if err := ctx.Admin.CreateDB(context.Background(), dbName); err != nil {
			ctx.Errorf("Failed to create test DB: %s", err)
			return
		}
		ctx.Run("group", func(ctx *kt.Context) {
			ctx.RunAdmin(func(ctx *kt.Context) {
				checkDBExists(ctx, ctx.Admin, dbName)
			})
			ctx.RunNoAuth(func(ctx *kt.Context) {
				checkDBExists(ctx, ctx.NoAuth, dbName)
			})
		})
	})
}

func checkDBExists(ctx *kt.Context, client *kivik.Client, dbName string) {
	ctx.Run(dbName, func(ctx *kt.Context) {
		ctx.Parallel()
		exists, err := client.DBExists(context.Background(), dbName)
		if !ctx.IsExpectedSuccess(err) {
			return
		}
		if ctx.MustBool("exists") != exists {
			ctx.Errorf("Expected: %t, actual: %t", ctx.Bool("exists"), exists)
		}
	})
}
