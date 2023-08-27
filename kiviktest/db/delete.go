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
	kt.Register("Delete", _delete)
}

func _delete(ctx *kt.Context) {
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
	dbName := ctx.TestDB()
	defer ctx.DestroyDB(dbName)
	admdb := ctx.Admin.DB(dbName, ctx.Options("db"))
	if err := admdb.Err(); err != nil {
		ctx.Errorf("Failed to connect to db as admin: %s", err)
	}
	db := client.DB(dbName, ctx.Options("db"))
	if err := db.Err(); err != nil {
		ctx.Errorf("Failed to connect to db: %s", err)
		return
	}

	doc := &deleteDoc{
		ID: ctx.TestDBName(),
	}
	rev, err := admdb.Put(context.Background(), doc.ID, doc)
	if err != nil {
		ctx.Errorf("Failed to create test doc: %s", err)
		return
	}
	doc.Rev = rev

	doc2 := &deleteDoc{
		ID: ctx.TestDBName(),
	}
	rev, err = admdb.Put(context.Background(), doc2.ID, doc2)
	if err != nil {
		ctx.Errorf("Failed to create test doc: %s", err)
		return
	}
	doc2.Rev = rev

	ddoc := &testDoc{
		ID: "_design/foo",
	}
	rev, err = admdb.Put(context.Background(), ddoc.ID, ddoc)
	if err != nil {
		ctx.Fatalf("Failed to create design doc in test db: %s", err)
	}
	ddoc.Rev = rev

	local := &testDoc{
		ID: "_local/foo",
	}
	rev, err = admdb.Put(context.Background(), local.ID, local)
	if err != nil {
		ctx.Fatalf("Failed to create local doc in test db: %s", err)
	}
	local.Rev = rev

	ctx.Run("group", func(ctx *kt.Context) {
		ctx.Run("WrongRev", func(ctx *kt.Context) {
			ctx.Parallel()
			_, err := db.Delete(context.Background(), doc2.ID, "1-9c65296036141e575d32ba9c034dd3ee")
			ctx.CheckError(err)
		})
		ctx.Run("InvalidRevFormat", func(ctx *kt.Context) {
			ctx.Parallel()
			_, err := db.Delete(context.Background(), doc2.ID, "invalid rev format")
			ctx.CheckError(err)
		})
		ctx.Run("MissingDoc", func(ctx *kt.Context) {
			ctx.Parallel()
			_, err := db.Delete(context.Background(), "missing doc", "1-9c65296036141e575d32ba9c034dd3ee")
			ctx.CheckError(err)
		})
		ctx.Run("ValidRev", func(ctx *kt.Context) {
			ctx.Parallel()
			_, err := db.Delete(context.Background(), doc.ID, doc.Rev)
			ctx.CheckError(err)
		})
		ctx.Run("DesignDoc", func(ctx *kt.Context) {
			ctx.Parallel()
			_, err := db.Delete(context.Background(), ddoc.ID, ddoc.Rev)
			ctx.CheckError(err)
		})
		ctx.Run("Local", func(ctx *kt.Context) {
			ctx.Parallel()
			_, err := db.Delete(context.Background(), local.ID, local.Rev)
			ctx.CheckError(err)
		})
	})
}
