package client

import (
	"context"
	"regexp"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	kt.Register("Version", version)
}

func version(ctx *kt.Context) {
	ctx.RunAdmin(func(ctx *kt.Context) {
		testServerInfo(ctx, ctx.Admin)
	})
	ctx.RunNoAuth(func(ctx *kt.Context) {
		testServerInfo(ctx, ctx.NoAuth)
	})
}

func testServerInfo(ctx *kt.Context, client *kivik.Client) {
	ctx.Parallel()
	info, err := client.Version(context.Background())
	if !ctx.IsExpectedSuccess(err) {
		return
	}
	version := regexp.MustCompile(ctx.MustString("version"))
	vendor := regexp.MustCompile(ctx.MustString("vendor"))
	if !version.MatchString(info.Version) {
		ctx.Errorf("Version '%s' does not match /%s/", info.Version, version)
	}
	if !vendor.MatchString(info.Vendor) {
		ctx.Errorf("Vendor '%s' doesnot match /%s/", info.Vendor, vendor)
	}
}
