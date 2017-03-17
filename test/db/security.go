package db

import (
	"github.com/flimzy/diff"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	kt.Register("Security", security)
	kt.Register("SetSecurity", setSecurity)
}

var sec = &kivik.Security{
	Admins: kivik.Members{
		Names: []string{"bob", "alice"},
		Roles: []string{"hipsters"},
	},
	Members: kivik.Members{
		Names: []string{"fred"},
		Roles: []string{"beatniks"},
	},
}

func security(ctx *kt.Context) {
	ctx.RunAdmin(func(ctx *kt.Context) {
		for _, dbname := range ctx.MustStringSlice("databases") {
			func(dbname string) {
				ctx.Run(dbname, func(ctx *kt.Context) {
					ctx.Parallel()
					testGetSecurity(ctx, ctx.Admin, dbname, nil)
				})
			}(dbname)
		}
	})
	ctx.RunNoAuth(func(ctx *kt.Context) {
		for _, dbname := range ctx.MustStringSlice("databases") {
			func(dbname string) {
				ctx.Run(dbname, func(ctx *kt.Context) {
					ctx.Parallel()
					testGetSecurity(ctx, ctx.NoAuth, dbname, nil)
				})
			}(dbname)
		}
	})
	ctx.RunRW(func(ctx *kt.Context) {
		dbname := ctx.TestDBName()
		defer ctx.Admin.DestroyDB(dbname)
		if err := ctx.Admin.CreateDB(dbname); err != nil {
			ctx.Fatalf("Failed to create db: %s", err)
		}
		db, err := ctx.Admin.DB(dbname)
		if err != nil {
			ctx.Fatalf("Failed to open db: %s", err)
		}
		if err = db.SetSecurity(sec); err != nil {
			ctx.Fatalf("Failed to set security: %s", err)
		}
		ctx.Run("group", func(ctx *kt.Context) {
			ctx.RunAdmin(func(ctx *kt.Context) {
				ctx.Parallel()
				testGetSecurity(ctx, ctx.Admin, dbname, sec)
			})
			ctx.RunNoAuth(func(ctx *kt.Context) {
				ctx.Parallel()
				testGetSecurity(ctx, ctx.NoAuth, dbname, sec)
			})
		})
	})
}

func setSecurity(ctx *kt.Context) {
	ctx.RunRW(func(ctx *kt.Context) {
		ctx.RunAdmin(func(ctx *kt.Context) {
			testSetSecurityTests(ctx, ctx.Admin)
		})
		ctx.RunNoAuth(func(ctx *kt.Context) {
			testSetSecurityTests(ctx, ctx.NoAuth)
		})
	})
}

func testSetSecurityTests(ctx *kt.Context, client *kivik.Client) {
	ctx.Run("Exists", func(ctx *kt.Context) {
		ctx.Parallel()
		dbname := ctx.TestDBName()
		if err := ctx.Admin.CreateDB(dbname); err != nil {
			ctx.Fatalf("Failed to create db: %s", err)
		}
		defer ctx.Admin.DestroyDB(dbname)
		testSetSecurity(ctx, client, dbname)
	})
	ctx.Run("NotExists", func(ctx *kt.Context) {
		ctx.Parallel()
		dbname := ctx.TestDBName()
		testSetSecurity(ctx, client, dbname)
	})
}

func testSetSecurity(ctx *kt.Context, client *kivik.Client, dbname string) {
	db, err := client.DB(dbname)
	if err != nil {
		ctx.Fatalf("Failed to open db: %s", err)
	}
	ctx.CheckError(db.SetSecurity(sec))
}

func testGetSecurity(ctx *kt.Context, client *kivik.Client, dbname string, expected *kivik.Security) {
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
}
