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
	dbname := ctx.TestDBName()
	defer ctx.Admin.DestroyDB(context.Background(), dbname)
	if err := ctx.Admin.CreateDB(context.Background(), dbname); err != nil {
		ctx.Fatalf("Failed to create test db: %s", err)
	}
	db, err := client.DB(context.Background(), dbname)
	if err != nil {
		ctx.Fatalf("Failed to connect to db: %s", err)
	}
	ctx.CheckError(db.Compact(context.Background()))
}
