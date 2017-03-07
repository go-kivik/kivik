package client

import (
	"fmt"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	kt.Register("UUIDs", uuids)
}

func uuids(ctx *kt.Context) {
	ctx.RunAdmin(func(ctx *kt.Context) {
		for _, c := range ctx.MustIntSlice("counts") {
			testUUIDs(ctx, ctx.Admin, c)
		}
	})
	ctx.RunNoAuth(func(ctx *kt.Context) {
		for _, c := range ctx.MustIntSlice("counts") {
			testUUIDs(ctx, ctx.NoAuth, c)
		}
	})
}

func testUUIDs(ctx *kt.Context, client *kivik.Client, count int) {
	ctx.Run(fmt.Sprintf("%dCount", count), func(ctx *kt.Context) {
		ctx.Parallel()
		uuids, err := client.UUIDs(count)
		if !ctx.IsExpectedSuccess(err) {
			return
		}
		if len(uuids) != count {
			ctx.Errorf("Requested %d UUIDs, but got %d", count, len(uuids))
		}
	})
}
