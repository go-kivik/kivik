package db

import (
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	kt.Register("DeleteAttachment", delAttachment)
}

func delAttachment(ctx *kt.Context) {
	ctx.RunRW(func(ctx *kt.Context) {
		dbname := ctx.TestDBName()
		defer ctx.Admin.DestroyDB(dbname)
		if err := ctx.Admin.CreateDB(dbname); err != nil {
			ctx.Fatalf("Failed to create db: %s", err)
		}
		ctx.Run("group", func(ctx *kt.Context) {
			ctx.RunAdmin(func(ctx *kt.Context) {
				ctx.Parallel()
				testDeleteAttachments(ctx, ctx.Admin, dbname, "foo.txt")
				testDeleteAttachments(ctx, ctx.Admin, dbname, "NotFound")
			})
			ctx.RunNoAuth(func(ctx *kt.Context) {
				ctx.Parallel()
				testDeleteAttachments(ctx, ctx.NoAuth, dbname, "foo.txt")
				testDeleteAttachments(ctx, ctx.NoAuth, dbname, "NotFound")
			})
		})
	})
}

func testDeleteAttachments(ctx *kt.Context, client *kivik.Client, dbname, filename string) {
	ctx.Run(filename, func(ctx *kt.Context) {
		ctx.Parallel()
		adb, err := ctx.Admin.DB(dbname)
		if err != nil {
			ctx.Fatalf("Failed to open db: %s", err)
		}
		docID := ctx.TestDBName()
		doc := map[string]interface{}{
			"_id": docID,
			"_attachments": map[string]interface{}{
				"foo.txt": map[string]interface{}{
					"content_type": "text/plain",
					"data":         "VGhpcyBpcyBhIGJhc2U2NCBlbmNvZGVkIHRleHQ=",
				},
			},
		}
		rev, err := adb.Put(docID, doc)
		if err != nil {
			ctx.Fatalf("Failed to create doc: %s", err)
		}
		db, err := client.DB(dbname)
		if err != nil {
			ctx.Fatalf("Failed to connect to db")
		}
		rev, err = db.DeleteAttachment(docID, rev, filename)
		if !ctx.IsExpectedSuccess(err) {
			return
		}
		var x struct{}
		if err := db.Get(docID, &x, map[string]interface{}{"rev": rev}); err != nil {
			ctx.Fatalf("Failed to get deleted doc: %s", err)
		}
	})
}
