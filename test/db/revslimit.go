package db

import (
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	kt.Register("RevsLimit", revsLimit)
}

func revsLimit(ctx *kt.Context) {
	ctx.RunAdmin(func(ctx *kt.Context) {
		for _, dbname := range ctx.MustStringSlice("databases") {
			func(dbname string) {
				ctx.Run(dbname, func(ctx *kt.Context) {
					ctx.Parallel()
					testRevsLimit(ctx, ctx.Admin, dbname, ctx.Int("revs_limit"))
				})
			}(dbname)
		}
	})
	ctx.RunNoAuth(func(ctx *kt.Context) {
		for _, dbname := range ctx.MustStringSlice("databases") {
			func(dbname string) {
				ctx.Run(dbname, func(ctx *kt.Context) {
					ctx.Parallel()
					testRevsLimit(ctx, ctx.NoAuth, dbname, ctx.Int("revs_limit"))
				})
			}(dbname)
		}
	})
	ctx.RunRW(func(ctx *kt.Context) {
		ctx.RunAdmin(func(ctx *kt.Context) {
			testRevsLimitRW(ctx, ctx.Admin)
		})
		ctx.RunNoAuth(func(ctx *kt.Context) {
			testRevsLimitRW(ctx, ctx.NoAuth)
		})
	})
}

func testRevsLimitRW(ctx *kt.Context, client *kivik.Client) {
	dbname := ctx.TestDBName()
	defer ctx.Admin.DestroyDB(dbname)
	if err := ctx.Admin.CreateDB(dbname); err != nil {
		ctx.Fatalf("Failed to create db: %s", err)
	}
	ctx.Run("Set", func(ctx *kt.Context) {
		testSetRevsLimit(ctx, client, dbname)
	})
}

const testLimit = 555

func testSetRevsLimit(ctx *kt.Context, client *kivik.Client, dbname string) {
	db, err := client.DB(dbname)
	if err != nil {
		ctx.Fatalf("Failed to open db: %s", err)
	}
	err = db.SetRevsLimit(testLimit)
	if !ctx.IsExpectedSuccess(err) {
		return
	}
	newLimit, err := db.RevsLimit()
	if err != nil {
		ctx.Fatalf("Failed to re-test limit: %s", err)
	}
	if newLimit != testLimit {
		ctx.Errorf("Limit set to %d, but got %d", testLimit, newLimit)
	}
}

func testRevsLimit(ctx *kt.Context, client *kivik.Client, dbname string, expected int) {
	db, err := client.DB(dbname)
	if err != nil {
		ctx.Fatalf("Failed to open db: %s", err)
	}
	limit, err := db.RevsLimit()
	if !ctx.IsExpectedSuccess(err) {
		return
	}
	if limit != expected {
		ctx.Errorf("Unexpected limit: Expected: %d, Actual: %d", expected, limit)
	}
}
