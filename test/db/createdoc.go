package db

import (
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	kt.Register("CreateDoc", createDoc)
}

func createDoc(ctx *kt.Context) {
	ctx.RunRW(func(ctx *kt.Context) {
		dbname := ctx.TestDBName()
		if err := ctx.Admin.CreateDB(dbname); err != nil {
			ctx.Fatalf("Failed to create test db: %s", err)
		}
		defer ctx.Admin.DestroyDB(dbname)
		ctx.Run("group", func(ctx *kt.Context) {
			ctx.RunAdmin(func(ctx *kt.Context) {
				ctx.Parallel()
				testCreate(ctx, ctx.Admin, dbname)
			})
			ctx.RunNoAuth(func(ctx *kt.Context) {
				ctx.Parallel()
				testCreate(ctx, ctx.NoAuth, dbname)
			})
		})
	})
}

func testCreate(ctx *kt.Context, client *kivik.Client, dbname string) {
	db, err := client.DB(dbname)
	if err != nil {
		ctx.Fatalf("Failed to connect to database: %s", err)
	}
	ctx.Run("WithoutID", func(ctx *kt.Context) {
		ctx.Parallel()
		_, _, err := db.CreateDoc(map[string]string{"foo": "bar"})
		ctx.CheckError(err)
	})
	ctx.Run("WithID", func(ctx *kt.Context) {
		ctx.Parallel()
		id := ctx.TestDBName()
		docID, _, err := db.CreateDoc(map[string]string{"foo": "bar", "_id": id})
		if !ctx.IsExpectedSuccess(err) {
			return
		}
		if id != docID {
			ctx.Errorf("CreateDoc didn't honor provided ID. Expected '%s', Got '%s'", id, docID)
		}
	})
}
