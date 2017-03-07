package client

import (
	"github.com/flimzy/diff"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	kt.Register("Membership", membership)
}

func membership(ctx *kt.Context) {
	ctx.RunAdmin(func(ctx *kt.Context) {
		testMembership(ctx, ctx.Admin)
	})
	ctx.RunNoAuth(func(ctx *kt.Context) {
		testMembership(ctx, ctx.NoAuth)
	})
}

func testMembership(ctx *kt.Context, client *kivik.Client) {
	ctx.Parallel()
	all, cluster, err := ctx.Admin.Membership()
	if !ctx.IsExpectedSuccess(err) {
		return
	}
	if ctx.IsSet("all_min_count") {
		min := ctx.Int("all_min_count")
		if len(all) < min {
			ctx.Errorf("Expected at least %d 'all' entries, got %d", min, len(all))
		}
	} else if d := diff.TextSlices(ctx.MustStringSlice("all"), all); d != "" {
		ctx.Errorf("Unexpected 'all' list:\n%s\n", d)
	}
	if ctx.IsSet("cluster_min_count") {
		min := ctx.Int("cluster_min_count")
		if len(cluster) < min {
			ctx.Errorf("Expected at least %d 'cluster' entries, got %d", min, len(cluster))
		}
	} else if d := diff.TextSlices(ctx.MustStringSlice("cluster"), cluster); d != "" {
		ctx.Errorf("Unexpected 'cluster' list:\n%s\n", d)
	}
}
