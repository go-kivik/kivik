package db

import (
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	kt.Register("Delete", delete)
}

func delete(ctx *kt.Context) {
	ctx.RunRW(func(ctx *kt.Context) {
		ctx.RunAdmin(func(ctx *kt.Context) {
			testDelete(ctx, ctx.Admin)
		})
		ctx.RunNoAuth(func(ctx *kt.Context) {
			testDelete(ctx, ctx.NoAuth)
		})
	})
}

type deleteDoc struct {
	ID      string `json:"_id"`
	Rev     string `json:"_rev,omitempty"`
	Deleted bool   `json:"_deleted"`
}

func testDelete(ctx *kt.Context, client *kivik.Client) {
	ctx.Parallel()
	dbName := ctx.TestDBName()
	if err := ctx.Admin.CreateDB(kt.CTX, dbName); err != nil {
		ctx.Errorf("Failed to create test db '%s': %s", dbName, err)
		return
	}
	defer ctx.Admin.DestroyDB(kt.CTX, dbName)
	admdb, err := ctx.Admin.DB(kt.CTX, dbName)
	if err != nil {
		ctx.Errorf("Failed to connect to db as admin: %s", err)
	}
	db, err := client.DB(kt.CTX, dbName)
	if err != nil {
		ctx.Errorf("Failed to connect to db: %s", err)
		return
	}

	doc := &deleteDoc{
		ID: ctx.TestDBName(),
	}
	rev, err := admdb.Put(kt.CTX, doc.ID, doc)
	if err != nil {
		ctx.Errorf("Failed to create test doc: %s", err)
		return
	}
	doc.Rev = rev

	doc2 := &deleteDoc{
		ID: ctx.TestDBName(),
	}
	rev, err = admdb.Put(kt.CTX, doc2.ID, doc2)
	if err != nil {
		ctx.Errorf("Failed to create test doc: %s", err)
		return
	}
	doc2.Rev = rev

	ddoc := &testDoc{
		ID: "_design/foo",
	}
	rev, err = admdb.Put(kt.CTX, ddoc.ID, ddoc)
	if err != nil {
		ctx.Fatalf("Failed to create design doc in test db: %s", err)
	}
	ddoc.Rev = rev

	local := &testDoc{
		ID: "_local/foo",
	}
	rev, err = admdb.Put(kt.CTX, local.ID, local)
	if err != nil {
		ctx.Fatalf("Failed to create local doc in test db: %s", err)
	}
	local.Rev = rev

	ctx.Run("group", func(ctx *kt.Context) {
		ctx.Run("WrongRev", func(ctx *kt.Context) {
			ctx.Parallel()
			_, err := db.Delete(kt.CTX, doc2.ID, "1-9c65296036141e575d32ba9c034dd3ee")
			ctx.CheckError(err)
		})
		ctx.Run("InvalidRevFormat", func(ctx *kt.Context) {
			ctx.Parallel()
			_, err := db.Delete(kt.CTX, doc2.ID, "invalid rev format")
			ctx.CheckError(err)
		})
		ctx.Run("MissingDoc", func(ctx *kt.Context) {
			ctx.Parallel()
			_, err := db.Delete(kt.CTX, "missing doc", "1-9c65296036141e575d32ba9c034dd3ee")
			ctx.CheckError(err)
		})
		ctx.Run("ValidRev", func(ctx *kt.Context) {
			ctx.Parallel()
			_, err := db.Delete(kt.CTX, doc.ID, doc.Rev)
			ctx.CheckError(err)
		})
		ctx.Run("DesignDoc", func(ctx *kt.Context) {
			ctx.Parallel()
			_, err := db.Delete(kt.CTX, ddoc.ID, ddoc.Rev)
			ctx.CheckError(err)
		})
		ctx.Run("Local", func(ctx *kt.Context) {
			ctx.Parallel()
			_, err := db.Delete(kt.CTX, local.ID, local.Rev)
			ctx.CheckError(err)
		})
	})
}
