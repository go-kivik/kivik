package db

import (
	"github.com/flimzy/diff"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	kt.Register("Security", security)
}

func security(ctx *kt.Context) {
	ctx.RunAdmin(func(ctx *kt.Context) {
		for _, dbname := range ctx.MustStringSlice("databases") {
			testGetSecurity(ctx, ctx.Admin, dbname, nil)
		}
	})
	ctx.RunNoAuth(func(ctx *kt.Context) {
		for _, dbname := range ctx.MustStringSlice("databases") {
			testGetSecurity(ctx, ctx.NoAuth, dbname, nil)
		}
	})
}

func testGetSecurity(ctx *kt.Context, client *kivik.Client, dbname string, expected *kivik.Security) {
	ctx.Run(dbname, func(ctx *kt.Context) {
		ctx.Parallel()
		db, err := client.DB(dbname)
		if err != nil {
			ctx.Fatalf("Failed to open db: %s", err)
		}
		sec, err := db.Security()
		if !ctx.IsExpectedSuccess(err) {
			return
		}
		if expected != nil {
			if d := diff.AsJSON(expected, sec); d != "" {
				ctx.Errorf("Security document differs from expected:\n%s\n", d)
			}
		}
	})
}
