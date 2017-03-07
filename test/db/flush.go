package db

import (
	"time"

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
			db, err := client.DB(dbName)
			if !ctx.IsExpectedSuccess(err) {
				return
			}
			ctx.Run("DoFlush", func(ctx *kt.Context) {
				ts, err := db.Flush()
				if !ctx.IsExpectedSuccess(err) {
					return
				}
				ctx.Run("Timestamp", func(ctx *kt.Context) {
					if !time.Now().After(ts) {
						ctx.Errorf("Timestamp in the future: %s", ts)
					}
					if !ts.After(time.Now().Add(-time.Hour * 9000)) { // About a year; just to make sure we're within an order of magnitude
						ctx.Errorf("Timestamp is in the distant past: %s", ts)
					}
				})
			})
		})
	}
}
