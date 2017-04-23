package db

import (
	"context"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	kt.Register("GetAttachmentMeta", getAttachmentMeta)
}

func getAttachmentMeta(ctx *kt.Context) {
	ctx.RunRW(func(ctx *kt.Context) {
		dbname := ctx.TestDB()
		defer ctx.Admin.DestroyDB(context.Background(), dbname, ctx.Options("db"))
		adb, err := ctx.Admin.DB(context.Background(), dbname, ctx.Options("db"))
		if err != nil {
			ctx.Fatalf("Failed to open db: %s", err)
		}

		doc := map[string]interface{}{
			"_id": "foo",
			"_attachments": map[string]interface{}{
				"foo.txt": map[string]interface{}{
					"content_type": "text/plain",
					"data":         "VGhpcyBpcyBhIGJhc2U2NCBlbmNvZGVkIHRleHQ=",
				},
			},
		}
		if _, err := adb.Put(context.Background(), "foo", doc); err != nil {
			ctx.Fatalf("Failed to create doc: %s", err)
		}

		ddoc := map[string]interface{}{
			"_id": "_design/foo",
			"_attachments": map[string]interface{}{
				"foo.txt": map[string]interface{}{
					"content_type": "text/plain",
					"data":         "VGhpcyBpcyBhIGJhc2U2NCBlbmNvZGVkIHRleHQ=",
				},
			},
		}
		if _, err := adb.Put(context.Background(), "_design/foo", ddoc); err != nil {
			ctx.Fatalf("Failed to create design doc: %s", err)
		}

		ctx.Run("group", func(ctx *kt.Context) {
			ctx.RunAdmin(func(ctx *kt.Context) {
				ctx.Parallel()
				testGetAttachmentMeta(ctx, ctx.Admin, dbname, "foo", "foo.txt")
				testGetAttachmentMeta(ctx, ctx.Admin, dbname, "foo", "NotFound")
				testGetAttachmentMeta(ctx, ctx.Admin, dbname, "_design/foo", "foo.txt")
			})
			ctx.RunNoAuth(func(ctx *kt.Context) {
				ctx.Parallel()
				testGetAttachmentMeta(ctx, ctx.NoAuth, dbname, "foo", "foo.txt")
				testGetAttachmentMeta(ctx, ctx.NoAuth, dbname, "foo", "NotFound")
				testGetAttachmentMeta(ctx, ctx.NoAuth, dbname, "_design/foo", "foo.txt")
			})
		})
	})
}

func testGetAttachmentMeta(ctx *kt.Context, client *kivik.Client, dbname, docID, filename string) {
	ctx.Run(docID+"/"+filename, func(ctx *kt.Context) {
		ctx.Parallel()
		db, err := client.DB(context.Background(), dbname, ctx.Options("db"))
		if err != nil {
			ctx.Fatalf("Failed to connect to db")
		}
		att, err := db.GetAttachmentMeta(context.Background(), docID, "", filename)
		if !ctx.IsExpectedSuccess(err) {
			return
		}
		if client.Driver() != "pouch" {
			if att.ContentType != "text/plain" {
				ctx.Errorf("Content-Type: Expected %s, Actual %s", "text/plain", att.ContentType)
			}
		}
	})
}
