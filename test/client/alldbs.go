package client

import (
	"github.com/flimzy/diff"
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	kt.Register("AllDBs", allDBs)
}

func allDBs(ctx *kt.Context) {
	if ctx.RW && ctx.Admin != nil {
		ctx.Run("RW", func(ctx *kt.Context) {
			testAllDBsRW(ctx)
		})
	}
	ctx.RunAdmin(func(ctx *kt.Context) {
		// t.Parallel()
		testAllDBs(ctx, ctx.Admin, ctx.StringSlice("expected"))
	})
	ctx.RunNoAuth(func(ctx *kt.Context) {
		testAllDBs(ctx, ctx.NoAuth, ctx.StringSlice("expected"))
	})
}

func testAllDBsRW(ctx *kt.Context) {
	admin := ctx.Admin
	dbName := ctx.TestDBName()
	defer admin.DestroyDB(dbName)
	if err := admin.CreateDB(dbName); err != nil {
		ctx.Errorf("Failed to create test DB '%s': %s", dbName, err)
		return
	}
	expected := append(ctx.StringSlice("expected"), dbName)
	ctx.RunAdmin(func(ctx *kt.Context) {
		testAllDBs(ctx, ctx.Admin, expected)
	})
	ctx.RunNoAuth(func(ctx *kt.Context) {
		testAllDBs(ctx, ctx.NoAuth, expected)
	})
}

func testAllDBs(ctx *kt.Context, client *kivik.Client, expected []string) {
	allDBs, err := client.AllDBs()
	if !ctx.IsExpectedSuccess(err) {
		return
	}
	if d := diff.TextSlices(expected, allDBs); d != "" {
		ctx.Errorf("AllDBs() returned unexpected list:\n%s\n", d)
	}
}
