package db

import (
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	kt.Register("Put", put)
}

func put(ctx *kt.Context) {
	ctx.RunRW(func(ctx *kt.Context) {
		ctx.RunAdmin(func(ctx *kt.Context) {
			testPut(ctx, ctx.Admin)
		})
		ctx.RunNoAuth(func(ctx *kt.Context) {
			testPut(ctx, ctx.NoAuth)
		})
	})
}

func testPut(ctx *kt.Context, client *kivik.Client) {
	ctx.Parallel()
	dbName := ctx.TestDBName()
	defer ctx.Admin.DestroyDB(dbName)
	_ = ctx.Admin.CreateDB(dbName)
	db, err := client.DB(dbName)
	if !ctx.IsExpectedSuccess(err) {
		return
	}

	doc := &testDoc{
		ID:   ctx.TestDBName(),
		Name: "Alberto",
		Age:  32,
	}
	ctx.Run("Create", func(ctx *kt.Context) {
		rev, err := db.Put(doc.ID, doc)
		if !ctx.IsExpectedSuccess(err) {
			return
		}
		doc.Rev = rev
		doc.Age = 40
		ctx.Run("Update", func(ctx *kt.Context) {
			_, err := db.Put(doc.ID, doc)
			ctx.CheckError(err)
		})
	})
}
