package db

import (
	"fmt"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	kt.Register("DBInfo", dbInfo)
}

func dbInfo(ctx *kt.Context) {
	ctx.RunAdmin(func(ctx *kt.Context) {
		ctx.Parallel()
		roTests(ctx, ctx.Admin)
	})
	ctx.RunNoAuth(func(ctx *kt.Context) {
		ctx.Parallel()
		roTests(ctx, ctx.NoAuth)
	})
	ctx.RunRW(func(ctx *kt.Context) {
		ctx.Parallel()
		ctx.RunAdmin(func(ctx *kt.Context) {
			ctx.Parallel()
			rwTests(ctx, ctx.Admin)
		})
		ctx.RunNoAuth(func(ctx *kt.Context) {
			ctx.Parallel()
			rwTests(ctx, ctx.NoAuth)
		})
	})
}

func rwTests(ctx *kt.Context, client *kivik.Client) {
	dbname := ctx.TestDBName()
	defer ctx.Admin.DestroyDB(dbname)
	if err := ctx.Admin.CreateDB(dbname); err != nil {
		ctx.Fatalf("Failed to create test db: %s", err)
	}
	db, err := ctx.Admin.DB(dbname)
	if err != nil {
		ctx.Fatalf("Failed to connect to db: %s", err)
	}
	for i := 0; i < 10; i++ {
		id := fmt.Sprintf("%d", i)
		rev, err := db.Put(id, struct{}{})
		if err != nil {
			ctx.Fatalf("Failed to create document ID %s: %s", id, err)
		}
		if i > 5 {
			if _, err = db.Delete(id, rev); err != nil {
				ctx.Fatalf("Failed to delete document ID %s: %s", id, err)
			}
		}
	}
	testDBInfo(ctx, client, dbname, 6)
}

func roTests(ctx *kt.Context, client *kivik.Client) {
	for _, dbname := range ctx.MustStringSlice("databases") {
		func(dbname string) {
			ctx.Run(dbname, func(ctx *kt.Context) {
				ctx.Parallel()
				testDBInfo(ctx, client, dbname, 0)
			})
		}(dbname)
	}
}

func testDBInfo(ctx *kt.Context, client *kivik.Client, dbname string, docCount int64) {
	db, err := client.DB(dbname)
	// Check against the same status for connecting, and db.Info() later, because
	// where the error might occur is backend-specific.
	var info *kivik.DBInfo
	if err == nil {
		info, err = db.Info()
	}
	if !ctx.IsExpectedSuccess(err) {
		return
	}
	if info.Name != dbname {
		ctx.Errorf("Name: Expected '%s', actual '%s'", dbname, info.Name)
	}
	if docCount > 0 {
		if docCount != info.DocCount {
			ctx.Errorf("DocCount: Expected %d, actual %d", docCount, info.DocCount)
		}
	}
}
