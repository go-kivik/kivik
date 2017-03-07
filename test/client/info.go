package client

import (
	"regexp"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	kt.Register("ServerInfo", info)
}

func info(ctx *kt.Context) {
	ctx.RunAdmin(func(ctx *kt.Context) {
		testServerInfo(ctx, ctx.Admin)
	})
	ctx.RunNoAuth(func(ctx *kt.Context) {
		testServerInfo(ctx, ctx.NoAuth)
	})
}

func testServerInfo(ctx *kt.Context, client *kivik.Client) {
	ctx.Parallel()
	info, err := client.ServerInfo()
	if !ctx.IsExpectedSuccess(err) {
		return
	}
	version := regexp.MustCompile(ctx.MustString("version"))
	vendor := regexp.MustCompile(ctx.MustString("vendor"))
	vendorVersion := regexp.MustCompile(ctx.MustString("vendor_version"))
	if !version.MatchString(info.Version()) {
		ctx.Errorf("Version '%s' does not match /%s/", info.Version(), version)
	}
	if !vendor.MatchString(info.Vendor()) {
		ctx.Errorf("Vendor '%s' doesnot match /%s/", info.Vendor(), vendor)
	}
	if !vendorVersion.MatchString(info.VendorVersion()) {
		ctx.Errorf("Vendor version '%s' does not match /%s", info.VendorVersion(), vendorVersion)
	}
}
