package db

import (
	"context"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	kt.Register("Flush", flush)
}

func flush(ctx *kt.Context) {
	ctx.RunAdmin(func(ctx *kt.Context) {
		flushTest(ctx, ctx.Admin)
	})
	ctx.RunNoAuth(func(ctx *kt.Context) {
		flushTest(ctx, ctx.NoAuth)
	})
}

func flushTest(ctx *kt.Context, client *kivik.Client) {
	ctx.Parallel()
	for _, dbName := range ctx.MustStringSlice("databases") {
		ctx.Run(dbName, func(ctx *kt.Context) {
			db, err := client.DB(context.Background(), dbName, ctx.Options("db"))
			if !ctx.IsExpectedSuccess(err) {
				return
			}
			ctx.Run("DoFlush", func(ctx *kt.Context) {
				err := db.Flush(context.Background())
				if !ctx.IsExpectedSuccess(err) {
					return
				}
			})
		})
	}
}
