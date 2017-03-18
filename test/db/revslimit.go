package db

import (
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	kt.Register("RevsLimit", revsLimit)
}

func revsLimit(ctx *kt.Context) {
	ctx.RunAdmin(func(ctx *kt.Context) {
		for _, dbname := range ctx.MustStringSlice("databases") {
			ctx.Run(dbname, func(ctx *kt.Context) {
				testRevsLimit(ctx, ctx.Admin, dbname)
			})
		}
	})
	ctx.RunNoAuth(func(ctx *kt.Context) {
		for _, dbname := range ctx.MustStringSlice("databases") {
			ctx.Run(dbname, func(ctx *kt.Context) {
				testRevsLimit(ctx, ctx.NoAuth, dbname)
			})
		}
	})
}

func testRevsLimit(ctx *kt.Context, client *kivik.Client, dbname string) {
	ctx.Parallel()
	db, err := client.DB(dbname)
	if err != nil {
		ctx.Fatalf("Failed to open db: %s", err)
	}
	limit, err := db.RevsLimit()
	if !ctx.IsExpectedSuccess(err) {
		return
	}
	expected := ctx.MustInt("revs_limit")
	if limit != expected {
		ctx.Errorf("Unexpected limit: Expected: %d, Actual: %d", expected, limit)
	}
}
