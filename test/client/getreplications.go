package client

import (
	"github.com/flimzy/diff"
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
	"golang.org/x/net/context"
)

func init() {
	kt.Register("GetReplications", getReplications)
}

func getReplications(ctx *kt.Context) {
	ctx.RunAdmin(func(ctx *kt.Context) {
		ctx.Parallel()
		testGetReplications(ctx, ctx.Admin, []struct{}{})
	})
	ctx.RunNoAuth(func(ctx *kt.Context) {
		ctx.Parallel()
		testGetReplications(ctx, ctx.NoAuth, []struct{}{})
	})
	ctx.RunRW(func(ctx *kt.Context) {
		ctx.RunAdmin(func(ctx *kt.Context) {
			ctx.Parallel()

		})
		ctx.RunNoAuth(func(ctx *kt.Context) {
			ctx.Parallel()

		})
	})
}

func testRWGetReplications(ctx *kt.Context, client *kivik.Client) {
	dbname1 := ctx.TestDB()
	dbname2 := ctx.TestDB()
	defer ctx.Admin.DestroyDB(context.Background(), dbname1, ctx.Options("db"))
	defer ctx.Admin.DestroyDB(context.Background(), dbname2, ctx.Options("db"))
	ctx.Run("group", func(ctx *kt.Context) {
		ctx.Run("ValidReplication", func(ctx *kt.Context) {
			// TODO
		})
		ctx.Run("MissingSource", func(ctx *kt.Context) {
			// TODO
		})
		ctx.Run("MissingTarget", func(ctx *kt.Context) {
			// TODO
		})
	})
}

func testGetReplications(ctx *kt.Context, client *kivik.Client, expected interface{}) {
	reps, err := client.GetReplications(kt.CTX)
	if !ctx.IsExpectedSuccess(err) {
		return
	}
	if d := diff.AsJSON(expected, reps); d != "" {
		ctx.Errorf("GetReplications results differ:\n%s\n", d)
	}
}
