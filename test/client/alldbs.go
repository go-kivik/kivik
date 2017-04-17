package client

import (
	"context"

	"github.com/flimzy/diff"
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	kt.Register("AllDBs", allDBs)
}

func allDBs(ctx *kt.Context) {
	ctx.RunAdmin(func(ctx *kt.Context) {
		testAllDBs(ctx, ctx.Admin, ctx.StringSlice("expected"))
	})
	ctx.RunNoAuth(func(ctx *kt.Context) {
		testAllDBs(ctx, ctx.NoAuth, ctx.StringSlice("expected"))
	})
	if ctx.RW && ctx.Admin != nil {
		ctx.Run("RW", func(ctx *kt.Context) {
			testAllDBsRW(ctx)
		})
	}
}

func testAllDBsRW(ctx *kt.Context) {
	admin := ctx.Admin
	dbName := ctx.TestDBName()
	defer admin.DestroyDB(context.Background(), dbName)
	if err := admin.CreateDB(context.Background(), dbName); err != nil {
		ctx.Errorf("Failed to create test DB '%s': %s", dbName, err)
		return
	}
	expected := append(ctx.StringSlice("expected"), dbName)
	ctx.Run("group", func(ctx *kt.Context) {
		ctx.RunAdmin(func(ctx *kt.Context) {
			testAllDBs(ctx, ctx.Admin, expected)
		})
		ctx.RunNoAuth(func(ctx *kt.Context) {
			testAllDBs(ctx, ctx.NoAuth, expected)
		})
	})
}

func testAllDBs(ctx *kt.Context, client *kivik.Client, expected []string) {
	ctx.Parallel()
	allDBs, err := client.AllDBs(context.Background())
	if !ctx.IsExpectedSuccess(err) {
		return
	}
	if d := diff.TextSlices(expected, allDBs); d != "" {
		ctx.Errorf("AllDBs() returned unexpected list:\n%s\n", d)
	}
}
