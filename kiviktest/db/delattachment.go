// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

package db

import (
	"context"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kiviktest/v4/kt"
)

func init() {
	kt.Register("DeleteAttachment", delAttachment)
}

func delAttachment(ctx *kt.Context) {
	ctx.RunRW(func(ctx *kt.Context) {
		dbname := ctx.TestDB()
		defer ctx.DestroyDB(dbname)
		ctx.Run("group", func(ctx *kt.Context) {
			ctx.RunAdmin(func(ctx *kt.Context) {
				ctx.Parallel()
				testDeleteAttachments(ctx, ctx.Admin, dbname, "foo.txt")
				testDeleteAttachments(ctx, ctx.Admin, dbname, "NotFound")
				testDeleteAttachmentsDDoc(ctx, ctx.Admin, dbname, "foo.txt")
				testDeleteAttachmentNoDoc(ctx, ctx.Admin, dbname)
			})
			ctx.RunNoAuth(func(ctx *kt.Context) {
				ctx.Parallel()
				testDeleteAttachments(ctx, ctx.NoAuth, dbname, "foo.txt")
				testDeleteAttachments(ctx, ctx.NoAuth, dbname, "NotFound")
				testDeleteAttachmentsDDoc(ctx, ctx.NoAuth, dbname, "foo.txt")
				testDeleteAttachmentNoDoc(ctx, ctx.NoAuth, dbname)
			})
		})
	})
}

func testDeleteAttachmentNoDoc(ctx *kt.Context, client *kivik.Client, dbname string) {
	db := client.DB(dbname, ctx.Options("db"))
	if err := db.Err(); err != nil {
		ctx.Fatalf("Failed to connect to db")
	}
	ctx.Run("NoDoc", func(ctx *kt.Context) {
		ctx.Parallel()
		_, err := db.DeleteAttachment(context.Background(), "nonexistantdoc", "2-4259cd84694a6345d6c534ed65f1b30b", "foo.txt")
		ctx.CheckError(err)
	})
}

func testDeleteAttachments(ctx *kt.Context, client *kivik.Client, dbname, filename string) {
	ctx.Run(filename, func(ctx *kt.Context) {
		doDeleteAttachmentTest(ctx, client, dbname, ctx.TestDBName(), filename)
	})
}

func testDeleteAttachmentsDDoc(ctx *kt.Context, client *kivik.Client, dbname, filename string) {
	ctx.Run("DesignDoc/"+filename, func(ctx *kt.Context) {
		doDeleteAttachmentTest(ctx, client, dbname, "_design/"+ctx.TestDBName(), filename)
	})
}

func doDeleteAttachmentTest(ctx *kt.Context, client *kivik.Client, dbname, docID, filename string) {
	db := client.DB(dbname, ctx.Options("db"))
	if err := db.Err(); err != nil {
		ctx.Fatalf("Failed to connect to db")
	}
	ctx.Parallel()
	adb := ctx.Admin.DB(dbname, ctx.Options("db"))
	if err := adb.Err(); err != nil {
		ctx.Fatalf("Failed to open db: %s", err)
	}
	doc := map[string]interface{}{
		"_id": docID,
		"_attachments": map[string]interface{}{
			"foo.txt": map[string]interface{}{
				"content_type": "text/plain",
				"data":         "VGhpcyBpcyBhIGJhc2U2NCBlbmNvZGVkIHRleHQ=",
			},
		},
	}
	rev, err := adb.Put(context.Background(), docID, doc)
	if err != nil {
		ctx.Fatalf("Failed to create doc: %s", err)
	}
	rev, err = db.DeleteAttachment(context.Background(), docID, rev, filename)
	if !ctx.IsExpectedSuccess(err) {
		return
	}
	var i interface{}
	if err = db.Get(context.Background(), docID, map[string]interface{}{"rev": rev}).ScanDoc(&i); err != nil {
		ctx.Fatalf("Failed to get deleted doc: %s", err)
	}
}
