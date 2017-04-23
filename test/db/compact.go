package db

import (
	"context"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	kt.Register("Compact", compact)
}

func compact(ctx *kt.Context) {
	ctx.RunRW(func(ctx *kt.Context) {
		ctx.RunAdmin(func(ctx *kt.Context) {
			ctx.Parallel()
			testCompact(ctx, ctx.Admin)
		})
		ctx.RunNoAuth(func(ctx *kt.Context) {
			ctx.Parallel()
			testCompact(ctx, ctx.NoAuth)
		})
	})
}

func testCompact(ctx *kt.Context, client *kivik.Client) {
	dbname := ctx.TestDB()
	defer ctx.Admin.DestroyDB(context.Background(), dbname, ctx.Options("db"))
	db, err := client.DB(context.Background(), dbname, ctx.Options("db"))
	if err != nil {
		ctx.Fatalf("Failed to connect to db: %s", err)
	}
	ctx.CheckError(db.Compact(context.Background()))
}
